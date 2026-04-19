package input

import (
	"sync"
)

// Key represents a physical keyboard key.
type Key int

const (
	KeyUnknown Key = iota
	KeyA
	KeyW
	KeyS
	KeyD
	KeySpace
	KeyShiftLeft
	KeyControlLeft
	// ... more as needed
)

// Modifier represent sticky keys like Shift, Ctrl.
type Modifier int

const (
	ModShift Modifier = 1 << iota
	ModControl
	ModAlt
	ModSuper
)

// MouseButton represents a mouse click button.
type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonRight
	MouseButtonMiddle
)

// KeyboardState handles thread-safe polling of keys.
type KeyboardState struct {
	mu       sync.RWMutex
	current  map[Key]bool
	previous map[Key]bool
}

func newKeyboardState() KeyboardState {
	return KeyboardState{
		current:  make(map[Key]bool),
		previous: make(map[Key]bool),
	}
}

func (k *KeyboardState) SetKey(key Key, down bool) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.current[key] = down
}

func (k *KeyboardState) Pressed(key Key) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.current[key]
}

func (k *KeyboardState) JustPressed(key Key) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.current[key] && !k.previous[key]
}

func (k *KeyboardState) JustReleased(key Key) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return !k.current[key] && k.previous[key]
}

func (k *KeyboardState) Modifier(m Modifier) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	// Simplified modifier check for benchmarks
	if m == ModControl {
		return k.current[KeyControlLeft]
	}
	if m == ModShift {
		return k.current[KeyShiftLeft]
	}
	return false
}

func (k *KeyboardState) AnyPressed() bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	for _, down := range k.current {
		if down {
			return true
		}
	}
	return false
}

func (k *KeyboardState) UpdateFrame() {
	k.mu.Lock()
	defer k.mu.Unlock()
	for key, down := range k.current {
		k.previous[key] = down
	}
}

// MouseState handles thread-safe polling of mouse position and buttons.
type MouseState struct {
	mu           sync.RWMutex
	x, y         float32
	px, py       float32
	buttons      [3]bool
	prevButtons  [3]bool
	scrollX, sy float32
}

func newMouseState() MouseState {
	return MouseState{}
}

func (m *MouseState) SetPosition(x, y float32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.x, m.y = x, y
}

func (m *MouseState) Position() (float32, float32) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.x, m.y
}

func (m *MouseState) Delta() (float32, float32) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.x - m.px, m.y - m.py
}

func (m *MouseState) SetButton(b MouseButton, down bool) {
	if b < 0 || b >= 3 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.buttons[b] = down
}

func (m *MouseState) Pressed(b MouseButton) bool {
	if b < 0 || b >= 3 {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.buttons[b]
}

func (m *MouseState) JustPressed(b MouseButton) bool {
	if b < 0 || b >= 3 {
		return false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.buttons[b] && !m.prevButtons[b]
}

func (m *MouseState) SetScroll(x, y float32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.scrollX, m.sy = x, y
}

func (m *MouseState) Scroll() (float32, float32) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.scrollX, m.sy
}

func (m *MouseState) UpdateFrame() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.px, m.py = m.x, m.y
	m.prevButtons = m.buttons
	// Reset scroll deltas since they are usually per-frame
	m.scrollX, m.sy = 0, 0
}
