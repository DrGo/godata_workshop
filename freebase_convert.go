package main

// This is a small script to convert an Excel file to csv format and
// save it as a gzip compressed file.
//
// Requires the xlsx library, to get it use:
//     go get https://github.com/tealeg/xlsx
//
// Adjust the file paths below as needed.

import (
	"compress/gzip"
	"encoding/csv"
	"os"
	"path"

	"github.com/tealeg/xlsx"
)

var (
	dpath string = "/nfs/kshedden/Freebase"
	fname string = "SchichDataS1_FB.xlsx"
)

func main() {
	// Open the Excel file for reading
	xlFile, err := xlsx.OpenFile(path.Join(dpath, fname))
	if err != nil {
		panic(err)
	}

	// Open the output file for writing
	outname := path.Join(dpath, "FB.csv.gz")
	out, err := os.Create(outname)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// Wrap the file writer in a gzip writer
	gout := gzip.NewWriter(out)
	defer gout.Close()

	// Wrap the gzip writer in a csv writer
	wtr := csv.NewWriter(gout)

	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			var x []string
			for _, cell := range row.Cells {
				x = append(x, cell.String())
			}
			wtr.Write(x)
		}
	}
}
