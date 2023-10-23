package chandy_lamport

import (
	"log"
)

type Server struct {
	Id            string
	Tokens        int
	sim           *Simulator
	outboundLinks map[string]*Link // key = link.dest
	inboundLinks  map[string]*Link // key = link.src
	recording     map[string]bool  // key = link.src, value = if recording for that link is on
	snapshot      *SnapshotState   // The state of the current snapshot
}

func NewServer(id string, tokens int, sim *Simulator) *Server {
	return &Server{
		Id:            id,
		Tokens:        tokens,
		sim:           sim,
		outboundLinks: make(map[string]*Link),
		inboundLinks:  make(map[string]*Link),
		recording:     make(map[string]bool),
		snapshot:      nil,
	}
}

// Existing methods remain the same...

// Start the chandy-lamport snapshot algorithm on this server.
// This should be called only once per server.
func (server *Server) StartSnapshot(snapshotId int) {
	server.snapshot = &SnapshotState{
		Id:      snapshotId,
		Tokens:  server.Tokens,
		Messages: make([]*SnapshotMessage, 0),
	}
	// Mark that recording is turned on for all outbound channels
	for _, link := range server.outboundLinks {
		server.recording[link.dest] = true
	}

	// Send marker messages to all outbound links
	for _, dest := range getSortedKeys(server.outboundLinks) {
		markerMessage := MarkerMessage{SnapshotId: snapshotId}
		server.SendToNeighbors(markerMessage)
	}
}

// Callback for when a message is received on this server.
func (server *Server) HandlePacket(src string, message interface{}) {
	switch msg := message.(type) {
	case TokenMessage:
		// Handling TokenMessage
		server.Tokens += msg.Tokens
		if server.recording[src] {
			server.snapshot.Messages = append(server.snapshot.Messages,
				&SnapshotMessage{From: src, Tokens: msg.Tokens})
		}
	case MarkerMessage:
		// Handling MarkerMessage
		if server.snapshot == nil {
			// First marker message for this snapshot
			server.StartSnapshot(msg.SnapshotId)
			server.sim.NotifySnapshotComplete(server.Id, msg.SnapshotId)
		} else {
			if server.recording[src] {
				// Stop recording on the incoming channel
				server.recording[src] = false
			}
		}
	}

	// Check if the snapshot is complete for this server
	isComplete := true
	for _, status := range server.recording {
		if status {
			isComplete = false
			break
		}
	}
	if isComplete {
		// Notify the simulator
		server.sim.NotifySnapshotComplete(server.Id, server.snapshot.Id)
	}
}
