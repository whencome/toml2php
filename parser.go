package toml2php

import (
    "errors"
    "strconv"
    "strings"
)

// parse Parse PHP Array
func parse(toml string) (*PHPArray, error) {
    phpArr := &PHPArray{}
    
    // normalize the toml string
    toml, err := normalize(toml)
    if err != nil {
        return nil, err
    }

    // split lines
    arrToml := strings.Split(toml, "\n")
    arrSize := len(arrToml)

    var recurseKeys []string
    for ln := 0; ln < arrSize; ln++ {
        line := []rune(strings.TrimSpace(arrToml[ln]))
        lineSize := len(line)

        // Skip commented and empty lines
        if lineSize == 0 || line[0] == '#' {
            continue
        }
        
        // Array of Tables
        if string(line[0:2]) == "[[" && string(line[lineSize-2:]) == "]]" {
            tableName := line[2:lineSize-2]
            aTables := parseTableName(tableName)
            if len(aTables) <= 0 {
                continue
            }
            recurseKeys = make([]string, len(aTables))
            copy(recurseKeys, aTables)
            phpArr.AddRecurseKeys(aTables)
        } else if string(line[0:1]) == "[" && string(line[lineSize-1:]) == "]" {
            tableName := line[1:lineSize-1]
            aTables := parseTableName(tableName)
            if len(aTables) <= 0 {
                continue
            }
            recurseKeys = make([]string, len(aTables))
            copy(recurseKeys, aTables)
            phpArr.AddRecurseKeys(aTables)
        } else if runesContains(line, '=') {
            rawLine := string(line)
            pos := strings.Index(rawLine, "=")
            field := strings.TrimSpace(rawLine[0:pos])
            val := strings.TrimSpace(rawLine[pos+1:])
            valSize := len(val)
            if valSize >= 3 && val[0:3] == `"""` {
                if valSize == 3 || (valSize > 3 && val[valSize-3:] != `"""`) {
                    for {
                        ln++
                        nextLine := strings.TrimSpace(arrToml[ln])
                        val += "\n"
                        val += arrToml[ln]
                        if nextLine == `"""` || (len(nextLine) > 3 && nextLine[len(nextLine)-3:] == `"""`) {
                            break
                        }
                    }
                }
            }
            if valSize >= 3 && val[0:3] == `'''` {
                if valSize == 3 || (valSize > 3 && val[valSize-3:] != `'''`) {
                    for {
                        ln++
                        nextLine := strings.TrimSpace(arrToml[ln])
                        val += "\n"
                        val += arrToml[ln]
                        if nextLine == `'''` || (len(nextLine) > 3 && nextLine[len(nextLine)-3:] == `'''`) {
                            break
                        }
                    }
                }
            }
            k := buildRecurseKey(recurseKeys, field)
            err = parsePHPKeyValue(phpArr, k, val)
            if err != nil {
                return nil, err
            }
        } else if string(line[0:1]) == "[" && string(line[lineSize-1:]) != "]" {
            return nil, errors.New("Key groups have to be on a line by themselves: " + string(line))
        } else {
            return nil, errors.New("Syntax error on: " + string(line))
        }
    }

    return phpArr, nil
}

func buildRecurseKey(precedingKeys []string, key string) string {
    if precedingKeys == nil || len(precedingKeys) == 0 {
        return key
    }
    return strings.Join(precedingKeys, ".") + "." + key
}

func parsePHPValue(val string) (*PHPValue, error) {
    val = strings.TrimSpace(val)
    if val == "" {
        return nil, errors.New("Empty value not allowed")
    }
    chars := []rune(val)
    charsSize := len(chars)
    parsedVal := make([]rune, 0)
    // boolean
    if val == "true" || val == "false" {
        return NewPHPBoolValue(val), nil
    }
    // numbers
    if isNumeric(val) {
        return NewPHPNumberValue(val), nil
    }
    // Literal multi-line string
    if string(chars[0:3]) == `'''` && string(chars[charsSize-3:charsSize]) == `'''` {
        parsedVal = chars[3:charsSize-3]
        if parsedVal[0] == '\n' {
            parsedVal = parsedVal[1:]
        }
        return NewPHPStringValue(string(parsedVal)), nil
    }
    if string(chars[0:3]) == `"""` && string(chars[charsSize-3:charsSize]) == `"""` {
        parsedVal = chars[3:charsSize-3]
        if parsedVal[0] == '\n' {
            parsedVal = parsedVal[1:]
        }
        return NewPHPStringValue(string(parsedVal)), nil
    }
    // Literal string
    if chars[0] == '\'' && chars[charsSize-1] == '\'' {
        if runesContains(chars, '\n') {
            return nil, errors.New("New lines not allowed on single line string literals.")
        }
        return NewPHPStringValue(string(chars[1:charsSize-1])), nil
    }
    // string
    if chars[0] == '"' && chars[charsSize-1] == '"' {
        return NewPHPStringValue(string(chars[1:charsSize-1])), nil
    }
    // Single line array (normalized)
    if chars[0] == '[' && chars[charsSize-1] == ']' {
        phpArr, err := parsePHPArray(chars)
        if err != nil {
            return nil, err
        }
        return NewPHPArrayValue(phpArr), nil
    }
    // Inline table (normalized)
    if chars[0] == '{' && chars[charsSize-1] == '}' {
        phpArr, err := parsePHPInlineTable(chars)
        if err != nil {
            return nil, err
        }
        return NewPHPArrayValue(phpArr), nil
    }
    return nil, errors.New("Unknown value type: "+val)
}

func parsePHPArray(chars []rune) (*PHPArray, error) {
    openBrackets    := 0
    openString      := false
    openCurlyBraces := 0
    openLString     := false
    buffer          := ""

    charsSize := len(chars)
    phpArr := NewPHPArray()
    keyPos := 0
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
                    phpVal, err := parsePHPValue(strings.TrimSpace(buffer))
                    if err != nil {
                        return nil, err
                    }
                    phpArr.AddChild(strconv.Itoa(keyPos), phpVal)
                    keyPos++
                }
                return phpArr, nil
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
                phpVal, err := parsePHPValue(strings.TrimSpace(buffer))
                if err != nil {
                    return nil, err
                }
                phpArr.AddChild(strconv.Itoa(keyPos), phpVal)
                keyPos++
            }
            buffer = ""
        } else {
            buffer += string(chars[i])
        }
    }
    return nil, errors.New("Wrong array definition:" + string(chars))
}

// parsePHPInlineTable Parse inline tables into common table array
func parsePHPInlineTable(chars []rune) (*PHPArray, error) {
    charsSize := len(chars)
    if chars[0] == '{' && chars[charsSize-1] == '}' {
        chars = chars[1: charsSize-1]
    } else {
        return nil, errors.New("Invalid inline table definition: " + string(chars))
    }

    charsSize = len(chars)
    openString := false
    openLString := false
    openBrackets := 0
    buffer := ""

    // save result
    phpArr := NewPHPArray()
    // keyPos := 0

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
            parsedPHPArr, err := parsePHPInlineTableFieldValue(buffer)
            if err != nil {
                return nil, err
            }
            // phpArr.AddChild(strconv.Itoa(keyPos), NewPHPArrayValue(parsedPHPArr))
            phpArr.MergeChilds(parsedPHPArr)
            // keyPos++
            buffer = ""
        } else {
            buffer += string(chars[i])
        }
    }

    // parse last buffer
    parsedPHPArr, err := parsePHPInlineTableFieldValue(buffer)
    if err != nil {
        return nil, err
    }
    // phpArr.AddChild(strconv.Itoa(keyPos), NewPHPArrayValue(parsedPHPArr))
    phpArr.MergeChilds(parsedPHPArr)
    // keyPos++
    return phpArr, nil
}

// parseInlineTableFieldValue 解析键值对内容
func parsePHPInlineTableFieldValue(snippet string) (*PHPArray, error) {
    pos := strings.Index(snippet, "=")
    if pos <= 0 {
        return nil, errors.New("[split] invalid inline toml table data: " + snippet)
    }
    field := strings.TrimSpace(snippet[0:pos])
    val := strings.TrimSpace(snippet[pos+1:])
    phpVal, err := parsePHPValue(val)
    if err != nil {
        return nil, errors.New("[parse] invalid inline toml table data: " + val)
    }
    phpArr := NewPHPArray()
    phpArr.AddDeepValue([]string{field}, phpVal)
    return phpArr, nil
}

// parsePHPKeyValue 解析键值对
func parsePHPKeyValue(phpArr *PHPArray, key, val string) error {
    openQuote := false
    openDoubleQuote := false
    buffer := ""

    keyChars := []rune(key)
    keyCharsSize := len(keyChars)

    recurseKeys := make([]string, 0)

    for i := 0; i < keyCharsSize; i++ {
        // Handle quoting
        if keyChars[i] == '"' {
            if !openQuote {
                openDoubleQuote = !openDoubleQuote
                continue
            }
        }
        if keyChars[i] == '\'' {
            if !openDoubleQuote {
                openQuote = !openQuote
                continue
            }
        }

        // Handle dotted keys
        if keyChars[i] == '.' && !openQuote && !openDoubleQuote {
            recurseKeys = append(recurseKeys, buffer)
            buffer = ""
            continue
        }

        buffer += string(keyChars[i])
    }

    if buffer != "" {
        recurseKeys = append(recurseKeys, buffer)
        buffer = ""
    }

    phpVal, err := parsePHPValue(val)
    if err != nil {
        return err
    }
    
    if len(recurseKeys) > 0 {
        phpArr.AddDeepValue(recurseKeys, phpVal)
    }
    return nil
}

