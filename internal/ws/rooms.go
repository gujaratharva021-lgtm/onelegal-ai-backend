package ws

import (
	"sync"

	"github.com/google/uuid"
)

// RoomRegistry restricts signaling relay for a given call_id to exactly the
// two authorized participants of that private video room (lawyer + client).
// Even though Hub.SendJSON already targets a single specific recipient, that
// alone doesn't stop a third authenticated user from crafting a message with
// a call_id they observed and inserting themselves into someone else's call
// — the registry closes that gap by checking both ends of every relayed
// call message against the room's actual two-person membership.
type RoomRegistry struct {
	mu    sync.RWMutex
	rooms map[string][2]uuid.UUID // roomID -> [lawyerID, clientUserID]
}

func NewRoomRegistry() *RoomRegistry {
	return &RoomRegistry{rooms: make(map[string][2]uuid.UUID)}
}

// Register marks a room active with exactly its two allowed participants.
func (r *RoomRegistry) Register(roomID string, lawyerID, clientUserID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rooms[roomID] = [2]uuid.UUID{lawyerID, clientUserID}
}

// Unregister closes a room — after this, no further signaling for that
// call_id is relayed by anyone.
func (r *RoomRegistry) Unregister(roomID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.rooms, roomID)
}

// Allowed reports whether userID and otherID are the two (and only the two)
// registered participants of roomID.
func (r *RoomRegistry) Allowed(roomID string, userID, otherID uuid.UUID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	pair, ok := r.rooms[roomID]
	if !ok {
		return false
	}
	return (userID == pair[0] && otherID == pair[1]) || (userID == pair[1] && otherID == pair[0])
}
