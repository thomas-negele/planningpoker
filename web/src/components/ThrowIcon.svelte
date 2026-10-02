<script lang="ts">
  import type { ThrowObject } from '../lib/protocol';
  import { heartColour } from '../lib/throw-objects';

  interface Props {
    object: ThrowObject;
    size?: number;
    seed?: number;
  }

  let { object, size = 32, seed = 0 }: Props = $props();
  const flowerVariant = $derived((seed >>> 0) % 3);
  const flowerPalette = $derived(Math.floor((seed >>> 0) / 3) % 4);
  const heartPalette = $derived(heartColour(seed));
</script>

<svg
  class="throw-icon"
  class:plane={object === 'paper-plane'}
  class:flower-coral={flowerPalette === 0}
  class:flower-blue={flowerPalette === 1}
  class:flower-yellow={flowerPalette === 2}
  class:flower-violet={flowerPalette === 3}
  data-heart={object === 'heart' ? heartPalette : undefined}
  viewBox="0 0 48 48"
  width={size}
  height={size}
  x={-size / 2}
  y={-size / 2}
  aria-hidden="true"
  focusable="false"
>
  {#if object === 'paper-ball'}
    <path
      class="paper ball"
      d="M23.5 4.8c5.1-1.1 9.4 2.1 12.2 5.7 4.8 1.8 7 6.6 6.2 11.2 2 4.5-.6 9.2-4.2 11.5-1 5.1-5.7 8.4-10.3 8.1-4.1 2.5-9.3.8-11.8-2.7-4.8-.5-8.3-4.5-8-9.1-3.1-3.7-1.8-9.2 1.6-12.1.2-4.8 4.3-8.3 8.8-8.4 1.4-2.1 3.2-3.5 5.5-4.2z"
    />
    <path class="ball-shadow" d="M8.2 29.5c5.8 3.9 10.8 2 15.2 4.4 4.2 2.3 8.9.4 14.3-.7-1 5.1-5.7 8.4-10.3 8.1-4.1 2.5-9.3.8-11.8-2.7-4.4-.5-7.7-4-7.4-9.1z" />
    <path class="ball-paper" d="M9.3 17.4l9.2-3.7 6 7.2-4.1 8.3-12.2.3m10.3-15.8l5-8.9 5.2 10.1-4.2 6m0 0l9.5-6.7 7.9 7.5-9.1 4.5m0 0l4.9 7-10.3 8.1-4-7.4m0 0l-7.8 4.7-5.1-8.9" />
    <path class="ball-crease" d="M13.2 20.9l7.2 8.3-6.5 3.6m10.6-11.9l2.9 13m5.4-8.7l-5.4 8.7m-8.9-20.2l5 7.2 5.2-6m-5.3 19l-7.8 4.7" />
    <path class="shine" d="M13.1 14.8c2.8-2.2 6-3 8.5-2.4" />
  {:else if object === 'paper-plane'}
    <path class="paper" d="M4 25L43 8 31 40l-8-12z" />
    <path class="plane-fold" d="M4 25l19 3L43 8 18 24m5 4l8 12" />
    <path class="shine" d="M10 23L36 12" />
  {:else if object === 'heart'}
    <path
      class="heart"
      d="M24 41.5C13.6 34 5.5 27 5.5 17.6 5.5 11.3 10.2 6.5 16 6.5c3.6 0 6.4 1.8 8 4.6 1.6-2.8 4.4-4.6 8-4.6 5.8 0 10.5 4.8 10.5 11.1C42.5 27 34.4 34 24 41.5z"
    />
    <path class="heart-light" d="M11 15.8c.4-3.3 2.8-5.3 5.8-5.2" />
  {:else if object === 'poo'}
    <path class="poo" d="M21.2 13.2c.4-4.4 3.7-7.6 8.9-8.6-1.3 2.4-.9 4.4.9 6.2 1.6 1.6 1.9 3.6 1 5" />
    <path class="poo" d="M8.6 44.5h30.8c3.9 0 6.6-2.7 6.6-6.1 0-3.5-2.8-6.2-6.5-6.2H8.5C4.8 32.2 2 34.9 2 38.4c0 3.4 2.7 6.1 6.6 6.1z" />
    <path class="poo" d="M13.6 33.4h20.8c3.4 0 5.9-2.4 5.9-5.6 0-3.1-2.5-5.5-5.8-5.5H13.5c-3.3 0-5.8 2.4-5.8 5.5 0 3.2 2.5 5.6 5.9 5.6z" />
    <path class="poo" d="M17.6 23.4h12.8c2.9 0 5-2 5-4.7s-2.1-4.8-5-4.8H17.6c-2.9 0-5 2.1-5 4.8s2.1 4.7 5 4.7z" />
    <path class="poo-seam" d="M11.8 33c7.6 1 16.8 1 24.4 0M16.4 23.1c5.1.7 10.1.7 15.2 0" />
    <path class="poo-light" d="M6.4 36.8c.7-1.6 2.1-2.4 3.9-2.5m1.2-8.5c.6-1.3 1.8-2 3.3-2.1m2.1-7.8c.5-1 1.4-1.6 2.6-1.6" />
    <ellipse class="eye" cx="18.6" cy="27.6" rx="3.3" ry="3.9" />
    <ellipse class="eye" cx="29.4" cy="27.6" rx="3.3" ry="3.9" />
    <circle class="pupil" cx="19.2" cy="28.3" r="1.8" />
    <circle class="pupil" cx="30" cy="28.3" r="1.8" />
    <path class="mouth" d="M15.2 36.2c2.2 5.3 15.4 5.3 17.6 0z" />
    <path class="tongue" d="M20.4 39.5c2.2-1.5 5-1.5 7.2 0-2.3.9-4.9.9-7.2 0z" />
  {:else}
    <path class="stem" d="M24 19c-1 9 .8 17 1.5 25" />
    <path class="leaf" d="M24.6 33c-7.2-7.3-11.7-2.6-10.8-1.1 2.1 3.6 6.5 4.6 10.8 4.1m.3 1.1c5.8-6.6 10.2-3.4 9.8-1.8-.8 3.3-4.7 5-9.5 4.7" />
    {#if flowerVariant === 0}
      <g class="petals daisy">
        <ellipse cx="24" cy="8.5" rx="4.1" ry="7" />
        <ellipse cx="24" cy="23.5" rx="4.1" ry="7" />
        <ellipse cx="16.5" cy="16" rx="7" ry="4.1" />
        <ellipse cx="31.5" cy="16" rx="7" ry="4.1" />
        <ellipse cx="18.6" cy="10.6" rx="4" ry="6.5" transform="rotate(-45 18.6 10.6)" />
        <ellipse cx="29.4" cy="21.4" rx="4" ry="6.5" transform="rotate(-45 29.4 21.4)" />
        <ellipse cx="29.4" cy="10.6" rx="4" ry="6.5" transform="rotate(45 29.4 10.6)" />
        <ellipse cx="18.6" cy="21.4" rx="4" ry="6.5" transform="rotate(45 18.6 21.4)" />
      </g>
      <circle class="centre" cx="24" cy="16" r="5.2" />
    {:else if flowerVariant === 1}
      <path class="petal tulip" d="M14 8c3.7.5 6.8 2.7 10 6.2C27.2 10.7 30.3 8.5 34 8c1.3 9.8-2.1 16-10 16S12.7 17.8 14 8z" />
      <path class="petal-light" d="M24 14.2c-2.1-3.1-4.2-5.5-6.3-7.1.1 6.7 2.1 11 6.3 16.9" />
    {:else}
      <g class="petals poppy">
        <path d="M24 16C12 13.5 10.1 5 16.2 4.5 20 4.2 22.4 8.8 24 16z" />
        <path d="M24 16C26.7 5 35 3.8 36.2 9.6c.7 3.7-4 5.5-12.2 6.4z" />
        <path d="M24 16c11.1 1.8 12.6 10.2 6.8 11.2-3.7.7-5.6-3.3-6.8-11.2z" />
        <path d="M24 16c-2 11-10.4 12.3-11.3 6.5C12.1 18.9 16 17.2 24 16z" />
      </g>
      <circle class="centre" cx="24" cy="16" r="4.2" />
    {/if}
  {/if}
</svg>

<style>
  .throw-icon {
    overflow: visible;
    display: block;
  }
  .paper {
    fill: #f4f1e8;
    stroke: #75839a;
    stroke-width: 1.5;
    stroke-linejoin: round;
  }
  .ball {
    stroke-width: 1.25;
  }
  .ball-shadow {
    fill: #c9ced5;
    opacity: 0.72;
  }
  .ball-paper {
    fill: #e4e5e2;
    stroke: #a4acb8;
    stroke-width: 0.9;
    stroke-linecap: round;
    stroke-linejoin: round;
    opacity: 0.78;
  }
  .ball-crease {
    fill: none;
    stroke: #7d8898;
    stroke-width: 1.15;
    stroke-linecap: round;
    stroke-linejoin: round;
    opacity: 0.82;
  }
  .plane-fold {
    fill: none;
    stroke: #9aa6b8;
    stroke-width: 1.35;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  .shine {
    fill: none;
    stroke: #fff;
    stroke-width: 1.7;
    stroke-linecap: round;
    opacity: 0.8;
  }
  .stem,
  .leaf {
    fill: none;
    stroke: #3e8c65;
    stroke-width: 2.3;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  .leaf {
    fill: #5da876;
    stroke-width: 1.4;
  }
  .petals,
  .petal {
    fill: var(--flower-petal);
    stroke: var(--flower-edge);
    stroke-width: 0.9;
    stroke-linejoin: round;
  }
  .daisy ellipse:nth-child(even),
  .poppy path:nth-child(even) {
    fill: var(--flower-petal-alt);
  }
  .petal-light {
    fill: none;
    stroke: var(--flower-light);
    stroke-width: 1.2;
    stroke-linecap: round;
  }
  .centre {
    fill: var(--flower-centre);
    stroke: var(--flower-edge);
    stroke-width: 0.9;
  }
  .heart {
    fill: var(--heart-fill);
    stroke: var(--heart-edge);
    stroke-width: 1.5;
    stroke-linejoin: round;
  }
  .heart-light {
    fill: none;
    stroke: var(--heart-light);
    stroke-width: 2.2;
    stroke-linecap: round;
  }
  [data-heart='red'] {
    --heart-fill: #ec4a5f;
    --heart-light: #ff9aa7;
    --heart-edge: #d7172f;
  }
  [data-heart='pink'] {
    --heart-fill: #f26bb2;
    --heart-light: #ffb6dc;
    --heart-edge: #d01276;
  }
  [data-heart='orange'] {
    --heart-fill: #f98d33;
    --heart-light: #ffc88f;
    --heart-edge: #af5205;
  }
  [data-heart='yellow'] {
    --heart-fill: #fbcd38;
    --heart-light: #fff1a8;
    --heart-edge: #886903;
  }
  [data-heart='green'] {
    --heart-fill: #48c071;
    --heart-light: #a9ecbd;
    --heart-edge: #2b7c46;
  }
  [data-heart='blue'] {
    --heart-fill: #4b93ef;
    --heart-light: #aed0fb;
    --heart-edge: #1369d7;
  }
  [data-heart='purple'] {
    --heart-fill: #9d68e6;
    --heart-light: #d5bcf7;
    --heart-edge: #8847e1;
  }
  /* The outline is kept at 3:1 against the dark table; the seams carry the tiers. */
  .poo {
    fill: #a86e3f;
    stroke: #956138;
    stroke-width: 1.4;
    stroke-linejoin: round;
  }
  .poo-seam {
    fill: none;
    stroke: #5a361b;
    stroke-width: 1.5;
    stroke-linecap: round;
  }
  .poo-light {
    fill: none;
    stroke: #d6a274;
    stroke-width: 1.6;
    stroke-linecap: round;
  }
  .eye {
    fill: #fff;
    stroke: #55311a;
    stroke-width: 1;
  }
  .pupil {
    fill: #2b1a0e;
  }
  .mouth {
    fill: #3b2010;
    stroke: #3b2010;
    stroke-width: 1;
    stroke-linejoin: round;
  }
  .tongue {
    fill: #e0727a;
  }
  .flower-coral {
    --flower-petal: #f36f8b;
    --flower-petal-alt: #ff9aad;
    --flower-edge: #8e3d59;
    --flower-light: #ffc0ca;
    --flower-centre: #f3c84b;
  }
  .flower-blue {
    --flower-petal: #719ee8;
    --flower-petal-alt: #a7c5f4;
    --flower-edge: #395f9c;
    --flower-light: #d8e6ff;
    --flower-centre: #ffd45c;
  }
  .flower-yellow {
    --flower-petal: #f2b94b;
    --flower-petal-alt: #ffd978;
    --flower-edge: #9a681f;
    --flower-light: #fff0b1;
    --flower-centre: #684326;
  }
  .flower-violet {
    --flower-petal: #aa7ae8;
    --flower-petal-alt: #d0adf4;
    --flower-edge: #664296;
    --flower-light: #ead8ff;
    --flower-centre: #f2c84b;
  }
</style>
