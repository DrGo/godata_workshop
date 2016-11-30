package main

// This script is a small command line utility for performing simple
// grep-like selections on a text file.
//
// Example usage:
//    ./nuclear_grep --country=China --units=7
//
// The above invocation of the script will print to stdout all the
// records for plants in China that have exactly 7 reactors.  Only the
// plants that are currently in-service are searched.
//
// See nuclear_count_russia.go for more information about preparing
// the input data.

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	// A country name
	country_name string

	// Name of the plant
	site_name string

	// Number of reactor units
	num_units int

	// The primary data source
	datafile string = "in_service.csv"
)

// readFile reads a CSV file and prints to stdout the lines that
// match the selection criteria.
func readFile() {

	// Open the file, panic on error, don't forget to close
	fid, err := os.Open(datafile)
	if err != nil {
		panic(err)
	}
	defer fid.Close()

	rdr := csv.NewReader(fid)

	// Assume the header can be read
	header, _ := rdr.Read()

	// Partial check of file structure
	if header[0] != "Power station" || header[1] != "# Units" || header[3] != "Country" {
		panic("non-conforming files structure")
	}

	for {
		// Get the next line
		record, err := rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		// Check the country name if needed
		if country_name != "" && record[3] != country_name {
			continue
		}

		// Check the site name if needed
		if site_name != "" && !strings.Contains(record[0], site_name) {
			continue
		}

		// Check the number of units if needed
		if num_units != -1 {
			n, err := strconv.Atoi(record[1])
			if err != nil || n != num_units {
				continue
			}
		}

		// If we reach here, this is a selected station
		fmt.Printf("%s\n", strings.Join(record, ","))
	}
}

func main() {

	// Get the search parameters
	flag.StringVar(&country_name, "country", "", "Name of country in which plant is located")
	flag.StringVar(&site_name, "site", "", "Name of site")
	flag.IntVar(&num_units, "units", -1, "Number of units")
	flag.Parse()

	// Scan the file
	readFile()
}
