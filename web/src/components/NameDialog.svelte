<script lang="ts">
  import { untrack } from 'svelte';

  import { MAX_NAME_LENGTH, NAME_VISIBILITY_HINT, REMEMBERED_FOR } from '../lib/name';

  // Edit the name and storage preference without blocking room updates.

  interface Props {
    name: string;
    remember: boolean;
    onsave: (name: string, remember: boolean) => void;
    onforget: () => void;
    oncancel: () => void;
  }

  let { name, remember, onsave, onforget, oncancel }: Props = $props();

  // Initialize the draft once so incoming snapshots cannot overwrite unsaved edits.
  let draft = $state(untrack(() => name));
  let keep = $state(untrack(() => remember));

  // Delete immediately, including when the dialog is later cancelled.
  function keepChanged(event: Event) {
    keep = (event.currentTarget as HTMLInputElement).checked;
    if (!keep) onforget();
  }

  function submit(event: SubmitEvent) {
    event.preventDefault();
    const trimmed = draft.trim();
    // The server validates submitted names.
    if (trimmed === '') return;
    onsave(trimmed, keep);
  }
</script>

<div
  class="backdrop"
  role="presentation"
  onclick={(e) => e.target === e.currentTarget && oncancel()}
>
  <div class="dialog" role="dialog" aria-modal="true" aria-labelledby="name-dialog-title">
    <h2 id="name-dialog-title">Your name</h2>

    <form onsubmit={submit}>
      <!-- svelte-ignore a11y_autofocus -->
      <input
        id="dialog-name"
        type="text"
        bind:value={draft}
        autofocus
        maxlength={MAX_NAME_LENGTH}
        autocomplete="off"
        autocapitalize="words"
        spellcheck="false"
        aria-label="Your name"
        onkeydown={(e) => e.key === 'Escape' && oncancel()}
      />

      <p class="hint">{NAME_VISIBILITY_HINT}</p>

      <div class="remember">
        <input id="dialog-remember" type="checkbox" checked={keep} onchange={keepChanged} />
        <label for="dialog-remember">
          Remember my name on this device
          <span class="detail">
            Stores the name you typed, in this browser, for {REMEMBERED_FOR}, so you do not have
            to type it again. Nothing else is stored. Untick to delete it.
          </span>
        </label>
      </div>

      <div class="actions">
        <button type="button" class="secondary" onclick={oncancel}>Cancel</button>
        <button type="submit" class="primary" disabled={draft.trim() === ''}>Save</button>
      </div>
    </form>
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
    background: var(--surface);
    padding: 1.75rem 2rem;
    border-radius: 16px;
    width: min(24rem, 100%);
  }

  h2 {
    margin: 0 0 1rem;
    font-size: 1.15rem;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  input[type='text'] {
    font: inherit;
    padding: 0.7rem 0.9rem;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: var(--background);
    color: var(--text);
  }

  input[type='text']:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }

  .hint {
    margin: 0.5rem 0 0;
    color: var(--text-dim);
    font-size: 0.78rem;
    line-height: 1.45;
  }

  .remember {
    margin-top: 0.5rem;
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.5rem;
    align-items: start;
  }

  .remember input {
    margin-top: 0.15rem;
    width: 1rem;
    height: 1rem;
    accent-color: var(--accent);
  }

  .remember label {
    font-size: 0.85rem;
    color: var(--text);
    cursor: pointer;
  }

  .detail {
    display: block;
    margin-top: 0.25rem;
    color: var(--text-dim);
    font-size: 0.78rem;
    line-height: 1.45;
  }

  .actions {
    margin-top: 1rem;
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }
</style>
