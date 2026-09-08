<script lang="ts">
  import type { Card, Deck, Results } from '../lib/protocol';

  // Render per-card counts using the deck's scale and ordering. Do not infer a
  // winning estimate or compute averages over categorical cards.

  interface Props {
    results: Results;
    /** The room's deck, which is what says which cards belong on the size scale. */
    deck: Deck;
  }

  let { results, deck }: Props = $props();

  interface Row {
    card: Card;
    count: number;
    /** Share of the longest bar in this round, 0–1. Zero when nobody played it. */
    fraction: number;
  }

  function countOf(card: Card): number {
    return results.tally.find((entry) => entry.card === card)?.count ?? 0;
  }

  // Normalize every bar against the largest count across all cards.
  const busiest = $derived(Math.max(0, ...results.tally.map((entry) => entry.count)));

  function rowsFor(cards: Card[]): Row[] {
    return cards.map((card) => {
      const count = countOf(card);
      return { card, count, fraction: busiest === 0 ? 0 : count / busiest };
    });
  }

  // Keep all size rows, including zero counts, with larger sizes above smaller ones.
  const scale = $derived(rowsFor([...deck.scale].reverse()));

  // Derive non-scale cards from the deck instead of hardcoding exceptions.
  const asides = $derived(rowsFor(deck.cards.filter((card) => !deck.scale.includes(card))));

  // Show non-scale cards only when at least one was played.
  const showAsides = $derived(asides.some((row) => row.count > 0));

  const voters = $derived(results.tally.reduce((total, entry) => total + entry.count, 0));
  const agreed = $derived(results.tally.length === 1 && voters > 1);
</script>

<div class="results">
  <!-- Expose card/count pairs to assistive technology; the bars are decorative. -->
  <dl class="chart">
    {#each scale as row (row.card)}
      <div class="row" class:none={row.count === 0}>
        <dt class="card">{row.card}</dt>
        <dd class="value">
          <span class="track" aria-hidden="true">
            <span class="bar" style:width="{row.fraction * 100}%"></span>
          </span>
          <span class="count">{row.count}</span>
        </dd>
      </div>
    {/each}
  </dl>

  {#if showAsides}
    <dl class="chart asides">
      {#each asides as row (row.card)}
        <div class="row" class:none={row.count === 0}>
          <dt class="card">{row.card}</dt>
          <dd class="value">
            <span class="track" aria-hidden="true">
              <span class="bar" style:width="{row.fraction * 100}%"></span>
            </span>
            <span class="count">{row.count}</span>
          </dd>
        </div>
      {/each}
    </dl>
  {/if}

  <p class="summary">
    {#if voters === 0}
      Nobody voted in this round.
    {:else if agreed}
      Everyone who voted chose the same card.
    {:else}
      {voters} {voters === 1 ? 'vote' : 'votes'}.
    {/if}
  </p>
</div>

<style>
  .results {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 0.45rem;
    width: 100%;
  }

  .chart {
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  /* Non-scale cards share one row beneath the size scale. */
  .chart.asides {
    flex-direction: row;
    justify-content: center;
    gap: 1.1rem;
    border-top: 1px solid var(--felt-edge);
    padding-top: 0.4rem;
    margin-top: 0.15rem;
  }

  .chart.asides .row {
    flex: 1;
    max-width: 9rem;
  }

  .row {
    display: flex;
    align-items: center;
    gap: 0.45rem;
  }

  dt.card {
    flex: none;
    width: 1.9rem;
    text-align: right;
    font-size: 0.8rem;
    font-weight: 600;
    color: var(--text);
  }

  dd.value {
    margin: 0;
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .track {
    flex: 1;
    min-width: 0;
    height: 0.75rem;
    border-radius: 3px;
    background: var(--felt-edge);
    overflow: hidden;
  }

  /* Use the same bar color for every card; length communicates the count. */
  .bar {
    display: block;
    height: 100%;
    border-radius: 3px;
    background: var(--card-face);
  }

  .count {
    flex: none;
    width: 0.9rem;
    font-size: 0.78rem;
    font-variant-numeric: tabular-nums;
    color: var(--text-dim);
  }

  /* Keep zero-count rows visible but subdued. */
  .row.none dt.card,
  .row.none .count {
    color: var(--text-dim);
    opacity: 0.6;
  }

  .summary {
    margin: 0.1rem 0 0;
    text-align: center;
    font-size: 0.78rem;
    color: var(--text-dim);
  }
</style>
