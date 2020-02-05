package commands

import (
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/rigelrozanski/common/colour"
	"github.com/rigelrozanski/thranch/quac"
)

const (
	caliSetStartX      = 100 // calibration set-x start
	caliSetEndX        = 120 // calibration set-x end
	caliSearchMinY     = 5   // calibration search y start
	thick              = 30  // pixels down and across to check for calibration and parsing
	acrylicPaintGperMl = 1.2

	// Allowable variance per R,G,or B, up or down from the
	// calibration variance to be considered that colour
	variance = 20 * 257
)

// Lock2yamlCmd represents the lock2yaml command
var (
	ColourCmd = &cobra.Command{
		Use:   "colour <filepath-to-image> [run-durations-seconds] [final-mix-volume-ml]",
		Short: "determine subtractive colour mixing, provided image must be completely the desired colour",
		RunE:  mixColourCmd,
		Args:  cobra.MaximumNArgs(3),
	}
)

func init() {
	RootCmd.AddCommand(
		ColourCmd,
	)
}

func getAvgColourFromFile(filpath string) colour.FRGB {
	file, err := os.Open(filpath)
	if err != nil {
		log.Fatal("Error: Image could not be decoded")
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	bounds := img.Bounds()
	return colour.LoadColours(0, bounds.Max.X, 0, bounds.Max.Y, img).AvgColour().ToFRGB()
}

func mixColourCmd(cmd *cobra.Command, args []string) error {

	goal := getAvgColourFromFile(args[0])

	runduration := 5
	mixMl := float64(50)

	if len(args) >= 2 {
		var err error
		runduration, err = strconv.Atoi(args[1])
		if err != nil {
			return err
		}
	}
	if len(args) >= 3 {
		ml, err := strconv.Atoi(args[2])
		if err != nil {
			return err
		}
		mixMl = float64(ml)
	}

	var inputNames []string
	var inputs []colour.FRGB
	quac.Initialize(os.ExpandEnv("$HOME/.thranch_config"))
	imagesIdeas := quac.GetImagesByTag([]string{"paints"})
	for i, idea := range imagesIdeas {
		input := getAvgColourFromFile(idea.Path())
		inputs = append(inputs, input)

		input.PrintColour("")
		colourName, _ := idea.GetTaggedValue("colour")
		inputNames = append(inputNames, colourName)
		fmt.Printf("input No %v, name: %v\n", i, colourName)
	}

	goal.PrintColour("")
	fmt.Printf("goal colour %v\n", goal)

	accRes := AccumulateRandResults(42, time.Duration(runduration)*time.Second, inputs, goal)

	// sort by keys
	keys := make([]float64, 0, len(accRes))
	for key := range accRes {
		keys = append(keys, key)
	}
	sort.Float64s(keys)

	// display top results
	i := 0
	for _, key := range keys {
		var cumulative, totalPortion float64
		proportion := accRes[key]
		col := colour.CalcMixedColour(inputs, proportion)
		col.PrintColour("")
		fmt.Printf("fit: %.2f\ncolour: %.0f\n proportions for %vml mixed paint:", key/257, col, mixMl)

		// calculate the totalPortion
		for j := 0; j < len(proportion); j++ {
			totalPortion += proportion[j]
		}
		totalGrams := mixMl * acrylicPaintGperMl
		gramPerPortion := totalGrams / totalPortion

		for j := 0; j < len(proportion); j++ {
			cumulative += proportion[j] * gramPerPortion
			fmt.Printf("\n\t%v:   \t%.2fg \t%.2f g", inputNames[j], proportion[j]*gramPerPortion, cumulative)
		}
		fmt.Println("")

		i++
		if i > 2 {
			break
		}
	}
	return nil
}

func getCalibrationColours(caliSetStartX, caliSetEndX, caliSearchMinY, caliSearchMaxY, thick int, variance uint32, img image.Image,
) (scannedColours []colour.FRGB, err error) {

	col, outsideY, err := colour.GetCalibrationColour(caliSetStartX, caliSetEndX, caliSearchMinY, caliSearchMaxY, thick, variance, img)
	if err != nil {
		return scannedColours, err
	}
	scannedColours = append(scannedColours, col.ToFRGB())

	for {
		col, outsideY, err = colour.GetCalibrationColour(caliSetStartX, caliSetEndX, outsideY, caliSearchMaxY, thick, variance, img)
		if err != nil { // no more colours to scan
			break
		}
		scannedColours = append(scannedColours, col.ToFRGB())
	}

	return scannedColours, nil
}

// Get a bunch of random results
func AccumulateRandResults(randSource int64, runDur time.Duration, inputs []colour.FRGB, goal colour.FRGB) (fits map[float64]colour.InputProportions) {

	fits = make(map[float64]colour.InputProportions)
	inputLen := len(inputs)
	rd := rand.New(rand.NewSource(randSource))

	timeStart := time.Now()
	for {

		var proportion colour.InputProportions
		for j := 0; j < inputLen; j++ {
			proportion = append(proportion, rd.Float64())
		}

		fit := colour.CalcFit(inputs, proportion, goal)
		fits[fit] = proportion

		if time.Since(timeStart) > runDur {
			break
		}
	}

	return fits
}
