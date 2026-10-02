import assert from 'node:assert/strict';
import test from 'node:test';
import { HEART_COLOURS, heartColour, isThrowObject, offeredThrowObjects } from './throw-objects.ts';

test('both new objects are known, anything else is not', () => {
  for (const object of ['paper-ball', 'paper-plane', 'flowers', 'heart', 'poo']) {
    assert.equal(isThrowObject(object), true, object);
  }
  for (const object of ['client-html', 'Heart', '', null, undefined, 3]) {
    assert.equal(isThrowObject(object), false, String(object));
  }
});

test('the picker offers what the server lists, in picker order', () => {
  assert.deepEqual(offeredThrowObjects(['paper-ball', 'paper-plane', 'flowers', 'heart']), [
    'paper-ball',
    'paper-plane',
    'flowers',
    'heart',
  ]);
  assert.deepEqual(offeredThrowObjects(['poo', 'heart', 'flowers', 'paper-plane', 'paper-ball']), [
    'paper-ball',
    'paper-plane',
    'flowers',
    'heart',
    'poo',
  ]);
});

test('objects this page cannot draw are left out', () => {
  assert.deepEqual(offeredThrowObjects(['heart', 'confetti']), ['heart']);
});

test('a policy without a usable list offers nothing, so throwing stays off', () => {
  for (const listed of [undefined, null, 'heart', {}, [], ['confetti']]) {
    assert.equal(offeredThrowObjects(listed), null, JSON.stringify(listed));
  }
});

test('heart seeds reach all seven cheerful colours and nothing else', () => {
  const seen = new Set<string>();
  for (let seed = 0; seed < 7; seed++) seen.add(heartColour(seed));
  assert.deepEqual([...seen].sort(), [...HEART_COLOURS].sort());

  for (const seed of [7, 1000, 123_456_789, 2 ** 31, 2 ** 32 - 1]) {
    assert.ok((HEART_COLOURS as readonly string[]).includes(heartColour(seed)), String(seed));
  }
  for (const excluded of ['black', 'white', 'grey', 'gray', 'brown']) {
    assert.ok(!(HEART_COLOURS as readonly string[]).includes(excluded), excluded);
  }
});

test('the same seed always yields the same heart, and the picker\'s is red', () => {
  for (const seed of [0, 3, 41, 2 ** 32 - 1]) assert.equal(heartColour(seed), heartColour(seed));
  assert.equal(heartColour(0), 'red');
});
