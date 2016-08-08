Go for data processing
======================

This site contains scripts used in a
[CSCAR](http://cscar.research.umich.edu) workshop on using
[Go](http://golang.org) for data processing.

Go resources
------------

* [Go project web site](http://golang.org)

* [Effective Go](https://golang.org/doc/effective_go.html)

* [Go tour](https://tour.golang.org/welcome/1)

* [Go playground](https://play.golang.org/)

* The Go [standard library](https://golang.org/pkg/)

* The Go project [wiki](https://github.com/golang/go/wiki)

Essential concepts
------------------

* Basic Go language

* Go tools

* Native Go data structures: slices, maps, structs

* Data input and output

    * Working with text files

    * Working with structured data using JSON and Gobs

    * Working with raw binary data

    * Binary data containers [Apache
      Arrow](https://github.com/apache/arrow) and
      [Feather](https://github.com/wesm/feather) -- won't be covered in
      August 2016 offering of the workshop

* Basic concurrency


Index of examples
-----------------

* [nuclear_count_russia.go](nuclear_count_russia.go) (basic file reading)

* [nuclear_make_map.go](nuclear_make_map.go) (csv reading, making and inverting maps)

* [nuclear_grep.go](nuclear_grep.go) (flags)

* [nuclear_json.go](nuclear_json.go) (json and gob serialization, structs)

* [streaming.go](streaming.go) (harvest Twitter streams)

* [gcos_monthly.go](gcos_monthly.go)

* [gcos_monthly_concurrent.go](gcos_monthly_concurrent.go)

* [gcos_columnize.go](gcos_columnize.go) (concurrency, serialization, binary data, file system manipulations)


Go libraries for data processing
--------------------------------

This workshop is primarily about writing programs to manipulate data
using the core Go language.  There are also some Go libraries that can
be used for more specialized data processing:

* [gonum](https://github.com/gonum) -- a collection of numerical libraries

* [go-twitter](https://github.com/dghubble/go-twitter) -- a library for accessing the Twitter API

* [gogeo](https://github.com/paulmach/go.geo) -- a geometry/geography library

* [awk](https://github.com/spakin/awk) -- Awk-like processing of text files

* [biogo](https://github.com/biogo/biogo) -- A bioinformatics library