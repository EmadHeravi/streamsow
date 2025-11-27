/*
 * SPDX-FileCopyrightText: Streamzeug Copyright ©
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package udp

import (
	"net"
	"time"

	"github.com/EmadHeravi/streamsow/logging"
	"github.com/EmadHeravi/streamsow/mainloop"
)

// StartReader starts the UDP socket listener and forwards packets into flow channel.
func (i *UdpInput) StartReader(ch chan<- mainloop.InputPacket) error {
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
				n, addr, err := conn.ReadFromUDP(buf)
				if err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout() {
						continue // timeout → retry
					}
					logger.Error().Err(err).Msg("UDP read error")
					continue
				}

				// Forward packet into mainloop
				pkt := mainloop.InputPacket{
					Data:      append([]byte(nil), buf[:n]...), // copy
					Timestamp: time.Now().UnixNano(),
					Source:    addr.String(),
				}

				select {
				case ch <- pkt:
				default:
					logger.Warn().Msg("mainloop channel full — dropping packet")
				}
			}
		}
	}()

	return nil
}
