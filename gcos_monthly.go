package main

// This script constructs monthly averages from daily temperature
// records for the GCOS surface network (GCOS GSN).
//
// This is the non-concurrent version of the script, see
// gcos_monthly_concurrent.go for the concurrent version.
//
// The data files are available from:
// ftp://ftp.ncdc.noaa.gov/pub/data/ghcn/daily/ghcnd_gsn.tar.gz
//
// Prior to running this script, the tarball should be fully extracted
// and the data files individually gzipped.  On a linux-like machine
// the following will download and prepare the data:
//
// wget ftp://ftp.ncdc.noaa.gov/pub/data/ghcn/daily/ghcnd_gsn.tar.gz
// gunzip ghcnd_gsn.tar.gz
// tar -xvf ghcnd_gsn.tar
// cd ghcnd_gsn
// gzip *
//
// The data_path and out_path variables below must be set to
// appropriate local directory paths.
//
// The data file format is available here:
// ftp://ftp.ncdc.noaa.gov/pub/data/ghcn/daily/readme.txt

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

var (
	// Location of the data in the local file system
	data_path = "/nfs/kshedden/GHCN/ghcnd_gsn"

	// Path where the output file is written
	out_path = "/nfs/kshedden/GHCN"

	// The temperature type to process, should be either "TMAX" or
	// "TMIN"
	eltype = "TMAX"

	// io.Writer for the output file
	wtr *gzip.Writer
)

// The data in one line of the input file
type lrec_t struct {
	Id      string    // The station id
	Year    int       // The year of the data point
	Month   int       // The month of the data point (1..12)
	Element string    // The data value type (TMAX or TMIN)
	Values  []float64 // The daily values
	IsValid []bool    // Validity flags for the data
}

// The summary record for one month
type mrec_t struct {
	Id     string  // The station id
	Year   int     // The year of the data point
	Month  int     // The month of the data point (1..12)
	Mean   float64 // The mean value
	Nvalid int     // The number of valid values in the mean
}

// Parse one line of a raw file and put the results into a structure.
func parse(line string) *lrec_t {

	var rec lrec_t
	var err error

	rec.Id = line[0:11]

	rec.Year, err = strconv.Atoi(line[11:15])
	if err != nil {
		panic(err)
	}

	rec.Month, err = strconv.Atoi(line[15:17])
	if err != nil {
		panic(err)
	}

	// Read all the daily temperature values.  See data format
	// document for parsing details
	for pos := 21; pos < len(line); pos += 8 {

		// First check the quality flag
		if line[26] != ' ' {
			rec.Values = append(rec.Values, 0)
			rec.IsValid = append(rec.IsValid, false)
			continue
		}

		sval := strings.TrimLeft(line[pos:pos+5], " ")
		v, err := strconv.ParseFloat(sval, 64)
		if err != nil {
			panic(err)
		}
		rec.Values = append(rec.Values, v)
		rec.IsValid = append(rec.IsValid, true)
	}

	return &rec
}

// Convert a raw data record into a monthly summary record
func summarize(lrec *lrec_t) *mrec_t {

	nvalid := 0
	sum := 0.0

	for j, x := range lrec.Values {
		if lrec.IsValid[j] && x != -9999 {
			sum += x
			nvalid++
		}
	}

	mean := sum / float64(nvalid)

	// In the raw data the temperature units are 0.1 degree C, we
	// want to convert to degrees C.
	mean /= 10

	mrec := &mrec_t{Id: lrec.Id,
		Year: lrec.Year, Month: lrec.Month, Nvalid: nvalid, Mean: mean}

	return mrec
}

// All processing for one data file (for one station)
func processFile(file os.FileInfo) {

	// A file reader for the input file
	fname := path.Join(data_path, file.Name())
	fid, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer fid.Close()

	// Wrap the file reader in a gzip reader
	rdr, err := gzip.NewReader(fid)
	if err != nil {
		panic(err)
	}
	defer rdr.Close()

	scanner := bufio.NewScanner(rdr)

	// Read the lines of the file
	for scanner.Scan() {

		line := scanner.Text()

		eltype_val := line[17:21]
		if eltype_val != eltype {
			continue
		}

		lrec := parse(line)
		mrec := summarize(lrec)

		outline := fmt.Sprintf("%s,%d,%d,%d,%.3f\n", mrec.Id, mrec.Year,
			mrec.Month, mrec.Nvalid, mrec.Mean)
		wtr.Write([]byte(outline))
	}
}

func main() {

	files, err := ioutil.ReadDir(data_path)
	if err != nil {
		panic(err)
	}

	// Create a file writer
	fname := fmt.Sprintf("gcos_monthly_%s.csv.gz", eltype)
	fname = path.Join(out_path, fname)
	oid, err := os.Create(fname)
	if err != nil {
		panic(err)
	}
	defer oid.Close()

	// Wrap the file output writer in a gzip writer.
	wtr = gzip.NewWriter(oid)
	defer wtr.Close()

	// Put a header into the output file
	header := "Id,Year,Month,Nvalid,Mean\n"
	wtr.Write([]byte(header))

	// Process each file
	for _, file := range files {
		processFile(file)
	}
}
