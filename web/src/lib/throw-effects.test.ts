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
