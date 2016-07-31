package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// This script is a small command line utility for performing simple
// grep-like selections on a text file.
//
// Example usage:
//    ./nuclear_grep --country=China --units=7
//
// See nuclear_count_russia.go for more information about the data.

var (
	// A country name
	country_name string

	// Name of the plant
	site_name string

	// Number of reactor units
	num_units int
)

// read_file reads a CSV file and prints to stdout the lines that
// match the selection criteria.
func read_file() {

	// Open the file, panic on error, don't forget to close
	fname := "in_service.csv"
	fid, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer fid.Close()

	rdr := csv.NewReader(fid)
	first := true

	for {
		// Get the next line
		record, err := rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		// Skip the header
		if first {
			first = false
			continue
		}

		// Check the country name if needed
		if country_name != "" && record[3] != country_name {
			continue
		}

		// Check the site name if needed
		if site_name != "" && record[0] != site_name {
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
	read_file()
}
