<script lang="ts">
  import { version } from '../../package.json';
  import { createGame } from '../lib/api';
  import { DECK_OPTIONS, type DeckName } from '../lib/decks';
  import { navigate, roomPath } from '../lib/router.svelte';

  let starting = $state(false);
  let problem = $state<string | null>(null);
  let deck = $state<DeckName>('t-shirt');

  async function start() {
    starting = true;
    problem = null;
    try {
      const roomId = await createGame(deck);
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

  <fieldset disabled={starting}>
    <legend>Deck</legend>
    <div class="deck-options">
      {#each DECK_OPTIONS as option (option.name)}
        <label class:chosen={deck === option.name}>
          <input type="radio" name="deck" value={option.name} bind:group={deck} />
          <span>{option.label}</span>
        </label>
      {/each}
    </div>
  </fieldset>

  <button class="primary" onclick={start} disabled={starting}>
    {starting ? 'Starting…' : 'Start a new game'}
  </button>

  {#if problem}
    <p class="problem" role="alert">{problem}</p>
  {/if}

  <p class="version">Version {version}</p>
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

  fieldset {
    border: 0;
    padding: 0;
    margin: 0 0 1.25rem;
  }

  legend {
    margin: 0 auto 0.55rem;
    color: var(--text-dim);
    font-size: 0.78rem;
  }

  .deck-options {
    display: flex;
    justify-content: center;
    gap: 0.5rem;
  }

  label {
    position: relative;
    display: inline-flex;
    cursor: pointer;
  }

  label span {
    padding: 0.55rem 0.8rem;
    border: 1px solid var(--border);
    border-radius: 8px;
    color: var(--text-dim);
    font-size: 0.85rem;
  }

  label.chosen span {
    border-color: var(--accent);
    color: var(--text);
    background: var(--surface-raised);
  }

  input {
    position: absolute;
    opacity: 0;
    pointer-events: none;
  }

  input:focus-visible + span {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  fieldset:disabled label {
    cursor: not-allowed;
    opacity: 0.45;
  }

  .problem {
    margin: 1.25rem 0 0;
    color: var(--bad);
    font-size: 0.9rem;
  }

  .version {
    margin: 2rem 0 0;
    color: var(--text-dim);
    font-size: 0.75rem;
  }
</style>
