package commands

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/cobra"
)

// TODO
// - rotation on a group element, to allow for bizzare-rotated columns
// - prickles ON the lines
// - blank element
// - wavey lines
// - sun element

var (
	SongsheetCmd = &cobra.Command{
		Use:   "songsheet [elements]",
		Short: "print songsheet",
		Long: `songsheet elements: 
	pillar:              PILLAR 
	horizontal-pillar:   HPILLAR
	cactus:              CACTUS
	horizontal-cactus:   HCACTUS
	lines:               LINES[spacing,angle-rad]    ex.: LINES; LINES[0.5]; LINES[0.5,0.1]
	vertical-lines:      VLINES[spacing,angle-rad]
	vertical-grid:       VGRID[spacing,size]  
	horizontal-grid:     HGRID[spacing,size]  
	group row:           ROW(elems..)
	group column:        COL(elems..)

pattern example: 
    COL(elem1;COL(elem2;elem3;ROW(elem4;elem5));elem6)`,
		Args: cobra.ExactArgs(1),
		RunE: songsheetCmd,
	}
	padding    = 0.25
	thickerLW  = 0.017
	thinLW     = 0.01
	thinishLW  = 0.0075
	thinishtLW = 0.005
	thinestLW  = 0.001

	headerFlag             bool
	mirrorStringsOrderFlag bool
)

func init() {
	SongsheetCmd.PersistentFlags().BoolVar(
		&headerFlag, "header", true, "include a header element")
	SongsheetCmd.PersistentFlags().BoolVar(
		&mirrorStringsOrderFlag, "mirror", false, "mirror string positions")
	RootCmd.AddCommand(SongsheetCmd)
}

// NOTE for all elements, padding is only added on the right and bottom sides
// so that the padding is never doubled

func songsheetCmd(cmd *cobra.Command, args []string) error {

	pdf := gofpdf.New("P", "in", "Letter", "")
	pdf.SetMargins(0, 0, 0)
	pdf.AddPage()

	elem, err := parseElem(args[0])
	if err != nil {
		return err
	}
	if elem == nil {
		return fmt.Errorf("could not parse %v", args[0])
	}

	bnd := bounds{padding, padding, 11, 8.5}
	if headerFlag {
		bnd = printHeader(pdf, bnd, nil)
	}

	_ = elem.printPDF(pdf, bnd)

	return pdf.OutputFileAndClose(fmt.Sprintf("songsheet_%v.pdf", args[0]))
}

// --------------------------------

func parseElem(text string) (elem ssElement, err error) {

	// NOTE all elements must be registered here
	elemKinds := []ssElement{
		pillar{},
		flowLines{},
		sines{},
		grid{},
		groupElem{parseFn: parseElem},
	}

	for _, elemKind := range elemKinds {
		elem, err := elemKind.parseText(text)
		if err != nil {
			return nil, err
		}
		if elem != nil {
			return elem, nil
		}
	}

	// could not be parsed
	return nil, nil
}

// --------------------------------

type bounds struct {
	top    float64
	left   float64
	bottom float64
	right  float64
}

func (b bounds) Width() float64 {
	return b.right - b.left
}

func (b bounds) Height() float64 {
	return b.bottom - b.top
}

// --------------------------------

type headerContent struct {
	title         string
	titleLine2    string
	date          string
	tuning        string
	capo          string
	bpm           string
	timesigTop    string
	timesigBottom string
}

func printHeader(pdf *gofpdf.Fpdf, bnd bounds, hc *headerContent) (reducedBounds bounds) {
	dateRightOffset := 2.3
	totalHeaderHeight := 1.0
	boxHeight := 0.25
	boxTextMargin := 0.06

	// flip string orientation if called for
	if mirrorStringsOrderFlag {
		thicknessesRev := make([]float64, len(thicknesses))
		j := len(thicknesses) - 1
		for i := 0; i < len(thicknesses); i++ {
			thicknessesRev[j] = thicknesses[i]
			j--
		}
		thicknesses = thicknessesRev
	}

	// print title
	pdf.SetFont("courier", "", 30)
	pdf.Text(bnd.left, bnd.top+1.5*padding, hc.title)

	// print date
	pdf.SetFont("courier", "", 14)
	pdf.Text(bnd.right-dateRightOffset, bnd.top+padding, "DATE:"+hc.date)

	// print box
	pdf.SetLineWidth(thinLW)

	pdf.Rect(bnd.left, bnd.top+totalHeaderHeight-boxHeight,
		bnd.right-bnd.left-padding, boxHeight, "")

	// print box contents
	conts := []string{"TUNING:", "CAPO:", "BPM:", "TIMESIG:", "FEEL:"}
	if hc != nil {
		conts = []string{
			"TUNING: " + hc.tuning,
			"CAPO: " + hc.capo,
			"BPM: " + hc.bpm,
			"TIMESIG: " + hc.timesigTop + "/" + hc.timesigBottom,
		}
	}

	// determine extra space available
	pdf.SetFont("courier", "", 14)
	charH := GetFontHeight(14)
	charW := GetCourierFontWidthFromHeight(charH)
	numChars := 0
	for _, c := range conts {
		numChars += len(c)
	}
	usedWidth := charW * float64(numChars)

	xTextAreaStart := bnd.left + boxTextMargin
	xTextAreaEnd := bnd.right - padding - boxTextMargin
	xTextIncr := 0.0
	freeWidthPerItem := (xTextAreaEnd - xTextAreaStart - usedWidth) / float64(len(conts))

	for _, cont := range conts {
		pdf.Text(xTextAreaStart+xTextIncr, bnd.top+totalHeaderHeight-boxTextMargin, cont)
		xTextIncr += float64(len(cont))*charW + freeWidthPerItem
	}

	return bounds{bnd.top + totalHeaderHeight + padding, bnd.left, bnd.bottom, bnd.right}
}

// songsheet element
type ssElement interface {
	printPDF(*gofpdf.Fpdf, bounds) (reduced bounds)
	getWidth() (isStatic bool, width float64)   // width is only valid if isStatic=true
	getHeight() (isStatic bool, height float64) // height is only valid if isStatic=true
	parseText(text string) (elem ssElement, err error)
}

// ----------------------------------------

type groupElem struct {
	kind    string // "row", "col", or "combo"
	elems   []ssElement
	parseFn func(text string) (ssElement, error)
}

var _ ssElement = groupElem{}

func (ge groupElem) parseText(text string) (ssElement, error) {
	if !strings.HasSuffix(text, ")") {
		return nil, nil
	}
	newGE := groupElem{}
	if strings.HasPrefix(text, "ROW(") {
		newGE.kind = "row"
		text = strings.TrimPrefix(text, "ROW(")
	} else if strings.HasPrefix(text, "COL(") {
		text = strings.TrimPrefix(text, "COL(")
		newGE.kind = "col"
	} else if strings.HasPrefix(text, "COMBO(") {
		text = strings.TrimPrefix(text, "COMBO(")
		newGE.kind = "combo"
	} else {
		return nil, nil
	}

	text = strings.TrimSuffix(text, ")")

	// split by top-level semi-colons
	split := []string{}
	bracketLevel := 0 // increases with (, decreases with )
	lastCutI := 0     // previous cut taken from text
	for i, c := range text {
		switch string(c) {
		case ";":
			if bracketLevel == 0 {
				split = append(split, text[lastCutI:i])
				lastCutI = i + 1
			}
		case "(":
			bracketLevel++
		case ")":
			bracketLevel--
		}
	}
	// last element doesn't need an ';' at the end
	split = append(split, text[lastCutI:])

	for _, elemText := range split {

		//elem, err := parseElem(elemText)
		elem, err := ge.parseFn(elemText)
		if err != nil {
			return nil, err
		}
		if elem == nil {
			return nil, fmt.Errorf("could not parse element: %v", elemText)
		}
		newGE.elems = append(newGE.elems, elem)
	}

	return newGE, nil
}

func (ge groupElem) getWidth() (isStatic bool, width float64) {
	if ge.kind == "combo" {
		return false, 0
	}

	total := 0.0
	for _, elem := range ge.elems {
		isStatic, elemWidth := elem.getWidth()
		if !isStatic {
			return false, 0
		}
		total += elemWidth
	}
	return true, total
}

func (ge groupElem) getHeight() (isStatic bool, height float64) {
	if ge.kind == "combo" {
		return false, 0
	}

	total := 0.0
	for _, elem := range ge.elems {
		isStatic, elemHeight := elem.getHeight()
		if !isStatic {
			return false, 0
		}
		total += elemHeight
	}
	return true, total
}

func (ge groupElem) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reducedBounds bounds) {

	if ge.kind == "combo" {
		for _, elem := range ge.elems {
			_ = elem.printPDF(pdf, bnd)
		}

		// uses entire bounds
		return bounds{bnd.bottom, bnd.left, bnd.bottom, bnd.left}
	}

	nonStaticWidth := bnd.Width()
	nonStaticHeight := bnd.Height()
	nonStaticWidthElems := 0
	nonStaticHeightElems := 0

	for _, elem := range ge.elems {
		switch ge.kind {
		case "row":
			isStaticW, elemWidth := elem.getWidth()
			if isStaticW {
				nonStaticWidth -= elemWidth
			} else {
				nonStaticWidthElems++
			}
		case "col":
			isStaticH, elemHeight := elem.getHeight()
			if isStaticH {
				nonStaticHeight -= elemHeight
			} else {
				nonStaticHeightElems++
			}
		}
	}

	// only used with groupElem row
	nonStaticWidthPerElem := nonStaticWidth / float64(nonStaticWidthElems)

	// only used with groupElem column
	nonStaticHeightPerElem := nonStaticHeight / float64(nonStaticHeightElems)

	for _, elem := range ge.elems {
		newBounds := bounds{bnd.top, bnd.left, bnd.bottom, bnd.right}
		isStaticW, elemWidth := elem.getWidth()
		if isStaticW {
			newBounds.left += elemWidth
		} else {
			switch ge.kind {
			case "row":
				bnd.right = bnd.left + nonStaticWidthPerElem
				newBounds.left += nonStaticWidthPerElem
			case "col":
				// just use all the width
				bnd.right = bnd.left + nonStaticWidth
			}
		}

		isStaticH, elemHeight := elem.getHeight()
		if isStaticH {
			newBounds.top += elemHeight
		} else {
			switch ge.kind {
			case "row":
				// just use all the height
				bnd.bottom = bnd.top + nonStaticHeight
			case "col":
				bnd.bottom = bnd.top + nonStaticHeightPerElem
				newBounds.top += nonStaticHeightPerElem
			}
		}

		_ = elem.printPDF(pdf, bnd)

		// manually adjust the bounds
		bnd = newBounds
	}

	return bnd
}

// ----------------------------------------

type pillar struct {
	isHorizontal bool
	hasPrickles  bool
}

// thicknesses of guitar strings from thick to thin
var thicknesses = []float64{0.0472, 0.0314, 0.0236, 0.0157, 0.0079, 0.0039}

var _ ssElement = pillar{}

func (pil pillar) parseText(text string) (ssElement, error) {
	switch text {
	case "PILLAR":
		return pillar{false, false}, nil
	case "HPILLAR":
		return pillar{true, false}, nil
	case "CACTUS":
		return pillar{false, true}, nil
	case "HCACTUS":
		return pillar{true, true}, nil
	}
	return nil, nil
}

func (pil pillar) elemThickness() float64 {
	cactusZoneWidth := 0.0 // TODO move to struct
	spacing := padding / 2 // TODO move to struct
	if pil.hasPrickles {
		cactusZoneWidth = padding
	}
	noLines := len(thicknesses)
	return 2*cactusZoneWidth + padding + (float64(noLines-1) * spacing)
}

func (pil pillar) getWidth() (isStatic bool, width float64) {
	if pil.isHorizontal {
		return false, 0
	}
	return true, pil.elemThickness()
}

func (pil pillar) getHeight() (isStatic bool, height float64) {
	if !pil.isHorizontal {
		return false, 0
	}
	return true, pil.elemThickness()
}

func (pil pillar) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reducedBounds bounds) {

	// the top zone of the pillar that shows the guitar string thicknesses
	thicknessIndicatorMargin := padding / 2

	spacing := padding / 2
	cactusZoneWidth := 0.0
	cactusPrickleSpacing := padding
	if pil.hasPrickles {
		cactusZoneWidth = padding // one for the cactus
	}

	noLines := len(thicknesses)

	// print thicknesses
	var xStart, xEnd, yStart, yEnd float64
	for i := 0; i < noLines; i++ {
		pdf.SetLineWidth(thicknesses[i])
		if pil.isHorizontal {
			yStart = bnd.top + cactusZoneWidth + (float64(i) * spacing)
			yEnd = yStart
			xStart = bnd.left
			xEnd = xStart + thicknessIndicatorMargin
		} else {
			xStart = bnd.left + cactusZoneWidth + (float64(i) * spacing)
			xEnd = xStart
			yStart = bnd.top
			yEnd = yStart + thicknessIndicatorMargin
		}

		pdf.Line(xStart, yStart, xEnd, yEnd)
	}

	// print seperator
	pdf.SetLineWidth(thinestLW)
	if pil.isHorizontal {
		yStart = bnd.top + cactusZoneWidth
		yEnd = yStart + float64(noLines-1)*spacing
		xStart = bnd.left + thicknessIndicatorMargin
		xEnd = xStart
	} else {
		xStart = bnd.left + cactusZoneWidth
		xEnd = xStart + float64(noLines-1)*spacing
		yStart = bnd.top + thicknessIndicatorMargin
		yEnd = yStart
	}
	pdf.Line(xStart, yStart, xEnd, yEnd)

	// print pillar lines
	for i := 0; i < noLines; i++ {
		pdf.SetLineWidth(thinestLW)
		if pil.isHorizontal {
			yStart = bnd.top + cactusZoneWidth + (float64(i) * spacing)
			yEnd = yStart
			xStart = bnd.left + thicknessIndicatorMargin
			xEnd = bnd.right - padding
		} else {
			xStart = bnd.left + cactusZoneWidth + (float64(i) * spacing)
			xEnd = xStart
			yStart = bnd.top + thicknessIndicatorMargin
			yEnd = bnd.bottom - padding
		}

		pdf.Line(xStart, yStart, xEnd, yEnd)
	}

	// print cactus prickles
	if pil.hasPrickles {
		if pil.isHorizontal {
			xStart := bnd.left + thicknessIndicatorMargin + cactusPrickleSpacing/2
			xEnd := bnd.right - padding
			for x := xStart; x < xEnd; x += cactusPrickleSpacing {
				pdf.SetLineWidth(thinestLW)
				yTopStart := bnd.top
				yTopEnd := yTopStart + cactusZoneWidth/2
				yBottomStart := bnd.top + cactusZoneWidth +
					(float64(noLines-1) * spacing) + cactusZoneWidth/2
				yBottomEnd := yBottomStart + cactusZoneWidth/2

				pdf.Line(x, yTopStart, x, yTopEnd)
				pdf.Line(x, yBottomStart, x, yBottomEnd)
			}
		} else {
			yStart := bnd.top + thicknessIndicatorMargin + cactusPrickleSpacing/2
			yEnd := bnd.bottom - padding
			for y := yStart; y < yEnd; y += cactusPrickleSpacing {
				pdf.SetLineWidth(thinestLW)
				xLeftStart := bnd.left
				xLeftEnd := xLeftStart + cactusZoneWidth/2
				xRightStart := bnd.left + cactusZoneWidth +
					(float64(noLines-1) * spacing) + cactusZoneWidth/2
				xRightEnd := xRightStart + cactusZoneWidth/2

				pdf.Line(xLeftStart, y, xLeftEnd, y)
				pdf.Line(xRightStart, y, xRightEnd, y)
			}
		}
	}

	if pil.isHorizontal {
		return bounds{bnd.top + pil.elemThickness(), bnd.left, bnd.bottom, bnd.right}
	} else {
		return bounds{bnd.top, bnd.left + pil.elemThickness(), bnd.bottom, bnd.right}
	}
}

// ---------------------

type grid struct {
	isHorizontal bool
	spacing      float64
	cells        float64
}

var _ ssElement = grid{}

func (g grid) parseText(text string) (elem ssElement, err error) {

	flOut := grid{}
	if strings.HasPrefix(text, "VGRID") {
		text = strings.TrimPrefix(text, "VGRID")
		flOut = grid{false, padding, 0}
	} else if strings.HasPrefix(text, "HGRID") {
		text = strings.TrimPrefix(text, "HGRID")
		flOut = grid{true, padding, 0}
	} else {
		return nil, nil
	}

	if len(text) == 0 {
		return flOut, nil
	}
	text = strings.TrimPrefix(text, "[")
	text = strings.TrimSuffix(text, "]")
	split := strings.Split(text, ",")

	switch len(split) {
	case 1:
		flOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
	case 2:
		flOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
		flOut.cells, err = strconv.ParseFloat(split[1], 64)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("bad input for %v", text)
	}
	return flOut, nil
}

func (g grid) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reducedBounds bounds) {

	//var usedBnd bounds
	//if g.isHorizontal {
	//usedBnd = bounds{bnd.top, bnd.left, bnd.top + g.size, bnd.right}
	//} else {
	//usedBnd = bounds{bnd.top, bnd.left, bnd.bottom, bnd.left + g.size}
	//}

	//NewFlowLines(true, g.spacing, 0, 0, 0).printPDF(pdf, usedBnd)
	//NewFlowLines(false, g.spacing, math.Pi/2, 0, 0).printPDF(pdf, usedBnd)

	var xBound, yBound float64 = 0, 0
	if g.isHorizontal {
		xBound = bnd.right - padding
		yBound = bnd.top + g.cells*g.spacing
	} else {
		xBound = bnd.left + g.cells*g.spacing
		yBound = bnd.bottom - padding
	}

	// horizontal lines
	for yStart := bnd.top; yStart <= yBound; yStart += g.spacing {
		xStart := bnd.left
		xEnd := xBound
		yEnd := yStart
		pdf.Line(xStart, yStart, xEnd, yEnd)
	}

	// vertical lines
	for xStart := bnd.left; xStart <= xBound; xStart += g.spacing {
		yStart := bnd.top
		yEnd := yBound
		xEnd := xStart
		pdf.Line(xStart, yStart, xEnd, yEnd)
	}

	if g.isHorizontal {
		return bounds{yBound, bnd.left, bnd.bottom, bnd.right}
	} else {
		return bounds{bnd.top, xBound, bnd.bottom, bnd.right}
	}
}

func (g grid) getWidth() (isStatic bool, width float64) {
	if g.isHorizontal {
		return false, 0
	}
	return true, g.cells * g.spacing
}

func (g grid) getHeight() (isStatic bool, height float64) {
	if !g.isHorizontal {
		return false, 0
	}
	return true, g.cells * g.spacing
}

// ---------------------

type flowLines struct {
	isHorizontal bool
	spacing      float64
	angleRad     float64 // in radians
	shift        float64 // startpoint shift
	midPoints    int64   // any midpoints labelled with circles
	ticks        int64   // any midpoints labelled with circles
}

// NewflowLines creates a new flowLines object
func NewFlowLines(isHorizontal bool, spacing, angleRad, shift float64,
	midPoints, ticks int64) flowLines {
	return flowLines{
		isHorizontal: isHorizontal,
		spacing:      spacing,
		angleRad:     angleRad,
		midPoints:    midPoints,
		ticks:        ticks,
	}
}

var _ ssElement = flowLines{}

func (fl flowLines) parseText(text string) (elem ssElement, err error) {
	flOut := flowLines{}
	if strings.HasPrefix(text, "LINES") {
		text = strings.TrimPrefix(text, "LINES")
		flOut = NewFlowLines(true, 2*padding, 0, padding/2, 0, 0)
	} else if strings.HasPrefix(text, "VLINES") {
		text = strings.TrimPrefix(text, "VLINES")
		flOut = NewFlowLines(false, 2*padding, math.Pi/2, padding/2, 0, 0)
	} else {
		return nil, nil
	}

	if len(text) == 0 {
		return flOut, nil
	}
	text = strings.TrimPrefix(text, "[")
	text = strings.TrimSuffix(text, "]")
	split := strings.Split(text, ",")

	switch len(split) {
	case 1:
		flOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
	case 2:
		flOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
		flOut.angleRad, err = strconv.ParseFloat(split[1], 64)
		if err != nil {
			return nil, err
		}
	case 3:
		flOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
		flOut.angleRad, err = strconv.ParseFloat(split[1], 64)
		if err != nil {
			return nil, err
		}
		flOut.shift, err = strconv.ParseFloat(split[2], 64)
		if err != nil {
			return nil, err
		}
	case 4:
		flOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
		flOut.angleRad, err = strconv.ParseFloat(split[1], 64)
		if err != nil {
			return nil, err
		}
		flOut.shift, err = strconv.ParseFloat(split[2], 64)
		if err != nil {
			return nil, err
		}
		flOut.midPoints, err = strconv.ParseInt(split[3], 10, 64)
		if err != nil {
			return nil, err
		}
	case 5:
		flOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
		flOut.angleRad, err = strconv.ParseFloat(split[1], 64)
		if err != nil {
			return nil, err
		}
		flOut.shift, err = strconv.ParseFloat(split[2], 64)
		if err != nil {
			return nil, err
		}
		flOut.midPoints, err = strconv.ParseInt(split[3], 10, 64)
		if err != nil {
			return nil, err
		}
		flOut.ticks, err = strconv.ParseInt(split[4], 10, 64)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("bad input for %v", text)
	}
	return flOut, nil
}

func (fl flowLines) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reducedBounds bounds) {

	pdf.SetLineWidth(thinestLW)
	var midPointXOffSet, midPointYOffSet float64 = 0, 0
	var ticksXOffSet, ticksYOffSet float64 = 0, 0
	midPointCircleRadius := padding / 10
	tickLength := fl.spacing / 10
	tickAngle := fl.angleRad + math.Pi/2 // perpendicular

	if fl.isHorizontal {
		yOverallStart := bnd.top + fl.shift

		yOverallEnd := bnd.bottom - padding
		xStart := bnd.left
		// print the lines and any midpoints
		for yStart := yOverallStart; yStart <= yOverallEnd; yStart += fl.spacing {
			xEnd := bnd.right - padding
			width := xEnd - xStart

			// tan = height/width
			// tan = (yend-ystart)/width
			// yend = width*tan+ystart
			yEnd := math.Tan(fl.angleRad)*width + yStart

			// need 0.001 for float imprecision on final line if on the boundary
			if yEnd > yOverallEnd+0.001 {
				yEnd = yOverallEnd
				// work backwards
				width = (yEnd - yStart) / math.Tan(fl.angleRad)
				xEnd = xStart + width
			}

			// if is first record
			// determine the x and y midpoint coordinate offsets
			if yStart == yOverallStart {
				midPointXOffSet = (xEnd - xStart) / float64(fl.midPoints+1)
				midPointYOffSet = (yEnd - yStart) / float64(fl.midPoints+1)
				ticksXOffSet = (xEnd - xStart) / float64(fl.ticks+1)
				ticksYOffSet = (yEnd - yStart) / float64(fl.ticks+1)
			}

			pdf.Line(xStart, yStart, xEnd, yEnd)
			// print any midpoints
			for i := int64(1); i <= fl.midPoints; i++ {
				mpX := xStart + float64(i)*midPointXOffSet
				mpY := yStart + float64(i)*midPointYOffSet
				if mpY > yOverallEnd+0.001 {
					break
				}
				pdf.Circle(mpX, mpY, midPointCircleRadius, "F")
				pdf.Circle(xStart, mpY, midPointCircleRadius, "F")
			}

			// print any ticks
			for i := int64(1); i <= fl.ticks; i++ {
				tickStartX := xStart + float64(i)*ticksXOffSet
				tickStartY := yStart + float64(i)*ticksYOffSet
				if tickStartY > yOverallEnd+0.001 {
					break
				}

				tickEndX := tickStartX + math.Cos(tickAngle)*tickLength
				tickEndY := tickStartY + math.Sin(tickAngle)*tickLength
				pdf.Line(tickStartX, tickStartY, tickEndX, tickEndY)
			}
		}
	} else {
		xOverallStart := bnd.left
		xOverallEnd := bnd.right - padding
		yStart := bnd.top + fl.shift
		for xStart := xOverallStart; xStart <= xOverallEnd; xStart += fl.spacing {
			yEnd := bnd.bottom - padding
			height := yEnd - yStart

			// tan = height/width
			// tan = height/(xend-xstart)
			// xend = height/tan +xstart
			xEnd := height/math.Tan(fl.angleRad) + xStart

			// need 0.001 for float imprecision on final line if on the boundary
			if xEnd > xOverallEnd+0.001 {
				xEnd = xOverallEnd
				// work backwards
				height = (xEnd - xStart) * math.Tan(fl.angleRad)
				yEnd = yStart + height
			}

			// if is first record
			// determine the x and y midpoint coordinate offsets
			if xStart == xOverallStart {
				midPointXOffSet = (xEnd - xStart) / float64(fl.midPoints+1)
				midPointYOffSet = (yEnd - yStart) / float64(fl.midPoints+1)
			}

			pdf.Line(xStart, yStart, xEnd, yEnd)

			// print any midpoints
			for i := int64(1); i <= fl.midPoints; i++ {
				mpX := xStart + float64(i)*midPointXOffSet
				mpY := yStart + float64(i)*midPointYOffSet
				if mpX > xOverallEnd+0.001 {
					break
				}
				pdf.Circle(mpX, mpY, midPointCircleRadius, "F")
				pdf.Circle(mpX, yStart, midPointCircleRadius, "F")
			}
		}
	}

	// uses entire bounds
	return bounds{bnd.bottom, bnd.left, bnd.bottom, bnd.left}
}

func (fl flowLines) getWidth() (isStatic bool, width float64) {
	return false, 0
}

func (fl flowLines) getHeight() (isStatic bool, height float64) {
	return false, 0
}

// ---------------------

type sines struct {
	isHorizontal bool
	spacing      float64
	shift        float64
	resolution   float64
	amplitude    float64
	frequency    float64
}

var _ ssElement = sines{}

func (s sines) parseText(text string) (elem ssElement, err error) {
	sOut := sines{}
	if strings.HasPrefix(text, "SINES") {
		sOut = sines{true, 2 * padding, padding / 2, padding / 10, padding, padding}
		text = strings.TrimPrefix(text, "SINES")
	} else if strings.HasPrefix(text, "HSINES") {
		sOut = sines{false, 2 * padding, padding / 2, padding / 10, padding, padding}
		text = strings.TrimPrefix(text, "HSINES")
	} else {
		return nil, nil
	}

	if len(text) == 0 {
		return sOut, nil
	}
	text = strings.TrimPrefix(text, "[")
	text = strings.TrimSuffix(text, "]")
	split := strings.Split(text, ",")

	switch len(split) {
	case 1:
		sOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
	case 2:
		sOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
		sOut.shift, err = strconv.ParseFloat(split[1], 64)
		if err != nil {
			return nil, err
		}
	case 3:
		sOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
		sOut.shift, err = strconv.ParseFloat(split[1], 64)
		if err != nil {
			return nil, err
		}
		sOut.resolution, err = strconv.ParseFloat(split[2], 64)
		if err != nil {
			return nil, err
		}
	case 4:
		sOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
		sOut.shift, err = strconv.ParseFloat(split[1], 64)
		if err != nil {
			return nil, err
		}
		sOut.resolution, err = strconv.ParseFloat(split[2], 64)
		if err != nil {
			return nil, err
		}
		sOut.amplitude, err = strconv.ParseFloat(split[3], 64)
		if err != nil {
			return nil, err
		}
	case 5:
		sOut.spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
		sOut.shift, err = strconv.ParseFloat(split[1], 64)
		if err != nil {
			return nil, err
		}
		sOut.resolution, err = strconv.ParseFloat(split[2], 64)
		if err != nil {
			return nil, err
		}
		sOut.amplitude, err = strconv.ParseFloat(split[3], 64)
		if err != nil {
			return nil, err
		}
		sOut.frequency, err = strconv.ParseFloat(split[4], 64)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("bad input for %v", text)
	}
	return sOut, nil
}

func (s sines) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reducedBounds bounds) {

	pdf.SetLineWidth(thinestLW)

	resolution := s.resolution

	if s.isHorizontal {
		yOverallStart := bnd.top + s.shift
		yOverallEnd := bnd.bottom - padding
		xStart := bnd.left
		xEnd := bnd.right - padding
		for yStart := yOverallStart; yStart <= yOverallEnd; yStart += s.spacing {
			lastPointX := xStart
			lastPointY := yStart
			for eqX := float64(0); true; eqX += resolution {
				if xStart+eqX > xEnd {
					break
				}
				eqY := s.amplitude * math.Cos(s.frequency*eqX)
				if eqX > 0 {
					pdf.Line(lastPointX, lastPointY, xStart+eqX, yStart+eqY)
				}
				lastPointX = xStart + eqX
				lastPointY = yStart + eqY
			}
		}
	} else {
		// TODO
	}

	// uses entire bounds
	return bounds{bnd.bottom, bnd.left, bnd.bottom, bnd.left}
}

func (s sines) getWidth() (isStatic bool, width float64) {
	return false, 0
}

func (s sines) getHeight() (isStatic bool, height float64) {
	return false, 0
}
