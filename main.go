package main

import (
	"context"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gen2brain/beeep"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const NUM_OF_WORK_PERIODS = 4
const WORK_DURATION_MIN = 25
const SHORT_BREAK_DURATION_MIN = 5
const FINAL_BREAK_DURATION = 30

func main() {
	f, err := os.Open("assets/beep.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	ctx := context.Background()
	ctxWithCancel, cancelFunction := context.WithCancel(ctx)

	// Clean all resources
	defer func() {
		fmt.Println("main: canceling context")
		cancelFunction()
	}()

	// define state change what will be used to stop and start 60-60-30 working cycles
	stateChan := make(chan int, 1)

	// state changes handler
	go func(ctx context.Context, stateChan chan int) {
		ctxWithCancel, cancelFunction := context.WithCancel(ctx)

		// clean all resources
		defer func() {
			fmt.Println("handler: canceling context")
			cancelFunction()
		}()

		for {
			select {
			// handle cancel from main
			case <-ctx.Done():
				break
			// handle state changes
			case state := <-stateChan:
				// start cycle
				if state == 1 {
					fmt.Println("start new cycle")
					go startCycle(ctxWithCancel, streamer)
				} else if state == 0 { // stop cycle
					fmt.Println("Stop!")
					cancelFunction()

					// previous context has already been canceled, so new
					// context required to have ability to start new working cycle
					ctxWithCancel, cancelFunction = context.WithCancel(ctx)
				}
			}
		}
	}(ctxWithCancel, stateChan)

	//TODO: trait icon + trait menu handler with start, stop and exit
	stateChan <- 1

	// awaiting to exit signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	s := <-ch
	log.Printf("Got signal: %v, exiting.", s)
}

func startCycle(ctx context.Context, streamer beep.StreamSeekCloser) {
	finalState := NUM_OF_WORK_PERIODS * 2

	for state := 1; state <= finalState; state++ {
		streamer.Seek(0)
		speaker.Play(beep.Seq(streamer, beep.Callback(func() {})))

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
				//fmt.Printf("Final state %v finished\n", state)
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
				//fmt.Printf("Break state %v finished\n", state)
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
