package openhue

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

// PrintJsonArray formats the input array as JSON and prints it. If the length of the array is equal to 1,
// it will print it as a single object
func PrintJsonArray[T any](array []T) {
	if len(array) == 1 {
		PrintJson(array[0])
	} else {
		PrintJson(array)
	}
}

func PrintJson[T any](array T) {
	var out []byte
	out, _ = json.MarshalIndent(array, "", "  ")
	fmt.Println(string(out))
}

// PrintTable prints each line of the objects contained in the table value
func PrintTable[T any](table []T, lineFn func(T) string, header ...string) {

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	for _, h := range header {
		_, _ = fmt.Fprint(w, h+"\t")
	}
	_, _ = fmt.Fprintln(w)

	for range header {
		_, _ = fmt.Fprint(w, "----\t")
	}
	_, _ = fmt.Fprintln(w)

	for _, l := range table {
		_, _ = fmt.Fprintln(w, lineFn(l))
	}

	_ = w.Flush()
}
