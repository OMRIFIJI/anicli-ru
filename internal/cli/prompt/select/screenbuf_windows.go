//go:build windows

package promptselect

import (
	"anicliru/internal/cli/ansi"
	"os"

	"github.com/shirou/gopsutil/v3/process"
)

var isPwsh = isPowershell()

func enterAltScreenBuf() {
	if isPwsh {
		ansi.EnterAltScreenBuf()
	}
}

func exitAltScreenBuf() {
	if isPwsh {
		ansi.ExitAltScreenBuf()
	}
}

// Сделать поаккуратнее
func isPowershell() bool {
	currentProcess, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return false
	}

	parentProcess, err := currentProcess.Parent()
	if err != nil {
		return false
	}

	parentName, err := parentProcess.Name()
	if err != nil {
		return false
	}

	return parentName == "powershell.exe" || parentName == "pwsh.exe"
}
