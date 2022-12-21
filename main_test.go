package main

import (
	"testing"

	"vec/model"
	"vec/processor"
)

func TestQuery(t *testing.T) {
	process := processor.NewStrToVec("name", "http://10.100.29.62:8789/str2vec")
	patent := model.Patent{Name: "橡胶填料高分散碳纳米管"}
	vec := process.ToVec(&patent)
	process.SaveVec(vec)
}
