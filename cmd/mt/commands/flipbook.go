package commands

import (
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/cobra"
)

// speedy todolists
var (
	FlipBook = &cobra.Command{
		Use:   "flipbook [book-pages] [repeat] [GIF]",
		Short: "calendar utilities",
		RunE:  FlipBookCmd,
	}
)

func init() {
	RootCmd.AddCommand(FlipBook)
}

func FlipBookCmd(cmd *cobra.Command, args []string) error {

	noPages, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	// repeat the gif images
	repeat, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	// get number of pages to create
	pdfPages := noPages / 9
	if noPages%9 != 0 {
		pdfPages++
	}

	pdf := gofpdf.New("P", "in", "Letter", "")
	pdf.SetMargins(0, 0, 0)

	// create the pdf pages
	for i := 0; i <= pdfPages; i++ {
		pdf.AddPage()
		AddPageCutMarks2(pdf)
	}

	var imgPaths []string

	gifPath := args[2]
	gifBase := path.Base(gifPath)
	giffile, err := os.Open(gifPath)
	if err != nil {
		return err
	}

	dir := strings.TrimSuffix(gifPath, path.Ext(gifPath)) + "_split"
	os.MkdirAll(dir, os.ModePerm)
	imgWidth, imgHeight, err := SplitAnimatedGIF(giffile, dir)
	if err != nil {
		return err
	}

	dirFiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range dirFiles {
		name := f.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if f.IsDir() {
			continue
		}
		imgPaths = append(imgPaths, path.Join(dir, name))
	}

	// repeat the entire series of images based on input
	var imgPathsRepeats []string // image paths with dups
	for i := 0; i < repeat; i++ {
		for j := 0; j < len(imgPaths); j++ {
			imgPathsRepeats = append(imgPathsRepeats, imgPaths[j])
		}
	}

	// duplicate images to correct number of frames
	noDups := noPages / len(imgPathsRepeats)
	var imgPathsDups []string // image paths with dups
	for i := 0; i < len(imgPathsRepeats); i++ {
		// TODO implement fading
		for j := 0; j < noDups; j++ {
			imgPathsDups = append(imgPathsDups, imgPathsRepeats[i])
		}
	}

	var opt gofpdf.ImageOptions
	for i := 0; i <= pdfPages; i++ {
		pdf.SetPage(i + 1)

		for cellY := 0; cellY < 3; cellY++ {
			for cellX := 0; cellX < 3; cellX++ {

				imgIndex := i + (pdfPages+1)*cellX + 3*(pdfPages+1)*cellY
				if imgIndex >= len(imgPathsDups) {
					continue
				}
				imgPath := imgPathsDups[imgIndex]
				scaledHeight := (8.5/3.0 - 0.6) * float64(imgHeight) / float64(imgWidth)
				yPosition := (float64(cellY+1)*11)/3.0 - scaledHeight - 0.3
				pdf.ImageOptions(imgPath, float64(cellX)*8.5/3.0+0.3, yPosition, 8.5/3.0-0.6, 0, false, opt, 0, "")
			}
		}
	}

	err = pdf.OutputFileAndClose(fmt.Sprintf("%v_flipbook.pdf", gifBase))
	if err != nil {
		return err
	}

	// delete the working folder
	os.RemoveAll(dir)

	return nil
	//// print the file
	//command1 := fmt.Sprintf("lp temp.pdf")
	//output1, err := cmn.Execute(command1)
	//fmt.Printf("%v\n%v\n", command1, output1)
	//if err != nil {
	//return err
	//}

	//// remove the temp file
	//return os.Remove("temp.pdf")
}

// write cut marks
func AddPageCutMarks2(pdf *gofpdf.Fpdf) {

	// -
	pdf.Line(0, (float64(11) / 3), 0.5, (float64(11) / 3))         // top-left
	pdf.Line(8, (float64(11) / 3), 8.5, (float64(11) / 3))         // top-right
	pdf.Line(0, (2 * float64(11) / 3), 0.5, (2 * float64(11) / 3)) // lower-left
	pdf.Line(8, (2 * float64(11) / 3), 8.5, (2 * float64(11) / 3)) // lower-right

	// |
	pdf.Line(8.5/3, 0, 8.5/3, 0.5)       // top-left
	pdf.Line(2*8.5/3, 0, 2*8.5/3, 0.5)   // top-right
	pdf.Line(8.5/3, 10.5, 8.5/3, 11)     // lower-left
	pdf.Line(2*8.5/3, 10.5, 2*8.5/3, 11) // lower-right

	// +
	pdf.Line((8.5/3 - 0.5), (float64(11) / 3), (8.5/3 + 0.5), (float64(11) / 3))             // top-left horizontal
	pdf.Line((8.5 / 3), (float64(11)/3 + 0.5), (8.5 / 3), (float64(11)/3 - 0.5))             // top-left vertical
	pdf.Line((2*8.5/3 - 0.5), (float64(11) / 3), (2*8.5/3 + 0.5), (float64(11) / 3))         // top-right horizontal
	pdf.Line((2 * 8.5 / 3), (float64(11)/3 + 0.5), (2 * 8.5 / 3), (float64(11)/3 - 0.5))     // top-right vertical
	pdf.Line((8.5/3 - 0.5), (2 * float64(11) / 3), (8.5/3 + 0.5), (2 * float64(11) / 3))     // lower-left horizontal
	pdf.Line((8.5 / 3), (2*float64(11)/3 + 0.5), (8.5 / 3), (2*float64(11)/3 - 0.5))         // lower-left vertical
	pdf.Line((2*8.5/3 - 0.5), (2 * float64(11) / 3), (2*8.5/3 + 0.5), (2 * float64(11) / 3)) // lower-right horizontal
	pdf.Line((2 * 8.5 / 3), (2*float64(11)/3 + 0.5), (2 * 8.5 / 3), (2*float64(11)/3 - 0.5)) // lower-right vertical
}

// Decode reads and analyzes the given reader as a GIF image
func SplitAnimatedGIF(reader io.Reader, writePath string) (imgWidth, imgHeight int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error while decoding: %s", r)
		}
	}()

	gif, err := gif.DecodeAll(reader)
	if err != nil {
		return 0, 0, err
	}

	imgWidth, imgHeight = getGifDimensions(gif)

	overpaintImage := image.NewRGBA(image.Rect(0, 0, imgWidth, imgHeight))
	draw.Draw(overpaintImage, overpaintImage.Bounds(), gif.Image[0], image.ZP, draw.Src)

	for i, srcImg := range gif.Image {
		draw.Draw(overpaintImage, overpaintImage.Bounds(), srcImg, image.ZP, draw.Over)

		// save current frame "stack". This will overwrite an existing file with that name
		file, err := os.Create(path.Join(writePath, strconv.Itoa(i)+".png"))
		if err != nil {
			return 0, 0, err
		}

		err = png.Encode(file, overpaintImage)
		if err != nil {
			return 0, 0, err
		}

		file.Close()
	}

	return imgWidth, imgHeight, nil
}

func getGifDimensions(gif *gif.GIF) (x, y int) {
	var lowestX int
	var lowestY int
	var highestX int
	var highestY int

	for _, img := range gif.Image {
		if img.Rect.Min.X < lowestX {
			lowestX = img.Rect.Min.X
		}
		if img.Rect.Min.Y < lowestY {
			lowestY = img.Rect.Min.Y
		}
		if img.Rect.Max.X > highestX {
			highestX = img.Rect.Max.X
		}
		if img.Rect.Max.Y > highestY {
			highestY = img.Rect.Max.Y
		}
	}

	return highestX - lowestX, highestY - lowestY
}
