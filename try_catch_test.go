package main

import (
	"testing"
)

// call matchKeyword and verify a success
func TestMatchKeywordSuccess(t *testing.T) {
    isMatch := matchKeyword("try", 0, "try")
	if !isMatch {
        t.Errorf(`the keyword 'try' did not match but it should have`);
	}
}

// call matchKeyword and verify a fail
func TestMatchKeywordFail(t *testing.T) {
    isMatch := matchKeyword("nottry", 0, "try")
	if isMatch {
        t.Errorf(`the keyword 'try' matched but it shouln't have`);
	}
    isMatch = matchKeyword("try", 3, "try")
	if isMatch {
        t.Errorf(`the keyword 'try' matched but it shouln't have`);
	}
    isMatch = matchKeyword("TRY", 3, "try")
	if isMatch {
        t.Errorf(`the keyword 'try' matched but it shouln't have`);
	}
    isMatch = matchKeyword("trysomemore", 0, "try")
	if isMatch {
        t.Errorf(`the keyword 'try' matched but it shouln't have`);
	}
}

func TestTranspileWithoutGollinCode(t *testing.T) {
	src := `
package main

import "os"

func main() {
    f := os.Open("test.txt")
}`

	expected := `
package main

import "os"

func main() {
    f := os.Open("test.txt")
}`
	byteString := string(transpile(src));

	if byteString != expected {
        t.Errorf(`Transpiling did not work, expected output was not produced`);
	}
}

func TestFullTranspile(t *testing.T) {
	src := `
package main

import "os"

func main() {
    f := try {
        os.Open("test.txt")
    } catch {
        fmt.Println("I could not open that text file");
    }
}`

	expected := `
package main

import "os"

func main() {
    f, err := os.Open("test.txt")
    if err != nil {
        fmt.Println("I could not open that text file");
    }

}`
	byteString := string(transpile(src));

	if byteString != expected {
        t.Errorf(`Transpiling did not work, expected output was not produced`);
	}
}
