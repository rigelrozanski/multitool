package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/cobra"

	cmn "github.com/rigelrozanski/common"
)

// speedy todolists
var (
	MasonLabels = &cobra.Command{
		Use:   "masonjar <common> <label1,#ofLabel1;label2,#ofLabel2;etc>",
		Short: "print 24 labels for mason jars",
		Args:  cobra.RangeArgs(1, 2),
		RunE:  MasonLabelsCmd,
	}
)

func init() {
	RootCmd.AddCommand(MasonLabels)
}

func MasonLabelsCmd(cmd *cobra.Command, args []string) error {

	var commonStr, specificsUnparsed string

	if len(args) == 2 {
		commonStr = args[0]
		specificsUnparsed = args[1]
	} else {
		specificsUnparsed = args[0]
	}

	var specifics []string
	splitLabels := strings.Split(specificsUnparsed, ";")
	for _, labelNo := range splitLabels {

		splitLabelNo := strings.Split(labelNo, ",")
		switch {
		case len(labelNo) == 0:
			continue
		case len(splitLabelNo) == 2:
			n, err := strconv.Atoi(splitLabelNo[1])
			if err != nil {
				return fmt.Errorf("error, converting %s from %s into integer, error: %s",
					splitLabelNo[1], labelNo, err)
			}
			for i := 0; i < n; i++ {
				specifics = append(specifics, splitLabelNo[0])
			}
		default:
			return fmt.Errorf("error, string %s not in the required format", labelNo)
		}
	}

	pdf := gofpdf.New("P", "in", "Letter", "")
	pdf.SetMargins(0, 0, 0)
	pdf.AddPage()

	pdf.SetFont("courier", "B", 12)
	dateStr := time.Now().Format(cmn.LayoutYYYYdMMdDD)

	i := 0
	for cellY := 0; cellY < 8 && i < len(specifics); cellY++ {
		for cellX := 0; cellX < 3 && i < len(specifics); cellX++ {
			pdf.Text(float64(cellX)*8.5/3.0+0.4, (float64(cellY)*11)/8.0+0.5, dateStr)
			if len(commonStr) > 0 {
				pdf.Text(float64(cellX)*8.5/3.0+0.4, (float64(cellY)*11)/8.0+0.7, commonStr)
			}

			pdf.Text(float64(cellX)*8.5/3.0+0.4, (float64(cellY)*11)/8.0+0.9, specifics[i])
			i++
		}
	}

	err := pdf.OutputFileAndClose("mason-labels.pdf")
	if err != nil {
		panic(err)
	}
	return nil
}
