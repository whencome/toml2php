package toml2php

// SetIndent 设置缩进字符
func SetIndent(indent string) {
	IndentString = indent
}

// ParseSingle 解析单个值，如整数、浮点数、字符串、布尔值等等
func ParseSingle(snippet string) (string, error) {
	// normalize the toml string
	toml, err := normalize(snippet)
	if err != nil {
		return "", err
	}
	phpVal, err := parsePHPValue(toml)
	if err != nil {
		return "", err
	}
	return phpVal.String(0), nil
}

// ParseTable 解析数组
func ParseTable(snippet string) (string, error) {
	phpArr, err := parse(snippet)
	if err != nil {
		return "", err
	}
	return phpArr.String(0), nil
}
