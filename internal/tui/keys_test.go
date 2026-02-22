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
}

func TestShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	bindings := km.ShortHelp()

	if len(bindings) == 0 {
		t.Error("ShortHelp returned empty")
	}
}

func TestFullHelp(t *testing.T) {
	km := DefaultKeyMap()
	groups := km.FullHelp()

	if len(groups) == 0 {
		t.Error("FullHelp returned empty")
	}
	for i, group := range groups {
		if len(group) == 0 {
			t.Errorf("FullHelp group %d is empty", i)
		}
	}
}
