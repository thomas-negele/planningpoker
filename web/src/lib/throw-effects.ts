import type { ThrowObject } from './protocol';

export const MIN_FLIGHT_MS = 600;
export const MAX_FLIGHT_MS = 1200;
export const REST_MS = 2000;
export const FADE_MS = 600;
export const MAX_LIVE_THROWS = 48;

// Per-object motion. The heavier the object, the shorter it slides after impact.
/**
 * Half an object's drawn extent: how far beyond the viewport edge a flight starts, and how far
 * from the edges of a narrow-list row it comes to rest.
 */
const RADIUS: Record<ThrowObject, number> = {
  'paper-ball': 18,
  'paper-plane': 24,
  flowers: 25,
  heart: 19,
  poo: 22,
};
/** Full turns during flight from one seeded draw in [0, 1); null draws nothing. */
const TURNS: Record<ThrowObject, ((draw: number) => number) | null> = {
  'paper-ball': (draw) => 1.8 + draw * 2.2,
  'paper-plane': null,
  flowers: (draw) => 0.3 + draw * 0.55,
  heart: (draw) => 0.2 + draw * 0.35,
  poo: null,
};
/** A heavy object flies flatter, a light one higher. */
const ARC_SCALE: Record<ThrowObject, number> = {
  'paper-ball': 1,
  'paper-plane': 1,
  flowers: 1,
  heart: 1.15,
  poo: 0.55,
};
const SLIDE_RANGES: Record<ThrowObject, [number, number]> = {
  'paper-ball': [20, 38],
  'paper-plane': [14, 30],
  flowers: [8, 20],
  heart: [8, 16],
  poo: [2, 6],
};
const BOUNCE: Record<ThrowObject, number> = {
  'paper-ball': 9,
  'paper-plane': 1.5,
  flowers: 3,
  heart: 2.5,
  poo: 2.5,
};
/** Resting tilt in degrees from one seeded draw in [0, 1). */
const RESTING_TILT: Record<Exclude<ThrowObject, 'paper-ball'>, (draw: number) => number> = {
  'paper-plane': (draw) => 5 + draw * 7,
  flowers: (draw) => 7 + draw * 10,
  heart: (draw) => 4 + draw * 8,
  poo: (draw) => 2 + draw * 4,
};
const POO_SQUASH_WIDEN = 0.25;
const POO_SQUASH_FLATTEN = 0.3;
const HEART_PULSE = 0.25;

export function canAdmitThrowEffect(
  liveCount: number,
  eventAgeMs: number,
  targetAvailable: boolean,
): boolean {
  return (
    targetAvailable &&
    Number.isFinite(eventAgeMs) &&
    eventAgeMs >= 0 &&
    eventAgeMs < MAX_FLIGHT_MS &&
    liveCount < MAX_LIVE_THROWS
  );
}

export interface Point {
  x: number;
  y: number;
}

export interface Flight {
  object: ThrowObject;
  duration: number;
  start: Point;
  impact: Point;
  landing: Point;
  impactOffset: Point;
  landingOffset: Point;
  arc: number;
  turns: number;
  /** Peak sideways sway in degrees for objects that do not tumble; 0 otherwise. */
  sway: number;
  impactFraction: number;
  restingRotation: number;
}

export interface Pose extends Point {
  rotation: number;
  /** Squash and pulse. Exactly 1 outside the settling part of a flight. */
  scaleX: number;
  scaleY: number;
  opacity: number;
  phase: 'flight' | 'rest' | 'fade' | 'expired';
}

export interface ActiveEffect {
  event: { target: string };
  flight: Flight;
  startedAt: number;
  skipFlight: boolean;
}

export interface RenderedEffect<T extends ActiveEffect> {
  effect: T;
  pose: Pose;
}

/**
 * Where a thrown object comes to rest, in viewport coordinates. Below a table seat it lands in a
 * band under `bottom`; in a narrow-list row it rests on the row's `middle`, between the end of
 * the name (`left`) and the row's controls (`right`).
 */
export type LandingZone =
  | { left: number; right: number; bottom: number }
  | { left: number; right: number; middle: number };

/** Objects below a table seat rest at least this far beneath it: 12 px impact plus 2 px sink. */
const TABLE_MIN_DEPTH = 14;

/** How far either side of a row's centre an object's centre may rest; 0 when space is tight. */
function rowRoom(zone: LandingZone, object: ThrowObject): number {
  return Math.max(0, (zone.right - zone.left) / 2 - RADIUS[object]);
}

/** The resting point for a landing offset, refitted when the zone's layout has changed. */
export function restingPoint(zone: LandingZone, object: ThrowObject, offset: Point): Point {
  const centre = (zone.left + zone.right) / 2;
  if ('middle' in zone) {
    const room = rowRoom(zone, object);
    return { x: centre + Math.min(room, Math.max(-room, offset.x)), y: zone.middle };
  }
  return { x: centre + offset.x, y: zone.bottom + Math.max(TABLE_MIN_DEPTH, offset.y) };
}

function random(seed: number): () => number {
  let value = seed >>> 0;
  return () => {
    value = (value + 0x6d2b79f5) | 0;
    let mixed = Math.imul(value ^ (value >>> 15), 1 | value);
    mixed = (mixed + Math.imul(mixed ^ (mixed >>> 7), 61 | mixed)) ^ mixed;
    return ((mixed ^ (mixed >>> 14)) >>> 0) / 4294967296;
  };
}

export function createFlight(
  seed: number,
  object: ThrowObject,
  viewport: { width: number; height: number },
  zone: LandingZone,
): Flight {
  const next = random(seed);
  const radius = RADIUS[object];
  const fromLeft = next() < 0.5;
  const inRow = 'middle' in zone;
  // In a row the impact spreads across the free space, and the slide stops at its edges.
  const room = inRow ? rowRoom(zone, object) : Infinity;
  const impactSpan = inRow ? room : Math.min(92, Math.max(58, (zone.right - zone.left) * 0.72));
  const spreadDraw = next();
  const depthDraw = next();
  const impactOffset = { x: (spreadDraw - 0.5) * impactSpan, y: inRow ? 0 : 12 + depthDraw * 17 };
  const slideRange = SLIDE_RANGES[object];
  const slideDistance = slideRange[0] + next() * (slideRange[1] - slideRange[0]);
  const direction = fromLeft ? 1 : -1;
  const sinkDraw = next();
  const landingOffset = {
    x: Math.min(room, Math.max(-room, impactOffset.x + direction * slideDistance)),
    y: inRow ? 0 : impactOffset.y + 2 + sinkDraw * 6,
  };
  const centre = (zone.left + zone.right) / 2;
  const line = 'middle' in zone ? zone.middle : zone.bottom;
  const impact = { x: centre + impactOffset.x, y: line + impactOffset.y };
  const landing = { x: centre + landingOffset.x, y: line + landingOffset.y };
  const spin = TURNS[object];
  const turns = spin === null ? 0 : spin(next());
  const arcScale = ARC_SCALE[object];
  return {
    object,
    duration: MIN_FLIGHT_MS + next() * (MAX_FLIGHT_MS - MIN_FLIGHT_MS),
    start: {
      x: fromLeft ? -radius - 2 : viewport.width + radius + 2,
      y: viewport.height * (0.18 + next() * 0.52),
    },
    impact,
    landing,
    impactOffset,
    landingOffset,
    arc: (70 + next() * Math.min(150, Math.max(80, viewport.height * 0.22))) * arcScale,
    turns,
    impactFraction: 0.72 + next() * 0.1,
    restingRotation:
      object === 'paper-ball'
        ? turns * 360 + direction * slideDistance * 4
        : direction * RESTING_TILT[object](next()),
    // Drawn last, so that the other objects' seeded flights do not depend on it.
    sway: object === 'poo' ? 3 + next() * 3 : 0,
  };
}

export function poseAt(flight: Flight, age: number, reducedMotion = false): Pose {
  const flightAge = reducedMotion ? flight.duration + Math.max(0, age) : Math.max(0, age);
  if (flightAge < flight.duration) {
    const t = flightAge / flight.duration;
    const approaching = t < flight.impactFraction;
    const progress = approaching ? t / flight.impactFraction : 1;
    let x = flight.start.x + (flight.impact.x - flight.start.x) * progress;
    const lineY = flight.start.y + (flight.impact.y - flight.start.y) * progress;
    let y = lineY - Math.sin(Math.PI * progress) * flight.arc;
    let rotation = flight.turns * 360 * t + flight.sway * Math.sin(2 * Math.PI * progress);
    let scaleX = 1;
    let scaleY = 1;

    if (flight.object === 'paper-plane') {
      const dx = flight.impact.x - flight.start.x;
      const dy = flight.impact.y - flight.start.y - Math.PI * flight.arc * Math.cos(Math.PI * progress);
      const tangent = (Math.atan2(dy, dx) * 180) / Math.PI;
      rotation = tangent;
    }

    if (!approaching) {
      const settle = (t - flight.impactFraction) / (1 - flight.impactFraction);
      const eased = 1 - (1 - settle) ** 3;
      const bounce = BOUNCE[flight.object];
      x = flight.impact.x + (flight.landing.x - flight.impact.x) * eased;
      y =
        flight.impact.y +
        (flight.landing.y - flight.impact.y) * eased -
        Math.sin(Math.PI * settle) * bounce * (1 - settle);
      rotation += (flight.restingRotation - rotation) * eased;

      if (flight.object === 'poo') {
        // Squashed flat at impact, then a damped wobble that is gone by the end.
        const wobble = Math.cos(3 * Math.PI * settle) * (1 - settle) ** 2;
        scaleX = 1 + POO_SQUASH_WIDEN * wobble;
        scaleY = 1 - POO_SQUASH_FLATTEN * wobble;
      } else if (flight.object === 'heart') {
        // One pulse: a single swell and return within the settling.
        scaleX = scaleY = 1 + HEART_PULSE * Math.sin(Math.PI * settle);
      }
    }
    return { x, y, rotation, scaleX, scaleY, opacity: 1, phase: 'flight' };
  }

  const settledAge = flightAge - flight.duration;
  const restingRotation = reducedMotion ? 0 : flight.restingRotation;
  if (settledAge < REST_MS) {
    return {
      ...flight.landing,
      rotation: restingRotation,
      scaleX: 1,
      scaleY: 1,
      opacity: 1,
      phase: 'rest',
    };
  }
  const fadeAge = settledAge - REST_MS;
  if (fadeAge < FADE_MS) {
    return {
      ...flight.landing,
      rotation: restingRotation,
      scaleX: 1,
      scaleY: 1,
      opacity: 1 - fadeAge / FADE_MS,
      phase: 'fade',
    };
  }
  return {
    ...flight.landing,
    rotation: 0,
    scaleX: 1,
    scaleY: 1,
    opacity: 0,
    phase: 'expired',
  };
}

/** Stop a moving throw without shortening the rest of its visible lifetime. */
export function settleForReducedMotion<T extends ActiveEffect>(effect: T, clock: number): T {
  if (effect.skipFlight) return effect;
  const elapsed = Math.max(0, clock - effect.startedAt);
  return {
    ...effect,
    skipFlight: true,
    // A flight starts its full rest now; an already settled object keeps its
    // existing rest/fade age while dropping its static rotation.
    startedAt: elapsed < effect.flight.duration ? clock : effect.startedAt + effect.flight.duration,
  };
}

/** Measure each resting target's zone once, then render only the surviving effects. */
export function computeThrowFrame<T extends ActiveEffect>(
  effects: T[],
  clock: number,
  landingZone: (target: string) => LandingZone | null,
): { active: T[]; rendered: RenderedEffect<T>[] } {
  const zones = new Map<string, LandingZone | null>();
  const active: T[] = [];
  const rendered: RenderedEffect<T>[] = [];
  for (const effect of effects) {
    const pose = poseAt(effect.flight, clock - effect.startedAt, effect.skipFlight);
    if (pose.phase === 'expired') continue;
    if (pose.phase !== 'flight') {
      const target = effect.event.target;
      if (!zones.has(target)) zones.set(target, landingZone(target));
      const zone = zones.get(target);
      if (!zone) continue;
      const point = restingPoint(zone, effect.flight.object, effect.flight.landingOffset);
      pose.x = point.x;
      pose.y = point.y;
    }
    active.push(effect);
    rendered.push({ effect, pose });
  }
  return { active, rendered };
}
