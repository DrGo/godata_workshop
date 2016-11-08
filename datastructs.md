Go has two built-in data structures that are used in most Go programs.

Arrays and slices
-----------------

A *slice* is the native Go data structure for homogeneous sequential
data.  A slice is backed by a data structure called an *array*, but in
practice nearly everything is a slice, so that is the main focus here.

Like most objects in Go, slices are typed:

```
x := []int{3, 4, 5}

y := []string{"cat", "dog", "mouse"}

z := []float64{3.5, 7, -1}
```

Go supports slicing and element access with the bracket `[ ]` syntax:

```
x := []int{1, 3, 5, 7, 9}

y := x[1:3] // a slice of ints sharing an aray with x

z := x[1]   // an int
```

You can append a scalar to a slice, and you can append a slice to a slice:

```
x := []int{1, 3, 5, 7, 9}

x = append(x, 3)

y := []int{2, 4, 6}

z := append(x, y...)
```

You can allocate a new slice, and copy data from another slice into
that space.

```
x := int{1, 3, 5, 7, 9}

y := make([]int, len(x))

copy(y, x)
```

Maps
----

A *map* is a data structure that associates unique keys with values.
The keys and values must belong to a declared type:

```
h := make(map[string]int)

h["apple"] = 5
h["banana"] = 6
```

Structs
-------

Go also has a compound data type called a *struct*, which is a data
structure that encapsulates several values of defined types.  An
example struct definition is:

```
type country struct {
    name       string
    capital    string
    population int
}
```

Here are a few ways to create and use a `country` object:

```
var x country = country{"Mexico", "Mexico City", 119530753}

var y country = country{population: 119530753, name: "Mexico", captital: "Mexico City"}

var y country
y.capital = "Mexico City"
y.population = 119530753
y.name = "Mexico"
```

We can also use `country` objects in slices and maps:

```
x := make([]country, 10)

for i, _ := range x {
    x[i] = country{"Mexico", "Mexico City", 119530753}
}
```

```
h := make(map[int]country)

h[3425] = country{"Mexico", "Mexico City", 119530753}
h[8374] = cuntry{"Hungary", "Budapest", 9855571}
```