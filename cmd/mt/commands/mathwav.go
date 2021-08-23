package commands

import (
	"math"
	"os"

	"github.com/spf13/cobra"
	wav "github.com/youpy/go-wav"
)

// file commands
var (
	MathWavCmd = &cobra.Command{
		Use:   "mathwav",
		Short: "generate a wav file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outfile, err := os.OpenFile("./mathaudio.wav", os.O_RDWR|os.O_CREATE, 0600)
			if err != nil {
				return err
			}

			defer func() {
				outfile.Close()
			}()

			var numSamples uint32 = 100000
			var numChannels uint16 = 1
			var sampleRate uint32 = 44100
			var bitsPerSample uint16 = 16

			writer := wav.NewWriter(outfile, numSamples, numChannels, sampleRate, bitsPerSample)
			samples := make([]wav.Sample, numSamples)

			//freq := 0.05
			amplitude := 300.0
			for i := 0; i < int(numSamples); i++ {
				// for 16 bit audio these values are each an int16
				// (min -32767, max 32767)
				//y := amplitude * math.Cos(freq*float64(i))
				//samples[i].Values[0] = int(int16(y))

				freq1 := 0.05 + 0.1*(float64(i)/float64(numSamples))
				y1 := amplitude * math.Cos(freq1*float64(i))

				freq2 := 0.07 + 0.07*(float64(i)/float64(numSamples))
				y2 := amplitude * math.Tan(freq2*float64(i))
				samples[i].Values[0] = int(int16(y1 + y2))
			}

			err = writer.WriteSamples(samples)
			if err != nil {
				return err
			}

			outfile.Close()
			return nil
		},
	}
)

func init() {
	RootCmd.AddCommand(MathWavCmd)
}
