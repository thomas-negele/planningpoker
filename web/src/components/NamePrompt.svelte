<script lang="ts">
  import {
    MAX_NAME_LENGTH,
    NAME_VISIBILITY_HINT,
    REMEMBERED_FOR,
    forgetName,
    nameIsRemembered,
    rememberName,
    rememberedName,
  } from '../lib/name';
  import type { Participant } from '../lib/protocol';

  // Joining requires confirmation. Name storage is optional and can also be changed
  // from the table. autocomplete=off requests that the browser avoid saving its own
  // copy.

  interface Props {
    participants: Participant[];
    refusal: string | null;
    disabled: boolean;
    onseat: (name: string, remember: boolean) => void;
  }

  let { participants, refusal, disabled, onseat }: Props = $props();

  let name = $state(rememberedName());
  let remember = $state(nameIsRemembered());

  // Delete immediately, even if the form is never submitted.
  function rememberChanged(event: Event) {
    remember = (event.currentTarget as HTMLInputElement).checked;
    if (!remember) forgetName();
  }

  function submit(event: SubmitEvent) {
    event.preventDefault();
    const trimmed = name.trim();
    // Reject an empty field locally; display other validation errors from the server.
    if (trimmed === '') return;
    if (remember) rememberName(trimmed);
    onseat(trimmed, remember);
  }
</script>

<main>
  <h1>Join the game</h1>

  <form onsubmit={submit}>
    <label for="name">Your name</label>
    <input
      id="name"
      type="text"
      bind:value={name}
      maxlength={MAX_NAME_LENGTH}
      placeholder="e.g. Thomas"
      autocomplete="off"
      autocapitalize="words"
      spellcheck="false"
    />
    <p class="hint">{NAME_VISIBILITY_HINT}</p>

    <div class="remember">
      <input id="remember" type="checkbox" checked={remember} onchange={rememberChanged} />
      <label for="remember">
        Remember my name on this device
        <span class="detail">
          Stores the name you typed, in this browser, for {REMEMBERED_FOR}, so you do not have to
          type it again. Nothing else is stored. Untick to delete it.
        </span>
      </label>
    </div>

    <button class="primary" type="submit" disabled={disabled || name.trim() === ''}>
      Take a seat
    </button>
  </form>

  {#if refusal}
    <p class="problem" role="alert">{refusal}</p>
  {/if}

  <section class="already">
    {#if participants.length === 0}
      <p class="dim">Nobody is at the table yet.</p>
    {:else}
      <p class="dim">Already here:</p>
      <ul>
        {#each participants as participant (participant.id)}
          <li class:away={participant.away}>
            {participant.name}{#if participant.away}<span class="tag">away</span>{/if}
          </li>
        {/each}
      </ul>
    {/if}
  </section>
</main>

<style>
  main {
    background: var(--surface);
    padding: 2.5rem 3rem;
    border-radius: 16px;
    max-width: 26rem;
    width: 100%;
  }

  h1 {
    margin: 0 0 1.5rem;
    font-size: 1.5rem;
    text-align: center;
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  label {
    font-size: 0.85rem;
    color: var(--text-dim);
  }

  input {
    font: inherit;
    padding: 0.7rem 0.9rem;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: var(--background);
    color: var(--text);
  }

  input:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 1px;
  }

  /* Keep the name-visibility advice visually secondary to the form. */
  .hint {
    margin: 0.5rem 0 0;
    color: var(--text-dim);
    font-size: 0.78rem;
    line-height: 1.45;
  }

  .remember {
    margin-top: 0.75rem;
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

  button {
    margin-top: 0.5rem;
  }

  .problem {
    margin: 1rem 0 0;
    color: var(--bad);
    font-size: 0.9rem;
  }

  .already {
    margin-top: 2rem;
    border-top: 1px solid var(--border);
    padding-top: 1.25rem;
  }

  .dim {
    margin: 0 0 0.5rem;
    color: var(--text-dim);
    font-size: 0.85rem;
  }

  ul {
    margin: 0;
    padding: 0;
    list-style: none;
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
  }

  li {
    background: var(--background);
    border-radius: 999px;
    padding: 0.25rem 0.75rem;
    font-size: 0.85rem;
  }

  li.away {
    color: var(--text-dim);
  }

  .tag {
    margin-left: 0.4rem;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-dim);
  }
</style>
