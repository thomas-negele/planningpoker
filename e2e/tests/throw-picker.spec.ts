import { expect, test, type APIRequestContext, type Browser, type Page } from '@playwright/test';
import { pooOrigin } from '../servers';

const FOUR = ['paper ball', 'paper plane', 'a flower', 'a heart'];
const FIVE = [...FOUR, 'a pile of poo'];

async function expectChoices(picker: ReturnType<Page['locator']>, labels: string[]) {
  const names = await picker.getByRole('button').evaluateAll((buttons) =>
    buttons.map((button) => button.getAttribute('aria-label')),
  );
  expect(names).toEqual(labels.map((label) => `Throw ${label} at Grace`));
}

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

test('mouse hover exposes only four choices and leaves no trigger after departure', async ({ browser, request }) => {
  const { alice, first, second } = await joinTwo(browser, request);
  try {
    const seat = alice.locator('[data-seat-id]').filter({ hasText: 'Grace' });
    const picker = seat.locator('.throw-picker');
    const trigger = alice.getByRole('button', { name: 'Throw something at Grace' });

    await seat.hover();
    await expect(picker).toBeVisible();
    await expect(picker.getByRole('button')).toHaveCount(4);
    await expectChoices(picker, FOUR);
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

test('narrow touch layout opens the same four choices through its trigger', async ({ browser, request }) => {
  const { alice, first, second } = await joinTwo(browser, request, true);
  try {
    const seat = alice.locator('[data-seat-id]').filter({ hasText: 'Grace' });
    const picker = seat.locator('.throw-picker');
    const trigger = alice.getByRole('button', { name: 'Throw something at Grace' });

    await expect(trigger).toBeVisible();
    await expect(picker).toBeHidden();
    await trigger.tap();
    await expect(trigger).toHaveAttribute('aria-expanded', 'true');
    await expect(picker.getByRole('button')).toHaveCount(4);
    await expectChoices(picker, FOUR);
    await expect(picker).toBeVisible();
    await picker.getByRole('button', { name: 'Throw a flower at Grace' }).tap();
    await expect(picker).toBeHidden();
  } finally {
    await first.close();
    await second.close();
  }
});

test.describe('with the pile of poo switched on', () => {
  test.use({ baseURL: pooOrigin });

  test('hover offers all five choices, the new ones last', async ({ browser, request }) => {
    const { alice, first, second } = await joinTwo(browser, request);
    try {
      const seat = alice.locator('[data-seat-id]').filter({ hasText: 'Grace' });
      const picker = seat.locator('.throw-picker');
      await seat.hover();
      await expect(picker).toBeVisible();
      await expectChoices(picker, FIVE);
    } finally {
      await first.close();
      await second.close();
    }
  });

  test('narrow touch layout offers the same five choices', async ({ browser, request }) => {
    const { alice, first, second } = await joinTwo(browser, request, true);
    try {
      const seat = alice.locator('[data-seat-id]').filter({ hasText: 'Grace' });
      const picker = seat.locator('.throw-picker');
      await alice.getByRole('button', { name: 'Throw something at Grace' }).tap();
      await expect(picker).toBeVisible();
      await expectChoices(picker, FIVE);
      const box = await picker.boundingBox();
      expect(box).not.toBeNull();
      expect(box!.x).toBeGreaterThanOrEqual(0);
      expect(box!.x + box!.width).toBeLessThanOrEqual(320);
    } finally {
      await first.close();
      await second.close();
    }
  });

  test('keyboard activation throws a heart and a pile of poo that arrive at the target', async ({ browser, request }) => {
    const { alice, first, second } = await joinTwo(browser, request);
    try {
      // A freshly seated page first saves up message allowance before it throws.
      await alice.waitForTimeout(1200);
      const trigger = alice.getByRole('button', { name: 'Throw something at Grace' });
      for (const label of ['a heart', 'a pile of poo']) {
        await trigger.focus();
        await alice.keyboard.press('Enter');
        await expect(trigger).toHaveAttribute('aria-expanded', 'true');
        const choice = alice.getByRole('button', { name: `Throw ${label} at Grace` });
        await choice.focus();
        await alice.keyboard.press('Enter');
        await expect(trigger).toBeFocused();
        await expect(alice.locator('.effects svg')).toHaveCount(1);
        await expect(alice.locator('.effects svg')).toHaveCount(0, { timeout: 6_000 });
      }
    } finally {
      await first.close();
      await second.close();
    }
  });
});

test('on the narrow list a thrown object comes to rest in its target\'s row', async ({ browser, request }) => {
  const { alice, first, second } = await joinTwo(browser, request);
  try {
    await alice.setViewportSize({ width: 320, height: 640 });
    // A freshly seated page first saves up message allowance before it throws.
    await alice.waitForTimeout(1200);
    const targetRow = alice.locator('[data-seat-id]').filter({ hasText: 'Grace' });
    const ownRow = alice.locator('[data-seat-id].you');
    // An offscreen target deliberately receives no effect, so bring both rows into view.
    await ownRow.scrollIntoViewIfNeeded();
    await targetRow.scrollIntoViewIfNeeded();
    await targetRow.hover();
    await alice.getByRole('button', { name: 'Throw a flower at Grace' }).click();

    // Flight and settling take at most 1.2 s; the object then rests for 2 s.
    const object = alice.locator('.effects svg');
    await expect(object).toHaveCount(1);
    await alice.waitForTimeout(1500);
    const rest = await object.boundingBox();
    const row = await targetRow.boundingBox();
    const own = await ownRow.boundingBox();
    expect(rest && row && own).toBeTruthy();
    const centre = rest!.y + rest!.height / 2;
    expect(centre).toBeGreaterThan(row!.y);
    expect(centre).toBeLessThan(row!.y + row!.height);
    expect(centre < own!.y || centre > own!.y + own!.height).toBe(true);
  } finally {
    await first.close();
    await second.close();
  }
});
