import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';

// Taking the movement away must not take what the movement carried. A thrown
// object still has to arrive at the seat it was aimed at; it simply must not
// travel there across the screen.

async function room(request: APIRequestContext): Promise<string> {
  const response = await request.post('/api/games', { data: { deck: 't-shirt' } });
  const body = (await response.json()) as { roomId: string };
  return `/g/${encodeURIComponent(body.roomId)}`;
}

async function takeSeat(page: Page, path: string, name: string): Promise<void> {
  await page.goto(path);
  await page.getByRole('textbox', { name: 'Your name' }).fill(name);
  await page.getByRole('button', { name: 'Take a seat' }).click();
  await expect(page.locator('[data-seat-id]').filter({ hasText: name })).toBeVisible();
}

function poseOf(page: Page) {
  return page.locator('.effects g').first().getAttribute('transform');
}

async function bothSeated(browser: Browser, request: APIRequestContext) {
  const path = await room(request);

  // Ada is the one who asked for less movement, and the one thrown at.
  const quiet = await browser.newContext({
    reducedMotion: 'reduce',
    viewport: { width: 1280, height: 860 },
  });
  const lively = await browser.newContext({ viewport: { width: 1280, height: 860 } });

  const ada = await quiet.newPage();
  const grace = await lively.newPage();
  await takeSeat(ada, path, 'Ada');
  await takeSeat(grace, path, 'Grace');
  await expect(ada.locator('[data-seat-id]').filter({ hasText: 'Grace' })).toBeVisible();
  await expect(grace.locator('[data-seat-id]').filter({ hasText: 'Ada' })).toBeVisible();

  // The client paces throws locally: ThrowPacer starts with no allowance and
  // refills at half the message rate, so a throw in the first fraction of a
  // second after seating is dropped before it reaches the socket. A person
  // taking aim is slower than a test, so the test waits for what a person would
  // have spent anyway.
  await grace.waitForTimeout(1000);

  return { ada, grace, quiet, lively };
}

test('a thrown object arrives without travelling', async ({ browser, request }) => {
  const { ada, grace, quiet, lively } = await bothSeated(browser, request);
  try {
    const seat = grace.locator('[data-seat-id]').filter({ hasText: 'Ada' });
    await seat.hover();
    await grace.getByRole('button', { name: 'Throw paper ball at Ada' }).click();

    // It is there, on the page of the person who asked for less movement.
    await expect(ada.locator('.effects g')).toHaveCount(1);

    const first = await poseOf(ada);
    await ada.waitForTimeout(250);
    const second = await poseOf(ada);

    // Same place a quarter of a second later: it settled rather than flew.
    expect(first).not.toBeNull();
    expect(second).toBe(first);

    // And it settled at the seat it was aimed at, not at the edge of the screen
    // where a flight would have begun.
    const target = await ada.locator('[data-seat-id]').filter({ hasText: 'Ada' }).boundingBox();
    const match = /translate\(([-\d.]+) ([-\d.]+)\)/.exec(first ?? '');
    expect(match).not.toBeNull();
    const x = Number(match![1]);
    const y = Number(match![2]);
    // Within about a seat's reach of its middle. The landing point carries a
    // small random offset by design, and a seat is not the same size in every
    // arrangement, so the claim being tested is "at that seat" rather than a
    // particular pixel.
    const reach = target!.width + target!.height;
    const away = Math.hypot(x - (target!.x + target!.width / 2), y - (target!.y + target!.height / 2));
    expect(away).toBeLessThan(reach);

    // And nowhere near an edge, which is where a flight would have begun.
    const viewport = ada.viewportSize()!;
    expect(Math.min(x, y, viewport.width - x, viewport.height - y)).toBeGreaterThan(60);
  } finally {
    await quiet.close();
    await lively.close();
  }
});

test('the same throw does travel when no preference was stated', async ({ browser, request }) => {
  // The other half of the claim: this is a preference being honoured, not the
  // animation having been removed for everybody.
  const { ada, grace, quiet, lively } = await bothSeated(browser, request);
  try {
    const seat = grace.locator('[data-seat-id]').filter({ hasText: 'Ada' });
    await seat.hover();
    await grace.getByRole('button', { name: 'Throw paper ball at Ada' }).click();

    await expect(grace.locator('.effects g')).toHaveCount(1);
    const first = await poseOf(grace);
    await grace.waitForTimeout(120);
    const second = await poseOf(grace);

    expect(first).not.toBeNull();
    expect(second).not.toBe(first);
  } finally {
    await quiet.close();
    await lively.close();
  }
});

test('the interface still shows every state it would have animated', async ({ browser, request }) => {
  const { ada, grace, quiet, lively } = await bothSeated(browser, request);
  try {
    // Playing a card raises it. The rise is the animation; being raised is the
    // information, and that has to survive.
    const played = ada.locator('footer button.card', { hasText: 'M' });
    await played.click();
    await expect(played).toHaveAttribute('aria-pressed', 'true');
    await expect(ada.locator('[data-seat-id]').filter({ hasText: 'Ada' }).locator('.face-down')).toBeVisible();

    await ada.getByRole('button', { name: 'Reveal' }).click();
    await expect(ada.locator('[data-seat-id]').filter({ hasText: 'Ada' }).locator('.face-up')).toHaveText('M');

    await ada.getByRole('button', { name: 'New round' }).click();
    await expect(ada.locator('[data-seat-id]').filter({ hasText: 'Ada' }).locator('.empty')).toBeVisible();
  } finally {
    await quiet.close();
    await lively.close();
  }
});
