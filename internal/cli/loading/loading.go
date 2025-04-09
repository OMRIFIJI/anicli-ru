package loading

import (
	"fmt"
	"time"

	"github.com/OMRIFIJI/anicli-ru/internal/cli/ansi"
)

func RestoreTerminal() {
	defer ansi.ShowCursor()
	defer ansi.ClearLine()
}

func DisplayLoading(quitChan chan struct{}) {
	flowerPhases := []string{"", "*", "❀", "🌸"}
	phasesCount := len(flowerPhases)
	bloomLen := len(flowerPhases[phasesCount-1])
	flowersMax := 3
	loadingStr := "Поиск аниме по вашему запросу... "

	ansi.HideCursor()
	fmt.Print(loadingStr)

	for {
		printSuccess := func(j int) bool {
			select {
			case <-quitChan:
				return false
			case <-time.After(bloomPhaseSleep * time.Millisecond):
				ansi.ClearLine()
				fmt.Print(loadingStr + flowerPhases[j])
				return true
			}
		}

		// Рост
		for range flowersMax {
			for j := 1; j < len(flowerPhases); j++ {
				if !printSuccess(j) {
					return
				}
			}
			loadingStr += flowerPhases[phasesCount-1]
		}

		// Увядание с развернутым циклом для балансировки сна
		loadingStr = loadingStr[:len(loadingStr)-bloomLen]
		for j := len(flowerPhases) - 2; j > 0; j-- {
			if !printSuccess(j) {
				return
			}
		}
		for range flowersMax - 1 {
			loadingStr = loadingStr[:len(loadingStr)-bloomLen]
			for j := len(flowerPhases) - 1; j > 0; j-- {
				if !printSuccess(j) {
					return
				}

			}
		}

		// Вывод текста без значков загрузки
		select {
		case <-quitChan:
			return
		case <-time.After(bloomPhaseSleep * time.Millisecond):
			ansi.ClearLine()
			fmt.Print(loadingStr)
		}
	}
}
