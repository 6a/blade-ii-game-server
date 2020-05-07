// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package apiinterface provides utilities for interacting with the Blade II Online REST API.
package apiinterface

// Winner represents the winner for a match.
type Winner uint8

// Enum values that represent the potential winners for a match.
const (
	Draw    Winner = 0
	Player1        = 1
	Player2        = 2
)

// MMRUpdateRequest describes the data needed to update the MMR for a pair of users.
type MMRUpdateRequest struct {
	Player1ID uint64 `json:"player1id"`
	Player2ID uint64 `json:"player2id"`
	Winner    Winner `json:"winner"`
}
