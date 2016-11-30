package main

// This script takes three csv files containing data about nuclear
// power plants, stores the data for each plant as a struct, then
// writes the structs to files in json and gob formats.
//
// Missing values are indicated with the zero type fo the
// corresponding type, which means a 0 for numeric variables.
//
// See nuclear_count_russia.go for more information about the data.

import (
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	// The names of the raw data files, which should be in the working directory.
	files []string = []string{"in_service.csv", "shut_down.csv", "under_construction.csv"}

	// Encoders for creating json and gob format files.
	jenc *json.Encoder
	genc *gob.Encoder
)

// A representation of the data for one power plant.
type powerplant struct {
	// The name of the plant
	Name string

	// The number of reactor units
	Units int64

	// The capacity in megawatts
	Capacity float64

	// The country where the plant is located
	Country string

	// The geospatial coordinates of the plant
	Location geopoint
}

// A simple representation of a location on the Earth's surface
type geopoint struct {
	// The latitude coordinate
	Latitude float64

	// The longitude coordinate
	Longitude float64
}

// processUnits takes the string form of the number of reactors and
// returns it into an int64.
func processUnits(raw string) int64 {
	// The raw value contains commas
	raw = strings.Replace(raw, ",", "", -1)
	n, err := strconv.Atoi(raw)
	if err != nil {
		panic(err)
	}
	return int64(n)
}

// processCapacity takes the string form of the plant capacity (in MW)
// and returns it as a float64.
func processCapacity(raw string) float64 {
	// The raw value contains commas
	raw = strings.Replace(raw, ",", "", -1)

	// The raw value contains trailing characters that must be
	// removed
	ii := -1
	for i, x := range raw {
		if !strings.ContainsRune("0123456789", x) {
			ii = i
			break
		}
	}
	if ii != -1 {
		raw = raw[0:ii]
	}

	c, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		c = 0
	}
	return c
}

// processLocation takes the string form of a power plant's location
// and returns it as a geopoint structure.  The raw form of the
// location is "... / ... / ### ; ### (...)", where the two ### values
// are the latitude and longitude of the plant, respectively.
func processLocation(raw string) geopoint {
	raw = strings.Split(raw, "/")[2]
	raw = strings.Split(raw, "(")[0]
	fields := strings.Split(raw, ";")
	for j, x := range fields {
		// There are \uFEFF (zero-width spaces) in the file
		fields[j] = strings.Trim(x, " \ufeff")
	}
	nfields := make([]float64, 2)
	for j, v := range fields {
		var err error
		nfields[j], err = strconv.ParseFloat(v, 64)
		if err != nil {
			fmt.Printf(":%v:\n", fields[1])
			panic(err)
		}
	}
	return geopoint{Latitude: nfields[0], Longitude: nfields[1]}
}

// findColumn returns a map from a column name to its numeric position
// in the file.
func findColumn(header []string) map[string]int {
	colix := make(map[string]int)
	for i, v := range header {
		colix[v] = i
	}
	return colix
}

// processFile handles reading, conversion, and output generation for
// all plants in one data file.
func processFile(fname string) {

	// Open the file, panic on error, don't forget to close
	fid, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer fid.Close()
	rdr := csv.NewReader(fid)

	// Assume we can always read the header
	header, _ := rdr.Read()

	// The capacity column is named inconsistently among the files
	// so we rename it to something consistent.
	for k, v := range header {
		if strings.Contains(strings.ToLower(v), "capacity") {
			header[k] = "Capacity (MW)"
		}
	}

	colix := findColumn(header)

	for {
		// Get the next line
		record, err := rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		var name string
		pos, ok := colix["Power station"]
		if ok {
			name = record[pos]
		}

		var n_units int64
		pos, ok = colix["# Units"]
		if ok {
			n_units = processUnits(record[1])
		}

		var location geopoint
		pos, ok = colix["Location"]
		if ok {
			location = processLocation(record[pos])
		}

		var capacity float64
		pos, ok = colix["Capacity (MW)"]
		if ok {
			capacity = processCapacity(record[2])
		} else {
			panic(fmt.Sprintf("%+v\n", colix))
		}

		var country string
		pos, ok = colix["Country"]
		if ok {
			country = record[pos]
		}

		plant := powerplant{Name: name, Units: n_units, Capacity: capacity,
			Country: country, Location: location}

		genc.Encode(plant)
		jenc.Encode(plant)
	}
}

func main() {

	// Set up the json encoder
	fid, err := os.Create("nuclear.json")
	if err != nil {
		panic(err)
	}
	defer fid.Close()
	jenc = json.NewEncoder(fid)

	// Set up the gob encoder
	fid, err = os.Create("nuclear.gob")
	if err != nil {
		panic(err)
	}
	defer fid.Close()
	genc = gob.NewEncoder(fid)

	for _, fname := range files {
		processFile(fname)
	}
}
