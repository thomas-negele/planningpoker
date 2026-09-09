<script lang="ts">
  import { onMount } from 'svelte';
  import { legalNoticesEnabled } from '../lib/legal';

  type State = 'asking' | 'enabled' | 'disabled' | 'unavailable';

  // Start in 'asking' and render nothing, so links never appear and then vanish.
  let state = $state<State>('asking');

  // Ask once per page load. This deliberately does not go through the room
  // connection: the answer is the same before anybody has taken a seat, on the
  // entry screen, and while the socket is down.
  onMount(async () => {
    try {
      state = (await legalNoticesEnabled()) ? 'enabled' : 'disabled';
    } catch {
      // A failure here concerns the footer only; it must not disturb the game.
      state = 'unavailable';
    }
  });
</script>

{#if state === 'enabled'}
  <!--
    Open in a new tab: following a link in the current one would tear down the
    room connection and the participant would come back to a reconnecting table.
    rel="noopener noreferrer" is what stops the opened document from reaching
    back into this one.
  -->
  <footer>
    <a href="/legal/privacy" target="_blank" rel="noopener noreferrer">Privacy</a>
    <a href="/legal/imprint" target="_blank" rel="noopener noreferrer">Legal notice</a>
  </footer>
{:else if state === 'unavailable'}
  <footer>
    <p class="unavailable" role="status">Legal information is unavailable.</p>
  </footer>
{/if}

<style>
  footer {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: center;
    gap: 0.35rem 1.25rem;
    /* Normal flow below the screen's content, so it never covers a control. */
    padding: 0.75rem 1.25rem 1rem;
    font-size: 0.85rem;
    color: var(--text-dim);
  }

  a {
    color: var(--text-dim);
    text-decoration: underline;
  }

  a:hover {
    color: var(--text);
  }

  a:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 3px;
    border-radius: 3px;
  }

  .unavailable {
    margin: 0;
  }
</style>
