package shell

import (
	"fmt"
)

// Info shows a system notification (stub for now to avoid dependency hell before validation)
func Info(title, message string) {
	fmt.Printf("SHELL NOTIFICATION: [%s] %s\n", title, message)
}

// OpenFile opens a native file dialog (stub)
func OpenFile(title string) (string, error) {
	fmt.Printf("SHELL DIALOG: %s\n", title)
	return "mock_file.txt", nil
}

// TrayIcon sets up the system tray (stub)
func TrayIcon(title string, onExit func()) {
	fmt.Printf("SHELL TRAY: %s activated\n", title)
}
