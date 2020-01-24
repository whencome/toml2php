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
   matched, err := regexp.MatchString(`^(\+|\-)?\d*(\.\d+)?$`, str)
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
    snippet = strings.ReplaceAll(snippet, "\n", "\\n")

    // Run, char by char.
    normalized     := ""
    openString     := false
    openLString    := false
    openMString    := false
    openMLString   := false
    openBrackets   := 0
    openKeygroup   := false
    lineBuffer     := ""

    chars := []rune(snippet)
    charsSize := len(chars)
    for i:=0; i<charsSize; i++ {
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
            if string(chars[i:i+3]) == `"""` {
                i += 2
                normalized += `"""`
                lineBuffer += `"""`
                keep = false
                openMString = !openMString
            } else if !openMString {
                openString = !openString
            }
        } else if chars[i] == '\'' && !openString && !openMString {
            if string(chars[i:i+3]) == "'''" {
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
                if i < charsSize && chars[i] != '\n' {
                    i++
                } else {
                    break
                }
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

    // chars := []rune(name)
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

// parseValue 将TOML值解析为php的语法字符串
func parseValue(val string) (string, error) {
    val = strings.TrimSpace(val)
    if val == "" {
        return "", errors.New("Empty value not allowed")
    }
    chars := []rune(val)
    charsSize := len(chars)
    parsedVal := make([]rune, 0)
    // boolean & numbers
    if val == "true" || val == "false" || isNumeric(val) {
        return val, nil
    }
    // Literal multi-line string
    if string(chars[0:3]) == `'''` && string(chars[charsSize-3:charsSize]) == `'''` {
        parsedVal = chars[3:charsSize-3]
        if parsedVal[0] == '\n' {
            parsedVal = parsedVal[1:]
        }
        return fmtPhpString(string(parsedVal)), nil
    }
    if string(chars[0:3]) == `"""` && string(chars[charsSize-3:charsSize]) == `"""` {
        parsedVal = chars[3:charsSize-3]
        if parsedVal[0] == '\n' {
            parsedVal = parsedVal[1:]
        }
        return fmtPhpString(string(parsedVal)), nil
    }
    // Literal string
    if chars[0] == '\'' && chars[charsSize-1] == '\'' {
        if runesContains(chars, '\n') {
            return "", errors.New("New lines not allowed on single line string literals.")
        }
        return fmtPhpString(string(chars[1:charsSize-1])), nil
    }
    // string
    if chars[0] == '"' && chars[charsSize-1] == '"' {
        return fmtPhpString(string(chars[1:charsSize-1])), nil
    }
    // datetime not supported temporarily for now

    // Single line array (normalized)
    if chars[0] == '[' && chars[charsSize-1] == ']' {
        return parseArray(chars)
    }
    // Inline table (normalized)
    if chars[0] == '{' && chars[charsSize-1] == '}' {
        return parseInlineTable(chars)
    }
    return "", errors.New("Unknown value type: "+val)
}

// parseArray 
func parseArray(chars []rune) (string, error) {
    // result to save parse result
    result := bytes.Buffer{}
    openBrackets    := 0
    openString      := false
    openCurlyBraces := 0
    openLString     := false
    buffer          := ""
    checkComma      := false

    charsSize := len(chars)
    result.WriteString("[")
    for i := 0; i < charsSize; i++ {
        if chars[i] == '[' && !openString && !openLString {
            openBrackets++
            if openBrackets == 1 {
                continue
            }
        } else if chars[i] == ']' && !openString && !openLString {
            openBrackets--
            if openBrackets == 0 {
                if strings.TrimSpace(buffer) != "" {
                    parsed, err := parseValue(strings.TrimSpace(buffer))
                    if err != nil {
                        return "", err
                    }
                    result.WriteString(parsed)
                }
                // 数据类型检查
                // 跳过
                result.WriteString("]")
                return result.String(), nil
            }
        } else if chars[i] == '"' && chars[i-1] != '\\' && !openLString {
            openString = !openString
        } else if chars[i] == '\'' && !openString {
            openLString = !openLString
        } else if chars[i] == '{' && !openString && !openLString {
            openCurlyBraces++
        } else if chars[i] == '}' && !openString && !openLString {
            openCurlyBraces--
        }

        if (chars[i] == ',' || chars[i] == '}') && !openString && !openLString && openBrackets == 1 && openCurlyBraces == 0 {
            if chars[i] == '}' {
                buffer += string(chars[i])
            }
            buffer = strings.TrimSpace(buffer)
            if buffer != "" {
                parsed, err := parseValue(strings.TrimSpace(buffer))
                if err != nil {
                    return "", err
                }
                if checkComma {
                    result.WriteString(", ")
                    checkComma = false
                }
                result.WriteString(parsed)
                // result.WriteString(", ")
                checkComma = true
            }
            buffer = ""
        } else {
            buffer += string(chars[i])
        }
    }
    return "", errors.New("Wrong array definition:" + string(chars))
}

// parseInlineTable Parse inline tables into common table array
func parseInlineTable(chars []rune) (string, error) {
    charsSize := len(chars)
    if chars[0] == '{' && chars[charsSize-1] == '}' {
        chars = chars[1: charsSize-1]
    } else {
        return "", errors.New("Invalid inline table definition: " + string(chars))
    }

    charsSize = len(chars)
    result := bytes.Buffer{}
    openString := false
    openLString := false
    openBrackets := 0
    checkComma := false
    buffer := ""

    result.WriteString("[")
    for i := 0; i < charsSize; i++ {
        if chars[i] == '"' && chars[i-1] != '\\' {
            openString = !openString
        } else if chars[i] == '\'' {
            openLString = !openLString
        } else if chars[i] == '[' && !openString && !openLString {
            openBrackets++
        } else if chars[i] == ']' && !openString && !openLString {
            openBrackets--
        }

        if chars[i] == ',' && !openString && !openLString && openBrackets == 0 {
            parsed, err := parseInlineTableFieldValue(buffer)
            if err != nil {
                return "", err
            }
            if checkComma {
                result.WriteString(", ")
                checkComma = false
            }
            result.WriteString(parsed)
            // result.WriteString(", ")
            checkComma = true
            buffer = ""
        } else {
            buffer += string(chars[i])
        }
    }

    // parse last buffer
    parsed, err := parseInlineTableFieldValue(buffer)
    if err != nil {
        return "", err
    }
    if checkComma {
        result.WriteString(", ")
        checkComma = false
    }
    result.WriteString(parsed)
    // 补充最后的括号
    result.WriteString("]")
    return result.String(), nil
}

// parseInlineTableFieldValue 解析键值对内容
func parseInlineTableFieldValue(snippet string) (string, error) {
    result := bytes.Buffer{}
    pos := strings.Index(snippet, "=")
    if pos <= 0 {
        return "", errors.New("[split] invalid inline toml table data: " + snippet)
    }
    field := strings.TrimSpace(snippet[0:pos])
    val := strings.TrimSpace(snippet[pos+1:])
    if isNumeric(field) {
        result.WriteString(field)
    } else {
        result.WriteString(`"`)
        result.WriteString(field)
        result.WriteString(`"`)
    }
    result.WriteString(" => ")
    parsed, err := parseValue(val)
    if err != nil {
        return "", errors.New("[parse] invalid inline toml table data: " + val)
    }
    result.WriteString(parsed)
    return result.String(), nil
}

// parseKeyValue takes a key expression and the current pointer to position in the right hierarchy.
// Then it sets the corresponding value on that position.
// func parseKeyValue(key, val string) (string, error) {
//     openQuote := false
//     openDoubleQuote := false
//     buffer := ""
//     result := bytes.Buffer{}
//
//     keyChars := []rune(key)
//     keyCharsSize := len(keyChars)
//
//     for i := 0; i < keyCharsSize; i++ {
//         // Handle quoting
//         if keyChars[i] == '"' {
//             if !openQuote {
//                 openDoubleQuote = !openDoubleQuote
//                 continue
//             }
//         }
//         if keyChars[i] == '\'' {
//             if !openDoubleQuote {
//                 openQuote = !openQuote
//                 continue
//             }
//         }
//
//         // Handle dotted keys
//         if keyChars[i] == '.' && !openQuote && !openDoubleQuote {
//
//         }
//     }
// }

