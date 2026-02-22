package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/josegonzalez/aws-eks-calculator/internal/calculator"
	"github.com/josegonzalez/aws-eks-calculator/internal/export"
	"github.com/josegonzalez/aws-eks-calculator/internal/prefs"
	"github.com/josegonzalez/aws-eks-calculator/internal/pricing"
	"github.com/josegonzalez/aws-eks-calculator/internal/tui/styles"
	"github.com/josegonzalez/aws-eks-calculator/internal/tui/views"
)

// viewState represents the current view in the TUI.
type viewState int

const (
	viewCapabilitySelector viewState = iota
	viewCalculator
	viewHelp
	viewRegions
)

// allRegions is the list of AWS regions available in the region picker.
var allRegions = []string{
	"us-east-1", "us-east-2", "us-west-1", "us-west-2",
	"eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1", "eu-central-2", "eu-north-1", "eu-south-1", "eu-south-2",
	"ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2", "ap-northeast-3", "ap-south-1", "ap-east-1",
	"sa-east-1", "ca-central-1", "me-south-1", "af-south-1",
}

// clearExportMsg is sent after a delay to clear the export status message.
type clearExportMsg struct{}

// pricingMsg carries the result of an async pricing fetch.
type pricingMsg struct {
	rates pricing.Rates
	err   error
}

// cacheWarmMsg is sent when background cache warming completes.
type cacheWarmMsg struct{}

// capabilityState holds per-capability TUI state.
type capabilityState struct {
	Inputs     []textinput.Model
	FocusIndex int
	Breakdown  calculator.CostBreakdown
}

// Model represents the main TUI application state.
type Model struct {
	width  int
	height int
	view viewState

	// Per-capability state
	activeCapability calculator.Capability
	capStates        map[calculator.Capability]*capabilityState

	// Capability selector
	capSelectorCursor int

	// Pricing state
	rates         pricing.Rates
	ratesLoading  bool
	ratesLoaded   bool
	ratesErr      error
	pricingRegion string
	cacheWarmed   bool

	// Region picker state
	regionCursor int
	allRegions   []string

	// Export
	exportDir string // directory for export files; empty means current dir
	exportMsg string

	quitting bool
}

// NewModel creates a new TUI model with default values.
func NewModel() Model {
	capStates := make(map[calculator.Capability]*capabilityState)
	for _, cap := range calculator.AllCapabilities {
		capStates[cap] = newCapabilityState(cap)
	}

	region := "us-east-1"
	if p := prefs.Load(); p.Region != "" && containsRegion(allRegions, p.Region) {
		region = p.Region
	}

	m := Model{
		activeCapability: calculator.CapabilityArgoCD,
		capStates:        capStates,
		allRegions:       allRegions,
		rates:            pricing.DefaultRates(),
		ratesLoading:     true,
		pricingRegion:    region,
		view:             viewCapabilitySelector,
	}

	m.applyLiveRates()
	m.recalculate()

	return m
}

func containsRegion(regions []string, region string) bool {
	for _, r := range regions {
		if r == region {
			return true
		}
	}
	return false
}

func newCapabilityState(cap calculator.Capability) *capabilityState {
	defaults := calculator.DefaultInput(cap)
	fields := views.InputFieldsForCapability(cap)
	n := len(fields)
	inputs := make([]textinput.Model, n)

	// Map inputs positionally based on capability
	idx := 0

	// Clusters
	inputs[idx] = newIntInput(fmt.Sprintf("%d", defaults.NumClusters))
	idx++

	// Resources/cluster
	inputs[idx] = newIntInput(fmt.Sprintf("%d", defaults.ResourcesPerCluster))
	idx++

	// Hours/month
	inputs[idx] = newFloatInput(fmt.Sprintf("%.0f", defaults.HoursPerMonth))
	idx++

	// ArgoCD-only: AppTemplates, ClustersPerTemplate
	if cap == calculator.CapabilityArgoCD {
		inputs[idx] = newIntInput("0")
		idx++
		inputs[idx] = newIntInput("0")
		idx++
	}

	// Self-managed inputs
	inputs[idx] = newFloatInput(fmt.Sprintf("%.1f", defaults.SelfManagedVCPUPerCluster))
	idx++
	inputs[idx] = newFloatInput(fmt.Sprintf("%.1f", defaults.SelfManagedMemGBPerCluster))
	idx++
	inputs[idx] = newFloatInput(fmt.Sprintf("%.4f", defaults.SelfManagedVCPUCostPerHour))
	idx++
	inputs[idx] = newFloatInput(fmt.Sprintf("%.4f", defaults.SelfManagedMemGBCostPerHour))

	inputs[0].Focus()
	inputs[0].TextStyle = styles.FocusedInputStyle

	return &capabilityState{
		Inputs: inputs,
	}
}

func newIntInput(value string) textinput.Model {
	ti := textinput.New()
	ti.SetValue(value)
	ti.Width = 10
	ti.CharLimit = 6
	return ti
}

func newFloatInput(value string) textinput.Model {
	ti := textinput.New()
	ti.SetValue(value)
	ti.Width = 10
	ti.CharLimit = 10
	return ti
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, fetchPricingCmd(m.pricingRegion))
}

func fetchPricingCmd(region string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		rates, err := pricing.FetchRates(ctx, region)
		return pricingMsg{rates: rates, err: err}
	}
}

// warmCacheCmd fetches pricing for all regions except skip, populating the
// on-disk cache so that future region switches are instant.
func warmCacheCmd(regions []string, skip string) tea.Cmd {
	return func() tea.Msg {
		for _, region := range regions {
			if region == skip {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			pricing.FetchRates(ctx, region)
			cancel()
		}
		return cacheWarmMsg{}
	}
}

// activeState returns the capabilityState for the currently active capability.
func (m *Model) activeState() *capabilityState {
	return m.capStates[m.activeCapability]
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case clearExportMsg:
		m.exportMsg = ""
		return m, nil

	case pricingMsg:
		m.ratesLoading = false
		m.rates = msg.rates
		m.applyLiveRates()
		m.recalculate()
		if msg.err == nil {
			m.ratesLoaded = true
			m.ratesErr = nil
			if !m.cacheWarmed {
				m.cacheWarmed = true
				return m, warmCacheCmd(m.allRegions, m.pricingRegion)
			}
		} else {
			m.ratesErr = msg.err
		}
		return m, nil

	case cacheWarmMsg:
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	// Update focused text input
	if m.view == viewCalculator {
		cs := m.activeState()
		if cs.FocusIndex < len(cs.Inputs) {
			var cmd tea.Cmd
			cs.Inputs[cs.FocusIndex], cmd = cs.Inputs[cs.FocusIndex].Update(msg)
			m.recalculate()
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.view {
	case viewCapabilitySelector:
		return m.handleSelectorKeys(msg)
	case viewCalculator:
		return m.handleCalculatorKeys(msg)
	case viewHelp:
		return m.handleHelpKeys(msg)
	case viewRegions:
		return m.handleRegionKeys(msg)
	}
	return m, nil
}

func (m Model) handleSelectorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.capSelectorCursor > 0 {
			m.capSelectorCursor--
		}
		return m, nil
	case "down", "j":
		if m.capSelectorCursor < len(calculator.AllCapabilities)-1 {
			m.capSelectorCursor++
		}
		return m, nil
	case "enter":
		m.activeCapability = calculator.AllCapabilities[m.capSelectorCursor]
		m.view = viewCalculator
		m.recalculate()
		return m, nil
	}
	return m, nil
}

func (m Model) handleCalculatorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cs := m.activeState()

	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit

	case "q":
		m.quitting = true
		return m, tea.Quit

	case "tab", "down":
		cs.FocusIndex = (cs.FocusIndex + 1) % len(cs.Inputs)
		cmd := m.updateFocus()
		return m, cmd

	case "shift+tab", "up":
		cs.FocusIndex = (cs.FocusIndex - 1 + len(cs.Inputs)) % len(cs.Inputs)
		cmd := m.updateFocus()
		return m, cmd

	case "[":
		idx := capabilityIndex(m.activeCapability)
		if idx > 0 {
			m.switchCapability(calculator.AllCapabilities[idx-1])
		} else {
			m.switchCapability(calculator.AllCapabilities[len(calculator.AllCapabilities)-1])
		}
		return m, nil

	case "]":
		idx := capabilityIndex(m.activeCapability)
		if idx < len(calculator.AllCapabilities)-1 {
			m.switchCapability(calculator.AllCapabilities[idx+1])
		} else {
			m.switchCapability(calculator.AllCapabilities[0])
		}
		return m, nil

	case "r":
		m.view = viewRegions
		m.regionCursor = 0
		return m, nil

	case "e":
		return m.doExport()

	case "?":
		m.view = viewHelp
		return m, nil
	}

	// Pass key to focused input
	var cmd tea.Cmd
	cs.Inputs[cs.FocusIndex], cmd = cs.Inputs[cs.FocusIndex].Update(msg)
	m.recalculate()
	return m, cmd
}

func (m *Model) switchCapability(cap calculator.Capability) {
	m.activeCapability = cap
	m.recalculate()
}

func capabilityIndex(cap calculator.Capability) int {
	for i, c := range calculator.AllCapabilities {
		if c == cap {
			return i
		}
	}
	return 0
}

func (m Model) handleRegionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.view = viewCalculator
		return m, nil
	case "up", "k":
		if m.regionCursor > 0 {
			m.regionCursor--
		}
		return m, nil
	case "down", "j":
		if m.regionCursor < len(m.allRegions)-1 {
			m.regionCursor++
		}
		return m, nil
	case "enter":
		selected := m.allRegions[m.regionCursor]
		m.view = viewCalculator
		if selected != m.pricingRegion {
			m.pricingRegion = selected
			m.ratesLoading = true
			_ = prefs.Save(prefs.Prefs{Region: selected})
			return m, fetchPricingCmd(selected)
		}
		return m, nil
	}
	return m, nil
}

func (m Model) handleHelpKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "?", "q":
		m.view = viewCalculator
		return m, nil
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) updateFocus() tea.Cmd {
	cs := m.activeState()
	cmds := make([]tea.Cmd, len(cs.Inputs))
	for i := range cs.Inputs {
		if i == cs.FocusIndex {
			cmds[i] = cs.Inputs[i].Focus()
			cs.Inputs[i].TextStyle = styles.FocusedInputStyle
		} else {
			cs.Inputs[i].Blur()
			cs.Inputs[i].TextStyle = styles.BlurredInputStyle
		}
	}
	return tea.Batch(cmds...)
}

func (m *Model) recalculate() {
	cs := m.activeState()
	input := m.buildInput()
	cs.Breakdown = calculator.Calculate(input)
}

func (m *Model) buildInput() calculator.ScenarioInput {
	cs := m.activeState()
	cap := m.activeCapability
	base, resource := m.rates.ForCapability(cap)

	input := calculator.ScenarioInput{
		Name:            "Custom",
		Capability:      cap,
		NumClusters:     parseInt(cs.Inputs[0].Value()),
		ResourcesPerCluster: parseInt(cs.Inputs[1].Value()),
		HoursPerMonth:   parseFloat(cs.Inputs[2].Value()),
		Region:          m.pricingRegion,
		BasePerHour:     base,
		ResourcePerHour: resource,
	}

	if cap == calculator.CapabilityArgoCD {
		input.AppTemplates = parseInt(cs.Inputs[3].Value())
		input.ClustersPerTemplate = parseInt(cs.Inputs[4].Value())
		input.SelfManagedVCPUPerCluster = parseFloat(cs.Inputs[5].Value())
		input.SelfManagedMemGBPerCluster = parseFloat(cs.Inputs[6].Value())
		input.SelfManagedVCPUCostPerHour = parseFloat(cs.Inputs[7].Value())
		input.SelfManagedMemGBCostPerHour = parseFloat(cs.Inputs[8].Value())
	} else {
		input.SelfManagedVCPUPerCluster = parseFloat(cs.Inputs[3].Value())
		input.SelfManagedMemGBPerCluster = parseFloat(cs.Inputs[4].Value())
		input.SelfManagedVCPUCostPerHour = parseFloat(cs.Inputs[5].Value())
		input.SelfManagedMemGBCostPerHour = parseFloat(cs.Inputs[6].Value())
	}

	return input
}

func (m *Model) applyLiveRates() {
	for _, cap := range calculator.AllCapabilities {
		cs := m.capStates[cap]
		if cap == calculator.CapabilityArgoCD {
			cs.Inputs[7].SetValue(fmt.Sprintf("%.4f", m.rates.FargateVCPUPerHour))
			cs.Inputs[8].SetValue(fmt.Sprintf("%.4f", m.rates.FargateMemGBPerHour))
		} else {
			cs.Inputs[5].SetValue(fmt.Sprintf("%.4f", m.rates.FargateVCPUPerHour))
			cs.Inputs[6].SetValue(fmt.Sprintf("%.4f", m.rates.FargateMemGBPerHour))
		}
	}
}

func (m Model) exportPath(name string) string {
	if m.exportDir != "" {
		return m.exportDir + "/" + name
	}
	return name
}

func (m Model) doExport() (Model, tea.Cmd) {
	input := m.buildInput()
	cs := m.activeState()
	scenario := export.Scenario{Input: input, Breakdown: cs.Breakdown}

	filename := fmt.Sprintf("%s-cost-estimate.csv", strings.ToLower(m.activeCapability.String()))
	path := m.exportPath(filename)
	if err := export.ToCSV([]export.Scenario{scenario}, path); err != nil {
		m.exportMsg = fmt.Sprintf("Export failed: %v", err)
	} else {
		m.exportMsg = fmt.Sprintf("Exported to %s", path)
	}
	return m, tea.Tick(3*time.Second, clearExportTick)
}

func clearExportTick(time.Time) tea.Msg {
	return clearExportMsg{}
}

// View renders the current view.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	b.WriteString(styles.TitleStyle.Render("AWS EKS Capabilities Cost Calculator"))
	b.WriteString("\n\n")

	if m.ratesLoading {
		b.WriteString(styles.MutedStyle.Render(fmt.Sprintf("Loading rates for %s...", m.pricingRegion)))
	} else {
		switch m.view {
		case viewCapabilitySelector:
			b.WriteString(views.RenderCapabilitySelector(m.capSelectorCursor))

		case viewCalculator:
			cs := m.activeState()
			input := m.buildInput()

			b.WriteString(views.RenderTabBar(m.activeCapability))
			b.WriteString("\n\n")
			b.WriteString(views.RenderCalculator(m.activeCapability, cs.Inputs, cs.FocusIndex, input, cs.Breakdown, m.width, m.height))
			b.WriteString("\n\n")

			hints := views.InputHintsForCapability(m.activeCapability)
			if cs.FocusIndex >= 0 && cs.FocusIndex < len(hints) {
				b.WriteString(styles.MutedStyle.Render(hints[cs.FocusIndex]))
				b.WriteString("\n")
			}
		case viewHelp:
			b.WriteString(views.RenderHelp())

		case viewRegions:
			b.WriteString(views.RenderRegions(m.allRegions, m.regionCursor))
		}

		var hint string
		switch m.view {
		case viewCapabilitySelector:
			hint = "↑/↓ navigate  enter select  q quit"
		case viewCalculator:
			hint = "↑/↓/tab navigate  [/] capability  r region  e export  ? help  q quit"
		case viewHelp:
			hint = "esc back  q quit"
		case viewRegions:
			hint = "↑/↓ navigate  enter select  esc cancel"
		}
		if hint != "" {
			b.WriteString("\n")
			b.WriteString(styles.HelpStyle.Render(hint))
		}
	}

	if m.exportMsg != "" {
		b.WriteString("\n\n")
		b.WriteString(styles.SuccessStyle.Render(m.exportMsg))
	}

	if m.ratesErr != nil {
		b.WriteString("\n\n")
		if m.ratesLoaded {
			b.WriteString(styles.WarningStyle.Render(
				fmt.Sprintf("⚠ Unable to fetch rates for %s. Using previously fetched rates.", m.pricingRegion)))
		} else {
			b.WriteString(styles.WarningStyle.Render(
				fmt.Sprintf("⚠ Unable to fetch rates for %s. Using default rates.", m.pricingRegion)))
		}
	}

	return b.String()
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	if v < 0 {
		return 0
	}
	return v
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	if v < 0 {
		return 0
	}
	return v
}
