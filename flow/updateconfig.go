/*
 * SPDX-FileCopyrightText: Streamzeug Copyright © 2021 ODMedia B.V. All right reserved.
 * SPDX-FileContributor: Author: Gijs Peskens <gijs@peskens.net>
 * SPDX-License-Identifier: GPL-3.0-or-later
 */

package flow

import (
	"reflect"
	"time"

	"github.com/odmedia/streamzeug/config"
	"github.com/odmedia/streamzeug/logging"
)

func (f *Flow) UpdateConfig(c *config.Flow) (err error) {
	f.configLock.Lock()
	shouldUnlock := true
	defer func() {
		if shouldUnlock {
			f.configLock.Unlock()
		}
	}()

	// Nothing changed
	if reflect.DeepEqual(f.config, *c) {
		return nil
	}

	logging.Log.Info().
		Str("identifier", f.config.Identifier).
		Msg("updating flow config")

	defer func() {
		if err == nil {
			logging.Log.Info().
				Str("identifier", f.config.Identifier).
				Msg("done updating config")
			return
		}
		logging.Log.Error().
			Str("identifier", f.config.Identifier).
			Err(err).
			Msgf("error configuring: %s", err)
	}()

	// RIST receiver settings changed: rebuild the whole flow
	if c.Latency != f.config.Latency ||
		c.RistProfile != f.config.RistProfile ||
		c.StreamID != f.config.StreamID {

		logging.Log.Info().
			Str("identifier", f.config.Identifier).
			Msg("rist settings changed, re-creating")

		f.Stop()
		f.Wait(5 * time.Millisecond)

		newflow, err := CreateFlow(f.rcontext, c)
		if err != nil {
			return err
		}
		shouldUnlock = false
		*f = *newflow
		return nil
	}

	// -----------------------------
	// INPUTS: add / remove / update
	// -----------------------------
	if !reflect.DeepEqual(c.Inputs, f.config.Inputs) {
		checkDelete := make(map[string]int)
		for _, ic := range c.Inputs {
			checkDelete[ic.URL] = 1
		}

		// Remove inputs that disappeared
		for url, in := range f.configuredInputs {
			if _, ok := checkDelete[url]; !ok {
				in.Close()
				delete(f.configuredInputs, url)
			}
		}

		// Add new inputs
		for _, ic := range c.Inputs {
			if _, ok := f.configuredInputs[ic.URL]; !ok {
				if err := f.setupInput(&ic); err != nil {
					return err
				}
			}
		}
	}

	f.config.Inputs = c.Inputs

	// If after input changes the configs are equal, we’re done
	if reflect.DeepEqual(f.config, *c) {
		return nil
	}

	// ------------------------------
	// OUTPUTS: add / remove / update
	// ------------------------------
	if !reflect.DeepEqual(c.Outputs, f.config.Outputs) {
		checkDelete := make(map[string]int)
		for _, oc := range c.Outputs {
			checkDelete[oc.URL] = 1
		}

		// Remove outputs that disappeared
		for url, oh := range f.configuredOutputs {
			if _, ok := checkDelete[url]; !ok {
				oh.out.Close()
				delete(f.configuredOutputs, url)
			}
		}

		// Add / reconfigure outputs
		for _, oc := range c.Outputs {
			if oh, ok := f.configuredOutputs[oc.URL]; !ok {
				// New output
				if err := f.setupOutput(&oc); err != nil {
					return err
				}
			} else {
				// Existing output, config changed
				if !reflect.DeepEqual(oh.conf, oc) {
					oh.out.Close()
					delete(f.configuredOutputs, oc.URL)
					if err := f.setupOutput(&oc); err != nil {
						return err
					}
				}
			}
		}
	}

	f.config = *c
	return nil
}
