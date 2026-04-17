package shell

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Notify sends a system notification.
// On macOS, it uses osascript (No CGO).
// On Linux, it uses notify-send (No CGO).
func Notify(title, message string) error {
	switch runtime.GOOS {
	case "darwin":
		// Added sound and escaped strings better
		script := fmt.Sprintf("display notification %q with title %q sound name \"Hero\"", message, title)
		return exec.Command("osascript", "-e", script).Run()
	case "linux":
		return exec.Command("notify-send", title, message).Run()
	default:
		fmt.Printf("NOTIFICATION: [%s] %s\n", title, message)
		return nil
	}
}

// SelectFile opens a native file picker.
func SelectFile(title string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		// Get POSIX path directly from AppleScript
		script := fmt.Sprintf("POSIX path of (choose file with prompt %q)", title)
		out, err := exec.Command("osascript", "-e", script).Output()
		if err != nil {
			// Check if it's a cancellation (User canceled. -128)
			if strings.Contains(err.Error(), "exit status 1") {
				return "", nil // Return empty, no error means silent cancel
			}
			return "", err
		}
		return strings.TrimSpace(string(out)), nil
	default:
		return "", fmt.Errorf("SelectFile not implemented for %s", runtime.GOOS)
	}
}

// OpenURL opens the system browser.
func OpenURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	default: // linux, freebsd, etc.
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Run()
}

// TrayIcon sets up the system tray (mocked).
func TrayIcon(title string, onExit func()) {
	fmt.Printf("SHELL TRAY: %s activated\n", title)
}
