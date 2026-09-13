import assert from 'node:assert/strict';
import test from 'node:test';
import { ThrowPacer } from './throw-pacing.ts';

const policy = {
  participantPerSecond: 3,
  roomPerSecond: 12,
  messagePerSecond: 10,
  messageBurst: 20,
};

test('throws are bounded by a rolling participant window without deferred work', () => {
  let now = 0;
  const pacer = new ThrowPacer(() => now);
  pacer.configure({ ...policy, messagePerSecond: 100 });
  now = 100;

  assert.equal(pacer.takeThrow(0), true);
  assert.equal(pacer.takeThrow(0), true);
  assert.equal(pacer.takeThrow(0), true);
  assert.equal(pacer.takeThrow(0), false);
  now = 999;
  assert.equal(pacer.takeThrow(0), false);
  now = 1100;
  assert.equal(pacer.takeThrow(0), true);
});

test('outgoing backlog and missing fresh policy suppress throws', () => {
  let now = 0;
  const pacer = new ThrowPacer(() => now);
  assert.equal(pacer.takeThrow(0), false);
  pacer.configure(policy);
  now = 1000;
  assert.equal(pacer.takeThrow(12), false);
  pacer.resetSocket();
  assert.equal(pacer.takeThrow(0), false);
});

test('reconnect starts empty while retaining the same-seat rolling history', () => {
  let now = 0;
  const pacer = new ThrowPacer(() => now);
  pacer.configure({ ...policy, messagePerSecond: 100 });
  now = 100;
  assert.equal(pacer.takeThrow(0), true);
  assert.equal(pacer.takeThrow(0), true);
  assert.equal(pacer.takeThrow(0), true);

  pacer.resetSocket();
  pacer.configure({ ...policy, messagePerSecond: 100 });
  assert.equal(pacer.takeThrow(0), false, 'the fresh socket budget starts empty');
  now = 999;
  assert.equal(pacer.takeThrow(0), false, 'reconnect does not reset the participant window');
  now = 1100;
  assert.equal(pacer.takeThrow(0), true);
});

test('the default rate accounts for ordinary intents and never queues a click', () => {
  let now = 0;
  const pacer = new ThrowPacer(() => now);
  pacer.configure(policy);
  now = 400;
  assert.equal(pacer.takeThrow(0), true);
  assert.equal(pacer.takeThrow(0), false);

  pacer.noteIntent();
  assert.equal(pacer.takeThrow(0), false);
  now = 800;
  assert.equal(pacer.takeThrow(0), true, 'only this new click is considered after refill');
});

test('low message rates reserve capacity for immediate game actions', () => {
  let now = 0;
  const pacer = new ThrowPacer(() => now);
  pacer.configure({ ...policy, messagePerSecond: 1, messageBurst: 2 });
  now = 4000;
  assert.equal(pacer.takeThrow(0), true);
  pacer.noteIntent();
  assert.equal(pacer.takeThrow(0), false);
  now = 6000;
  assert.equal(pacer.takeThrow(0), false);
  now = 8000;
  assert.equal(pacer.takeThrow(0), true);
});
