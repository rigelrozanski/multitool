package commands

import (
	"errors"

	"github.com/jung-kurt/gofpdf"
	wb "github.com/rigelrozanski/wb/lib"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(PrintTodoCmd)
}

// speedy todolists
var PrintTodoCmd = &cobra.Command{
	Use: "print-todo",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, found := wb.GetWB("paper-todo")
		if !found {
			return errors.New("can't find wb paper-todo")
		}
		pdf := gofpdf.New("P", "mm", "Letter", "")
		pdf.AddPage()
		pdf.SetFont("Arial", "", 14)

		for i, item := range items {
			bullet := " - " + item
			pdf.Cell(3, float64(3+12*i), bullet)
			pdf.Cell(110, float64(3+12*i), bullet)
		}

		err := pdf.OutputFileAndClose("hello2.pdf")
		if err != nil {
			return err
		}
		return nil
	},
}
