import { expect, test } from '@playwright/test';

test.describe('Flags', () => {
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
      await expect(
        page.getByText('Successfully created variant')
      ).toBeVisible();
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
      await expect(
        page.getByText('Successfully created variant')
      ).toBeVisible();
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
      await expect(
        page.getByText('Successfully updated variant')
      ).toBeVisible();
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
      await expect(
        page.getByText('Successfully updated variant')
      ).toBeVisible();
    });
  });
});

test.describe('Flags - Read Only', () => {
  test.beforeEach(async ({ page }) => {
    await page.route(/\/meta\/config/, async (route) => {
      const response = await route.fetch();
      const json = await response.json();
      json.storage = { type: 'git' };
      // Fulfill using the original response, while patching the
      // response body with our changes to mock git storage for read only mode
      await route.fulfill({ response, json });
    });

    await page.goto('/');
    await page.getByRole('link', { name: 'Flags' }).click();
  });

  test('can not create flag', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'New Flag' })).toBeDisabled();
  });

  test('can not update flag', async ({ page }) => {
    await page.getByRole('link', { name: 'test-flag' }).click();
    await page.getByLabel('Description').click();
    await page.getByLabel('Description').fill('Test flag description 2');
    await expect(page.getByRole('switch', { name: 'Enabled' })).toBeDisabled();
    await expect(page.getByRole('button', { name: 'Update' })).toBeDisabled();
  });

  test('can not add variants to flag', async ({ page }) => {
    await page.getByRole('link', { name: 'test-flag' }).click();

    await expect(
      page.getByRole('button', { name: 'New Variant' })
    ).toBeDisabled();
  });

  test('can not edit variant description', async ({ page }) => {
    await page.getByRole('link', { name: 'test-flag' }).click();
    await expect(page.getByRole('link', { name: 'Edit ,chrome' })).toBeHidden();
  });

  test('can not delete flag', async ({ page }) => {
    await page.getByRole('link', { name: 'test-flag' }).click();
    await page.getByRole('button', { name: 'Actions' }).click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();
    // assert nothing happens
    await expect(page.getByRole('menuitem', { name: 'Delete' })).toBeHidden();
  });
});
