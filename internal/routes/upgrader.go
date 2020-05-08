// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package routes defines http endpoint handlers for http/websocket connections to the server.
package routes

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// upgrader is a custom websocket upgrader with a origin checking function that always returns true. This
// is definitely not safe, but it works.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // TODO replace with a proper method
}
