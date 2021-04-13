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
// - prickles ON the lines
// - blank element
// - wavey lines
// - vertical lines
// - grid element
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
	lines:               LINES[spacing,angle-rad]    examples: LINES; LINES[0.5]; LINES[0.5,pi/10]
	group row:           ROW(elems..)
	group column:        COL(elems..)

pattern example: 
    COL(elem1;COL(elem2;elem3;ROW(elem4;elem5));elem6)`,
		Args: cobra.ExactArgs(1),
		RunE: songsheetCmd,
	}
	padding   = 0.25
	thinLW    = 0.01
	thinestLW = 0.001

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
		bnd = printHeader(pdf, bnd)
	}

	_ = elem.printPDF(pdf, bnd)

	return pdf.OutputFileAndClose("songsheet.pdf")
}

// --------------------------------

func parseElem(text string) (elem ssElement, err error) {

	// NOTE all elements must be registered here
	elemKinds := []ssElement{
		pillar{},
		flowLines{},
		groupElem{},
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

func printHeader(pdf *gofpdf.Fpdf, bnd bounds) (reducedBounds bounds) {
	dateRightOffset := 2.0
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

	// print date
	pdf.SetFont("courier", "", 14)
	pdf.Text(bnd.right-dateRightOffset, bnd.top, "DATE:")

	// print box
	pdf.SetLineWidth(thinLW)

	pdf.Rect(bnd.left, bnd.top+totalHeaderHeight-boxHeight,
		bnd.right-bnd.left-padding, boxHeight, "")

	// print box contents
	conts := []string{"TUNING:", "CAPO:", "BPM:", "TIMESIG:", "FEEL:"}
	xTextAreaStart := bnd.left + boxTextMargin
	xTextAreaEnd := bnd.right - padding - boxTextMargin
	xTextIncr := (xTextAreaEnd - xTextAreaStart) / float64(len(conts))
	for i, cont := range conts {
		pdf.Text(xTextAreaStart+float64(i)*xTextIncr, bnd.top+totalHeaderHeight-boxTextMargin, cont)
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
	isRow bool // otherwise is col
	elems []ssElement
}

var _ ssElement = groupElem{}

func (ge groupElem) parseText(text string) (ssElement, error) {
	if !strings.HasSuffix(text, ")") {
		return nil, nil
	}
	newGE := groupElem{}
	if strings.HasPrefix(text, "ROW(") {
		newGE.isRow = true
		text = strings.TrimPrefix(text, "ROW(")
	} else if strings.HasPrefix(text, "COL(") {
		text = strings.TrimPrefix(text, "COL(")
		newGE.isRow = false
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

	//fmt.Printf("debug split: %v\n", split)

	for _, elemText := range split {
		elem, err := parseElem(elemText)
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
	nonStaticWidth := bnd.Width()
	nonStaticHeight := bnd.Height()
	nonStaticWidthElems := 0
	nonStaticHeightElems := 0

	for _, elem := range ge.elems {
		if ge.isRow {
			isStaticW, elemWidth := elem.getWidth()
			if isStaticW {
				nonStaticWidth -= elemWidth
			} else {
				nonStaticWidthElems++
			}
		} else {
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
			if ge.isRow {
				bnd.right = bnd.left + nonStaticWidthPerElem
				newBounds.left += nonStaticWidthPerElem
			} else {
				// just use all the width
				bnd.right = bnd.left + nonStaticWidth
			}
		}

		isStaticH, elemHeight := elem.getHeight()
		if isStaticH {
			newBounds.top += elemHeight
		} else {
			if ge.isRow {
				// just use all the height
				bnd.bottom = bnd.top + nonStaticHeight
			} else {
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

type flowLines struct {
	//isHorizontal bool
	spacing  float64
	angleRad float64 // in radians
}

var _ ssElement = flowLines{}

func (fl flowLines) parseText(text string) (elem ssElement, err error) {
	if !strings.HasPrefix(text, "LINES") {
		return nil, nil
	}
	trim := strings.TrimPrefix(text, "LINES")
	spacing := 0.5
	angle := 0.0
	if len(trim) == 0 {
		return flowLines{spacing, angle}, nil
	}
	trim = strings.TrimPrefix(trim, "[")
	trim = strings.TrimSuffix(trim, "]")
	split := strings.Split(trim, ",")

	switch len(split) {
	case 1:
		spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
	case 2:

		spacing, err = strconv.ParseFloat(split[0], 64)
		if err != nil {
			return nil, err
		}
		angle, err = strconv.ParseFloat(split[1], 64)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("bad input for %v", text)
	}
	return flowLines{spacing, angle}, nil
}

func (fl flowLines) printPDF(pdf *gofpdf.Fpdf, bnd bounds) (reducedBounds bounds) {

	pdf.SetLineWidth(thinestLW)

	yOverallStart := bnd.top + fl.spacing
	yOverallEnd := bnd.bottom - padding
	for yStart := yOverallStart; yStart < yOverallEnd; yStart += fl.spacing {
		xStart := bnd.left
		xEnd := bnd.right - padding
		width := xEnd - xStart

		// tan = height/width
		// tan = (yend-ystart)/width
		// yend = width*tan+ystart
		yEnd := math.Tan(fl.angleRad)*width + yStart

		pdf.Line(xStart, yStart, xEnd, yEnd)
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
