package commands

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

// Lock2yamlCmd represents the lock2yaml command
var (
	ColourCmd = &cobra.Command{
		Use:   "colour",
		Short: "determine subtractive colour mixing",
		RunE:  colourCmd,
	}
)

func init() {
	RootCmd.AddCommand(
		ColourCmd,
	)
}

type RGB struct {
	R float64
	G float64
	B float64
}

func colourCmd(cmd *cobra.Command, args []string) error {

	inputs := []RGB{
		{10, 11, 230},
		{23, 222, 50},
		{240, 30, 30},
		{0, 0, 0},
		{256, 256, 256},
	}
	goal := RGB{123, 123, 123}
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
		fmt.Printf("fit: %v\ncolour: %v\nproportions: %v\n\n", key, calcMixedColour(inputs, proportion), proportion)

		i++
		if i > 5 {
			break
		}
	}
	return nil
}

// map[inputNo]proportions
type InputProportions []float64

//
func AccumulateRandResults(randSource int64, runDur time.Duration, inputs []RGB, goal RGB) (fits map[float64]InputProportions) {

	fits = make(map[float64]InputProportions)
	inputLen := len(inputs)
	rd := rand.New(rand.NewSource(randSource))

	timeStart := time.Now()
	for {

		var proportion InputProportions
		for j := 0; j < inputLen; j++ {
			proportion = append(proportion, rd.Float64())
		}

		fit := calcFit(inputs, proportion, goal)
		fits[fit] = proportion

		if time.Since(timeStart) > runDur {
			break
		}
	}

	return fits
}

func calcFit(inputs []RGB, proportion InputProportions, goal RGB) float64 {
	mixed := calcMixedColour(inputs, proportion)
	return math.Abs(mixed.R-goal.R) +
		math.Abs(mixed.G-goal.G) +
		math.Abs(mixed.B-goal.B)
}

// subtractive colour model get the resulting colour
func calcMixedColour(inputs []RGB, proportion InputProportions) (mixed RGB) {

	inputLen := len(inputs)
	var totalsR, totalsG, totalsB, totalAmt float64

	for i := 0; i < inputLen; i++ {
		totalAmt += proportion[i]
		totalsR += float64(inputs[i].R) * proportion[i]
		totalsG += float64(inputs[i].G) * proportion[i]
		totalsB += float64(inputs[i].B) * proportion[i]
	}

	mixed = RGB{
		totalsR / totalAmt,
		totalsG / totalAmt,
		totalsB / totalAmt,
	}
	return mixed
}
