package srt

import (
	"errors"
	"net"

	"github.com/haivision/srtgo"
)

// Socket is a wrapper around srtgo.SrtSocket.
type Socket struct {
	inner *srtgo.SrtSocket
}

// NewSocket creates a new SRT socket with the given host, port and options.
func NewSocket(host string, port uint16, options map[string]string) (*Socket, error) {
	s := srtgo.NewSrtSocket(host, port, options)
	if s == nil {
		return nil, errors.New("srt: NewSocket returned nil underlying socket")
	}
	return &Socket{inner: s}, nil
}

// Mode returns the mode of the socket.
func (s *Socket) Mode() Mode {
	if s == nil || s.inner == nil {
		return ModeFailure
	}
	return Mode(s.inner.Mode())
}

// Listen puts the socket into listening mode.
func (s *Socket) Listen(backlog int) error {
	if s == nil || s.inner == nil {
		return errors.New("srt: Listen on nil socket")
	}
	return s.inner.Listen(backlog)
}

// Accept waits for an incoming connection and returns a new Socket.
func (s *Socket) Accept() (*Socket, *net.UDPAddr, error) {
	if s == nil || s.inner == nil {
		return nil, nil, errors.New("srt: Accept on nil socket")
	}
	newS, addr, err := s.inner.Accept()
	if err != nil {
		return nil, addr, err
	}
	if newS == nil {
		return nil, addr, errors.New("srt: Accept returned nil underlying socket")
	}
	return &Socket{inner: newS}, addr, nil
}

// Connect connects the socket in caller mode.
func (s *Socket) Connect() error {
	if s == nil || s.inner == nil {
		return errors.New("srt: Connect on nil socket")
	}
	return s.inner.Connect()
}

// Close closes the socket.
func (s *Socket) Close() {
	if s == nil || s.inner == nil {
		return
	}
	s.inner.Close()
}

// Write writes data to the socket.
func (s *Socket) Write(b []byte) (int, error) {
	if s == nil || s.inner == nil {
		return 0, errors.New("srt: Write on nil socket")
	}
	return s.inner.Write(b)
}

// Stats retrieves the socket statistics.
func (s *Socket) Stats() (*Stats, error) {
	if s == nil || s.inner == nil {
		return nil, errors.New("srt: Stats on nil socket")
	}
	st, err := s.inner.Stats()
	if err != nil {
		return nil, err
	}
	return (*Stats)(st), nil
}

// Underlying exposes the underlying srtgo.SrtSocket for internal use.
func (s *Socket) Underlying() *srtgo.SrtSocket {
	if s == nil {
		return nil
	}
	return s.inner
}
