package conf

import "testing"

func TestLoadConf(t *testing.T) {
	c := &PolyConf{}
	err := LoadConf(c, ".conf_ex/poly.json")
	if err != nil {
		t.Fatal(err)
	}
}