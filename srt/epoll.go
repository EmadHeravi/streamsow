package srt

// Epoll functionality is not yet wrapped. This file provides
// placeholders so the API surface is ready when needed.

// EpollID represents an epoll instance identifier.
type EpollID int

// EpollEvent represents epoll event flags.
type EpollEvent int

const (
	EpollIn  EpollEvent = 1 << 0
	EpollOut EpollEvent = 1 << 1
)

// EpollCreate is a stub for future epoll integration.
func EpollCreate() (EpollID, error) {
	return 0, nil
}

// EpollAdd is a stub for future epoll integration.
func EpollAdd(_ EpollID, _ *Socket, _ EpollEvent) error {
	return nil
}

// EpollWait is a stub for future epoll integration.
func EpollWait(_ EpollID, _ int) ([]*Socket, error) {
	return nil, nil
}
