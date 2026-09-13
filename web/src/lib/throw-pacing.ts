import type { ThrowPolicy } from './protocol';

/** Local best-effort guard. The room remains authoritative across connections. */
export class ThrowPacer {
  #policy: ThrowPolicy | null = null;
  #history: number[] = [];
  #tokens = 0;
  #last = 0;
  readonly #now: () => number;

  constructor(now: () => number = () => performance.now()) {
    this.#now = now;
  }

  configure(policy: ThrowPolicy): void {
    this.#policy = policy;
    this.#tokens = 0;
    this.#last = this.#now();
  }

  resetSocket(): void {
    this.#policy = null;
    this.#tokens = 0;
    this.#last = this.#now();
  }

  get ready(): boolean {
    return this.#policy !== null;
  }

  /** Ordinary actions bypass the advisory gate but spend from its budget. */
  noteIntent(): void {
    this.#refill();
    this.#tokens -= 1;
  }

  /** Accept now or discard. No deferred work is retained. */
  takeThrow(outgoingBytes: number): boolean {
    const policy = this.#policy;
    if (policy === null || outgoingBytes > 0) return false;

    const now = this.#now();
    this.#history = this.#history.filter((at) => now - at < 1000 || now - at < 0);
    this.#refill(now);
    if (this.#history.length >= policy.participantPerSecond || this.#tokens < 2) return false;

    this.#history.push(now);
    this.#tokens -= 1;
    return true;
  }

  #refill(now = this.#now()): void {
    const policy = this.#policy;
    if (policy === null) return;
    const elapsed = Math.max(0, now - this.#last) / 1000;
    this.#tokens = Math.min(policy.messageBurst, this.#tokens + elapsed * policy.messagePerSecond * 0.5);
    this.#last = now;
  }
}
