<script lang="ts">
  // The visitor-mode choice, used both when taking a seat and when editing one's own
  // name. Visitor mode belongs to the seat, not to the remembered-name preference,
  // so it is never stored in a cookie.
  //
  // The explanation uses the same InfoControl as the other choices on these two
  // forms. It had its own once, which meant two differently drawn "i" symbols
  // three rows apart on the join screen, and that one kept its note at
  // visibility:hidden while closed — which takes it out of the accessibility
  // tree, so the aria-describedby below resolved to nothing until the button was
  // pressed. Sharing the control settles both.

  import InfoControl from './InfoControl.svelte';

  interface Props {
    /** The current choice. Bound, so the parent decides when to send it. */
    checked: boolean;
    /** Distinguishes the input ids when both forms exist on one page. */
    idPrefix: string;
  }

  let { checked = $bindable(), idPrefix }: Props = $props();

  const checkboxId = $derived(`${idPrefix}-visitor`);
  const hintId = $derived(`${idPrefix}-visitor-hint`);
</script>

<div class="visitor">
  <input id={checkboxId} type="checkbox" aria-describedby={hintId} bind:checked />
  <div class="text">
    <label for={checkboxId}>Visitor mode</label>

    <InfoControl id={hintId} label="What visitor mode means" text="Visitors cannot vote." />
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

</style>
