// Step 10: Complete Styling & Theming System - enables global theming

package main

import (
	"fmt"
	"runtime"

	"github.com/AllenDang/cimgui-go/backend"
	"github.com/AllenDang/cimgui-go/backend/glfwbackend"
	"github.com/AllenDang/cimgui-go/imgui"
)

var (
	userName     = "Enter your name"
	showGreeting = false
	counter      = 0

	sliderValue   = float32(50.0)
	colorValue    = [3]float32{1.0, 0.5, 0.2} // Orange color
	progress      = float32(0.0)
	hoverCount    = 0
	clickCount    = 0
	eventMessage  = "Interact with widgets to see events"
	totalCounters = 0

	// Styling state
	currentTheme = 0
	customColor  = RGB(100, 150, 200)
)

var (
	ColorRed    = RGB(255, 0, 0)
	ColorGreen  = RGB(0, 255, 0)
	ColorBlue   = RGB(0, 0, 255)
	ColorYellow = RGB(255, 255, 0)
	ColorWhite  = RGB(255, 255, 255)
	ColorBlack  = RGB(0, 0, 0)
	ColorGray   = RGB(128, 128, 128)
)

// MasterWindow represents the main application window
type MasterWindow struct {
	backend backend.Backend[glfwbackend.GLFWWindowFlags]
	title   string
	width   int
	height  int
}

// Global status display instance
var globalStatus *StatusDisplayWidget

// Global theme tracking
var currentThemeObject *Theme

// LogStatus adds a message to the global status display
func LogStatus(message string) {
	if globalStatus != nil {
		globalStatus.AddMessage(message)
	}
	fmt.Printf("[STATUS] %s\n", message)
}

// FIXED: Proper global theme application
func SetGlobalTheme(theme *Theme) {
	currentThemeObject = theme
	LogStatus(fmt.Sprintf("Theme set to: %s", theme.name))
}

// Widget interface - every GUI element implements this
type Widget interface {
	Build()
}

// Layout is a collection of widgets that implements Widget itself
type Layout []Widget

func (l Layout) Build() {
	for _, widget := range l {
		if widget != nil {
			widget.Build()
		}
	}
}

type EventWidget struct {
	onHover       func()
	onClick       func()
	onDoubleClick func()
	onRightClick  func()
	onKeyPress    func(key int)
}

// Event creates an event handler widget
func Event() *EventWidget {
	return &EventWidget{}
}

func (e *EventWidget) OnHover(onHover func()) *EventWidget {
	e.onHover = onHover
	return e
}

func (e *EventWidget) OnClick(onClick func()) *EventWidget {
	e.onClick = onClick
	return e
}

func (e *EventWidget) OnDoubleClick(onDoubleClick func()) *EventWidget {
	e.onDoubleClick = onDoubleClick
	return e
}

func (e *EventWidget) OnRightClick(onRightClick func()) *EventWidget {
	e.onRightClick = onRightClick
	return e
}

func (e *EventWidget) OnKeyPress(onKeyPress func(key int)) *EventWidget {
	e.onKeyPress = onKeyPress
	return e
}

func (e *EventWidget) Build() {
	// Check if previous item was hovered
	if imgui.IsItemHovered() && e.onHover != nil {
		e.onHover()
	}

	// Check for mouse clicks on previous item
	if imgui.IsItemClicked() && e.onClick != nil {
		e.onClick()
	}

	if imgui.IsItemHovered() && imgui.IsMouseDoubleClicked(imgui.MouseButtonLeft) && e.onDoubleClick != nil {
		e.onDoubleClick()
	}

	if imgui.IsItemHovered() && imgui.IsMouseDown(imgui.MouseButtonRight) && e.onRightClick != nil {
		e.onRightClick()
	}

	// Check for key presses when item is focused
	if imgui.IsItemFocused() && e.onKeyPress != nil {
		// Check some common keys
		if imgui.IsKeyPressedBoolV(imgui.KeyEnter, true) {
			e.onKeyPress(int(imgui.KeyEnter))
		}
		if imgui.IsKeyPressedBoolV(imgui.KeyEscape, true) {
			e.onKeyPress(int(imgui.KeyEscape))
		}
		if imgui.IsKeyPressedBoolV(imgui.KeySpace, true) {
			e.onKeyPress(int(imgui.KeySpace))
		}
	}
}

type TooltipWidget struct {
	text string
}

// Tooltip creates a tooltip widget
func Tooltip(text string) *TooltipWidget {
	return &TooltipWidget{text: text}
}

// Build shows the tooltip if previous item is hovered
func (t *TooltipWidget) Build() {
	if imgui.IsItemHovered() {
		imgui.SetTooltip(t.text)
	}
}

type LabelWidget struct {
	text string
}

func Label(text string) *LabelWidget {
	return &LabelWidget{text: text}
}

func (l *LabelWidget) Build() {
	imgui.Text(l.text)
}

type ButtonWidget struct {
	text    string
	onClick func()
	width   float32
	height  float32
}

func Button(text string) *ButtonWidget {
	return &ButtonWidget{text: text,
		width: 0, height: 0}
}

func (b *ButtonWidget) OnClick(fn func()) *ButtonWidget {
	b.onClick = fn
	return b
}

func (b *ButtonWidget) Build() {
	var clicked bool
	if b.width > 0 && b.height > 0 {
		clicked = imgui.ButtonV(b.text, imgui.Vec2{X: b.width, Y: b.height})
	} else {
		clicked = imgui.Button(b.text)
	}
	if clicked && b.onClick != nil {
		b.onClick()
	}
}

func (b *ButtonWidget) Size(width, height float32) *ButtonWidget {
	b.width = width
	b.height = height
	return b
}

type RowWidget struct {
	Widgets []Widget
}

func Row(widgets ...Widget) *RowWidget {
	row := &RowWidget{Widgets: widgets}
	return row
}

func (r *RowWidget) Build() {
	if len(r.Widgets) == 0 {
		return
	}

	// For simple horizontal layout, use a table
	if imgui.BeginTableV("#row_table", int32(len(r.Widgets)), imgui.TableFlagsNone, imgui.Vec2{}, 0.0) {
		imgui.TableNextRow()

		for _, widget := range r.Widgets {
			imgui.TableNextColumn()
			widget.Build()
		}

		imgui.EndTable()
	}
}

type SpacingWidget struct{}

func Spacing() *SpacingWidget {
	return &SpacingWidget{}
}

func (s *SpacingWidget) Build() {
	imgui.Spacing()
}

// HotkeyWidget handles global keyboard shortcuts
type HotkeyWidget struct {
	key      int
	ctrl     bool
	shift    bool
	alt      bool
	callback func()
}

// Hotkey creates a global hotkey handler
func Hotkey(key int) *HotkeyWidget {
	return &HotkeyWidget{key: key}
}

// Ctrl adds Ctrl modifier (builder pattern)
func (h *HotkeyWidget) Ctrl() *HotkeyWidget {
	h.ctrl = true
	return h
}

// Shift adds Shift modifier (builder pattern)
func (h *HotkeyWidget) Shift() *HotkeyWidget {
	h.shift = true
	return h
}

// Alt adds Alt modifier (builder pattern)
func (h *HotkeyWidget) Alt() *HotkeyWidget {
	h.alt = true
	return h
}

// OnPress sets the callback for when hotkey is pressed (builder pattern)
func (h *HotkeyWidget) OnPress(callback func()) *HotkeyWidget {
	h.callback = callback
	return h
}

// Build checks for hotkey presses
func (h *HotkeyWidget) Build() {
	// Check if the key combination is pressed
	if imgui.IsKeyDown(imgui.Key(h.key)) {
		ctrlPressed := imgui.IsKeyDown(imgui.KeyLeftCtrl) || imgui.IsKeyDown(imgui.KeyRightCtrl)
		shiftPressed := imgui.IsKeyDown(imgui.KeyLeftShift) || imgui.IsKeyDown(imgui.KeyRightShift)
		altPressed := imgui.IsKeyDown(imgui.KeyLeftAlt) || imgui.IsKeyDown(imgui.KeyRightAlt)

		// Check if modifiers match
		if h.ctrl == ctrlPressed && h.shift == shiftPressed && h.alt == altPressed {
			if h.callback != nil {
				h.callback()
			}
		}
	}
}

type Sizeable interface {
	Size(width, height float32) Widget
}

// SeparatorWidget adds a horizontal line
type SeparatorWidget struct{}

// Separator creates a separator widget
func Separator() *SeparatorWidget {
	return &SeparatorWidget{}
}

// Build renders separator using ImGui
func (s *SeparatorWidget) Build() {
	imgui.Separator()
}

// NewMasterWindow creates a new master window
func NewMasterWindow(title string, width, height int) *MasterWindow {
	runtime.LockOSThread() // Required for OpenGL context

	// Create ImGui context
	imgui.CreateContext()

	// Create GLFW backend
	glfwBackend := glfwbackend.NewGLFWBackend()

	// Create the backend wrapper
	backendInstance, err := backend.CreateBackend(glfwBackend)
	if err != nil {
		panic(err)
	}

	// Create the window
	backendInstance.CreateWindow(title, width, height)

	return &MasterWindow{
		backend: backendInstance,
		title:   title,
		width:   width,
		height:  height,
	}
}

// FIXED: Proper theme application in the main loop
func (w *MasterWindow) Run(loopFunc func()) {
	w.backend.Run(func() {
		// Apply global theme at the start of each frame
		var colorCount, varCount int32
		if currentThemeObject != nil {
			// Push theme colors
			for colorID, color := range currentThemeObject.colors {
				imgui.PushStyleColorVec4(imgui.Col(colorID), color)
				colorCount++
			}

			// Push theme variables
			for varID, value := range currentThemeObject.vars {
				imgui.PushStyleVarFloat(imgui.StyleVar(varID), value)
				varCount++
			}
		}

		// Execute user's UI definition
		loopFunc()

		// Pop theme styles at the end of the frame
		if varCount > 0 {
			imgui.PopStyleVarV(varCount)
		}
		if colorCount > 0 {
			imgui.PopStyleColorV(colorCount)
		}
	})
}

func onHelloClick() {
	println("Hello button was clicked!")
}

func onGoodbyeClick() {
	println("Goodbye button was clicked!")
}

type InputTextWidget struct {
	id       string
	label    string
	text     *string
	width    float32
	onChange func()
}

func InputText(label string, text *string) *InputTextWidget {
	id := fmt.Sprintf("%s##input", label)

	return &InputTextWidget{
		id:    id,
		label: label,
		text:  text,
		width: 0,
	}
}

func (i *InputTextWidget) Size(width float32) *InputTextWidget {
	i.width = width
	return i
}

func (i *InputTextWidget) OnChange(onChange func()) *InputTextWidget {
	i.onChange = onChange
	return i
}

func (i *InputTextWidget) Build() {
	if i.width > 0 {
		imgui.SetNextItemWidth(i.width)
	}

	oldText := *i.text
	changed := imgui.InputTextWithHint(i.id, "", i.text, 0, nil)

	if changed && oldText != *i.text && i.onChange != nil {
		i.onChange()
	}
}

// Context manages global state for our GUI framework
type Context struct {
	widgetCounter int
	stateMap      map[string]interface{}
}

// Global context instance
var GlobalContext = &Context{
	widgetCounter: 0,
	stateMap:      make(map[string]interface{}),
}

// GenAutoID generates unique IDs for widgets
func GenAutoID(prefix string) string {
	GlobalContext.widgetCounter++
	return fmt.Sprintf("%s##%d", prefix, GlobalContext.widgetCounter)
}

type CheckboxWidget struct {
	id       string
	onChange func()
	label    string
	checked  *bool
}

func Checkbox(label string, checked *bool) *CheckboxWidget {
	id := fmt.Sprintf("%s##checkbox", label)

	return &CheckboxWidget{
		id:      id,
		label:   label,
		checked: checked,
	}
}

func (c *CheckboxWidget) OnChange(fn func()) *CheckboxWidget {
	c.onChange = fn
	return c
}

func (c *CheckboxWidget) Build() {
	if c.checked == nil {
		panic("c.checked is nil in Build method!")
	}

	oldValue := *c.checked
	imgui.Checkbox(c.label, c.checked)

	if oldValue != *c.checked && c.onChange != nil {
		fmt.Printf("Checkbox changed from %t to %t, calling onChange\n", oldValue, *c.checked)
		c.onChange()
	}
}

// SingleWindowWidget fills the entire master window
type SingleWindowWidget struct {
	widgets []Widget
}

func SingleWindow() *SingleWindowWidget {
	return &SingleWindowWidget{
		widgets: []Widget{},
	}
}

func (s *SingleWindowWidget) Layout(widgets ...Widget) *SingleWindowWidget {
	s.widgets = widgets
	return s
}

func (s *SingleWindowWidget) Build() {
	viewport := imgui.MainViewport()
	pos := viewport.Pos()
	size := viewport.Size()

	imgui.SetNextWindowPos(pos)
	imgui.SetNextWindowSize(size)

	flags := imgui.WindowFlagsNoTitleBar |
		imgui.WindowFlagsNoResize |
		imgui.WindowFlagsNoMove |
		imgui.WindowFlagsNoCollapse |
		imgui.WindowFlagsNoScrollbar

	imgui.BeginV("##SingleWindow", nil, imgui.WindowFlags(flags))

	for _, widget := range s.widgets {
		if widget != nil {
			widget.Build()
		}
	}

	imgui.End()
}

// ColumnWidget arranges widgets vertically
type ColumnWidget struct {
	widgets []Widget
}

func Column(widgets ...Widget) *ColumnWidget {
	return &ColumnWidget{widgets: widgets}
}

func (c *ColumnWidget) Build() {
	for _, widget := range c.widgets {
		if widget != nil {
			widget.Build()
		}
	}
}

// SliderWidget represents a value slider
type SliderWidget struct {
	id       string
	label    string
	value    *float32
	min, max float32
	onChange func()
}

func SliderFloat(label string, value *float32, min, max float32) *SliderWidget {
	id := fmt.Sprintf("%s##slider", label)
	return &SliderWidget{
		id:    id,
		label: label,
		value: value,
		min:   min,
		max:   max,
	}
}

func (s *SliderWidget) OnChange(onChange func()) *SliderWidget {
	s.onChange = onChange
	return s
}

func (s *SliderWidget) Build() {
	oldValue := *s.value

	if imgui.SliderFloatV(s.label, s.value, s.min, s.max, "%.2f", 0) {
		if oldValue != *s.value && s.onChange != nil {
			s.onChange()
		}
	}
}

// ColorEditWidget represents a color picker
type ColorEditWidget struct {
	id       string
	label    string
	color    *[3]float32
	onChange func()
}

func ColorEdit(label string, color *[3]float32) *ColorEditWidget {
	id := fmt.Sprintf("%s##color", label)
	return &ColorEditWidget{
		id:    id,
		label: label,
		color: color,
	}
}

func (c *ColorEditWidget) OnChange(onChange func()) *ColorEditWidget {
	c.onChange = onChange
	return c
}

func (c *ColorEditWidget) Build() {
	oldColor := *c.color

	if imgui.ColorEdit3V(c.label, c.color, 0) {
		if oldColor != *c.color && c.onChange != nil {
			c.onChange()
		}
	}
}

// ProgressBarWidget represents a progress bar
type ProgressBarWidget struct {
	progress float32
	width    float32
	height   float32
	overlay  string
}

func ProgressBar(progress float32) *ProgressBarWidget {
	return &ProgressBarWidget{
		progress: progress,
		width:    -1,
		height:   0,
	}
}

func (p *ProgressBarWidget) Size(width, height float32) *ProgressBarWidget {
	p.width = width
	p.height = height
	return p
}

func (p *ProgressBarWidget) Overlay(text string) *ProgressBarWidget {
	p.overlay = text
	return p
}

func (p *ProgressBarWidget) Build() {
	size := imgui.Vec2{X: p.width, Y: p.height}
	imgui.ProgressBarV(p.progress, size, p.overlay)
}

// counterState holds internal state for CounterWidget
type counterState struct {
	value int
	step  int
}

func (s *counterState) Dispose() {
	// Nothing to clean up for this simple state
}

// CounterWidget is a custom widget that manages its own counter state
type CounterWidget struct {
	id       string
	label    string
	minValue int
	maxValue int
	onChange func(int)
}

func Counter(label string) *CounterWidget {
	id := fmt.Sprintf("%s##counter", label)
	return &CounterWidget{
		id:       id,
		label:    label,
		minValue: 0,
		maxValue: 100,
	}
}

func (c *CounterWidget) Min(min int) *CounterWidget {
	c.minValue = min
	return c
}

func (c *CounterWidget) Max(max int) *CounterWidget {
	c.maxValue = max
	return c
}

func (c *CounterWidget) OnChange(onChange func(int)) *CounterWidget {
	c.onChange = onChange
	return c
}

func (c *CounterWidget) getState() *counterState {
	if existingState, exists := GlobalContext.stateMap[c.id]; exists {
		if state, ok := existingState.(*counterState); ok {
			return state
		}
	}

	newState := &counterState{
		value: c.minValue,
		step:  1,
	}
	GlobalContext.stateMap[c.id] = newState
	return newState
}

func (c *CounterWidget) Build() {
	state := c.getState()

	if imgui.BeginTableV("##counter_table", 4, imgui.TableFlagsNone, imgui.Vec2{}, 0.0) {
		imgui.TableNextRow()

		imgui.TableNextColumn()
		imgui.Text(c.label)

		imgui.TableNextColumn()
		if imgui.Button("-") && state.value > c.minValue {
			oldValue := state.value
			state.value--
			if c.onChange != nil {
				c.onChange(state.value)
			}
			fmt.Printf("%s: %d -> %d\n", c.label, oldValue, state.value)
		}

		imgui.TableNextColumn()
		imgui.Text(fmt.Sprintf(" %d ", state.value))

		imgui.TableNextColumn()
		if imgui.Button("+") && state.value < c.maxValue {
			oldValue := state.value
			state.value++
			if c.onChange != nil {
				c.onChange(state.value)
			}
			fmt.Printf("%s: %d -> %d\n", c.label, oldValue, state.value)
		}

		imgui.EndTable()
	}
}

func (c *CounterWidget) GetValue() int {
	state := c.getState()
	return state.value
}

func (c *CounterWidget) SetValue(value int) {
	state := c.getState()
	if value >= c.minValue && value <= c.maxValue {
		state.value = value
	}
}

// timerState holds internal state for TimerWidget
type timerState struct {
	startTime   float64
	elapsedTime float64
	isRunning   bool
	isPaused    bool
}

func (s *timerState) Dispose() {
	// Nothing to clean up
}

// TimerWidget shows elapsed time with start/stop/reset controls
type TimerWidget struct {
	id    string
	label string
}

func Timer(label string) *TimerWidget {
	return &TimerWidget{
		id:    fmt.Sprintf("%s##timer", label),
		label: label,
	}
}

func (t *TimerWidget) getState() *timerState {
	if existingState, exists := GlobalContext.stateMap[t.id]; exists {
		if state, ok := existingState.(*timerState); ok {
			return state
		}
	}

	newState := &timerState{
		startTime:   imgui.Time(),
		elapsedTime: 0.0,
		isRunning:   false,
		isPaused:    false,
	}
	GlobalContext.stateMap[t.id] = newState
	return newState
}

func (t *TimerWidget) Build() {
	state := t.getState()
	currentTime := imgui.Time()

	if state.isRunning && !state.isPaused {
		state.elapsedTime = currentTime - state.startTime
	}

	imgui.Text(fmt.Sprintf("%s: %.1fs", t.label, state.elapsedTime))

	if imgui.BeginTableV("##timer_controls", 3, imgui.TableFlagsNone, imgui.Vec2{}, 0.0) {
		imgui.TableNextRow()

		imgui.TableNextColumn()
		if !state.isRunning {
			if imgui.Button("Start") {
				state.startTime = currentTime - state.elapsedTime
				state.isRunning = true
				state.isPaused = false
			}
		} else {
			if !state.isPaused {
				if imgui.Button("Pause") {
					state.isPaused = true
				}
			} else {
				if imgui.Button("Resume") {
					state.startTime = currentTime - state.elapsedTime
					state.isPaused = false
				}
			}
		}

		imgui.TableNextColumn()
		if imgui.Button("Stop") {
			state.isRunning = false
			state.isPaused = false
			state.elapsedTime = 0.0
		}

		imgui.TableNextColumn()
		if imgui.Button("Reset") {
			state.startTime = currentTime
			state.elapsedTime = 0.0
		}

		imgui.EndTable()
	}
}

func (t *TimerWidget) GetElapsed() float64 {
	state := t.getState()
	return state.elapsedTime
}

// statusState holds the status display state
type statusState struct {
	messages    []string
	timestamps  []float64
	maxMessages int
}

func (s *statusState) Dispose() {
	s.messages = nil
	s.timestamps = nil
}

// StatusDisplayWidget shows a scrolling list of status messages
type StatusDisplayWidget struct {
	id     string
	height float32
}

func StatusDisplay() *StatusDisplayWidget {
	return &StatusDisplayWidget{
		id:     "##status_display",
		height: 100,
	}
}

func (s *StatusDisplayWidget) Height(height float32) *StatusDisplayWidget {
	s.height = height
	return s
}

func (s *StatusDisplayWidget) getState() *statusState {
	if existingState, exists := GlobalContext.stateMap[s.id]; exists {
		if state, ok := existingState.(*statusState); ok {
			return state
		}
	}

	newState := &statusState{
		messages:    make([]string, 0),
		timestamps:  make([]float64, 0),
		maxMessages: 100,
	}
	GlobalContext.stateMap[s.id] = newState
	return newState
}

func (s *StatusDisplayWidget) AddMessage(message string) {
	state := s.getState()
	currentTime := imgui.Time()

	state.messages = append(state.messages, message)
	state.timestamps = append(state.timestamps, currentTime)

	if len(state.messages) > state.maxMessages {
		state.messages = state.messages[1:]
		state.timestamps = state.timestamps[1:]
	}
}

func (s *StatusDisplayWidget) Build() {
	state := s.getState()
	currentTime := imgui.Time()

	for i := len(state.messages) - 1; i >= 0; i-- {
		age := currentTime - state.timestamps[i]

		if age < 10.0 {
			timeStr := fmt.Sprintf("[%.1fs] %s", age, state.messages[i])
			imgui.Text(timeStr)
		}
	}
}

// FIXED: StyleSetter with proper stack management
type StyleSetter struct {
	colors  map[int]imgui.Vec4
	vars    map[int]float32
	widgets []Widget
}

func Style() *StyleSetter {
	return &StyleSetter{
		colors:  make(map[int]imgui.Vec4),
		vars:    make(map[int]float32),
		widgets: make([]Widget, 0),
	}
}

func (s *StyleSetter) SetColor(colorID int, color imgui.Vec4) *StyleSetter {
	s.colors[colorID] = color
	return s
}

func (s *StyleSetter) SetVar(varID int, value float32) *StyleSetter {
	s.vars[varID] = value
	return s
}

func (s *StyleSetter) To(widgets ...Widget) *StyleSetter {
	s.widgets = widgets
	return s
}

// FIXED: Proper stack management
func (s *StyleSetter) Build() {
	// Count what we're pushing
	colorCount := int32(len(s.colors))
	varCount := int32(len(s.vars))

	// Push all style colors
	for colorID, color := range s.colors {
		imgui.PushStyleColorVec4(imgui.Col(colorID), color)
	}

	// Push all style variables
	for varID, value := range s.vars {
		imgui.PushStyleVarFloat(imgui.StyleVar(varID), value)
	}

	// Render child widgets with applied styles
	for _, widget := range s.widgets {
		if widget != nil {
			widget.Build()
		}
	}

	// Pop in reverse order (IMPORTANT!)
	if varCount > 0 {
		imgui.PopStyleVarV(varCount)
	}
	if colorCount > 0 {
		imgui.PopStyleColorV(colorCount)
	}
}

// Theme represents a complete UI theme
type Theme struct {
	name   string
	colors map[int]imgui.Vec4
	vars   map[int]float32
}

// FIXED: Better theme definitions
var (
	DarkTheme = &Theme{
		name: "Dark",
		colors: map[int]imgui.Vec4{
			int(imgui.ColWindowBg):       {X: 0.06, Y: 0.06, Z: 0.06, W: 1.00},
			int(imgui.ColButton):         {X: 0.20, Y: 0.25, Z: 0.29, W: 1.00},
			int(imgui.ColButtonHovered):  {X: 0.26, Y: 0.59, Z: 0.98, W: 0.40},
			int(imgui.ColButtonActive):   {X: 0.26, Y: 0.59, Z: 0.98, W: 0.67},
			int(imgui.ColText):           {X: 1.00, Y: 1.00, Z: 1.00, W: 1.00},
			int(imgui.ColFrameBg):        {X: 0.16, Y: 0.29, Z: 0.48, W: 0.54},
			int(imgui.ColFrameBgHovered): {X: 0.26, Y: 0.59, Z: 0.98, W: 0.40},
			int(imgui.ColFrameBgActive):  {X: 0.26, Y: 0.59, Z: 0.98, W: 0.67},
		},
		vars: map[int]float32{
			int(imgui.StyleVarWindowRounding): 5.0,
			int(imgui.StyleVarFrameRounding):  3.0,
		},
	}

	LightTheme = &Theme{
		name: "Light",
		colors: map[int]imgui.Vec4{
			int(imgui.ColWindowBg):       {X: 0.94, Y: 0.94, Z: 0.94, W: 1.00},
			int(imgui.ColButton):         {X: 0.74, Y: 0.74, Z: 0.74, W: 1.00},
			int(imgui.ColButtonHovered):  {X: 0.86, Y: 0.86, Z: 0.86, W: 1.00},
			int(imgui.ColButtonActive):   {X: 0.64, Y: 0.64, Z: 0.64, W: 1.00},
			int(imgui.ColText):           {X: 0.00, Y: 0.00, Z: 0.00, W: 1.00},
			int(imgui.ColFrameBg):        {X: 1.00, Y: 1.00, Z: 1.00, W: 0.54},
			int(imgui.ColFrameBgHovered): {X: 0.86, Y: 0.86, Z: 0.86, W: 0.40},
			int(imgui.ColFrameBgActive):  {X: 0.76, Y: 0.76, Z: 0.76, W: 0.67},
		},
		vars: map[int]float32{
			int(imgui.StyleVarWindowRounding): 2.0,
			int(imgui.StyleVarFrameRounding):  2.0,
		},
	}

	BlueTheme = &Theme{
		name: "Blue",
		colors: map[int]imgui.Vec4{
			int(imgui.ColWindowBg):       {X: 0.11, Y: 0.15, Z: 0.25, W: 1.00},
			int(imgui.ColButton):         {X: 0.26, Y: 0.59, Z: 0.98, W: 0.40},
			int(imgui.ColButtonHovered):  {X: 0.26, Y: 0.59, Z: 0.98, W: 1.00},
			int(imgui.ColButtonActive):   {X: 0.06, Y: 0.53, Z: 0.98, W: 1.00},
			int(imgui.ColText):           {X: 1.00, Y: 1.00, Z: 1.00, W: 1.00},
			int(imgui.ColFrameBg):        {X: 0.16, Y: 0.29, Z: 0.48, W: 0.54},
			int(imgui.ColFrameBgHovered): {X: 0.26, Y: 0.59, Z: 0.98, W: 0.40},
			int(imgui.ColFrameBgActive):  {X: 0.26, Y: 0.59, Z: 0.98, W: 0.67},
		},
		vars: map[int]float32{
			int(imgui.StyleVarWindowRounding): 8.0,
			int(imgui.StyleVarFrameRounding):  4.0,
		},
	}
)

// GetAvailableThemes returns all available themes
func GetAvailableThemes() []*Theme {
	return []*Theme{DarkTheme, LightTheme, BlueTheme}
}

// Color helper functions for easier color creation
func RGB(r, g, b float32) imgui.Vec4 {
	return imgui.Vec4{X: r / 255.0, Y: g / 255.0, Z: b / 255.0, W: 1.0}
}

func RGBA(r, g, b, a float32) imgui.Vec4 {
	return imgui.Vec4{X: r / 255.0, Y: g / 255.0, Z: b / 255.0, W: a / 255.0}
}

func ColorFromHex(hex string) imgui.Vec4 {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}

	if len(hex) != 6 {
		return imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}
	}

	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)

	return RGB(float32(r), float32(g), float32(b))
}

// FIXED: Working theme switching and styling demo
func loop() {
	if globalStatus == nil {
		globalStatus = StatusDisplay().Height(120)
	}

	SingleWindow().Layout(
		Label("✅ FIXED: Complete Styling & Theming System - Step 10"),
		Separator(),

		// FIXED: Working theme selector
		Label("🎨 Global Theme Selection:"),
		Row(
			Button("Dark Theme").OnClick(func() {
				currentTheme = 0
				SetGlobalTheme(DarkTheme)
			}),
			Button("Light Theme").OnClick(func() {
				currentTheme = 1
				SetGlobalTheme(LightTheme)
			}),
			Button("Blue Theme").OnClick(func() {
				currentTheme = 2
				SetGlobalTheme(BlueTheme)
			}),
		),

		Spacing(),

		// Individual widget styling examples
		Label("🖌️ Individual Widget Styling:"),

		Style().
			SetColor(int(imgui.ColButton), ColorRed).
			SetColor(int(imgui.ColText), ColorWhite).
			To(
				Button("Red Button").OnClick(func() {
					LogStatus("Red button clicked!")
				}),
			),

		Style().
			SetColor(int(imgui.ColButton), ColorGreen).
			SetColor(int(imgui.ColText), ColorBlack).
			To(
				Button("Green Button").OnClick(func() {
					LogStatus("Green button clicked!")
				}),
			),

		Style().
			SetColor(int(imgui.ColButton), customColor).
			SetVar(int(imgui.StyleVarFrameRounding), 10.0).
			To(
				Button("Custom Rounded").OnClick(func() {
					LogStatus("Custom styled button clicked!")
				}),
			),

		Separator(),

		// Nested styling demonstration
		Label("🎭 Nested Styling:"),
		Style().
			SetColor(int(imgui.ColButton), ColorBlue).
			To(
				Row(
					Button("Blue 1"),
					Style().
						SetColor(int(imgui.ColButton), ColorYellow).
						SetColor(int(imgui.ColText), ColorBlack).
						To(
							Button("Yellow Override"),
						),
					Button("Blue 2"),
				),
			),

		Separator(),

		// Interactive styling
		Label("⚙️ Interactive Color Customization:"),
		Row(
			Column(
				Label("Custom Color:"),
				ColorEdit("Custom", &colorValue).OnChange(func() {
					customColor = RGB(colorValue[0]*255, colorValue[1]*255, colorValue[2]*255)
					LogStatus("Custom color updated!")
				}),
			),
			Column(
				Label("Styled Controls:"),
				Style().
					SetColor(int(imgui.ColSliderGrab), customColor).
					SetColor(int(imgui.ColSliderGrabActive), customColor).
					To(
						SliderFloat("Value", &sliderValue, 0.0, 100.0),
					),
			),
		),

		Separator(),

		// Advanced styling examples
		Label("🚀 Advanced Styling:"),
		Row(
			Style().
				SetColor(int(imgui.ColButton), RGB(255, 100, 100)).
				SetVar(int(imgui.StyleVarFrameRounding), 15.0).
				To(
					Button("Rounded Red"),
				),
			Style().
				SetColor(int(imgui.ColButton), RGB(100, 255, 100)).
				SetVar(int(imgui.StyleVarFrameRounding), 0.0).
				To(
					Button("Square Green"),
				),
			Style().
				SetColor(int(imgui.ColButton), RGB(100, 100, 255)).
				SetVar(int(imgui.StyleVarFrameRounding), 25.0).
				To(
					Button("Very Rounded Blue"),
				),
		),

		Separator(),

		// Status information
		func() Widget {
			themeName := "None"
			if currentThemeObject != nil {
				themeName = currentThemeObject.name
			}
			return Column(
				Label(fmt.Sprintf("✅ Current Global Theme: %s", themeName)),
				Label("✅ Individual widget styling: WORKING"),
				Label("✅ Nested styling: WORKING"),
				Label("✅ Stack management: FIXED"),
				Label("✅ Global theming: FIXED"),
			)
		}(),

		Separator(),

		// Event log with consistent styling
		Label("📝 Event Log:"),
		func() Widget {
			if globalStatus == nil {
				globalStatus = StatusDisplay().Height(120)
			}
			return globalStatus
		}(),

		Spacing(),
		Label("💡 Try switching themes to see global styling in action!"),
		Label("💡 Individual styled widgets override global theme colors."),
	).Build()
}

func main() {
	// Set initial theme
	SetGlobalTheme(DarkTheme)

	// Create master window
	wnd := NewMasterWindow("Step 10: Complete Styling System", 900, 700)

	// Run the application
	wnd.Run(loop)
}
