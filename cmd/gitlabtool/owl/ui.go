package main

import (
	"fmt"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

func runUILoop() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)

	drawSplash()
	<-ready
	redrawAll()

	// capture and process events from the CLI
	eventChan := make(chan termbox.Event, 16)
	go handleEvents(eventChan)
	go func() {
		for {
			ev := termbox.PollEvent()
			eventChan <- ev
		}
	}()

	for {
		select {
		case <-newEvt:
			drawNotify()
			time.Sleep(time.Second)
			redrawAll()

		case err := <-errCh:
			termbox.Close()
			fmt.Println(err)
			return

		case <-quit:
			return
		}
	}

}

func handleEvents(eventChan chan termbox.Event) {
	for ev := range eventChan {
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeySpace:
				// page down
				if !detailView {
					selectedRow += h
					lock.Lock()
					totalN := len(events)
					lock.Unlock()
					if selectedRow >= totalN {
						selectedRow = totalN - 1
					} else {
						page++
					}
					redrawAll()
				}
				continue

			case termbox.KeyEsc:
				if detailView {
					detailView = false
					redrawAll()
				} else {
					termbox.Close()
					os.Exit(0)
				}
			}

			switch ev.Ch {
			case 'j':
				lock.Lock()
				totalN := len(events)
				lock.Unlock()
				if selectedRow < totalN {
					selectedRow++
					if selectedRow%pageSize == 0 {
						page++
					}
					redrawAll()
				}

			case 'k':
				if selectedRow > 0 {
					if selectedRow%pageSize == 0 {
						page--
					}
					selectedRow--
					redrawAll()
				}

			case 'd':
				// detail page
				if detailView {
					redrawAll()
				} else {
					drawDetail()
				}
				detailView = !detailView

			case 'b':
				// page up
				if !detailView {
					selectedRow -= pageSize
					if selectedRow < 0 {
						selectedRow = 0
					} else {
						page--
					}
					redrawAll()
				}

			case 'q':
				if detailView {
					detailView = false
					redrawAll()
				} else {
					termbox.Close()
					os.Exit(0)
				}

			}

		case termbox.EventError:
			panic(ev.Err)

		}
	}
}