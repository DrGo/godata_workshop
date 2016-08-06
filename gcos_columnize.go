package main

// This script takes the text input file of GHCN (Global Historical
// Climatology Network) data and converts it to compressed binary
// column vectors, in long format by year.
//
// The raw data are in wide format by station month, see the data file
// format here:
//     ftp://ftp.ncdc.noaa.gov/pub/data/ghcn/daily/readme.txt
//
// We pivot the data to create files by year, combining data for all
// stations and all days for a single year.  The resulting data are
// stored in directories for each year, e.g.:
//
// 1909/
//      ids.gz
//      dates.gz
//      values.gz
//
// The three data files are aligned, i.e. ids[i], dates[i], values[i]
// correspond to an observation.  The data file formats are:
//
// ids.gz: the station identifiers as a newline (\n) delimited
//     sequence of text values
// dates.gz: the date of each observation, as a newline delimited
//     sequence of iso formatted dates (e.g. 1909-03-15) for March
//     15th 1909
// values.gz: the temperature values, as a stream of native float64
//     values
//
// The output files are sorted first by station then by date.
//
// The data_path and out_path variables below should be set to
// writeable directory paths in the file system.
//
// The script can be used to extract either daily maximum or daily
// minimum temperature values, by setting the "eltype" variable below
// to either "TMAX" or "TMIN" respectively.
//
// The script uses external libraries that can be obtained using:
//     go get github.com/kshedden/ziparray

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/kshedden/ziparray"
)

var (
	// Location of the data in the local file system
	data_path = "/nfs/kshedden/GHCN/ghcnd_gsn"

	// Path where the output file is written
	out_path = "/nfs/kshedden/GHCN_tmp"

	// The temperature type to process, should be either "TMAX" or
	// "TMIN"
	eltype = "TMAX"

	// A gob encoder for the data in each year
	year_gob map[int]*gob.Encoder

	// The backing data structure for each gob encoder.  We can't
	// connect to the files directly because we can't have
	// enough many open files.
	year_buf map[int]*bytes.Buffer

	// We don't know up-front which years appear in the file, so
	// we create data structures as we encounter each year for the
	// first time.  seen tells us whether we are seeing a year for
	// the first time and therefore need to set things up.
	seen map[int]bool

	wg sync.WaitGroup

	// Number of values to write to a channel before flushing to disk
	buf_size int = 1e7
)

// One data point
type rec_t struct {
	Id    string  // The station id
	Year  int     // The year of the data point
	Month int     // The month of the data point (1..12)
	Day   int     // The day within the month (1..31)
	Value float64 // The data value (TMAX or TMIN)
}

// We will need to sort slices of rec_t values
type recslice []rec_t

func (a recslice) Len() int {
	return len(a)
}
func (a recslice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a recslice) Less(i, j int) bool {
	if a[i].Id < a[j].Id {
		return true
	} else if a[i].Id > a[j].Id {
		return false
	}
	if a[i].Year < a[j].Year {
		return true
	} else if a[i].Year > a[j].Year {
		return false
	}
	if a[i].Month < a[j].Month {
		return true
	} else if a[i].Month > a[j].Month {
		return false
	}
	if a[i].Day < a[j].Day {
		return true
	} else if a[i].Day > a[j].Day {
		return false
	}
	return false
}

func setupYear(year int) {
	year_buf[year] = new(bytes.Buffer)
	year_gob[year] = gob.NewEncoder(year_buf[year])

	// Make sure the output path exists
	dname := path.Join(out_path, fmt.Sprintf("%d", year))
	err := os.MkdirAll(dname, 0700)
	if err != nil {
		panic(err)
	}

	// Reset the output file
	fn := tfileName(year)
	_, err = os.Create(fn)
	if err != nil {
		panic(err)
	}

	seen[year] = true
}

func parse(line string) {

	var err error

	id := line[0:11]

	year, err := strconv.Atoi(line[11:15])
	if err != nil {
		panic(err)
	}

	// If we have not seen this year before, need to set up.
	if !seen[year] {
		setupYear(year)
	}

	month, err := strconv.Atoi(line[15:17])
	if err != nil {
		panic(err)
	}

	// Read all the daily temperature values.  See data format
	// document for parsing details
	for pos := 21; pos < len(line); pos += 8 {

		// Skip if low quality
		if line[26] != ' ' {
			continue
		}

		sval := strings.TrimLeft(line[pos:pos+5], " ")
		v, err := strconv.ParseFloat(sval, 64)
		if err != nil {
			panic(err)
		}

		// This represents a missing value
		if v == -9999 {
			continue
		}

		// The raw data are in 0.1 degrees C, convert to degrees C
		v /= 10

		day := (pos-21)/8 + 1
		r := rec_t{Id: id, Year: year, Month: month, Day: day, Value: v}
		year_gob[year].Encode(r)

		// If the buffer is full, flush it
		if year_buf[year].Len() > buf_size {
			flush(year)
		}
	}
}

// All processing for one data file (for one station)
func processFile(file os.FileInfo) {

	fmt.Printf("Reading %v\n", file.Name())

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

		parse(line)
	}
}

// Returns the name of the temporary data file for each year
func tfileName(year int) string {
	return path.Join(out_path, fmt.Sprintf("%d", year), "raw.bin")
}

func flush(year int) {

	fmt.Printf("Flushing %d\n", year)
	fname := tfileName(year)
	fid, err := os.OpenFile(fname, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer fid.Close()

	_, err = fid.Write(year_buf[year].Bytes())
	if err != nil {
		panic(err)
	}
	year_buf[year].Truncate(0)
}

func step1() {

	year_buf = make(map[int]*bytes.Buffer)
	year_gob = make(map[int]*gob.Encoder)
	seen = make(map[int]bool)

	// Reset the output directory
	err := os.RemoveAll(out_path)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(out_path, 0700)
	if err != nil {
		panic(err)
	}

	// Get a list of the input data file names
	files, err := ioutil.ReadDir(data_path)
	if err != nil {
		panic(err)
	}

	// Process each file
	for _, file := range files {
		processFile(file)
	}

	for year, _ := range year_buf {
		flush(year)
	}
}

func doSortWrite(year int) {

	defer func() { wg.Done() }()

	year_s := fmt.Sprintf("%d", year)
	fn := path.Join(out_path, year_s, "raw.bin")
	fid, err := os.Open(fn)
	if err != nil {
		panic(err)
	}
	dec := gob.NewDecoder(fid)
	var x []rec_t
	for {
		var z rec_t
		err = dec.Decode(&z)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		x = append(x, z)
	}

	sort.Sort(recslice(x))

	ids := make([]string, len(x))
	values := make([]float64, len(x))
	dates := make([]string, len(x))

	for i, y := range x {
		ids[i] = y.Id
		values[i] = y.Value
		da := fmt.Sprintf("%4d-%02d-%02d", y.Year, y.Month, y.Day)
		dates[i] = da
	}

	fname := path.Join(out_path, year_s, "ids.gz")
	ziparray.WriteString(ids, fname)

	fname = path.Join(out_path, year_s, "values.gz")
	ziparray.WriteFloat64(values, fname)

	fname = path.Join(out_path, year_s, "dates.gz")
	ziparray.WriteString(dates, fname)

	err = os.Remove(tfileName(year))
	if err != nil {
		panic(err)
	}
}

func recsort() {

	fmt.Printf("Sorting and writing output...\n")

	// Get a list of the directory names (a directory for each
	// year)
	dirs, err := ioutil.ReadDir(out_path)
	if err != nil {
		panic(err)
	}

	// Reset since we may have used it already in step 1
	wg = sync.WaitGroup{}

	for _, di := range dirs {
		if !di.IsDir() {
			continue
		}

		year, err := strconv.Atoi(di.Name())
		if err != nil {
			panic(err)
		}

		wg.Add(1)
		go doSortWrite(year)
	}

	wg.Wait()
}

func main() {
	step1()
	recsort()
}
