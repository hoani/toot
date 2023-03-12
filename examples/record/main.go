package main

import (
	"context"
	"fmt"
	"os"

	"github.com/faiface/beep/wav"
	"github.com/gordonklaus/portaudio"
	"github.com/hoani/toot"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// initialize toot
	terminate, err := toot.Initialize()
	if err != nil {
		panic(err)
	}
	defer terminate()

	devices, _ := portaudio.Devices()
	for i, d := range devices {
		fmt.Printf("%d: %#v\n", i, d)
	}
	os.Exit(0)

	m, err := toot.NewMicrophone()
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := m.Start(ctx); err != nil {
		panic(err)
	}

	a := toot.NewAnalyzer(m, int(m.Format().SampleRate))

	f, err := os.Create("test.wav")

	go func() {
		err = wav.Encode(f, a, m.Format())
		if err != nil {
			panic(err)
		}
	}()

	fmt.Print("\nPress [ENTER] to finish recording! ")
	fmt.Scanln()
	m.Close()

	fmt.Print("computing power series...\n")
	freqs, powerSeries := a.GetF0()
	fmt.Print("plotting...\n")
	Plot(freqs, powerSeries)
	fmt.Print("transcribing...\n")
	c := openai.NewClient(os.Getenv("OPENAI_KEY"))

	req := openai.AudioRequest{
		Model:    openai.Whisper1,
		FilePath: "./test.wav",
	}
	resp, err := c.CreateTranscription(ctx, req)
	if err != nil {
		fmt.Printf("Transcription error: %v\n", err)
		return
	}
	fmt.Println(resp.Text)

}
