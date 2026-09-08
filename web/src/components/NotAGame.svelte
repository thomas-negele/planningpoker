<script lang="ts">
  import { createGame } from '../lib/api';
  import { navigate, roomPath } from '../lib/router.svelte';

  // Shown for invalid room IDs. Missing rooms are recreated on connection.

  interface Props {
    /** The server's reason, already a sentence. */
    reason: string | null;
  }

  let { reason }: Props = $props();

  let starting = $state(false);
  let problem = $state<string | null>(null);

  async function start() {
    starting = true;
    problem = null;
    try {
      navigate(roomPath(await createGame()));
    } catch (error) {
      problem = error instanceof Error ? error.message : 'Could not start a game.';
      starting = false;
    }
  }
</script>

<main>
  <h1>That link is not a game</h1>

  <p class="rule">
    {reason ??
      'A room name needs at least 5 characters, and may contain only letters, digits, hyphens and underscores.'}
  </p>

  <button class="primary" onclick={start} disabled={starting}>
    {starting ? 'Starting…' : 'Start a new game'}
  </button>

  {#if problem}
    <p class="problem" role="alert">{problem}</p>
  {/if}
</main>

<style>
  main {
    background: var(--surface);
    padding: 2.5rem 3rem;
    border-radius: 16px;
    text-align: center;
    max-width: 30rem;
  }

  h1 {
    margin: 0 0 0.75rem;
    font-size: 1.5rem;
  }

  .rule {
    margin: 0 0 1.75rem;
    color: var(--text-dim);
    line-height: 1.55;
    font-size: 0.9rem;
  }

  .problem {
    margin: 1rem 0 0;
    color: var(--bad);
    font-size: 0.9rem;
  }
</style>
