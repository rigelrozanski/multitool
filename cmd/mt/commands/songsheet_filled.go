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

	lyricFontPt float64

	printTitleFlag   bool
	spacingRatioFlag float64
	numColumnsFlag   uint16
)

func init() {
	SongsheetFilledCmd.PersistentFlags().BoolVar(
		&mirrorStringsOrderFlag, "mirror", false, "mirror string positions")
	SongsheetFilledCmd.PersistentFlags().Uint16Var(
		&numColumnsFlag, "columns", 2, "number of columns to print song into")

	SongsheetFilledCmd.PersistentFlags().Float64Var(
		&spacingRatioFlag, "spacing-ratio", 1.0, "ratio of the spacing to the lyric-lines")
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
	if !printTitleFlag {
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
	lyricFontPt, err = determineLyricFontPt(lines, bndsCols[0])
	if err != nil {
		return err
	}

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

	// print the songsheet elements
	//  - use a temp pdf to test whether the borders are exceeded within
	//    the current column, if so move to the next column
	for _, el := range parsedElems {
		pdfTemp := &gofpdf.Fpdf{}
		*pdfTemp = *pdf
		bndNew := el.printPDF(pdfTemp, bndsCols[bndsColsIndex])
		if bndNew.Height() < 0 || bndNew.Width() < 0 {
			bndsColsIndex++
			if bndsColsIndex >= len(bndsCols) {
				return errors.New("song doesn't fit on one sheet (functionality not built yet for multiple sheets)") // TODO
			}
			bndsCols[bndsColsIndex] = el.printPDF(pdf, bndsCols[bndsColsIndex])
		} else {
			bndsCols[bndsColsIndex] = bndNew
			*pdf = *pdfTemp
		}
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
		for _, fld := range flds {
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
	parseText(lines []string) (reducedLines []string, elem tssElement, err error)
}

// ---------------------

type chordChart struct {
	chords      []Chord
	labelFontPt float64
}

type Chord struct {
	name      string // must be 1 or 2 characters
	positions []int  // from thick to thin guitar strings
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

	cOut := chordChart{labelFontPt: 12}
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
			num, err := strconv.Atoi(word)
			if err != nil {
				return lines, elem,
					fmt.Errorf("bad number conversion for %v at %v,%v",
						word, j, i)
			}
			newChord.positions = append(newChord.positions, num)
		}
		cOut.chords = append(cOut.chords, newChord)
	}

	// chop off the first 9 lines
	return lines[9:], cOut, nil
}

func (c chordChart) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reduced bounds) {

	// the top zone of the pillar that shows the guitar string thicknesses
	thicknessIndicatorMargin := padding / 2

	spacing := padding / 2
	cactusZoneWidth := 0.0
	cactusPrickleSpacing := padding
	cactusZoneWidth = padding // one for the cactus

	noLines := len(thicknesses)

	// print thicknesses
	var xStart, xEnd, yStart, yEnd float64
	for i := 0; i < noLines; i++ {
		pdf.SetLineWidth(thicknesses[i])
		yStart = bnd.top + cactusZoneWidth + (float64(i) * spacing)
		yEnd = yStart
		xStart = bnd.left
		xEnd = xStart + thicknessIndicatorMargin
		pdf.Line(xStart, yStart, xEnd, yEnd)
	}

	// print seperator
	pdf.SetLineWidth(thinestLW)
	yStart = bnd.top + cactusZoneWidth
	yEnd = yStart + float64(noLines-1)*spacing
	xStart = bnd.left + thicknessIndicatorMargin
	xEnd = xStart
	pdf.Line(xStart, yStart, xEnd, yEnd)

	// print pillar lines
	for i := 0; i < noLines; i++ {
		pdf.SetLineWidth(thinestLW)
		yStart = bnd.top + cactusZoneWidth + (float64(i) * spacing)
		yEnd = yStart
		xStart = bnd.left + thicknessIndicatorMargin
		xEnd = bnd.right - padding
		pdf.Line(xStart, yStart, xEnd, yEnd)
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
		xStartLabel := xStart - fontWidth/2
		if len(chd.name) == 2 {
			xStartLabel = xStart - fontWidth
		}
		yStartLabel := yBottomEnd + fontHeight + labelPadding
		pdf.Text(xStartLabel, yStartLabel, chd.name)

		// print positions
		xStartPositions := xStart - fontWidth/2
		for i := 0; i < noLines; i++ {
			yStartPositions := bnd.top + cactusZoneWidth +
				(float64(i) * spacing) + fontHeight/2
			pdf.Text(xStartPositions, yStartPositions, strconv.Itoa(chd.positions[i]))
		}

		chordIndex++
	}

	elemThickness := 2*cactusZoneWidth + padding +
		(float64(noLines-1) * spacing) + fontHeight + labelPadding
	return bounds{bnd.top + elemThickness, bnd.left, bnd.bottom, bnd.right}
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

func (s singleSpacing) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reduced bounds) {
	lineHeight := GetFontHeight(lyricFontPt) * spacingRatioFlag
	return bounds{bnd.top + lineHeight, bnd.left, bnd.bottom, bnd.right}
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
		return lines, elem,
			fmt.Errorf("improper number of input lines,"+
				" want 1 have %v", len(lines))
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
	// XXX

	return bounds{bnd.bottom, bnd.left, bnd.bottom, bnd.left} // XXX
}

// ---------------------

func GetFontPt(heightInches float64) float64 {
	return heightInches * 72.281142957923
}

func GetFontHeight(fontPt float64) (heightInches float64) {
	return fontPt / 72.281142957923
}

func GetCourierFontWidthFromHeight(height float64) float64 {
	return 0.43 * height
}

func GetCourierFontHeightFromWidth(width float64) float64 {
	return width / 0.43
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

func determineLyricFontPt(
	lines []string, bnd bounds) (fontPt float64, err error) {

	humpsChars := 0
	for i := 0; i < len(lines)-1; i++ {
		if strings.HasPrefix(lines[i], "_   _   _") &&
			strings.HasPrefix(lines[i+1], " \\_/ \\_/ \\_/") {

			humpsChars = len(strings.TrimSpace(lines[1]))
			secondLineLen := len(strings.TrimSpace(lines[2])) + 1 // +1 for the leading space just trimmed
			if humpsChars < secondLineLen {
				humpsChars = secondLineLen
			}
			break
		}
	}
	if humpsChars == 0 {
		return 0, errors.New("could not find a sine curve to determine the lyric font pt")
	}

	xStart := bnd.left
	xEnd := bnd.right - padding
	width := xEnd - xStart
	fontWidth := width / float64(humpsChars)
	fontHeight := GetCourierFontHeightFromWidth(fontWidth)
	return GetFontPt(fontHeight), nil
}

func (s singleAnnotatedSine) parseText(lines []string) (reduced []string, elem tssElement, err error) {

	// the annotated sine must come in 4 lines
	//    ex.   desciption
	// 1) F              bolded central annotations
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
	amplitude := GetFontHeight(lyricFontPt) / 2
	xStart := bnd.left
	xEnd := bnd.right - padding
	width := xEnd - xStart
	frequency := math.Pi * 2 * s.humps / width
	yStart := bnd.top + amplitude // because the equation may be negative sending to bnd.top
	lastPointX := xStart
	lastPointY := yStart
	for eqX := float64(0); true; eqX += resolution {
		if xStart+eqX > xEnd {
			break
		}
		eqY := amplitude * math.Cos(frequency*eqX)
		if eqX > 0 {

			// -eqY because starts from topleft corner
			pdf.Line(lastPointX, lastPointY, xStart+eqX, yStart-eqY)
		}
		lastPointX = xStart + eqX
		lastPointY = yStart + eqY
	}

	// print the bold central
	fontH := amplitude / 2
	fontPt := GetFontPt(fontH)
	pdf.SetFont("courier", "B", fontPt)
	for _, bs := range s.boldedCentral {
		X := xStart + bs.position
		Y := yStart + fontH/2 // so the text is centered along the sine axis
		pdf.Text(X, Y, string(bs.ch))
	}

	// print the characters along the sine curve
	chhbs := amplitude / 4
	for _, as := range s.alongSine {
		if as.ch == ' ' {
			continue
		}

		// determine hump position
		eqY := amplitude * math.Cos(frequency*as.position)

		// character height which extends beyond the sine curve

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
			w := GetCourierFontWidthFromHeight(h) // font width

			// we want the character to be centered about the sine curve
			pdf.SetFont("courier", "", fontPt)
			tipX := xStart + as.position
			tipY := yStart + eqY
			pdf.Text(tipX-(w/2), tipY+(h/2), string(as.ch))
		}
	}

	usedHeight := 2 * (amplitude + chhbs)
	return bounds{bnd.top + usedHeight, bnd.left, bnd.bottom, bnd.right}
}
