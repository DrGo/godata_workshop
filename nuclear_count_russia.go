package main

// This is a simple introductory script in which we read a small
// text file into Go and do some very basic processing with it.
//
// The data set describes all the nuclear power plants in the world,
// and are available here:
//    https://en.wikipedia.org/wiki/List_of_nuclear_power_stations
//
// We will need the data in CSV format.  You can do that using the tool
// at this site:
//    http://wikitables.geeksta.net/
//
// Or just follow this direct link:
//    http://wikitables.geeksta.net/url/?url=https%3A%2F%2Fen.wikipedia.org%2Fwiki%2FList_of_nuclear_power_stations
//
// You should download all three files and name them "in_service.csv",
// "under_construction.csv", and "shut_down.csv", respectively.

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// countRussia counts the number of lines that mention "Russia" and
// returns the count
func countRussia(name string) int {

	// Open the file, panic on error
	fid, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	// This makes sure we don't forget to close the file (a resource leak)
	defer fid.Close()

	// This is a utility class to help us read through text files
	scanner := bufio.NewScanner(fid)

	// Read the file by line
	n := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Russia") {
			n++
		}
	}

	return n
}

func main() {

	fname := "in_service.csv"
	n := countRussia(fname)

	fmt.Printf("%d lines contain the word \"Russia\"\n", n)
}
