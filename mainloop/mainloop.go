/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia B.V.
 * SPDX-FileContributor: Author: Gijs Peskens <gijs@peskens.net>
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package mainloop

import (
	"context"
	"sync"
	"time"

	"code.videolan.org/rist/ristgo"
	"github.com/EmadHeravi/streamsow/logging"
	"github.com/EmadHeravi/streamsow/output"
	"github.com/rs/zerolog"
)

// InputPacket previously used for UDP/RTP input has been deprecated.
// All input sources (UDP, SRT, RIST) are now normalized to RIST blocks
// before entering this mainloop.

// inputstatus tracks statistics for the primary input.
type inputstatus struct {
	packetcount        int
	packetcountsince   int
	bytesSince         int
	discontinuitycount int
	lastPacketTime     time.Time
}

// Mainloop is the central receiver loop that takes RIST blocks
// from a ristgo.ReceiverFlow and forwards them to registered outputs.
type Mainloop struct {
	ctx                context.Context
	flow               ristgo.ReceiverFlow
	logger             zerolog.Logger
	outputs            map[int]*out
	outPutAdd          chan output.Output
	outPutRemove       chan output.Output
	outRemoveIdx       chan int
	wg                 sync.WaitGroup
	statusLock         sync.Mutex
	primaryInputStatus inputstatus
	lastStatusCall     time.Time
}

// removeOutputByID schedules removal of an output by index.
func (m *Mainloop) removeOutputByID(idx int) {
	select {
	case <-m.ctx.Done():
		return
	default:
	}
	m.outRemoveIdx <- idx
}

// RemoveOutput schedules removal of an output by object.
func (m *Mainloop) RemoveOutput(o output.Output) {
	select {
	case <-m.ctx.Done():
		return
	default:
	}
	m.outPutRemove <- o
}

// deleteOutput closes the data channel and removes it from the map.
func (m *Mainloop) deleteOutput(idx int, o output.Output) {
	m.logger.Info().Msgf("deleting output: %s", o.String())
	close(m.outputs[idx].dataChan)
	delete(m.outputs, idx)
}

// AddOutput adds a new output writer to the mainloop.
func (m *Mainloop) AddOutput(o output.Output) {
	m.logger.Info().Msgf("adding output %s", o.String())
	select {
	case <-m.ctx.Done():
		return
	default:
	}
	m.outPutAdd <- o
}

// Wait blocks until the mainloop goroutines complete or timeout expires.
func (m *Mainloop) Wait(timeout time.Duration) {
	c := make(chan bool)
	go func() {
		m.wg.Wait()
		c <- true
	}()
	select {
	case <-c:
		return
	case <-time.After(timeout):
		return
	}
}

// NewMainloop wires a RIST ReceiverFlow into the main processing loop.
// All packet sources are normalized to RIST and appear in the same flow.
func NewMainloop(ctx context.Context, flow ristgo.ReceiverFlow, identifier string) *Mainloop {
	m := &Mainloop{
		ctx:          ctx,
		flow:         flow,
		logger:       logging.Log.With().Str("identifier", identifier).Logger(),
		outputs:      make(map[int]*out),
		outPutAdd:    make(chan output.Output, 4),
		outPutRemove: make(chan output.Output, 4),
		outRemoveIdx: make(chan int, 16),
	}
	go receiveLoop(m)
	return m
}

// receiveLoop consumes RIST data blocks from the flow and forwards to outputs.
func receiveLoop(m *Mainloop) {
	outputidx := 0
	m.primaryInputStatus.lastPacketTime = time.Now()
	m.lastStatusCall = m.primaryInputStatus.lastPacketTime
	expectedSeq := uint16(0)
	m.logger.Info().Msg("receiver mainloop started")
	m.wg.Add(1)
	lastDiscontinuityMsg := time.Time{}
	discontinuitiesSinceLastMsg := 0

main:
	for {
		select {
		case <-m.ctx.Done():
			break main

		// Unified RIST input path
		case rb, ok := <-m.flow.DataChannel():
			if !ok {
				break main
			}

			discontinuity := false
			if rb.Discontinuity {
				discontinuity = true
			}
			if rb.SeqNo != uint32(expectedSeq) {
				discontinuity = true
			}
			if discontinuity {
				m.primaryInputStatus.discontinuitycount++
				discontinuitiesSinceLastMsg++
			}

			if discontinuitiesSinceLastMsg > 0 &&
				time.Since(lastDiscontinuityMsg) >= 5*time.Second {
				m.logger.Error().
					Int("count", discontinuitiesSinceLastMsg).
					Msg("discontinuity!")
				lastDiscontinuityMsg = time.Now()
				discontinuitiesSinceLastMsg = 0
			}

			expectedSeq = uint16(rb.SeqNo) + 1

			m.statusLock.Lock()
			m.primaryInputStatus.packetcount++
			m.primaryInputStatus.packetcountsince++
			m.primaryInputStatus.lastPacketTime = time.Now()
			m.primaryInputStatus.bytesSince += len(rb.Data)
			m.statusLock.Unlock()

			m.writeOutputs(rb)

		case o := <-m.outPutAdd:
			m.statusLock.Lock()
			m.addOutput(o, outputidx)
			outputidx++
			m.statusLock.Unlock()

		case idx := <-m.outRemoveIdx:
			m.statusLock.Lock()
			o, ok := m.outputs[idx]
			if ok {
				m.deleteOutput(idx, o.w)
			} else {
				m.logger.Error().
					Msgf("couldn't delete output at index: %d, notfound", idx)
			}
			m.statusLock.Unlock()

		case o := <-m.outPutRemove:
			found := false
			m.statusLock.Lock()
			for idx, out := range m.outputs {
				if out.w == o {
					m.deleteOutput(idx, o)
					found = true
					break
				}
			}
			m.statusLock.Unlock()
			if !found {
				m.logger.Error().
					Msgf("couldn't delete output: %s, notfound", o.String())
			}
		}
	}

	close(m.outPutAdd)
	close(m.outPutRemove)
	close(m.outRemoveIdx)
	m.logger.Info().Msg("mainloop terminated")
	m.wg.Done()
}
