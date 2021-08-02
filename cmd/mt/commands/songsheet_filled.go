package commands

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/cobra"
)

var (
	SongsheetFilledCmd = &cobra.Command{
		Use:   "songsheet-filled [qu-id]",
		Short: "print filled songsheet from qu id",
		Args:  cobra.ExactArgs(1),
		RunE:  songsheetFilledCmd,
	}
)

func init() {
	SongsheetFilledCmd.PersistentFlags().BoolVar(
		&mirrorStringsOrderFlag, "mirror", false, "mirror string positions")
	RootCmd.AddCommand(SongsheetFilledCmd)
}

func songsheetFilledCmd(cmd *cobra.Command, args []string) error {

	pdf := gofpdf.New("P", "in", "Letter", "")
	pdf.SetMargins(0, 0, 0)
	pdf.AddPage()

	elemKinds := []tssElement{
		singleSpacing{},
		singleAnnotatedSine{},
		singleLyricLine{},
	}

	// XXX
	//quid, err := parseElem(args[0])
	//if err != nil {
	//return err
	//}

	// get contents of songsheet

	bnd := bounds{padding, padding, 11, 8.5}
	if headerFlag {
		bnd = printHeader(pdf, bnd, nil) // XXX
	}
	_ = elem.printPDF(pdf, bnd)
	return pdf.OutputFileAndClose(fmt.Sprintf("songsheet_%v.pdf", title))
}

// ---------------------

// whole text songsheet element
type tssElement interface {
	ssElement
	parseText(text string) (reducedText string, elem tssElement, err error)
}

// ---------------------

type chordChart struct {
	chords []Chord
}

type Chord struct {
	name   string // must be 1 or 2 characters
	chords []int  // from thick to thin guitar strings
}

var _ tssElement = chordChart{}

func (c chordChart) parseText(text string) (reduced string, elem tssElement, err error) {
	lines := strings.Split(text, "\n")
	if len(lines) < 9 {
		return text, elem, fmt.Errorf("improper number of input lines, want 1 have %v", len(lines))
	}

	// checking form, must be in the pattern as such:
	//  |  |  |
	//- 1  3
	//- 0  2
	//- 3  0
	//- 0  0
	//- 1  1
	//- 0  0
	//  |  |  |
	//  F  G  C
	if !strings.HasPrefix(lines[0], "  |  |  |") {
		return text, elem, fmt.Errorf("not a chord chart (line 1)")
	}
	if !strings.HasPrefix(lines[7], "  |  |  |") {
		return text, elem, fmt.Errorf("not a chord chart (line 7)")
	}
	for i := 1; i <= 6; i++ {
		if !strings.HasPrefix(lines[i], "- ") {
			return text, elem, fmt.Errorf("not a chord chart (line %v)", i)
		}
	}

	// get the chords
	chordNames := lines[8]
	for j := 2; j < len(chordNames); j += 3 {

		if chordNames[j] == ' ' {
			// this chord is not labelled, must be the end of the chords
			break
		}

		c := chord{name: string(chordNames[j])}

		// add the second character to the name (if it exists)
		if j+1 < len(chordNames) {
			if chordNames[j+1] != ' ' {
				c.name = append(c.name, string(chordNames[j+1]))
			}
		}

		// add all the guitar strings
		for i := 1; i <= 6; i++ {
			word := string(lines[i][j])

			if j+1 < len(lines[i]) {
				if lines[i][j+1] != ' ' {
					word := append(word, string(lines[i][j+1]))
				}
			}
			num, err := strconv.Atoi(word)
			if err != nil {
				return text, elem, fmt.Errorf("bad number conversion for %v at %v,%v", word, j, i)
			}
			c.chords = append(c.chords, num)
		}
	}

	// chop off the first 9 lines
	return text[9:], c, nil
}

func (c chordChart) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reduced bounds) {
	return bounds{bnd.bottom, bnd.left, bnd.bottom, bnd.left} // XXX
}

func (c chordChart) getWidth() (isStatic bool, width float64) {
	return false, 0
}

func (c chordChart) getHeight() (isStatic bool, height float64) {
	return true, 0 // XXX WHAT IS IT
}

// ---------------------

type singleSpacing struct{}

var _ tssElement = singleSpacing{}

func (s singleSpacing) parseText(text string) (elem tssElement, err error) {
	lines := strings.Split(text, "\n")
	if len(lines) != 1 {
		return elem, fmt.Errorf("improper number of input lines, want 1 have %v", len(lines))
	}
	if len(strings.TrimSpace(lines[0])) != 0 {
		return elem, errors.New("blank line contains content")
	}
	return singleSpacing{}, nil
}

func (s singleSpacing) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reduced bounds) {
	return bounds{bnd.bottom, bnd.left, bnd.bottom, bnd.left} // XXX
}

func (s singleSpacing) getWidth() (isStatic bool, width float64) {
	return false, 0
}

func (s singleSpacing) getHeight() (isStatic bool, height float64) {
	return true, 0 // XXX WHAT IS IT
}

// ---------------------

type singleLineLyrics struct {
	lyricLocos  []stringInt
	melodyLocos []stringInt
}

var _ tssElement = singleLineLyrics{}

type stringInt struct {
	s string
	i int
}

// split up a string into its constituent
// words and write record each words locations
func getLocos(str string) (out []stringInt) {

	locationCount := 0

	remaining := str
	for {
		spl := strings.SplitN(remaining, " ", 2)
		locationCount++ // for the space
		if len(spl) > 0 {
			if len(spl[0]) == 0 {
				continue
			}
			out = append(out, stringInt{spl[0], locationCount})
			locationCount += len(spl[0])

			if len(spl) == 1 {
				break
			}
			if len(spl) == 2 {
				remaining = spl[1]
			}
		} else {
			break
		}
	}

}

func (s singleLineLyrics) parseText(text string) (elem tssElement, err error) {
	lines := strings.Split(text, "\n")
	if len(lines) != 2 {
		return elem, fmt.Errorf("improper number of input lines, want 1 have %v", len(lines))
	}

	// lyrics

	lyricWords := strings.Split(text, " ")
	lyricMelodies := strings.Split(text, " ")

	if len(lyricWords) != len(lyricMelodies) {
		return elem, fmt.Errorf("different number of lyrics (%v) and melodies (%v) for:\n%v",
			lyricWords, lyricMelodies, text)
	}

	return singleLineLyrics{}, nil
}

func (s singleLineLyrics) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reduced bounds) {
	return bounds{bnd.bottom, bnd.left, bnd.bottom, bnd.left} // XXX
}

func (s singleLineLyrics) getWidth() (isStatic bool, width float64) {
	return false, 0
}

func (s singleLineLyrics) getHeight() (isStatic bool, height float64) {
	return true, 0 // XXX WHAT IS IT
}

// ---------------------

type singleAnnotatedSine struct {
	humps         float64
	boldedCentral []sineAnnotation
	alongSine     []sineAnnotation
}

type sineAnnotation struct {
	position float64 // in humps
	ch       rune    // annotation text
}

var _ tssElement = singleAnnotatedSine{}

func (s singleAnnotatedSine) parseText(text string) (elem tssElement, err error) {

	// the annotated sine must come in 4 lines
	//    ex.   desciption
	// 1)   _    text representation of the sine humps (top)
	// 2) _/ \_  text representation of the sine humps (bottom)
	// 3) F      bolded central annotations
	// 4)   ^ v  annotations along the sine curve

	lines := strings.Split(text, "\n")
	if len(lines) != 4 {
		return elem, fmt.Errorf("improper number of input lines, want 4 have %v", len(lines))
	}

	humpsChars := len(lines[0])
	if humpsChars < len(lines[1]) {
		humpsChars = len(lines[1])
	}
	humps := float64(humpsChars) / 4

	boldedCentral := make([]sineAnnotation)
	for pos, ch := range lines[2] {
		if ch == ' ' {
			continue
		}
		boldedCentral = append(boldedCentral,
			[]sineAnnotation{float64(pos) / 4, ch})
	}

	alongSine := make([]sineAnnotation)
	for pos, ch := range lines[3] {
		if ch == ' ' {
			continue
		}
		alongSine = append(alongSine,
			[]sineAnnotation{float64(pos) / 4, ch})
	}
}

func (s singleAnnotatedSine) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reduced bounds) {

	// print ^ and v in special way as they represent up and down strokes

	// Print the sine function
	pdf.SetLineWidth(thinestLW)
	resolution := 10
	amplitude := 0.04 // XXX
	width := bnd.right - bnd.left - padding
	frequency := math.Pi * 2 * s.humps / width
	yStart := bnd.top + s.shift
	xStart := bnd.left
	xEnd := bnd.right - padding
	lastPointX := xStart
	lastPointY := yStart
	for eqX := float64(0); true; eqX += 10 {
		if xStart+eqX > xEnd {
			break
		}
		eqY := amplitude * math.Cos(frequency*eqX)
		if eqX > 0 {
			pdf.Line(lastPointX, lastPointY, xStart+eqX, yStart+eqY)
		}
		lastPointX = xStart + eqX
		lastPointY = yStart + eqY
	}

	// Print the bold central
	// TODO

	// print the floating

	return bounds{bnd.bottom, bnd.left, bnd.bottom, bnd.left}
}

func (s singleAnnotatedSine) getWidth() (isStatic bool, width float64) {
	return false, 0
}

func (s singleAnnotatedSine) getHeight() (isStatic bool, height float64) {
	return true, 0 // XXX WHAT IS IT
}
