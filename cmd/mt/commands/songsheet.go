package commands

import (
	"math"

	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/cobra"
)

// speedy todolists
var (
	SongsheetCmd = &cobra.Command{
		Use:   "songsheet",
		Short: "print songsheet",
		RunE:  songsheetCmd,
	}
)

//var marginMin float64
//var gridSide float64
//var lineWidth float64

func init() {
	//GridpaperCmd.PersistentFlags().Float64VarP(&marginMin, "margin", "m", 0.4, "define the minimum margin (in inches)")
	//GridpaperCmd.PersistentFlags().Float64VarP(&gridSide, "gridSide", "g", 0.2, "define the side-length of a gridcell (in inches)")
	//GridpaperCmd.PersistentFlags().Float64VarP(&lineWidth, "linewidth", "l", 0.002, "define the line width (in inches)")
	RootCmd.AddCommand(SongsheetCmd)
}

func songsheetCmd(cmd *cobra.Command, args []string) error {

	maxWideCells := int64((8.5 - 2*marginMin) / gridSide)
	maxHeightCells := int64((11.0 - 2*marginMin) / gridSide)
	revisedSideMargins := (8.5 - float64(maxWideCells)*gridSide) / 2
	revisedUpperLowerMargins := (11.0 - float64(maxHeightCells)*gridSide) / 2

	pdf := gofpdf.New("P", "in", "Letter", "")
	pdf.SetMargins(revisedSideMargins, revisedUpperLowerMargins, revisedSideMargins)
	pdf.SetLineWidth(lineWidth)
	pdf.AddPage()

	// XXX add left bounds here
	overallBound := bounds{0, 0, 11, 8.5}
	bnd := printHeader(pdf, overallBound)

	col := column{false, true, true}
	col2 := column{false, false, false}
	fl := flowLines{0.5, math.Pi / 8}

	bnd = col.printColumn(pdf, bnd)
	bnd = col2.printColumn(pdf, bnd)
	fl.printToPdf(pdf, bnd)

	err := pdf.OutputFileAndClose("songsheet.pdf")
	if err != nil {
		return err
	}

	return nil
}

type bounds struct {
	top    float64
	left   float64
	bottom float64
	right  float64
}

func printHeader(pdf *gofpdf.Fpdf, bnd bounds) (reducedBounds bounds) {
	margin := 0.25
	dateRightOffset := 2.0
	totalHeaderHeight := 1.0
	boxHeight := 0.25
	boxTextMargin := 0.06

	// print date
	pdf.SetFont("courier", "", 14)
	pdf.Text(bnd.right-dateRightOffset, bnd.top+margin, "DATE:")

	// print box
	pdf.SetLineWidth(0.01)

	pdf.Rect(bnd.left+margin, bnd.top+totalHeaderHeight-boxHeight,
		bnd.right-bnd.left-2*margin, boxHeight, "")

	// print box contents
	conts := []string{"TUNING:", "CAPO:", "BPM:", "TIMESIG:", "FEEL:"}
	xTextAreaStart := bnd.left + margin + boxTextMargin
	xTextAreaEnd := bnd.right - margin - boxTextMargin
	xTextIncr := (xTextAreaEnd - xTextAreaStart) / float64(len(conts))
	for i, cont := range conts {
		pdf.Text(xTextAreaStart+float64(i)*xTextIncr, totalHeaderHeight-boxTextMargin, cont)
	}

	return bounds{totalHeaderHeight, bnd.left, bnd.bottom, bnd.right}
}

type column struct {
	hasReverseStrings bool
	isHorizontal      bool
	hasPrickles       bool
}

func (col column) printColumn(pdf *gofpdf.Fpdf, bnd bounds) (reducedBounds bounds) {

	margin := 0.25
	thicknessIndicatorMargin := 0.125 // the top zone of the column that shows the guitar string thicknesses
	spacing := 0.125
	cactusZoneWidth := 0.0
	cactusPrickleSpacing := 2 * spacing
	if col.hasPrickles {
		cactusZoneWidth = 2 * spacing // one for the cactus
	}

	// thicknesses of guitar strings from thick to thin
	thicknesses := []float64{0.0472, 0.0314, 0.0236, 0.0157, 0.0079, 0.0039}
	noLines := len(thicknesses)
	if col.hasReverseStrings {
		thicknessesRev := make([]float64, len(thicknesses))
		j := len(thicknesses) - 1
		for i := 0; i < len(thicknesses); i++ {
			thicknessesRev[j] = thicknesses[i]
			j--
		}
		thicknesses = thicknessesRev
	}

	// print thicknesses
	var xStart, xEnd, yStart, yEnd float64
	for i := 0; i < noLines; i++ {
		pdf.SetLineWidth(thicknesses[i])
		if col.isHorizontal {
			yStart = bnd.top + margin + cactusZoneWidth + (float64(i) * spacing)
			yEnd = yStart
			xStart = bnd.left + margin
			xEnd = xStart + thicknessIndicatorMargin
		} else {
			xStart = bnd.left + margin + cactusZoneWidth + (float64(i) * spacing)
			xEnd = xStart
			yStart = bnd.top + margin
			yEnd = yStart + thicknessIndicatorMargin
		}

		pdf.Line(xStart, yStart, xEnd, yEnd)
	}

	// print seperator
	pdf.SetLineWidth(0.001)
	if col.isHorizontal {
		yStart = bnd.top + margin + cactusZoneWidth
		yEnd = yStart + float64(noLines-1)*spacing
		xStart = bnd.left + margin + thicknessIndicatorMargin
		xEnd = xStart
	} else {
		xStart = bnd.left + margin + cactusZoneWidth
		xEnd = xStart + float64(noLines-1)*spacing
		yStart = bnd.top + margin + thicknessIndicatorMargin
		yEnd = yStart
	}
	pdf.Line(xStart, yStart, xEnd, yEnd)

	// print column lines
	for i := 0; i < noLines; i++ {
		pdf.SetLineWidth(0.001)
		if col.isHorizontal {
			yStart = bnd.top + margin + cactusZoneWidth + (float64(i) * spacing)
			yEnd = yStart
			xStart = bnd.left + margin + thicknessIndicatorMargin
			xEnd = bnd.right - margin
		} else {
			xStart = bnd.left + margin + cactusZoneWidth + (float64(i) * spacing)
			xEnd = xStart
			yStart = bnd.top + margin + thicknessIndicatorMargin
			yEnd = bnd.bottom - margin
		}

		pdf.Line(xStart, yStart, xEnd, yEnd)
	}

	// print cactus prickles
	if col.hasPrickles {
		if col.isHorizontal {
			xStart := bnd.left + margin + thicknessIndicatorMargin + cactusPrickleSpacing/2
			xEnd := bnd.right - margin
			for x := xStart; x < xEnd; x += cactusPrickleSpacing {
				pdf.SetLineWidth(0.001)
				yTopStart := bnd.top + margin
				yTopEnd := yTopStart + cactusZoneWidth/2
				yBottomStart := bnd.top + margin + cactusZoneWidth + (float64(noLines-1) * spacing) + cactusZoneWidth/2
				yBottomEnd := yBottomStart + cactusZoneWidth/2

				pdf.Line(x, yTopStart, x, yTopEnd)
				pdf.Line(x, yBottomStart, x, yBottomEnd)
			}
		} else {
			yStart := bnd.top + margin + thicknessIndicatorMargin + cactusPrickleSpacing/2
			yEnd := bnd.bottom - margin
			for y := yStart; y < yEnd; y += cactusPrickleSpacing {
				pdf.SetLineWidth(0.001)
				xLeftStart := bnd.left + margin
				xLeftEnd := xLeftStart + cactusZoneWidth/2
				xRightStart := bnd.left + margin + cactusZoneWidth + (float64(noLines-1) * spacing) + cactusZoneWidth/2
				xRightEnd := xRightStart + cactusZoneWidth/2

				pdf.Line(xLeftStart, y, xLeftEnd, y)
				pdf.Line(xRightStart, y, xRightEnd, y)
			}
		}
	}

	loss := 2*cactusZoneWidth + 2*margin + (float64(noLines-1) * spacing)
	if col.isHorizontal {
		return bounds{bnd.top + loss, bnd.left, bnd.bottom, bnd.right}
	} else {
		return bounds{bnd.top, bnd.left + loss, bnd.bottom, bnd.right}
	}
}

type flowLines struct {
	//isHorizontal bool
	spacing  float64
	angleRad float64 // in radians
}

func (fl flowLines) printToPdf(pdf *gofpdf.Fpdf, bnd bounds) {

	margin := 0.25
	pdf.SetLineWidth(0.001)

	yOverallStart := bnd.top + margin
	yOverallEnd := bnd.bottom - margin
	for yStart := yOverallStart; yStart < yOverallEnd; yStart += fl.spacing {
		xStart := bnd.left //+ margin // XXX TODO need to only pad if previously unpadded on right... should be done in main level
		xEnd := bnd.right - margin
		width := xEnd - xStart
		// tan = height/width
		// tan = (yend-ystart)/width
		// yend = width*tan+ystart
		yEnd := math.Tan(fl.angleRad)*width + yStart

		pdf.Line(xStart, yStart, xEnd, yEnd)
	}
}
