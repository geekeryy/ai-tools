package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/gordonklaus/portaudio"
)

var (
	recording bool
	wg        sync.WaitGroup
	stopChan  chan struct{}
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Audio Recorder")

	startButton := widget.NewButton("Start Recording", func() {
		if recording {
			return
		}

		startRecording()
	})

	stopButton := widget.NewButton("Stop Recording", func() {
		if !recording {
			return
		}

		stopRecording()
	})

	myWindow.SetContent(container.NewVBox(
		startButton,
		stopButton,
	))

	myWindow.ShowAndRun()
}

func startRecording() {
	recording = true
	stopChan = make(chan struct{})
	wg.Add(1)

	go func() {
		defer wg.Done()

		// Initialize PortAudio
		portaudio.Initialize()
		defer portaudio.Terminate()

		sampleRate := 44100
		bufferSize := 2048
		input := make([]int16, bufferSize)

		// Open audio stream
		stream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), len(input), &input)
		if err != nil {
			fmt.Println("Error opening stream:", err)
			return
		}
		defer stream.Close()

		// Create WAV file
		outFile, err := os.Create("output.wav")
		if err != nil {
			fmt.Println("Error creating WAV file:", err)
			return
		}
		defer outFile.Close()

		encoder := wav.NewEncoder(outFile, sampleRate, 16, 1, 1)

		if err := stream.Start(); err != nil {
			fmt.Println("Error starting stream:", err)
			return
		}

		fmt.Println("Recording...")

		for {
			select {
			case <-stopChan:
				// 确保 WAV 编码器正确关闭
				if err := encoder.Close(); err != nil {
					fmt.Println("Error closing encoder:", err)
				}
				return
			default:
				if err := stream.Read(); err != nil {
					fmt.Println("Error reading stream:", err)
					return
				}

				// Convert int16 data to int
				intData := make([]int, len(input))
				for i, v := range input {
					intData[i] = int(v)
				}

				intBuffer := &audio.IntBuffer{Data: intData, Format: &audio.Format{SampleRate: sampleRate, NumChannels: 1}}

				if err := encoder.Write(intBuffer); err != nil {
					fmt.Println("Error writing to WAV file:", err)
					return
				}

				time.Sleep(20 * time.Millisecond)
			}
		}
	}()
}

func stopRecording() {
	close(stopChan)
	recording = false

	wg.Wait()
	fmt.Println("Recording stopped. Saved as output.wav.")
}
