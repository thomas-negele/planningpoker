package hub

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"time"

	"de.thomasnegele.planningpoker/internal/game"
)

// ThrowObject is one of the deliberately small set of playful table effects.
type ThrowObject string

const (
	ThrowPaperBall  ThrowObject = "paper-ball"
	ThrowPaperPlane ThrowObject = "paper-plane"
	ThrowFlowers    ThrowObject = "flowers"
	ThrowHeart      ThrowObject = "heart"
	// ThrowPoo is accepted only when Limits.PooThrows is set.
	ThrowPoo ThrowObject = "poo"

	// Throw limits are fixed product rules. They sit below the configurable raw
	// connection limit and excess effects are silently discarded.
	ThrowsPerParticipant = 3
	ThrowsPerRoom        = 12
	throwWindow          = time.Second
	throwInboxSize       = 1
	throwDeliverySize    = 1
)

var (
	ErrUnknownThrowObject = errors.New("unknown throw object")
	ErrThrowAtSelf        = errors.New("a participant cannot throw at themselves")
	ErrThrowTargetAbsent  = errors.New("the throw target is not present in this room")
)

// ThrowEvent is immutable transient data. It is never stored in a game snapshot.
type ThrowEvent struct {
	ID         string
	Sender     game.ParticipantID
	Target     game.ParticipantID
	Object     ThrowObject
	Seed       uint32
	AcceptedAt time.Time
}

// ThrowObjects lists the objects a server with these limits accepts, in the
// order the browser offers them.
func (l Limits) ThrowObjects() []ThrowObject {
	objects := []ThrowObject{ThrowPaperBall, ThrowPaperPlane, ThrowFlowers, ThrowHeart}
	if l.PooThrows {
		objects = append(objects, ThrowPoo)
	}
	return objects
}

type throwCommand struct {
	conn   *Conn
	target game.ParticipantID
	object ThrowObject
}

func (c throwCommand) apply(s *roomState) {
	sender, seated := s.conns[c.conn]
	if !seated || sender == "" {
		s.refuse(c.conn, game.ErrUnknownParticipant)
		return
	}
	if !slices.Contains(s.throwObjects, c.object) {
		s.refuse(c.conn, ErrUnknownThrowObject)
		return
	}
	if c.target == sender {
		s.refuse(c.conn, ErrThrowAtSelf)
		return
	}

	present := false
	for _, participant := range s.game.View().Participants {
		if participant.ID == c.target && !participant.Away {
			present = true
			break
		}
	}
	if !present {
		s.refuse(c.conn, ErrThrowTargetAbsent)
		return
	}

	now := s.clock.Now()
	participantHistory := recentThrows(s.participantThrows[sender], now)
	roomHistory := recentThrows(s.roomThrows, now)
	// Check both before spending either allowance.
	if len(participantHistory) >= ThrowsPerParticipant || len(roomHistory) >= ThrowsPerRoom {
		s.participantThrows[sender] = participantHistory
		s.roomThrows = roomHistory
		return
	}

	var seedBytes [4]byte
	if _, err := io.ReadFull(s.random, seedBytes[:]); err != nil {
		s.refuse(c.conn, fmt.Errorf("%w: %v", game.ErrShortRandomRead, err))
		return
	}

	s.throwSequence++
	event := ThrowEvent{
		ID:         strconv.FormatUint(s.throwSequence, 10),
		Sender:     sender,
		Target:     c.target,
		Object:     c.object,
		Seed:       binary.BigEndian.Uint32(seedBytes[:]),
		AcceptedAt: now,
	}
	s.participantThrows[sender] = append(participantHistory, now)
	s.roomThrows = append(roomHistory, now)
	for conn := range s.conns {
		conn.tryThrow(event)
	}
}

// recentThrows returns at most the fixed ceiling's worth of timestamps. A
// backwards-moving test clock cannot accidentally grant extra allowance.
func recentThrows(history []time.Time, now time.Time) []time.Time {
	first := 0
	for first < len(history) {
		age := now.Sub(history[first])
		if age < throwWindow || age < 0 {
			break
		}
		first++
	}
	return history[first:]
}
