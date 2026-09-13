import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';

async function roomPath(request: APIRequestContext): Promise<string> {
  const response = await request.post('/api/games', { data: { deck: 't-shirt' } });
  expect(response.ok()).toBe(true);
  const body = (await response.json()) as { roomId: string };
  return `/g/${encodeURIComponent(body.roomId)}`;
}

async function takeSeat(page: Page, path: string, name: string): Promise<void> {
  await page.goto(path);
  await page.getByRole('textbox', { name: 'Your name' }).fill(name);
  await page.getByRole('button', { name: 'Take a seat' }).click();
  await expect(page.locator('[data-seat-id]').filter({ hasText: name })).toBeVisible();
}

async function joinTwo(browser: Browser, request: APIRequestContext, touch = false) {
  const path = await roomPath(request);
  const first = await browser.newContext({
    viewport: touch ? { width: 320, height: 480 } : { width: 1280, height: 800 },
    hasTouch: touch,
    isMobile: touch,
  });
  const second = await browser.newContext();
  try {
    const alice = await first.newPage();
    const grace = await second.newPage();
    await takeSeat(alice, path, 'Ada');
    await takeSeat(grace, path, 'Grace');
    await expect(alice.locator('[data-seat-id]').filter({ hasText: 'Grace' })).toBeVisible();
    return { alice, first, second };
  } catch (error) {
    await first.close();
    await second.close();
    throw error;
  }
}

test('mouse hover exposes only three choices and leaves no trigger after departure', async ({ browser, request }) => {
  const { alice, first, second } = await joinTwo(browser, request);
  try {
    const seat = alice.locator('[data-seat-id]').filter({ hasText: 'Grace' });
    const picker = seat.locator('.throw-picker');
    const trigger = alice.getByRole('button', { name: 'Throw something at Grace' });

    await seat.hover();
    await expect(picker).toBeVisible();
    await expect(picker.getByRole('button')).toHaveCount(3);
    await expect(trigger).toHaveCSS('opacity', '0');

    await picker.getByRole('button', { name: 'Throw paper plane at Grace' }).hover();
    await expect(picker).toBeVisible();
    await alice.mouse.move(2, 2);
    await expect(picker).toBeHidden();
    await expect(trigger).toHaveCSS('opacity', '0');
  } finally {
    await first.close();
    await second.close();
  }
});

test('narrow touch layout opens the same three choices through its trigger', async ({ browser, request }) => {
  const { alice, first, second } = await joinTwo(browser, request, true);
  try {
    const seat = alice.locator('[data-seat-id]').filter({ hasText: 'Grace' });
    const picker = seat.locator('.throw-picker');
    const trigger = alice.getByRole('button', { name: 'Throw something at Grace' });

    await expect(trigger).toBeVisible();
    await expect(picker).toBeHidden();
    await trigger.tap();
    await expect(trigger).toHaveAttribute('aria-expanded', 'true');
    await expect(picker.getByRole('button')).toHaveCount(3);
    await expect(picker).toBeVisible();
    await picker.getByRole('button', { name: 'Throw a flower at Grace' }).tap();
    await expect(picker).toBeHidden();
  } finally {
    await first.close();
    await second.close();
  }
});
