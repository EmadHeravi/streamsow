/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia B.V. All right reserved.
 * SPDX-FileContributor: Author: Gijs Peskens <gijs@peskens.net>
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package mainloop

import (
	"context"
	"sync"
	"time"

	"code.videolan.org/rist/ristgo"
	"github.com/odmedia/streamzeug/logging"
	"github.com/odmedia/streamzeug/output"
	"github.com/rs/zerolog"
)

// InputPacket is a generic packet structure used for non-RIST
// inputs such as plain UDP/RTP. It allows UDP readers to push
// data into the mainloop without changing the existing RIST
// receiver and output wiring.
type InputPacket struct {
	Data      []byte
	Timestamp int64
	Source    string
}

type inputstatus struct {
	packetcount        int
	packetcountsince   int
	bytesSince         int
	discontinuitycount int
	lastPacketTime     time.Time
}

type Mainloop struct {
	ctx                context.Context
	flow               ristgo.ReceiverFlow
	logger             zerolog.Logger
	outputs            map[int]*out
	outPutAdd          chan output.Output
	outPutRemove       chan output.Output
	outRemoveIdx       chan int
	udpChan            chan InputPacket
	wg                 sync.WaitGroup
	statusLock         sync.Mutex
	primaryInputStatus inputstatus
	lastStatusCall     time.Time
}

// UDPChannel returns the internal UDP packet channel. UDP inputs
// should send InputPacket instances to this channel from their
// reader goroutines.
func (m *Mainloop) UDPChannel() chan<- InputPacket {
	return m.udpChan
}

func (m *Mainloop) removeOutputByID(idx int) {
	select {
	case <-m.ctx.Done():
		return
	default:
	}
	m.outRemoveIdx <- idx
}

func (m *Mainloop) RemoveOutput(o output.Output) {
	select {
	case <-m.ctx.Done():
		return
	default:
	}
	m.outPutRemove <- o
}

func (m *Mainloop) deleteOutput(idx int, o output.Output) {
	m.logger.Info().Msgf("deleting output: %s", o.String())
	close(m.outputs[idx].dataChan)
	delete(m.outputs, idx)
}

func (m *Mainloop) AddOutput(o output.Output) {
	m.logger.Info().Msgf("adding output %s", o.String())
	select {
	case <-m.ctx.Done():
		return
	default:
	}
	m.outPutAdd <- o
}

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
// UDP/RTP inputs can additionally push packets into udpChan for stats.
func NewMainloop(ctx context.Context, flow ristgo.ReceiverFlow, identifier string) *Mainloop {
	m := &Mainloop{
		ctx:          ctx,
		flow:         flow,
		logger:       logging.Log.With().Str("identifier", identifier).Logger(),
		outputs:      make(map[int]*out),
		outPutAdd:    make(chan output.Output, 4),
		outPutRemove: make(chan output.Output, 4),
		outRemoveIdx: make(chan int, 16),
		udpChan:      make(chan InputPacket, 512),
	}
	go receiveLoop(m)
	return m
}

func receiveLoop(m *Mainloop) {
	outputidx := 0
	m.primaryInputStatus.lastPacketTime = time.Now()
	m.lastStatusCall = m.primaryInputStatus.lastPacketTime
	expectedSec := uint16(0)
	m.logger.Info().Msg("receiver mainloop started")
	m.wg.Add(1)
	lastDiscontinuityMsg := time.Time{}
	discontinuitiesSinceLastMsg := 0

main:
	for {
		select {
		case <-m.ctx.Done():
			break main

		// RIST receiver path (existing behaviour)
		case rb, ok := <-m.flow.DataChannel():
			if !ok {
				break main
			}
			discontinuity := false
			if rb.Discontinuity {
				discontinuity = true
			}
			if rb.SeqNo != uint32(expectedSec) {
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

			expectedSec = uint16(rb.SeqNo) + 1

			m.statusLock.Lock()
			m.primaryInputStatus.packetcount++
			m.primaryInputStatus.packetcountsince++
			m.primaryInputStatus.lastPacketTime = time.Now()
			m.primaryInputStatus.bytesSince += len(rb.Data)
			m.statusLock.Unlock()

			m.writeOutputs(rb)

		// UDP/RTP packet path (stats only, no output write)
		case pkt := <-m.udpChan:
			if pkt.Data == nil {
				break
			}
			m.statusLock.Lock()
			m.primaryInputStatus.packetcount++
			m.primaryInputStatus.packetcountsince++
			m.primaryInputStatus.lastPacketTime = time.Now()
			m.primaryInputStatus.bytesSince += len(pkt.Data)
			m.statusLock.Unlock()

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
	close(m.udpChan)
	m.logger.Info().Msg("mainloop terminated")
	m.wg.Done()
}
