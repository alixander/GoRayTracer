package main

import "testing"

func TestBasic(t *testing.T) {
    if (emptyVector().X != 0) {
        t.Error("Failed")
    }
}
