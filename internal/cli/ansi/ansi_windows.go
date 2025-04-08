//go:build windows

package ansi

import (
	"os"

	"github.com/shirou/gopsutil/v3/process"
)

var isPwsh = isPowershell()

func EnterAltScreenBuf() {
	if isPwsh {
		enterAltScreenBufCommon()
	}
}

func ExitAltScreenBuf() {
	if isPwsh {
		exitAltScreenBufCommon()
	}
}

// TODO: сделать аккуратнее
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
