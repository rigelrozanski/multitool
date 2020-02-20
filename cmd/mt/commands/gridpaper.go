package commands

import (
	"github.com/jung-kurt/gofpdf"
	cmn "github.com/rigelrozanski/common"
	"github.com/spf13/cobra"
)

// speedy todolists
var (
	GridpaperCmd = &cobra.Command{
		Use:   "grid",
		Short: "print gridpaper",
		RunE:  gridpaperCmd,
	}
)

var marginMin float64
var gridSide float64
var lineWidth float64

func init() {
	GridpaperCmd.PersistentFlags().Float64VarP(&marginMin, "margin", "m", 0.4, "define the minimum margin (in inches)")
	GridpaperCmd.PersistentFlags().Float64VarP(&gridSide, "gridSide", "g", 0.2, "define the side-length of a gridcell (in inches)")
	GridpaperCmd.PersistentFlags().Float64VarP(&lineWidth, "linewidth", "l", 0.002, "define the line width (in inches)")
	RootCmd.AddCommand(GridpaperCmd)
}

func gridpaperCmd(cmd *cobra.Command, args []string) error {

	maxWideCells := int64((8.5 - 2*marginMin) / gridSide)
	maxHeightCells := int64((11.0 - 2*marginMin) / gridSide)
	revisedSideMargins := (8.5 - float64(maxWideCells)*gridSide) / 2
	revisedUpperLowerMargins := (11.0 - float64(maxHeightCells)*gridSide) / 2
	xMin := revisedSideMargins
	xMax := xMin + float64(maxWideCells)*gridSide
	yMin := revisedUpperLowerMargins
	yMax := yMin + float64(maxHeightCells)*gridSide

	pdf := gofpdf.New("P", "in", "Letter", "")
	pdf.SetMargins(revisedSideMargins, revisedUpperLowerMargins, revisedSideMargins)
	pdf.SetLineWidth(lineWidth)
	pdf.AddPage()

	// |
	for x := xMin; x <= xMax+gridSide/10; x += gridSide { // need to use +gridSide/10 to make up for rounding error
		pdf.Line(x, yMin, x, yMax)
	}

	// -
	for y := yMin; y <= yMax+gridSide/10; y += gridSide {
		pdf.Line(xMin, y, xMax, y)
	}

	err := pdf.OutputFileAndClose("gridpaper.pdf")
	if err != nil {
		return err
	}

	_, err = cmn.Execute("open gridpaper.pdf")
	if err != nil {
		return err
	}

	return nil
}
