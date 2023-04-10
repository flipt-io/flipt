import { expect, test } from '@playwright/test';

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
    await page.getByRole('button', { name: 'Create' }).click();
  });

  await test.step('create segment', async () => {
    await page.getByRole('link', { name: 'Segments' }).click();
    await page.getByRole('button', { name: 'New Segment' }).click();
    await page.getByLabel('Name').fill('Test Rule');
    await page.getByLabel('Description').click();
    await page.getByRole('button', { name: 'Create' }).click();
  });

  await test.step('create rule', async () => {
    await page.reload();
    await page.getByRole('link', { name: 'Flags' }).click();
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Evaluation' }).click();
    await page.getByRole('button', { name: 'New Rule' }).click();

    await page.locator('#segmentKey-select-button').click();
    await page.getByText('test-rule').click();
    await page.getByLabel('Multi-Variant').check();
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Successfully created rule')).toBeVisible();
  });
});

test('can update rule', async ({ page }) => {
  await page.getByRole('link', { name: 'test-rule' }).click();
  await page.getByRole('link', { name: 'Evaluation' }).click();
  await page.getByRole('link', { name: 'Edit |' }).click();
  await page
    .getByRole('dialog', { name: 'Edit Rule' })
    .getByText('123')
    .click();

  await page.locator('input[name="\\31 23"]').click();
  await page.locator('input[name="\\31 23"]').fill('40');
  await page.locator('input[name="\\31 23"]').press('Tab');
  await page.locator('input[name="\\34 56"]').fill('60');
  await page.getByRole('button', { name: 'Update' }).click();
  await page.getByText('Successfully updated rule').click();
});
