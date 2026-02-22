package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/josegonzalez/aws-eks-calculator/internal/calculator"
	"github.com/josegonzalez/aws-eks-calculator/internal/prefs"
	"github.com/josegonzalez/aws-eks-calculator/internal/pricing"
)

// newReadyModel returns a NewModel with ratesLoading cleared and already
// in the calculator view, simulating the state after the initial pricing
// fetch completes and the user selected a capability.
func newReadyModel() Model {
	m := NewModel()
	m.ratesLoading = false
	m.view = viewCalculator
	return m
}

func TestNewModel(t *testing.T) {
	prefs.SetDir(t.TempDir())
	defer prefs.SetDir("")

	m := NewModel()

	if len(m.capStates) != 3 {
		t.Errorf("expected 3 capability states, got %d", len(m.capStates))
	}
	argoState := m.capStates[calculator.CapabilityArgoCD]
	if len(argoState.Inputs) != 9 {
		t.Errorf("ArgoCD: expected 9 inputs, got %d", len(argoState.Inputs))
	}
	ackState := m.capStates[calculator.CapabilityACK]
	if len(ackState.Inputs) != 7 {
		t.Errorf("ACK: expected 7 inputs, got %d", len(ackState.Inputs))
	}
	kroState := m.capStates[calculator.CapabilityKro]
	if len(kroState.Inputs) != 7 {
		t.Errorf("kro: expected 7 inputs, got %d", len(kroState.Inputs))
	}

	if m.view != viewCapabilitySelector {
		t.Errorf("expected viewCapabilitySelector, got %d", m.view)
	}
	if m.pricingRegion != "us-east-1" {
		t.Errorf("expected default pricingRegion us-east-1, got %q", m.pricingRegion)
	}
	if len(m.allRegions) == 0 {
		t.Error("no regions loaded")
	}
	if !m.ratesLoading {
		t.Error("ratesLoading should be true on new model")
	}
}

func TestInit(t *testing.T) {
	m := NewModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a batch command")
	}
}

func TestUpdateWindowSize(t *testing.T) {
	m := NewModel()
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := updated.(Model)
	if model.width != 120 || model.height != 40 {
		t.Errorf("expected 120x40, got %dx%d", model.width, model.height)
	}
}

func TestUpdateClearExportMsg(t *testing.T) {
	m := NewModel()
	m.exportMsg = "test"
	updated, _ := m.Update(clearExportMsg{})
	model := updated.(Model)
	if model.exportMsg != "" {
		t.Errorf("expected empty exportMsg, got %q", model.exportMsg)
	}
}

func TestUpdatePricingMsg(t *testing.T) {
	m := NewModel()
	rates := pricing.Rates{
		ArgoCDBasePerHour:   0.03,
		ArgoCDAppPerHour:    0.002,
		ACKBasePerHour:      0.03,
		ACKResourcePerHour:  0.002,
		KroBasePerHour:      0.03,
		KroRGDPerHour:       0.002,
		FargateVCPUPerHour:  0.05,
		FargateMemGBPerHour: 0.005,
	}

	updated, cmd := m.Update(pricingMsg{rates: rates})
	model := updated.(Model)

	if !model.ratesLoaded {
		t.Error("ratesLoaded should be true")
	}
	if model.ratesLoading {
		t.Error("ratesLoading should be false after pricing msg")
	}
	if model.rates.ArgoCDBasePerHour != 0.03 {
		t.Errorf("expected ArgoCDBasePerHour 0.03, got %f", model.rates.ArgoCDBasePerHour)
	}

	// ArgoCD Fargate inputs are at indices 7, 8
	argoState := model.capStates[calculator.CapabilityArgoCD]
	if argoState.Inputs[7].Value() != "0.0500" {
		t.Errorf("expected Fargate vCPU input updated to 0.0500, got %q", argoState.Inputs[7].Value())
	}
	if argoState.Inputs[8].Value() != "0.0050" {
		t.Errorf("expected Fargate mem input updated to 0.0050, got %q", argoState.Inputs[8].Value())
	}

	// ACK Fargate inputs are at indices 5, 6
	ackState := model.capStates[calculator.CapabilityACK]
	if ackState.Inputs[5].Value() != "0.0500" {
		t.Errorf("expected ACK Fargate vCPU input updated to 0.0500, got %q", ackState.Inputs[5].Value())
	}

	if cmd == nil {
		t.Error("first successful pricing msg should return cache warm command")
	}
	if !model.cacheWarmed {
		t.Error("cacheWarmed should be true after first success")
	}
}

func TestUpdatePricingMsgSuccess(t *testing.T) {
	m := NewModel()
	defaults := pricing.DefaultRates()

	updated, cmd := m.Update(pricingMsg{rates: defaults, err: nil})
	model := updated.(Model)

	if model.ratesErr != nil {
		t.Errorf("expected nil error, got %v", model.ratesErr)
	}
	if !model.ratesLoaded {
		t.Error("ratesLoaded should be true after successful fetch")
	}
	if cmd == nil {
		t.Error("first successful pricing msg should return cache warm command")
	}
}

func TestUpdatePricingMsgError(t *testing.T) {
	m := NewModel()
	// First do a successful fetch
	goodRates := pricing.Rates{
		ArgoCDBasePerHour:   0.03,
		ArgoCDAppPerHour:    0.002,
		ACKBasePerHour:      0.03,
		ACKResourcePerHour:  0.002,
		KroBasePerHour:      0.03,
		KroRGDPerHour:       0.002,
		FargateVCPUPerHour:  0.05,
		FargateMemGBPerHour: 0.005,
	}
	updated, _ := m.Update(pricingMsg{rates: goodRates})
	m = updated.(Model)

	// Now send a failed fetch with default rates
	badRates := pricing.DefaultRates()
	fetchErr := errors.New("network timeout")
	updated, _ = m.Update(pricingMsg{rates: badRates, err: fetchErr})
	model := updated.(Model)

	// Rates are always applied now — msg.rates are default rates (safe fallback)
	if model.rates.ArgoCDBasePerHour != badRates.ArgoCDBasePerHour {
		t.Errorf("expected rates applied from msg, got %f", model.rates.ArgoCDBasePerHour)
	}
	if model.ratesErr == nil {
		t.Error("ratesErr should be set")
	}
	if !model.ratesLoaded {
		t.Error("ratesLoaded should stay true")
	}
}

func TestUpdatePricingMsgErrorFirstFetch(t *testing.T) {
	m := NewModel()
	defaults := pricing.DefaultRates()

	fetchErr := errors.New("no credentials")
	updated, _ := m.Update(pricingMsg{rates: defaults, err: fetchErr})
	model := updated.(Model)

	if model.ratesLoaded {
		t.Error("ratesLoaded should stay false on first fetch failure")
	}
	if model.ratesLoading {
		t.Error("ratesLoading should be false after pricing msg with error")
	}
	if model.ratesErr == nil {
		t.Error("ratesErr should be set")
	}
	// Rates should still be applied even on error (they contain safe defaults)
	if model.rates.ArgoCDBasePerHour != defaults.ArgoCDBasePerHour {
		t.Errorf("rates should be applied even on error, got %f", model.rates.ArgoCDBasePerHour)
	}
}

func TestUpdatePricingMsgPartialError(t *testing.T) {
	m := NewModel()
	// Simulate partial success: ArgoCD live rates fetched, ACK/kro used defaults
	partialRates := pricing.Rates{
		ArgoCDBasePerHour:   0.03,
		ArgoCDAppPerHour:    0.0015,
		ACKBasePerHour:      0.02771, // defaults
		ACKResourcePerHour:  0.00136, // defaults
		KroBasePerHour:      0.02771, // defaults
		KroRGDPerHour:       0.00136, // defaults
		FargateVCPUPerHour:  0.05,
		FargateMemGBPerHour: 0.005,
	}
	partialErr := errors.New("EKS ACK: zero rates for region us-east-1, using defaults")

	updated, _ := m.Update(pricingMsg{rates: partialRates, err: partialErr})
	model := updated.(Model)

	// Rates should be applied despite the error
	if model.rates.ArgoCDBasePerHour != 0.03 {
		t.Errorf("expected ArgoCDBasePerHour 0.03, got %f", model.rates.ArgoCDBasePerHour)
	}
	if model.rates.ACKBasePerHour != 0.02771 {
		t.Errorf("expected ACKBasePerHour default 0.02771, got %f", model.rates.ACKBasePerHour)
	}
	// Fargate rates should be applied to inputs
	argoState := model.capStates[calculator.CapabilityArgoCD]
	if argoState.Inputs[7].Value() != "0.0500" {
		t.Errorf("expected Fargate vCPU input updated, got %q", argoState.Inputs[7].Value())
	}
	// Error should be set
	if model.ratesErr == nil {
		t.Error("ratesErr should be set for partial error")
	}
	// ratesLoaded should not be set (error path)
	if model.ratesLoaded {
		t.Error("ratesLoaded should not be true on partial error")
	}
}

func TestUpdateFallthrough(t *testing.T) {
	m := newReadyModel()
	m.view = viewHelp
	updated, _ := m.Update(tea.MouseMsg{})
	model := updated.(Model)
	if model.view != viewHelp {
		t.Error("view should not change on unhandled msg")
	}
}

func TestUpdateInputForwarding(t *testing.T) {
	m := newReadyModel()
	m.view = viewCalculator
	updated, _ := m.Update(tea.MouseMsg{})
	_ = updated.(Model) // should not panic
}

// Capability selector tests

func TestSelectorKeysNavigation(t *testing.T) {
	m := NewModel()
	m.ratesLoading = false

	// Down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(Model)
	if model.capSelectorCursor != 1 {
		t.Errorf("down should move cursor to 1, got %d", model.capSelectorCursor)
	}

	// Down with j
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(Model)
	if model.capSelectorCursor != 2 {
		t.Errorf("j should move cursor to 2, got %d", model.capSelectorCursor)
	}

	// Down at bottom should stay
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	if model.capSelectorCursor != 2 {
		t.Errorf("down at bottom should stay at 2, got %d", model.capSelectorCursor)
	}

	// Up
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(Model)
	if model.capSelectorCursor != 1 {
		t.Errorf("up should move cursor to 1, got %d", model.capSelectorCursor)
	}

	// Up with k
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(Model)
	if model.capSelectorCursor != 0 {
		t.Errorf("k should move cursor to 0, got %d", model.capSelectorCursor)
	}

	// Up at top should stay
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(Model)
	if model.capSelectorCursor != 0 {
		t.Errorf("up at top should stay at 0, got %d", model.capSelectorCursor)
	}
}

func TestSelectorEnter(t *testing.T) {
	m := NewModel()
	m.ratesLoading = false
	m.capSelectorCursor = 1 // ACK

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.view != viewCalculator {
		t.Error("enter should switch to viewCalculator")
	}
	if model.activeCapability != calculator.CapabilityACK {
		t.Errorf("expected ACK, got %v", model.activeCapability)
	}
}

func TestSelectorQuit(t *testing.T) {
	m := NewModel()
	m.ratesLoading = false

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model := updated.(Model)
	if !model.quitting {
		t.Error("q should quit from selector")
	}
	if cmd == nil {
		t.Error("q should return tea.Quit")
	}
}

func TestSelectorCtrlC(t *testing.T) {
	m := NewModel()
	m.ratesLoading = false

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(Model)
	if !model.quitting {
		t.Error("ctrl+c should quit from selector")
	}
	if cmd == nil {
		t.Error("ctrl+c should return tea.Quit")
	}
}

func TestSelectorUnhandled(t *testing.T) {
	m := NewModel()
	m.ratesLoading = false

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model := updated.(Model)
	if model.view != viewCapabilitySelector {
		t.Error("unhandled key should not change view")
	}
}

// Calculator tests

func TestCalculatorKeysCtrlC(t *testing.T) {
	m := newReadyModel()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(Model)
	if !model.quitting {
		t.Error("ctrl+c should set quitting")
	}
	if cmd == nil {
		t.Error("ctrl+c should return tea.Quit")
	}
}

func TestCalculatorKeysTab(t *testing.T) {
	m := newReadyModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	model := updated.(Model)
	cs := model.activeState()
	if cs.FocusIndex != 1 {
		t.Errorf("tab should advance focus to 1, got %d", cs.FocusIndex)
	}
}

func TestCalculatorKeysDown(t *testing.T) {
	m := newReadyModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(Model)
	cs := model.activeState()
	if cs.FocusIndex != 1 {
		t.Errorf("down should advance focus to 1, got %d", cs.FocusIndex)
	}
}

func TestCalculatorKeysShiftTab(t *testing.T) {
	m := newReadyModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model := updated.(Model)
	cs := model.activeState()
	// Wraps from 0 to last
	expectedLast := len(cs.Inputs) - 1
	if cs.FocusIndex != expectedLast {
		t.Errorf("shift+tab should wrap to %d, got %d", expectedLast, cs.FocusIndex)
	}
}

func TestCalculatorKeysUp(t *testing.T) {
	m := newReadyModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	model := updated.(Model)
	cs := model.activeState()
	expectedLast := len(cs.Inputs) - 1
	if cs.FocusIndex != expectedLast {
		t.Errorf("up should wrap to %d, got %d", expectedLast, cs.FocusIndex)
	}
}

func TestCalculatorKeysRegions(t *testing.T) {
	m := newReadyModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model := updated.(Model)
	if model.view != viewRegions {
		t.Error("r should switch to viewRegions")
	}
	if model.regionCursor != 0 {
		t.Errorf("region cursor should be 0, got %d", model.regionCursor)
	}
}

func TestCalculatorKeysHelp(t *testing.T) {
	m := newReadyModel()
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model := updated.(Model)
	if model.view != viewHelp {
		t.Error("? should switch to viewHelp")
	}
}

func TestCalculatorKeysExport(t *testing.T) {
	m := newReadyModel()
	m.exportDir = t.TempDir()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model := updated.(Model)
	if model.exportMsg == "" {
		t.Error("e should set exportMsg")
	}
	if cmd == nil {
		t.Error("e should return a tick command")
	}
}

func TestCalculatorKeysQWithFocus(t *testing.T) {
	m := newReadyModel()
	// 'q' should always quit, even with a focused input (inputs are numeric-only)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model := updated.(Model)
	if !model.quitting {
		t.Error("q should quit even with focused input")
	}
	if cmd == nil {
		t.Error("q should return tea.Quit")
	}
}

func TestCalculatorKeysPassthrough(t *testing.T) {
	m := newReadyModel()
	// Type a digit into the focused input
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	model := updated.(Model)
	cs := model.activeState()
	if !strings.Contains(cs.Inputs[0].Value(), "3") {
		t.Error("digit should be forwarded to focused input")
	}
}

func TestCalculatorTabSwitching(t *testing.T) {
	m := newReadyModel()

	// Start at ArgoCD, press ] to go to ACK
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	model := updated.(Model)
	if model.activeCapability != calculator.CapabilityACK {
		t.Errorf("] from ArgoCD should switch to ACK, got %v", model.activeCapability)
	}

	// Press ] to go to kro
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	model = updated.(Model)
	if model.activeCapability != calculator.CapabilityKro {
		t.Errorf("] from ACK should switch to kro, got %v", model.activeCapability)
	}

	// Press [ to go back to ACK
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	model = updated.(Model)
	if model.activeCapability != calculator.CapabilityACK {
		t.Errorf("[ from kro should switch to ACK, got %v", model.activeCapability)
	}

	// Press [ to go back to ArgoCD
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	model = updated.(Model)
	if model.activeCapability != calculator.CapabilityArgoCD {
		t.Errorf("[ from ACK should switch to ArgoCD, got %v", model.activeCapability)
	}
}

func TestCalculatorBracketWrapsAround(t *testing.T) {
	m := newReadyModel()

	// [ from ArgoCD (first) should wrap to kro (last)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}})
	model := updated.(Model)
	if model.activeCapability != calculator.CapabilityKro {
		t.Errorf("[ from ArgoCD should wrap to kro, got %v", model.activeCapability)
	}

	// ] from kro (last) should wrap to ArgoCD (first)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	model = updated.(Model)
	if model.activeCapability != calculator.CapabilityArgoCD {
		t.Errorf("] from kro should wrap to ArgoCD, got %v", model.activeCapability)
	}
}

func TestCalculatorInputsPersistAcrossTabs(t *testing.T) {
	m := newReadyModel()

	// Change a value in ArgoCD
	cs := m.activeState()
	cs.Inputs[0].SetValue("5")

	// Switch to ACK
	m.switchCapability(calculator.CapabilityACK)

	// Switch back to ArgoCD
	m.switchCapability(calculator.CapabilityArgoCD)

	cs = m.activeState()
	if cs.Inputs[0].Value() != "5" {
		t.Errorf("ArgoCD input should persist, got %q", cs.Inputs[0].Value())
	}
}

// Region picker tests

func TestRegionPickerNavigation(t *testing.T) {
	m := newReadyModel()
	m.view = viewRegions

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(Model)
	if model.regionCursor != 1 {
		t.Errorf("down should move cursor to 1, got %d", model.regionCursor)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(Model)
	if model.regionCursor != 2 {
		t.Errorf("j should move cursor to 2, got %d", model.regionCursor)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(Model)
	if model.regionCursor != 1 {
		t.Errorf("up should move cursor to 1, got %d", model.regionCursor)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(Model)
	if model.regionCursor != 0 {
		t.Errorf("k should move cursor to 0, got %d", model.regionCursor)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(Model)
	if model.regionCursor != 0 {
		t.Errorf("up at top should stay at 0, got %d", model.regionCursor)
	}
}

func TestRegionPickerDownAtBottom(t *testing.T) {
	m := newReadyModel()
	m.view = viewRegions
	m.regionCursor = len(m.allRegions) - 1

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(Model)
	if model.regionCursor != len(m.allRegions)-1 {
		t.Error("down at bottom should stay at last")
	}
}

func TestRegionPickerEsc(t *testing.T) {
	m := newReadyModel()
	m.view = viewRegions
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)
	if model.view != viewCalculator {
		t.Error("esc should close region picker")
	}
}

func TestRegionPickerQ(t *testing.T) {
	m := newReadyModel()
	m.view = viewRegions
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model := updated.(Model)
	if model.view != viewCalculator {
		t.Error("q should close region picker")
	}
}

func TestRegionPickerSelectDifferentRegion(t *testing.T) {
	prefs.SetDir(t.TempDir())
	defer prefs.SetDir("")

	m := newReadyModel()
	m.view = viewRegions
	m.regionCursor = 1

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.view != viewCalculator {
		t.Error("enter should return to calculator")
	}
	if model.pricingRegion != "us-east-2" {
		t.Errorf("expected pricingRegion us-east-2, got %q", model.pricingRegion)
	}
	if cmd == nil {
		t.Error("selecting a different region should trigger a pricing fetch")
	}
	if !model.ratesLoading {
		t.Error("ratesLoading should be true after selecting a different region")
	}
}

func TestRegionPickerSelectSameRegion(t *testing.T) {
	dir := t.TempDir()
	prefs.SetDir(dir)
	defer prefs.SetDir("")

	m := newReadyModel()
	m.view = viewRegions
	m.regionCursor = 0

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.view != viewCalculator {
		t.Error("enter should return to calculator")
	}
	if model.pricingRegion != "us-east-1" {
		t.Errorf("expected pricingRegion us-east-1, got %q", model.pricingRegion)
	}
	if cmd != nil {
		t.Error("selecting the same region should not trigger a pricing fetch")
	}
}

func TestRegionPickerUnhandled(t *testing.T) {
	m := newReadyModel()
	m.view = viewRegions
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model := updated.(Model)
	if model.view != viewRegions {
		t.Error("unhandled key should not change view")
	}
}

// Help tests

func TestHelpKeysEsc(t *testing.T) {
	m := newReadyModel()
	m.view = viewHelp
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(Model)
	if model.view != viewCalculator {
		t.Error("esc should close help")
	}
}

func TestHelpKeysQuestion(t *testing.T) {
	m := newReadyModel()
	m.view = viewHelp
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model := updated.(Model)
	if model.view != viewCalculator {
		t.Error("? should close help")
	}
}

func TestHelpKeysQ(t *testing.T) {
	m := newReadyModel()
	m.view = viewHelp
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model := updated.(Model)
	if model.view != viewCalculator {
		t.Error("q should close help")
	}
}

func TestHelpKeysCtrlC(t *testing.T) {
	m := newReadyModel()
	m.view = viewHelp
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model := updated.(Model)
	if !model.quitting {
		t.Error("ctrl+c should quit")
	}
	if cmd == nil {
		t.Error("ctrl+c should return tea.Quit")
	}
}

func TestHelpKeysUnhandled(t *testing.T) {
	m := newReadyModel()
	m.view = viewHelp
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model := updated.(Model)
	if model.view != viewHelp {
		t.Error("unhandled key should not change view")
	}
}

// View tests

func TestViewCalculator(t *testing.T) {
	m := newReadyModel()
	m.width = 120
	m.height = 40
	output := m.View()

	if !strings.Contains(output, "AWS EKS Capabilities Cost Calculator") {
		t.Error("missing title")
	}
	if !strings.Contains(output, "navigate") {
		t.Error("missing help bar")
	}
	if !strings.Contains(output, "[/]") {
		t.Error("missing capability switching hint")
	}
}

func TestViewCapabilitySelector(t *testing.T) {
	m := NewModel()
	m.ratesLoading = false
	output := m.View()

	if !strings.Contains(output, "Select EKS Capability") {
		t.Error("missing capability selector")
	}
	if !strings.Contains(output, "enter select") {
		t.Error("missing selector hints")
	}
}

func TestViewHelp(t *testing.T) {
	m := newReadyModel()
	m.view = viewHelp
	output := m.View()

	if !strings.Contains(output, "Keyboard Shortcuts") {
		t.Error("missing help content")
	}
	if !strings.Contains(output, "esc back") {
		t.Error("missing help view hints")
	}
}

func TestViewRegions(t *testing.T) {
	m := newReadyModel()
	m.view = viewRegions
	output := m.View()

	if !strings.Contains(output, "Select Region") {
		t.Error("missing region picker")
	}
	if !strings.Contains(output, "us-east-1") {
		t.Error("missing default region in picker")
	}
	if !strings.Contains(output, "esc cancel") {
		t.Error("missing region picker hints")
	}
}

func TestViewQuitting(t *testing.T) {
	m := NewModel()
	m.quitting = true
	output := m.View()
	if output != "" {
		t.Errorf("quitting should return empty string, got %q", output)
	}
}

func TestViewWithExportMsg(t *testing.T) {
	m := newReadyModel()
	m.exportMsg = "Exported!"
	m.width = 120
	m.height = 40
	output := m.View()

	if !strings.Contains(output, "Exported!") {
		t.Error("export message not shown")
	}
}

func TestViewHintOutOfRange(t *testing.T) {
	m := newReadyModel()
	cs := m.activeState()
	cs.FocusIndex = -1
	m.width = 120
	m.height = 40
	// Should not panic
	_ = m.View()
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"0", 0},
		{"5", 5},
		{"100", 100},
		{"-1", 0},
		{"abc", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := parseInt(tt.input)
		if got != tt.want {
			t.Errorf("parseInt(%q): got %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"0", 0},
		{"1.5", 1.5},
		{"0.04048", 0.04048},
		{"-1.0", 0},
		{"abc", 0},
		{"", 0},
	}
	for _, tt := range tests {
		got := parseFloat(tt.input)
		if got != tt.want {
			t.Errorf("parseFloat(%q): got %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestDoExportSuccess(t *testing.T) {
	m := newReadyModel()
	m.exportDir = t.TempDir()
	result, cmd := m.doExport()
	if !strings.Contains(result.exportMsg, "Exported to") {
		t.Errorf("expected success message, got %q", result.exportMsg)
	}
	if !strings.Contains(result.exportMsg, "argocd-cost-estimate.csv") {
		t.Errorf("expected argocd filename, got %q", result.exportMsg)
	}
	if cmd == nil {
		t.Error("should return tick command")
	}
}

func TestDoExportACK(t *testing.T) {
	m := newReadyModel()
	m.activeCapability = calculator.CapabilityACK
	m.exportDir = t.TempDir()
	result, _ := m.doExport()
	if !strings.Contains(result.exportMsg, "ack-cost-estimate.csv") {
		t.Errorf("expected ack filename, got %q", result.exportMsg)
	}
}

func TestDoExportKro(t *testing.T) {
	m := newReadyModel()
	m.activeCapability = calculator.CapabilityKro
	m.exportDir = t.TempDir()
	result, _ := m.doExport()
	if !strings.Contains(result.exportMsg, "kro-cost-estimate.csv") {
		t.Errorf("expected kro filename, got %q", result.exportMsg)
	}
}

func TestDoExportError(t *testing.T) {
	m := newReadyModel()
	m.exportDir = "/nonexistent/path"
	result, cmd := m.doExport()
	if !strings.Contains(result.exportMsg, "Export failed") {
		t.Errorf("expected error message, got %q", result.exportMsg)
	}
	if cmd == nil {
		t.Error("should return tick command even on error")
	}
}



func TestExportPath(t *testing.T) {
	m := NewModel()
	if m.exportPath("test.csv") != "test.csv" {
		t.Error("empty exportDir should use filename directly")
	}
	m.exportDir = "/tmp/exports"
	if m.exportPath("test.csv") != "/tmp/exports/test.csv" {
		t.Errorf("got %q", m.exportPath("test.csv"))
	}
}

func TestClearExportTick(t *testing.T) {
	msg := clearExportTick(time.Time{})
	if _, ok := msg.(clearExportMsg); !ok {
		t.Errorf("expected clearExportMsg, got %T", msg)
	}
}

func TestHandleKeyPressUnknownView(t *testing.T) {
	m := newReadyModel()
	m.view = viewState(99) // unknown view
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = updated.(Model) // should not panic
}

func TestCalculatorExportViaKeybinding(t *testing.T) {
	m := newReadyModel()
	m.exportDir = t.TempDir()
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	model := updated.(Model)
	if !strings.Contains(model.exportMsg, "Exported to") {
		t.Errorf("expected success, got %q", model.exportMsg)
	}
	if cmd == nil {
		t.Error("should return tick command")
	}
}

func TestViewWithRatesError(t *testing.T) {
	m := newReadyModel()
	m.width = 120
	m.height = 40
	m.ratesLoaded = true
	m.ratesErr = errors.New("timeout")
	m.pricingRegion = "eu-west-1"
	output := m.View()

	if !strings.Contains(output, "Unable to fetch rates for eu-west-1") {
		t.Error("missing rates error warning")
	}
	if !strings.Contains(output, "previously fetched rates") {
		t.Error("should mention previously fetched rates when ratesLoaded is true")
	}
}

func TestViewWithRatesErrorFirstFetch(t *testing.T) {
	m := newReadyModel()
	m.width = 120
	m.height = 40
	m.ratesLoaded = false
	m.ratesErr = errors.New("no credentials")
	m.pricingRegion = "us-east-1"
	output := m.View()

	if !strings.Contains(output, "Unable to fetch rates for us-east-1") {
		t.Error("missing rates error warning")
	}
	if !strings.Contains(output, "default rates") {
		t.Error("should mention default rates when ratesLoaded is false")
	}
}

func TestViewNoWarningWhenRatesOk(t *testing.T) {
	m := newReadyModel()
	m.width = 120
	m.height = 40
	m.ratesLoaded = true
	m.ratesErr = nil
	output := m.View()

	if strings.Contains(output, "Unable to fetch rates") {
		t.Error("should not show warning when rates are OK")
	}
}

func TestViewLoading(t *testing.T) {
	prefs.SetDir(t.TempDir())
	defer prefs.SetDir("")

	m := NewModel()
	m.width = 120
	m.height = 40
	output := m.View()

	if !strings.Contains(output, "Loading rates for us-east-1") {
		t.Error("should show loading message")
	}
	if strings.Contains(output, "navigate") {
		t.Error("should not show calculator content while loading")
	}
}

func TestViewLoadingAfterRegionChange(t *testing.T) {
	prefs.SetDir(t.TempDir())
	defer prefs.SetDir("")

	m := newReadyModel()
	m.view = viewRegions
	m.regionCursor = 1 // us-east-2

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)
	model.width = 120
	model.height = 40
	output := model.View()

	if !strings.Contains(output, "Loading rates for us-east-2") {
		t.Error("should show loading message after region change")
	}
}

func TestViewLoadingClearedAfterPricingMsg(t *testing.T) {
	m := NewModel()
	m.width = 120
	m.height = 40

	updated, _ := m.Update(pricingMsg{rates: pricing.DefaultRates()})
	model := updated.(Model)
	model.width = 120
	model.height = 40
	output := model.View()

	if strings.Contains(output, "Loading rates") {
		t.Error("should not show loading after pricing msg")
	}
}

func TestCacheWarmTriggeredOnce(t *testing.T) {
	m := NewModel()

	updated, cmd := m.Update(pricingMsg{rates: pricing.DefaultRates()})
	model := updated.(Model)
	if cmd == nil {
		t.Fatal("first success should return cache warm command")
	}
	if !model.cacheWarmed {
		t.Fatal("cacheWarmed should be true")
	}

	updated, cmd = model.Update(pricingMsg{rates: pricing.DefaultRates()})
	model = updated.(Model)
	if cmd != nil {
		t.Error("second success should not return cache warm command")
	}
}

func TestCacheWarmNotTriggeredOnError(t *testing.T) {
	m := NewModel()

	fetchErr := errors.New("timeout")
	updated, cmd := m.Update(pricingMsg{rates: pricing.DefaultRates(), err: fetchErr})
	model := updated.(Model)

	if cmd != nil {
		t.Error("error should not return a command")
	}
	if model.cacheWarmed {
		t.Error("cacheWarmed should remain false on error")
	}
}

func TestCacheWarmMsgHandled(t *testing.T) {
	m := newReadyModel()
	updated, cmd := m.Update(cacheWarmMsg{})
	_ = updated.(Model) // should not panic
	if cmd != nil {
		t.Error("cacheWarmMsg should not return a command")
	}
}

// Preference persistence tests

func TestNewModelWithSavedRegion(t *testing.T) {
	dir := t.TempDir()
	prefs.SetDir(dir)
	defer prefs.SetDir("")

	_ = prefs.Save(prefs.Prefs{Region: "eu-west-1"})

	m := NewModel()
	if m.pricingRegion != "eu-west-1" {
		t.Errorf("expected saved region eu-west-1, got %q", m.pricingRegion)
	}
}

func TestNewModelIgnoresInvalidSavedRegion(t *testing.T) {
	dir := t.TempDir()
	prefs.SetDir(dir)
	defer prefs.SetDir("")

	_ = prefs.Save(prefs.Prefs{Region: "invalid-region"})

	m := NewModel()
	if m.pricingRegion != "us-east-1" {
		t.Errorf("expected default us-east-1 for invalid saved region, got %q", m.pricingRegion)
	}
}

func TestNewModelDefaultsWithNoPrefs(t *testing.T) {
	dir := t.TempDir()
	prefs.SetDir(dir)
	defer prefs.SetDir("")

	m := NewModel()
	if m.pricingRegion != "us-east-1" {
		t.Errorf("expected default us-east-1 with no prefs, got %q", m.pricingRegion)
	}
}

func TestRegionPickerSavesPreference(t *testing.T) {
	dir := t.TempDir()
	prefs.SetDir(dir)
	defer prefs.SetDir("")

	m := newReadyModel()
	m.view = viewRegions
	m.regionCursor = 1 // us-east-2

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(Model)

	if model.pricingRegion != "us-east-2" {
		t.Fatalf("expected us-east-2, got %q", model.pricingRegion)
	}

	saved := prefs.Load()
	if saved.Region != "us-east-2" {
		t.Errorf("expected saved region us-east-2, got %q", saved.Region)
	}
}

func TestRegionPickerSameRegionDoesNotSave(t *testing.T) {
	dir := t.TempDir()
	prefs.SetDir(dir)
	defer prefs.SetDir("")

	// Save eu-west-1 as the preferred region
	_ = prefs.Save(prefs.Prefs{Region: "eu-west-1"})

	m := newReadyModel()
	// Model loads eu-west-1 from prefs. Select the same region (index 4).
	m.view = viewRegions
	m.regionCursor = 4 // eu-west-1

	m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Prefs should still have the old value since same-region selection skips save
	saved := prefs.Load()
	if saved.Region != "eu-west-1" {
		t.Errorf("expected prefs unchanged (eu-west-1), got %q", saved.Region)
	}
}
