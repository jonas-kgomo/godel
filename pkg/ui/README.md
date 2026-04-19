# Gödel UI System

The Gödel UI system is a high-performance, reactive component library built on top of Cogent Core. It combines the speed of GPU-accelerated rendering via WebGPU with a modern, signal-based state management system.

## Design Philosophy

- **Zero-CGO**: All components run using system-native bridges.
- **Signal-First**: State is managed via `Signal` and `Computed` types for fine-grained reactivity.
- **AutoPilot Ready**: Every widget can be targetable by the engine's simulation ghost using `.ID()`.
- **Modern Aesthetic**: Built-in support for HSL/Hex colors, cards, and smooth spacing.

## Core Primitives

### Layout
- `HStack(...)`: Horizontal flex container.
- `VStack(...)`: Vertical flex container.
- `Spacer(w, h)`: Rigid or flexible spacing.
- `Container(w)`: Versatile wrapper for padding, backgrounds, and rounding.

### Typography
- `Label(text)`: Standard static text.
- `LabelSignal(signal)`: Reactive text that updates automatically when the signal changes.

### Interactions
- `Button(cfg)`: Fully styleable button with reactive states.
- `TextInput(cfg)`: Dual-pipeline text input (filtered against hardware ghosting).
- `CheckBox(cfg)`: Reactive toggle state.

### Advanced
- `Dynamic(signal)`: The "View Portal". Swaps entire widget trees based on state changes.
- `ScrollView(w)`: GPU-accelerated clipped scrolling.

## Simulation & Testing

Gödel UI is uniquely designed for automation. You can label any high-level widget with an ID for use in simulation scripts:

```go
ui.WithID("login-btn", ui.Button(...))
```

The AutoPilot engine will use these IDs to calculate screen coordinates and inject virtual hardware events.
