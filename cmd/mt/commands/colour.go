package commands

import (
	"fmt"
	"image"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/rigelrozanski/common/colour"
)

const (
	caliSetStartX  = 100 // calibration set-x start
	caliSetEndX    = 120 // calibration set-x end
	caliSearchMinY = 5   // calibration search y start
	thick          = 30  // pixels down and across to check for calibration and parsing

	// Allowable variance per R,G,or B, up or down from the
	// calibration variance to be considered that colour
	variance = 20 * 257
)

// Lock2yamlCmd represents the lock2yaml command
var (
	ColourCmd = &cobra.Command{
		Use:   "colour",
		Short: "determine subtractive colour mixing",
		RunE:  colourCmd,
		Args:  cobra.ExactArgs(1),
	}
)

func init() {
	RootCmd.AddCommand(
		ColourCmd,
	)
}

func colourCmd(cmd *cobra.Command, args []string) error {

	scanPath := args[0]

	file, err := os.Open(scanPath)
	if err != nil {
		log.Fatal("Error: Image could not be decoded")
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	bounds := img.Bounds()

	scans, err := getCalibrationColours(caliSetStartX, caliSetEndX, caliSearchMinY, bounds.Max.Y, thick, variance, img)
	if err != nil {
		return err
	}

	goal := scans[0]
	inputs := scans[1:]

	goal.PrintColour("")
	fmt.Printf("goal colour %v\n", goal)

	accRes := AccumulateRandResults(42, 5*time.Second, inputs, goal)

	// sort by keys
	keys := make([]float64, 0, len(accRes))
	for key := range accRes {
		keys = append(keys, key)
	}
	sort.Float64s(keys)

	// display top 5 results
	i := 0
	for _, key := range keys {
		proportion := accRes[key]
		col := colour.CalcMixedColour(inputs, proportion)
		col.PrintColour("")
		fmt.Printf("fit: %v\ncolour: %v\nproportions: %v\n\n", key, col, proportion)

		i++
		if i > 5 {
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
