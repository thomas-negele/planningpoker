<script lang="ts">
  // Explain which capacity ceiling was reached and offer a manual retry.

  interface Props {
    /** Which thing was full. */
    scope: 'server' | 'room';
    /** The server's reason, already a sentence. */
    reason: string | null;
  }

  let { scope, reason }: Props = $props();

  const WORDING = {
    server: {
      title: 'This server is full',
      fallback:
        'This server is running as many games as it can right now. Please try again in a few minutes.',
      detail:
        'Your link is fine — nothing is wrong with it. It will work again once one of the games running now has finished.',
    },
    room: {
      title: 'This room is full',
      fallback: 'This room already has as many connections open as it allows.',
      detail:
        'Your link is fine, and the server is not busy — this one room simply has too many connections open. If you have it open in another tab or window, closing that one frees a place immediately.',
    },
  } as const;

  const words = $derived(WORDING[scope]);
</script>

<main>
  <h1>{words.title}</h1>

  <p class="reason">{reason ?? words.fallback}</p>

  <p class="detail">{words.detail}</p>

  <button class="primary" onclick={() => window.location.reload()}>Try this link again</button>
</main>

<style>
  main {
    background: var(--surface);
    padding: 2.5rem 3rem;
    border-radius: 16px;
    text-align: center;
    max-width: 30rem;
  }

  h1 {
    margin: 0 0 0.75rem;
    font-size: 1.5rem;
  }

  .reason {
    margin: 0 0 0.75rem;
    line-height: 1.55;
  }

  .detail {
    margin: 0 0 1.75rem;
    color: var(--text-dim);
    line-height: 1.55;
    font-size: 0.9rem;
  }
</style>
