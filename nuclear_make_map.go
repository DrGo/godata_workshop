package main

// This script illustrates some basic processing of a text file
// containing information about all the nuclear power plants in the
// world.  The focus here is on reading the data from a CSV file and
// placing it into Go data structures, and then doing some simple
// manipulations of the data structures.
//
// See nuclear_count_russia.go for more information about the data.

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	// Map from each power plant site name to the number of
	// reactors
	num_reactors map[string]int

	// Map from each possible reactor size to the corresponding
	// list of reactor names.
	by_size map[int][]string
)

// make_map populates the map named num_reactors, that maps each site
// name to the corresponding number of reactors at the site.
func makeMap(fname string) {

	// Open the file, panic on error, don't forget close
	fid, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer fid.Close()

	rdr := csv.NewReader(fid)

	// Assume the header can be read
	header, _ := rdr.Read()

	// Partial check of file structure
	if header[0] != "Power station" || header[1] != "# Units" {
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

		// The number of reactors is assumed to be in position
		// 1.
		n, err := strconv.Atoi(record[1])
		if err != nil {
			panic(err)
		}

		// The name is always in position 0
		num_reactors[record[0]] = n
	}

	// Print the first 5 locations and their reactor count
	n := 0
	for k, v := range num_reactors {
		fmt.Printf("%s  %d\n", k, v)
		n++
		if n > 5 {
			break
		}
	}
	fmt.Printf("\n\n")
}

// invert_map populates a map called by_size (a global variable) that
// maps each possible plant size (number of reactors) to the list of
// names of sites having that size.  These lists are sorted.
func invertMap() {

	// Fill in the map
	for name, size := range num_reactors {
		by_size[size] = append(by_size[size], name)
	}

	// Sort each list of plant names
	for _, name := range by_size {
		sort.StringSlice(name).Sort()
	}

	// Print the first 5 plants that have exactly two reactors
	fmt.Printf("%s\n", strings.Join(by_size[2][0:5], "\n"))
}

func main() {

	// Maps must be "made" before they can be used
	num_reactors = make(map[string]int)
	by_size = make(map[int][]string)

	makeMap("in_service.csv")

	invertMap()
}
