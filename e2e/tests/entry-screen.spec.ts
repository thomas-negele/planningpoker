import { expect, test, type Page } from '@playwright/test';

// The entry screen names each deck's cards, and takes them from the server so
// that what it promises is what the table deals. These tests are about that
// promise and about what happens when the server does not answer.

function option(page: Page, label: string) {
  return page.getByRole('radio', { name: new RegExp(label) });
}

// The radio itself is visually hidden and takes no pointer events; what a person
// clicks is the row around it, so that is what these tests click too.
function choose(page: Page, label: string) {
  return page.locator('label', { hasText: label }).click();
}

test('each deck is listed with its cards, comma separated', async ({ page }) => {
  await page.goto('/');

  const tshirt = page.locator('li', { hasText: 'T-shirt sizes' });
  const fibonacci = page.locator('li', { hasText: 'Fibonacci' });

  await expect(tshirt.locator('.cards')).toHaveText('XS, S, M, L, XL, ?, ☕');
  await expect(fibonacci.locator('.cards')).toHaveText('0, ½, 1, 2, 3, 5, 8, 13, 21, ?, ☕');
});

test('the cards listed are the cards the table deals', async ({ page }) => {
  await page.goto('/');

  const listed = await page.locator('li', { hasText: 'Fibonacci' }).locator('.cards').textContent();
  const wanted = (listed ?? '').split(',').map((card) => card.trim());

  await choose(page, 'Fibonacci');
  await page.getByRole('button', { name: 'Start a new game' }).click();
  await page.getByRole('textbox', { name: 'Your name' }).fill('Ada');
  await page.getByRole('button', { name: 'Take a seat' }).click();

  const dealt = page.locator('footer button.card');
  await expect(dealt).toHaveCount(wanted.length);
  expect(await dealt.allInnerTexts()).toEqual(wanted);
});

test('a game can still be started when the cards cannot be fetched', async ({ page }) => {
  // Not knowing what is in a deck is a smaller failure than not being able to
  // start one, so the screen keeps working without the values.
  await page.route('**/api/decks', (route) => route.abort('failed'));
  await page.goto('/');

  await expect(option(page, 'T-shirt sizes')).toBeVisible();
  await expect(option(page, 'Fibonacci')).toBeVisible();
  await expect(page.locator('.cards')).toHaveCount(0);

  await choose(page, 'Fibonacci');
  await expect(option(page, 'Fibonacci')).toBeChecked();
  await page.getByRole('button', { name: 'Start a new game' }).click();

  await expect(page).toHaveURL(/\/g\//);
  await expect(page.getByRole('button', { name: 'Take a seat' })).toBeVisible();
});

test('the deck can be chosen and the game started with the keyboard alone', async ({ page }) => {
  await page.goto('/');

  // Reach the group, move within it the way a radio group is moved, and start.
  await page.getByRole('radio', { name: /T-shirt sizes/ }).focus();
  await page.keyboard.press('ArrowDown');
  await expect(option(page, 'Fibonacci')).toBeChecked();

  await page.keyboard.press('Shift+Tab');
  await expect(page.getByRole('button', { name: 'Start a new game' })).toBeFocused();
  await page.keyboard.press('Enter');

  await page.getByRole('textbox', { name: 'Your name' }).fill('Grace');
  await page.getByRole('button', { name: 'Take a seat' }).click();

  // The Fibonacci deck is the one that arrived, chosen without a pointer.
  await expect(page.locator('footer button.card').first()).toHaveText('0');
});
