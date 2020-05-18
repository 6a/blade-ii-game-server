// Copyright 2020 James Einosuke Stanton. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE.md file.

// Package matchmaking implements the Blade II Online matchmaking server.
package matchmaking

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/6a/blade-ii-game-server/internal/database"
	"github.com/6a/blade-ii-game-server/internal/protocol"
)

const (

	// BufferSize is the size of each message queue's buffer.
	BufferSize = 2048

	// readyCheckTime is maximum time to wait for a ready check.
	readyCheckTime = time.Second * 20

	// How frequently to update the matchmaking queue (minimum wait between iterations).
	pollTime = 250 * time.Millisecond
)

// Queue is a wrapper for the matchmaking queue
type Queue struct {

	// A slice of indices used to keep track of the order of the clients in the matchmaking queue,
	// as maps are not ordered in golang.
	clientIndex []uint64

	// The current ID number - each client is given a unique ID when they join the server. This
	// value should be incremented per client, so that it can be used to sort them.
	nextClientID uint64

	// Mutex lock for getting the next client ID
	clientIDMutex sync.Mutex

	// A slice of client pairs, containing client pairs that have been matched together, and
	// and are currently performing a ready check (or waiting for one to start).
	matchedPairs []ClientPair

	// A map containing all the clients that are currently matchmaking - essentially the matchmaking queue itself.
	queue map[uint64]*MMClient

	// Channel for new client's that have been successfully authenticated
	connect chan *MMClient

	// Channel for client's that are to be disconnected.
	disconnect chan DisconnectRequest

	// Channel for messages that should be broadcasted to all clients.
	broadcast chan protocol.Message

	// Channel for server commands.
	commands chan protocol.Command
}

// Init initializes the matchmaking server including starting the internal loop.
func (queue *Queue) Init() {

	// Initialize the client index slice. (used to keep track of the order clients in the matchmaking queue, as maps are not ordered in golang).
	queue.clientIndex = make([]uint64, 0)

	// Initialize the matched pairs slice.
	queue.matchedPairs = make([]ClientPair, 0)

	// Initialize the actual queue.
	queue.queue = make(map[uint64]*MMClient)

	// Initialize the various channels.
	queue.connect = make(chan *MMClient, BufferSize)
	queue.disconnect = make(chan DisconnectRequest, BufferSize)
	queue.broadcast = make(chan protocol.Message, BufferSize)
	queue.commands = make(chan protocol.Command, BufferSize)

	go queue.MainLoop()
}

// MainLoop is the main logic loop for the queue.
func (queue *Queue) MainLoop() {

	// Loop forever.
	for {

		// Log the start time for this server tick - so that we can introduce a wait if the tick takes less time than
		// the minimum wait, to reduce server load.
		start := time.Now()

		// Make an empty slice of disconnect requests, so that client disconnects can be handled later.
		toRemove := make([]DisconnectRequest, 0)

		// If any of the queues have something in them, process their data until all the queues are empty.
		for len(queue.connect)+len(queue.disconnect)+len(queue.broadcast)+len(queue.commands) > 0 {
			select {
			case client := <-queue.connect:

				// If a client with the same DBID already exists, we need to set it to be removed, and then
				// update the new clients values to match
				if oldClient, ok := queue.queue[client.DBID]; ok {

					// Disconnect the old client
					queue.Remove(oldClient, protocol.WSCDuplicateConnection, "Removing stale connection")

					// Set the client ID on the new client to match the old one
					client.ClientID = oldClient.ClientID

				} else {

					// Add the key to the client index, which doubles as a record of the join order of the clients.
					queue.clientIndex = append(queue.clientIndex, client.DBID)

					// Set the client ID on the client wit a new ID
					client.ClientID = queue.getNextClientID()
				}

				// Add the client to the queue
				queue.queue[client.DBID] = client

				// Send a message to the client informing it that it has joined the matchmaking queue.
				client.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCJoinedQueue, "Added to matchmaking queue"))

				log.Printf("Client [%s] joined the matchmaking queue. Total clients: %v", client.PublicID, len(queue.queue))

				break
			case disconnectRequest := <-queue.disconnect:

				// Disconnect are handled later, so just add it to the removal queue.
				toRemove = append(toRemove, disconnectRequest)

				break
			case message := <-queue.broadcast:

				// Broadcasted messages are simply broadcasted to all matches in the match map.
				for _, client := range queue.queue {
					client.SendMessage(message)

				}
				break
			case command := <-queue.commands:

				// Process the command.
				queue.processCommand(command)

				break
			}
		}

		// Sort the pending removal slice in ascending order so that we can iterate through the queue index once (by going backwards).
		sort.Slice(toRemove, func(i, j int) bool { return toRemove[i].Client.ClientID < toRemove[j].Client.ClientID })

		// Initialise a variable to use so that we can retain our position when iterating down the queue index.
		indexIterator := len(queue.clientIndex) - 1

		// Remove any clients that are pending removal.
		for index := len(toRemove) - 1; index >= 0; index-- {

			// If a client with the key (queue index) is found in the matchmaking queue...
			if client, ok := queue.queue[toRemove[index].Client.DBID]; ok {

				// Get the public ID for the client.
				deletedClientPID := client.PublicID

				// Close the connection.
				client.Close(protocol.NewMessage(protocol.WSMTText, toRemove[index].Reason, toRemove[index].Message))

				// Check to see if the connection identifier is the same - if it is, then we remove it.
				// If not, it means that this client is actually a stale connection, and it has already
				// been removed from the matchmaking queue.
				if client.connection.UUID == toRemove[index].Client.connection.UUID {
					// Delete the client from the matchmaking queue.
					delete(queue.queue, toRemove[index].Client.DBID)

					// Iterate down the client index slice, backwards, using the iterator that that declared earlier.
					for indexIterator >= 0 {

						// If the current value of the index iterator is the same as the client index of the client
						// that is to be disconnected, remove the index from the queue index slice.
						if queue.clientIndex[index] == toRemove[index].Client.DBID {

							// If the queue only has 1 client, handle this as an edge case because we cant shrink it - instead
							// we just overwrite it with a new empty slice. Otherwise, remove the slice member at the position
							// indicated by the index iterator.
							if len(queue.clientIndex) == 1 {
								queue.clientIndex = make([]uint64, 0)
							} else {
								queue.clientIndex = append(queue.clientIndex[:indexIterator], queue.clientIndex[indexIterator+1:]...)
							}
							break
						}

						// Decrement the index iterator by 1.
						indexIterator--
					}
				}

				log.Printf("Client [%s] left the matchmaking queue. Total clients: %v", deletedClientPID, len(queue.queue))
			}
		}

		// Tick all clients
		for _, client := range queue.queue {
			client.Tick()
		}

		// Get a container containing all the clients that were paired up for a match.
		newMatchedPairs := queue.matchMake()

		// Append all the new matchmade pairs to the matched pairs slice.
		queue.matchedPairs = append(queue.matchedPairs, newMatchedPairs...)

		// Iterate backwards over the matched pairs - backwards so that they can be removed from the slice while
		// iterating.
		for index := len(queue.matchedPairs) - 1; index >= 0; index-- {

			// if the pair is not currently ready checking, it means that it is new to the matched pairs slice. In
			// this case, we start inform the clients that a match was found, and start the ready checking process.
			if !queue.matchedPairs[index].IsReadyChecking {
				queue.matchedPairs[index].SendMatchFoundMessage()
			}

			// Poll the ready check for the matched pair at the current index. If the function returns true,
			// It means that the process has finished, and this pair should be removed from the matched pairs slice.
			if queue.pollReadyCheck(queue.matchedPairs[index]) {

				// If the slice only has 1 client pair, handle this as an edge case because we cant shrink it -
				// instead we just overwrite it with a new empty slice. Otherwise, remove the last member of the
				// matched pairs slice.
				if len(queue.matchedPairs) == 1 {
					queue.matchedPairs = make([]ClientPair, 0)
				} else {
					queue.matchedPairs = queue.matchedPairs[:len(queue.matchedPairs)-1]
				}
			}
		}

		// Add a delay before the next iteration if the time taken is less than the designated poll time.
		elapsed := time.Now().Sub(start)
		remainingPollTime := pollTime - elapsed
		if remainingPollTime > 0 {
			time.Sleep(remainingPollTime)
		}
	}
}

// AddClient takes a client and adds it to the matchmaking server to be processed later.
func (queue *Queue) AddClient(client *MMClient) {
	queue.connect <- client
}

// Remove adds a client to the disconnect queue, to be disconnected next later, along with a reason code and a message.
func (queue *Queue) Remove(client *MMClient, reason protocol.B2Code, message string) {

	// Create a new disconnect request
	disconnectRequest := DisconnectRequest{
		Client:  client,
		Reason:  reason,
		Message: message,
	}

	// Add it to the disconnect queue
	queue.disconnect <- disconnectRequest
}

// Broadcast adds a message to the broadcast queue, to be sent to all connected clients.
func (queue *Queue) Broadcast(message protocol.Message) {
	queue.broadcast <- message
}

// pollReadyCheck checks if the ready check for the specified client pair is complete. If complete, returns true.
// This function also handles the ready checking logic, such as checking for failures, updating the a client that
// the other client has "readied up".
func (queue *Queue) pollReadyCheck(clientPair ClientPair) (finished bool) {

	// Determine if this ready check has finished, by means of timing out.
	timedOut := time.Now().Sub(clientPair.ReadyStart) > readyCheckTime

	// Determine the ready validity for each client. Essentially, a client is ready if they confirmed that they
	// where ready within the ready check maximum time. The ready flag is checked first as it's fast and allows for an early exit.
	client1ReadyValid := clientPair.Client1.Ready && clientPair.Client1.ReadyTime.Sub(clientPair.ReadyStart) <= readyCheckTime
	client2ReadyValid := clientPair.Client2.Ready && clientPair.Client2.ReadyTime.Sub(clientPair.ReadyStart) <= readyCheckTime

	// If the ready check is complete (either both clients are ready and valid, or the ready check ended with one or more
	// clients not confirming that they where ready)...
	if (client1ReadyValid && client2ReadyValid) || timedOut {

		// If the request timed out, and one of the clients was invalid, the match cannot be created.
		if timedOut && (!client1ReadyValid || !client2ReadyValid) {

			// For each client, if they failed the ready check, boot them from the queue. Otherwise, make
			// them elligible for matchmaking again.

			if !client1ReadyValid {

				// Remove the client from the matchmaking queue.
				queue.Remove(clientPair.Client1, protocol.WSCReadyCheckFailed, "")
			} else {

				// Reset their ready checking flags, so that they can be picked up by the matchmaking function again.
				clientPair.Client1.IsReadyChecking = false
				clientPair.Client1.Ready = false

				// Then send a message to the client informing them that their opponent did not accept the match.
				clientPair.Client1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCOpponentDidNotAccept, ""))
			}

			if !client2ReadyValid {

				// Remove the client from the matchmaking queue.
				queue.Remove(clientPair.Client2, protocol.WSCReadyCheckFailed, "")
			} else {

				// Reset their ready checking flags, so that they can be picked up by the matchmaking function again.
				clientPair.Client2.IsReadyChecking = false
				clientPair.Client2.Ready = false

				// Then send a message to the client informing them that their opponent did not accept the match.
				clientPair.Client2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCOpponentDidNotAccept, ""))
			}

			// Return true, indicating that the specified client pair should be removed from the matched pairs
			// slice.
			return true
		}

		// If we reach here, then both clients accepted the match and therefore a match can be created.

		// Create a match, and get the returned match ID. Failures are not not handled properly at the moment.
		matchID, err := database.CreateMatch(clientPair.Client1.DBID, clientPair.Client2.DBID)
		if err != nil {

			// In the event of an error, the match was not created properly, so just boot the players out
			// with a ready check failed code and hope they try again.
			queue.Remove(clientPair.Client1, protocol.WSCReadyCheckFailed, "")
			queue.Remove(clientPair.Client2, protocol.WSCReadyCheckFailed, "")

			log.Printf("Failed to create a match: %s", err.Error())
		}

		// Send the match confirmation message to both clients, with the newly created match's ID.
		clientPair.SendMatchConfirmedMessage(matchID)

		// Remove both clients from the matchmaking queue.
		queue.Remove(clientPair.Client1, protocol.WSCNone, "Match found - closing connection")
		queue.Remove(clientPair.Client2, protocol.WSCNone, "Match found - closing connection")

		// Return true, indicating that the specified client pair should be removed from the matched pairs slice.
		return true
	} else if client1ReadyValid != client2ReadyValid {

		// If the ready check is still incomplete, but not timed out, check to see if one of the clients
		// has become ready since the last time we checked. If this is the case, set a flag (to avoid sending the message
		// multiple times), and inform the non-ready client that the other one is ready.

		if client1ReadyValid && !clientPair.Client1.AcceptMessageSentToOpponent {

			// Set the internal flag to prevent this happening each time this function is called.
			clientPair.Client1.AcceptMessageSentToOpponent = true

			// Send a message to the OTHER client informing them that THIS client is ready.
			clientPair.Client2.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCOpponentAccepted, ""))
		} else if client2ReadyValid && !clientPair.Client2.AcceptMessageSentToOpponent {

			// Set the internal flag to prevent this happening each time this function is called.
			clientPair.Client2.AcceptMessageSentToOpponent = true

			// Send a message to the OTHER client informing them that THIS client is ready.
			clientPair.Client1.SendMessage(protocol.NewMessage(protocol.WSMTText, protocol.WSCOpponentAccepted, ""))
		}
	}

	// Reaching this portion of code indicates that the ready check is still in progress - so return false.
	return false
}

// matchMake goes through the matchmaking queue and pairs up clients based various factors*
//
// Note - Currently just works on a first come first serve basis, but should be changed to take into account ELO, queue
// size, wait time, position in the queue etc.. An empty return array indicates that no clients were paired up.
func (queue *Queue) matchMake() (pairs []ClientPair) {

	// Initialize an empty slice to return.
	pairs = make([]ClientPair, 0)

	// Initialize an empty client pair, which will be replaced after being filled.
	currentPair := ClientPair{}

	// Iterate over all the clients indices in the client index slice.
	for _, clientIndex := range queue.clientIndex {

		// Attempt to get the client - validate the index first. Invalid indices are ignored.
		if client, ok := queue.queue[clientIndex]; ok {

			// Ignore the client if it is currently ready checking as this means it is not eligible for matchmaking.
			if !client.IsReadyChecking {

				// If the client pair declared earlier has a nil value for client 1, set this client as client 1.
				// Otherwise, set it as client 2, append it to the pairs slice, and then reset the client pair
				// back to an empty one.
				if currentPair.Client1 == nil {
					currentPair.Client1 = client
				} else {
					currentPair.Client2 = client
					pairs = append(pairs, currentPair)
					currentPair = ClientPair{}
				}
			}
		}
	}

	// Return the pairs that were found
	return pairs
}

// processCommand handles server commands.
//
// Note - not yet implemented, but prints out some diagonstics and returns with a noop.
func (queue *Queue) processCommand(command protocol.Command) {
	log.Printf("Processing command of type [ %v ] with data [ %v ]", command.Type, command.Data)
}

//
func (queue *Queue) getNextClientID() uint64 {

	queue.clientIDMutex.Lock()
	defer queue.clientIDMutex.Unlock()

	returnID := queue.nextClientID
	queue.nextClientID++

	return returnID
}
