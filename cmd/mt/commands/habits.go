package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/cobra"

	cmn "github.com/rigelrozanski/common"
	"github.com/rigelrozanski/thranch/quac"
)

// speedy todolists
var (
	Habits = &cobra.Command{
		Use:   "habits [YYYY-MM-DD]",
		Short: "print daily activities sheet beginning today (or at the specified date)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  HabitsCmd,
	}
)

func init() {
	RootCmd.AddCommand(Habits)
}

func HabitsCmd(cmd *cobra.Command, args []string) error {

	// get the date
	startDate := time.Now()
	if len(args) == 1 {
		var err error
		startDate, err = cmn.ParseYYYYdMMdDD(args[0])
		if err != nil {
			return err
		}
	}

	pageWidthInch := 8.5
	margin := 0.3
	pdf := gofpdf.New("P", "in", "Letter", "")
	pdf.SetMargins(0, 0, 0)
	pdf.SetFont("courier", "", 12)
	pdf.AddPage()

	// read in activities
	quac.Initialize(os.ExpandEnv("$HOME/.thranch_config"))
	activitiesRaw := quac.GetForApp("habits")
	activities := strings.Split(activitiesRaw, "\n")

	minX := margin
	maxX := pageWidthInch - margin

	xActivities := minX
	yActivities := float64(1.5)
	maxStrWidth := float64(0)
	maxYActivity := float64(0)

	for i, activity := range activities {
		if len(activity) < 2 {
			continue
		}

		// print text
		y := yActivities + float64(i)*0.2
		pdf.Text(xActivities, y, activity)

		// update dimention saves
		width := pdf.GetStringWidth(activity)
		if maxStrWidth < width {
			maxStrWidth = width
		}
		maxYActivity = y + 0.07
	}

	xDates := maxStrWidth + xActivities + 0.1
	yDates := float64(1.3)
	maxXDates := float64(0)

	// print first vertical line
	pdf.Line(xDates+0.07-0.2, 0.3, xDates+0.07-0.2, maxYActivity)

	for i := 0; ; i++ {

		// print date
		pdf.TransformBegin()
		pdf.TransformRotate(90, xDates, yDates)
		date := startDate.Add(time.Duration(i) * 24 * time.Hour)
		dateStr := date.Format(cmn.LayoutYYYYdMMdDD)
		pdf.Text(xDates, yDates, dateStr)
		pdf.TransformEnd()

		// print vertical line
		pdf.Line(xDates+0.07, 0.3, xDates+0.07, maxYActivity)

		// break if past the max X
		xDates = xDates + 0.2
		if xDates > maxX {
			break
		}

		maxXDates = xDates + 0.07
	}

	// print horizontal lines
	pdf.Line(minX, yActivities-0.13, maxXDates, yActivities-0.13)
	for i, activity := range activities {
		if len(activity) < 2 {
			continue
		}
		y := yActivities + float64(i)*0.2
		pdf.Line(minX, y+0.07, maxXDates, y+0.07)

	}

	err := pdf.OutputFileAndClose("temp.pdf")
	if err != nil {
		return err
	}

	// print the file
	command1 := fmt.Sprintf("lp -d Brother_HL_L2320D_series temp.pdf")
	output1, err := cmn.Execute(command1)
	fmt.Printf("%v\n%v\n", command1, output1)
	if err != nil {
		return err
	}

	// remove the temp file
	return os.Remove("temp.pdf")
}
