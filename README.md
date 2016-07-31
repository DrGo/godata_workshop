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
      [Feather](https://github.com/wesm/feather/issues)

* Basic concurrency

Index of examples
-----------------

* [nuclear_count_russia.go](nuclear_count_russia.go)

* nuclear_make_map.go

Running the workshop scripts
----------------------------

Note that most of the scripts have global variables that you must edit
to point to directories in your file system where you have permission
to read and write files.  In addition, for most scripts some data must
be downloaded, and a bit of pre-processing is necessary.

* `data_path`: the data are read from here, see the comments at the
  top of each script to see how to obtain and pre-process any data
  needed by the script.

* `out_path`: the results are written here, this can usually be any
  directory where you have write permission

### Data sets

#### GCOS surface network climate data

```
wget ftp://ftp.ncdc.noaa.gov/pub/data/ghcn/daily/ghcnd_gsn.tar.gz
gunzip ghcnd_gsn.tar.gz
tar -xvf ghcnd_gsn.tar
cd ghcnd_gsn
gzip *
```