import { expect, test, type APIRequestContext, type Page } from '@playwright/test';

async function roomPath(request: APIRequestContext): Promise<string> {
  const response = await request.post('/api/games', { data: { deck: 't-shirt' } });
  expect(response.ok()).toBe(true);
  const body = (await response.json()) as { roomId: string };
  return `/g/${encodeURIComponent(body.roomId)}`;
}

/** Takes a seat, ticking the visitor box first when asked to. */
async function takeSeat(page: Page, path: string, name: string, visitor = false): Promise<void> {
  await page.goto(path);
  await page.getByRole('textbox', { name: 'Your name' }).fill(name);
  if (visitor) await page.getByRole('checkbox', { name: 'Visitor mode' }).check();
  await page.getByRole('button', { name: 'Take a seat' }).click();
  await expect(page.locator('[data-seat-id]').filter({ hasText: name })).toBeVisible();
}

function seat(page: Page, name: string) {
  return page.locator('[data-seat-id]').filter({ hasText: name });
}

/** One card of the deck along the bottom edge, which a visitor is never offered. */
function deckCard(page: Page, card: string) {
  return page.locator('footer button.card').filter({ hasText: new RegExp(`^${card}$`) });
}

/** Opens one's own name dialog, which is reached by clicking one's own name. */
async function openOwnNameDialog(page: Page, name: string): Promise<void> {
  await seat(page, name).locator('button.name').click();
  await expect(page.getByRole('dialog')).toBeVisible();
}

test('the join form explains visitor mode to a keyboard user and seats them as one', async ({
  page,
  request,
}) => {
  const path = await roomPath(request);
  await page.goto(path);

  const box = page.getByRole('checkbox', { name: 'Visitor mode' });
  await expect(box).not.toBeChecked();

  // The explanation is reachable without a pointer: focusing the control shows it.
  // The closed state is opacity alone rather than visibility, deliberately: a
  // hidden element leaves the accessibility tree, and this note is what the
  // checkbox above points at with aria-describedby. Visibility is therefore the
  // wrong thing to assert here — opacity is what changes.
  const info = page.getByRole('button', { name: 'What visitor mode means' });
  const hint = page.getByText('Visitors cannot vote.');
  await expect(hint).toHaveCSS('opacity', '0');
  await info.focus();
  await expect(hint).toHaveCSS('opacity', '1');

  await page.getByRole('textbox', { name: 'Your name' }).fill('Ada');
  await box.check();
  await page.getByRole('button', { name: 'Take a seat' }).click();

  // A visitor is offered no cards, and their seat says so instead of holding one.
  await expect(seat(page, 'Ada')).toContainText('Visitor');
  await expect(deckCard(page, 'M')).toHaveCount(0);
  // The deck's place carries a short label rather than falling empty.
  await expect(page.locator('footer.visiting')).toHaveText('Visitor mode');
  await expect(page.getByText('No voters this round')).toBeVisible();

  // The round is still theirs to run.
  await expect(page.getByRole('button', { name: 'Reveal' })).toBeEnabled();
  await expect(page.getByRole('button', { name: 'New round' })).toBeEnabled();
});

test('leaving the box unticked seats a voter with the deck to hand', async ({ page, request }) => {
  const path = await roomPath(request);
  await takeSeat(page, path, 'Ada');

  await expect(deckCard(page, 'M')).toBeVisible();
  await expect(seat(page, 'Ada')).not.toContainText('Visitor');
  await expect(page.getByText('Waiting for votes…')).toBeVisible();
});

test('cancelling the name dialog leaves the mode and the vote alone', async ({ page, request }) => {
  const path = await roomPath(request);
  await takeSeat(page, path, 'Ada');

  await deckCard(page, 'M').click();
  await expect(page.getByText('Everyone has voted.')).toBeVisible();

  await openOwnNameDialog(page, 'Ada');
  await page.getByRole('dialog').getByRole('checkbox', { name: 'Visitor mode' }).check();
  await page.getByRole('button', { name: 'Cancel' }).click();

  await expect(page.getByRole('dialog')).toHaveCount(0);
  await expect(seat(page, 'Ada')).not.toContainText('Visitor');
  await expect(page.getByText('Everyone has voted.')).toBeVisible();
  await expect(deckCard(page, 'M')).toBeVisible();
});

test('saving visitor mode during a hidden round drops the vote for both browsers', async ({
  browser,
  request,
}) => {
  const path = await roomPath(request);
  const first = await browser.newContext({ viewport: { width: 1280, height: 800 } });
  const second = await browser.newContext({ viewport: { width: 1280, height: 800 } });

  try {
    const ada = await first.newPage();
    const grace = await second.newPage();
    await takeSeat(ada, path, 'Ada');
    await takeSeat(grace, path, 'Grace');

    await deckCard(ada, 'M').click();
    // The other browser sees a face-down card, never its value.
    await expect(seat(grace, 'Ada').locator('.card.face-down')).toBeVisible();

    await openOwnNameDialog(ada, 'Ada');
    await ada.getByRole('dialog').getByRole('checkbox', { name: 'Visitor mode' }).check();
    await ada.getByRole('button', { name: 'Save' }).click();

    await expect(seat(ada, 'Ada')).toContainText('Visitor');
    await expect(seat(grace, 'Ada')).toContainText('Visitor');
    await expect(seat(grace, 'Ada').locator('.card')).toHaveCount(0);
    await expect(deckCard(ada, 'M')).toHaveCount(0);

    // Switching back does not bring the old card back; the deck shows none played.
    await openOwnNameDialog(ada, 'Ada');
    await ada.getByRole('dialog').getByRole('checkbox', { name: 'Visitor mode' }).uncheck();
    await ada.getByRole('button', { name: 'Save' }).click();

    await expect(seat(ada, 'Ada')).not.toContainText('Visitor');
    await expect(seat(grace, 'Ada').locator('.card.empty')).toBeVisible();
    await expect(deckCard(ada, 'M')).not.toHaveClass(/played/);
  } finally {
    await first.close();
    await second.close();
  }
});

test('a card revealed before the switch stays on the table beside the visitor label', async ({
  page,
  request,
}) => {
  const path = await roomPath(request);
  await takeSeat(page, path, 'Ada');

  await deckCard(page, 'M').click();
  await page.getByRole('button', { name: 'Reveal' }).click();
  await expect(seat(page, 'Ada').locator('.card.face-up')).toHaveText('M');

  await openOwnNameDialog(page, 'Ada');
  await page.getByRole('dialog').getByRole('checkbox', { name: 'Visitor mode' }).check();
  await page.getByRole('button', { name: 'Save' }).click();

  await expect(seat(page, 'Ada').locator('.card.face-up')).toHaveText('M');
  await expect(seat(page, 'Ada')).toContainText('visitor');
  await expect(page.getByRole('button', { name: 'New round' })).toBeEnabled();
});

test('a visitor keeps their seat and mode across a reload, in the narrow layout too', async ({
  browser,
  request,
}) => {
  const path = await roomPath(request);
  const context = await browser.newContext({ viewport: { width: 320, height: 640 } });

  try {
    const ada = await context.newPage();
    await takeSeat(ada, path, 'Ada', true);
    await expect(seat(ada, 'Ada')).toContainText('Visitor');

    await ada.reload();

    // The seat cookie returns the same seat rather than asking for a name again.
    await expect(seat(ada, 'Ada')).toContainText('Visitor');
    await expect(ada.locator('[data-seat-id]')).toHaveCount(1);
    await expect(deckCard(ada, 'M')).toHaveCount(0);
    await expect(ada.getByRole('button', { name: 'Reveal' })).toBeEnabled();
  } finally {
    await context.close();
  }
});
