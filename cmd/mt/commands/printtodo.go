package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/jung-kurt/gofpdf"
	"github.com/rigelrozanski/common"
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
		pdf.SetFont("courier", "", 14)

		for i, item := range items {
			bullet := " - " + item
			pdf.Text(5, float64(10+5*i), bullet)
			pdf.Text(110, float64(10+5*i), bullet)
			pdf.Text(5, float64(155+5*i), bullet)
			pdf.Text(110, float64(155+5*i), bullet)
		}

		err := pdf.OutputFileAndClose("temp.pdf")
		if err != nil {
			return err
		}

		// print the file
		command1 := fmt.Sprintf("lp temp.pdf")
		output1, err := common.Execute(command1)
		fmt.Printf("%v\n%v\n", command1, output1)
		if err != nil {
			return err
		}

		// remove the temp file
		return os.Remove("temp.pdf")
	},
}
