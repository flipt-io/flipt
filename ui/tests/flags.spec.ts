import { expect, test } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/');
  await page.getByRole('link', { name: 'Flags' }).click();
});

test('can create flag', async ({ page }) => {
  await page.getByRole('button', { name: 'New Flag' }).click();
  await page.getByLabel('Name').fill('Test Flag');
  await page.getByLabel('Description').click();
  await page.getByRole('button', { name: 'Create' }).click();
  await expect(page.getByText('Successfully created flag')).toBeVisible();
});

test('can update flag', async ({ page }) => {
  await page.getByRole('link', { name: 'test-flag' }).click();
  await page.getByLabel('Description').click();
  await page.getByLabel('Description').fill('Test flag description');
  await page.getByRole('button', { name: 'Update' }).click();
  await expect(page.getByText('Successfully updated flag')).toBeVisible();
});

test('can add variants to flag', async ({ page }) => {
  await page.getByRole('link', { name: 'test-flag' }).click();

  await test.step('add variant', async () => {
    await page.getByRole('button', { name: 'New Variant' }).click();
    await page
      .getByRole('dialog', { name: 'New Variant' })
      .locator('#key')
      .click();
    await page
      .getByRole('dialog', { name: 'New Variant' })
      .locator('#key')
      .fill('chrome');
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Successfully created variant')).toBeVisible();
  });

  await test.step('add another variant', async () => {
    await page.getByRole('button', { name: 'New Variant' }).click();
    await page
      .getByRole('dialog', { name: 'New Variant' })
      .locator('#key')
      .click();
    await page
      .getByRole('dialog', { name: 'New Variant' })
      .locator('#key')
      .fill('firefox');
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Successfully created variant')).toBeVisible();
  });

  await test.step('edit variant description', async () => {
    await page.getByRole('link', { name: 'Edit ,chrome' }).click();
    await page
      .getByRole('dialog', { name: 'Edit Variant' })
      .locator('#description')
      .click();
    await page
      .getByRole('dialog', { name: 'Edit Variant' })
      .locator('#description')
      .fill('chrome browser');
    await page
      .getByRole('dialog', { name: 'Edit Variant' })
      .getByRole('button', { name: 'Update' })
      .click();
    await expect(page.getByText('Successfully updated variant')).toBeVisible();
  });

  await test.step('edit other variant description', async () => {
    await page.getByRole('link', { name: 'Edit ,firefox' }).click();
    await page
      .getByRole('dialog', { name: 'Edit Variant' })
      .locator('#description')
      .click();
    await page
      .getByRole('dialog', { name: 'Edit Variant' })
      .locator('#description')
      .fill('firefox browser');
    await page
      .getByRole('dialog', { name: 'Edit Variant' })
      .getByRole('button', { name: 'Update' })
      .click();
    await expect(page.getByText('Successfully updated variant')).toBeVisible();
  });
});
