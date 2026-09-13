<script lang="ts">
  import { tick } from 'svelte';
  import type { Card, Participant, ThrowObject } from '../lib/protocol';
  import ThrowIcon from './ThrowIcon.svelte';

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
    canThrow: boolean;
    onthrow: (object: ThrowObject) => void;
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
    canThrow,
    onthrow,
  }: Props = $props();

  // The parent owns the rename dialog and storage preference.
  let open = $state(false);
  let root = $state<HTMLElement>();
  let trigger = $state<HTMLButtonElement>();

  function toggle() {
    if (canThrow) open = !open;
  }

  async function choose(object: ThrowObject) {
    if (!canThrow) return;
    onthrow(object);
    await closeAndRestoreFocus();
  }

  async function closeAndRestoreFocus() {
    open = false;
    await tick();
    trigger?.focus();
  }

  function outside(event: PointerEvent) {
    if (open && event.target instanceof Node && !root?.contains(event.target)) open = false;
  }

  function keydown(event: KeyboardEvent) {
    if (open && event.key === 'Escape') {
      event.preventDefault();
      closeAndRestoreFocus();
    }
  }

  async function focusout() {
    if (!open) return;
    await tick();
    if (root && !root.contains(document.activeElement)) open = false;
  }

  $effect(() => {
    if (!canThrow || participant.away) open = false;
  });
</script>

<svelte:window onpointerdown={outside} onkeydown={keydown} />

<!-- RoomView supplies the seat geometry through CSS custom properties. -->
<li
  bind:this={root}
  class="seat"
  class:you={isYou}
  class:away={participant.away}
  style:--edge-x={edgeX}
  style:--edge-y={edgeY}
  style:--push-x={pushX}
  style:--push-y={pushY}
  style:--push={push}
  data-seat-id={participant.id}
  onfocusout={focusout}
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

  {#if !isYou && !participant.away}
    <div class="throw-control" class:open>
      <button
        bind:this={trigger}
        class="throw-trigger"
        aria-label={`Throw something at ${participant.name}`}
        aria-expanded={open}
        disabled={!canThrow}
        onclick={toggle}
      >
        <ThrowIcon object="paper-ball" size={18} />
      </button>
      <div class="throw-picker" role="group" aria-label={`Throw at ${participant.name}`}>
        <button aria-label={`Throw paper ball at ${participant.name}`} onclick={() => choose('paper-ball')}>
          <ThrowIcon object="paper-ball" size={24} />
        </button>
        <button aria-label={`Throw paper plane at ${participant.name}`} onclick={() => choose('paper-plane')}>
          <ThrowIcon object="paper-plane" size={25} />
        </button>
        <button aria-label={`Throw a flower at ${participant.name}`} onclick={() => choose('flowers')}>
          <ThrowIcon object="flowers" size={25} />
        </button>
      </div>
    </div>
  {/if}
</li>

<style>
  .seat {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.4rem;
    text-align: center;
  }

  .throw-control {
    position: absolute;
    left: calc(100% - 0.2rem);
    bottom: 0.1rem;
    z-index: 7;
  }

  .throw-trigger {
    display: grid;
    place-items: center;
    width: 1.8rem;
    height: 1.8rem;
    padding: 0;
    border-radius: 999px;
    opacity: 0;
    pointer-events: none;
    transform: scale(0.9);
    transition: opacity 120ms ease, transform 120ms ease;
  }

  .throw-trigger:focus-visible {
    opacity: 1;
    transform: scale(1);
  }

  .throw-picker {
    position: absolute;
    left: 50%;
    bottom: calc(100% + 0.35rem);
    display: flex;
    gap: 0.25rem;
    padding: 0.35rem;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: var(--surface-raised);
    box-shadow: 0 0.45rem 1.2rem rgb(0 0 0 / 28%);
    opacity: 0;
    visibility: hidden;
    pointer-events: none;
    transform: translate(-50%, 0.25rem) scale(0.96);
    transition: opacity 100ms ease, transform 100ms ease, visibility 100ms;
  }

  .throw-picker::after {
    content: '';
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    height: 0.5rem;
  }

  .seat:hover .throw-picker,
  .throw-control.open .throw-picker {
    opacity: 1;
    visibility: visible;
    pointer-events: auto;
    transform: translate(-50%, 0) scale(1);
  }

  .throw-picker button {
    display: grid;
    place-items: center;
    width: 2rem;
    height: 2rem;
    padding: 0;
    border: 0;
    border-radius: 999px;
    background: transparent;
  }

  .throw-picker button:hover,
  .throw-picker button:focus-visible {
    background: var(--surface);
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

    .throw-control {
      left: auto;
      right: 2.9rem;
      bottom: 50%;
      transform: translateY(50%);
    }

    .throw-picker {
      left: auto;
      right: 0;
      bottom: calc(100% + 0.35rem);
      transform: translate(0, 0.25rem) scale(0.96);
    }

    .seat:hover .throw-picker,
    .throw-control:focus-within .throw-picker,
    .throw-control.open .throw-picker {
      transform: translate(0, 0) scale(1);
    }
  }

  @media (hover: none) {
    .throw-trigger { opacity: 1; pointer-events: auto; transform: scale(1); }
    .seat:hover .throw-picker {
      opacity: 0;
      visibility: hidden;
      pointer-events: none;
    }
    .throw-control.open .throw-picker { opacity: 1; visibility: visible; pointer-events: auto; }
  }
</style>
