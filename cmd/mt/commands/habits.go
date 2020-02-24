package commands

import (
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
	activitiesUnfiltered := strings.Split(activitiesRaw, "\n")
	var activities []string
	for _, activity := range activitiesUnfiltered {
		if len(activity) > 1 {
			activities = append(activities, activity)
		}
	}

	minX := margin
	maxX := pageWidthInch - margin

	xActivities := minX
	yActivities := float64(1.5)
	maxStrWidth := float64(0)

	maxYActivity := float64(0)
	maxXDates := float64(0)

	xDates := maxStrWidth + xActivities + 0.1
	yDates := float64(1.3)

	// determine maxXDates, maxYActivity
	stringWidthMargin := 0.1
	for i, activity := range activities {
		width := pdf.GetStringWidth(activity) + stringWidthMargin
		if maxStrWidth < width {
			maxStrWidth = width
		}
		y := yActivities + float64(i)*0.2
		maxYActivity = y + 0.07
	}
	for i := 0; ; i++ {
		// break if past the max X
		xDates = xDates + 0.2
		if xDates > maxX {
			break
		}
		maxXDates = xDates + 0.07
	}

	// fill colour for the tiled rows
	pdf.SetFillColor(230, 230, 230)

	// print horizontal lines
	pdf.Line(minX, yActivities-0.13, maxXDates, yActivities-0.13)
	for i, _ := range activities {
		y := yActivities + float64(i)*0.2
		if i%2 == 0 && (i+1) < len(activities) {
			y2 := yActivities + float64(i+1)*0.2
			width := maxXDates - minX
			height := y2 - y
			pdf.Rect(minX, y+0.07, width, height, "F")
		}
		pdf.Line(minX, y+0.07, maxXDates, y+0.07)
	}

	// print horizontal text habit items
	for i, activity := range activities {
		y := yActivities + float64(i)*0.2
		pdf.Text(xActivities, y, activity)
	}

	// print vertical lines and dates
	xDates = maxStrWidth + xActivities + 0.1
	pdf.Line(xDates+0.07-0.2, 0.3, xDates+0.07-0.2, maxYActivity) // print first vertical line
	for i := 0; ; i++ {
		// print date
		pdf.TransformBegin()
		pdf.TransformRotate(90, xDates, yDates)
		date := startDate.Add(time.Duration(i) * 24 * time.Hour)
		dateStr := date.Format(cmn.LayoutYYYYdMMdDD)
		pdf.Text(xDates, yDates, dateStr)
		pdf.TransformEnd()
		pdf.Line(xDates+0.07, 0.3, xDates+0.07, maxYActivity)
		// break if past the max X
		xDates = xDates + 0.2
		if xDates > maxX {
			break
		}
	}

	// ___________________ OUTPUT
	err := pdf.OutputFileAndClose("temp.pdf")
	if err != nil {
		return err
	}

	_, err = cmn.Execute("open temp.pdf")
	if err != nil {
		return err
	}

	return nil

	// TODO integrate with flag
	//// print the file
	//command1 := fmt.Sprintf("lp -d Brother_HL_L2320D_series temp.pdf")
	//output1, err := cmn.Execute(command1)
	//fmt.Printf("%v\n%v\n", command1, output1)
	//if err != nil {
	//return err
	//}
	//// remove the temp file
	//return os.Remove("temp.pdf")
}
