package main

import (
	"testing"
)

func TestTernaryOperator(t *testing.T) {
	src := `
package main

import "fmt"

func main() {
    holes := 2
    type := holes >= 1 ? "swiss" : "cheddar"
    fmt.Println(type)
}`

	expected := `
package main

import "fmt"

func main() {
    holes := 2
    type := func() any { if holes >= 1 { return "swiss" }; return "cheddar" }()
    fmt.Println(type)
}`
	byteString := string(transpile(src));

	if byteString != expected {
        t.Errorf(`%s`, byteString);
        t.Errorf(`Transpiling did not work, expected output was not produced`);
	}
}
