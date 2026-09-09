package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"text/template"

	"github.com/caerbannogwhite/enchanter/meta"
)

const (
	GOROUTINES                  = 4
	RESULT_VAR_NAME             = "result"
	RESULT_SIZE_VAR_NAME        = "resultSize"
	RESULT_NULL_MASK_VAR_NAME   = "resultNullMask"
	RESULT_IS_NULLABLE_VAR_NAME = "resultIsNullable"
	FINAL_RETURN_FMT            = "Errors{fmt.Sprintf(\"Cannot %s %%s and %%s\", s.Type().String(), o.Type().String())}"
)

type BuildInfo struct {
	OpCode        meta.OPCODE
	Op1Scalar     bool
	Op2Scalar     bool
	Op1VarName    string
	Op1SeriesType string
	Op1InnerType  meta.BaseType
	Op2VarName    string
	Op2SeriesType string
	Op2InnerType  meta.BaseType
	ResInnerType  meta.BaseType
	MakeOperation MakeOperationType
}

func (bi BuildInfo) UpdateScalarInfo(Op1Scalar, Op2Scalar bool) BuildInfo {
	bi.Op1Scalar = Op1Scalar
	bi.Op2Scalar = Op2Scalar
	return bi
}

// Generate the code that defines the result inner array and computes the
// result size and null mask. The second return value is the expression the
// return statement must use for the result's IsNullable_ field.
//
// Nullability is resolved at run time by the binaryNullMask helper in the
// series package, which collapses what used to be four generated variants
// (one per nullability combination) into a single call per length case.
func generateMakeResultStmt(info BuildInfo) ([]ast.Stmt, string) {
	// The result size is the length of the non-scalar operand; with two
	// scalars either length is 1, and the second operand is used.
	resSizeVariable := info.Op1VarName
	if info.Op1Scalar {
		resSizeVariable = info.Op2VarName
	}

	resultGoType := info.ResInnerType.ToGoType()

	// Special case for the result type
	if info.ResInnerType == meta.StringType {
		resultGoType = "[]*string"
	}

	// assign the result size
	stmts := []ast.Stmt{&ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: RESULT_SIZE_VAR_NAME},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.Ident{Name: fmt.Sprintf("%s.Len()", resSizeVariable)},
		},
	}}

	// The result is NAs: only the size is needed.
	if info.ResInnerType == meta.NullType {
		return stmts, "false"
	}

	// make the result array
	stmts = append(stmts, &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: RESULT_VAR_NAME},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.CallExpr{
				Fun: &ast.Ident{Name: "make"},
				Args: []ast.Expr{
					&ast.Ident{Name: resultGoType},
					&ast.Ident{Name: RESULT_SIZE_VAR_NAME},
				},
			},
		},
	})

	// Special case: one operand is NAs but the result is a typed series. The
	// result inherits a copy of the typed operand's null mask and is marked
	// not nullable, preserving the historical behavior of this case (see
	// naOperandNullMask).
	if info.Op1InnerType == meta.NullType || info.Op2InnerType == meta.NullType {
		nonNullOperand := info.Op1VarName
		nonNullOperandIsScalar := info.Op1Scalar
		if info.Op1InnerType == meta.NullType {
			nonNullOperand = info.Op2VarName
			nonNullOperandIsScalar = info.Op2Scalar
		}

		stmts = append(stmts, &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{Name: RESULT_NULL_MASK_VAR_NAME},
			},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.Ident{Name: fmt.Sprintf(
					"naOperandNullMask(%s.IsNullable_, %s.NullMask_, %v, %s)",
					nonNullOperand, nonNullOperand, nonNullOperandIsScalar, RESULT_SIZE_VAR_NAME)},
			},
		})

		// For the older operations this branch keeps the historical contract
		// of naOperandNullMask: the mask is copied but the result is marked
		// not nullable. Coalesce is new and has no history to preserve, so it
		// keeps the typed operand's nullability flag.
		if info.OpCode == meta.OP_BINARY_COALESCE {
			return stmts, fmt.Sprintf("%s.IsNullable_", nonNullOperand)
		}
		return stmts, "false"
	}

	// General case: one call resolves the mask and the nullability flag.
	// Coalesce is the one operation whose result is null only where both
	// operands are null, so it uses the AND-combining helper.
	maskHelper := "binaryNullMask"
	if info.OpCode == meta.OP_BINARY_COALESCE {
		maskHelper = "coalesceNullMask"
	}
	stmts = append(stmts, &ast.AssignStmt{
		Lhs: []ast.Expr{
			&ast.Ident{Name: RESULT_NULL_MASK_VAR_NAME},
			&ast.Ident{Name: RESULT_IS_NULLABLE_VAR_NAME},
		},
		Tok: token.DEFINE,
		Rhs: []ast.Expr{
			&ast.Ident{Name: fmt.Sprintf(
				"%s(%s.IsNullable_, %s.NullMask_, %v, %s.IsNullable_, %s.NullMask_, %v, %s)",
				maskHelper, info.Op1VarName, info.Op1VarName, info.Op1Scalar,
				info.Op2VarName, info.Op2VarName, info.Op2Scalar, RESULT_SIZE_VAR_NAME)},
		},
	})
	return stmts, RESULT_IS_NULLABLE_VAR_NAME
}

// Generate the code to compute the operation
func generateOperationLoop(info BuildInfo) []ast.Stmt {

	statements := make([]ast.Stmt, 0)

	if info.Op1Scalar && info.Op2Scalar {
		statements = append(statements, &ast.ExprStmt{
			X: info.MakeOperation(RESULT_VAR_NAME, "0", info.Op1VarName, "0", info.Op2VarName, "0"),
		})
	} else {

		op1Index := "i"
		op2Index := "i"
		if info.Op1Scalar {
			op1Index = "0"
		}

		if info.Op2Scalar {
			op2Index = "0"
		}

		statements = append(statements, &ast.ForStmt{
			Init: &ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.Ident{Name: "i"},
				},
				Tok: token.DEFINE,
				Rhs: []ast.Expr{
					&ast.Ident{Name: "0"},
				},
			},
			Cond: &ast.BinaryExpr{
				X:  &ast.Ident{Name: "i"},
				Op: token.LSS,
				Y:  &ast.Ident{Name: RESULT_SIZE_VAR_NAME},
			},
			Post: &ast.IncDecStmt{
				X:   &ast.Ident{Name: "i"},
				Tok: token.INC,
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{
						X: info.MakeOperation(RESULT_VAR_NAME, "i", info.Op1VarName, op1Index, info.Op2VarName, op2Index),
					},
				},
			},
		})
	}

	return statements
}

func generateOperation(info BuildInfo) []ast.Stmt {
	resSeriesType := computeResSeriesType(info.OpCode, info.Op1InnerType, info.Op2InnerType)

	// 1 - Generate the result inner data array, its size and its null mask.
	// isNullableExpr is the expression the return statement below must use
	// for the result's IsNullable_ field.
	statements, isNullableExpr := generateMakeResultStmt(info)

	// 2 - Generate the loop to compute the operation
	if resSeriesType != "NAs" {
		statements = append(statements, generateOperationLoop(info)...)
	}

	// 3 - Generate the return statement with the result series
	params := []ast.Expr{
		&ast.KeyValueExpr{
			Key:   &ast.Ident{Name: "IsNullable_"},
			Value: &ast.Ident{Name: isNullableExpr},
		},
		&ast.KeyValueExpr{
			Key:   &ast.Ident{Name: "NullMask_"},
			Value: &ast.Ident{Name: RESULT_NULL_MASK_VAR_NAME},
		},
	}

	switch resSeriesType {

	// NA: the only parameter is the size of the result series
	case "NAs":
		params = []ast.Expr{
			&ast.KeyValueExpr{
				Key:   &ast.Ident{Name: "size"},
				Value: &ast.Ident{Name: RESULT_SIZE_VAR_NAME},
			},
		}

	// BOOL Memory optimized: convert the result to a binary vector and add the size to the result series
	case "SeriesBoolMemOpt":
		params = append(params, &ast.KeyValueExpr{
			Key:   &ast.Ident{Name: "Data_"},
			Value: &ast.Ident{Name: fmt.Sprintf("boolVecToBinVec(%s)", RESULT_VAR_NAME)},
		})

		params = append(params, &ast.KeyValueExpr{
			Key:   &ast.Ident{Name: "size"},
			Value: &ast.Ident{Name: RESULT_SIZE_VAR_NAME},
		})

	// Default: just add the data to the result series
	default:
		params = append(params, &ast.KeyValueExpr{
			Key:   &ast.Ident{Name: "Data_"},
			Value: &ast.Ident{Name: RESULT_VAR_NAME},
		})

		params = append(params, &ast.KeyValueExpr{
			Key:   &ast.Ident{Name: "Ctx_"},
			Value: &ast.Ident{Name: fmt.Sprintf("%s.Ctx_", info.Op1VarName)},
		})
	}

	statements = append(statements, &ast.ReturnStmt{
		Results: []ast.Expr{
			&ast.CompositeLit{
				Type: &ast.Ident{Name: resSeriesType},
				Elts: params,
			},
		},
	})

	return statements
}

// Generate the flat switch over the four length cases: scalar-scalar,
// scalar-vector, vector-scalar and equal-length vectors. Every case ends in
// a return, so the trailing default return is reached only on a length
// mismatch between two vectors - exactly as the previous nested-if form
// behaved.
func generateSizeCheck(info BuildInfo, defaultReturn ast.Stmt) []ast.Stmt {
	lenCase := func(cond string, op1Scalar, op2Scalar bool) *ast.CaseClause {
		return &ast.CaseClause{
			List: []ast.Expr{ast.NewIdent(cond)},
			Body: generateOperation(info.UpdateScalarInfo(op1Scalar, op2Scalar)),
		}
	}

	op1Len := fmt.Sprintf("%s.Len()", info.Op1VarName)
	op2Len := fmt.Sprintf("%s.Len()", info.Op2VarName)

	return []ast.Stmt{
		&ast.SwitchStmt{
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					lenCase(fmt.Sprintf("%s == 1 && %s == 1", op1Len, op2Len), true, true),
					lenCase(fmt.Sprintf("%s == 1", op1Len), true, false),
					lenCase(fmt.Sprintf("%s == 1", op2Len), false, true),
					lenCase(fmt.Sprintf("%s == %s", op1Len, op2Len), false, false),
				},
			},
		},
		defaultReturn,
	}
}

// Generate the switch statement to handle the different types of the second operand
func generateSwitchType(
	operation Operation, op1SeriesType string, op1InnerType meta.BaseType,
	op1VarName, op2VarName string, defaultReturn ast.Stmt) []ast.Stmt {

	// Generate the preliminary type check, to check the type of second operand
	// is a Series or a raw value
	otherSeriesDefiniton := &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{{Name: "otherSeries"}},
					Type:  &ast.Ident{Name: "Series"},
				},
			},
		},
	}

	typeCheck := &ast.IfStmt{
		Init: &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("_, ok")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{ast.NewIdent("other.(Series)")},
		},
		Cond: &ast.Ident{Name: "ok"},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("otherSeries")},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{ast.NewIdent("other.(Series)")},
				},
			},
		},
		Else: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("otherSeries")},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{ast.NewIdent("NewSeries(other, nil, false, false, s.Ctx_)")},
				},
			},
		},
	}

	// Generate the context check
	contextCheck := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  &ast.Ident{Name: fmt.Sprintf("%s.Ctx_", op1VarName)},
			Op: token.NEQ,
			Y:  &ast.Ident{Name: fmt.Sprintf("%s.GetContext()", "otherSeries")},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{ast.NewIdent(fmt.Sprintf("Errors{fmt.Sprintf(\"Cannot operate on series with different contexts: %%v and %%v\", s.Ctx_, %s.GetContext())}", "otherSeries"))},
				},
			},
		},
	}

	op2VarNameTyped := "o"
	bigSwitch := &ast.TypeSwitchStmt{
		Assign: &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(op2VarNameTyped)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{ast.NewIdent(fmt.Sprintf("%s.(type)", "otherSeries"))},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		},
	}

	// Generate the switch cases for each type of the second operand
	for _, op2 := range operation.ApplyTo {
		bigSwitch.Body.List = append(bigSwitch.Body.List,
			&ast.CaseClause{
				List: []ast.Expr{ast.NewIdent(op2.SeriesName)},
				Body: generateSizeCheck(BuildInfo{
					OpCode:        operation.OpCode,
					Op1VarName:    op1VarName,
					Op1SeriesType: op1SeriesType,
					Op1InnerType:  op1InnerType,
					Op2VarName:    op2VarNameTyped,
					Op2SeriesType: op2.SeriesName,
					Op2InnerType:  op2.SeriesType,
					ResInnerType:  ComputeResInnerType(operation.OpCode, op1InnerType, op2.SeriesType),
					MakeOperation: op2.MakeOperation,
				}, defaultReturn),
			},
		)
	}

	bigSwitch.Body.List = append(bigSwitch.Body.List, &ast.CaseClause{
		List: nil,
		Body: []ast.Stmt{defaultReturn},
	})

	return []ast.Stmt{otherSeriesDefiniton, typeCheck, contextCheck, bigSwitch}
}

func computeResSeriesType(opCode meta.OPCODE, op1, op2 meta.BaseType) string {
	switch ComputeResInnerType(opCode, op1, op2) {
	case meta.NullType:
		return "NAs"
	case meta.BoolType:
		return "Bools"
	case meta.IntType:
		return "Ints"
	case meta.Int64Type:
		return "Int64s"
	case meta.Float32Type:
		return "SeriesFloat32"
	case meta.Float64Type:
		return "Float64s"
	case meta.StringType:
		return "Strings"
	case meta.TimeType:
		return "Times"
	case meta.DurationType:
		return "Durations"
	}
	return "Errors"
}

func ComputeResInnerType(opCode meta.OPCODE, op1, op2 meta.BaseType) meta.BaseType {
	return opCode.GetBinaryOpResultType(meta.Primitive{Base: op1}, meta.Primitive{Base: op2}).Base
}

func generateOperations() {
	for filename, info := range GenerateOperationsData() {

		src, err := os.ReadFile(filepath.Join(SERIES_FOLDER, filename))
		if err != nil {
			panic(err)
		}

		// Parse the file.
		fset := token.NewFileSet()
		fast, err := parser.ParseFile(fset, filepath.Join(SERIES_FOLDER, filename), src, parser.ParseComments)
		if err != nil {
			panic(err)
		}

		// // Add the utils package
		// fast.Decls = append(fast.Decls, &ast.GenDecl{
		// 	Tok: token.IMPORT,
		// 	Specs: []ast.Spec{
		// 		&ast.ImportSpec{Path: &ast.BasicLit{Value: `"github.com/caerbannogwhite/enchanter/utils"`}},
		// 	},
		// })

		for i, decl := range fast.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				switch funcDecl.Name.Name {
				case "And":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["And"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "AND"))},
						})

				case "Or":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Or"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "OR"))},
						})

				case "Mul":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Mul"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "multiply"))},
						})

				case "Div":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Div"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "divide"))},
						})

				case "Mod":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Mod"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "use modulo"))},
						})

				case "Exp":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Exp"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "use exponentiation"))},
						})

				case "Add":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Add"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "sum"))},
						})

				case "Sub":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Sub"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "subtract"))},
						})

				case "Eq":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Eq"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "compare for equality"))},
						})

				case "Ne":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Ne"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "compare for inequality"))},
						})

				case "Lt":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Lt"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "compare for less than"))},
						})

				case "Le":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Le"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "compare for less than or equal to"))},
						})

				case "Gt":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Gt"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "compare for greater than"))},
						})

				case "Ge":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Ge"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "compare for greater than or equal to"))},
						})

				case "Coalesce":
					fast.Decls[i].(*ast.FuncDecl).Body.List = generateSwitchType(
						info.Operations["Coalesce"], info.SeriesName, info.SeriesType, "s", "other",
						&ast.ReturnStmt{
							Results: []ast.Expr{ast.NewIdent(fmt.Sprintf(FINAL_RETURN_FMT, "coalesce"))},
						})
				}
			}

			buf := new(bytes.Buffer)
			err = format.Node(buf, fset, fast)
			if err != nil {
				panic(err)
			}

			err = os.WriteFile(filepath.Join(SERIES_FOLDER, filename), buf.Bytes(), 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}

func generateBase() {
	for filename, info := range DATA_BASE_METHODS {
		tmplBase, err := template.New("base").Parse(TEMPLATE_BASIC_ACCESSORS)
		if err != nil {
			panic(err)
		}

		tmplFilters, err := template.New("filters").Parse(TEMPLATE_FILTERS)
		if err != nil {
			panic(err)
		}

		tmplMaps, err := template.New("maps").Parse(TEMPLATE_MAPS)
		if err != nil {
			panic(err)
		}

		f, err := os.Create(filepath.Join(SERIES_FOLDER, filename))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		err = tmplBase.Execute(f, info)
		if err != nil {
			panic(err)
		}

		err = tmplFilters.Execute(f, info)
		if err != nil {
			panic(err)
		}

		err = tmplMaps.Execute(f, info)
		if err != nil {
			panic(err)
		}
	}
}

var SERIES_FOLDER = filepath.Join("..", "series")

func main() {
	generateBase()
	generateOperations()
}
