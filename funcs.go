package toml2php

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
)

// runeInArray 判断给定的rune是否在数组中
func runeInArray(r rune, arr []rune) bool {
	size := len(arr)
	if size == 0 {
		return false
	}
	for i := 0; i < size; i++ {
		if r == arr[i] {
			return true
		}
	}
	return false
}

// runesContains 判断rune列表中是否包含某个值
func runesContains(arr []rune, r rune) bool {
	for _, v := range arr {
		if v == r {
			return true
		}
	}
	return false
}

// isNumeric 判断给定的字符串是否是数字
func isNumeric(str string) bool {
	matched, err := regexp.MatchString(`^(\+|\-)?(0|[1-9]\d*)((\.\d+)?((e|E)(\+|\-)?[1-9]\d*)?)?$`, str)
	if err != nil {
		return false
	}
	return matched
}

func isPositiveIntNumeric(str string) bool {
	matched, err := regexp.MatchString(`^(0|[1-9]+)$`, str)
	if err != nil {
		return false
	}
	return matched
}

// fmtPhpString 格式化为PHP字符串形式
func fmtPhpString(str string) string {
	str = strings.TrimSpace(str)
	if str == "" {
		return "\"\""
	}
	chars := []rune(str)
	charsSize := len(chars)
	buffer := bytes.Buffer{}
	buffer.WriteRune('"')
	for i := 0; i < charsSize; i++ {
		if chars[i] == '"' && chars[i-1] != '\\' {
			buffer.WriteRune('\\')
			buffer.WriteRune('"')
			continue
		}
		if chars[i] == '\n' {
			buffer.WriteRune('\\')
			buffer.WriteRune('n')
			continue
		}
		if (i == 0 && chars[i] == '\\' && (i+1 == charsSize || chars[i+1] == '\n')) ||
			(i > 0 && chars[i] == '\\' && chars[i-1] != '\\' && (i+1 == charsSize || chars[i+1] == '\n')) {
			continue
		}
		buffer.WriteRune(chars[i])
	}
	buffer.WriteRune('"')
	return buffer.String()
}

// normalize 对输入的配置进行标准化处理，以便于后续解析
func normalize(snippet string) (string, error) {
	// 处理换行符
	snippet = strings.ReplaceAll(snippet, "\r\n", "\n")
	snippet = strings.ReplaceAll(snippet, "\n\r", "\n")
	// 处理制表符（\t）
	snippet = strings.ReplaceAll(snippet, "\t", "")

	// Run, char by char.
	normalized := ""
	openString := false
	openLString := false
	openMString := false
	openMLString := false
	openBrackets := 0
	openKeygroup := false
	lineBuffer := ""

	chars := []rune(snippet)
	charsSize := len(chars)
	for i := 0; i < charsSize; i++ {
		keep := true
		if chars[i] == '[' && !openString && !openLString && !openMString && !openMLString {
			openBrackets++
			if openBrackets == 1 && strings.TrimSpace(lineBuffer) == "" {
				openKeygroup = true
			}
		} else if chars[i] == ']' && !openString && !openLString && !openMString && !openMLString {
			if openBrackets > 0 {
				openBrackets--
				if openKeygroup {
					openKeygroup = false
				}
			} else {
				return "", errors.New("Unexpected ']' on : " + lineBuffer)
			}
		} else if openBrackets > 0 && chars[i] == '\n' {
			if openKeygroup {
				return "", errors.New("Multi-line keygroup definition is not allowed on: " + lineBuffer)
			}
			keep = false
		} else if (openString || openLString) && chars[i] == '\n' {
			return "", errors.New("Multi-line string not allowed on: " + lineBuffer)
		} else if chars[i] == '"' && chars[i-1] != '\\' && !openLString && !openMLString {
			if charsSize >= i+3 && string(chars[i:i+3]) == `"""` {
				i += 2
				normalized += `"""`
				lineBuffer += `"""`
				keep = false
				openMString = !openMString
			} else if !openMString {
				openString = !openString
			}
		} else if chars[i] == '\'' && !openString && !openMString {
			if charsSize >= i+3 && string(chars[i:i+3]) == "'''" {
				i += 2
				normalized += "'''"
				lineBuffer += "'''"
				keep = false
				openMLString = !openMLString
			} else if !openMLString {
				openLString = !openLString
			}
		} else if chars[i] == '\\' && chars[i-1] != '\\' && !runeInArray(chars[i+1], []rune{'b', 't', 'n', 'f', 'r', 'u', 'U', '"', '\\', ' '}) {
			if openString {
				return "", errors.New("Reserved special characters inside strings are not allowed: " + string(chars[i]) + string(chars[i+1]))
			}
			if openMString {
				for {
					if chars[i] == '\n' || chars[i+1] == ' ' {
						i++
						keep = false
					} else {
						break
					}
				}
			}
		} else if chars[i] == '#' && !openString && !openKeygroup {
			for {
				if i >= charsSize || chars[i] == '\n' {
					break
				}
				i++
			}
			keep = openBrackets == 0
		}

		// raw lines
		if i < charsSize {
			lineBuffer += string(chars[i])
			if chars[i] == '\n' {
				lineBuffer = ""
			}
			if keep {
				normalized += string(chars[i])
			}
		}
	}

	// Something went wrong.
	if openBrackets > 0 {
		return "", errors.New("Syntax error found on TOML document. Missing closing bracket.")
	}
	if openString {
		return "", errors.New("Syntax error found on TOML document. Missing closing string delimiter.")
	}
	if openMString {
		return "", errors.New("Syntax error found on TOML document. Missing closing multi-line string delimiter.")
	}
	if openLString {
		return "", errors.New("Syntax error found on TOML document. Missing closing literal string delimiter.")
	}
	if openMLString {
		return "", errors.New("Syntax error found on TOML document. Missing closing multi-line literal string delimiter.")
	}
	if openKeygroup {
		return "", errors.New("Syntax error found on TOML document. Missing closing key group delimiter.")
	}

	return normalized, nil
}

// parseTableName Parses TOML table names and returns the hierarchy array of table names.
func parseTableName(chars []rune) []string {
	buffer := bytes.Buffer{}
	strOpen := false
	names := make([]string, 0)

	charsSize := len(chars)
	for i := 0; i < charsSize; i++ {
		if chars[i] == '"' {
			if !strOpen || (strOpen && chars[i-1] != '\\') {
				strOpen = !strOpen
			}
		} else if chars[i] == '.' && !strOpen {
			names = append(names, buffer.String())
			buffer.Reset()
			continue
		}
		buffer.WriteRune(chars[i])
	}
	if buffer.Len() > 0 {
		names = append(names, buffer.String())
	}
	return names
}
