package commands

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
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
- move on-sine annotations to either the top or bottom if they
  directly intersect with a central-axis annotation
- title font size/multiple lines?
- create an amplitude decay factor (flag) allow for decays
  to happen in the middle of sine
   - also allow for pauses (no sine at all)
- eliminate cactus prickles where no label exists
- use of sine instead of cos with different text hump pattern:
     _
    / \_
- allow for comments within a songsheet maybe same style as golang with `//`
- make new file format (and search using qu OR in the current directory for files
  with this new type)
- break this program out to a new repo
*/

var (
	SongsheetFilledCmd = &cobra.Command{
		Use:   "songsheetfilled [qu-id]",
		Short: "print filled songsheet from qu id",
		Args:  cobra.ExactArgs(1),
		RunE:  songsheetFilledCmd,
	}

	SongsheetPlaybackCmd = &cobra.Command{
		Use:   "songsheet-filled-playback [qu-id] [cursor-x] [cursor-y]",
		Short: "move the vim cursor to account for time passed after 1/4 of a sine duration",
		Args:  cobra.ExactArgs(2),
		RunE:  songsheetPlaybackCmd,
	}

	SongsheetPlaybackTimeCmd = &cobra.Command{
		Use:   "songsheet-filled-playback-time [qu-id] [cursor-x] [cursor-y]",
		Short: "return the playback time (mm:ss:[cs][cs]) for the current position",
		Args:  cobra.ExactArgs(2),
		RunE:  songsheetPlaybackTimeCmd,
	}

	lyricFontPt  float64
	longestHumps float64

	printTitleFlag         bool
	spacingRatioFlag       float64
	sineAmplitudeRatioFlag float64
	numColumnsFlag         uint16

	subsupSizeMul = 0.65 // size of sub and superscript relative to thier root's size
)

func init() {
	quac.Initialize(os.ExpandEnv("$HOME/.thranch_config"))

	SongsheetFilledCmd.PersistentFlags().BoolVar(
		&mirrorStringsOrderFlag, "mirror", false,
		"mirror string positions")
	SongsheetFilledCmd.PersistentFlags().Uint16Var(
		&numColumnsFlag, "columns", 2,
		"number of columns to print song into")
	SongsheetFilledCmd.PersistentFlags().Float64Var(
		&spacingRatioFlag, "spacing-ratio", 1.5,
		"ratio of the spacing to the lyric-lines")
	SongsheetFilledCmd.PersistentFlags().Float64Var(
		&sineAmplitudeRatioFlag, "amp-ratio", 0.8,
		"ratio of amplitude of the sine curve to the lyric text")
	RootCmd.AddCommand(SongsheetFilledCmd)
	RootCmd.AddCommand(SongsheetPlaybackCmd)
	RootCmd.AddCommand(SongsheetPlaybackTimeCmd)
}

func songsheetPlaybackCmd(cmd *cobra.Command, args []string) error {
	output, err := songsheetPlayback(cmd, args, false, true)
	fmt.Printf(output)
	return err
}

func songsheetPlaybackTimeCmd(cmd *cobra.Command, args []string) error {
	output, err := songsheetPlayback(cmd, args, true, false)
	fmt.Printf(output)
	return err

}

func songsheetPlayback(cmd *cobra.Command, args []string, getTime, getMovement bool) (
	output string, err error) {

	if getTime && getMovement || (!getTime && !getMovement) {
		panic("bad use of command one and only one must be true")
	}

	initTime := time.Now()

	// get the relevant file
	quid, err := strconv.Atoi(args[0])
	if err != nil {
		return err.Error(), err
	}
	content, found := quac.GetContentByID(uint32(quid))
	if !found {
		err = fmt.Errorf("could not find anything under id: %v", quid)
		return err.Error(), err
	}
	lines := strings.Split(string(content), "\n")

	curX, err := strconv.Atoi(args[1])
	if err != nil {
		return err.Error(), err
	}

	curY, err := strconv.Atoi(args[2])
	if err != nil {
		return err.Error(), err
	}

	// Determine the adjacent time markers,
	// scan outward from this position to
	// find the previous and next ones.

	// the middle sas is the sas which the cursor is
	// considered to be nearest to... it is from
	// this line that the cursor movements will occur
	middleSasLineNo := int16(-1)

	// moving backwards
	yI := int16(curY)
	var surrSas lineAndSasses
	el := singleAnnotatedSine{} // dummy element to make the call
	for {
		workingLines := lines[yI:]
		_, sasEl, err := el.parseText(workingLines)
		sas := sasEl.(singleAnnotatedSine)
		if err == nil {
			ls := lineAndSas{yI, sas}
			surrSas = append([]lineAndSas{ls}, surrSas...)
			if middleSasLineNo == -1 {
				middleSasLineNo = yI + int16(lineNoOffsetToMiddleHump)
			}
			if sas.hasPlaybackTime {
				break
			}
		}
		yI--
		if yI < 0 {
			break
		}
	}
	middleSasIndex := len(surrSas) - 1

	// moving forwards
	yI = int16(curY)
	for {
		workingLines := lines[yI:]
		_, sasEl, err := el.parseText(workingLines)
		sas := sasEl.(singleAnnotatedSine)
		if err == nil {
			ls := lineAndSas{yI, sas}
			surrSas = append(surrSas, ls)
			if sas.hasPlaybackTime {
				break
			}
		}
		yI++
		if int(yI) >= len(lines) {
			break
		}
	}

	// check that the first and last elements have playback times
	if len(surrSas) <= 1 {
		err = errors.New("not enough sas's to determine the playback times")
		return err.Error(), err
	}
	if !surrSas[0].sas.hasPlaybackTime {
		err = errors.New("first sas doesn't have playback time")
		return err.Error(), err
	}
	if !surrSas[len(surrSas)-1].sas.hasPlaybackTime {
		err = errors.New("last sas doesn't have playback time")
		return err.Error(), err
	}
	ptFirst, ptLast := surrSas[0].sas.pt, surrSas[len(surrSas)-1].sas.pt

	// determine the total number of sine humps we're dealing with
	totalHumps := 0.0
	for _, s := range surrSas {
		totalHumps += s.sas.humps
	}

	// determine the duration of time passing per hump
	totalDur := float64(ptLast.t.Sub(ptFirst.t))
	durPerOneForthHump := time.Duration((totalDur / totalHumps) / float64(charsToaHump))

	endCalcTime := time.Now()
	diffTime := endCalcTime.Sub(initTime)

	// potentially have to move more than one
	// character if this calculation has taken to long
	charMovements := 1
	var finalSleep time.Duration
	if diffTime <= durPerOneForthHump {
		finalSleep = durPerOneForthHump - diffTime
	} else {
		charMovements += int(diffTime / durPerOneForthHump)
		finalSleep = diffTime % durPerOneForthHump
	}
	time.Sleep(finalSleep)

	finalCurX, finalCurY, endReached := surrSas.getNextPosition(
		curX, middleSasIndex, charMovements)

	// set the position of the cursor
	// -1 because so that first position is 1
	if endReached {
		output = fmt.Sprintf("<esc>")
		return output, nil
	}

	output = fmt.Sprintf("%vgg%vl", finalCurY, (finalCurX - 1))
	return output, nil
}

var (
	// line offset from the top of a text-sine (with chords and
	// everything) to the middle of the text-based-sine-curve
	lineNoOffsetToMiddleHump = 2

	charsToaHump = 4.0 // 4 character positions to a hump in a text-based sine wave
)

type lineAndSas struct {
	lineNo int16
	sas    singleAnnotatedSine
}

type lineAndSasses []lineAndSas

// gets the next position of the cursor which is moving hump-movements
// along the text-sine-curve. If there are not enough positions within
// the current hump, then recurively call for the next hump
func (ls lineAndSasses) getNextPosition(startCurX, startHumpIndex, charMovement int) (
	endCurX, endCurY int, endReached bool) {

	if startHumpIndex >= len(ls) {
		return 0, 0, true
	}

	// get the current line hump
	clh := ls[startHumpIndex].sas.totalHumps()

	if startCurX+charMovement <= int(clh*charsToaHump) {
		endCurX = startCurX + charMovement
		endCurY = int(ls[startHumpIndex].lineNo) + lineNoOffsetToMiddleHump
		return endCurX, endCurY, false
	}

	reducedCM := (startCurX + charMovement - int(clh*charsToaHump))
	return ls.getNextPosition(0, startHumpIndex+1, reducedCM)
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

		// add the second and third character to the name (if it exists)
		if j+1 < len(chordNames) && chordNames[j+1] != ' ' {
			newChord.name += string(chordNames[j+1])
			if j+2 < len(chordNames) && chordNames[j+2] != ' ' {
				newChord.name += string(chordNames[j+2])
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

// test to see whether or not the second and third inputs are
// superscript and/subscript to the first input if it is a chord
func determineChordsSubscriptSuperscript(ch1, ch2, ch3 rune) (subscript, superscript rune) {
	if !(unicode.IsLetter(ch1) && unicode.IsUpper(ch1)) {
		return ' ', ' '
	}
	subscript, superscript = ' ', ' '
	if unicode.IsNumber(ch2) || (unicode.IsLetter(ch2) && unicode.IsLower(ch2)) {
		subscript = ch2
	}
	if unicode.IsNumber(ch3) || (unicode.IsLetter(ch3) && unicode.IsLower(ch3)) {
		superscript = ch3
	}
	return subscript, superscript
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

		ch1, ch2, ch3 := ' ', ' ', ' '
		switch len(chd.name) {
		case 3:
			ch3 = rune(chd.name[2])
			fallthrough
		case 2:
			ch2 = rune(chd.name[1])
			fallthrough
		case 1:
			ch1 = rune(chd.name[0])
		}

		subscriptCh, superscriptCh := determineChordsSubscriptSuperscript(
			ch1, ch2, ch3)

		xLabel := x - fontWidth/2
		yLabel := yBottomEnd + fontHeight + labelPadding
		pdf.SetFont("courier", "", c.labelFontPt)
		pdf.Text(xLabel, yLabel, string(ch1))

		if subscriptCh != ' ' {
			pdf.SetFont("courier", "", c.labelFontPt*subsupSizeMul)
			pdf.Text(xLabel+fontWidth, yLabel, string(subscriptCh))
		}
		if superscriptCh != ' ' {
			panic("chords labels cannot have superscript")
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

	// whether to display:
	//  '(' = the number with tight brackets
	//  '/' = to add a modfier for a slide up
	//  '\' = to add a modifier for a slide down
	extra rune
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

const (
	mod1         = '.'
	mod2         = '-'
	mod3         = '~'
	extraBrac    = '('
	extraSldUp   = '\\'
	extraSldDown = '/'
)

func runeIsMod(r rune) bool {
	return r == mod1 || r == mod2 || r == mod3
}

func runeIsExtra(r rune) bool {
	return r == extraBrac || r == extraSldUp || r == extraSldDown
}

// contains at least one modifier,
// and only modifiers, extras, or spaces
func stringOnlyContainsMelodyModifiersAndExtras(s string) bool {
	found := false
	for _, b := range s {
		r := rune(b)
		if !(runeIsExtra(r) || runeIsMod(r) || unicode.IsSpace(r)) {
			return false
		}
		if runeIsExtra(r) || runeIsMod(r) {
			found = true
		}
	}
	return found
}

func (s singleLineMelody) parseText(lines []string) (reduced []string, elem tssElement, err error) {
	if len(lines) < 2 {
		return lines, elem,
			fmt.Errorf("improper number of input lines,"+
				" want 1 have %v", len(lines))
	}

	// determine which lines should be used for the melody modifiers
	melodyNums, upper, lower := "", "", ""
	switch {
	// numbers then modifiers/extras
	case stringOnlyContainsNumbersAndSpaces(lines[0]) &&
		stringOnlyContainsMelodyModifiersAndExtras(lines[1]):
		melodyNums, lower = lines[0], lines[1]

	// modifiers/extras, then numbers, then modifiers/extras
	case len(lines) >= 3 &&
		stringOnlyContainsMelodyModifiersAndExtras(lines[0]) &&
		stringOnlyContainsNumbersAndSpaces(lines[1]) &&
		stringOnlyContainsMelodyModifiersAndExtras(lines[2]):
		upper, melodyNums, lower = lines[0], lines[1], lines[2]

	// modifiers/extras then numbers then either not modfiers or no third line
	case stringOnlyContainsMelodyModifiersAndExtras(lines[0]) &&
		stringOnlyContainsNumbersAndSpaces(lines[1]) &&
		((len(lines) >= 3 && !stringOnlyContainsMelodyModifiersAndExtras(lines[2])) ||
			len(lines) == 2):
		upper, melodyNums = lines[0], lines[1]
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

		m := melody{blank: false, num: r, modifier: ' ', extra: ' '}
		chAbove, chBelow := ' ', ' '
		if len(upper) > i && !unicode.IsSpace(rune(upper[i])) {
			chAbove = rune(upper[i])
		}
		if len(lower) > i && !unicode.IsSpace(rune(lower[i])) {
			chBelow = rune(lower[i])
		}
		if runeIsExtra(chAbove) {
			m.extra = chAbove
		}
		if runeIsExtra(chBelow) {
			m.extra = chBelow
		}
		if runeIsMod(chAbove) {
			m.modifierIsAboveNum = true
			m.modifier = chAbove
		} else if runeIsMod(chBelow) {
			m.modifierIsAboveNum = false
			m.modifier = chBelow
		}

		// ensure that the melody modifier has a valid rune
		if unicode.IsSpace(m.modifier) {
			return lines, elem, fmt.Errorf("no melody modifier for the melody")
		}
		if !runeIsMod(m.modifier) {
			return lines, elem, fmt.Errorf(
				"bad modifier not '%s', '%s', or '%s' (have %v)",
				mod1, mod2, mod3, m.modifier)
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
		case mod1:
			xMod := xNum + melodyFontW/2
			yMod := yNum + melodyHPadding*1.5
			if melody.modifierIsAboveNum {
				yMod = yNum - melodyFontH - melodyHPadding/1.5
			}
			pdf.Circle(xMod, yMod, melodyHPadding/1.5, "F")
		case mod2:
			xModStart := xNum
			xModEnd := xNum + melodyFontW
			yMod := yNum + melodyHPadding
			if melody.modifierIsAboveNum {
				yMod = yNum - melodyFontH - melodyHPadding
			}
			pdf.SetLineWidth(thinishLW)
			pdf.Line(xModStart, yMod, xModEnd, yMod)
		case mod3:
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

		// print extra decorations

		// offset factors per number (specific to font)
		XStart0, YStart0, XEnd0, YEnd0 := 0.55, 0.10, 0.45, 0.60
		XStart1, YStart1, XEnd1, YEnd1 := 0.75, 0.14, 0.30, 0.75
		XStart2, YStart2, XEnd2, YEnd2 := 0.75, 0.14, 0.30, 0.80
		XStart3, YStart3, XEnd3, YEnd3 := 0.65, 0.14, 0.30, 0.75
		XStart4, YStart4, XEnd4, YEnd4 := 0.75, 0.14, 0.60, 0.53
		XStart5, YStart5, XEnd5, YEnd5 := 0.65, 0.14, 0.30, 0.60
		XStart6, YStart6, XEnd6, YEnd6 := 0.65, 0.14, 0.50, 0.65
		XStart7, YStart7, XEnd7, YEnd7 := 0.50, 0.10, 0.30, 0.65
		XStart8, YStart8, XEnd8, YEnd8 := 0.65, 0.14, 0.30, 0.65
		XStart9, YStart9, XEnd9, YEnd9 := 0.55, 0.25, 0.30, 0.65

		switch melody.extra {
		case extraBrac:
			xBrac1 := xNum - fontW*0.5
			xBrac2 := xNum + fontW*0.5
			yBrac := yNum - melodyHPadding/2 // shift for looks
			pdf.Text(xBrac1, yBrac, "(")
			pdf.Text(xBrac2, yBrac, ")")

		case extraSldUp:

			// get the next melody number
			nextMelodyNum := ' '
			j := i + 1
			for ; j < len(s.melodies); j++ {
				nmn := s.melodies[j].num
				if unicode.IsNumber(nmn) {
					nextMelodyNum = nmn
					break
				}
			}

			draw := true

			// determine the starting location
			xSldStart, ySldStart := 0.0, 0.0
			switch melody.num {
			case '0':
				xSldStart = xNum + fontW*XStart0
				ySldStart = yNum - melodyHPadding*YStart0
			case '1':
				xSldStart = xNum + fontW*XStart1
				ySldStart = yNum - melodyHPadding*YStart1
			case '2':
				xSldStart = xNum + fontW*XStart2
				ySldStart = yNum - melodyHPadding*YStart2
			case '3':
				xSldStart = xNum + fontW*XStart3
				ySldStart = yNum - melodyHPadding*YStart3
			case '4':
				xSldStart = xNum + fontW*XStart4
				ySldStart = yNum - melodyHPadding*YStart4
			case '5':
				xSldStart = xNum + fontW*XStart5
				ySldStart = yNum - melodyHPadding*YStart5
			case '6':
				xSldStart = xNum + fontW*XStart6
				ySldStart = yNum - melodyHPadding*YStart6
			case '7':
				xSldStart = xNum + fontW*XStart7
				ySldStart = yNum - melodyHPadding*YStart7
			case '8':
				xSldStart = xNum + fontW*XStart8
				ySldStart = yNum - melodyHPadding*YStart8
			case '9':
				xSldStart = xNum + fontW*XStart9
				ySldStart = yNum - melodyHPadding*YStart9
			default:
				draw = false
			}

			// determine the ending location
			xNumNext := xLyricStart + float64(j)*fontW + melodyWPadding
			xSldEnd, ySldEnd := 0.0, 0.0
			switch nextMelodyNum {
			case '0':
				xSldEnd = xNumNext + fontW*XEnd0
				ySldEnd = yNum - fontH + melodyHPadding*YEnd0
			case '1':
				xSldEnd = xNumNext + fontW*XEnd1
				ySldEnd = yNum - fontH + melodyHPadding*YEnd1
			case '2':
				xSldEnd = xNumNext + fontW*XEnd2
				ySldEnd = yNum - fontH + melodyHPadding*YEnd2
			case '3':
				xSldEnd = xNumNext + fontW*XEnd3
				ySldEnd = yNum - fontH + melodyHPadding*YEnd3
			case '4':
				xSldEnd = xNumNext + fontW*XEnd4
				ySldEnd = yNum - fontH + melodyHPadding*YEnd4
			case '5':
				xSldEnd = xNumNext + fontW*XEnd5
				ySldEnd = yNum - fontH + melodyHPadding*YEnd5
			case '6':
				xSldEnd = xNumNext + fontW*XEnd6
				ySldEnd = yNum - fontH + melodyHPadding*YEnd6
			case '7':
				xSldEnd = xNumNext + fontW*XEnd7
				ySldEnd = yNum - fontH + melodyHPadding*YEnd7
			case '8':
				xSldEnd = xNumNext + fontW*XEnd8
				ySldEnd = yNum - fontH + melodyHPadding*YEnd8
			case '9':
				xSldEnd = xNumNext + fontW*XEnd9
				ySldEnd = yNum - fontH + melodyHPadding*YEnd9
			default:
				draw = false
			}

			if draw {
				pdf.SetLineCapStyle("round")
				defer pdf.SetLineCapStyle("")
				pdf.SetLineWidth(thinishtLW)
				pdf.Line(xSldStart, ySldStart, xSldEnd, ySldEnd)
			}

		}
	}
	// the greatest use of height is from the midpoint of the mod3 modifier
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
	hasPlaybackTime bool
	pt              playbackTime
	humps           float64
	trailingHumps   float64 // the sine curve reduces its amplitude to zero during these
	alongAxis       []sineAnnotation
	alongSine       []sineAnnotation
}

// Dummiii TODO
func (sas singleAnnotatedSine) totalHumps() float64 {
	return sas.humps + sas.trailingHumps
}

type sineAnnotation struct {
	position    float64 // in humps
	bolded      bool    // whether the whole unit is bolded
	ch          rune    // main character
	subscript   rune    // following subscript character
	superscript rune    // following superscript character
}

// NewsineAnnotation creates a new sineAnnotation object
func NewSineAnnotation(position float64, bolded bool,
	ch, subscript, superscript rune) sineAnnotation {
	return sineAnnotation{
		position:    position,
		bolded:      bolded,
		ch:          ch,
		subscript:   subscript,
		superscript: superscript,
	}
}

var _ tssElement = singleAnnotatedSine{}

func determineLyricFontPt(
	lines []string, bnd bounds) (maxHumps, fontPt float64, err error) {

	// find the longest set of humps among them all
	humpsChars := 0
	for i := 0; i < len(lines)-1; i++ {
		if strings.HasPrefix(lines[i], "_") &&
			strings.HasPrefix(lines[i+1], " \\_/") {

			humpsCharsNew := len(strings.TrimSpace(lines[i]))

			// +1 for the leading space just trimmed
			secondLineLen := len(strings.TrimSpace(lines[i+1])) + 1

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
	return float64(humpsChars) / charsToaHump, GetFontPt(fontHeight), nil
}

// TODO better name for this function
func IsTopLineSine(lines []string) (sas singleAnnotatedSine, yesitis bool, err error) {

	// the annotated sine must come in 4 OR 5 Lines
	//    ex.   desciption
	// 1) F              along axis annotations
	// 2) _   _   _   _  text representation of the sine humps (top)
	// 3)  \_/ \_/ \_/   text representation of the sine humps (bottom)
	// 4)   ^   ^ 1   v  annotations along the sine curve
	// 5)     00:03.14   (optional) playback time position

	if len(lines) < 4 {
		return sas, false, fmt.Errorf("improper number of input lines,"+
			"want 4 have %v", len(lines))
	}

	// ensure that the second and third lines start with at least 1 sine hump
	//_
	// \_/
	if !(strings.HasPrefix(lines[1], "_") && strings.HasPrefix(lines[2], " \\_/")) {
		return sas, false, fmt.Errorf("first lines are not sine humps")
	}

	// get the playback time if it exists
	if len(lines) > 4 {
		pt, ptFound := getPlaybackTimeFromLine(lines[4])
		sas = singleAnnotatedSine{
			hasPlaybackTime: ptFound,
			pt:              pt,
		}
	}

	return sas, true, nil
}

type playbackTime struct {
	mins      uint8
	secs      uint8
	centiSecs uint8  // 1/100th of a second
	str       string // string representation
	t         time.Time
}

// 00:00.00
func getPlaybackTimeFromLine(line string) (pt playbackTime, found bool) {
	tr := strings.TrimSpace(line)
	if len(tr) != 8 {
		return pt, false
	}
	str := tr
	spl1 := strings.SplitN(tr, ":", 2)
	if len(spl1) != 2 {
		return pt, false
	}
	spl2 := strings.SplitN(spl1[1], ".", 2)
	if len(spl1) != 2 {
		return pt, false
	}

	mins, err := strconv.Atoi(spl1[0])
	if err != nil {
		return pt, false
	}
	secs, err := strconv.Atoi(spl2[0])
	if err != nil {
		return pt, false
	}
	centiSecs, err := strconv.Atoi(spl2[0])
	if err != nil {
		return pt, false
	}

	// get the time in the golang time format
	dur := time.Minute * time.Duration(mins)
	dur += time.Second * time.Duration(secs)
	dur += time.Millisecond * 10 * time.Duration(centiSecs)
	t := time.Time{}.Add(dur)

	return playbackTime{
		mins:      uint8(mins),
		secs:      uint8(secs),
		centiSecs: uint8(centiSecs),
		str:       str,
		t:         t,
	}, true
}

func (s singleAnnotatedSine) parseText(lines []string) (reduced []string, elem tssElement, err error) {

	sas, _, err := IsTopLineSine(lines)
	if err != nil {
		return lines, elem, err
	}

	humpsChars := len(strings.TrimSpace(lines[1]))
	secondLineTrimTrail := strings.TrimRight(lines[2], ".")
	secondLineLen := len(strings.TrimSpace(secondLineTrimTrail)) + 1 // +1 for the leading space just trimmed
	if humpsChars < secondLineLen {
		humpsChars = secondLineLen
	}
	humps := float64(humpsChars) / charsToaHump

	trailingHumpsChars := strings.Count(lines[2], ".")
	trailingHumps := float64(trailingHumpsChars) / charsToaHump

	// parse along axis text
	alongAxis := []sineAnnotation{}
	fl := lines[0]
	for pos := 0; pos < len(fl); pos++ {
		ch := rune(fl[pos])
		if ch == ' ' {
			continue
		}
		bolded := false

		if unicode.IsLetter(ch) &&
			unicode.IsUpper(ch) {

			bolded = true
		}

		ch2, ch3 := ' ', ' '
		if pos+1 < len(fl) {
			ch2 = rune(fl[pos+1])
		}
		if pos+2 < len(fl) {
			ch3 = rune(fl[pos+2])
		}

		subscript, superscript := determineChordsSubscriptSuperscript(
			ch, ch2, ch3)

		alongAxis = append(alongAxis,
			NewSineAnnotation(float64(pos)/4, bolded, ch,
				subscript, superscript))

		if subscript != ' ' {
			pos++
		}
		if superscript != ' ' {
			pos++
		}
	}

	// parse along sine text
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
			NewSineAnnotation(float64(pos)/4, bolded, ch, ' ', ' '))
	}

	sas.humps = humps
	sas.trailingHumps = trailingHumps
	sas.alongAxis = alongAxis
	sas.alongSine = alongSine
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
	trailingWidth := 0.0
	if s.humps < longestHumps {
		trailingWidth = width * s.trailingHumps / longestHumps
		width = width * s.humps / longestHumps
	}
	frequency := math.Pi * 2 * s.humps / width
	yStart := bnd.top + usedHeight/2
	lastPointX := xStart
	lastPointY := yStart
	pdf.SetLineWidth(thinestLW)

	// regular sinepart
	eqX := 0.0
	for ; true; eqX += resolution {
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

	// trailing sine part
	maxWidth := width + trailingWidth
	for ; true; eqX += resolution {
		if eqX > maxWidth {
			break
		}

		// trailing amplitude
		ta := amplitude * (maxWidth - eqX) / trailingWidth

		eqY := ta * math.Cos(frequency*eqX)

		if eqX > 0 {
			// -eqY because starts from topleft corner
			pdf.Line(lastPointX, lastPointY, xStart+eqX, yStart-eqY)
		}
		lastPointX = xStart + eqX
		lastPointY = yStart - eqY
	}

	///////////////
	// print the text along axis

	// (max multiplier would be 2 as the text is
	// centered between the positive and neg amplitude)
	fontH := amplitude * 1.7

	fontW := GetCourierFontWidthFromHeight(fontH)
	fontPt := GetFontPt(fontH)
	fontHSubSup := fontH * subsupSizeMul
	fontPtSubSup := GetFontPt(fontHSubSup)

	XsubsupCrunch := fontW * 0.1 // squeeze the sub and super script into the chord a bit

	for _, aa := range s.alongAxis {

		X := xStart + (aa.position/s.humps)*width - fontW/2
		Y := yStart + fontH/2 // so the text is centered along the sine axis
		bolded := ""
		if aa.bolded {
			bolded = "B"
		}
		pdf.SetFont("courier", bolded, fontPt)
		pdf.Text(X, Y, string(aa.ch))

		// print sub or super script if exists
		if aa.subscript != ' ' || aa.superscript != ' ' {
			Xsubsup := X + fontW - XsubsupCrunch
			pdf.SetFont("courier", bolded, fontPtSubSup)
			if aa.subscript != ' ' {
				Ysub := Y - fontH/2 + fontHSubSup
				pdf.Text(Xsubsup, Ysub, string(aa.subscript))
			}
			if aa.superscript != ' ' {
				Ysuper := Y - fontH/2
				pdf.Text(Xsubsup, Ysuper, string(aa.superscript))
			}
		}

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
