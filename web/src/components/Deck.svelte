<script lang="ts">
  import type { Card, Deck } from '../lib/protocol';

  // Render the server-supplied deck in its original order.

  interface Props {
    deck: Deck;
    /** The card this participant is currently holding, if any. */
    played: Card | null;
    disabled: boolean;
    onplay: (card: Card) => void;
  }

  let { deck, played, disabled, onplay }: Props = $props();
</script>

<footer>
  <ul class:compact={deck.cards.length > 7}>
    {#each deck.cards as card (card)}
      <li>
        <button
          class="card"
          class:played={played === card}
          {disabled}
          aria-pressed={played === card}
          onclick={() => onplay(card)}
        >
          {card}
        </button>
      </li>
    {/each}
  </ul>
</footer>

<style>
  footer {
    position: sticky;
    bottom: 0;
    padding: 1rem 1rem 1.25rem;
    background: linear-gradient(to top, var(--background) 65%, transparent);
  }

  ul {
    margin: 0 auto;
    padding: 0;
    list-style: none;
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.6rem;
  }

  .card {
    width: 3.4rem;
    height: 4.8rem;
    border-radius: 9px;
    border: 1px solid var(--border);
    background: var(--surface-raised);
    color: var(--text);
    font: inherit;
    font-size: 1.2rem;
    font-weight: 600;
    cursor: pointer;
    transition:
      transform 120ms ease,
      border-color 120ms ease;
  }

  .card:hover:not(:disabled),
  .card:focus-visible:not(:disabled) {
    transform: translateY(-6px);
    border-color: var(--accent);
  }

  .card:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  /* The played card stays raised, so what you are holding is visible at a glance. */
  .card.played {
    transform: translateY(-12px);
    border-color: var(--accent);
    background: var(--accent);
    color: var(--accent-text);
  }

  .card:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  ul.compact {
    gap: 0.4rem;
  }

  ul.compact .card {
    width: 2.75rem;
    height: 4rem;
    font-size: 1rem;
  }

  @media (max-width: 30rem) {
    footer {
      padding-inline: 0.5rem;
    }
  }
</style>
