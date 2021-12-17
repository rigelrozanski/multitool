package commands

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"strings"

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

				freq1 := 2.3 //+ 0.1*(float64(i)/float64(numSamples))
				y1 := amplitude * math.Sin(freq1*float64(i))
				y2 := amplitude * math.Cos(freq1*float64(i)/2) * math.Cos(freq1*float64(i))
				//y3 := amplitude * 1.0 / math.Pow(10.0, math.Sin(freq1*float64(i)/2))
				y3 := amplitude*math.Sin(freq1*float64(i)) + amplitude*freq1*math.Sin(float64(i)/freq1)

				_ = y1
				_ = y2
				_ = y3

				//freq2 := 0.07 + 0.07*(float64(i)/float64(numSamples))
				//y2 := amplitude * math.Tan(freq2*float64(i))
				samples[i].Values[0] = int(int16(y1))
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

// file commands
// intensity == saturation*Luminosity
//    physically speaking, we should probably turn
//    the remaining grey (1-saturation)*Luminosity
//    into white noise... but lets not right now
// hue == frequency
var (
	ImgWavCmd = &cobra.Command{
		Use:   "imgwav [file] [k-samples] [loops]",
		Short: "generate a wav file from an image",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {

			volumeMul := 1.0
			freqMin := 0.004 //+ 0.1*(float64(i)/float64(numSamples))
			freqMax := 2.3   // could be as high as 2.5

			samplesArg, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			loopsArg, err := strconv.Atoi(args[2])
			if err != nil {
				return err
			}

			imgFile, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer imgFile.Close()

			image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
			image.RegisterFormat("gif", "gif", gif.Decode, gif.DecodeConfig)
			image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

			// Consider using the general image.Decode as it can sniff and decode any registered image format.
			img, _, err := image.Decode(imgFile)
			if err != nil {
				log.Fatal(err)
			}

			type AF struct {
				ampl float64
				freq float64
			}

			var AFs []AF
			for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
				for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
					rI, gI, bI, _ := img.At(x, y).RGBA()

					// convert to floats with values between 0 and 1
					r, g, b := (float64(rI) / 65535), (float64(gI) / 65535), (float64(bI) / 65535)

					max := math.Max(math.Max(r, g), b)
					min := math.Min(math.Min(r, g), b)

					luminosity := (max + min) / 2
					saturation := 0.0
					diff := max - min
					if diff == 0 { // it's gray ignore this pixel
						continue
					}
					if luminosity < 0.5 { // it's not gray
						saturation = diff / (max + min)
					} else {
						saturation = diff / (2 - max - min)
					}
					audioIntensity := saturation * luminosity * volumeMul
					_ = audioIntensity

					// hue which will be between 0 and 1
					hue := 0.0
					r2 := (((max - r) / 6) + (diff / 2)) / diff
					g2 := (((max - g) / 6) + (diff / 2)) / diff
					b2 := (((max - b) / 6) + (diff / 2)) / diff
					switch {
					case r == max:
						hue = b2 - g2
					case g == max:
						hue = (1.0 / 3.0) + r2 - b2
					case b == max:
						hue = (2.0 / 3.0) + g2 - r2
					}
					switch { // fix wraparounds
					case hue < 0:
						hue += 1
					case hue > 1:
						hue -= 1
					}
					freq := freqMin + hue*(freqMax-freqMin)

					af := AF{
						ampl: audioIntensity,
						freq: freq,
					}
					AFs = append(AFs, af)
				}
			}
			fmt.Println("calculated AFs")

			///////////////////////////

			rootName := strings.Split(path.Base(args[0]), ".")[0]
			fp := fmt.Sprintf("./%v.wav", rootName)
			outfile, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE, 0600)
			if err != nil {
				return err
			}
			defer func() {
				outfile.Close()
			}()

			timeMultiplier := uint32(loopsArg)
			numSamples := uint32(samplesArg) * 1000
			var numChannels uint16 = 1
			var sampleRate uint32 = 44100
			var bitsPerSample uint16 = 16
			finalSamples := numSamples * timeMultiplier
			writer := wav.NewWriter(outfile, finalSamples, numChannels, sampleRate, bitsPerSample)
			samples := make([]wav.Sample, finalSamples)

			// saved samples
			ssamples := make([]int, numSamples)

			for i := 0; i < int(numSamples); i++ {
				if int(numSamples/10) != 0 && i%int(numSamples/10) == 0 {
					per := int(100 * (float64(i) / float64(numSamples)))
					fmt.Printf("sampling %v percent\n", per)
				}
				sum := 0.0
				for _, af := range AFs {
					sum += af.ampl * math.Sin(af.freq*float64(i))
				}
				ssamples[i] = int(int16(sum))
			}

			// write the samples with time multiplier
			ii := 0
			for n := 0; n < int(timeMultiplier); n++ {
				if int(timeMultiplier/10) != 0 && n%int(timeMultiplier/10) == 0 {
					per := int(100 * (float64(n) / float64(timeMultiplier)))
					fmt.Printf("writting %v percent\n", per)
				}
				for i := 0; i < int(numSamples); i++ {
					samples[ii].Values[0] = int(int16(ssamples[i]))
					ii++
				}
			}

			fmt.Printf("saving audio...\n")
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
	RootCmd.AddCommand(ImgWavCmd)
}
