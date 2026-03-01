package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"
)

type TryCatch struct {
	assignment      string
	funcCall        string
	catchBody       string
	catchBodyIndent string
	hasAssign       bool
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("gollin file path required.")
		os.Exit(1)
	}
	path := os.Args[1]

	var gollinPath string

	if strings.HasSuffix(path, ".gol") {
		gollinPath = path
	} else {
		var gollinBuilder strings.Builder
		gollinBuilder.WriteString(path)
		gollinBuilder.WriteString(".gol")
		gollinPath = gollinBuilder.String()
	}

	gollinCode, err := os.ReadFile(gollinPath)
	if err != nil {
		log.Fatal(err)
	}

	goCode := transpile(string(gollinCode))

	newFilePath := strings.Split(gollinPath, ".")[0]
	var goPath strings.Builder
	goPath.WriteString(newFilePath)
	goPath.WriteString(".go")

	os.WriteFile(goPath.String(), goCode, 0777)
}

func transpile(src string) []byte {
	var output strings.Builder
	i := 0

	// Loop over full src to transpile it
	for i < len(src) {
		if matchKeyword(src, i, "try") {
			indent := getIndentAt(src, i)

			// Capture any assignment before "try" like "result := "
			assignment := getAssignmentBefore(output.String())
			if assignment != "" {
				trimAssignmentFromOutput(&output)
			}

			i += len("try") // pass over the full try keyword
			i = skipWhitespace(src, i)

			// Expect opening brace
			if i >= len(src) || src[i] != '{' {
				log.Fatalf("transpile error: 'try' keyword not followed by '{' at position %d\n", i)
			}
			i++ // consume the opening {

			// Collect try body
			tryBody, next := collectBlock(src, i)
			i = next

			i = skipWhitespace(src, i)

			// Check for optional catch
			var catchBody string
			var catchBodyIndent string
			if matchKeyword(src, i, "catch") {
				i += len("catch")
				i = skipWhitespace(src, i)

				if i < len(src) && src[i] == '{' {
					i++ // consume the opening {
					catchBody, i = collectBlock(src, i)
				}

				trimmed := strings.TrimLeft(catchBody, "\n\r")
				catchBodyIndent = trimmed[:len(trimmed)-len(strings.TrimLeft(trimmed, " \t"))]
			}

			tc := TryCatch{
				assignment:      strings.TrimSpace(assignment),
				funcCall:        strings.TrimSpace(tryBody),
				catchBody:       strings.TrimSpace(catchBody),
				catchBodyIndent: catchBodyIndent,
				hasAssign:       strings.TrimSpace(assignment) != "",
			}

			output.WriteString(renderTryCatch(tc, indent))
			continue
		}

		output.WriteByte(src[i])
		i++
	}

	return []byte(output.String())
}

// collectBlock reads characters until the matching closing brace,
// handling nested braces.
func collectBlock(src string, i int) (string, int) {
	var body strings.Builder
	depth := 1

	for i < len(src) && depth > 0 {
		ch := src[i]
		if ch == '{' {
			depth++
			body.WriteByte(ch)
		} else if ch == '}' {
			depth--
			if depth > 0 {
				body.WriteByte(ch)
			}
			// at depth == 0 just stop
		} else {
			body.WriteByte(ch)
		}
		i++
	}

	return body.String(), i
}

// matchKeyword returns true if src[i:] starts with the given keyword
// followed by a non-identifier character (so "tryout" doesn't match "try")
func matchKeyword(src string, i int, keyword string) bool {
	end := i + len(keyword)
	if end > len(src) {
		return false
	}
	if src[i:end] != keyword {
		return false
	}
	// Make sure it's not part of a larger identifier
	if end < len(src) && isIdentChar(rune(src[end])) {
		return false
	}
	return true
}

// extracts any assignment prefix e.g. "result := " or "result = "
func getAssignmentBefore(written string) string {
	lastNewline := strings.LastIndex(written, "\n")
	var currentLine string
	if lastNewline == -1 {
		currentLine = written
	} else {
		currentLine = written[lastNewline+1:]
	}

	currentLine = strings.TrimSpace(currentLine)
	if currentLine == "" {
		return ""
	}

	// Check it ends with := or = but not ==
	if strings.HasSuffix(currentLine, ":=") {
		return currentLine
	}
	if strings.HasSuffix(currentLine, "=") && !strings.HasSuffix(currentLine, "==") {
		return currentLine
	}

	return ""
}

// trimAssignmentFromOutput removes the last partial line from the output builder
// since we'll re-emit it as part of renderTryCatch
func trimAssignmentFromOutput(output *strings.Builder) {
	s := output.String()
	lastNewline := strings.LastIndex(s, "\n")
	output.Reset()
	if lastNewline != -1 {
		output.WriteString(s[:lastNewline+1])
	}
}

func getIndentAt(src string, i int) string {
	lineStart := i
	for lineStart > 0 && src[lineStart-1] != '\n' {
		lineStart--
	}
	end := lineStart
	for end < len(src) && (src[end] == ' ' || src[end] == '\t') {
		end++
	}
	return src[lineStart:end]
}

func skipWhitespace(src string, i int) int {
    for i < len(src) && unicode.IsSpace(rune(src[i])) {
        i++
    }
    return i
}

func isIdentChar(ch rune) bool {
    return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}

func renderTryCatch(tc TryCatch, indent string) string {
	var b strings.Builder

	if tc.hasAssign {
		op := ":="
		if strings.HasSuffix(tc.assignment, "=") && !strings.HasSuffix(tc.assignment, ":=") {
			op = "="
		}
		lhs := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(tc.assignment), "="), ":")
		lhs = strings.TrimSpace(lhs)
		b.WriteString(fmt.Sprintf("%s%s, err %s %s\n", indent, lhs, op, tc.funcCall))
	} else {
		b.WriteString(fmt.Sprintf("%s_, err := %s\n", indent, tc.funcCall))
	}

	b.WriteString(fmt.Sprintf("%sif err != nil {\n", indent))
	if tc.catchBody != "" {
		b.WriteString(fmt.Sprintf("%s%s\n", tc.catchBodyIndent, tc.catchBody))
	} else {
		b.WriteString(fmt.Sprintf("%s\treturn err\n", indent))
	}

	b.WriteString(fmt.Sprintf("%s}\n", indent))

	return b.String()
}
