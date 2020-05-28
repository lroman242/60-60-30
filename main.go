package main

import (
	"context"
	"fmt"
	//"github.com/faiface/beep"
	//"github.com/faiface/beep/mp3"
	//"github.com/faiface/beep/speaker"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
	"time"
)

const NUM_OF_WORK_PERIODS = 4
const WORK_DURATION_MIN = 25
const SHORT_BREAK_DURATION_MIN = 5
const FINAL_BREAK_DURATION = 30

//var audioStream beep.StreamSeekCloser
var isAudioEnabled bool = false
var isAutoRestartEnabled bool = false

func main() {
	//f, err := os.Open("assets/beep.mp3")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//audioStream, format, err := mp3.Decode(f)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(audioStream)
	//defer audioStream.Close()
	//speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))


	systray.Run(onReady, func() {
		fmt.Println("onExitHandler")
	})
}

// Run working cycle
func startCycle(ctx context.Context) {
	finalState := NUM_OF_WORK_PERIODS * 2

	for state := 1; state <= finalState; state++ {
		fmt.Println(state)
		//playBeep()
		if state == finalState {
			err := beeep.Alert("60-60-30 ЦИКЛЫ", "Отдохни 30мин", "assets/warning.png")
			if err != nil {
				panic(err)
			}

			select {
			// stop state received. break loop
			case <-ctx.Done():
				// stop func execution
				return
			case <-time.After(FINAL_BREAK_DURATION * time.Second):
				// do nothing
				//TODO: auto restart option + 3 x playBeep
			}
		} else if state%2 == 0 { // short break period
			err := beeep.Alert("60-60-30 ЦИКЛЫ", "Отдохни 5мин", "assets/warning.png")
			if err != nil {
				panic(err)
			}

			select {
			// stop state received. break loop
			case <-ctx.Done():
				// stop func execution
				return
			case <-time.After(SHORT_BREAK_DURATION_MIN * time.Second):
				// do nothing. go to next period
			}
		} else { // work period
			err := beeep.Alert("60-60-30 ЦИКЛЫ", "Начался рабочий цикл в 25мин", "assets/warning.png")
			if err != nil {
				panic(err)
			}

			select {
			// stop state received. break loop
			case <-ctx.Done():
				// stop func execution
				return
			case <-time.After(WORK_DURATION_MIN * time.Second):
				// do nothing. go to next period
			}
		}
	}
}

//func playBeep() {
//	fmt.Println(audioStream)
//
//	audioStream.Seek(0)
//	speaker.Play(beep.Seq(audioStream, beep.Callback(func() {})))
//}

func onReady() {
	ctx := context.Background()
	ctxWithCancel, _ := context.WithCancel(ctx)

	systray.SetTemplateIcon(Icon, Icon)
	systray.SetIcon(Icon)
	systray.SetTooltip("60-60-30 work cycle")

	btnStart := systray.AddMenuItem("Start", "Start Cycle")
	btnStop := systray.AddMenuItem("Stop", "Stop Current Cycle")
	btnRestart := systray.AddMenuItem("Restart", "Restart Cycle")
	systray.AddSeparator()
	btnEnableAutoRestart := systray.AddMenuItem("Enable Auto Restart", "Enable Auto Restart")
	btnDisableAutoRestart := systray.AddMenuItem("Disable Auto Restart", "Disable Auto Restart")

	btnEnableAudio := systray.AddMenuItem("Enable Audio Signal", "Enable Audio Signal")
	btnDisableAudio := systray.AddMenuItem("Disable Audio Signal", "Enable Audio Signal")
	systray.AddSeparator()
	btnQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	go func(ctx context.Context) {
		ctxWithCancel, cancelFunc := context.WithCancel(ctx)
		defer cancelFunc()

		btnStart.SetIcon(StartIcon)
		btnStart.Enable()

		btnStop.SetIcon(StopIcon)
		btnStop.Disable()

		btnRestart.SetIcon(RestartIcon)
		btnRestart.Disable()

		if isAudioEnabled {
			btnEnableAudio.Hide()
			btnDisableAudio.Show()
		} else {
			btnEnableAudio.Show()
			btnDisableAudio.Hide()
		}

		if isAutoRestartEnabled {
			btnEnableAutoRestart.Hide()
			btnDisableAutoRestart.Show()
		} else {
			btnEnableAutoRestart.Show()
			btnDisableAutoRestart.Hide()
		}

		for {
			select {
			// handle cancel from main
			case <-ctx.Done():
				cancelFunc()
				break
			case <-btnStart.ClickedCh:
				btnStart.Disable()
				btnStop.Enable()
				btnRestart.Enable()

				go startCycle(ctxWithCancel)
			case <-btnStop.ClickedCh:
				btnStart.Enable()
				btnStop.Disable()
				btnRestart.Disable()

				cancelFunc()

				// previous context has already been canceled, so new
				// context required to have ability to start new working cycle
				ctxWithCancel, cancelFunc = context.WithCancel(ctx)
			case <-btnRestart.ClickedCh:
				//Stop existed
				cancelFunc()

				// previous context has already been canceled, so new
				// context required to have ability to start new working cycle
				ctxWithCancel, cancelFunc = context.WithCancel(ctx)

				//Run new
				go startCycle(ctxWithCancel)
			case <-btnEnableAudio.ClickedCh:
				isAudioEnabled = true
				btnEnableAudio.Hide()
				btnDisableAudio.Show()
			case <-btnDisableAudio.ClickedCh:
				isAudioEnabled = false
				btnEnableAudio.Show()
				btnDisableAudio.Hide()
			case <-btnEnableAutoRestart.ClickedCh:
				isAutoRestartEnabled = true
				btnEnableAutoRestart.Hide()
				btnDisableAutoRestart.Show()
			case <-btnDisableAutoRestart.ClickedCh:
				isAutoRestartEnabled = false
				btnEnableAutoRestart.Show()
				btnDisableAutoRestart.Hide()
			case <-btnQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}(ctxWithCancel)
}
