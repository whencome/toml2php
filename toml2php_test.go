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
    tomlInlineTable := `{title = "Games", url = "/games", childs = [{title = "Game A", url = "/games/game-a", childs = []}, {title = "Game B", url = "/games/game-b", childs = []}]}`
    tomlRunes := []rune(tomlInlineTable)
    parsed, err := parsePHPInlineTable(tomlRunes)
    if err != nil {
        t.Logf("parsePHPInlineTable failed: %s \n", err)
        t.Fail()
    }
    t.Log("parsed: ", parsed.String(0))
}
