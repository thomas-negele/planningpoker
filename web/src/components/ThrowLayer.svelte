<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import type { RoomConnection } from '../lib/connection.svelte';
  import {
    canAdmitThrowEffect,
    computeThrowFrame,
    createFlight,
    settleForReducedMotion,
    type Flight,
    type LandingZone,
    type RenderedEffect,
  } from '../lib/throw-effects';
  import type { ThrowObject, ThrownMessage } from '../lib/protocol';
  import ThrowIcon from './ThrowIcon.svelte';

  const FLIGHT_SIZES: Record<ThrowObject, number> = {
    'paper-ball': 36,
    'paper-plane': 36,
    flowers: 42,
    heart: 36,
    poo: 40,
  };
  // The pile of poo squashes onto its base, which sits this far below the icon's
  // centre as a fraction of its size; everything else scales about its centre.
  const POO_BASE = 0.44;

  // Below this width seats form a list instead of a table; the same breakpoint as the layout
  // in RoomView.svelte and Seat.svelte.
  const NARROW_LIST = '(max-width: 57.999rem)';
  // Keeps a resting object clear of the name and the controls it lies between.
  const ROW_GAP = 4;

  interface Props { connection: RoomConnection }
  interface Effect {
    event: ThrownMessage;
    flight: Flight;
    startedAt: number;
    skipFlight: boolean;
  }

  let { connection }: Props = $props();
  let effects = $state<Effect[]>([]);
  let rendered = $state<RenderedEffect<Effect>[]>([]);
  let frame = 0;
  let unsubscribe = () => {};
  let reduced = false;
  let motion: MediaQueryList | null = null;
  let narrowList: MediaQueryList | null = null;

  function landingZone(target: string): LandingZone | null {
    const node = document.querySelector<HTMLElement>(`[data-seat-id="${CSS.escape(target)}"]`);
    if (!node) return null;
    const rect = node.getBoundingClientRect();
    if (
      rect.bottom <= 0 ||
      rect.top >= window.innerHeight ||
      rect.right <= 0 ||
      rect.left >= window.innerWidth
    ) return null;
    if (!narrowList?.matches) return { left: rect.left, right: rect.right, bottom: rect.bottom };

    // In a list row: from the end of the name's content to the first of the row's controls.
    const name = node.querySelector('[data-landing-after]');
    const nameEnd = Math.max(
      rect.left,
      ...Array.from(name?.children ?? [], (child) => child.getBoundingClientRect().right),
    );
    const controlsStart = Math.min(
      rect.right,
      ...Array.from(node.querySelectorAll('[data-landing-before]'), (control) =>
        control.getBoundingClientRect().left,
      ),
    );
    return {
      left: nameEnd + ROW_GAP,
      right: controlsStart - ROW_GAP,
      middle: (rect.top + rect.bottom) / 2,
    };
  }

  function receive(event: ThrownMessage) {
    const zone = landingZone(event.target);
    if (!canAdmitThrowEffect(effects.length, event.ageMs, zone !== null) || zone === null) return;
    const flight = createFlight(
      event.seed,
      event.object,
      { width: window.innerWidth, height: window.innerHeight },
      zone,
    );
    effects = [...effects, { event, flight, startedAt: performance.now(), skipFlight: reduced }];
    startFrames();
  }

  function updateVisuals(clock: number) {
    const scene = computeThrowFrame(effects, clock, landingZone);
    effects = scene.active;
    rendered = scene.rendered;
  }

  function tick(clock: number) {
    frame = 0;
    updateVisuals(clock);
    startFrames();
  }

  function startFrames() {
    if (frame === 0 && effects.length > 0 && connection.canAct) {
      frame = requestAnimationFrame(tick);
    }
  }

  function layoutChanged() {
    const clock = performance.now();
    effects = effects.filter(
      (effect) => effect.skipFlight || clock - effect.startedAt >= effect.flight.duration,
    );
    updateVisuals(clock);
    startFrames();
  }

  function motionChanged() {
    reduced = motion?.matches ?? false;
    if (reduced) {
      // Once an in-flight object is settled for reduced motion, it must not resume
      // travelling if the preference changes back before its original flight time.
      const clock = performance.now();
      effects = effects.map((effect) => settleForReducedMotion(effect, clock));
      updateVisuals(clock);
    }
    startFrames();
  }

  $effect(() => {
    if (!connection.canAct) {
      effects = [];
      rendered = [];
      if (frame !== 0) cancelAnimationFrame(frame);
      frame = 0;
    }
  });

  onMount(() => {
    unsubscribe = connection.onThrow(receive);
    narrowList = matchMedia(NARROW_LIST);
    motion = matchMedia('(prefers-reduced-motion: reduce)');
    motionChanged();
    motion.addEventListener('change', motionChanged);
    window.addEventListener('resize', layoutChanged);
    window.addEventListener('scroll', layoutChanged, true);
  });

  onDestroy(() => {
    unsubscribe();
    motion?.removeEventListener('change', motionChanged);
    window.removeEventListener('resize', layoutChanged);
    window.removeEventListener('scroll', layoutChanged, true);
    if (frame !== 0) cancelAnimationFrame(frame);
  });
</script>

<svg class="effects" width="100%" height="100%" aria-hidden="true">
  {#each rendered as item (item.effect.event.id)}
    {@const size = FLIGHT_SIZES[item.effect.event.object]}
    {@const anchor = item.effect.event.object === 'poo' ? size * POO_BASE : 0}
    <g
      transform={`translate(${item.pose.x} ${item.pose.y}) rotate(${item.pose.rotation}) ` +
        `translate(0 ${anchor}) scale(${item.pose.scaleX} ${item.pose.scaleY}) translate(0 ${-anchor})`}
      opacity={item.pose.opacity}
    >
      <ThrowIcon object={item.effect.event.object} seed={item.effect.event.seed} {size} />
    </g>
  {/each}
</svg>

<style>
  .effects {
    position: fixed;
    inset: 0;
    overflow: hidden;
    pointer-events: none;
    z-index: 5;
  }
</style>
