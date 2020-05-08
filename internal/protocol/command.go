// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package protocol provides utilities for handling websocket messages.
package protocol

// Queue Command types.
const (
	QCTBroadcastMessage uint16 = iota
	QCTDropAll
	QCTChangePollTime
)

// Command is a wrapper for a queue command and any accompanying data.
type Command struct {
	Type uint16
	Data string
}
