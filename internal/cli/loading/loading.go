package loading

import (
	"anicliru/internal/cli/ansi"
	"fmt"
	"sync"
	"time"
)

func DisplayLoading(quitChan chan bool, wg *sync.WaitGroup) {
	flowerPhases := []string{"", "*", "❀", "🌸"}
	phasesCount := len(flowerPhases)
	bloomLen := len(flowerPhases[phasesCount-1])
	flowersMax := 3
	loadingStr := "Поиск аниме по вашему запросу... "

	ansi.HideCursor()
	fmt.Print(loadingStr)

	defer ansi.ShowCursor()
	defer ansi.ClearLine()
	defer wg.Done()

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
		for i := 0; i < flowersMax; i++ {
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
		for i := 0; i < flowersMax-1; i++ {
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
