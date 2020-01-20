package toml2php

import (
    "bytes"
    "strings"

    "toml2php/util"
)

// define php data type
const (
    PhpTypeNumber = iota
    PhpTypeBoolean
    PhpTypeString
    PhpTypeValue
    PhpTypeArray
)


type PHPStruct struct {

}

// PHPValue define a php value
type PHPValue struct {
    Value       interface{}
    Type        int  // indicate the value type, can be int,float,string,bool,array
}

type PHPKey struct {
    Value       string
    IsNumeric   bool  // 键值是否是数字
}

type PHPKeyValuePair struct {
    Key         string
    Value       interface{}
    Type        int
}

// PHPArray define a php array （array & map）
type PHPArray struct {
    // Childs          map[string]*PHPArray
    // Key             *PHPKey
    // Value           *PHPValue
    Values          []*PHPKeyValuePair
}

func NewPHPArray() *PHPArray {
    return &PHPArray{
        // Childs:make([]*PHPArray, 0),
        Values:make([]*PHPKeyValuePair, 0),
    }
}

// 定义一个数组值
func NewPHPValue() *PHPValue {
    return &PHPValue{}
}

// NewPHPBoolValue create a PHP boolean value
func NewPHPBoolValue(val string) *PHPValue {
    return &PHPValue{
        Value:val,
        Type:PhpTypeBoolean,
    }
}

// NewPHPNumberValue create a PHP number value
func NewPHPNumberValue(val string) *PHPValue {
    return &PHPValue{
        Value:val,
        Type:PhpTypeNumber,
    }
}

// NewPHPStringValue create a PHP boolean value
func NewPHPStringValue(val string) *PHPValue {
    return &PHPValue{
        Value:val,
        Type:PhpTypeString,
    }
}

// NewPHPArrayValue create a PHP boolean value
func NewPHPArrayValue(val *PHPArray) *PHPValue {
    return &PHPValue{
        Value:val,
        Type:PhpTypeArray,
    }
}

// String format PHPValue as php code
func (phpVal *PHPValue) String(depth int) string {
    switch phpVal.Type {
    case PhpTypeBoolean:
        return util.NewValue(phpVal.Value).String()
    case PhpTypeNumber:
        return util.NewValue(phpVal.Value).String()
    case PhpTypeString:
        return fmtPhpString(util.NewValue(phpVal.Value).String())
    case PhpTypeArray:
        phpArr := phpVal.Value.(*PHPArray)
        return phpArr.String(depth)
    case PhpTypeValue:
        phpVal := phpVal.Value.(*PHPValue)
        return phpVal.String(depth)
    }
    return ""
}

func (phpKV *PHPKeyValuePair) GetValue(depth int) string {
    switch phpKV.Type {
    case PhpTypeBoolean:
        return util.NewValue(phpKV.Value).String()
    case PhpTypeNumber:
        return util.NewValue(phpKV.Value).String()
    case PhpTypeString:
        return fmtPhpString(util.NewValue(phpKV.Value).String())
    case PhpTypeArray:
        phpArr := phpKV.Value.(*PHPArray)
        return phpArr.String(depth)
    case PhpTypeValue:
        phpVal := phpKV.Value.(*PHPValue)
        return phpVal.String(depth)
    }
    return ""
}

func (phpKV *PHPKeyValuePair) String(depth int) string {
    buf := bytes.Buffer{}
    buf.WriteString(strings.Repeat("\t", depth))
    if isPositiveIntNumeric(phpKV.Key) {
        buf.WriteString(phpKV.Key)
    } else {
        buf.WriteString(fmtPhpString(phpKV.Key))
    }
    buf.WriteString(" => ")
    buf.WriteString(phpKV.GetValue(depth))
    return buf.String()
}

func NewNumberKey(v string) *PHPKey {
    return &PHPKey{
        Value:v,
        IsNumeric:true,
    }
}

func NewStringKey(v string) *PHPKey {
    return &PHPKey{
        Value:v,
        IsNumeric:false,
    }
}

func (phpKey *PHPKey) String() string {
    if phpKey.IsNumeric {
        return phpKey.Value
    }
    return fmtPhpString(phpKey.Value)
}

// AddRecurseKeys add/initialize recursed keys
func (phpArr *PHPArray) AddRecurseKeys(fields []string) {
    fieldsSize := len(fields)
    if fieldsSize == 0 {
        return
    }
    refPhpArr := phpArr
    for i := 0; i < fieldsSize; i++ {
        field := fields[i]
        found := false
        for _, v := range refPhpArr.Values {
            if v.Key == field && v.Type == PhpTypeArray {
                refPhpArr = v.Value.(*PHPArray)
                found = true
            }
        }
        if !found {
            arr := NewPHPArray()
            kvPair := &PHPKeyValuePair{
                Key:field,
                Type:PhpTypeArray,
                Value:arr,
            }
            refPhpArr.Values = append(refPhpArr.Values, kvPair)
            refPhpArr = arr
        }
    }
}

// AddDeepValue add value for specified key, which may be in a deep length
func (phpArr *PHPArray) AddDeepValue(paths []string, key string, val *PHPValue) {
    pathSize := len(paths)
    refPhpArr := phpArr
    // 为0，表示在当前节点添加
    if pathSize == 0 {
        phpArr.AddChild(key, val)
        return
    }
    // 深度添加
    var refKVPair *PHPKeyValuePair
    for i := 0; i < pathSize; i++ {
        field := paths[i]
        found := false
        for _, v := range refPhpArr.Values {
            if v.Key == field && v.Type == PhpTypeArray {
                refPhpArr = v.Value.(*PHPArray)
                found = true
                if i == pathSize - 1 {
                    refKVPair = v
                }
            }
        }
        if !found {
            arr := NewPHPArray()
            kvPair := &PHPKeyValuePair{
                Key:field,
            }
            refPhpArr.Values = append(refPhpArr.Values, kvPair)
            refPhpArr = arr
            if i == pathSize - 1 {
                refKVPair = kvPair
            }
        }
    }
    refKVPair.Value = val
    refKVPair.Type = PhpTypeValue
}

func (phpArr *PHPArray) AddChild(key string, val *PHPValue) {
    kvPair := &PHPKeyValuePair{
        Key:key,
        Type:PhpTypeValue,
        Value:val,
    }
    phpArr.Values = append(phpArr.Values, kvPair)
}

func (phpArr *PHPArray) MergeChilds(arr *PHPArray) {
    if arr == nil || len(arr.Values) == 0 {
        return
    }
    for _, v := range arr.Values {
        found := false
        for _, ov := range phpArr.Values {
            if v.Key == ov.Key {
                ov.Value = v.Value
                found = true
                break
            }
        }
        if !found {
            phpArr.Values = append(phpArr.Values, v)
        }
    }
}

func (phpArr *PHPArray) String(depth int) string {
    result := bytes.Buffer{}
    result.WriteString("[")
    valSize := len(phpArr.Values)
    if phpArr.Values != nil && valSize > 0 {
        result.WriteString("\n")
        for i, kv := range phpArr.Values {
            result.WriteString("\t")
            result.WriteString(kv.String(depth+1))
            if i != valSize - 1 {
                result.WriteString(",")
            }
            result.WriteString("\n")
        }
        result.WriteString(strings.Repeat("\t", depth))
    }
    result.WriteString("]")
    return result.String()
}



