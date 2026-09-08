// Owns the room socket, reconnection state and received snapshots. Components send
// intents through this class; the server remains authoritative.

import {
  CLOSE_INVALID_ROOM_ID,
  CODE_AT_CAPACITY,
  CODE_INVALID_ROOM_ID,
  CODE_TOO_MANY_CONNECTIONS,
  refusalText,
  type ClientMessage,
  type Room,
  type ServerMessage,
} from './protocol';

export type ConnectionStatus =
  /** A socket is being opened for the first time. */
  | 'connecting'
  /** The socket is open; the latest received snapshot is rendered. */
  | 'open'
  /** Lost, and being re-established on its own. What is on screen is stale. */
  | 'reconnecting'
  /** The server rejected the room ID. Automatic retries stop. */
  | 'refused'
  /** The server or room is at capacity. Automatic retries stop until manual retry. */
  | 'full';

// Use increasing delays to limit retry traffic and jitter to spread reconnects.
// The base delay is capped before jitter, which can extend it by 25%.
const FIRST_DELAY_MS = 500;
const GROWTH_FACTOR = 1.8;
const MAX_DELAY_MS = 20_000;
const JITTER = 0.25;

function nextDelay(previous: number): number {
  const grown = Math.min(previous * GROWTH_FACTOR, MAX_DELAY_MS);
  const spread = grown * JITTER;
  return grown - spread + Math.random() * spread * 2;
}

export class RoomConnection {
  /** Connection status used to mark stale UI and disable actions. */
  status = $state<ConnectionStatus>('connecting');

  /** The last snapshot the server sent. Null until the first one arrives. */
  room = $state<Room | null>(null);

  /** This connection's own participant identifier, empty until it has a seat. */
  you = $state('');

  /** The most recent refusal, already translated into a sentence. */
  refusal = $state<string | null>(null);

  /** Distinguishes a full server from a room with no free connections. */
  fullScope = $state<'server' | 'room' | null>(null);

  /**
   * The previous participant no longer exists in a received snapshot, indicating
   * that the room was recreated and its round was lost.
   */
  roundLost = $state(false);

  /**
   * Local record of the card sent by this page; hidden snapshots contain no value.
   * Retained across socket reconnects, but unavailable after a page reload.
   * This is not an acknowledgement of a specific vote.
   */
  myCard = $state<string | null>(null);

  readonly roomId: string;

  /** Most recently submitted card, used when a snapshot reports an existing vote. */
  #pendingCard: string | null = null;

  #socket: WebSocket | null = null;
  #delay = FIRST_DELAY_MS;
  #retry: ReturnType<typeof setTimeout> | null = null;
  #closed = false;

  constructor(roomId: string) {
    this.roomId = roomId;
    this.#open();
  }

  /** True once this connection has taken a seat at the table. */
  get seated(): boolean {
    return this.you !== '';
  }

  /** Intents may only be sent over an established connection. */
  get canAct(): boolean {
    return this.status === 'open' && this.#socket?.readyState === WebSocket.OPEN;
  }

  seat(name: string): void {
    this.#send({ type: 'seat', name });
  }

  vote(card: string): void {
    if (!this.canAct) return;
    this.#pendingCard = card;
    this.#send({ type: 'vote', card });
  }

  reveal(): void {
    this.#send({ type: 'reveal' });
  }

  newRound(): void {
    this.#send({ type: 'newRound' });
  }

  rename(name: string): void {
    this.#send({ type: 'rename', name });
  }

  /** Clears the refusal currently on screen, once it has been read. */
  dismissRefusal(): void {
    this.refusal = null;
  }

  /** Dismisses the notice that the round was lost. */
  dismissRoundLost(): void {
    this.roundLost = false;
  }

  /** Stops for good. Called when the page leaves this room. */
  close(): void {
    this.#closed = true;
    if (this.#retry !== null) clearTimeout(this.#retry);
    this.#socket?.close(1000, 'left the room');
    this.#socket = null;
  }

  #send(message: ClientMessage): void {
    if (!this.canAct) return;
    this.#socket?.send(JSON.stringify(message));
  }

  #url(): string {
    const url = new URL(`/ws/${encodeURIComponent(this.roomId)}`, window.location.href);
    url.protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return url.toString();
  }

  #open(): void {
    if (this.#closed) return;

    let socket: WebSocket;
    try {
      socket = new WebSocket(this.#url());
    } catch {
      this.#scheduleRetry();
      return;
    }
    this.#socket = socket;

    socket.addEventListener('open', () => {
      // The first fresh snapshot may arrive after the open event.
      this.status = 'open';
      this.#delay = FIRST_DELAY_MS;
    });

    socket.addEventListener('message', (event) => {
      let message: ServerMessage;
      try {
        message = JSON.parse(String(event.data)) as ServerMessage;
      } catch {
        console.error('[planningpoker] could not read a message from the server');
        return;
      }
      this.#receive(message);
    });

    socket.addEventListener('close', (event) => {
      this.#socket = null;
      if (this.#closed) return;

      if (this.status === 'full') {
        // Capacity refusal stops automatic retries.
        return;
      }

      if (event.code === CLOSE_INVALID_ROOM_ID) {
        // Invalid names cannot become valid through retrying.
        this.status = 'refused';
        return;
      }

      // Retry transient failures; the next snapshot reveals whether the room survived.
      this.#scheduleRetry();
    });

    socket.addEventListener('error', () => {
      // The close event handles retry decisions; browser error events omit the cause.
    });
  }

  #receive(message: ServerMessage): void {
    if (message.type === 'state') {
      // A recreated room no longer contains this page's previous participant.
      if (this.you !== '' && !message.room.participants.some((p) => p.id === this.you)) {
        this.roundLost = true;
        this.you = '';
      }

      this.room = message.room;
      if (message.you !== '') this.you = message.you;
      this.#reconcileMyCard();
      return;
    }

    if (message.type === 'error') {
      if (message.code === CODE_INVALID_ROOM_ID) {
        this.status = 'refused';
        this.refusal = refusalText(message.code);
        return;
      }

      if (message.code === CODE_AT_CAPACITY || message.code === CODE_TOO_MANY_CONNECTIONS) {
        // Both capacity refusals stop retries, with different advice for each scope.
        this.status = 'full';
        this.fullScope = message.code === CODE_AT_CAPACITY ? 'server' : 'room';
        this.refusal = refusalText(message.code);
        return;
      }
      this.refusal = refusalText(message.code);
    }
  }

  /**
   * Clear the local card when no vote exists; otherwise use the pending value.
   * A voted flag does not acknowledge a specific card. Error handling currently
   * leaves pending votes intact, so a rejected change may be displayed later.
   */
  #reconcileMyCard(): void {
    const me = this.room?.participants.find((p) => p.id === this.you);

    if (!me?.voted) {
      this.#pendingCard = null;
      this.myCard = null;
      return;
    }

    if (this.#pendingCard !== null) {
      this.myCard = this.#pendingCard;
    }
    // Without a local submission, a hidden card remains unknown to this page.
  }

  #scheduleRetry(): void {
    if (this.#closed || this.status === 'refused') return;

    this.status = 'reconnecting';
    this.#retry = setTimeout(() => {
      this.#retry = null;
      this.#open();
    }, this.#delay);
    this.#delay = nextDelay(this.#delay);
  }
}
