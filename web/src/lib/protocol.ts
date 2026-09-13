// Wire types and refusal codes mirror internal/transport/protocol.go.

import type { DeckName } from './decks';

/** A card is whatever the room's deck offers. The page never assumes which. */
export type Card = string;

export type ThrowObject = 'paper-ball' | 'paper-plane' | 'flowers';

export interface ThrowPolicy {
  participantPerSecond: number;
  roomPerSecond: number;
  messagePerSecond: number;
  messageBurst: number;
}

export interface Deck {
  name: DeckName;
  cards: Card[];
  /**
   * Ordered subset of cards that express a size. Non-scale cards are derived by
   * difference.
   */
  scale: Card[];
}

/** Public participant state. Hidden snapshots disclose only whether a vote exists. */
export interface Participant {
  id: string;
  name: string;
  away: boolean;
  voted: boolean;
}

export interface ParticipantCard {
  id: string;
  /** Empty for somebody who did not vote, which is not the same as playing "?". */
  card: Card;
}

export interface CardCount {
  card: Card;
  count: number;
}

/** Present only once the round is revealed; absent entirely while it is hidden. */
export interface Results {
  cards: ParticipantCard[];
  tally: CardCount[];
}

export interface Room {
  id: string;
  deck: Deck;
  /** Present only after a revealed round selected a deck for the next round. */
  pendingDeck?: Deck;
  revealed: boolean;
  participants: Participant[];
  everyonePresentHasVoted: boolean;
  results?: Results;
}

export interface StateMessage {
  type: 'state';
  /** This connection's own participant, empty until it has taken a seat. */
  you: string;
  room: Room;
  throwPolicy?: ThrowPolicy;
}

export interface ThrownMessage {
  type: 'thrown';
  id: string;
  sender: string;
  target: string;
  object: ThrowObject;
  seed: number;
  ageMs: number;
}

export interface ErrorMessage {
  type: 'error';
  code: string;
  message: string;
}

export type ServerMessage = StateMessage | ErrorMessage | ThrownMessage;

/** Everything a client may ask for. */
export type ClientMessage =
  | { type: 'seat'; name: string }
  | { type: 'vote'; card: Card }
  | { type: 'reveal' }
  | { type: 'newRound' }
  | { type: 'setDeck'; deck: DeckName }
  | { type: 'rename'; name: string }
  | { type: 'throw'; target: string; object: ThrowObject };

/** Application close code for an invalid room ID; browsers can read it after upgrade. */
export const CLOSE_INVALID_ROOM_ID = 4400;

/** The refusal code meaning the identifier may not name a room. */
export const CODE_INVALID_ROOM_ID = 'invalid_room_id';

/** Process room capacity reached; automatic retries stop. */
export const CODE_AT_CAPACITY = 'at_capacity';

/** Per-room connection capacity reached, distinct from room and seat ceilings. */
export const CODE_TOO_MANY_CONNECTIONS = 'too_many_connections';

/** User-facing messages indexed by server refusal code. */
const REFUSALS: Record<string, string> = {
  not_seated: 'You are not seated at this table. Reload the page to take a seat.',
  name_empty: 'Please enter a name.',
  name_too_long: 'That name is too long. Please use a shorter one.',
  card_not_in_deck: 'That card is not in this deck.',
  unknown_deck: 'That deck is not available.',
  deck_locked: 'The deck cannot be changed while voting is in progress.',
  round_revealed: 'The round has been revealed. Start a new round to vote again.',
  bad_message: 'The server did not understand that. Please reload the page.',
  invalid_room_id:
    'A room name needs at least 5 characters, and may contain only letters, digits, hyphens and underscores.',
  room_full: 'This table is full. Ask somebody to close their tab, then try again.',
  at_capacity:
    'This server is running as many games as it can right now. Please try again in a few minutes.',
  too_fast: 'That arrived faster than the server accepts. Please slow down.',
  unknown_throw: 'That throw is not available.',
  throw_at_self: 'Choose another participant for that throw.',
  throw_target_absent: 'That participant is not currently at the table.',
  too_many_connections:
    'This room already has as many connections open as it allows. If you have it open in another tab, close that one and try again.',
  server_error: 'Something went wrong on the server. Please try again.',
};

export function refusalText(code: string): string {
  const known = REFUSALS[code];
  if (known !== undefined) return known;

  // Report unmapped codes so protocol drift is visible.
  console.error(`[planningpoker] no message for refusal code "${code}"`);
  return `Unexpected refusal from the server (${code}).`;
}

/** Known refusal codes exposed for protocol consistency checks. */
export const KNOWN_REFUSAL_CODES = Object.keys(REFUSALS);
