<script lang="ts">
  import type { Card, Participant } from '../lib/protocol';

  // Render one participant. Card values are supplied only after reveal.

  interface Props {
    participant: Participant;
    /** The card this participant played, once the round is revealed. */
    card: Card | null;
    revealed: boolean;
    isYou: boolean;
    /** Ellipse edge coordinates and outward displacement supplied by RoomView. */
    edgeX: number;
    edgeY: number;
    pushX: number;
    pushY: number;
    push: string;
    canRename: boolean;
    onedit: () => void;
  }

  let {
    participant,
    card,
    revealed,
    isYou,
    edgeX,
    edgeY,
    pushX,
    pushY,
    push,
    canRename,
    onedit,
  }: Props = $props();

  // The parent owns the rename dialog and storage preference.
</script>

<!-- RoomView supplies the seat geometry through CSS custom properties. -->
<li
  class="seat"
  class:you={isYou}
  class:away={participant.away}
  style:--edge-x={edgeX}
  style:--edge-y={edgeY}
  style:--push-x={pushX}
  style:--push-y={pushY}
  style:--push={push}
>
  <div class="card-slot">
    {#if revealed}
      {#if card}
        <span class="card face-up">{card}</span>
      {:else}
        <span class="card empty" aria-label="did not vote">—</span>
      {/if}
    {:else if participant.voted}
      <span class="card face-down" aria-label="has played a card"></span>
    {:else}
      <span class="card empty" aria-label="has not played a card"></span>
    {/if}
  </div>

  <div class="who">
    {#if isYou && canRename}
      <button class="name editable" onclick={onedit} title="Change your name and how it is stored">
        {participant.name}
        <svg viewBox="0 0 16 16" width="11" height="11" aria-hidden="true" focusable="false">
          <path
            d="M11.5 1.5l3 3L5 14H2v-3l9.5-9.5z"
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linejoin="round"
          />
        </svg>
      </button>
    {:else}
      <span class="name">{participant.name}</span>
    {/if}

    {#if participant.away}
      <span class="away-tag">away</span>
    {/if}
  </div>
</li>

<style>
  .seat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.4rem;
    text-align: center;
  }

  .card-slot {
    height: 4.2rem;
    display: flex;
    align-items: flex-end;
  }

  .card {
    width: 2.9rem;
    height: 4.1rem;
    border-radius: 7px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.05rem;
    font-weight: 600;
    border: 1px solid var(--border);
    transition: transform 120ms ease;
  }

  .card.face-down {
    background: repeating-linear-gradient(
      45deg,
      var(--card-back) 0 6px,
      var(--card-back-alt) 6px 12px
    );
    border-color: var(--accent-dim);
  }

  .card.face-up {
    background: var(--card-face);
    color: var(--card-face-text);
    border-color: var(--card-face-border);
  }

  .card.empty {
    background: transparent;
    border-style: dashed;
    color: var(--text-dim);
    font-weight: 400;
  }

  /* Keep the full accepted name visible; do not truncate it. */
  .who {
    display: flex;
    align-items: center;
    gap: 0.35rem;
  }

  .name {
    font-size: 0.9rem;
    white-space: nowrap;
  }

  button.name {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    background: none;
    border: none;
    padding: 0.1rem 0.3rem;
    border-radius: 5px;
    color: inherit;
    font: inherit;
    font-size: 0.9rem;
    cursor: pointer;
  }

  button.name:hover,
  button.name:focus-visible {
    background: var(--surface-raised);
  }

  .seat.you .name {
    color: var(--accent);
    font-weight: 600;
  }

  .seat.away {
    opacity: 0.55;
  }

  .away-tag {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-dim);
    border: 1px solid var(--border);
    border-radius: 999px;
    padding: 0 0.35rem;
  }

  /* Use a horizontal seat row in the narrow-screen list. */
  @media (max-width: 57.999rem) {
    .seat {
      flex-direction: row;
      align-items: center;
      justify-content: space-between;
      width: 100%;
      gap: 0.75rem;
      padding: 0.3rem 0.6rem;
      border-radius: 8px;
      background: var(--surface);
      text-align: left;
    }

    .seat.you {
      background: var(--surface-raised);
    }

    .card-slot {
      order: 2;
      height: auto;
      align-items: center;
      flex: none;
    }

    .card {
      width: 1.8rem;
      height: 2.5rem;
      border-radius: 5px;
      font-size: 0.8rem;
    }

    .who {
      order: 1;
      flex: 1;
      min-width: 0;
    }

    .name {
      font-size: 0.95rem;
    }
  }
</style>
