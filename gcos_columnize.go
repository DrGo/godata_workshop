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

// Configurable values
var (
	// Location of the data in the local file system
	data_path = "/nfs/kshedden/GHCN/ghcnd_gsn"

	// Path where the output file is written
	out_path = "/nfs/kshedden/GHCN_tmp"

	// The temperature type to process, should be either "TMAX" or
	// "TMIN"
	eltype = "TMAX"
)

var (
	// A gob encoder for the data in each year
	year_gob map[int]*gob.Encoder

	// The backing data buffer for each gob encoder.  We can't
	// store the files directly because of limits on the number of
	// simultaneously open files.
	year_buf map[int]*bytes.Buffer

	// Safely send results from goroutines back to parent
	rec_chan chan rec_t

	// Semaphore, used to limit the number of input files being
	// processed simultaneously (since a limited number of files
	// can be open at the same time).
	sem chan bool

	// The number of input files that can be simultaneously processed
	sem_size int = 50

	// Used to manage concurrency
	wg sync.WaitGroup

	// Number of bytes per year to save in-memory before flushing to disk
	buf_size int = 1e7
)

// One data value (either maximum or minimum daily temperature)
type rec_t struct {
	Id    string  // The station id
	Year  int     // The year of the data point
	Month int     // The month of the data point (1..12)
	Day   int     // The day within the month (1..31)
	Value float64 // The data value (TMAX or TMIN)
}

// We will need to sort slices of rec_t values
type recslice []rec_t

// Needed for sorting.  Less implies sorting by station first, then
// by date.
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

// setupYear creates data structures to handle all the data we
// encounter for one year.  It also truncates the file that will be
// used for temporary data storage for the year's data.
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
}

// parse processes one row of data from a raw input file (i.e. data
// for all days in one month for one station).
func parse(line string) {

	var err error

	// The station id
	id := line[0:11]

	// The year for the data value
	year, err := strconv.Atoi(line[11:15])
	if err != nil {
		panic(err)
	}

	// The month for the data value
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

		// The data value (maximum or minimum temperature)
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
		rec_chan <- r
	}
}

// processFile handles all processing for one data file (for one station).
func processFile(file os.FileInfo) {

	defer func() {
		<-sem
		wg.Done()
	}()

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

		// Check the element type first so we can skip the
		// line if not being used.
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

// flush writes the data from one in-memory buffer to a file on disk.
func flush(year int) {

	fmt.Printf("Flushing %d\n", year)
	fname := tfileName(year)
	fid, err := os.OpenFile(fname, os.O_APPEND|os.O_WRONLY, 0700)
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

// processRaw processes the raw data into native go data structures.
func processRaw() {

	year_buf = make(map[int]*bytes.Buffer)
	year_gob = make(map[int]*gob.Encoder)
	rec_chan = make(chan rec_t)
	sem = make(chan bool, sem_size)

	// Reset the output directory
	err := os.RemoveAll(out_path)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(out_path, 0700)
	if err != nil {
		panic(err)
	}

	for year := 1833; year <= 2016; year++ {
		setupYear(year)
	}

	// Get a list of the input data file names
	files, err := ioutil.ReadDir(data_path)
	if err != nil {
		panic(err)
	}

	// Process each file
	go func() {
		for _, file := range files {
			wg.Add(1)

			// We will only be able to put sem_size true's
			// into the semaphore channel at once.  When a
			// call to processFile completes, we remove
			// one value from sem so that this loop can
			// proceed to the next file.
			sem <- true

			go processFile(file)
		}
		wg.Wait()

		// Close the channel to signal that we can stop reading.
		close(rec_chan)
	}()

	// Read until the channel is closed
	for r := range rec_chan {
		year := r.Year
		if year_gob[year] == nil {
			fmt.Printf("%v\n", year)
			panic("")
		}
		err = year_gob[year].Encode(r)
		if err != nil {
			panic(err)
		}
		if year_buf[year].Len() > buf_size {
			flush(year)
		}
	}

	// Write whatever is left to disk
	for year, _ := range year_buf {
		flush(year)
	}
}

// doSortWrite takes the native go data structure for one year, sorts
// by station then by date, and finally creates the final output
// files.
func doSortWrite(year int) {

	defer func() { wg.Done() }()

	// Read the sequence of rec_t values from disk into an array.
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

	// Sort by station then by date
	sort.Sort(recslice(x))

	// Split the go struct into arrays for each field.
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

	// Remove the temporary data file.
	err = os.Remove(tfileName(year))
	if err != nil {
		panic(err)
	}
}

// recsort loops over the years and manages the process of sorting and
// generating final output.
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
	processRaw()
	recsort()
}
