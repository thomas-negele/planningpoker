<script lang="ts">
  import { onMount } from 'svelte';
  import { version } from '../../package.json';
  import { createGame, fetchStartableDecks } from '../lib/api';
  import { DECK_OPTIONS, isDeckName, type DeckName } from '../lib/decks';
  import { navigate, roomPath } from '../lib/router.svelte';

  interface Option {
    name: DeckName;
    label: string;

    /** The deck's cards, once the server has named them. Null until then, and
        null for good if the request fails. */
    cards: string[] | null;
  }

  // Start from the names alone, so the screen is complete before any request
  // finishes. If the server answers, each option gains its cards; if it never
  // does, both decks are still offered and a game can still be started.
  let options = $state<Option[]>(
    DECK_OPTIONS.map((option) => ({ name: option.name, label: option.label, cards: null })),
  );

  let starting = $state(false);
  let problem = $state<string | null>(null);
  let deck = $state<DeckName>('t-shirt');

  onMount(async () => {
    try {
      const listed = await fetchStartableDecks();

      // Take what the server sends, in the order it sends it: it is the only
      // place a deck is defined. A name this page does not know is dropped
      // rather than offered, because page and server ship together and an
      // unknown name means something is wrong, not that something is new.
      const known = listed.filter((entry) => isDeckName(entry.name));
      if (known.length > 0) {
        options = known.map((entry) => ({
          name: entry.name as DeckName,
          label: entry.label,
          cards: entry.cards ?? null,
        }));
        if (!options.some((option) => option.name === deck)) deck = options[0].name;
      }
    } catch {
      // Deliberately silent. The fallback above is already on screen and the
      // visitor can do everything they came to do; an error about card values
      // would be noise in front of a working button.
    }
  });

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

  <button class="primary start" onclick={start} disabled={starting}>
    {starting ? 'Starting…' : 'Start a new game'}
  </button>

  {#if problem}
    <p class="problem" role="alert">{problem}</p>
  {/if}

  <fieldset disabled={starting}>
    <legend>with this deck</legend>
    <ul class="decks">
      {#each options as option (option.name)}
        <li>
          <label class:chosen={deck === option.name}>
            <input type="radio" name="deck" value={option.name} bind:group={deck} />
            <span class="marker" aria-hidden="true"></span>
            <span class="named">
              <span class="label">{option.label}</span>
              {#if option.cards}
                <span class="cards">{option.cards.join(', ')}</span>
              {/if}
            </span>
          </label>
        </li>
      {/each}
    </ul>
  </fieldset>

  <p class="version">Version {version}</p>
</main>

<style>
  main {
    background: var(--surface);
    padding: 3rem;
    border-radius: var(--radius-panel);
    border: 1px solid var(--panel-border);
    box-shadow: var(--shadow-panel);
    text-align: center;
    max-width: 30rem;
  }

  h1 {
    margin: 0 0 0.5rem;
    font-size: 2rem;
    font-weight: var(--display-weight);
    letter-spacing: var(--display-tracking);
  }

  .sub {
    margin: 0 0 2rem;
    color: var(--text-dim);
    line-height: 1.5;
  }

  .start {
    font-size: 1rem;
    padding: 0.7rem 1.4rem;
  }

  fieldset {
    border: 0;
    padding: 0;
    margin: 2rem 0 0;
  }

  legend {
    margin: 0 auto 0.75rem;
    color: var(--text-dim);
    font-size: 0.78rem;
  }

  /* The choice reads as a list rather than as a row of buttons, because each
     option now carries a second line naming its cards. */
  .decks {
    margin: 0;
    padding: 0;
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    text-align: left;
  }

  label {
    display: flex;
    align-items: flex-start;
    gap: 0.7rem;
    padding: 0.7rem 0.9rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    cursor: pointer;
  }

  label.chosen {
    border-color: var(--accent);
    background: var(--surface-raised);
  }

  /* The bullet is the selection: one mark that says both "an option" and
     "this one". */
  .marker {
    flex: none;
    width: 0.85rem;
    height: 0.85rem;
    margin-top: 0.15rem;
    border-radius: 50%;
    border: 1px solid var(--border);
  }

  label.chosen .marker {
    border-color: var(--accent);
    background: var(--accent);
    box-shadow: inset 0 0 0 3px var(--surface-raised);
  }

  .named {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    min-width: 0;
  }

  .label {
    font-size: 0.9rem;
    color: var(--text-dim);
  }

  label.chosen .label {
    color: var(--text);
  }

  /* The cards themselves. Tabular figures keep the Fibonacci deck's numbers
     from shifting about as the selection changes. */
  .cards {
    font-size: 0.78rem;
    line-height: 1.45;
    color: var(--text-dim);
    font-variant-numeric: tabular-nums;
  }

  input {
    position: absolute;
    opacity: 0;
    pointer-events: none;
  }

  /* The real radio keeps arrow-key selection; this is what shows its focus. */
  input:focus-visible + .marker {
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
