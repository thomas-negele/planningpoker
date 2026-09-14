<script lang="ts">
  import { untrack } from 'svelte';

  import {
    MAX_NAME_LENGTH,
    NAME_STORAGE_CHOICE,
    NAME_STORAGE_DETAIL,
    NAME_VISIBILITY_HINT,
  } from '../lib/name';
  import InfoControl from './InfoControl.svelte';
  import VisitorModeField from './VisitorModeField.svelte';

  // Edit the name, the storage preference and visitor mode without blocking room
  // updates.

  interface Props {
    name: string;
    remember: boolean;
    visitor: boolean;
    onsave: (name: string, remember: boolean, visitor: boolean) => void;
    onforget: () => void;
    oncancel: () => void;
  }

  let { name, remember, visitor, onsave, onforget, oncancel }: Props = $props();

  // Initialize the draft once so incoming snapshots cannot overwrite unsaved edits.
  let draft = $state(untrack(() => name));
  let keep = $state(untrack(() => remember));

  // The mode is a draft until Save, so cancelling leaves the seat as it is.
  let asVisitor = $state(untrack(() => visitor));

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
    onsave(trimmed, keep, asVisitor);
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
      <span class="titled">
        <span class="field">Your name</span>
        <InfoControl
          id="dialog-name-visibility"
          label="Who can see this name"
          text={NAME_VISIBILITY_HINT}
        />
      </span>

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
        aria-describedby="dialog-name-visibility"
        onkeydown={(e) => e.key === 'Escape' && oncancel()}
      />

      <div class="remember">
        <input
          id="dialog-remember"
          type="checkbox"
          checked={keep}
          aria-describedby="dialog-name-storage"
          onchange={keepChanged}
        />
        <span class="titled">
          <label for="dialog-remember">{NAME_STORAGE_CHOICE}</label>
          <InfoControl
            id="dialog-name-storage"
            label="What is stored, and how to delete it"
            text={NAME_STORAGE_DETAIL}
          />
        </span>
      </div>

      <VisitorModeField bind:checked={asVisitor} idPrefix="dialog" />

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
    border-radius: var(--radius-panel);
    border: 1px solid var(--panel-border);
    box-shadow: var(--shadow-panel);
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
    border-radius: var(--radius-control);
    border: 1px solid var(--border);
    background: var(--background);
    color: var(--text);
  }

  input[type='text']:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }

  /* A label and the information control that belongs to it, kept on one line. */
  .titled {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
  }

  .field {
    font-size: 0.85rem;
    color: var(--text-dim);
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

  .remember .titled {
    align-items: baseline;
  }

  .actions {
    margin-top: 1rem;
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
  }
</style>
