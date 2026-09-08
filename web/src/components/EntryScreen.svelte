<script lang="ts">
  import { createGame } from '../lib/api';
  import { navigate, roomPath } from '../lib/router.svelte';

  let starting = $state(false);
  let problem = $state<string | null>(null);

  async function start() {
    starting = true;
    problem = null;
    try {
      const roomId = await createGame();
      navigate(roomPath(roomId));
    } catch (error) {
      problem = error instanceof Error ? error.message : 'Could not start a game.';
      starting = false;
    }
  }
</script>

<main>
  <h1>Planning Poker</h1>
  <p class="sub">Start a game, share the link, estimate together.</p>

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
    padding: 3rem;
    border-radius: 16px;
    text-align: center;
    max-width: 30rem;
  }

  h1 {
    margin: 0 0 0.5rem;
    font-size: 2rem;
    letter-spacing: -0.02em;
  }

  .sub {
    margin: 0 0 2rem;
    color: var(--text-dim);
    line-height: 1.5;
  }

  .problem {
    margin: 1.25rem 0 0;
    color: var(--bad);
    font-size: 0.9rem;
  }
</style>
