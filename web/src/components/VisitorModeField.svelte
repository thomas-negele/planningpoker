<script lang="ts">
  // The visitor-mode choice, used both when taking a seat and when editing one's own
  // name. Visitor mode belongs to the seat, not to the remembered-name preference,
  // so it is never stored in a cookie.

  interface Props {
    /** The current choice. Bound, so the parent decides when to send it. */
    checked: boolean;
    /** Distinguishes the input ids when both forms exist on one page. */
    idPrefix: string;
  }

  let { checked = $bindable(), idPrefix }: Props = $props();

  const checkboxId = $derived(`${idPrefix}-visitor`);
  const hintId = $derived(`${idPrefix}-visitor-hint`);

  // The explanation opens on click and also appears while the control is hovered or
  // focused, so a keyboard user reads it without having to activate anything.
  let explaining = $state(false);
</script>

<div class="visitor">
  <input id={checkboxId} type="checkbox" aria-describedby={hintId} bind:checked />
  <div class="text">
    <label for={checkboxId}>Visitor mode</label>

    <!-- The explanation is positioned out of the flow on purpose. A hint that took up
         space would shrink the form again the moment the pointer left the control,
         moving the checkbox out from under the click that was on its way to it. -->
    <span class="info-anchor">
      <button
        type="button"
        class="info"
        aria-label="What visitor mode means"
        aria-describedby={hintId}
        aria-expanded={explaining}
        aria-controls={hintId}
        onclick={() => (explaining = !explaining)}
      >
        i
      </button>
      <span id={hintId} class="detail" role="note">Visitors cannot vote.</span>
    </span>
  </div>
</div>

<style>
  .visitor {
    margin-top: 0.75rem;
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.5rem;
    align-items: start;
  }

  .visitor input {
    margin-top: 0.15rem;
    width: 1rem;
    height: 1rem;
    accent-color: var(--accent);
  }

  .text {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    /* The label sits on the same line as the checkbox it names. */
    min-height: 1.3rem;
  }

  label {
    font-size: 0.85rem;
    color: var(--text);
    cursor: pointer;
  }

  .info-anchor {
    position: relative;
    display: inline-flex;
  }

  .info {
    display: inline-grid;
    place-items: center;
    width: 1.1rem;
    height: 1.1rem;
    padding: 0;
    border: 1px solid var(--border);
    border-radius: 999px;
    background: none;
    color: var(--text-dim);
    font-size: 0.7rem;
    font-style: italic;
    font-weight: 600;
    line-height: 1;
    cursor: pointer;
  }

  .info:hover,
  .info:focus-visible {
    color: var(--text);
    border-color: var(--accent-dim);
  }

  /* Kept in the document for assistive technology, which reaches it through
     aria-describedby, and shown on hover, focus or activation. */
  .detail {
    position: absolute;
    top: calc(100% + 0.4rem);
    left: -0.5rem;
    z-index: 5;
    width: max-content;
    max-width: 14rem;
    padding: 0.4rem 0.55rem;
    border: 1px solid var(--border);
    border-radius: 6px;
    background: var(--surface-raised);
    color: var(--text);
    font-size: 0.75rem;
    line-height: 1.4;
    opacity: 0;
    visibility: hidden;
    transition: opacity 100ms ease, visibility 100ms;
  }

  .info:hover ~ .detail,
  .info:focus-visible ~ .detail,
  .info[aria-expanded='true'] ~ .detail {
    opacity: 1;
    visibility: visible;
  }
</style>
