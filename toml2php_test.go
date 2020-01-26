package toml2php

import (
    "io/ioutil"
    "os"
    "testing"
)

func TestNormalize(t *testing.T) {
    tomlFile := "example.toml"
    file, err := os.Open(tomlFile)
    if err != nil {
        t.Logf("open file %s failed \n", tomlFile)
        t.Fail()
    }
    defer file.Close()

    tomlBytes, err := ioutil.ReadAll(file)
    if err != nil {
        t.Log("read file content failed\n")
        t.Fail()
    }

    toml, err := normalize(string(tomlBytes))
    if err != nil {
        t.Logf("normalize toml failed: %s\n", err)
        t.Fail()
    }

    t.Log(toml)
}

func TestParseArray(t *testing.T) {
    tomlArr := `[ 'literal,', 'strings', 'quo"ted' ]`
    tomlRunes := []rune(tomlArr)
    parsed, err := parsePHPArray(tomlRunes)
    if err != nil {
        t.Logf("parsePHPArray failed: %s \n", err)
        t.Fail()
    }
    t.Log("parsed: ", parsed.String(0))
}

func TestParseInlineTable(t *testing.T) {
    tomlInlineTable := `PR17 = [
  {title = "Home", url = "/", childs = []},
  {title = "Games", url = "/games", childs = [{title = "Game A", url = "/games/game-a", childs = []}, {title = "Game B", url = "/games/game-b", childs = []}]},
  {title = "About us", url = "/about", childs = []}
]`
    parsed, err := parsePHPInlineTableFieldValue(tomlInlineTable)
    if err != nil {
        t.Logf("parsePHPInlineTable failed: %s \n", err)
        t.Fail()
        return
    }
    t.Log("parsed: ", parsed.String(0))
}

func TestParsePHPKeyValue(t *testing.T) {
    phpArr := &PHPArray{}
    key := "PR17"
    val := `[
  {title = "Home", url = "/", childs = []},
  {title = "Games", url = "/games", childs = [{title = "Game A", url = "/games/game-a", childs = []}, {title = "Game B", url = "/games/game-b", childs = []}]},
  {title = "About us", url = "/about", childs = []}
]`
    err := parsePHPKeyValue(phpArr, key, val)
    if err != nil {
        t.Logf("parsePHPKeyValue failed: %s \n", err)
        t.Fail()
        return
    }
    t.Log("parsed: ", phpArr.String(0))
}

func TestParseMultiLine(t *testing.T) {
    toml := `key3 = """
One
Two"""`
    rs, err := parse(toml)
    if err != nil {
        t.Logf("parse %s failed: %s\n", toml, err)
        t.Fail()
        return
    }
    t.Logf("parse %s success: %s\n", toml, rs.String(0))
}

func TestParseSingle(t *testing.T) {
    var tomls = []string{
        // number
        `123554`,
        `23.4056`,
        `0.2399`,
        // boolean
        `true`,
        `false`,
        // string
        `"good"`,
        `"hello,world"`,
    }
    for _, toml := range tomls {
        rs, err := ParseSingle(toml)
        if err != nil {
            t.Logf("parse %s failed: %s\n", toml, err)
            t.Fail()
            continue
        }
        t.Logf("parse %s success: %s\n", toml, rs)
    }
}

func TestParseTable(t *testing.T) {
    tomlFile := "example.toml"
    file, err := os.Open(tomlFile)
    if err != nil {
        t.Logf("open file %s failed \n", tomlFile)
        t.Fail()
    }
    defer file.Close()

    tomlBytes, err := ioutil.ReadAll(file)
    if err != nil {
        t.Log("read file content failed\n")
        t.Fail()
    }

    rs, err := ParseTable(string(tomlBytes))
    if err != nil {
        t.Logf("parse table failed: %s\n", err)
        t.Fail()
    }

    t.Logf("\n=======================\n")

    t.Log(rs)
}
