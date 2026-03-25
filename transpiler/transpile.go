package transpiler

import (
	"fmt"
	"log"
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

func Transpile(src string) []byte {
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

		// collect strings for string interpolation
		if src[i] == '"' {
			i++ // consume opening quote
			rawString, next := collectString(src, i)
			i = next

			if strings.Contains(rawString, `\{`) {
				trimPrecedingCall(&output)
				output.WriteString(renderInterpolation(rawString))

				// Consume the closing ')' that belonged to the original call
				if i < len(src) && src[i] == ')' {
					i++
				}
			} else {
				// no interpolation, just leave the string alone
				output.WriteByte('"')
				output.WriteString(rawString)
				output.WriteByte('"')
			}
			continue
		}

		// Ternary operator
		if src[i] == '?' {
			condition := getTernaryCondition(output.String())
			if condition != "" {
				trimTernaryCondition(&output)
			}
			i++ // consume '?'
			i = skipWhitespace(src, i)

			trueBranch, next := collectUntil(src, i, ':')
			i = next
			i++ // consme ':'
			i = skipWhitespace(src, i)

			falseBranch, next2 := collectUntil(src, i, '\n', ';')

			i = next2

			output.WriteString(renderTernary(
				strings.TrimSpace(condition),
				strings.TrimSpace(trueBranch),
				strings.TrimSpace(falseBranch),
			))
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

// collectString reads characters from src[i:] up to (but not including) the
// closing unescaped double-quote
func collectString(src string, i int) (string, int) {
	var body strings.Builder
	for i < len(src) {
		ch := src[i]
		// Honor backslash escapes so \" doesn't end the string.
		if ch == '\\' && i+1 < len(src) {
			next := src[i+1]
			// \{ is our interpolation sigil — keep it as-is so
			// renderInterpolation can find it later.
			body.WriteByte(ch)
			body.WriteByte(next)
			i += 2
			continue
		}
		if ch == '"' {
			i++ // consume closing quote
			break
		}
		body.WriteByte(ch)
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

// Trim to the end of the preceding line for string interpolation
func trimPrecedingCall(output *strings.Builder) {
	s := output.String()
	// Find the opening paren of the call
	parenIdx := strings.LastIndex(s, "(")
	lineStart := strings.LastIndex(s[:parenIdx], "\n")
	output.Reset()
	if lineStart >= 0 {
		// Keep everything up to and including the newline
		output.WriteString(s[:lineStart+1])
	}
	// Re-emit the indent
	line := s[lineStart+1 : parenIdx]
	for _, ch := range line {
		if ch == ' ' || ch == '\t' {
			output.WriteRune(ch)
		} else {
			break
		}
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

// renderInterpolation converts a raw string to a fmt.Sprintf call if the
// string interpolation \{variable} formatting is found.
func renderInterpolation(raw string) string {
	var format strings.Builder
	var args []string

	i := 0
	for i < len(raw) {
		// Look for the interpolation sigil \{
		if raw[i] == '\\' && i+1 < len(raw) && raw[i+1] == '{' {
			i += 2 // skip \{

			// Collect everything up to the closing }
			var expr strings.Builder
			for i < len(raw) && raw[i] != '}' {
				expr.WriteByte(raw[i])
				i++
			}
			i++ // consume the closing }

			format.WriteString("%v")
			args = append(args, strings.TrimSpace(expr.String()))
			continue
		}

		if raw[i] == '%' {
			// escape any bare % for Sprintf
			format.WriteString("%%")
		} else {
			// Just pass on regular characters
			format.WriteByte(raw[i])
		}
		i++
	}

	return fmt.Sprintf(`fmt.Sprintf("%s", %s)`,
		format.String(),
		strings.Join(args, ", "),
	)
}

// collectUntil reads characters until the given stop byte
func collectUntil(src string, i int, stops ...byte) (string, int) {
	var body strings.Builder
	for i < len(src) && !isByteIn(src[i], stops) {
		body.WriteByte(src[i])
		i++
	}
	return body.String(), i
}

func isByteIn(b byte, set []byte) bool {
	for _, s := range set {
		if b == s {
			return true
		}
	}
	return false
}

// getTernaryCondition extracts the condition expression
func getTernaryCondition(written string) string {
	lastNewline := strings.LastIndex(written, "\n")
	var line string
	if lastNewline == -1 {
		line = written
	} else {
		line = written[lastNewline+1:]
	}

	if idx := strings.LastIndex(line, ":="); idx != -1 {
		return strings.TrimSpace(line[idx+2:])
	}
	if idx := strings.LastIndex(line, "="); idx != -1 {
		return strings.TrimSpace(line[idx+1:])
	}

	return strings.TrimSpace(line)
}

// trimTernaryCondition removes the condition from the end of the
// output buffer
func trimTernaryCondition(output *strings.Builder) {
	s := output.String()
	lastNewline := strings.LastIndex(s, "\n")
	line := s[lastNewline+1:]

	var keepUntil int
	if idx := strings.LastIndex(line, ":="); idx != -1 {
		keepUntil = lastNewline + 1 + idx + len(":=") + 1
	} else if idx := strings.LastIndex(line, "="); idx != -1 {
		keepUntil = lastNewline + 1 + idx + len("=") + 1
	} else {
		keepUntil = lastNewline + 1
	}

	output.Reset()
	output.WriteString(s[:keepUntil])
}

func renderTernary(condition, trueBranch, falseBranch string) string {
	return fmt.Sprintf("func() any { if %s { return %s }; return %s }()",
		condition, trueBranch, falseBranch)
}
