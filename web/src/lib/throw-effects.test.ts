import assert from 'node:assert/strict';
import test from 'node:test';
import {
  FADE_MS,
  MAX_FLIGHT_MS,
  MAX_LIVE_THROWS,
  MIN_FLIGHT_MS,
  REST_MS,
  canAdmitThrowEffect,
  computeThrowFrame,
  createFlight,
  poseAt,
  restingPoint,
  settleForReducedMotion,
  type ActiveEffect,
} from './throw-effects.ts';

const viewport = { width: 1000, height: 700 };
const target = { left: 440, right: 560, bottom: 420 };

test('seeded flights begin fully offscreen and finish at their seeded resting point', () => {
  for (const object of ['paper-ball', 'paper-plane', 'flowers'] as const) {
    for (const seed of [1, 2, 3, 4, 5, 99]) {
      const flight = createFlight(seed, object, viewport, target);
      const radius = object === 'flowers' ? 25 : object === 'paper-plane' ? 24 : 18;
      assert.ok(flight.start.x < -radius || flight.start.x > viewport.width + radius);
      assert.ok(flight.duration >= MIN_FLIGHT_MS && flight.duration <= MAX_FLIGHT_MS);
      const end = poseAt(flight, flight.duration);
      assert.equal(end.phase, 'rest');
      assert.equal(end.x, flight.landing.x);
      assert.equal(end.y, flight.landing.y);
      assert.ok(Math.abs(flight.impactOffset.x) <= 46);
      assert.ok(Math.abs(flight.landingOffset.x) <= 84);
      assert.ok(flight.landingOffset.y >= 14 && flight.landingOffset.y <= 35);
      for (const fraction of [0, 0.25, 0.5, 0.9, 1]) {
        const pose = poseAt(flight, flight.duration * fraction);
        assert.ok(Number.isFinite(pose.x));
        assert.ok(Number.isFinite(pose.y));
        assert.ok(Number.isFinite(pose.rotation));
      }
    }
  }
});

test('different seeds vary trajectory and both entry sides occur', () => {
  const flights = Array.from({ length: 20 }, (_, seed) =>
    createFlight(seed, 'paper-plane', viewport, target),
  );
  assert.ok(flights.some((flight) => flight.start.x < 0));
  assert.ok(flights.some((flight) => flight.start.x > viewport.width));
  assert.ok(new Set(flights.map((flight) => flight.duration.toFixed(3))).size > 10);
  assert.ok(new Set(flights.map((flight) => flight.impactOffset.x.toFixed(2))).size > 15);
  assert.ok(new Set(flights.map((flight) => flight.landingOffset.x.toFixed(2))).size > 15);
  for (const flight of flights) {
    const pose = poseAt(flight, flight.duration * 0.5);
    const headingX = Math.cos((pose.rotation * Math.PI) / 180);
    assert.equal(Math.sign(headingX), Math.sign(flight.impact.x - flight.start.x));
  }
});

test('impact continues into an object-specific decelerating slide', () => {
  const ball = createFlight(7, 'paper-ball', viewport, target);
  const plane = createFlight(7, 'paper-plane', viewport, target);
  const flowers = createFlight(7, 'flowers', viewport, target);
  const slideLength = (flight: typeof ball) => Math.hypot(
    flight.landing.x - flight.impact.x,
    flight.landing.y - flight.impact.y,
  );

  assert.ok(slideLength(ball) > slideLength(plane));
  assert.ok(slideLength(plane) > slideLength(flowers));
  for (const flight of [ball, plane, flowers]) {
    const impact = poseAt(flight, flight.duration * flight.impactFraction);
    assert.ok(Math.abs(impact.x - flight.impact.x) < 0.000_001);
    assert.ok(Math.abs(impact.y - flight.impact.y) < 0.000_001);
    const halfway = poseAt(
      flight,
      flight.duration * (flight.impactFraction + (1 - flight.impactFraction) / 2),
    );
    assert.equal(
      Math.sign(halfway.x - flight.impact.x),
      Math.sign(flight.landing.x - flight.impact.x),
    );
    assert.ok(Math.abs(halfway.x - flight.impact.x) < Math.abs(flight.landing.x - flight.impact.x));
  }
});

test('objects rest, fade, expire, and reduced motion skips flight', () => {
  const flight = createFlight(42, 'flowers', viewport, target);
  const reduced = poseAt(flight, 0, true);
  assert.equal(reduced.phase, 'rest');
  assert.equal(reduced.rotation, 0);
  assert.deepEqual({ x: reduced.x, y: reduced.y }, flight.landing);
  assert.equal(poseAt(flight, REST_MS + FADE_MS, true).phase, 'expired');
  assert.equal(poseAt(flight, 400, true).phase, 'rest');
  assert.equal(poseAt(flight, flight.duration + REST_MS - 1).phase, 'rest');
  const fading = poseAt(flight, flight.duration + REST_MS + FADE_MS / 2);
  assert.equal(fading.phase, 'fade');
  assert.equal(fading.opacity, 0.5);
  assert.equal(poseAt(flight, flight.duration + REST_MS + FADE_MS).phase, 'expired');
  assert.equal(MAX_LIVE_THROWS, 48);
});

test('effect admission rejects missing targets, stale events, and a full live-object bound', () => {
  assert.equal(canAdmitThrowEffect(MAX_LIVE_THROWS - 1, MAX_FLIGHT_MS - 1, true), true);
  assert.equal(canAdmitThrowEffect(MAX_LIVE_THROWS, 0, true), false);
  assert.equal(canAdmitThrowEffect(0, MAX_FLIGHT_MS, true), false);
  assert.equal(canAdmitThrowEffect(0, -1, true), false);
  assert.equal(canAdmitThrowEffect(0, Number.NaN, true), false);
  assert.equal(canAdmitThrowEffect(0, 0, false), false);
});

test('switching to reduced motion gives moving objects a full rest without reviving settled ones', () => {
  const flight = createFlight(42, 'paper-ball', viewport, target);
  const effect: ActiveEffect = {
    event: { target: 'ada' },
    flight,
    startedAt: 100,
    skipFlight: false,
  };
  const switchedAt = 100 + flight.duration / 2;
  const stopped = settleForReducedMotion(effect, switchedAt);
  assert.equal(stopped.startedAt, switchedAt);
  const firstRest = poseAt(flight, 0, stopped.skipFlight);
  assert.equal(firstRest.phase, 'rest');
  assert.deepEqual({ x: firstRest.x, y: firstRest.y }, flight.landing);
  assert.equal(firstRest.rotation, 0);
  assert.equal(poseAt(flight, REST_MS - 1, stopped.skipFlight).phase, 'rest');
  assert.equal(poseAt(flight, REST_MS, stopped.skipFlight).phase, 'fade');
  assert.equal(poseAt(flight, REST_MS + FADE_MS, stopped.skipFlight).phase, 'expired');
  assert.equal(settleForReducedMotion(stopped, switchedAt + 10), stopped);

  for (const settledAge of [REST_MS / 2, REST_MS + FADE_MS / 2]) {
    const clock = effect.startedAt + flight.duration + settledAge;
    const before = poseAt(flight, clock - effect.startedAt);
    const after = settleForReducedMotion(effect, clock);
    const pose = poseAt(flight, clock - after.startedAt, after.skipFlight);
    assert.equal(pose.phase, before.phase);
    assert.equal(pose.opacity, before.opacity);
    assert.equal(pose.rotation, 0);
    assert.equal(after.startedAt, effect.startedAt + flight.duration);
  }
});

test('one frame measures each resting target once and discards missing or expired effects', () => {
  const flight = createFlight(7, 'flowers', viewport, target);
  const effect = (targetID: string, startedAt: number): ActiveEffect => ({
    event: { target: targetID },
    flight,
    startedAt,
    skipFlight: false,
  });
  const clock = 5000;
  const calls: string[] = [];
  const scene = computeThrowFrame(
    [
      effect('ada', clock - flight.duration - 100),
      effect('ada', clock - flight.duration - 200),
      effect('missing', clock - flight.duration - 100),
      effect('ada', clock - flight.duration - REST_MS - FADE_MS),
      effect('moving', clock - 100),
    ],
    clock,
    (id) => {
      calls.push(id);
      return id === 'ada' ? target : null;
    },
  );
  assert.deepEqual(calls, ['ada', 'missing']);
  assert.equal(scene.active.length, 3);
  assert.equal(scene.rendered.length, 3);
  assert.deepEqual(scene.rendered.map(({ pose }) => pose.phase), ['rest', 'rest', 'flight']);
  for (const { pose } of scene.rendered.slice(0, 2)) {
    assert.equal(pose.x, (target.left + target.right) / 2 + flight.landingOffset.x);
    assert.equal(pose.y, target.bottom + flight.landingOffset.y);
  }
  const afterSleep = computeThrowFrame(scene.active, clock + flight.duration + REST_MS + FADE_MS, () => {
    throw new Error('expired effects should not measure the DOM');
  });
  assert.deepEqual(afterSleep.active, []);
  assert.deepEqual(afterSleep.rendered, []);
});

const ALL_OBJECTS = ['paper-ball', 'paper-plane', 'flowers', 'heart', 'poo'] as const;

// Samples the settling part of a flight, from impact up to the last moment before rest.
function settling(flight: ReturnType<typeof createFlight>, steps = 200) {
  const impactAge = flight.duration * flight.impactFraction;
  return Array.from({ length: steps }, (_, i) =>
    poseAt(flight, impactAge + ((flight.duration - impactAge) * i) / steps),
  );
}

test('existing objects and every settled pose keep their normal size', () => {
  for (const seed of [1, 7, 42, 99]) {
    for (const object of ['paper-ball', 'paper-plane', 'flowers'] as const) {
      const flight = createFlight(seed, object, viewport, target);
      for (let age = 0; age < flight.duration; age += 10) {
        const pose = poseAt(flight, age);
        assert.equal(pose.scaleX, 1, `${object} at ${age}`);
        assert.equal(pose.scaleY, 1, `${object} at ${age}`);
      }
    }
    for (const object of ALL_OBJECTS) {
      const flight = createFlight(seed, object, viewport, target);
      for (const age of [flight.duration, flight.duration + REST_MS + FADE_MS / 2]) {
        const pose = poseAt(flight, age);
        assert.deepEqual([pose.scaleX, pose.scaleY], [1, 1], `${object} ${pose.phase}`);
      }
      for (const age of [0, 300, REST_MS + FADE_MS / 2]) {
        const reduced = poseAt(flight, age, true);
        assert.deepEqual([reduced.scaleX, reduced.scaleY], [1, 1], `${object} reduced`);
        assert.equal(reduced.rotation, 0);
      }
    }
  }
});

test('the new objects begin offscreen and stay within the existing flight bounds', () => {
  for (const object of ['heart', 'poo'] as const) {
    for (const seed of [1, 2, 3, 4, 5, 99]) {
      const flight = createFlight(seed, object, viewport, target);
      assert.ok(flight.start.x < -18 || flight.start.x > viewport.width + 18);
      assert.ok(flight.duration >= MIN_FLIGHT_MS && flight.duration <= MAX_FLIGHT_MS);
      assert.ok(Math.abs(flight.landingOffset.x) <= 84);
      assert.ok(flight.landingOffset.y >= 14 && flight.landingOffset.y <= 35);
      const end = poseAt(flight, flight.duration);
      assert.equal(end.phase, 'rest');
      assert.deepEqual({ x: end.x, y: end.y }, flight.landing);
    }
  }
});

test('the pile of poo squashes on impact, wobbles back and has settled before resting', () => {
  for (const seed of [1, 7, 42, 99, 1234]) {
    const flight = createFlight(seed, 'poo', viewport, target);
    for (let age = 0; age < flight.duration * flight.impactFraction; age += 10) {
      assert.equal(poseAt(flight, age).scaleY, 1, 'no squash before impact');
    }
    const poses = settling(flight);
    assert.ok(poses[0].scaleY < 0.75 && poses[0].scaleX > 1.2, 'flattened and widened at impact');
    // The wobble overshoots at least once: taller than normal for a moment.
    assert.ok(poses.some((pose) => pose.scaleY > 1));
    const last = poses[poses.length - 1];
    assert.ok(Math.abs(last.scaleY - 1) < 0.001 && Math.abs(last.scaleX - 1) < 0.001);
  }
});

test('the pile of poo slides less than the paper ball and sways instead of tumbling', () => {
  const slide = (flight: ReturnType<typeof createFlight>) =>
    Math.abs(flight.landing.x - flight.impact.x);
  for (let seed = 0; seed < 40; seed++) {
    const poo = createFlight(seed, 'poo', viewport, target);
    const ball = createFlight(seed, 'paper-ball', viewport, target);
    assert.ok(slide(poo) < slide(ball), `seed ${seed}`);
    assert.ok(slide(poo) <= 6);
    assert.ok(poo.turns === 0 && poo.sway > 0 && poo.sway <= 6);
  }
  const meanArc = (object: 'poo' | 'paper-ball') =>
    Array.from({ length: 40 }, (_, seed) => createFlight(seed, object, viewport, target).arc)
      .reduce((sum, arc) => sum + arc, 0) / 40;
  assert.ok(meanArc('poo') < meanArc('paper-ball') * 0.7);
});

test('the heart pulses exactly once during settling and rests at its normal size', () => {
  for (const seed of [1, 7, 42, 99, 1234]) {
    const flight = createFlight(seed, 'heart', viewport, target);
    const scales = settling(flight).map((pose) => pose.scaleX);
    for (const pose of settling(flight)) assert.equal(pose.scaleX, pose.scaleY);
    let maxima = 0;
    for (let i = 1; i < scales.length - 1; i++) {
      if (scales[i] > scales[i - 1] && scales[i] >= scales[i + 1]) maxima++;
    }
    assert.equal(maxima, 1);
    assert.ok(Math.max(...scales) > 1.2);
    // The swell fades continuously into the rest, where the size is exactly normal.
    assert.ok(Math.abs(scales[scales.length - 1] - 1) < 0.01);
    assert.equal(poseAt(flight, flight.duration).scaleX, 1);
  }
});

test('repeated hearts and piles of poo still vary path, speed and resting point', () => {
  for (const object of ['heart', 'poo'] as const) {
    const flights = Array.from({ length: 20 }, (_, seed) => createFlight(seed, object, viewport, target));
    assert.ok(flights.some((flight) => flight.start.x < 0));
    assert.ok(flights.some((flight) => flight.start.x > viewport.width));
    assert.ok(new Set(flights.map((flight) => flight.duration.toFixed(3))).size > 10);
    assert.ok(new Set(flights.map((flight) => flight.arc.toFixed(2))).size > 10);
    assert.ok(new Set(flights.map((flight) => flight.landingOffset.x.toFixed(2))).size > 15);
  }
});

// The free space of a narrow-list row: from the end of the name to the row's controls.
const row = { left: 120, right: 250, middle: 300 };

test('in a narrow-list row every object rests inside the free space on the row\'s middle', () => {
  for (const object of ALL_OBJECTS) {
    for (let seed = 0; seed < 200; seed++) {
      const flight = createFlight(seed, object, viewport, row);
      for (const point of [flight.impact, flight.landing]) {
        assert.equal(point.y, row.middle, `${object} seed ${seed}`);
        // No object's drawn half-width is below 18 px, so its whole body stays in the row.
        assert.ok(point.x >= row.left + 18 && point.x <= row.right - 18, `${object} seed ${seed}`);
      }
      const rest = poseAt(flight, flight.duration);
      assert.deepEqual({ x: rest.x, y: rest.y }, flight.landing);
    }
  }
});

test('a row narrower than the object centres it without spreading', () => {
  const tight = { left: 200, right: 230, middle: 300 };
  for (const object of ALL_OBJECTS) {
    for (const seed of [1, 7, 42, 99]) {
      const flight = createFlight(seed, object, viewport, tight);
      assert.deepEqual(flight.landing, { x: 215, y: 300 }, `${object} seed ${seed}`);
    }
  }
});

test('a resting object is refitted when the layout switches between table and list', () => {
  const seat = { left: 440, right: 560, bottom: 420 };
  const fromTable = createFlight(7, 'paper-ball', viewport, seat);
  const inRow = restingPoint(row, 'paper-ball', fromTable.landingOffset);
  assert.equal(inRow.y, row.middle);
  assert.ok(inRow.x >= row.left + 18 && inRow.x <= row.right - 18);

  const fromRow = createFlight(7, 'paper-ball', viewport, row);
  const belowSeat = restingPoint(seat, 'paper-ball', fromRow.landingOffset);
  assert.ok(belowSeat.y >= seat.bottom + 14, 'back below the seat, not on its edge');

  // Within one layout the resting point is exactly where the flight landed.
  assert.deepEqual(restingPoint(seat, 'paper-ball', fromTable.landingOffset), fromTable.landing);
  assert.deepEqual(restingPoint(row, 'paper-ball', fromRow.landingOffset), fromRow.landing);
});
