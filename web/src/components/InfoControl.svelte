<script lang="ts">
  import { tick } from 'svelte';

  // A small "i" beside a control, holding the sentence that explains it.
  //
  // Two things this must get right, because the storage choice depends on it.
  // It has to be reachable by every means of operating the page — hover alone
  // would put the text out of reach on a touchscreen and for anybody working
  // without a mouse. And the text has to stay in the accessibility tree even
  // while it is invisible, so that a screen reader announces it together with
  // the control it belongs to; that is why the closed state is opacity alone
  // and never visibility or display, both of which would hide it from
  // assistive technology as well as from the eye.

  interface Props {
    /** The sentence to disclose. */
    text: string;

    /** What the trigger is called, e.g. "Who can see this name". */
    label: string;

    /** Identifier the explained control points at with aria-describedby. */
    id: string;
  }

  let { text, label, id }: Props = $props();

  // Three independent reasons to be open, so that moving the pointer away does
  // not close a bubble the keyboard is still in, and tapping does not close one
  // the pointer happens to rest on.
  let hovered = $state(false);
  let focused = $state(false);
  let pinned = $state(false);

  // Escape closes the text while the trigger keeps the focus, which is what a
  // person expects; without this the focus would immediately reopen it.
  let dismissed = $state(false);

  const open = $derived(!dismissed && (hovered || focused || pinned));

  /**
   * Whether this focus came from the keyboard.
   *
   * Tapping a button focuses it too. If focus counted regardless, a tap would
   * open the text and no second tap could ever close it again, because the
   * focus would still be holding it open. :focus-visible is the browser's own
   * answer to "did they arrive here by keyboard", so it is the one used.
   */
  function focusedByKeyboard(element: HTMLElement): boolean {
    try {
      return element.matches(':focus-visible');
    } catch {
      // Very old engines do not know the selector; treating focus as keyboard
      // focus there is the harmless way to be wrong.
      return true;
    }
  }

  let root = $state<HTMLElement>();
  let trigger = $state<HTMLButtonElement>();
  let bubble = $state<HTMLElement>();

  // How far the bubble has been nudged sideways to stay on screen.
  let shift = $state(0);

  /**
   * Keep the bubble inside the panel it belongs to, and inside the viewport in
   * any case. It is centred on a trigger that may sit close to either edge, and
   * a text that hangs off the card — or worse, off the screen — explains
   * nothing. Whichever of the two is narrower wins.
   */
  async function place() {
    shift = 0;
    await tick();
    if (!bubble) return;

    const margin = 8;
    const panel = root?.closest('main, .dialog')?.getBoundingClientRect();
    const left = Math.max(margin, (panel?.left ?? 0) + margin);
    const right = Math.min(window.innerWidth - margin, (panel?.right ?? window.innerWidth) - margin);

    const box = bubble.getBoundingClientRect();
    let delta = 0;
    if (box.right > right) delta = right - box.right;
    if (box.left + delta < left) delta = left - box.left;
    shift = delta;
  }

  $effect(() => {
    if (open) void place();
  });

  async function dismiss() {
    dismissed = true;
    pinned = false;
    hovered = false;
    await tick();
    trigger?.focus();
  }

  function outside(event: PointerEvent) {
    if (pinned && event.target instanceof Node && !root?.contains(event.target)) pinned = false;
  }

  function keydown(event: KeyboardEvent) {
    if (open && event.key === 'Escape') {
      event.preventDefault();
      dismiss();
    }
  }
</script>

<svelte:window onpointerdown={outside} onkeydown={keydown} onresize={() => open && place()} />

<span class="info" bind:this={root}>
  <button
    bind:this={trigger}
    type="button"
    class="trigger"
    aria-label={label}
    aria-expanded={open}
    aria-describedby={id}
    onfocus={(event) => (focused = focusedByKeyboard(event.currentTarget))}
    onblur={() => {
      focused = false;
      dismissed = false;
    }}
    onpointerenter={(event) => {
      if (event.pointerType !== 'mouse') return;
      hovered = true;
      dismissed = false;
    }}
    onpointerleave={() => (hovered = false)}
    onclick={() => {
      dismissed = false;
      pinned = !pinned;
    }}
  >
    <svg viewBox="0 0 16 16" width="13" height="13" aria-hidden="true" focusable="false">
      <circle cx="8" cy="8" r="6.6" fill="none" stroke="currentColor" stroke-width="1.3" />
      <circle cx="8" cy="4.9" r="0.85" fill="currentColor" />
      <path
        d="M8 7.2v4.3"
        fill="none"
        stroke="currentColor"
        stroke-width="1.4"
        stroke-linecap="round"
      />
    </svg>
  </button>

  <span
    class="bubble"
    class:shown={open}
    {id}
    role="note"
    bind:this={bubble}
    style:--shift="{shift}px"
  >
    {text}
  </span>
</span>

<style>
  .info {
    position: relative;
    display: inline-flex;
    vertical-align: middle;
  }

  .trigger {
    display: grid;
    place-items: center;
    width: 1.25rem;
    height: 1.25rem;
    padding: 0;
    border: none;
    border-radius: 50%;
    background: none;
    color: var(--text-dim);
    cursor: pointer;
  }

  .trigger:hover,
  .trigger:focus-visible {
    color: var(--text);
  }

  .bubble {
    position: absolute;
    bottom: calc(100% + 0.4rem);
    left: 50%;
    z-index: 5;
    width: max-content;
    max-width: min(17rem, 70vw);
    padding: 0.55rem 0.7rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-control);
    background: var(--surface-raised);
    box-shadow: var(--shadow-panel);
    color: var(--text);
    font-size: 0.78rem;
    font-weight: 400;
    line-height: 1.45;
    text-align: left;
    white-space: normal;

    /* Closed is opacity alone. Anything that removes the element from the
       accessibility tree would also silence the description this control
       exists to provide. */
    opacity: 0;
    pointer-events: none;

    /* The sideways nudge that keeps this inside its panel lives in the same
       transform as any entrance movement would. Animating the transform would
       therefore animate the position, sliding the text across the screen on its
       way to where it belongs. Only the opacity is allowed to change over
       time. */
    transform: translate(calc(-50% + var(--shift, 0px)), 0);
    transition: opacity 110ms ease;
  }

  .bubble.shown {
    opacity: 1;
  }

  @media (prefers-reduced-motion: reduce) {
    .bubble {
      transition: none;
    }
  }
</style>
