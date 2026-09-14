

import type { DeckName } from './decks';

interface CreateGameResponse {
  roomId: string;
}

/**
 * One deck a game can be started with, as the server describes it.
 *
 * The name is typed as a plain string because this is the wire, and the wire is
 * not to be trusted into a narrower type on arrival. Callers narrow it with
 * isDeckName before using it.
 */
export interface StartableDeck {
  /** The identifier createGame expects, once it has been checked. */
  name: string;

  /** The wording to show. */
  label: string;

  /** The deck's cards in the deck's own order. */
  cards: string[];
}

/**
 * Ask which decks a game can be started with. The card values come from the
 * server rather than from a list in this page, so that what the entry screen
 * promises is what the table deals. Callers must be prepared for this to fail:
 * not knowing what is in a deck is a smaller failure than not being able to
 * start a game, so the entry screen carries on without the values.
 */
export async function fetchStartableDecks(): Promise<StartableDeck[]> {
  const response = await fetch('/api/decks');
  if (!response.ok) {
    throw new Error(`the server did not list the decks (status ${response.status})`);
  }
  const body = (await response.json()) as StartableDeck[];
  if (!Array.isArray(body)) {
    throw new Error('the server listed the decks in a form this page does not understand');
  }
  return body;
}

/** Create a room without seating anyone; return its invitation identifier. */
export async function createGame(deck: DeckName = 't-shirt'): Promise<string> {
  const response = await fetch('/api/games', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ deck }),
  });
  if (response.status === 503) {
    // Report capacity separately from other HTTP failures.
    throw new Error(
      'This server is running as many games as it can right now. Please try again in a few minutes.',
    );
  }
  if (!response.ok) {
    throw new Error(`the server refused to start a game (status ${response.status})`);
  }
  const body = (await response.json()) as CreateGameResponse;
  if (!body.roomId) {
    throw new Error('the server started a game but returned no room');
  }
  return body.roomId;
}
