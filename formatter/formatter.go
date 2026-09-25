package formatter

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/caerbannogwhite/enchanter"
	"github.com/caerbannogwhite/enchanter/utils"
	"github.com/charmbracelet/lipgloss"
)

type Formatter interface {
	Compute()
	GetMaxWidth() int
	Push(val any)
	Format(width int, val any, isNa bool) string
}

const (
	cutoffDelta                = 3.14159265358979323846 * 1e-6 // pi * 10^-6 used in rounding last digits close to 5
	defaultDecimalDigits       = 11
	defaultThreshold           = -8
	defaultScientificThreshold = 9
	defaultMaxDigits           = 11
	defaultMovingDigits        = 3
	defaultNaText              = enchanter.NA_TEXT
	defaultInfText             = enchanter.INF_TEXT
)

type NumericFormatter struct {
	maxWidth            int       // Maximum width required to print the number.
	values              []float64 // The values to print.
	decimalDigits       int       // The number of digits to print after the decimal point.
	threshold           int       // The number of digits to print before the decimal point.
	scientificThreshold int       // The number of digits to print before switching to scientific notation.
	maxDigits           int       // The maximum number of digits to print.
	movingDigits        int       // The number of digits to print before the decimal point for very small numbers.
	naText              string    // The text to print for NaNs.
	infText             string    // The text to print for Infs.
	useExpFormat        bool      // Whether to use scientific notation.
	hasNegative         bool      // Whether negative numbers are present.
	useLipGloss         bool      // Whether to use lipgloss.
	justifyLeft         bool      // Whether to justify left.
	truncateOutput      bool      // Whether to truncate output.

	styleBold   lipgloss.Style
	styleItalic lipgloss.Style
	styleNa     lipgloss.Style
	styleNum    lipgloss.Style
	styleNumNeg lipgloss.Style
	styleNone   lipgloss.Style
}

func NewNumericFormatter() *NumericFormatter {
	return &NumericFormatter{
		maxWidth:            0,
		values:              make([]float64, 0),
		decimalDigits:       defaultDecimalDigits,
		threshold:           defaultThreshold,
		scientificThreshold: defaultScientificThreshold,
		maxDigits:           defaultMaxDigits,
		movingDigits:        defaultMovingDigits,
		naText:              defaultNaText,
		infText:             defaultInfText,
		useExpFormat:        false,
		hasNegative:         false,
		useLipGloss:         false,
		justifyLeft:         false,
		truncateOutput:      false,

		styleBold:   lipgloss.NewStyle().Bold(true),
		styleItalic: lipgloss.NewStyle().Italic(true),
		styleNa:     lipgloss.NewStyle().Bold(true).Copy().Foreground(lipgloss.Color("#c00020")),
		styleNum:    lipgloss.NewStyle().Foreground(lipgloss.Color("#0080c0")),
		styleNumNeg: lipgloss.NewStyle().Foreground(lipgloss.Color("#c04000")),
		styleNone:   lipgloss.NewStyle(),
	}
}

func (f *NumericFormatter) SetDecimalDigits(decimalDigits int) *NumericFormatter {
	f.decimalDigits = decimalDigits
	return f
}

func (f *NumericFormatter) SetThreshold(threshold int) *NumericFormatter {
	f.threshold = threshold
	return f
}

func (f *NumericFormatter) SetScientificThreshold(scientificThreshold int) *NumericFormatter {
	f.scientificThreshold = scientificThreshold
	return f
}

func (f *NumericFormatter) SetMaxDigits(maxDigits int) *NumericFormatter {
	f.maxDigits = maxDigits
	return f
}

func (f *NumericFormatter) SetMovingDigits(movingDigits int) *NumericFormatter {
	f.movingDigits = movingDigits
	return f
}

func (f *NumericFormatter) SetNaText(naText string) *NumericFormatter {
	f.naText = naText
	return f
}

func (f *NumericFormatter) SetInfText(infText string) *NumericFormatter {
	f.infText = infText
	return f
}

func (f *NumericFormatter) SetUseLipGloss(useLipGloss bool) *NumericFormatter {
	f.useLipGloss = useLipGloss
	return f
}

func (f *NumericFormatter) SetJustifyLeft(justifyLeft bool) *NumericFormatter {
	f.justifyLeft = justifyLeft
	return f
}

func (f *NumericFormatter) SetTruncateOutput(truncateOutput bool) *NumericFormatter {
	f.truncateOutput = truncateOutput
	return f
}

func (f *NumericFormatter) Compute() {
	f.maxWidth = 0

	useLipGloss := f.useLipGloss
	truncateOutput := f.truncateOutput

	f.useLipGloss = false
	f.truncateOutput = false

	var s string
	for _, val := range f.values {
		s = f.Format(1, val, false)
		f.maxWidth = max(f.maxWidth, len(s))
	}

	f.useLipGloss = useLipGloss
	f.truncateOutput = truncateOutput
}

func (f *NumericFormatter) GetMaxWidth() int {
	return f.maxWidth
}

func (f *NumericFormatter) Push(val any) {
	var num float64
	switch val.(type) {
	case int:
		num = float64(val.(int))
	case int64:
		num = float64(val.(int64))
	case float64:
		num = val.(float64)
	default:
		return
	}

	f.values = append(f.values, num)

	if math.IsNaN(num) {
		f.maxWidth = max(f.maxWidth, len(f.naText))
		return
	}

	if math.IsInf(num, 1) {
		f.maxWidth = max(f.maxWidth, len(f.infText))
		return
	}

	if math.IsInf(num, -1) {
		f.maxWidth = max(f.maxWidth, len(f.infText)+1)
		f.hasNegative = true
		return
	}

	absNum := math.Abs(num)
	if f.useExpFormat || absNum <= 1e-50 {
		return
	}

	// Set exponential format if very large.
	if absNum >= math.Pow(10, float64(f.scientificThreshold))-0.5-0.1*cutoffDelta { // 999 999 999.4999...
		f.useExpFormat = true
		return
	}

	// Check if very small, here more care is needed because of rounding.
	signif, exponent := sigAndExp(num)

	// If the number would be rounded up, there might be less digits needed / no switch to exp format.
	// To check this, determine the number of digits that would be needed for this number:
	digits := min(f.maxDigits, max(0, f.movingDigits-exponent))

	// If this is 0, no need to check anything, otherwise find out where the rounding digit is related to the exponent.
	if digits > 0 && (math.Abs(signif) >= 10-math.Pow(0.1, float64(min(f.maxDigits+exponent, f.movingDigits))+1)*(5+cutoffDelta)) { // this is for rounding issues
		signif = 1 //nolint:ineffassign // dead store kept intentionally, mirrors the equivalent rounding-correction branch in render() for readability
		exponent++
	}

	if exponent < 2-f.scientificThreshold { // -7
		f.useExpFormat = true

	} else if exponent > f.threshold && f.decimalDigits > 0 {
		f.decimalDigits = min(f.maxDigits, max(0, f.movingDigits-exponent))
		f.threshold = exponent
	}
}

func (f *NumericFormatter) Format(width int, val any, isNa bool) string {
	var num float64
	switch val.(type) {
	case int:
		num = float64(val.(int))
	case int64:
		num = float64(val.(int64))
	case float64:
		num = val.(float64)
	default:
		return f.naText
	}

	if math.IsInf(num, -1) {
		return f.render(width, fmt.Sprintf("-%s", f.infText), f.styleNumNeg)
	}
	if math.IsInf(num, 1) {
		return f.render(width, f.infText, f.styleNum)
	}
	if math.IsNaN(num) || isNa {
		return f.render(width, f.naText, f.styleNa)
	}

	// Very small numbers, which are treated as zero, are formatted as 0.
	if math.Abs(num) <= 1e-50 {
		return f.render(width, "0", f.styleNum)
	}

	signif, exponent := sigAndExp(num)
	if f.useExpFormat {
		if math.Abs(signif) >= 9.9995-1e-4*cutoffDelta { // to avoid printing "10.000 x 10^exp" when signif rounded up, eg. 9.99999
			if signif < 0 {
				signif = -1.0
			} else {
				signif = 1.0
			}
			exponent++
		}

		return f.render(width, fmt.Sprintf("%se%d", fmt.Sprintf("%.*f", 3, signif), exponent), f.styleNum)
	}

	if exponent < -f.decimalDigits && f.decimalDigits != f.maxDigits {
		minNeededDigits := -exponent - 1
		if math.Abs(signif) >= 9.5-cutoffDelta { // if the last digit would be 9, check if rounded up or down
			minNeededDigits++
		}

		return f.render(width, fmt.Sprintf("%.*f", minNeededDigits, num), f.styleNum)
	}

	return f.render(width, fmt.Sprintf("%.*f", f.decimalDigits, num), f.styleNum)
}

func (f *NumericFormatter) render(width int, s string, style lipgloss.Style) string {
	if f.truncateOutput {
		s = utils.Truncate(s, width)
	}

	if f.useLipGloss {
		if f.justifyLeft {
			return style.Render(fmt.Sprintf("%-*s", width, s))
		} else {
			return style.Render(fmt.Sprintf("%*s", width, s))
		}
	}

	if f.justifyLeft {
		return fmt.Sprintf("%-*s", width, s)
	} else {
		return fmt.Sprintf("%*s", width, s)
	}
}

func sigAndExp(num float64) (float64, int) {
	exponent := int(math.Floor(math.Log10(math.Abs(num))))
	signif := num * math.Pow(10, -float64(exponent))
	if math.Abs(signif) >= 10 { // extra check for possible rounding wierdness
		exponent++
		signif /= 10
	}

	return signif, exponent
}

type StringFormatter struct {
	useLipGloss bool
	maxWidth    int
	lengths     []int

	styleString lipgloss.Style
	styleNa     lipgloss.Style
}

func NewStringFormatter() *StringFormatter {
	return &StringFormatter{
		useLipGloss: false,
		lengths:     make([]int, 0),
		styleString: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#98AFC7")),
		styleNa:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#c00020")),
	}
}

func (f *StringFormatter) SetUseLipGloss(useLipGloss bool) *StringFormatter {
	f.useLipGloss = useLipGloss
	return f
}

// Compute sets the column width to the 80th-percentile string length, so a
// few long outliers get truncated instead of widening the whole column.
func (f *StringFormatter) Compute() {
	if len(f.lengths) == 0 {
		f.maxWidth = 0
		return
	}
	sort.IntSlice(f.lengths).Sort()
	idx := max(int(math.Floor(0.8*float64(len(f.lengths))))-1, 0)
	f.maxWidth = f.lengths[idx]
}

func (f *StringFormatter) GetMaxWidth() int {
	return f.maxWidth
}

func (f *StringFormatter) Push(val any) {
	if s, ok := val.(string); ok {
		f.lengths = append(f.lengths, len(toPrintable(s)))
	}
}

func (f *StringFormatter) Format(width int, val any, isNa bool) string {
	if s, ok := val.(string); ok && !isNa {
		s = toPrintable(s)
		if f.useLipGloss {
			return f.styleString.Render(fmt.Sprintf("%-*s", width, utils.Truncate(s, width)))
		} else {
			return fmt.Sprintf("%-*s", width, utils.Truncate(s, width))
		}
	}
	if f.useLipGloss {
		return f.styleNa.Render(fmt.Sprintf("%-*s", width, utils.Truncate(enchanter.NA_TEXT, width)))
	}
	return fmt.Sprintf("%-*s", width, utils.Truncate(enchanter.NA_TEXT, width))
}

func toPrintable(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return '.'
	}, s)
}
