package toot

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/faiface/beep"
	"github.com/gordonklaus/portaudio"
)

var (
	initialized bool
	initlock    sync.Mutex
)

type Microphone interface {
	beep.Streamer
	Start(ctx context.Context) error
	Close() error
	Format() beep.Format
}

// Initialize initializes the mic library. We cannot instantiate a microphone without this.
// Returns a close function which must be called on exit, otherwise your program will leak resources.
func initialize() (close func() error, err error) {
	initlock.Lock()
	defer initlock.Unlock()
	if initialized {
		return nil, errors.New("mic already initialized")
	}
	if err := portaudio.Initialize(); err != nil {
		return nil, fmt.Errorf("mic failed to initialize portaudio: %w", err)
	}
	initialized = true
	return terminate, nil
}

func terminate() error {
	initlock.Lock()
	defer initlock.Unlock()
	initialized = false
	return portaudio.Terminate()
}

type mic struct {
	stream *portaudio.Stream
	ctx    context.Context
	cancel context.CancelFunc
	buffer chan float64
	wg     sync.WaitGroup
	err    error
	format beep.Format
}

func NewMicrophone() (Microphone, error) {
	initlock.Lock()
	defer initlock.Unlock()
	if !initialized {
		return nil, errors.New("mic library must be initialized with microphone.Initialize")
	}
	return &mic{
		buffer: make(chan float64),
	}, nil
}

func (m *mic) Stream(samples [][2]float64) (int, bool) {

	for i := range samples {
		select {
		case <-m.ctx.Done():
			return 0, false
		case v := <-m.buffer:
			samples[i][0] = v
			samples[i][1] = v
		}
	}
	return len(samples), true
}

func (m *mic) Err() error {
	return m.err
}

func (m *mic) Start(ctx context.Context) error {
	m.ctx, m.cancel = context.WithCancel(ctx)

	// get the default input device
	device, _ := portaudio.DefaultInputDevice()

	// set the sample rate and buffer size
	bufferSize := 512

	fmt.Printf("%v", device)

	m.format = beep.Format{
		SampleRate:  beep.SampleRate(device.DefaultSampleRate),
		NumChannels: 1,
		Precision:   3,
	}

	var err error
	m.stream, err = portaudio.OpenDefaultStream(1, 0, device.DefaultSampleRate, bufferSize,
		func(in []float32) {
			for _, input := range in {
				select {
				case m.buffer <- float64(input):
				case <-m.ctx.Done():
					return
				}
			}
		})
	if err != nil {
		return err
	}

	m.wg.Add(1)
	go func() {
		<-m.ctx.Done()
		if err := m.stream.Stop(); err != nil {
			m.err = err
		}
		if err = m.stream.Close(); err != nil {
			m.err = err
		}
		m.wg.Done()
	}()

	return m.stream.Start()
}

func (m *mic) Close() error {
	if m.cancel == nil {
		return errors.New("not started")
	}
	m.cancel()
	m.wg.Wait()
	return m.err
}

func (m *mic) Format() beep.Format {
	return m.format
}
