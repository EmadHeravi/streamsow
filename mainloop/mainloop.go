/*
 * SPDX-FileCopyrightText: Streamzeug Copyright ©
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package mainloop

import (
	"context"
	"sync"
	"time"

	"code.videolan.org/rist/ristgo"
	libristwrapper "code.videolan.org/rist/ristgo/libristwrapper"

	"github.com/odmedia/streamzeug/logging"
	"github.com/odmedia/streamzeug/output"
	"github.com/rs/zerolog"
)

// -------------------------------
// Generic UDP/RTP packet wrapper
// -------------------------------
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

// -------------------------------
// Mainloop struct
// -------------------------------
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

// -------------------------------
// UDP Channel
// -------------------------------
func (m *Mainloop) UDPChannel() chan<- InputPacket {
	return m.udpChan
}

// -------------------------------
// Output manipulation
// -------------------------------
func (m *Mainloop) removeOutputByID(idx int) {
	select {
	case <-m.ctx.Done():
		return
	default:
	}
	m.outRemoveIdx <- idx
}

func (m *Mainloop) RemoveOutput(output output.Output) {
	select {
	case <-m.ctx.Done():
		return
	default:
	}
	m.outPutRemove <- output
}

func (m *Mainloop) deleteOutput(idx int, output output.Output) {
	m.logger.Info().Msgf("deleting output: %s", output.String())
	close(m.outputs[idx].dataChan)
	delete(m.outputs, idx)
}

func (m *Mainloop) AddOutput(output output.Output) {
	m.logger.Info().Msgf("adding output %s", output.String())
	select {
	case <-m.ctx.Done():
		return
	default:
	}
	m.outPutAdd <- output
}

func (m *Mainloop) Wait(timeout time.Duration) {
	c := make(chan bool)
	go func() {
		m.wg.Wait()
		c <- true
	}()
	select {
	case <-c:
	case <-time.After(timeout):
	}
}

// -------------------------------
// Create new mainloop
// -------------------------------
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

// /////////////////////////////////////////////
// MPEG-TS continuity counter helper
// /////////////////////////////////////////////
func detectTsDiscontinuity(pkt []byte, lastCC *int) bool {
	if len(pkt) < 188 {
		return false
	}
	if pkt[0] != 0x47 {
		return false
	}

	cc := int(pkt[3] & 0x0F)

	if *lastCC == -1 {
		*lastCC = cc
		return false
	}

	expected := (*lastCC + 1) & 0x0F
	*lastCC = cc

	return cc != expected
}

// /////////////////////////////////////////////
// Main receiver loop
// /////////////////////////////////////////////
func receiveLoop(m *Mainloop) {
	outputidx := 0
	m.primaryInputStatus.lastPacketTime = time.Now()
	m.lastStatusCall = m.primaryInputStatus.lastPacketTime

	expectedSec := uint16(0)
	lastUdpCC := -1 // NEW: UDP TS CC tracker
	lastDiscontinuityMsg := time.Time{}
	discontinuitiesSinceLastMsg := 0

	m.logger.Info().Msg("receiver mainloop started")
	m.wg.Add(1)

main:
	for {
		select {

		case <-m.ctx.Done():
			break main

		// --------------------------------------------------------
		// RIST INPUT — unchanged behaviour
		// --------------------------------------------------------
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
				m.logger.Error().Int("count", discontinuitiesSinceLastMsg).Msg("discontinuity!")
				lastDiscontinuityMsg = time.Now()
				discontinuitiesSinceLastMsg = 0
			}

			expectedSec = uint16(rb.SeqNo) + 1

			m.statusLock.Lock()
			m.primaryInputStatus.packetcount++
			m.primaryInputStatus.packetcountsince++
			m.primaryInputStatus.bytesSince += len(rb.Data)
			m.primaryInputStatus.lastPacketTime = time.Now()
			m.statusLock.Unlock()

			m.writeOutputs(rb)

		// --------------------------------------------------------
		// UDP INPUT — NEW full forwarding + TS discontinuity
		// --------------------------------------------------------
		case pkt := <-m.udpChan:

			if pkt.Data == nil {
				break
			}

			// --- new discontinuity detection ---
			if detectTsDiscontinuity(pkt.Data, &lastUdpCC) {
				m.primaryInputStatus.discontinuitycount++
				discontinuitiesSinceLastMsg++
			}

			m.statusLock.Lock()
			m.primaryInputStatus.packetcount++
			m.primaryInputStatus.packetcountsince++
			m.primaryInputStatus.bytesSince += len(pkt.Data)
			m.primaryInputStatus.lastPacketTime = time.Now()
			m.statusLock.Unlock()

			// Wrap raw UDP packet into RistDataBlock for outputs
			rb := &libristwrapper.RistDataBlock{}
			rb.Data = append([]byte(nil), pkt.Data...)

			m.writeOutputs(rb)

		// --------------------------------------------------------
		// OUTPUT MANAGEMENT
		// --------------------------------------------------------
		case output := <-m.outPutAdd:
			m.statusLock.Lock()
			m.addOutput(output, outputidx)
			outputidx++
			m.statusLock.Unlock()

		case idx := <-m.outRemoveIdx:
			m.statusLock.Lock()
			output, ok := m.outputs[idx]
			if ok {
				m.deleteOutput(idx, output.w)
			} else {
				m.logger.Error().Msgf("couldn't delete output at index %d", idx)
			}
			m.statusLock.Unlock()

		case output := <-m.outPutRemove:
			found := false
			m.statusLock.Lock()
			for idx, o := range m.outputs {
				if o.w == output {
					m.deleteOutput(idx, output)
					found = true
					break
				}
			}
			m.statusLock.Unlock()
			if !found {
				m.logger.Error().Msgf("couldn't delete output: %s", output.String())
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
