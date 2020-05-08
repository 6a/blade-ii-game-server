// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package protocol provides utilities for handling websocket messages.
package protocol

// Type is a type definition for websocket message types.
type Type uint16

// Types of Websocket Message.
const (
	WSMTContinuation Type = 0
	WSMTText              = 1
	WSMTBinary            = 2
)
