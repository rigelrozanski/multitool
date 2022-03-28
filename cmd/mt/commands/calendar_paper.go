package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/rigelrozanski/common"
	"github.com/spf13/cobra"
)

// Lock2yamlCmd represents the lock2yaml command
var (
	PaperCalCmd = &cobra.Command{
		Use:   "paper-cal [year]",
		Short: "make a paper calendar for your year",
		Args:  cobra.ExactArgs(1),
		RunE:  paperCalCmd,
	}
)

func init() {
	RootCmd.AddCommand(PaperCalCmd)
}

func paperCalCmd(cmd *cobra.Command, args []string) error {

	year, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	// setup the page ordering
	pgs := []struct {
		A monthSec // top left
		B monthSec // top right
		C monthSec // bottom left
		D monthSec // bottome right
	}{
		{
			monthSec{true, 8, year},
			monthSec{true, 8, year},
			monthSec{true, 4, year},
			monthSec{true, 12, year},
		}, {
			monthSec{true, 9, year},
			monthSec{true, 7, year},
			monthSec{false, 0, 0},
			monthSec{true, 3, year},
		}, {
			monthSec{true, 7, year},
			monthSec{true, 9, year},
			monthSec{true, 3, year},
			monthSec{false, 0, 0},
		}, {
			monthSec{true, 10, year},
			monthSec{true, 6, year},
			monthSec{false, 0, 0},
			monthSec{true, 2, year},
		}, {
			monthSec{true, 6, year},
			monthSec{true, 10, year},
			monthSec{true, 2, year},
			monthSec{false, 0, 0},
		}, {
			monthSec{true, 11, year},
			monthSec{true, 5, year},
			monthSec{false, 0, 0},
			monthSec{true, 1, year},
		}, {
			monthSec{true, 5, year},
			monthSec{true, 11, year},
			monthSec{true, 1, year},
			monthSec{false, 0, 0},
		}, {
			monthSec{true, 12, year},
			monthSec{true, 4, year},
			monthSec{false, 0, 0},
			monthSec{false, 0, 0},
		},
	}

	FNs := []string{}
	for i, pg := range pgs {
		fileContents := svgContents(pg.A, pg.B, pg.C, pg.D)
		ioutil.WriteFile("./temp.svg", []byte(fileContents), 0644)
		fn := fmt.Sprintf("page%v.pdf", i+1)
		common.Execute("/Applications/Inkscape.app/Contents/MacOS/inkscape temp.svg --export-area-drawing" +
			" --batch-process --export-type=pdf --export-filename=" + fn)
		FNs = append(FNs, fn)
	}

	// merge files
	outCalFN := fmt.Sprintf("./%v_calendar.pdf", year)
	api.MergeFile(FNs, outCalFN, pdfcpu.NewDefaultConfiguration())

	// remove temp files
	err = os.Remove("./temp.svg")
	if err != nil {
		panic(err)
	}
	for _, fn := range FNs {
		err = os.Remove("./" + fn)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("print off the pdf (2-sided, borderless, 97% Scale [TODO-FIX]),\n" +
		"cut top and bottom then stack the tops on the bottoms")
	return nil
}

type monthSec struct {
	show    bool
	monthNo int
	year    int
}

func svgContents(monthA, monthB, monthC, monthD monthSec) string {

	out := `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<svg
   xmlns:dc="http://purl.org/dc/elements/1.1/"
   xmlns:cc="http://creativecommons.org/ns#"
   xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
   xmlns:svg="http://www.w3.org/2000/svg"
   xmlns="http://www.w3.org/2000/svg"
   xmlns:sodipodi="http://sodipodi.sourceforge.net/DTD/sodipodi-0.dtd"
   xmlns:inkscape="http://www.inkscape.org/namespaces/inkscape"
   width="8.5in"
   height="11in"
   viewBox="0 0 215.9 279.4"
   version="1.1"
   id="svg8"
   sodipodi:docname="month_template.svg"
   inkscape:version="1.0.2 (e86c8708, 2021-01-15)">
  <defs
     id="defs2" />
  <sodipodi:namedview
     id="base"
     pagecolor="#ffffff"
     bordercolor="#666666"
     borderopacity="1.0"
     inkscape:pageopacity="0.0"
     inkscape:pageshadow="2"
     inkscape:zoom="0.90028971"
     inkscape:cx="228.20345"
     inkscape:cy="70.157678"
     inkscape:document-units="in"
     inkscape:current-layer="layer1"
     inkscape:document-rotation="0"
     showgrid="false"
     units="in"
     inkscape:window-width="1252"
     inkscape:window-height="855"
     inkscape:window-x="44"
     inkscape:window-y="23"
     inkscape:window-maximized="0"
     showguides="true">
    <inkscape:grid
       type="xygrid"
       id="grid833"
       units="in"
       spacingx="107.95"
       spacingy="139.7"
       enabled="false" />
    <inkscape:grid
       type="xygrid"
       id="grid835"
       units="in"
       spacingx="6.35"
       spacingy="6.35"
       enabled="true"
       visible="true"
       empspacing="4"
       originx="2.54"
       originy="13.97" />
  </sodipodi:namedview>
  <metadata
     id="metadata5">
    <rdf:RDF>
      <cc:Work
         rdf:about="">
        <dc:format>image/svg+xml</dc:format>
        <dc:type
           rdf:resource="http://purl.org/dc/dcmitype/StillImage" />
        <dc:title />
      </cc:Work>
    </rdf:RDF>
  </metadata>
  <g
     inkscape:label="Layer 1"
     inkscape:groupmode="layer"
     id="layer1">
    <path
       style="fill:none;stroke:#cccccc;stroke-width:0.26458300000000001px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
       d="M 107.95,0 V 279.4"
       id="path841" />
    <path
       style="fill:none;stroke:#cccccc;stroke-width:0.26458300000000001px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
       d="M 0,139.7 H 215.9"
       id="path843" />`

	ContentSecA := svgContentsSecA(monthA)
	ContentSecB := svgContentsSecB(monthB)
	ContentSecC := svgContentsSecC(monthC)
	ContentSecD := svgContentsSecD(monthD)

	outEnd := `  </g>
</svg>`
	return out + ContentSecA + ContentSecB + ContentSecC + ContentSecD + outEnd
}

func monthNameTimeConst(m monthSec) (string, time.Month) {
	switch m.monthNo {
	case 1:
		return fmt.Sprintf("January - %v", m.year), time.January
	case 2:
		return "February", time.February
	case 3:
		return "March", time.March
	case 4:
		return "April", time.April
	case 5:
		return "May", time.May
	case 6:
		return "June", time.June
	case 7:
		return "July", time.July
	case 8:
		return "August", time.August
	case 9:
		return "September", time.September
	case 10:
		return "October", time.October
	case 11:
		return "November", time.November
	case 12:
		return "December", time.December
	}
	panic("bad month number")
	return "", time.January
}

// occurance = 1 for the first occurance
// for occurance 1 of each weekday this function will return the prev month's
// dayNumber for each weekday before the weekday which is the first day of the
// provided month
func GetDayForNthWeekday(
	year int, month time.Month, wd time.Weekday, occurance int) (dayNumber int) {

	occCount := 0
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	// shift everything up for the first week
	switch firstOfMonth.Weekday() {
	case time.Monday:
	case time.Tuesday:
		if wd == time.Monday {
			occCount++
		}
	case time.Wednesday:
		if wd == time.Monday || wd == time.Tuesday {
			occCount++
		}
	case time.Thursday:
		if wd == time.Monday || wd == time.Tuesday || wd == time.Wednesday {
			occCount++
		}
	case time.Friday:
		if wd == time.Monday || wd == time.Tuesday || wd == time.Wednesday ||
			wd == time.Thursday {
			occCount++
		}
	case time.Saturday:
		if wd == time.Monday || wd == time.Tuesday || wd == time.Wednesday ||
			wd == time.Thursday || wd == time.Friday {
			occCount++
		}
	case time.Sunday:
		if wd == time.Monday || wd == time.Tuesday || wd == time.Wednesday ||
			wd == time.Thursday || wd == time.Friday || wd == time.Saturday {
			occCount++
		}
	}

	// get from previous month if appropriate
	// if occCount == 1 then must be on prev month (from above switch)
	if occurance == 1 && occCount == 1 {
		md := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		prevMD := md.AddDate(0, -1, 0)
		return GetDayForLastWeekday(prevMD, wd)
	}

	// get for this month
	for dn := 1; dn <= lastOfMonth.Day(); dn++ {
		d := time.Date(year, month, dn, 0, 0, 0, 0, time.UTC)
		thisWD := d.Weekday()
		if thisWD == wd {
			occCount++
		}
		if occCount == occurance {
			return dn
			break
		}
	}

	// if haven't returned already must be beyond the end of the month
	md := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	nextMD := md.AddDate(0, 1, 0)
	return GetDayForFirstWeekday(nextMD, wd)
}

func GetDayForLastWeekday(
	monthsDate time.Time, wd time.Weekday) (dayNumber int) {

	DnAtOcc := 0
	firstOfMonth := time.Date(monthsDate.Year(), monthsDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	for dn := 20; dn <= lastOfMonth.Day(); dn++ {
		d := time.Date(monthsDate.Year(), monthsDate.Month(), dn, 0, 0, 0, 0, time.UTC)
		thisWD := d.Weekday()
		if thisWD == wd {
			DnAtOcc = dn
		}
	}
	return DnAtOcc
}

func GetDayForFirstWeekday(
	monthsDate time.Time, wd time.Weekday) (dayNumber int) {

	firstOfMonth := time.Date(monthsDate.Year(), monthsDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	for dn := 1; dn <= lastOfMonth.Day(); dn++ {
		d := time.Date(monthsDate.Year(), monthsDate.Month(), dn, 0, 0, 0, 0, time.UTC)
		thisWD := d.Weekday()
		if thisWD == wd {
			return dn
		}
	}
	return 0
}

// TODO seperate all coordinates
var (
	AMonthCoorX = "46.225574"
	AMonthCoorY = "6.6202736"
)

func svgContentsSecA(m monthSec) string {

	if m.show == false {
		return ""
	}

	out := `
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="46.225574"
       y="6.6202736"
       id="text1075"><tspan
         sodipodi:role="line"
         id="tspan1073"
         x="46.225574"
         y="6.6202736"
         style="stroke-width:0.264583">MONTHA</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="2.54"
       y="11.206953"
       id="text847"><tspan
         sodipodi:role="line"
         id="tspan845"
         x="2.54"
         y="11.206953"
         style="stroke-width:0.264583">Mon</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="27.939999"
       y="11.206953"
       id="text867"><tspan
         sodipodi:role="line"
         id="tspan865"
         x="27.939999"
         y="11.206953"
         style="stroke-width:0.264583">Tue</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="53.34"
       y="11.206953"
       id="text871"><tspan
         sodipodi:role="line"
         id="tspan869"
         x="53.34"
         y="11.206953"
         style="stroke-width:0.264583">Wed</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="78.739998"
       y="11.206953"
       id="text875"><tspan
         sodipodi:role="line"
         id="tspan873"
         x="78.739998"
         y="11.206953"
         style="stroke-width:0.264583">Thu</tspan></text>
	<g
       id="group_rect_A">
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 2.5399999,12.088617 H 104.14 l 1e-5,127.000003 H 2.5399999 Z"
         id="path1870" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 27.94,12.088617 -3e-6,127.000003"
         id="path1872" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 53.339999,12.088617 V 139.08862"
         id="path1874" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 78.739996,12.088617 8e-6,127.000003"
         id="path1876" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 2.5399999,37.488608 104.14,37.488614"
         id="path1878" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 2.5400023,62.888613 104.14,62.888611"
         id="path1880" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 2.5400023,88.28861 H 104.14"
         id="path1882" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 2.5399985,113.68861 H 104.14"
         id="path1884" />
    </g>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="23.118895"
       y="15.030036"
       id="text1215"><tspan
         sodipodi:role="line"
         id="tspan1213"
         x="23.118895"
         y="15.030036"
         style="stroke-width:0.264583">AMo0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="48.518917"
       y="15.030036"
       id="text1313"><tspan
         sodipodi:role="line"
         x="48.518917"
         y="15.030036"
         style="stroke-width:0.264583"
         id="tspan1447">ATu0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="73.917442"
       y="15.030036"
       id="text1317"><tspan
         sodipodi:role="line"
         id="tspan1315"
         x="73.917442"
         y="15.030036"
         style="stroke-width:0.264583">AWe0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="99.317444"
       y="15.030036"
       id="text1321"><tspan
         sodipodi:role="line"
         id="tspan1319"
         x="99.317444"
         y="15.030036"
         style="stroke-width:0.264583">ATh0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="23.118895"
       y="40.431011"
       id="text1453"><tspan
         sodipodi:role="line"
         id="tspan1451"
         x="23.118895"
         y="40.431011"
         style="stroke-width:0.264583">AMo1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="48.518917"
       y="40.431011"
       id="text1457"><tspan
         sodipodi:role="line"
         x="48.518917"
         y="40.431011"
         style="stroke-width:0.264583"
         id="tspan1455">ATu1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="73.917442"
       y="40.431011"
       id="text1461"><tspan
         sodipodi:role="line"
         id="tspan1459"
         x="73.917442"
         y="40.431011"
         style="stroke-width:0.264583">AWe1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="99.317444"
       y="40.431011"
       id="text1465"><tspan
         sodipodi:role="line"
         id="tspan1463"
         x="99.317444"
         y="40.431011"
         style="stroke-width:0.264583">ATh1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="23.118895"
       y="65.831009"
       id="text1481"><tspan
         sodipodi:role="line"
         id="tspan1479"
         x="23.118895"
         y="65.831009"
         style="stroke-width:0.264583">AMo2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="48.518917"
       y="65.831009"
       id="text1485"><tspan
         sodipodi:role="line"
         x="48.518917"
         y="65.831009"
         style="stroke-width:0.264583"
         id="tspan1483">ATu2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="73.917442"
       y="65.831009"
       id="text1489"><tspan
         sodipodi:role="line"
         id="tspan1487"
         x="73.917442"
         y="65.831009"
         style="stroke-width:0.264583">AWe2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="99.317444"
       y="65.831009"
       id="text1493"><tspan
         sodipodi:role="line"
         id="tspan1491"
         x="99.317444"
         y="65.831009"
         style="stroke-width:0.264583">ATh2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="23.118895"
       y="91.23101"
       id="text1509"><tspan
         sodipodi:role="line"
         id="tspan1507"
         x="23.118895"
         y="91.23101"
         style="stroke-width:0.264583">AMo3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="48.518917"
       y="91.23101"
       id="text1513"><tspan
         sodipodi:role="line"
         x="48.518917"
         y="91.23101"
         style="stroke-width:0.264583"
         id="tspan1511">ATu3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="73.917442"
       y="91.23101"
       id="text1517"><tspan
         sodipodi:role="line"
         id="tspan1515"
         x="73.917442"
         y="91.23101"
         style="stroke-width:0.264583">AWe3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="99.317444"
       y="91.23101"
       id="text1521"><tspan
         sodipodi:role="line"
         id="tspan1519"
         x="99.317444"
         y="91.23101"
         style="stroke-width:0.264583">ATh3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="23.118895"
       y="116.63101"
       id="text1537"><tspan
         sodipodi:role="line"
         id="tspan1535"
         x="23.118895"
         y="116.63101"
         style="stroke-width:0.264583">AMo4</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="48.518917"
       y="116.63101"
       id="text1541"><tspan
         sodipodi:role="line"
         x="48.518917"
         y="116.63101"
         style="stroke-width:0.264583"
         id="tspan1539">ATu4</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="73.917442"
       y="116.63101"
       id="text1545"><tspan
         sodipodi:role="line"
         id="tspan1543"
         x="73.917442"
         y="116.63101"
         style="stroke-width:0.264583">AWe4</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="99.317444"
       y="116.63101"
       id="text1549"><tspan
         sodipodi:role="line"
         id="tspan1547"
         x="99.317444"
         y="116.63101"
         style="stroke-width:0.264583">ATh4</tspan></text>
	`

	mn, mt := monthNameTimeConst(m)
	out = strings.Replace(out, "MONTHA", mn, 1)
	out = strings.Replace(out, "AMo0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Monday, 1)), 1)
	out = strings.Replace(out, "ATu0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Tuesday, 1)), 1)
	out = strings.Replace(out, "AWe0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Wednesday, 1)), 1)
	out = strings.Replace(out, "ATh0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Thursday, 1)), 1)
	out = strings.Replace(out, "AMo1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Monday, 2)), 1)
	out = strings.Replace(out, "ATu1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Tuesday, 2)), 1)
	out = strings.Replace(out, "AWe1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Wednesday, 2)), 1)
	out = strings.Replace(out, "ATh1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Thursday, 2)), 1)
	out = strings.Replace(out, "AMo2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Monday, 3)), 1)
	out = strings.Replace(out, "ATu2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Tuesday, 3)), 1)
	out = strings.Replace(out, "AWe2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Wednesday, 3)), 1)
	out = strings.Replace(out, "ATh2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Thursday, 3)), 1)
	out = strings.Replace(out, "AMo3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Monday, 4)), 1)
	out = strings.Replace(out, "ATu3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Tuesday, 4)), 1)
	out = strings.Replace(out, "AWe3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Wednesday, 4)), 1)
	out = strings.Replace(out, "ATh3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Thursday, 4)), 1)
	out = strings.Replace(out, "AMo4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Monday, 5)), 1)
	out = strings.Replace(out, "ATu4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Tuesday, 5)), 1)
	out = strings.Replace(out, "AWe4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Wednesday, 5)), 1)
	out = strings.Replace(out, "ATh4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Thursday, 5)), 1)

	return out
}

func svgContentsSecB(m monthSec) string {

	if m.show == false {
		return ""
	}

	out := `
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="112.07753"
       y="11.206891"
       id="text847-8"><tspan
         sodipodi:role="line"
         id="tspan845-5"
         x="112.07753"
         y="11.206891"
         style="stroke-width:0.264583">Fri</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="137.47739"
       y="11.206891"
       id="text867-3"><tspan
         sodipodi:role="line"
         id="tspan865-7"
         x="137.47739"
         y="11.206891"
         style="stroke-width:0.264583">Sat</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="162.87704"
       y="11.206891"
       id="text871-6"><tspan
         sodipodi:role="line"
         id="tspan869-8"
         x="162.87704"
         y="11.206891"
         style="stroke-width:0.264583">Sun</tspan></text>
    <g
       id="group_rect_B">
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 112.07753,12.088558 h 76.19951 V 139.08938 h -76.19951 z"
         id="path1847" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 137.47738,12.088558 V 139.08938"
         id="path1849" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 162.87738,12.088558 V 139.08938"
         id="path1851" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 112.07753,37.488555 h 76.19951"
         id="path1853" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 112.07753,62.888563 h 76.19951"
         id="path1855" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 112.07753,88.288999 76.19951,-2.64e-4"
         id="path1857" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 112.07753,113.68939 76.19951,-2.6e-4"
         id="path1859" />
    </g>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="132.65486"
       y="15.030036"
       id="text1389"><tspan
         sodipodi:role="line"
         id="tspan1387"
         x="132.65486"
         y="15.030036"
         style="stroke-width:0.264583">BFr0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="158.04651"
       y="15.030036"
       id="text1393"><tspan
         sodipodi:role="line"
         id="tspan1391"
         x="158.04651"
         y="15.030036"
         style="stroke-width:0.264583">BSa0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="183.4465"
       y="15.030036"
       id="text1397"><tspan
         sodipodi:role="line"
         id="tspan1395"
         x="183.4465"
         y="15.030036"
         style="stroke-width:0.264583">BSu0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="132.65486"
       y="40.431011"
       id="text1469"><tspan
         sodipodi:role="line"
         id="tspan1467"
         x="132.65486"
         y="40.431011"
         style="stroke-width:0.264583">BFr1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="158.04651"
       y="40.431011"
       id="text1473"><tspan
         sodipodi:role="line"
         id="tspan1471"
         x="158.04651"
         y="40.431011"
         style="stroke-width:0.264583">BSa1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="183.4465"
       y="40.431011"
       id="text1477"><tspan
         sodipodi:role="line"
         id="tspan1475"
         x="183.4465"
         y="40.431011"
         style="stroke-width:0.264583">BSu1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="132.65486"
       y="65.831009"
       id="text1497"><tspan
         sodipodi:role="line"
         id="tspan1495"
         x="132.65486"
         y="65.831009"
         style="stroke-width:0.264583">BFr2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="158.04651"
       y="65.831009"
       id="text1501"><tspan
         sodipodi:role="line"
         id="tspan1499"
         x="158.04651"
         y="65.831009"
         style="stroke-width:0.264583">BSa2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="183.4465"
       y="65.831009"
       id="text1505"><tspan
         sodipodi:role="line"
         id="tspan1503"
         x="183.4465"
         y="65.831009"
         style="stroke-width:0.264583">BSu2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="132.65486"
       y="91.23101"
       id="text1525"><tspan
         sodipodi:role="line"
         id="tspan1523"
         x="132.65486"
         y="91.23101"
         style="stroke-width:0.264583">BFr3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="158.04651"
       y="91.23101"
       id="text1529"><tspan
         sodipodi:role="line"
         id="tspan1527"
         x="158.04651"
         y="91.23101"
         style="stroke-width:0.264583">BSa3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="183.4465"
       y="91.23101"
       id="text1533"><tspan
         sodipodi:role="line"
         id="tspan1531"
         x="183.4465"
         y="91.23101"
         style="stroke-width:0.264583">BSu3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="132.65486"
       y="116.63101"
       id="text1553"><tspan
         sodipodi:role="line"
         id="tspan1551"
         x="132.65486"
         y="116.63101"
         style="stroke-width:0.264583">BFr4</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="158.04651"
       y="116.63101"
       id="text1557"><tspan
         sodipodi:role="line"
         id="tspan1555"
         x="158.04651"
         y="116.63101"
         style="stroke-width:0.264583">BSa4</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="183.4465"
       y="116.63101"
       id="text1561"><tspan
         sodipodi:role="line"
         id="tspan1559"
         x="183.4465"
         y="116.63101"
         style="stroke-width:0.264583">BSu4</tspan></text>
	`

	_, mt := monthNameTimeConst(m)
	out = strings.Replace(out, "BFr0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Friday, 1)), 1)
	out = strings.Replace(out, "BSa0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Saturday, 1)), 1)
	out = strings.Replace(out, "BSu0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Sunday, 1)), 1)
	out = strings.Replace(out, "BFr1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Friday, 2)), 1)
	out = strings.Replace(out, "BSa1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Saturday, 2)), 1)
	out = strings.Replace(out, "BSu1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Sunday, 2)), 1)
	out = strings.Replace(out, "BFr2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Friday, 3)), 1)
	out = strings.Replace(out, "BSa2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Saturday, 3)), 1)
	out = strings.Replace(out, "BSu2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Sunday, 3)), 1)
	out = strings.Replace(out, "BFr3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Friday, 4)), 1)
	out = strings.Replace(out, "BSa3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Saturday, 4)), 1)
	out = strings.Replace(out, "BSu3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Sunday, 4)), 1)
	out = strings.Replace(out, "BFr4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Friday, 5)), 1)
	out = strings.Replace(out, "BSa4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Saturday, 5)), 1)
	out = strings.Replace(out, "BSu4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Sunday, 5)), 1)

	return out
}

func svgContentsSecC(m monthSec) string {

	if m.show == false {
		return ""
	}

	out := `
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="46.225574"
       y="145.79091"
       id="text1621"><tspan
         sodipodi:role="line"
         id="tspan1619"
         x="46.225574"
         y="145.79091"
         style="stroke-width:0.264583">MONTHC</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="2.54"
       y="150.37755"
       id="text1567"><tspan
         sodipodi:role="line"
         id="tspan1565"
         x="2.54"
         y="150.37755"
         style="stroke-width:0.264583">Mon</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="27.939999"
       y="150.37755"
       id="text1571"><tspan
         sodipodi:role="line"
         id="tspan1569"
         x="27.939999"
         y="150.37755"
         style="stroke-width:0.264583">Tue</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="53.34"
       y="150.37755"
       id="text1575"><tspan
         sodipodi:role="line"
         id="tspan1573"
         x="53.34"
         y="150.37755"
         style="stroke-width:0.264583">Wed</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="78.739998"
       y="150.37755"
       id="text1579"><tspan
         sodipodi:role="line"
         id="tspan1577"
         x="78.739998"
         y="150.37755"
         style="stroke-width:0.264583">Thu</tspan></text>
    <g
       id="group_rect_C"
       transform="translate(0,139.17367)">
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 2.5399999,12.088617 H 104.14 l 1e-5,127.000003 H 2.5399999 Z"
         id="path1907" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 27.94,12.088617 -3e-6,127.000003"
         id="path1909" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 53.339999,12.088617 V 139.08862"
         id="path1911" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 78.739996,12.088617 8e-6,127.000003"
         id="path1913" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 2.5399999,37.488608 104.14,37.488614"
         id="path1915" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 2.5400023,62.888613 104.14,62.888611"
         id="path1917" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 2.5400023,88.28861 H 104.14"
         id="path1919" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 2.5399985,113.68861 H 104.14"
         id="path1921" />
    </g>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="23.118895"
       y="154.20056"
       id="text1667"><tspan
         sodipodi:role="line"
         id="tspan1665"
         x="23.118895"
         y="154.20056"
         style="stroke-width:0.264583">CMo0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="48.518917"
       y="154.20056"
       id="text1671"><tspan
         sodipodi:role="line"
         x="48.518917"
         y="154.20056"
         style="stroke-width:0.264583"
         id="tspan1669">CTu0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="73.917442"
       y="154.20056"
       id="text1675"><tspan
         sodipodi:role="line"
         id="tspan1673"
         x="73.917442"
         y="154.20056"
         style="stroke-width:0.264583">CWe0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="99.317444"
       y="154.20056"
       id="text1679"><tspan
         sodipodi:role="line"
         id="tspan1677"
         x="99.317444"
         y="154.20056"
         style="stroke-width:0.264583">CTh0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="23.118895"
       y="179.60118"
       id="text1695"><tspan
         sodipodi:role="line"
         id="tspan1693"
         x="23.118895"
         y="179.60118"
         style="stroke-width:0.264583">CMo1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="48.518917"
       y="179.60118"
       id="text1699"><tspan
         sodipodi:role="line"
         x="48.518917"
         y="179.60118"
         style="stroke-width:0.264583"
         id="tspan1697">CTu1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="73.917442"
       y="179.60118"
       id="text1703"><tspan
         sodipodi:role="line"
         id="tspan1701"
         x="73.917442"
         y="179.60118"
         style="stroke-width:0.264583">CWe1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="99.317444"
       y="179.60118"
       id="text1707"><tspan
         sodipodi:role="line"
         id="tspan1705"
         x="99.317444"
         y="179.60118"
         style="stroke-width:0.264583">CTh1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="23.118895"
       y="205.00084"
       id="text1723"><tspan
         sodipodi:role="line"
         id="tspan1721"
         x="23.118895"
         y="205.00084"
         style="stroke-width:0.264583">CMo2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="48.518917"
       y="205.00084"
       id="text1727"><tspan
         sodipodi:role="line"
         x="48.518917"
         y="205.00084"
         style="stroke-width:0.264583"
         id="tspan1725">CTu2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="73.917442"
       y="205.00084"
       id="text1731"><tspan
         sodipodi:role="line"
         id="tspan1729"
         x="73.917442"
         y="205.00084"
         style="stroke-width:0.264583">CWe2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="99.317444"
       y="205.00084"
       id="text1735"><tspan
         sodipodi:role="line"
         id="tspan1733"
         x="99.317444"
         y="205.00084"
         style="stroke-width:0.264583">CTh2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="23.118895"
       y="230.40048"
       id="text1751"><tspan
         sodipodi:role="line"
         id="tspan1749"
         x="23.118895"
         y="230.40048"
         style="stroke-width:0.264583">CMo3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="48.518917"
       y="230.40048"
       id="text1755"><tspan
         sodipodi:role="line"
         x="48.518917"
         y="230.40048"
         style="stroke-width:0.264583"
         id="tspan1753">CTu3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="73.917442"
       y="230.40048"
       id="text1759"><tspan
         sodipodi:role="line"
         id="tspan1757"
         x="73.917442"
         y="230.40048"
         style="stroke-width:0.264583">CWe3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="99.317444"
       y="230.40048"
       id="text1763"><tspan
         sodipodi:role="line"
         id="tspan1761"
         x="99.317444"
         y="230.40048"
         style="stroke-width:0.264583">CTh3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="23.118895"
       y="255.80016"
       id="text1779"><tspan
         sodipodi:role="line"
         id="tspan1777"
         x="23.118895"
         y="255.80016"
         style="stroke-width:0.264583">CMo4</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="48.518917"
       y="255.80016"
       id="text1783"><tspan
         sodipodi:role="line"
         x="48.518917"
         y="255.80016"
         style="stroke-width:0.264583"
         id="tspan1781">CTu4</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="73.917442"
       y="255.80016"
       id="text1787"><tspan
         sodipodi:role="line"
         id="tspan1785"
         x="73.917442"
         y="255.80016"
         style="stroke-width:0.264583">CWe4</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="99.317444"
       y="255.80016"
       id="text1791"><tspan
         sodipodi:role="line"
         id="tspan1789"
         x="99.317444"
         y="255.80016"
         style="stroke-width:0.264583">CTh4</tspan></text>
	`

	mn, mt := monthNameTimeConst(m)
	out = strings.Replace(out, "MONTHC", mn, 1)
	out = strings.Replace(out, "CMo0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Monday, 1)), 1)
	out = strings.Replace(out, "CTu0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Tuesday, 1)), 1)
	out = strings.Replace(out, "CWe0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Wednesday, 1)), 1)
	out = strings.Replace(out, "CTh0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Thursday, 1)), 1)
	out = strings.Replace(out, "CMo1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Monday, 2)), 1)
	out = strings.Replace(out, "CTu1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Tuesday, 2)), 1)
	out = strings.Replace(out, "CWe1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Wednesday, 2)), 1)
	out = strings.Replace(out, "CTh1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Thursday, 2)), 1)
	out = strings.Replace(out, "CMo2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Monday, 3)), 1)
	out = strings.Replace(out, "CTu2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Tuesday, 3)), 1)
	out = strings.Replace(out, "CWe2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Wednesday, 3)), 1)
	out = strings.Replace(out, "CTh2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Thursday, 3)), 1)
	out = strings.Replace(out, "CMo3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Monday, 4)), 1)
	out = strings.Replace(out, "CTu3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Tuesday, 4)), 1)
	out = strings.Replace(out, "CWe3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Wednesday, 4)), 1)
	out = strings.Replace(out, "CTh3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Thursday, 4)), 1)
	out = strings.Replace(out, "CMo4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Monday, 5)), 1)
	out = strings.Replace(out, "CTu4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Tuesday, 5)), 1)
	out = strings.Replace(out, "CWe4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Wednesday, 5)), 1)
	out = strings.Replace(out, "CTh4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Thursday, 5)), 1)

	return out
}

func svgContentsSecD(m monthSec) string {

	if m.show == false {
		return ""
	}

	out := `
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="112.07753"
       y="150.37744"
       id="text1627"><tspan
         sodipodi:role="line"
         id="tspan1625"
         x="112.07753"
         y="150.37744"
         style="stroke-width:0.264583">Fri</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="137.47739"
       y="150.37744"
       id="text1631"><tspan
         sodipodi:role="line"
         id="tspan1629"
         x="137.47739"
         y="150.37744"
         style="stroke-width:0.264583">Sat</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:4.23333px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="162.87704"
       y="150.37744"
       id="text1635"><tspan
         sodipodi:role="line"
         id="tspan1633"
         x="162.87704"
         y="150.37744"
         style="stroke-width:0.264583">Sun</tspan></text>
    <g
       id="group_rect_D"
       transform="translate(0,139.17373)">
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 112.07753,12.088558 h 76.19951 V 139.08938 h -76.19951 z"
         id="path1954" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 137.47738,12.088558 V 139.08938"
         id="path1956" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="M 162.87738,12.088558 V 139.08938"
         id="path1958" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 112.07753,37.488555 h 76.19951"
         id="path1960" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 112.07753,62.888563 h 76.19951"
         id="path1962" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 112.07753,88.288999 76.19951,-2.64e-4"
         id="path1964" />
      <path
         style="fill:none;stroke:#000000;stroke-width:0.264583px;stroke-linecap:butt;stroke-linejoin:miter;stroke-opacity:1"
         d="m 112.07753,113.68939 76.19951,-2.6e-4"
         id="path1966" />
    </g>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="132.65486"
       y="154.20056"
       id="text1683"><tspan
         sodipodi:role="line"
         id="tspan1681"
         x="132.65486"
         y="154.20056"
         style="stroke-width:0.264583">DFr0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="158.04651"
       y="154.20056"
       id="text1687"><tspan
         sodipodi:role="line"
         id="tspan1685"
         x="158.04651"
         y="154.20056"
         style="stroke-width:0.264583">DSa0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="183.4465"
       y="154.20056"
       id="text1691"><tspan
         sodipodi:role="line"
         id="tspan1689"
         x="183.4465"
         y="154.20056"
         style="stroke-width:0.264583">DSu0</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="132.65486"
       y="179.60118"
       id="text1711"><tspan
         sodipodi:role="line"
         id="tspan1709"
         x="132.65486"
         y="179.60118"
         style="stroke-width:0.264583">DFr1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="158.04651"
       y="179.60118"
       id="text1715"><tspan
         sodipodi:role="line"
         id="tspan1713"
         x="158.04651"
         y="179.60118"
         style="stroke-width:0.264583">DSa1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="183.4465"
       y="179.60118"
       id="text1719"><tspan
         sodipodi:role="line"
         id="tspan1717"
         x="183.4465"
         y="179.60118"
         style="stroke-width:0.264583">DSu1</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="132.65486"
       y="205.00084"
       id="text1739"><tspan
         sodipodi:role="line"
         id="tspan1737"
         x="132.65486"
         y="205.00084"
         style="stroke-width:0.264583">DFr2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="158.04651"
       y="205.00084"
       id="text1743"><tspan
         sodipodi:role="line"
         id="tspan1741"
         x="158.04651"
         y="205.00084"
         style="stroke-width:0.264583">DSa2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="183.4465"
       y="205.00084"
       id="text1747"><tspan
         sodipodi:role="line"
         id="tspan1745"
         x="183.4465"
         y="205.00084"
         style="stroke-width:0.264583">DSu2</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="132.65486"
       y="230.40048"
       id="text1767"><tspan
         sodipodi:role="line"
         id="tspan1765"
         x="132.65486"
         y="230.40048"
         style="stroke-width:0.264583">DFr3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="158.04651"
       y="230.40048"
       id="text1771"><tspan
         sodipodi:role="line"
         id="tspan1769"
         x="158.04651"
         y="230.40048"
         style="stroke-width:0.264583">DSa3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="183.4465"
       y="230.40048"
       id="text1775"><tspan
         sodipodi:role="line"
         id="tspan1773"
         x="183.4465"
         y="230.40048"
         style="stroke-width:0.264583">DSu3</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="132.65486"
       y="255.80016"
       id="text1795"><tspan
         sodipodi:role="line"
         id="tspan1793"
         x="132.65486"
         y="255.80016"
         style="stroke-width:0.264583">DFr4</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="158.04651"
       y="255.80016"
       id="text1799"><tspan
         sodipodi:role="line"
         id="tspan1797"
         x="158.04651"
         y="255.80016"
         style="stroke-width:0.264583">DSa4</tspan></text>
    <text
       xml:space="preserve"
       style="font-size:3.52777px;line-height:1.25;font-family:Courier;-inkscape-font-specification:Courier;stroke-width:0.264583"
       x="183.4465"
       y="255.80016"
       id="text1803"><tspan
         sodipodi:role="line"
         id="tspan1801"
         x="183.4465"
         y="255.80016"
         style="stroke-width:0.264583">DSu4</tspan></text>
	`

	_, mt := monthNameTimeConst(m)
	out = strings.Replace(out, "DFr0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Friday, 1)), 1)
	out = strings.Replace(out, "DSa0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Saturday, 1)), 1)
	out = strings.Replace(out, "DSu0", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Sunday, 1)), 1)
	out = strings.Replace(out, "DFr1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Friday, 2)), 1)
	out = strings.Replace(out, "DSa1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Saturday, 2)), 1)
	out = strings.Replace(out, "DSu1", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Sunday, 2)), 1)
	out = strings.Replace(out, "DFr2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Friday, 3)), 1)
	out = strings.Replace(out, "DSa2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Saturday, 3)), 1)
	out = strings.Replace(out, "DSu2", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Sunday, 3)), 1)
	out = strings.Replace(out, "DFr3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Friday, 4)), 1)
	out = strings.Replace(out, "DSa3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Saturday, 4)), 1)
	out = strings.Replace(out, "DSu3", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Sunday, 4)), 1)
	out = strings.Replace(out, "DFr4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Friday, 5)), 1)
	out = strings.Replace(out, "DSa4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Saturday, 5)), 1)
	out = strings.Replace(out, "DSu4", fmt.Sprintf("%v", GetDayForNthWeekday(m.year, mt, time.Sunday, 5)), 1)

	return out
}
