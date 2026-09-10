<script lang="ts">
  import { onMount } from 'svelte';
  import { DECK_OPTIONS, deckLabel, type DeckName } from '../lib/decks';
  import type { Deck } from '../lib/protocol';

  interface Props {
    active: Deck;
    pending?: Deck;
    revealed: boolean;
    disabled: boolean;
    onselect: (deck: DeckName) => void;
    oncancel: () => void;
  }

  let { active, pending, revealed, disabled, onselect, oncancel }: Props = $props();
  const selected = $derived(pending?.name ?? active.name);
  let dialog: HTMLDivElement;

  onMount(() => dialog.querySelector<HTMLInputElement>('input:checked')?.focus());
</script>

<div
  class="backdrop"
  role="presentation"
  onclick={(event) => event.target === event.currentTarget && oncancel()}
>
  <div
    bind:this={dialog}
    class="dialog"
    role="dialog"
    tabindex="-1"
    aria-modal="true"
    aria-labelledby="settings-title"
    onkeydown={(event) => event.key === 'Escape' && oncancel()}
  >
    <h2 id="settings-title">Room settings</h2>
    <p class="current">Current deck: {deckLabel(active.name)}</p>
    {#if revealed}
      <p class="hint">
        {#if pending}
          {deckLabel(pending.name)} is selected for the next round.
        {:else}
          Choose a deck for the next round. The current results stay unchanged.
        {/if}
      </p>
    {/if}

    <fieldset {disabled}>
      <legend>{revealed ? 'Deck for next round' : 'Deck'}</legend>
      {#each DECK_OPTIONS as option (option.name)}
        <label>
          <input
            type="radio"
            name="room-deck"
            value={option.name}
            checked={selected === option.name}
            onchange={() => onselect(option.name)}
          />
          <span>
            <strong>{option.label}</strong>
            <small>{option.name === 't-shirt' ? 'XS · S · M · L · XL · ? · ☕' : '0 · ½ · 1 · 2 · 3 · 5 · 8 · 13 · 21 · ? · ☕'}</small>
          </span>
        </label>
      {/each}
    </fieldset>

    <div class="actions">
      <button type="button" onclick={oncancel}>Close</button>
    </div>
  </div>
</div>

<style>
  .backdrop {
    position: fixed;
    inset: 0;
    background: rgb(0 0 0 / 55%);
    display: grid;
    place-items: center;
    padding: 1rem;
    z-index: 50;
  }

  .dialog {
    width: min(25rem, 100%);
    border-radius: 16px;
    background: var(--surface);
    padding: 1.5rem;
  }

  h2 {
    margin: 0;
    font-size: 1.15rem;
  }

  .current,
  .hint {
    margin: 0.45rem 0 0;
    color: var(--text-dim);
    font-size: 0.8rem;
    line-height: 1.45;
  }

  fieldset {
    display: grid;
    gap: 0.55rem;
    border: 0;
    padding: 0;
    margin: 1.2rem 0 0;
  }

  legend {
    margin-bottom: 0.55rem;
    font-size: 0.8rem;
    color: var(--text-dim);
  }

  label {
    position: relative;
    display: block;
    cursor: pointer;
  }

  input {
    position: absolute;
    opacity: 0;
    pointer-events: none;
  }

  label span {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.75rem 0.85rem;
    border: 1px solid var(--border);
    border-radius: 9px;
    color: var(--text-dim);
  }

  input:checked + span {
    border-color: var(--accent);
    background: var(--surface-raised);
    color: var(--text);
  }

  input:focus-visible + span {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  small {
    font-size: 0.72rem;
    font-weight: 400;
    line-height: 1.4;
  }

  fieldset:disabled label {
    cursor: not-allowed;
    opacity: 0.45;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    margin-top: 1rem;
  }
</style>
