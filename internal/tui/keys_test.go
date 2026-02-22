package tui

import "testing"

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	if len(km.NextInput.Keys()) == 0 {
		t.Error("NextInput has no keys")
	}
	if len(km.PrevInput.Keys()) == 0 {
		t.Error("PrevInput has no keys")
	}
	if len(km.Quit.Keys()) == 0 {
		t.Error("Quit has no keys")
	}
	if len(km.PrevTab.Keys()) == 0 {
		t.Error("PrevTab has no keys")
	}
	if len(km.NextTab.Keys()) == 0 {
		t.Error("NextTab has no keys")
	}
	if len(km.Export.Keys()) == 0 {
		t.Error("Export has no keys")
	}
	if len(km.Help.Keys()) == 0 {
		t.Error("Help has no keys")
	}
}
