package net

// WSMessageType is a type definition for websocket message types
type WSMessageType int

// Types of Websocket Message
const (
	WSMTContinuation = 0
	WSMTText         = 1
	WSMTBinary       = 2
)
