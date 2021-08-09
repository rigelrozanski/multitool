package commands

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/jung-kurt/gofpdf"
	"github.com/rigelrozanski/thranch/quac"
	"github.com/spf13/cobra"
)

/*
Design of audio playback in songsheet
 - the sine now needs:
   - playback time (first line in format "00:00.00" however can be
     ANYWHERE in line, for instance to line up with a lyric would often be the easiest)
   - relative position to all other sines
 - qu ability to record companion audio track for an
   idea file... should tag the new id:
      some-tag,uses-audio=1142,some-other-tag
 - command in vim to call qu record using sox and dispatch
 - audio playback at the hump location
   - another filename tag for vim to enable playback with: song-playback
   - ability play slowed down (using speed, or tempo option within sox)
   - This should be done with sequencial golang program calls (one call for
     each cursor movement)
 - move the cursor with the hump locations as the playback occurs
 - ability to hit escape during playback to stop the playback and cursor
 - adjust cursor movement speed which represent the playback based on all the
 checkpoints at the humps and the actual time vs the amount of time the humps
 would suggest. The first and last checkpoints normalize the position of the
 audio, and create the average BPM.  additional checkpoints within the track
 will create movement speed variations of the cursor.
*/

/*
TODO
- flags for colours!
  - about the sine curve colours
- bass line strings annotations
- chord chart to be able to sqeeze a few more chords in beyond
  the standard spacing
- chord chart to be able to go across the top if the squeeze doesn't work
*/

var (
	SongsheetFilledCmd = &cobra.Command{
		Use:   "songsheet-filled [qu-id]",
		Short: "print filled songsheet from qu id",
		Args:  cobra.ExactArgs(1),
		RunE:  songsheetFilledCmd,
	}

	lyricFontPt  float64
	longestHumps float64

	printTitleFlag         bool
	spacingRatioFlag       float64
	sineAmplitudeRatioFlag float64
	numColumnsFlag         uint16

	subscriptFactor = 0.6
)

func init() {
	quac.Initialize(os.ExpandEnv("$HOME/.thranch_config"))

	SongsheetFilledCmd.PersistentFlags().BoolVar(
		&mirrorStringsOrderFlag, "mirror", false, "mirror string positions")
	SongsheetFilledCmd.PersistentFlags().Uint16Var(
		&numColumnsFlag, "columns", 2, "number of columns to print song into")

	SongsheetFilledCmd.PersistentFlags().Float64Var(
		&spacingRatioFlag, "spacing-ratio", 1.5, "ratio of the spacing to the lyric-lines")
	SongsheetFilledCmd.PersistentFlags().Float64Var(
		&sineAmplitudeRatioFlag, "amp-ratio", 0.75, "ratio of amplitude of the sine curve to the lyric text")
	RootCmd.AddCommand(SongsheetFilledCmd)

}

func songsheetFilledCmd(cmd *cobra.Command, args []string) error {

	pdf := gofpdf.New("P", "in", "Letter", "")
	pdf.SetMargins(0, 0, 0)
	pdf.AddPage()

	// each line of text from the input file
	// is attempted to be fit into elements
	// in the order provided within elemKinds
	elemKinds := []tssElement{
		singleSpacing{},
		chordChart{},
		singleAnnotatedSine{},
		singleLineMelody{},
		singleLineLyrics{},
	}

	if numColumnsFlag < 1 {
		return errors.New("numColumnsFlag must be greater than 1")
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
	filename := fmt.Sprintf("songsheet_%v.pdf", hc.title)

	bnd := bounds{padding, padding, 11, 8.5}
	if printTitleFlag {
		hc.title = ""
	}
	bnd = printHeader(pdf, bnd, &hc)

	//seperate out remaining bounds into columns
	bndsColsIndex := 0
	bndsCols := splitBoundsIntoColumns(bnd, numColumnsFlag)
	if len(bndsCols) == 0 {
		panic("no bound columns")
	}

	//determine lyricFontPt
	longestHumps, lyricFontPt, err = determineLyricFontPt(lines, bndsCols[0])
	if err != nil {
		return err
	}

	// get contents of songsheet
	// parse all the elems
	parsedElems := []tssElement{}

OUTER:
	if len(lines) > 0 {
		allErrs := []string{}
		for _, elem := range elemKinds {
			reduced, newElem, err := elem.parseText(lines)
			if err == nil {
				lines = reduced
				parsedElems = append(parsedElems, newElem)
				goto OUTER
			} else {
				allErrs = append(allErrs, err.Error())
			}
		}
		return fmt.Errorf("could not parse song at line %+v\n all errors%+v\n", lines, allErrs)
	}

	// print the songsheet elements
	//  - use a dummy pdf to test whether the borders are exceeded within
	//    the current column, if so move to the next column
	for _, el := range parsedElems {
		dummy := dummyPdf{}
		bndNew := el.printPDF(dummy, bndsCols[bndsColsIndex])
		if bndNew.Height() < padding {
			bndsColsIndex++
			if bndsColsIndex >= len(bndsCols) {
				return errors.New("song doesn't fit on one sheet " +
					"(functionality not built yet for multiple sheets)") // TODO
			}
		}
		bndsCols[bndsColsIndex] = el.printPDF(pdf, bndsCols[bndsColsIndex])
	}

	return pdf.OutputFileAndClose(filename)
}

func splitBoundsIntoColumns(bnd bounds, numCols uint16) (splitBnds []bounds) {
	width := (bnd.right - bnd.left) / float64(numCols)
	for i := uint16(0); i < numCols; i++ {
		b := bounds{
			top:    bnd.top,
			bottom: bnd.bottom,
			left:   bnd.left + float64(i)*width,
			right:  bnd.left + float64(i+1)*width,
		}
		splitBnds = append(splitBnds, b)
	}
	return splitBnds
}

func parseHeader(lines []string) (reduced []string, hc headerContent, err error) {
	if len(lines) < 2 {
		return lines, hc, fmt.Errorf("improper number of "+
			"input lines, want at least 2 have %v", len(lines))
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
		if _, found := keywords[fld]; found {
			fldIsKeyword[i] = true
		}
	}

	for keyword, _ := range keywords {
		keywordFound := false
		keywordText := ""
	FLDSLOOP:
		for i, fld := range flds {
			switch {
			case !keywordFound && keyword == fld:
				keywordFound = true
				continue
			case keywordFound && fldIsKeyword[i]:
				break FLDSLOOP
			case keywordFound && !fldIsKeyword[i] && keywordText == "":
				keywordText = fld
			case keywordFound && !fldIsKeyword[i] && keywordText != "":
				keywordText += " " + fld
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

type Pdf interface {
	SetLineWidth(width float64)
	SetLineCapStyle(styleStr string)
	Line(x1, y1, x2, y2 float64)
	SetFont(familyStr, styleStr string, size float64)
	Text(x, y float64, txtStr string)
	Circle(x, y, r float64, styleStr string)
	Curve(x0, y0, cx, cy, x1, y1 float64, styleStr string)
}

// dummyPdf fulfills the interface Pdf (DNETL)
type dummyPdf struct{}

var _ Pdf = dummyPdf{}

func (d dummyPdf) SetLineWidth(width float64)                            {}
func (d dummyPdf) SetLineCapStyle(styleStr string)                       {}
func (d dummyPdf) Line(x1, y1, x2, y2 float64)                           {}
func (d dummyPdf) SetFont(familyStr, styleStr string, size float64)      {}
func (d dummyPdf) Text(x, y float64, txtStr string)                      {}
func (d dummyPdf) Circle(x, y, r float64, styleStr string)               {}
func (d dummyPdf) Curve(x0, y0, cx, cy, x1, y1 float64, styleStr string) {}

// ---------------------

// whole text songsheet element
type tssElement interface {
	printPDF(Pdf, bounds) (reduced bounds)
	parseText(lines []string) (reducedLines []string, elem tssElement, err error)
}

// ---------------------

type chordChart struct {
	chords          []Chord
	labelFontPt     float64
	positionsFontPt float64
}

type Chord struct {
	name      string   // must be 1 or 2 characters
	positions []string // from thick to thin guitar strings
}

var _ tssElement = chordChart{}

func (c chordChart) parseText(lines []string) (reduced []string, elem tssElement, err error) {
	if len(lines) < 9 {
		return lines, elem,
			fmt.Errorf("improper number of input lines,"+
				" want at least 9 have %v", len(lines))
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

	cOut := chordChart{
		labelFontPt:     12,
		positionsFontPt: 10,
	}
	// get the chords
	chordNames := lines[8]
	for j := 2; j < len(chordNames); j += 3 {

		if chordNames[j] == ' ' {
			// this chord is not labelled, must be the end of the chords
			break
		}

		newChord := Chord{name: string(chordNames[j])}

		// add the second character to the name (if it exists)
		if j+1 < len(chordNames) {
			if chordNames[j+1] != ' ' {
				newChord.name += string(chordNames[j+1])
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
			newChord.positions = append(newChord.positions, word)
		}
		cOut.chords = append(cOut.chords, newChord)
	}

	// chop off the first 9 lines
	return lines[9:], cOut, nil
}

func isChordWithSubscript(ch1, ch2 rune) bool {
	return unicode.IsLetter(ch1) &&
		unicode.IsUpper(ch1) &&
		(unicode.IsNumber(ch2) || (unicode.IsLetter(ch2) && unicode.IsLower(ch2)))
}

func (c chordChart) printPDF(pdf Pdf, bnd bounds) (reduced bounds) {

	usedHeight := 0.0

	// the top zone of the pillar that shows the guitar string thicknesses
	thicknessIndicatorMargin := padding / 2

	spacing := padding / 2
	cactusZoneWidth := 0.0
	cactusPrickleSpacing := padding
	cactusZoneWidth = padding // one for the cactus

	noLines := len(thicknesses)

	// print thicknesses
	var xStart, xEnd, y float64
	for i := 0; i < noLines; i++ {
		pdf.SetLineWidth(thicknesses[i])
		y = bnd.top + cactusZoneWidth + (float64(i) * spacing)
		xStart = bnd.left
		xEnd = xStart + thicknessIndicatorMargin
		pdf.Line(xStart, y, xEnd, y)
	}
	usedHeight += cactusZoneWidth + float64(noLines)*spacing

	// print seperator
	pdf.SetLineWidth(thinestLW)
	yStart := bnd.top + cactusZoneWidth
	yEnd := yStart + float64(noLines-1)*spacing
	xStart = bnd.left + thicknessIndicatorMargin
	xEnd = xStart
	pdf.Line(xStart, yStart, xEnd, yEnd)

	// print pillar lines
	for i := 0; i < noLines; i++ {
		pdf.SetLineWidth(thinestLW)
		y = bnd.top + cactusZoneWidth + (float64(i) * spacing)
		xStart = bnd.left + thicknessIndicatorMargin
		xEnd = bnd.right - padding
		pdf.Line(xStart, y, xEnd, y)
	}

	// print prickles
	xStart = bnd.left + thicknessIndicatorMargin + cactusPrickleSpacing/2
	xEnd = bnd.right - padding
	chordIndex := 0
	pdf.SetFont("courier", "", c.labelFontPt)
	fontHeight := GetFontHeight(c.labelFontPt)
	labelPadding := fontHeight * 0.1
	fontWidth := GetCourierFontWidthFromHeight(fontHeight)
	for x := xStart; x < xEnd; x += cactusPrickleSpacing {
		pdf.SetLineWidth(thinestLW)
		yTopStart := bnd.top
		yTopEnd := yTopStart + cactusZoneWidth/2
		yBottomStart := bnd.top + cactusZoneWidth +
			(float64(noLines-1) * spacing) + cactusZoneWidth/2
		yBottomEnd := yBottomStart + cactusZoneWidth/2

		pdf.Line(x, yTopStart, x, yTopEnd)
		pdf.Line(x, yBottomStart, x, yBottomEnd)

		// print labels
		if chordIndex >= len(c.chords) {
			continue
		}
		chd := c.chords[chordIndex]
		if len(chd.name) == 2 &&
			isChordWithSubscript(rune(chd.name[0]), rune(chd.name[1])) {

			xLabel := x - fontWidth/2
			yLabel := yBottomEnd + fontHeight + labelPadding
			pdf.SetFont("courier", "", c.labelFontPt)
			pdf.Text(xLabel, yLabel, string(chd.name[0]))
			pdf.SetFont("courier", "", c.labelFontPt*subscriptFactor)
			pdf.Text(xLabel+fontWidth, yLabel, string(chd.name[1]))
		} else {
			xLabel := x - fontWidth/2
			if len(chd.name) == 2 {
				xLabel = x - fontWidth
			}
			yLabel := yBottomEnd + fontHeight + labelPadding
			pdf.SetFont("courier", "", c.labelFontPt)
			pdf.Text(xLabel, yLabel, chd.name)
		}

		// print positions
		pdf.SetFont("courier", "", c.positionsFontPt)
		posFontH := GetFontHeight(c.positionsFontPt)
		posFontW := GetCourierFontWidthFromHeight(posFontH)
		//xPositions := x - fontWidth/2 // maybe incorrect, but looks better
		xPositions := x - posFontW/2
		for i := 0; i < noLines; i++ {
			yPositions := bnd.top + cactusZoneWidth +
				(float64(i) * spacing) + posFontH/2

			if chd.positions[i] == "x" {
				ext := posFontW / 2
				y := yPositions - posFontH/2
				pdf.Line(x-ext, y-ext, x+ext, y+ext)
				pdf.Line(x-ext, y+ext, x+ext, y-ext)
				continue
			}

			pdf.Text(xPositions, yPositions, chd.positions[i])
		}

		chordIndex++
	}
	// for the lower prickles and labels
	// (upper prickles already accounted for in previous usedHeight accumulation)
	usedHeight += cactusZoneWidth + fontHeight + labelPadding

	return bounds{bnd.top + usedHeight, bnd.left, bnd.bottom, bnd.right}
}

// ---------------------

type singleSpacing struct{}

var _ tssElement = singleSpacing{}

func (s singleSpacing) parseText(lines []string) (reduced []string, elem tssElement, err error) {
	if len(lines) < 1 {
		return lines, elem,
			fmt.Errorf("improper number of input lines, want 1 have %v", len(lines))
	}
	if len(strings.TrimSpace(lines[0])) != 0 {
		return lines, elem, errors.New("blank line contains content")
	}
	return lines[1:], singleSpacing{}, nil
}

func (s singleSpacing) printPDF(pdf Pdf, bnd bounds) (reduced bounds) {
	lineHeight := GetFontHeight(lyricFontPt) * spacingRatioFlag
	return bounds{bnd.top + lineHeight, bnd.left, bnd.bottom, bnd.right}
}

// ---------------------
type singleLineMelody struct {
	melodies []melody
}

var _ tssElement = singleLineMelody{}

type melody struct {
	blank              bool // no melody here, this is just a placeholder
	num                rune
	modifierIsAboveNum bool // otherwise below
	modifier           rune // either '.', '-', or '~'
}

// contains at least one number,
// and only numbers or spaces
func stringOnlyContainsNumbersAndSpaces(s string) bool {
	numFound := false
	for _, b := range s {
		r := rune(b)
		if !(unicode.IsSpace(r) || unicode.IsNumber(r)) {
			return false
		}
		if unicode.IsNumber(r) {
			numFound = true
		}
	}
	return numFound
}

// contains at least one modifier,
// and only modifiers or spaces
func stringOnlyContainsMelodyModifiers(s string) bool {
	modFound := false
	for _, b := range s {
		r := rune(b)
		if !(r == '.' || r == '-' || r == '~' || unicode.IsSpace(r)) {
			return false
		}
		if r == '.' || r == '-' || r == '~' {
			modFound = true
		}
	}
	return modFound
}

func (s singleLineMelody) parseText(lines []string) (reduced []string, elem tssElement, err error) {
	if len(lines) < 2 {
		return lines, elem,
			fmt.Errorf("improper number of input lines,"+
				" want 1 have %v", len(lines))
	}

	// determine which lines should be used for the melody modifiers
	melodyNums, upperMods, lowerMods := "", "", ""
	switch {
	// numbers then modifiers
	case stringOnlyContainsNumbersAndSpaces(lines[0]) &&
		stringOnlyContainsMelodyModifiers(lines[1]):
		melodyNums, lowerMods = lines[0], lines[1]

	// modifiers, then numbers, then modifiers
	case len(lines) >= 3 &&
		stringOnlyContainsMelodyModifiers(lines[0]) &&
		stringOnlyContainsNumbersAndSpaces(lines[1]) &&
		stringOnlyContainsMelodyModifiers(lines[2]):
		upperMods, melodyNums, lowerMods = lines[0], lines[1], lines[2]

	// modifiers then numbers then either not modfiers or no third line
	case stringOnlyContainsMelodyModifiers(lines[0]) &&
		stringOnlyContainsNumbersAndSpaces(lines[1]) &&
		((len(lines) >= 3 && !stringOnlyContainsMelodyModifiers(lines[2])) ||
			len(lines) == 2):
		upperMods, melodyNums = lines[0], lines[1]
	default:
		return lines, elem, fmt.Errorf("could not determine melody number line and modifier line")
	}

	slm := singleLineMelody{}
	melodiesFound := false
	for i, r := range melodyNums {
		if !(unicode.IsSpace(r) || unicode.IsNumber(r)) {
			return lines, elem, fmt.Errorf(
				"melodies line contains something other"+
					"than numbers and spaces (rune: %v, col: %v)", r, i)
		}
		if unicode.IsSpace(r) {
			slm.melodies = append(slm.melodies, melody{blank: true})
			continue
		}

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

		slm.melodies = append(slm.melodies, m)
		melodiesFound = true
	}

	if !melodiesFound {
		return lines, elem, fmt.Errorf("no melodies found")
	}

	return lines[3:], slm, nil
}

func (s singleLineMelody) printPDF(pdf Pdf, bnd bounds) (reduced bounds) {

	// accumulate all the used height as it's used
	usedHeight := 0.0

	// lyric font info
	fontH := GetFontHeight(lyricFontPt)
	fontW := GetCourierFontWidthFromHeight(fontH)
	xLyricStart := bnd.left - fontW/2 // - because of slight right shift in sine annotations

	// print the melodies
	melodyFontPt := lyricFontPt
	pdf.SetFont("courier", "", melodyFontPt)
	melodyFontH := GetFontHeight(melodyFontPt)
	melodyFontW := GetCourierFontWidthFromHeight(melodyFontH)
	melodyHPadding := melodyFontH * 0.3
	melodyWPadding := 0.0
	yNum := bnd.top + melodyFontH + melodyHPadding*2
	usedHeight += melodyFontH + melodyHPadding*3
	for i, melody := range s.melodies {
		if melody.blank {
			continue
		}

		// print number
		xNum := xLyricStart + float64(i)*fontW + melodyWPadding
		pdf.Text(xNum, yNum, string(melody.num))

		// print modifier
		switch melody.modifier {
		case '.':
			xMod := xNum + melodyFontW/2
			yMod := yNum + melodyHPadding*1.5
			if melody.modifierIsAboveNum {
				yMod = yNum - melodyFontH - melodyHPadding/1.5
			}
			pdf.Circle(xMod, yMod, melodyHPadding/1.5, "F")
		case '-':
			xModStart := xNum
			xModEnd := xNum + melodyFontW
			yMod := yNum + melodyHPadding
			if melody.modifierIsAboveNum {
				yMod = yNum - melodyFontH - melodyHPadding
			}
			pdf.SetLineWidth(thinishLW)
			pdf.Line(xModStart, yMod, xModEnd, yMod)
		case '~':
			xModStart := xNum
			xModMid := xNum + melodyFontW/2
			xModEnd := xNum + melodyFontW
			yMod := yNum + melodyHPadding/2
			yModMid := yMod + melodyHPadding*2
			if melody.modifierIsAboveNum {
				yMod = yNum - melodyFontH - melodyHPadding/2
				yModMid = yMod - melodyHPadding*2
			}
			pdf.SetLineWidth(thinishLW)
			pdf.Curve(xModStart, yMod, xModMid, yModMid, xModEnd, yMod, "")
		default:
			panic(fmt.Errorf("unknown modifier %v", melody.modifier))
		}
	}
	// the greatest use of height is from the midpoint of the ~ modifier
	usedHeight += melodyHPadding/2 + melodyHPadding*3

	return bounds{bnd.top + usedHeight, bnd.left, bnd.bottom, bnd.right}
}

// ---------------------

type singleLineLyrics struct {
	lyrics string
}

var _ tssElement = singleLineLyrics{}

func (s singleLineLyrics) parseText(lines []string) (reduced []string, elem tssElement, err error) {
	if len(lines) < 1 {
		return lines, elem,
			fmt.Errorf("improper number of input lines,"+
				" want 1 have %v", len(lines))
	}

	sll := singleLineLyrics{}
	sll.lyrics = lines[0]
	return lines[1:], sll, nil
}

func (s singleLineLyrics) printPDF(pdf Pdf, bnd bounds) (reduced bounds) {

	// accumulate all the used height as it's used
	usedHeight := 0.0

	// print the lyric
	pdf.SetFont("courier", "", lyricFontPt)
	fontH := GetFontHeight(lyricFontPt)
	fontW := GetCourierFontWidthFromHeight(fontH)
	xLyricStart := bnd.left - fontW/2 // - because of slight right shift in sine annotations
	yLyric := bnd.top + 1.3*fontH     // 3/2 because of tall characters extending beyond height calculation

	// the lyrics could just be printed in one go,
	// however do to the inaccuracies of determining
	// font heights and widths (boohoo) it will look
	// better to just print out each char individually
	for i, ch := range s.lyrics {
		xLyric := xLyricStart + float64(i)*fontW
		pdf.Text(xLyric, yLyric, string(ch))
	}
	usedHeight += 1.3 * fontH
	return bounds{bnd.top + usedHeight, bnd.left, bnd.bottom, bnd.right}
}

// ---------------------

const ( // empirically determined
	ptToHeight    = 100  //72
	widthToHeight = 0.82 //
)

func GetFontPt(heightInches float64) float64 {
	return heightInches * ptToHeight
}

func GetFontHeight(fontPt float64) (heightInches float64) {
	return fontPt / ptToHeight
}

func GetCourierFontWidthFromHeight(height float64) float64 {
	return widthToHeight * height
}

func GetCourierFontHeightFromWidth(width float64) float64 {
	return width / widthToHeight
}

// ---------------------

type singleAnnotatedSine struct {
	humps     float64
	alongAxis []sineAnnotation
	alongSine []sineAnnotation
}

type sineAnnotation struct {
	position  float64 // in humps
	ch        rune    // annotation text
	subscript bool    // should be printed as subscript
	bolded    bool    // should be printed as subscript
}

// NewsineAnnotation creates a new sineAnnotation object
func NewSineAnnotation(position float64, ch rune, subscript, bolded bool) sineAnnotation {
	return sineAnnotation{
		position:  position,
		ch:        ch,
		subscript: subscript,
		bolded:    bolded,
	}
}

var _ tssElement = singleAnnotatedSine{}

func determineLyricFontPt(
	lines []string, bnd bounds) (maxHumps, fontPt float64, err error) {

	// find the longest set of humps among them all
	humpsChars := 0
	for i := 0; i < len(lines)-1; i++ {
		if strings.HasPrefix(lines[i], "_   _   _") &&
			strings.HasPrefix(lines[i+1], " \\_/ \\_/ \\_/") {

			humpsCharsNew := len(strings.TrimSpace(lines[i]))
			secondLineLen := len(strings.TrimSpace(lines[i+1])) + 1 // +1 for the leading space just trimmed
			if humpsCharsNew < secondLineLen {
				humpsCharsNew = secondLineLen
			}
			if humpsCharsNew > humpsChars {
				humpsChars = humpsCharsNew
			}
		}
	}
	if humpsChars == 0 {
		return 0, 0, errors.New("could not find a sine curve to determine the lyric font pt")
	}

	xStart := bnd.left
	xEnd := bnd.right - padding
	width := xEnd - xStart
	fontWidth := width / float64(humpsChars)
	fontHeight := GetCourierFontHeightFromWidth(fontWidth)
	return float64(humpsChars) / 4, GetFontPt(fontHeight), nil
}

func (s singleAnnotatedSine) parseText(lines []string) (reduced []string, elem tssElement, err error) {

	// the annotated sine must come in 4 lines
	//    ex.   desciption
	// 1) F              along axis annotations
	// 2) _   _   _   _  text representation of the sine humps (top)
	// 3)  \_/ \_/ \_/   text representation of the sine humps (bottom)
	// 4)   ^   ^ 1   v  annotations along the sine curve

	if len(lines) < 4 {
		return lines, elem, fmt.Errorf("improper number of input lines,"+
			"want 4 have %v", len(lines))
	}

	// ensure that the first two lines start with at least 3 sine humps
	//_   _   _
	// \_/ \_/ \_/
	if !(strings.HasPrefix(lines[1], "_   _   _") &&
		strings.HasPrefix(lines[2], " \\_/ \\_/ \\_/")) {
		return lines, elem, fmt.Errorf("first lines are not sine humps")
	}

	humpsChars := len(strings.TrimSpace(lines[1]))
	secondLineLen := len(strings.TrimSpace(lines[2])) + 1 // +1 for the leading space just trimmed
	if humpsChars < secondLineLen {
		humpsChars = secondLineLen
	}
	humps := float64(humpsChars) / 4

	alongAxis := []sineAnnotation{}
	prevCh := ' '
	for pos, ch := range lines[0] {
		if ch == ' ' {
			continue
		}
		subscript := false
		bolded := false

		if unicode.IsLetter(ch) &&
			unicode.IsUpper(ch) {

			bolded = true
		}

		if isChordWithSubscript(prevCh, ch) {

			subscript = true
			bolded = true
		}

		alongAxis = append(alongAxis,
			NewSineAnnotation(float64(pos)/4, ch, subscript, bolded))

		prevCh = ch
	}

	alongSine := []sineAnnotation{}
	for pos, ch := range lines[3] {
		if ch == ' ' {
			continue
		}

		bolded := false
		if ch == 'V' {
			ch = 'v'
			bolded = true
		}
		if ch == 'A' {
			ch = '^'
			bolded = true
		}

		alongSine = append(alongSine,
			NewSineAnnotation(float64(pos)/4, ch, false, bolded))
	}

	sas := singleAnnotatedSine{
		humps:     humps,
		alongAxis: alongAxis,
		alongSine: alongSine,
	}

	return lines[4:], sas, nil
}

func (s singleAnnotatedSine) printPDF(pdf Pdf, bnd bounds) (reduced bounds) {

	// Print the sine function
	pdf.SetLineWidth(thinLW)
	resolution := 0.01
	lfh := GetFontHeight(lyricFontPt)
	amplitude := sineAmplitudeRatioFlag * lfh
	chhbs := lfh / 3      // char height beyond sine
	tipHover := chhbs / 2 // char hover when on the sine tip

	usedHeight := 2 * ( // times 2 because both sides of the sine
	amplitude +         // for the sine curve
		chhbs + // for the text extending out of the sine curve
		tipHover) // for the floating text extendion out of the sine tips

	xStart := bnd.left
	xEnd := bnd.right - padding
	width := xEnd - xStart
	if s.humps < longestHumps {
		width = width * s.humps / longestHumps
	}
	frequency := math.Pi * 2 * s.humps / width
	yStart := bnd.top + usedHeight/2
	lastPointX := xStart
	lastPointY := yStart
	pdf.SetLineWidth(thinestLW)
	for eqX := float64(0); true; eqX += resolution {
		if eqX > width {
			break
		}
		eqY := amplitude * math.Cos(frequency*eqX)

		if eqX > 0 {

			// -eqY because starts from topleft corner
			pdf.Line(lastPointX, lastPointY, xStart+eqX, yStart-eqY)
		}
		lastPointX = xStart + eqX
		lastPointY = yStart - eqY
	}

	// print the text along axis

	// (max multiplier would be 2 as the text is
	// centered between the positive and neg amplitude)
	fontH := amplitude * 1.5
	fontW := GetCourierFontWidthFromHeight(fontH)
	fontPt := GetFontPt(fontH)
	fontHSub := fontH * subscriptFactor
	fontPtSub := GetFontPt(fontHSub)
	for _, aa := range s.alongAxis {

		pt := fontPt
		if aa.subscript {
			pt = fontPtSub
		}

		X := xStart + (aa.position/s.humps)*width - fontW/2
		Y := yStart + fontH/2 // so the text is centered along the sine axis
		bolded := ""
		if aa.bolded {
			bolded = "B"
		}
		pdf.SetFont("courier", bolded, pt)
		pdf.Text(X, Y, string(aa.ch))
	}

	// print the characters along the sine curve
	pdf.SetLineCapStyle("square")
	defer pdf.SetLineCapStyle("")
	for _, as := range s.alongSine {
		if as.ch == ' ' {
			continue
		}

		// determine hump position
		eqX := (as.position / s.humps) * width
		eqY := amplitude * math.Cos(frequency*eqX)

		// determine bold params
		bolded := ""
		if as.bolded {
			pdf.SetLineWidth(thickerLW)
			bolded = "B"
		} else {
			pdf.SetLineWidth(thinishLW)
		}

		// character height which extends beyond the sine curve
		switch as.ch {
		case 'v':
			tipX := xStart + eqX
			tipY := yStart - eqY
			dec := (as.position) - math.Trunc(as.position)
			if dec == 0 || dec == 0.5 {
				tipY -= tipHover
			}
			// 45deg angles to the tip
			pdf.Line(tipX-chhbs, tipY-chhbs, tipX, tipY)
			pdf.Line(tipX, tipY, tipX+chhbs, tipY-chhbs)
		case '^':
			tipX := xStart + eqX
			tipY := yStart - eqY
			dec := (as.position) - math.Trunc(as.position)
			if dec == 0 || dec == 0.5 {
				tipY += tipHover
			}
			// 45deg angles to the tip
			pdf.Line(tipX-chhbs, tipY+chhbs, tipX, tipY)
			pdf.Line(tipX, tipY, tipX+chhbs, tipY+chhbs)
		case '|':
			x := xStart + eqX
			pdf.Line(x, yStart-amplitude-chhbs, x, yStart+amplitude+chhbs)

		default:
			h := 2 * chhbs // font height in inches
			fontPt := GetFontPt(h)
			w := GetCourierFontWidthFromHeight(h) // font width

			// we want the character to be centered about the sine curve
			pdf.SetFont("courier", bolded, fontPt)
			tipX := xStart + eqX
			tipY := yStart - eqY
			pdf.Text(tipX-(w/2), tipY+(h/2), string(as.ch))
		}
	}

	return bounds{bnd.top + usedHeight, bnd.left, bnd.bottom, bnd.right}
}
