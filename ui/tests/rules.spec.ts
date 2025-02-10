import { expect, test } from '@playwright/test';

test.describe('Rules', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: 'Flags' }).click();
  });

  test('can create rule', async ({ page }) => {
    await test.step('create flag', async () => {
      await page.getByRole('button', { name: 'New Flag' }).click();
      await page.getByLabel('Name').fill('Test Rule');
      await page.getByRole('switch', { name: 'Enabled' }).click();
      await page.getByRole('button', { name: 'Create' }).click();
      await page.getByRole('button', { name: 'New Variant' }).click();
      await page
        .getByRole('dialog', { name: 'New Variant' })
        .locator('#key')
        .fill('123');
      await page.getByRole('button', { name: 'Create' }).click();
      await page.getByRole('button', { name: 'New Variant' }).click();
      await page
        .getByRole('dialog', { name: 'New Variant' })
        .locator('#key')
        .fill('456');
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'Update' }).click();
    });

    await test.step('create segment', async () => {
      await page.getByRole('link', { name: 'Segments' }).click();
      await page.getByRole('button', { name: 'New Segment' }).click();
      await page.getByLabel('Name').fill('Test Rule');
      await page.getByLabel('Description').click();
      await page.getByRole('button', { name: 'Create' }).click();
    });

    await test.step('create multi-variate rule', async () => {
      await page.reload();
      await page.getByRole('link', { name: 'Flags' }).click();
      await page.getByRole('link', { name: 'test-rule' }).click();
      await page.getByRole('link', { name: 'Rules' }).click();
      await page.getByRole('button', { name: 'New Rule' }).click();
      await page.locator('#segmentKey-0-select-button').click();
      await page.getByLabel('New Rule').getByText('Test Rule').click();
      await page.getByLabel('Multi-Variate').check();
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });

    await test.step('create single-variant rule', async () => {
      await page.getByRole('button', { name: 'New Rule' }).click();
      await page
        .getByLabel('New Rule')
        .locator('#segmentKey-0-select-button')
        .click();
      await page.getByLabel('New Rule').getByText('Test Rule').click();
      await page.getByLabel('Single Variant').check();
      await page.locator('#variant-select-button').click();
      await page.getByLabel('New Rule').getByText('123').click();
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });

    await test.step('can update multi-variate rule', async () => {
      await page
        .locator('input[name="rollouts\\.\\[0\\]\\.distribution\\.rollout"]')
        .click();
      await page
        .locator('input[name="rollouts\\.\\[0\\]\\.distribution\\.rollout"]')
        .fill('40');
      await page
        .locator('input[name="rollouts\\.\\[1\\]\\.distribution\\.rollout"]')
        .click();
      await page
        .getByRole('listitem')
        .filter({ hasText: 'Multi-Variate' })
        .getByRole('button', { name: 'Done' })
        .click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });

    await test.step('can update single-variant rule', async () => {
      await page.locator('#quick-variant-2-select-button').click();
      await page
        .locator('li')
        .filter({ hasText: 'Single Variant' })
        .locator('li')
        .filter({ hasText: '456' })
        .click();
      await page
        .getByRole('listitem')
        .filter({ hasText: 'Single Variant' })
        .getByRole('button', { name: 'Update' })
        .click();
      await expect(page.getByText('Successfully updated rule')).toBeVisible();
    });
  });

  test('has default rule', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();
    await expect(
      page.getByRole('heading', { name: 'Default Rule' })
    ).toBeVisible();
    await page.locator('#variant-default-select-button').click();
    await page.getByLabel('', { exact: true }).getByText('456').click(); // TODO: should get variant by label
    await page.getByRole('button', { name: 'Update' }).last().click();
    await expect(
      page.getByText('Successfully updated default variant')
    ).toBeVisible();
  });
});
