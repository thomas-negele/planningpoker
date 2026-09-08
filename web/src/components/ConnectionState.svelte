<script lang="ts">
  import type { ConnectionStatus } from '../lib/connection.svelte';

  // Expose connection status through text as well as color.

  interface Props {
    status: ConnectionStatus;
  }

  let { status }: Props = $props();

  const WORDS: Record<ConnectionStatus, string> = {
    connecting: 'Connecting…',
    open: 'Connected',
    reconnecting: 'Connection lost — reconnecting…',
    refused: 'That link is not a game',
    full: 'This server is full',
  };
</script>

<p class="state" data-status={status} role="status" aria-live="polite">
  <span class="dot" aria-hidden="true"></span>
  {WORDS[status]}
</p>

<style>
  .state {
    margin: 0;
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    font-size: 0.8rem;
    color: var(--text-dim);
  }

  .dot {
    width: 0.5rem;
    height: 0.5rem;
    border-radius: 50%;
    background: var(--text-dim);
    flex: none;
  }

  /* Colour is an addition to the words, never the only carrier of the meaning. */
  .state[data-status='open'] .dot {
    background: var(--good);
  }

  .state[data-status='connecting'] .dot,
  .state[data-status='reconnecting'] .dot {
    background: var(--warn);
  }

  .state[data-status='reconnecting'] {
    color: var(--warn);
  }

  .state[data-status='refused'] .dot,
  .state[data-status='full'] .dot {
    background: var(--bad);
  }
</style>
