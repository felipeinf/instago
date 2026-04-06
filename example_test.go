package ig_test

import (
	"fmt"

	ig "github.com/felipeinf/instago"
)

func ExampleNewClient() {
	c := ig.NewClient()
	fmt.Println(c != nil)
	// Output:
	// true
}
