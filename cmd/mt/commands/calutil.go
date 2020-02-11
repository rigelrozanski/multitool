package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/cobra"

	cmn "github.com/rigelrozanski/common"
)

// speedy todolists
var (
	CalUtil = &cobra.Command{
		Use:   "calutil",
		Short: "calendar utilities",
	}
	RipDays = &cobra.Command{
		Use:   "ripdays <YYYY-MM-DD> <YYYY-MM-DD>",
		Short: "print a rip daily calendar between the specified dates inclusive",
		Args:  cobra.ExactArgs(2),
		RunE:  RipDaysCmd,
	}
)

func init() {
	CalUtil.AddCommand(RipDays)
	RootCmd.AddCommand(CalUtil)
}

func RipDaysCmd(cmd *cobra.Command, args []string) error {

	startDate, err := cmn.ParseYYYYdMMdDD(args[0])
	if err != nil {
		return err
	}
	endDate, err := cmn.ParseYYYYdMMdDD(args[1])
	if err != nil {
		return err
	}

	// get the no of days
	days := 0
	for i := startDate; !i.After(endDate); i = i.Add(24 * time.Hour) {
		days++
	}

	// get number of pages to create
	noPages := days / 9
	if days%9 != 0 {
		noPages++
	}

	pdf := gofpdf.New("P", "in", "Letter", "")
	pdf.SetMargins(0, 0, 0)

	// create the pdf pages
	for i := 0; i <= noPages; i++ {
		pdf.AddPage()
		AddPageCutMarks(pdf)
	}

	//var opt gofpdf.ImageOptions
	//pdf.ImageOptions(idea.Path(), -10, 1, 0, 0, true, opt, 0, "")

	//for i := startDate; !i.After(endDate); i = i.Add(24 * time.Hour) {
	for cellY := 0; cellY < 3; cellY++ {
		for cellX := 0; cellX < 3; cellX++ {
			for i := 0; i <= noPages; i++ {
				pdf.SetPage(i + 1)
				daysToAdd := time.Duration(i + (noPages+1)*cellX + 3*(noPages+1)*cellY)
				date := startDate.Add(daysToAdd * 24 * time.Hour)
				if date.After(endDate) {
					continue
				}

				pdf.SetFont("times", "B", 20)
				dateStr := date.Format("Monday")
				pdf.Text(float64(cellX)*8.5/3.0+0.3, (float64(cellY)*11)/3.0+0.6, dateStr)

				pdf.SetFont("courier", "B", 12)
				dateStr = date.Format("January 2")
				pdf.Text(float64(cellX)*8.5/3.0+0.4, (float64(cellY)*11)/3.0+0.8, dateStr)

				pdf.SetFont("courier", "B", 7)
				dateStr = date.Format(cmn.LayoutYYYYdMMdDD)
				pdf.Text(float64(cellX)*8.5/3.0+2, (float64(cellY)*11)/3.0+0.5, dateStr)
			}
		}
	}
	//}

	err = pdf.OutputFileAndClose("temp.pdf")
	if err != nil {
		return err
	}

	// print the file
	command1 := fmt.Sprintf("lp temp.pdf")
	output1, err := cmn.Execute(command1)
	fmt.Printf("%v\n%v\n", command1, output1)
	if err != nil {
		return err
	}

	// remove the temp file
	return os.Remove("temp.pdf")
}

// write cut marks
func AddPageCutMarks(pdf *gofpdf.Fpdf) {

	// -
	pdf.Line(0, (float64(11) / 3), 0.5, (float64(11) / 3))         // top-left
	pdf.Line(8, (float64(11) / 3), 8.5, (float64(11) / 3))         // top-right
	pdf.Line(0, (2 * float64(11) / 3), 0.5, (2 * float64(11) / 3)) // lower-left
	pdf.Line(8, (2 * float64(11) / 3), 8.5, (2 * float64(11) / 3)) // lower-right

	// |
	pdf.Line(8.5/3, 0, 8.5/3, 0.5)       // top-left
	pdf.Line(2*8.5/3, 0, 2*8.5/3, 0.5)   // top-right
	pdf.Line(8.5/3, 10.5, 8.5/3, 11)     // lower-left
	pdf.Line(2*8.5/3, 10.5, 2*8.5/3, 11) // lower-right

	// +
	pdf.Line((8.5/3 - 0.5), (float64(11) / 3), (8.5/3 + 0.5), (float64(11) / 3))             // top-left horizontal
	pdf.Line((8.5 / 3), (float64(11)/3 + 0.5), (8.5 / 3), (float64(11)/3 - 0.5))             // top-left vertical
	pdf.Line((2*8.5/3 - 0.5), (float64(11) / 3), (2*8.5/3 + 0.5), (float64(11) / 3))         // top-right horizontal
	pdf.Line((2 * 8.5 / 3), (float64(11)/3 + 0.5), (2 * 8.5 / 3), (float64(11)/3 - 0.5))     // top-right vertical
	pdf.Line((8.5/3 - 0.5), (2 * float64(11) / 3), (8.5/3 + 0.5), (2 * float64(11) / 3))     // lower-left horizontal
	pdf.Line((8.5 / 3), (2*float64(11)/3 + 0.5), (8.5 / 3), (2*float64(11)/3 - 0.5))         // lower-left vertical
	pdf.Line((2*8.5/3 - 0.5), (2 * float64(11) / 3), (2*8.5/3 + 0.5), (2 * float64(11) / 3)) // lower-right horizontal
	pdf.Line((2 * 8.5 / 3), (2*float64(11)/3 + 0.5), (2 * 8.5 / 3), (2*float64(11)/3 - 0.5)) // lower-right vertical
}
