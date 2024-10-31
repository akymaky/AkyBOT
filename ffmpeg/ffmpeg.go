package ffmpeg

import (
	"bufio"
	"io"
	"os/exec"
	"strconv"

	"github.com/disgoorg/disgo/voice"
	"github.com/jonas747/ogg"
)

const (
	// Exec is the default path to the ffmpeg executable
	Exec       = "ffmpeg"
	Channels   = 2
	SampleRate = 48000
	BufferSize = 65307
)

var _ voice.OpusFrameProvider = (*AudioProvider)(nil)

func New(uri string, opts ...ConfigOpt) (*AudioProvider, error) {
	cfg := DefaultConfig()
	cfg.Apply(opts)

	cmd := exec.Command(cfg.Exec,
		"-i",
		uri,
		"-c:a", "libopus",
		"-ac", strconv.Itoa(cfg.Channels),
		"-ar", strconv.Itoa(cfg.SampleRate),
		"-f", "ogg",
		"-b:a",
		"96K",
		"pipe:1",
	)

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return nil, err
	}

	return &AudioProvider{
		cmd:  cmd,
		pipe: pipe,
		d:    ogg.NewPacketDecoder(ogg.NewDecoder(bufio.NewReaderSize(pipe, cfg.BufferSize))),
	}, nil
}

type AudioProvider struct {
	cmd  *exec.Cmd
	pipe io.Closer
	d    *ogg.PacketDecoder
}

func (p *AudioProvider) Close() {
	return
}

func (p *AudioProvider) ProvideOpusFrame() ([]byte, error) {
	data, _, err := p.d.Decode()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (p *AudioProvider) Wait() error {
	return p.cmd.Wait()
}
