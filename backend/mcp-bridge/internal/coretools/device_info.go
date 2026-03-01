package coretools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

// DeviceInfoTool returns information about the local device.
type DeviceInfoTool struct{}

func NewDeviceInfoTool() *DeviceInfoTool {
	return &DeviceInfoTool{}
}

func (t *DeviceInfoTool) Name() string { return "get_device_info" }

func (t *DeviceInfoTool) Description() string {
	return "Retrieve information about the local device this companion is running on. " +
		"Returns the hostname (device name), operating system, CPU architecture, " +
		"current username, home directory, working directory, default shell, " +
		"and whether the process is running with elevated (root/sudo) privileges. " +
		"Use this to understand the environment before running commands, " +
		"to tailor file paths and package manager commands to the OS, " +
		"or to identify the machine in multi-device setups."
}

func (t *DeviceInfoTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
}

func (t *DeviceInfoTool) Execute(_ context.Context, _ map[string]interface{}) (string, error) {
	var sb strings.Builder

	// Hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "(unknown)"
	}
	fmt.Fprintf(&sb, "Hostname:      %s\n", hostname)

	// OS and architecture
	fmt.Fprintf(&sb, "OS:            %s\n", runtime.GOOS)
	fmt.Fprintf(&sb, "Architecture:  %s\n", runtime.GOARCH)

	// Kernel/OS version (best-effort)
	if version := getOSVersion(); version != "" {
		fmt.Fprintf(&sb, "OS Version:    %s\n", version)
	}

	// Current user
	currentUser, err := user.Current()
	if err == nil {
		fmt.Fprintf(&sb, "Username:      %s\n", currentUser.Username)
		fmt.Fprintf(&sb, "Home:          %s\n", currentUser.HomeDir)
		if currentUser.Name != "" && currentUser.Name != currentUser.Username {
			fmt.Fprintf(&sb, "Display Name:  %s\n", currentUser.Name)
		}
	}

	// If running under sudo, show the original user
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		fmt.Fprintf(&sb, "Sudo User:     %s (running as root via sudo)\n", sudoUser)
	}

	// Elevated privileges
	isRoot := os.Getuid() == 0
	fmt.Fprintf(&sb, "Elevated:      %v\n", isRoot)

	// Working directory
	if wd, err := os.Getwd(); err == nil {
		fmt.Fprintf(&sb, "Working Dir:   %s\n", wd)
	}

	// Shell
	if shell := os.Getenv("SHELL"); shell != "" {
		fmt.Fprintf(&sb, "Shell:         %s\n", shell)
	}

	// Number of CPUs
	fmt.Fprintf(&sb, "CPU Cores:     %d\n", runtime.NumCPU())

	return sb.String(), nil
}

// getOSVersion returns a human-readable OS version string.
func getOSVersion() string {
	switch runtime.GOOS {
	case "linux":
		// Try /etc/os-release first
		data, err := os.ReadFile("/etc/os-release")
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					val := strings.TrimPrefix(line, "PRETTY_NAME=")
					val = strings.Trim(val, "\"")
					return val
				}
			}
		}
		// Fallback to uname
		out, err := exec.Command("uname", "-r").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	case "darwin":
		out, err := exec.Command("sw_vers", "-productVersion").Output()
		if err == nil {
			return "macOS " + strings.TrimSpace(string(out))
		}
	case "windows":
		out, err := exec.Command("cmd", "/c", "ver").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return ""
}
