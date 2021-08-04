package commands

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/jung-kurt/gofpdf"
	"github.com/rigelrozanski/thranch/quac"
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
		chordChart{},
		singleSpacing{},
		singleAnnotatedSine{},
		singleLineLyrics{},
	}

	quid, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	content, found := quac.GetContentByID(uint32(quid))
	if !found {
		return fmt.Errorf("could not find anything under id: %v", quid)
	}
	lines := strings.Split(string(content), "\n")

	// get the header
	lines, hc, err := parseHeader(lines)

	// get contents of songsheet
	// parse all the elems
	parsedElems := []tssElement{}

OUTER:
	if len(lines) > 0 {
		for _, elem := range elemKinds {
			reduced, newElem, err := elem.parseText(lines)
			if err == nil {
				lines = reduced
				parsedElems = append(parsedElems, newElem)
				goto OUTER
			}
		}
		return fmt.Errorf("could not parse song at line %v", lines[0])
	}

	// XXX print the songsheet

	bnd := bounds{padding, padding, 11, 8.5}
	if headerFlag {
		bnd = printHeader(pdf, bnd, hc)
	}
	//_ = elem.printPDF(pdf, bnd)
	//return pdf.OutputFileAndClose(fmt.Sprintf("songsheet_%v.pdf", title))
	return nil // XXX
}

func parseHeader(lines []string) (reduced []string, hc headerContent, err error) {
	if len(lines) < 2 {
		return lines, hc, fmt.Errorf("improper number of input lines, want at least 2 have %v", len(lines))
	}

	splt := strings.SplitN(lines[0], "DATE:", 2)
	switch len(splt) {
	case 1:
		hc.title = splt[0]
	case 2:
		hc.title = splt[0]
		hc.date = splt[1]
	default:
		panic("")
	}

	flds := strings.Fields(lines[1])
	keywords := map[string]string{
		"TUNING:":  "",
		"CAPO:":    "",
		"TIMESIG:": "",
		"FEEL:":    "",
	}
	fldIsKeyword := make([]bool, len(flds))
	for i, fld := range flds {
		if _, found := keywords[keywords]; found {
			fldIsKeyword[i] = true
		}
	}

	for keyword, _ := range keywords {
		keywordFound := false
		keywordText := ""
		for i, fld := range flds {
			switch {
			case !keywordFound && keyword == fld:
				keywordFound = true
				continue
			case keywordFound && keyword == fld:
				break
			case keywordFound && keyword != fld && keywordText == "":
				keywordText = keyword
			case keywordFound && keyword != fld && keywordText != "":
				keywordText += " " + keyword
			}
		}
		if keywordFound {
			keywords[keyword] = keywordText
		}
	}

	// assign keywords to the header content
	hc.tuning = keywords["TUNING:"]
	hc.capo = keywords["CAPO:"]
	hc.timesig = keywords["TIMESIG:"]
	hc.feel = keywords["FEEL:"]

	return lines[2:], hc, nil
}

// ---------------------

// whole text songsheet element
type tssElement interface {
	printPDF(*gofpdf.Fpdf, bounds) (reduced bounds)
	getWidth() (isStatic bool, width float64)   // width is only valid if isStatic=true
	getHeight() (isStatic bool, height float64) // height is only valid if isStatic=true
	parseText(lines []string) (reducedLines []string, elem tssElement, err error)
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

func (c chordChart) parseText(lines []string) (reduced []string, elem tssElement, err error) {
	if len(lines) < 9 {
		return lines, elem, fmt.Errorf("improper number of input lines, want at least 9 have %v", len(lines))
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
		return lines, elem, fmt.Errorf("not a chord chart (line 1)")
	}
	if !strings.HasPrefix(lines[7], "  |  |  |") {
		return lines, elem, fmt.Errorf("not a chord chart (line 7)")
	}
	for i := 1; i <= 6; i++ {
		if !strings.HasPrefix(lines[i], "- ") {
			return lines, elem, fmt.Errorf("not a chord chart (line %v)", i)
		}
	}

	// get the chords
	chordNames := lines[8]
	for j := 2; j < len(chordNames); j += 3 {

		if chordNames[j] == ' ' {
			// this chord is not labelled, must be the end of the chords
			break
		}

		c := Chord{name: string(chordNames[j])}

		// add the second character to the name (if it exists)
		if j+1 < len(chordNames) {
			if chordNames[j+1] != ' ' {
				c.name += string(chordNames[j+1])
			}
		}

		// add all the guitar strings
		for i := 1; i <= 6; i++ {
			word := string(lines[i][j])

			if j+1 < len(lines[i]) {
				if lines[i][j+1] != ' ' {
					word += string(lines[i][j+1])
				}
			}
			num, err := strconv.Atoi(word)
			if err != nil {
				return lines, elem, fmt.Errorf("bad number conversion for %v at %v,%v", word, j, i)
			}
			c.chords = append(c.chords, num)
		}
	}

	// chop off the first 9 lines
	return lines[9:], c, nil
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

func (s singleSpacing) parseText(lines []string) (reduced []string, elem tssElement, err error) {
	if len(lines) < 1 {
		return lines, elem, fmt.Errorf("improper number of input lines, want 1 have %v", len(lines))
	}
	if len(strings.TrimSpace(lines[0])) != 0 {
		return lines, elem, errors.New("blank line contains content")
	}
	return lines[1:], singleSpacing{}, nil
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
	lyrics   string
	melodies []melody
}

type melody struct {
	blank              bool // no melody here, this is just a placeholder
	num                rune
	modifierIsAboveNum bool // otherwise below
	modifier           rune // either '.', '-', or '~'
}

var _ tssElement = singleLineLyrics{}

func (s singleLineLyrics) parseText(lines []string) (reduced []string, elem tssElement, err error) {
	if len(lines) < 4 {
		return lines, elem, fmt.Errorf("improper number of input lines, want 1 have %v", len(lines))
	}

	sll := singleLineLyrics{}
	sll.lyrics = lines[0]

	upperMods := lines[1]
	melodyNums := lines[2]
	lowerMods := lines[3]
	for i, r := range melodyNums {
		if !(unicode.IsSpace(r) || unicode.IsNumber(r)) {
			return lines, elem, fmt.Errorf(
				"melodies line contains something other"+
					"than numbers and spaces (rune: %v, col: %v)", r, i)
		}
		if unicode.IsSpace(r) {
			sll.melodies = append(sll.melodies, melody{blank: true})
		} else {
			m := melody{blank: false, num: r}
			if len(upperMods) > i && !unicode.IsSpace(rune(upperMods[i])) {
				m.modifierIsAboveNum = true
				m.modifier = rune(upperMods[i])
			} else if len(lowerMods) > i && !unicode.IsSpace(rune(lowerMods[i])) {
				m.modifierIsAboveNum = false
				m.modifier = rune(lowerMods[i])
			} else { // there must be no modifier
				return lines, elem, fmt.Errorf("no melody modifier for the melody")
			}

			// ensure that the melody modifier has a valid rune
			if !(m.modifier == '.' || m.modifier == '-' || m.modifier == '~') {
				return lines, elem, fmt.Errorf(
					"bad modifier not '.', '-', or '~' (have %v)", m.modifier)
			}

			sll.melodies = append(sll.melodies, m)
		}

	}

	return lines[4:], sll, nil
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

func GetFontPt(heightInches float64) float64 {
	return heightInches * 72.281142957923
}

func GetCourierFontWidth(height float64) float64 {
	return 0.43 * height
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

func (s singleAnnotatedSine) parseText(lines []string) (reduced []string, elem tssElement, err error) {

	// the annotated sine must come in 4 lines
	//    ex.   desciption
	// 1) F              bolded central annotations
	// 2) _   _   _   _  text representation of the sine humps (top)
	// 3)  \_/ \_/ \_/   text representation of the sine humps (bottom)
	// 4)   ^   ^ 1   v  annotations along the sine curve

	if len(lines) < 4 {
		return lines, elem, fmt.Errorf("improper number of input lines, want 4 have %v", len(lines))
	}

	// ensure that the first two lines start with at least 3 sine humps
	//_   _   _
	// \_/ \_/ \_/
	if !(lines[1].HasPrefix("_   _   _") &&
		lines[2].HasPrefix(" \\_/ \\_/ \\_/")) {
		return lines, elem, fmt.Errorf("first lines are not sine humps")
	}

	humpsChars := len(lines[1]) // TODO may run into issues here with suffix whitespace
	if humpsChars < len(lines[2]) {
		humpsChars = len(lines[2])
	}
	humps := float64(humpsChars) / 4

	boldedCentral := []sineAnnotation{}
	for pos, ch := range lines[0] {
		if ch == ' ' {
			continue
		}
		boldedCentral = append(boldedCentral,
			sineAnnotation{float64(pos) / 4, ch})
	}

	alongSine := []sineAnnotation{}
	for pos, ch := range lines[3] {
		if ch == ' ' {
			continue
		}
		alongSine = append(alongSine,
			sineAnnotation{float64(pos) / 4, ch})
	}

	sas := singleAnnotatedSine{
		humps:         humps,
		boldedCentral: boldedCentral,
		alongSine:     alongSine,
	}

	return lines[4:], sas, nil
}

func (s singleAnnotatedSine) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reduced bounds) {

	// Print the sine function
	pdf.SetLineWidth(thinestLW)
	resolution := 10.0
	amplitude := 0.04 // XXX
	width := bnd.right - bnd.left - padding
	frequency := math.Pi * 2 * s.humps / width
	yStart := bnd.top + amplitude // because the equation may be negative sending to bnd.top
	xStart := bnd.left
	xEnd := bnd.right - padding
	lastPointX := xStart
	lastPointY := yStart
	for eqX := float64(0); true; eqX += resolution {
		if xStart+eqX > xEnd {
			break
		}
		eqY := amplitude * math.Cos(frequency*eqX)
		if eqX > 0 {
			pdf.Line(lastPointX, lastPointY, xStart+eqX, yStart-eqY) // -eqY because starts from topleft corner
		}
		lastPointX = xStart + eqX
		lastPointY = yStart + eqY
	}

	// print the bold central
	fontH := amplitude / 2
	fontPt := GetFontPt(fontH)
	pdf.SetFont("courier", "B", fontPt)
	for i, bs := range s.boldedCentral {
		X := xStart + bs.position
		Y := yStart + fontH/2 // so the text is centered along the sine axis
		pdf.Text(X, Y, string(bs.ch))
	}

	// print the characters along the sine curve
	for i, as := range s.alongSine {
		if as.ch == ' ' {
			continue
		}

		// determine hump position
		eqY := amplitude * math.Cos(frequency*as.position)

		// character height which extends beyond the sine curve
		chhbs := amplitude / 4

		// TODO figure out some way to allow for emphasised lines
		//      maybe doubling the lines up
		//      maybe using V and A  ???
		switch as.ch {
		case 'v':
			tipX := xStart + as.position
			tipY := yStart + eqY
			// 45deg angles to the tip
			pdf.Line(tipX-chhbs, tipY-chhbs, tipX, tipY)
			pdf.Line(tipX, tipY, tipX+chhbs, tipY-chhbs)
		case '^':
			tipX := xStart + as.position
			tipY := yStart + eqY
			// 45deg angles to the tip
			pdf.Line(tipX-chhbs, tipY+chhbs, tipX, tipY)
			pdf.Line(tipX, tipY, tipX+chhbs, tipY+chhbs)
		case '|':
			x := xStart + as.position
			pdf.Line(x, yStart-amplitude-chhbs, x, yStart+amplitude+chhbs)

		default:
			h := 2 * chhbs // font height in inches
			fontPt := GetFontPt(h)
			w := GetCourierFontWidth(h) // font width

			// we want the character to be centered about the sine curve
			pdf.SetFont("courier", "", fontPt)
			tipX := xStart + as.position
			tipY := yStart + eqY
			pdf.Text(tipX-(w/2), tipY+(h/2), string(as.ch))
		}
	}

	return bounds{bnd.bottom, bnd.left, bnd.bottom, bnd.left} // XXX
}

func (s singleAnnotatedSine) getWidth() (isStatic bool, width float64) {
	return false, 0
}

func (s singleAnnotatedSine) getHeight() (isStatic bool, height float64) {
	return true, 0 // XXX WHAT IS IT
}
