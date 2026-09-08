<script lang="ts">
  import { roomURL } from '../lib/router.svelte';

  // Keep the invitation control in a fixed location regardless of participant count.

  interface Props {
    roomId: string;
  }

  let { roomId }: Props = $props();

  let copied = $state(false);
  let fallbackURL = $state<string | null>(null);
  let timer: ReturnType<typeof setTimeout> | null = null;

  async function copy() {
    const url = roomURL(roomId);
    try {
      await navigator.clipboard.writeText(url);
      copied = true;
      fallbackURL = null;
      if (timer !== null) clearTimeout(timer);
      timer = setTimeout(() => (copied = false), 2000);
    } catch {
      // Offer a selectable URL when clipboard access fails or is unavailable.
      fallbackURL = url;
      copied = false;
    }
  }
</script>

<div class="invite">
  <button onclick={copy} title="Copy the invitation link">
    <svg viewBox="0 0 16 16" width="13" height="13" aria-hidden="true" focusable="false">
      <path
        d="M6.5 9.5a3 3 0 004.2 0l2.3-2.3a3 3 0 00-4.2-4.2L7.6 4.2M9.5 6.5a3 3 0 00-4.2 0L3 8.8a3 3 0 004.2 4.2l1.2-1.2"
        fill="none"
        stroke="currentColor"
        stroke-width="1.4"
        stroke-linecap="round"
      />
    </svg>
    {copied ? 'Link copied' : 'Invite players'}
  </button>

  {#if fallbackURL}
    <label class="fallback">
      <span>Copy this link:</span>
      <input type="text" readonly value={fallbackURL} onfocus={(e) => e.currentTarget.select()} />
    </label>
  {/if}
</div>

<style>
  .invite {
    /* Position feedback and fallback without shifting the invitation button. */
    position: relative;
  }

  button {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    background: none;
    border: 1px solid var(--border);
    border-radius: 999px;
    padding: 0.35rem 0.8rem;
    color: var(--text-dim);
    font: inherit;
    font-size: 0.82rem;
    cursor: pointer;
    white-space: nowrap;
  }

  button:hover,
  button:focus-visible {
    color: var(--text);
    border-color: var(--accent);
  }

  .fallback {
    position: absolute;
    top: calc(100% + 0.4rem);
    right: 0;
    width: 18rem;
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    font-size: 0.75rem;
    color: var(--text-dim);
    background: var(--surface);
    border-radius: 6px;
    padding: 0.4rem;
    z-index: 3;
  }

  .fallback input {
    font: inherit;
    font-size: 0.75rem;
    padding: 0.3rem 0.4rem;
    border-radius: 5px;
    border: 1px solid var(--border);
    background: var(--background);
    color: var(--text);
    width: 100%;
  }
</style>
