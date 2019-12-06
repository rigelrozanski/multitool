package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

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
	}
	BookPDFCmd = &cobra.Command{
		Use:   "book [pdf-file]",
		Short: "first half on right, last half on left",
		RunE:  bookCmd,
	}
)

func init() {
	PDFCmd.AddCommand(
		DoublePDFCmd,
		BookPDFCmd,
	)
	RootCmd.AddCommand(
		PDFCmd,
	)
}

func extract(args []string) (
	inFile, tempDir string, pgCount int, config *pdfcpu.Configuration, err error) {

	config = pdfcpu.NewDefaultConfiguration()

	if len(args) != 1 {
		return "", "", 0, config, errors.New("must provide path to pdf file")
	}
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

	midPage := pgCount/2 + modPg

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
