package main

// This script calculates some simple summaries from the data on
// births and deaths of notable people, available here:
// http://science.sciencemag.org/content/suppl/2014/07/30/345.6196.558.DC1
//
// We use only the dataset labeled "Data S1", credited to Freebase.com
//
// The "freebase_convert.go" script takes the data and converts it
// from Excel to gzipped csv format.  The resulting file is assumed to
// be called FB.csv.gz and should be located at the directory path
// called "dpath" below.
//
// This script takes the geographical latitude and longitude
// coordinates for each person's birth and death location and
// calculates the distance in km between these two points.  It then
// prints a set of quantiles of the distribution of these distances to
// stdout.

import (
	"compress/gzip"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"

	"github.com/paulmach/go.geo"
)

var (
	// The FB.csv.gz file should be located here
	dpath string = "/nfs/kshedden/Freebase"

	// Raw data, map from person's name to birth and death
	// locations
	rdata map[string]*rec_t
)

// Location information for one person
type rec_t struct {
	BLoc   *geo.Point // Birth location
	DLoc   *geo.Point // Death location
	BDDist float64    // Distance from birth to death location
}

// readData reads the raw data file and creates a map from the
// person's name to an instance of the rec_t struct containing birth
// and death location information.  The BDDist field is not filled in
// here.
func readData() {

	// A file reader for the input data file
	fname := path.Join(dpath, "FB.csv.gz")
	fid, err := os.Open(fname)
	if err != nil {
		panic(err)
	}

	// Wrap the file reader in a gzip reader to decompress the
	// contents
	rdr, err := gzip.NewReader(fid)
	if err != nil {
		panic(err)
	}
	cdr := csv.NewReader(rdr)

	// This is only necessary because this csv file has some
	// formatting issues that would otherwise create problems.
	cdr.FieldsPerRecord = -1

	// Make a map from column name to column index
	colix := make(map[string]int)
	x, err := cdr.Read()
	if err != nil {
		panic(err)
	}
	for i, v := range x {
		colix[v] = i
	}

	// Get the indices for columns of interest
	var ii []int
	for _, v := range []string{"PrsLabel", "BLocLat", "BLocLong", "DLocLat", "DLocLong"} {
		col, ok := colix[v]
		if !ok {
			msg := fmt.Sprintf("Can't find %s", v)
			panic(msg)
		}
		ii = append(ii, col)
	}

	// Populate rdata
	tx := make([]float64, 4)
	rdata = make(map[string]*rec_t)
	for {
		rec, err := cdr.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		// File seems to be slightly malformed
		if len(rec) != 19 {
			continue
		}

		// Birth and death coordinates
		for j, i := range ii[1:5] {
			v, err := strconv.ParseFloat(rec[i], 64)
			if err != nil {
				panic(err)
			}
			tx[j] = v
		}

		// Convert the coordinates to geo.Point objects
		bloc := geo.NewPointFromLatLng(tx[0], tx[1])
		dloc := geo.NewPointFromLatLng(tx[2], tx[3])

		r := &rec_t{BLoc: bloc, DLoc: dloc}
		rdata[rec[ii[0]]] = r
	}
}

// getDistances calculates the distance in km between birth and death
// locations for each person.
func getDistances() {
	for _, v := range rdata {
		di := v.BLoc.GeoDistanceFrom(v.DLoc)
		v.BDDist = di / 1000 // Convert to km
	}
}

// sumaries prints some statistical summaries of the data.  The
// summaries are a sequence of quantiles of the distribution of
// distances between birth and death location.
func summaries() {

	// Extract the distances into an array
	dx := make([]float64, len(rdata))
	ii := 0
	for _, v := range rdata {
		dx[ii] = v.BDDist
		ii++
	}

	// Sort the distances
	sort.Float64Slice(dx).Sort()

	// The quantiles to display
	qtl := []float64{0.1, 0.25, 0.5, 0.75, 0.9}

	// Calculate and display the quantiles
	for _, q := range qtl {
		pos := int(q * float64(len(dx)-1))
		fmt.Printf("%5.2f %9.2f\n", q, dx[pos])
	}
}

func main() {
	readData()
	getDistances()
	summaries()
}
