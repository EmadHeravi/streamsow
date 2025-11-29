/*
 * SPDX-FileCopyrightText: Streamzeug Copyright ©
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package udp

import (
	"net"
	"time"

	"github.com/EmadHeravi/streamsow/include_srt/libristwrapper"
	"github.com/EmadHeravi/streamsow/input/normalizer"
	"github.com/EmadHeravi/streamsow/logging"
)

// StartReader starts the UDP socket listener and forwards
// packets into the unified RIST flow using a local RIST sender.
func (i *UdpInput) StartReader() error {
	logger := logging.Log.With().
		Str("module", "udp-reader").
		Str("identifier", i.identifier).
		Logger()

	// ----------- Resolve UDP Addr ------------
	udpAddr, err := net.ResolveUDPAddr("udp", i.url.Host)
	if err != nil {
		logger.Error().Err(err).Msg("failed to resolve UDP address")
		return err
	}

	// ----------- Listen ----------------------
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		logger.Error().Err(err).Msg("failed to open UDP socket")
		return err
	}

	logger.Info().Msgf("UDP listening on %s", i.url.Host)

	// ----------- Initialize RIST Sender ------------
	sender := libristwrapper.InitSender(0) // profile 0 = simple main profile
	defer sender.Free()

	// Optionally configure the local receiver address if needed
	// (e.g., "udp://127.0.0.1:9000" where your RIST receiver flow listens)
	sender.SetOutputIP("udp://127.0.0.1:9000")

	// ----------- Reader Loop -----------------
	go func() {
		defer conn.Close()

		buf := make([]byte, 2048) // MPEG-TS fits in 1316 but 2k safer
		for {
			select {
			case <-i.ctx.Done():
				logger.Info().Msg("UDP reader shutting down")
				return

			default:
				// Non-blocking read with timeout
				conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
				n, _, err := conn.ReadFromUDP(buf)
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						continue // timeout → retry
					}
					logger.Error().Err(err).Msg("UDP read error")
					continue
				}

				// Wrap UDP data into a RIST-compatible block
				rb := normalizer.WrapToRist(buf[:n])
				if rb == nil {
					continue
				}

				// Send to RIST flow (loopback or configured peer)
				ret := sender.SendData(rb)
				rb.Return()

				if ret != 0 {
					logger.Warn().Msg("RIST send returned non-zero (dropped or error)")
				}
			}
		}
	}()

	return nil
}
