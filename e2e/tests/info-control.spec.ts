import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';

// The two sentences that used to sit on the join screen now live behind an "i".
// That is only acceptable if the "i" can be reached by every means of operating
// the page, so these tests open it by pointer, by keyboard and by tap, and check
// that a screen reader is told the text belongs to the control it explains.

const STORAGE = 'What is stored, and how to delete it';

async function room(request: APIRequestContext): Promise<string> {
  const response = await request.post('/api/games', { data: { deck: 't-shirt' } });
  const body = (await response.json()) as { roomId: string };
  return `/g/${encodeURIComponent(body.roomId)}`;
}

function bubbleOf(page: Page, trigger: string) {
  return page.locator(`.info:has(button[aria-label="${trigger}"]) .bubble`);
}

test('a pointer discloses the text', async ({ page, request }) => {
  await page.goto(await room(request));

  const bubble = bubbleOf(page, STORAGE);
  await expect(bubble).toHaveCSS('opacity', '0');

  await page.getByRole('button', { name: STORAGE }).hover();
  await expect(bubble).toHaveCSS('opacity', '1');
  await expect(bubble).toContainText('Nothing else is stored');

  // Moving away closes it again.
  await page.mouse.move(2, 2);
  await expect(bubble).toHaveCSS('opacity', '0');
});

test('the keyboard discloses the text, and Escape closes it', async ({ page, request }) => {
  await page.goto(await room(request));

  const trigger = page.getByRole('button', { name: STORAGE });
  const bubble = bubbleOf(page, STORAGE);

  // Tab to it rather than calling focus(), so the browser counts this as
  // keyboard focus — which is the only kind that holds the text open. The
  // storage control follows the checkbox it explains.
  await page.locator('#remember').focus();
  await page.keyboard.press('Tab');
  await expect(trigger).toBeFocused();
  await expect(bubble).toHaveCSS('opacity', '1');

  // Escape closes it while the focus stays where it was.
  await page.keyboard.press('Escape');
  await expect(bubble).toHaveCSS('opacity', '0');
  await expect(trigger).toBeFocused();
});

test('a tap discloses the text where there is no hovering', async ({ browser, request }) => {
  const context = await browser.newContext({
    viewport: { width: 360, height: 720 },
    hasTouch: true,
    isMobile: true,
  });
  try {
    const page = await context.newPage();
    await page.goto(await room(request));

    const bubble = bubbleOf(page, STORAGE);
    await expect(bubble).toHaveCSS('opacity', '0');

    await page.getByRole('button', { name: STORAGE }).tap();
    await expect(bubble).toHaveCSS('opacity', '1');

    // A second tap puts it away again.
    await page.getByRole('button', { name: STORAGE }).tap();
    await expect(bubble).toHaveCSS('opacity', '0');
  } finally {
    await context.close();
  }
});

test('the bubble stays inside the panel it belongs to', async ({ page, request }) => {
  await page.goto(await room(request));
  await page.getByRole('button', { name: STORAGE }).hover();

  const bubble = await bubbleOf(page, STORAGE).boundingBox();
  // Two main elements are on screen here: the room's waiting shell and the join
  // card inside it. The card is the panel the bubble is clamped to.
  const panel = await page.locator('main', { hasText: 'Join the game' }).last().boundingBox();
  expect(bubble).not.toBeNull();
  expect(panel).not.toBeNull();

  expect(bubble!.x).toBeGreaterThanOrEqual(panel!.x);
  expect(bubble!.x + bubble!.width).toBeLessThanOrEqual(panel!.x + panel!.width);
});

test('the explained control carries the description', async ({ page, request }) => {
  await page.goto(await room(request));

  // The checkbox points at the text, so the two are announced together rather
  // than the text being an icon nobody can reach.
  const described = await page.locator('#remember').getAttribute('aria-describedby');
  expect(described).toBe('join-name-storage');
  await expect(page.locator('#join-name-storage')).toContainText('Nothing else is stored');

  const field = await page.locator('#name').getAttribute('aria-describedby');
  expect(field).toBe('join-name-visibility');
  await expect(page.locator('#join-name-visibility')).toContainText('everyone with the link');
});

test('joining and renaming word the same fact the same way', async ({
  browser,
  request,
}: {
  browser: Browser;
  request: APIRequestContext;
}) => {
  const context = await browser.newContext();
  try {
    const page = await context.newPage();
    const path = await room(request);
    await page.goto(path);

    const atJoin = await page.locator('#join-name-visibility').textContent();
    const storageAtJoin = await page.locator('#join-name-storage').textContent();

    await page.getByRole('textbox', { name: 'Your name' }).fill('Ada');
    await page.getByRole('button', { name: 'Take a seat' }).click();
    await page.locator('[data-seat-id]').filter({ hasText: 'Ada' }).waitFor();

    await page.getByRole('button', { name: 'Ada' }).click();
    await expect(page.getByRole('dialog')).toBeVisible();

    expect(await page.locator('#dialog-name-visibility').textContent()).toBe(atJoin);
    expect(await page.locator('#dialog-name-storage').textContent()).toBe(storageAtJoin);
  } finally {
    await context.close();
  }
});

test('Escape does not drag the focus to the control', async ({ page, request }) => {
  // The bubble can be open purely because a pointer rests on the trigger. Escape
  // then belongs to whatever the person is actually doing, not to this control.
  await page.goto(await room(request));

  const field = page.getByRole('textbox', { name: 'Your name' });
  await field.click();
  await page.getByRole('button', { name: STORAGE }).hover();
  await expect(bubbleOf(page, STORAGE)).toHaveCSS('opacity', '1');

  await page.keyboard.press('Escape');
  await expect(bubbleOf(page, STORAGE)).toHaveCSS('opacity', '0');
  await expect(field).toBeFocused();
});
