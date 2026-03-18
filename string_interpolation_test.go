package main

import (
	"testing"
)

func TestRegularStringIsLeftAlone(t *testing.T) {
	src := `
package main

import "fmt"

func main() {
	cheese := "cheddar"
	n := 1000
	fmt.Printf("I like %s times %d", cheese, n)
}`

	expected := `
package main

import "fmt"

func main() {
	cheese := "cheddar"
	n := 1000
	fmt.Printf("I like %s times %d", cheese, n)
}`
	byteString := string(transpile(src));

	if byteString != expected {
        t.Errorf(`Transpiling did not work, expected output was not produced`);
	}
}

func TestStringInterpolation(t *testing.T) {
	src := `
package main

import "fmt"

func main() {
	cheese := "cheddar"
	fmt.Printf("I like \{cheese}")
}`

	expected := `
package main

import "fmt"

func main() {
	cheese := "cheddar"
	fmt.Sprintf("I like %v", cheese)
}`
	byteString := string(transpile(src));

	if byteString != expected {
        t.Errorf(`Transpiling did not work, expected output was not produced`);
	}
}

func TestMultipleStringInterpolation(t *testing.T) {
	src := `
package main

import "fmt"

func main() {
	cheese := "cheddar"
    crackers := "ritz"
	fmt.Printf("I like \{cheese} and \{crackers}")
}`

	expected := `
package main

import "fmt"

func main() {
	cheese := "cheddar"
    crackers := "ritz"
	fmt.Sprintf("I like %v and %v", cheese, crackers)
}`
	byteString := string(transpile(src));

	if byteString != expected {
        t.Errorf(`Transpiling did not work, expected output was not produced`);
	}
}


func TestStringAndIntInterpolation(t *testing.T) {
	src := `
package main

import "fmt"

func main() {
	cheese := "cheddar"
	n := 1000
	fmt.Printf("I like \{cheese} times \{n}")
}`

	expected := `
package main

import "fmt"

func main() {
	cheese := "cheddar"
	n := 1000
	fmt.Sprintf("I like %v times %v", cheese, n)
}`
	byteString := string(transpile(src));

	if byteString != expected {
        t.Errorf(`Transpiling did not work, expected output was not produced`);
	}
}
