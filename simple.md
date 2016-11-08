A very simple Go program
------------------------

Begin by copying this short program into a file named `simple.go`:

```
package main

func main() {

        fmt.Printf("Hello!\n")

}
```

If you have configured your editor to use `goimports`, upon saving the
file the editor will rewrite the contents as:

```
package main

import "fmt"

func main() {

        fmt.Printf("Hello!\n")

}
```

If your editor does not make this change, you should enter the
`import` line as shown above manually.

Next you can run the program from the command line by typing:

```
go run simple.go
```

A slightly less simple Go program
---------------------------------

Copy the following into a file called `less_simple.go`:

```
package main

func print_string(x string) {
        fmt.Printf("%s\n", x)
}

func print_int(x int) {
        fmt.Printf("%d\n", x)
}

func main() {

        print_string("cat")

        print_int(3)
}
```

Again, you will need to manually enter the `import` line if you don't
have your editor configured to enter it for you.