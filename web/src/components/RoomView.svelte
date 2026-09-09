<script lang="ts">
  import { onDestroy } from 'svelte';
  import { RoomConnection } from '../lib/connection.svelte';
  import { forgetName, nameIsRemembered, rememberName } from '../lib/name';
  import type { Card } from '../lib/protocol';
  import ConnectionState from './ConnectionState.svelte';
  import Deck from './Deck.svelte';
  import NotAGame from './NotAGame.svelte';
  import ServerFull from './ServerFull.svelte';
  import InviteControl from './InviteControl.svelte';
  import NamePrompt from './NamePrompt.svelte';
  import Results from './Results.svelte';
  import NameDialog from './NameDialog.svelte';
  import Seat from './Seat.svelte';

  interface Props {
    roomId: string;
  }

  let { roomId }: Props = $props();

  // App keys this component by roomId, so navigation creates a new connection.
  // svelte-ignore state_referenced_locally
  const connection = new RoomConnection(roomId);
  onDestroy(() => connection.close());

  const room = $derived(connection.room);
  const participants = $derived(room?.participants ?? []);
  const seated = $derived(connection.seated);

  // Use revealed server results when available; otherwise show this page's local card.
  const myCard = $derived<Card | null>(
    room?.revealed
      ? room?.results?.cards.find((entry) => entry.id === connection.you)?.card || null
      : connection.myCard,
  );

  // Seat half-extents in rem, including the name. These are fixed layout budgets,
  // not live measurements. Keep them aligned with the name limit, side padding and
  // 58rem breakpoint; width was sized using fifteen capital Ms.
  const SEAT_HALF_WIDTH = 5.9;
  const SEAT_HALF_HEIGHT = 2.95;
  const SEAT_GAP = 1.1;

  /**
   * Place seats on an ellipse, offset along its outward normal by the seat's
   * projected half-extent and the gap to the table.
   */
  function seatPosition(index: number, total: number) {
    const angle = total === 0 ? -Math.PI / 2 : (index / total) * 2 * Math.PI - Math.PI / 2;
    const cos = Math.cos(angle);
    const sin = Math.sin(angle);

    // The outward normal of an ellipse whose width:height is 16:9. Multiplying by the
    // opposite radius is what makes this the normal rather than the radius itself.
    const nx = cos * 9;
    const ny = sin * 16;
    const length = Math.hypot(nx, ny) || 1;
    const unitX = nx / length;
    const unitY = ny / length;

    // The seat's own half-extent in the direction it is being pushed, plus a gap.
    const push =
      Math.abs(unitX) * SEAT_HALF_WIDTH + Math.abs(unitY) * SEAT_HALF_HEIGHT + SEAT_GAP;

    return { cos, sin, unitX, unitY, push: `${push.toFixed(2)}rem` };
  }

  // Read the stored preference even when the seat cookie skips the join prompt.
  let rememberingName = $state(nameIsRemembered());

  function seat(name: string, remember: boolean) {
    // The prompt handles the cookie; retain its choice for subsequent renames.
    rememberingName = remember;
    connection.dismissRefusal();
    connection.seat(name);
  }

  let editingName = $state(false);

  function saveName(name: string, remember: boolean) {
    editingName = false;
    rememberingName = remember;
    // Saving can opt in; opting out already deleted the cookie in the dialog.
    if (remember) rememberName(name);
    if (name === myName) return;
    connection.dismissRefusal();
    connection.rename(name);
  }

  function forgetMyName() {
    rememberingName = false;
    forgetName();
  }

  const myName = $derived(
    room?.participants.find((p) => p.id === connection.you)?.name ?? '',
  );
</script>

{#if connection.status === 'refused'}
  <NotAGame reason={connection.refusal} />
{:else if connection.status === 'full'}
  <ServerFull scope={connection.fullScope ?? 'server'} reason={connection.refusal} />
{:else if room === null}
  <main class="waiting">
    <ConnectionState status={connection.status} />
  </main>
{:else if !seated}
  <main class="waiting">
    {#if connection.roundLost}
      <p class="notice standalone" role="status">
        The round was lost — the table was started again. Take a seat to carry on.
        <button class="dismiss" onclick={() => connection.dismissRoundLost()}>Dismiss</button>
      </p>
    {/if}
    <NamePrompt
      participants={room.participants}
      refusal={connection.refusal}
      disabled={!connection.canAct}
      onseat={seat}
    />
    <ConnectionState status={connection.status} />
  </main>
{:else}
  <!-- Mark the last snapshot stale and disable actions while disconnected. -->
  <div class="room" class:stale={!connection.canAct}>
    <!-- Equal outer columns keep New round centered as the connection label changes. -->
    <header>
      <div class="left"><ConnectionState status={connection.status} /></div>

      <div class="middle">
        <button disabled={!connection.canAct} onclick={() => connection.newRound()}>
          New round
        </button>
      </div>

      <div class="right"><InviteControl {roomId} /></div>
    </header>

    <section class="table-area" aria-label="the table">
      <div class="arena">
        <div class="table">

          <!-- Reserve the Reveal slot so hiding it does not shift the status or results. -->
          <div class="action" class:spent={room.revealed}>
            <button
              class="primary"
              disabled={!connection.canAct}
              tabindex={room.revealed ? -1 : 0}
              onclick={() => connection.reveal()}
            >
              Reveal
            </button>
          </div>

          <div class="info">
            {#if room.revealed && room.results}
              <Results results={room.results} deck={room.deck} />
            {:else}
              <p class="status">
                {#if room.everyonePresentHasVoted}
                  Everyone has voted.
                {:else}
                  Waiting for votes…
                {/if}
              </p>
            {/if}
          </div>
        </div>

        <ul class="seats">
          {#each participants as participant, index (participant.id)}
            {@const position = seatPosition(index, participants.length)}
            <Seat
              {participant}
              card={room.results?.cards.find((entry) => entry.id === participant.id)?.card || null}
              revealed={room.revealed}
              isYou={participant.id === connection.you}
              edgeX={position.cos}
              edgeY={position.sin}
              pushX={position.unitX}
              pushY={position.unitY}
              push={position.push}
              canRename={connection.canAct}
              onedit={() => (editingName = true)}
            />
          {/each}
        </ul>
      </div>
    </section>

    <!-- Float notices above the deck to avoid shifting cards. -->
    <!-- Refusals concern an action; room resets discard the whole round. -->
    {#if connection.roundLost}
      <p class="notice" role="status">
        The round was lost — the table was started again. Take a seat to carry on.
        <button class="dismiss" onclick={() => connection.dismissRoundLost()}>Dismiss</button>
      </p>
    {/if}

    {#if connection.refusal}
      <p class="refusal" role="alert">
        {connection.refusal}
        <button class="dismiss" onclick={() => connection.dismissRefusal()}>Dismiss</button>
      </p>
    {/if}

    <Deck
      deck={room.deck}
      played={myCard}
      disabled={!connection.canAct || room.revealed}
      onplay={(card) => {
        connection.dismissRefusal();
        connection.vote(card);
      }}
    />
  </div>

  {#if editingName}
    <NameDialog
      name={myName}
      remember={rememberingName}
      onsave={saveName}
      onforget={forgetMyName}
      oncancel={() => (editingName = false)}
    />
  {/if}
{/if}

<style>
  .waiting {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    width: 100%;
    max-width: 26rem;
  }

  .room {
    display: flex;
    flex-direction: column;
    /* Fill the height the app shell gives this screen, which is the viewport less
       the legal footer when there is one. */
    align-self: stretch;
    width: 100%;
    position: relative;
    transition: opacity 150ms ease;
  }

  .room.stale {
    opacity: 0.45;
  }

  header {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    align-items: start;
    gap: 1rem;
    padding: 1rem 1.25rem 0;
  }

  .left {
    min-width: 0;
  }

  .middle {
    display: flex;
    justify-content: center;
  }

  .right {
    display: flex;
    justify-content: flex-end;
    /* Anchor the clipboard fallback to this header. */
    position: relative;
    min-width: 0;
  }

  .table-area {
    flex: 1;
    display: grid;
    place-items: center;
    /* Reserve space for seats outside the table. */
    padding: 7rem 7.5rem 5rem;
  }

  /*
   * Keep a fixed 16:9 table large enough for the chart. Seats are positioned
   * relative to it, so its dimensions must remain stable across reveal.
   */
  .arena {
    position: relative;
    width: 32rem;
    aspect-ratio: 16 / 9;
  }

  .table {
    position: absolute;
    inset: 0;
    background: var(--felt);
    border: 1px solid var(--felt-edge);
    border-radius: 48% / 34%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.9rem;
    /* Inset the chart from the rounded table edges. */
    padding: 1.9rem 3rem;
    text-align: center;
  }

  /* Reserve chart height and constrain width to keep rows within the rounded felt. */
  .info {

    min-height: 9.5rem;
    width: 100%;
    max-width: 17rem;
    display: flex;
    flex-direction: column;
    /* Align the short status line directly beneath Reveal. */
    justify-content: flex-start;
  }

  .status {
    margin: 0;
    color: var(--text-dim);
    font-size: 0.85rem;
  }

  /*
   * The action and results form one centered group. Keep the action slot when
   * Reveal is hidden to prevent layout shifts.
   */
  .action {

    height: 2.4rem;
    display: flex;
    align-items: center;
  }

  .action.spent {
    visibility: hidden;
  }

  .seats {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .notice {
    position: absolute;
    left: 50%;
    bottom: 11rem;
    transform: translateX(-50%);
    margin: 0;
    padding: 0.6rem 0.9rem;
    border-radius: 8px;
    background: var(--surface-raised);
    border: 1px solid var(--border);
    color: var(--text);
    font-size: 0.85rem;
    display: flex;
    align-items: center;
    gap: 1rem;
    max-width: min(30rem, 90vw);
    z-index: 2;
  }

  /* On the name prompt there is nothing to float above, so it simply sits there. */
  .notice.standalone {
    position: static;
    transform: none;
    width: 100%;
  }

  .refusal {
    position: absolute;
    left: 50%;
    bottom: 7.5rem;
    transform: translateX(-50%);
    margin: 0;
    padding: 0.6rem 0.9rem;
    border-radius: 8px;
    background: var(--bad-surface);
    border: 1px solid var(--bad);
    color: var(--bad);
    font-size: 0.85rem;
    display: flex;
    align-items: center;
    gap: 1rem;
    max-width: min(30rem, 90vw);
    z-index: 2;
  }

  .dismiss {
    background: none;
    border: none;
    color: inherit;
    font: inherit;
    font-size: 0.8rem;
    text-decoration: underline;
    cursor: pointer;
    padding: 0;
    flex: none;
  }

  /* Use the ring only when there is room for the table and full-width seats. */
  @media (min-width: 58rem) {
    .seats > :global(.seat) {
      position: absolute;
      /* Offset each seat from the table edge along the outward normal. */
      left: calc(50% + var(--edge-x) * 50% + var(--push-x) * var(--push));
      top: calc(50% + var(--edge-y) * 50% + var(--push-y) * var(--push));
      transform: translate(-50%, -50%);
    }
  }

  /* On narrow screens, list seats below the table instead of positioning a ring. */
  @media (max-width: 57.999rem) {
    /* Retain centering without reserving space for a ring. */
    .table-area {
      padding: 1.5rem;
    }

    .arena {
      width: min(22rem, 90vw);
      aspect-ratio: auto;
      display: flex;
      flex-direction: column;
      gap: 1.25rem;
    }

    .table {
      /* Restore flow positioning so the table does not overlap the seat list. */
      position: relative;
      order: 1;
      aspect-ratio: auto;
      /* Keep enough vertical space for the results chart. */
      min-height: 20rem;
    }

    .seats {
      order: 2;
      display: flex;
      flex-direction: column;
      width: 100%;
      gap: 0.3rem;
    }

    /* Allow header controls to fit on small screens. */
    header {
      grid-template-columns: auto auto;
      grid-template-areas:
        "status invite"
        "round round";
      justify-content: center;
      align-items: center;
      gap: 0.6rem 1rem;
    }

    .left {
      grid-area: status;
    }

    .right {
      grid-area: invite;
      justify-content: center;
    }

    .middle {
      grid-area: round;
    }
  }
</style>
