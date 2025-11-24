package srt

import "github.com/haivision/srtgo"

// SRTErrno wraps the srtgo.SRTErrno type.
type SRTErrno = srtgo.SRTErrno

// Common SRT error codes used by streamzeug.
const (
	ErrNoConn  SRTErrno = srtgo.ENoConn
	ErrInvSock SRTErrno = srtgo.EInvSock
)

// SocketClosed is an alias for the SrtSocketClosed error type.
type SocketClosed = srtgo.SrtSocketClosed
