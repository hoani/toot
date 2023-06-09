package toot

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/faiface/beep"
	"github.com/gordonklaus/portaudio"
)

type Microphone interface {
	beep.Streamer
	Start(ctx context.Context) error
	Close() error
	Format() beep.Format
	DeviceInfo() portaudio.DeviceInfo
	SetSampleRate(rate beep.SampleRate)
}

type InputDevice struct {
	Name   string
	Stereo bool
}

type mic struct {
	stream     *portaudio.Stream
	ctx        context.Context
	cancel     context.CancelFunc
	buffer     chan float64
	wg         sync.WaitGroup
	err        error
	device     *portaudio.DeviceInfo
	devicelock sync.Mutex
	sampleRate beep.SampleRate
}

func GetInputDevices() ([]InputDevice, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}
	defer portaudio.Terminate()

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, err
	}

	details := make([]InputDevice, 0)
	for _, device := range devices {
		if device.MaxInputChannels > 0 {
			details = append(details, InputDevice{
				Name:   device.Name,
				Stereo: device.MaxInputChannels > 1,
			})
		}
	}
	return details, nil
}

func NewDefaultMicrophone() (Microphone, error) {
	return NewMicrophone("")
}

func NewMicrophone(name string) (Microphone, error) {
	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}

	var device *portaudio.DeviceInfo
	if name == "" {
		var err error
		if device, err = portaudio.DefaultInputDevice(); err != nil {
			return nil, err
		}
	} else {
		devices, err := portaudio.Devices()
		if err != nil {
			return nil, err
		}
		for _, d := range devices {
			if strings.Contains(d.Name, name) && d.MaxInputChannels > 0 {
				device = d
			}
		}
	}

	if device == nil {
		return nil, fmt.Errorf("could not find input device %s", name)
	}

	m := &mic{
		ctx:        context.Background(),
		device:     device,
		buffer:     make(chan float64),
		sampleRate: beep.SampleRate(device.DefaultSampleRate),
	}

	return m, nil
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

func (m *mic) DeviceInfo() portaudio.DeviceInfo {
	return *m.device
}

func (m *mic) Err() error {
	return m.err
}

func (m *mic) Start(ctx context.Context) error {
	m.ctx, m.cancel = context.WithCancel(ctx)

	// get the default input device

	// set the sample rate and buffer size
	bufferSize := 512

	var err error
	m.stream, err = portaudio.OpenStream(
		portaudio.StreamParameters{
			Input: portaudio.StreamDeviceParameters{
				Device:   m.device,
				Channels: 1,
				Latency:  m.device.DefaultLowInputLatency,
			},
			SampleRate:      float64(m.sampleRate),
			FramesPerBuffer: bufferSize,
		},
		func(in []float32) {
			for _, input := range in {
				select {
				case m.buffer <- float64(input):
				case <-m.ctx.Done():
					return
				}
			}
		},
	)
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
	m.devicelock.Lock()
	defer m.devicelock.Unlock()

	if m.device == nil {
		return nil // Already closed.
	}

	// Cancel streaming.
	if m.cancel == nil {
		return errors.New("not started")
	}
	m.cancel()
	m.wg.Wait()

	if err := portaudio.Terminate(); err != nil {
		m.err = err
	}
	m.device = nil

	return m.err
}

func (m *mic) Format() beep.Format {
	m.devicelock.Lock()
	defer m.devicelock.Unlock()
	return beep.Format{
		SampleRate:  m.sampleRate,
		NumChannels: 1,
		Precision:   3,
	}
}

func (m *mic) SetSampleRate(rate beep.SampleRate) {
	m.devicelock.Lock()
	defer m.devicelock.Unlock()
	m.sampleRate = rate
}
