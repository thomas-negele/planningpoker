<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import type { RoomConnection } from '../lib/connection.svelte';
  import {
    canAdmitThrowEffect,
    computeThrowFrame,
    createFlight,
    settleForReducedMotion,
    type Flight,
    type RenderedEffect,
  } from '../lib/throw-effects';
  import type { ThrownMessage } from '../lib/protocol';
  import ThrowIcon from './ThrowIcon.svelte';

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

  function targetRect(target: string): DOMRect | null {
    const node = document.querySelector<HTMLElement>(`[data-seat-id="${CSS.escape(target)}"]`);
    if (!node) return null;
    const rect = node.getBoundingClientRect();
    if (
      rect.bottom <= 0 ||
      rect.top >= window.innerHeight ||
      rect.right <= 0 ||
      rect.left >= window.innerWidth
    ) return null;
    return rect;
  }

  function receive(event: ThrownMessage) {
    const rect = targetRect(event.target);
    if (!canAdmitThrowEffect(effects.length, event.ageMs, rect !== null) || rect === null) return;
    const flight = createFlight(
      event.seed,
      event.object,
      { width: window.innerWidth, height: window.innerHeight },
      rect,
    );
    effects = [...effects, { event, flight, startedAt: performance.now(), skipFlight: reduced }];
    startFrames();
  }

  function updateVisuals(clock: number) {
    const scene = computeThrowFrame(effects, clock, targetRect);
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
    <g
      transform={`translate(${item.pose.x} ${item.pose.y}) rotate(${item.pose.rotation})`}
      opacity={item.pose.opacity}
    >
      <ThrowIcon
        object={item.effect.event.object}
        seed={item.effect.event.seed}
        size={item.effect.event.object === 'flowers' ? 42 : 36}
      />
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
