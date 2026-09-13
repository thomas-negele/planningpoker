import type { ThrowObject } from './protocol';

export const MIN_FLIGHT_MS = 600;
export const MAX_FLIGHT_MS = 1200;
export const REST_MS = 2000;
export const FADE_MS = 600;
export const MAX_LIVE_THROWS = 48;

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
  impactFraction: number;
  restingRotation: number;
}

export interface Pose extends Point {
  rotation: number;
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

interface TargetRect {
  left: number;
  right: number;
  bottom: number;
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
  target: { left: number; right: number; bottom: number },
): Flight {
  const next = random(seed);
  const radius = object === 'flowers' ? 25 : object === 'paper-plane' ? 24 : 18;
  const fromLeft = next() < 0.5;
  const impactSpan = Math.min(92, Math.max(58, (target.right - target.left) * 0.72));
  const impactOffset = { x: (next() - 0.5) * impactSpan, y: 12 + next() * 17 };
  const impact = {
    x: (target.left + target.right) / 2 + impactOffset.x,
    y: target.bottom + impactOffset.y,
  };
  const slideRange = object === 'paper-ball' ? [20, 38] : object === 'paper-plane' ? [14, 30] : [8, 20];
  const slideDistance = slideRange[0] + next() * (slideRange[1] - slideRange[0]);
  const direction = fromLeft ? 1 : -1;
  const landingOffset = {
    x: impactOffset.x + direction * slideDistance,
    y: impactOffset.y + 2 + next() * 6,
  };
  const landing = {
    x: (target.left + target.right) / 2 + landingOffset.x,
    y: target.bottom + landingOffset.y,
  };
  const turns = object === 'paper-ball' ? 1.8 + next() * 2.2 : object === 'flowers' ? 0.3 + next() * 0.55 : 0;
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
    arc: 70 + next() * Math.min(150, Math.max(80, viewport.height * 0.22)),
    turns,
    impactFraction: 0.72 + next() * 0.1,
    restingRotation:
      object === 'paper-ball'
        ? turns * 360 + direction * slideDistance * 4
        : direction * (object === 'paper-plane' ? 5 + next() * 7 : 7 + next() * 10),
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
    let rotation = flight.turns * 360 * t;

    if (flight.object === 'paper-plane') {
      const dx = flight.impact.x - flight.start.x;
      const dy = flight.impact.y - flight.start.y - Math.PI * flight.arc * Math.cos(Math.PI * progress);
      const tangent = (Math.atan2(dy, dx) * 180) / Math.PI;
      rotation = tangent;
    }

    if (!approaching) {
      const settle = (t - flight.impactFraction) / (1 - flight.impactFraction);
      const eased = 1 - (1 - settle) ** 3;
      const bounce = flight.object === 'paper-ball' ? 9 : flight.object === 'flowers' ? 3 : 1.5;
      x = flight.impact.x + (flight.landing.x - flight.impact.x) * eased;
      y =
        flight.impact.y +
        (flight.landing.y - flight.impact.y) * eased -
        Math.sin(Math.PI * settle) * bounce * (1 - settle);
      rotation += (flight.restingRotation - rotation) * eased;
    }
    return { x, y, rotation, opacity: 1, phase: 'flight' };
  }

  const settledAge = flightAge - flight.duration;
  const restingRotation = reducedMotion ? 0 : flight.restingRotation;
  if (settledAge < REST_MS) {
    return { ...flight.landing, rotation: restingRotation, opacity: 1, phase: 'rest' };
  }
  const fadeAge = settledAge - REST_MS;
  if (fadeAge < FADE_MS) {
    return {
      ...flight.landing,
      rotation: restingRotation,
      opacity: 1 - fadeAge / FADE_MS,
      phase: 'fade',
    };
  }
  return { ...flight.landing, rotation: 0, opacity: 0, phase: 'expired' };
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

/** Measure each resting target once, then render only the surviving effects. */
export function computeThrowFrame<T extends ActiveEffect>(
  effects: T[],
  clock: number,
  targetRect: (target: string) => TargetRect | null,
): { active: T[]; rendered: RenderedEffect<T>[] } {
  const targets = new Map<string, TargetRect | null>();
  const active: T[] = [];
  const rendered: RenderedEffect<T>[] = [];
  for (const effect of effects) {
    const pose = poseAt(effect.flight, clock - effect.startedAt, effect.skipFlight);
    if (pose.phase === 'expired') continue;
    if (pose.phase !== 'flight') {
      const target = effect.event.target;
      if (!targets.has(target)) targets.set(target, targetRect(target));
      const rect = targets.get(target);
      if (!rect) continue;
      pose.x = (rect.left + rect.right) / 2 + effect.flight.landingOffset.x;
      pose.y = rect.bottom + effect.flight.landingOffset.y;
    }
    active.push(effect);
    rendered.push({ effect, pose });
  }
  return { active, rendered };
}
