package commands

import (
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/spf13/cobra"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

// Lock2yamlCmd represents the lock2yaml command
var (
	PDFCmd = &cobra.Command{
		Use:   "pdf",
		Short: "reorder pdf pages for booklet printing",
	}
	DoublePDFCmd = &cobra.Command{
		Use:   "double [pdf-file]",
		Short: "double all the pages in a pdf file",
		RunE:  doubleCmd,
		Args:  cobra.ExactArgs(1),
	}
	BookPDFCmd = &cobra.Command{
		Use:   "book [pdf-file]",
		Short: "first half on right, last half on left",
		RunE:  bookCmd,
		Args:  cobra.ExactArgs(1),
	}
	AltBookPDFCmd = &cobra.Command{
		Use:   "alt-book [pdf-file]",
		Short: "first half on right, last half on left",
		RunE:  altBookCmd,
		Args:  cobra.ExactArgs(1),
	}
)

func init() {
	PDFCmd.AddCommand(
		DoublePDFCmd,
		BookPDFCmd,
		AltBookPDFCmd,
	)
	RootCmd.AddCommand(
		PDFCmd,
	)
}

func extract(args []string) (
	inFile, tempDir string, pgCount int, config *pdfcpu.Configuration, err error) {

	config = pdfcpu.NewDefaultConfiguration()

	inFile = args[0]
	if path.Ext(inFile) != ".pdf" {
		return "", "", 0, config, errors.New("not a pdf file")
	}

	pgCount, err = api.PageCountFile(inFile)
	if err != nil {
		return "", "", 0, config, err
	}

	tempDir, err = ioutil.TempDir("", "_mt_pdf_combine")
	if err != nil {
		return "", "", 0, config, err
	}

	// extract all pdf pages to single pages
	var allPgs []string
	for i := 0; i < pgCount; i++ {
		allPgs = append(allPgs, strconv.Itoa(i+1))
	}
	for i := 0; i < pgCount; i++ {
		outFile := path.Join(tempDir, strconv.Itoa(i+1)+".pdf")
		rmPgs := make([]string, len(allPgs))
		copy(rmPgs, allPgs)
		rmPgs = append(rmPgs[:i], rmPgs[i+1:]...)
		api.RemovePagesFile(inFile, outFile, rmPgs, config)
	}

	return inFile, tempDir, pgCount, config, nil
}

func doubleCmd(cmd *cobra.Command, args []string) error {

	inFile, tempDir, pgCount, config, err := extract(args)
	if err != nil {
		return err
	}

	// double each page
	var orderedFiles []string
	for i := 0; i < pgCount; i++ {
		inFile := path.Join(tempDir, strconv.Itoa(i+1)+".pdf")
		orderedFiles = append(orderedFiles, inFile, inFile)
	}

	combinedFile := path.Dir(inFile) + "/" + strings.Split(path.Base(inFile), ".")[0] + "_doubled.pdf"
	api.MergeFile(orderedFiles, combinedFile, config)
	fmt.Printf("new file created at: %s\n", combinedFile)

	os.RemoveAll(tempDir)
	return nil
}

func bookCmd(cmd *cobra.Command, args []string) error {

	inFile, tempDir, pgCount, config, err := extract(args)
	if err != nil {
		return err
	}

	// add extra padding files
	modPg := pgCount % 4
	if modPg != 0 {
		numInserts := 4 - modPg
		rectFile := path.Join(tempDir, "rect.pdf")
		api.InsertPagesFile(inFile, rectFile, []string{"1"}, config)

		var rmPgs []string
		for i := 1; i < pgCount+1; i++ {
			rmPgs = append(rmPgs, strconv.Itoa(i+1))
		}
		for i := 0; i < numInserts; i++ {
			outFile := path.Join(tempDir, strconv.Itoa(pgCount+1+i)+".pdf")
			api.RemovePagesFile(rectFile, outFile, rmPgs, config)
		}
		pgCount += numInserts
	}

	midPage := pgCount/2 + 1
	fmt.Printf("debug modPg: %v\n", modPg)

	var orderedFiles []string
	for i := 0; midPage+i < pgCount; i += 2 {
		inFile1 := path.Join(tempDir, strconv.Itoa(i+1)+".pdf")
		inFile2 := path.Join(tempDir, strconv.Itoa(midPage+i)+".pdf")
		inFile3 := path.Join(tempDir, strconv.Itoa(midPage+i+1)+".pdf")
		inFile4 := path.Join(tempDir, strconv.Itoa(i+2)+".pdf")
		orderedFiles = append(orderedFiles, inFile1, inFile2, inFile3, inFile4)
	}

	combinedFile := path.Dir(inFile) + "/" + strings.Split(path.Base(inFile), ".")[0] + "_reordered.pdf"
	api.MergeFile(orderedFiles, combinedFile, config)
	fmt.Printf("new file created at: %s\n", combinedFile)

	os.RemoveAll(tempDir)
	return nil
}

func altBookCmd(cmd *cobra.Command, args []string) error {

	//inFile := args[0]
	//if path.Ext(inFile) != ".pdf" {
	//return errors.New("not a pdf file")
	//}

	dir := args[0]
	dirFiles, err := ioutil.ReadDir(args[0])
	if err != nil {
		return err
	}

	var imgPaths []string

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

	// get number of pages to create
	li := len(imgPaths)
	for ; li%4 != 0; li++ {
		imgPaths = append(imgPaths, "")
	}
	pdfPages := li / 2

	pdf := gofpdf.New("L", "in", "Letter", "")
	pdf.SetMargins(0, 0, 0)

	// create the pdf pages
	for i := 0; i <= pdfPages; i++ {
		pdf.AddPage()
	}

	var opt gofpdf.ImageOptions

	// get the im to read height/width ratio
	// all images should be the same dimentions
	reader, err := os.Open(imgPaths[0])
	defer reader.Close()
	if err != nil {
		return err
	}
	im, _, err := image.DecodeConfig(reader)
	if err != nil {
		return err
	}

	// process a whole sheet front and back at once
	yMargin := float64(0.3)
	scaledWidth := (8.5 - 2*yMargin) * float64(im.Width) / float64(im.Height)
	xPositionLeft := (float64(11)/2.0 - scaledWidth) / 2
	xPositionRight := float64(11)/2.0 + xPositionLeft
	if xPositionLeft < 0 {
		panic("bad borders")
	}

	imgIndex := 0
	for i := 0; i < pdfPages; i += 2 {

		// front left (pg1)
		pdf.SetPage(i + 1)
		if imgIndex >= len(imgPaths) {
			continue
		}
		imgPath := imgPaths[imgIndex]
		if imgPath != "" {
			pdf.ImageOptions(imgPath, xPositionLeft, yMargin, 0, 8.5-2*yMargin, false, opt, 0, "")
		}
		imgIndex++

		// back right (pg2)
		pdf.SetPage(i + 2)
		if imgIndex >= len(imgPaths) {
			continue
		}
		imgPath = imgPaths[imgIndex]
		if imgPath != "" {
			pdf.ImageOptions(imgPath, xPositionRight, yMargin, 0, 8.5-2*yMargin, false, opt, 0, "")
		}
		imgIndex++
	}

	for i := 0; i < pdfPages; i += 2 {

		// front right (pgMID)
		pdf.SetPage(i + 1)
		if imgIndex >= len(imgPaths) {
			continue
		}
		imgPath := imgPaths[imgIndex]
		if imgPath != "" {
			pdf.ImageOptions(imgPath, xPositionRight, yMargin, 0, 8.5-2*yMargin, false, opt, 0, "")
		}
		imgIndex++

		// back left (pgMID+1)
		pdf.SetPage(i + 2)
		if imgIndex >= len(imgPaths) {
			continue
		}
		imgPath = imgPaths[imgIndex]
		if imgPath != "" {
			pdf.ImageOptions(imgPath, xPositionLeft, yMargin, 0, 8.5-2*yMargin, false, opt, 0, "")
		}
		imgIndex++
	}

	err = pdf.OutputFileAndClose(fmt.Sprintf("%v_printable_book.pdf", strings.TrimSuffix(dir, "/")))
	if err != nil {
		return err
	}

	// TODO
	//// delete the working folder
	//os.RemoveAll(dir)

	return nil
}
