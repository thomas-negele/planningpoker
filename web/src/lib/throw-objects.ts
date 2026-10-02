/** Every object this page can draw, in picker order. The wire type derives from it. */
export const KNOWN_THROW_OBJECTS = [
  'paper-ball',
  'paper-plane',
  'flowers',
  'heart',
  'poo',
] as const;

export type ThrowObject = (typeof KNOWN_THROW_OBJECTS)[number];

export function isThrowObject(value: unknown): value is ThrowObject {
  return (KNOWN_THROW_OBJECTS as readonly unknown[]).includes(value);
}

/**
 * The server's list reduced to objects this page can draw, in picker order.
 * Null when nothing usable is listed, which leaves throwing unavailable.
 */
export function offeredThrowObjects(listed: unknown): ThrowObject[] | null {
  if (!Array.isArray(listed)) return null;
  const offered = KNOWN_THROW_OBJECTS.filter((object) => listed.includes(object));
  return offered.length > 0 ? offered : null;
}

/** Only cheerful colours; no black, white, grey, brown or muted hearts. */
export const HEART_COLOURS = [
  'red',
  'pink',
  'orange',
  'yellow',
  'green',
  'blue',
  'purple',
] as const;

export type HeartColour = (typeof HEART_COLOURS)[number];

/** Every client derives the same colour from the shared seed. Seed 0, the picker's, is red. */
export function heartColour(seed: number): HeartColour {
  return HEART_COLOURS[(seed >>> 0) % HEART_COLOURS.length];
}
