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

  test('can copy flag to new namespace', async ({ page }) => {
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('link', { name: 'Namespaces' }).click();
    await page.getByRole('button', { name: 'New Namespace' }).click();
    await page.getByLabel('Name', { exact: true }).fill('copy flag');
    await page.getByLabel('Description').fill('Copy Namespace');
    await page.getByRole('button', { name: 'Create' }).click();
    await page.getByRole('link', { name: 'Flags' }).click();
    await page.getByRole('link', { name: 'test-flag' }).click();

    // perform copy to new namespace
    await page.getByRole('button', { name: 'Actions' }).click();
    await page.getByRole('menuitem', { name: 'Copy to Namespace' }).click();
    await page.locator('#copyToNamespace-select-button').click();
    await page.getByRole('option', { name: 'copy flag', exact: true }).click();
    await page.getByRole('button', { name: 'Copy', exact: true }).click();
    await expect(page.getByText('Successfully copied flag')).toBeVisible();

    // switch to new namespace
    await page.getByRole('link', { name: 'Flags', exact: true }).click();
    await page.getByRole('button', { name: 'Default' }).click();
    await page.getByText('copy flag').click();

    // verify flag was copied
    await page.getByRole('link', { name: 'test-flag' }).click();
    await expect(page.getByText('Test Flag')).toBeVisible();

    // verify variants were copied
    await expect(
      page.getByRole('cell', { name: 'chrome', exact: true })
    ).toBeVisible();
    await expect(
      page.getByRole('cell', { name: 'firefox', exact: true })
    ).toBeVisible();
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
    await expect(page.getByLabel('Description')).toBeDisabled();
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

  test('can not copy flag to new namespace', async ({ page }) => {
    await page.getByRole('link', { name: 'test-flag' }).click();
    await page.getByRole('button', { name: 'Actions' }).click();
    await page.getByRole('menuitem', { name: 'Copy to Namespace' }).click();
    // assert nothing happens
    await expect(
      page.getByRole('menuitem', { name: 'Copy to Namespace' })
    ).toBeHidden();
  });
});
