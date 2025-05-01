import { expect, test } from '@playwright/test';

test.describe('Namespaces', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.getByRole('button', { name: 'Continue to Dashboard' }).click();
  });

  test('can create namespace', async ({ page }) => {
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('link', { name: 'Namespaces' }).click();
    await expect(
      page.getByRole('heading', { name: 'Namespaces' })
    ).toBeVisible();
    await page.getByRole('button', { name: 'New Namespace' }).click();
    await page.getByLabel('Name', { exact: true }).fill('staging');
    await page.getByLabel('Description').fill('Staging Namespace');
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(
      page.getByText('Successfully created namespace')
    ).toBeVisible();
  });

  test('can switch to newly created namespace', async ({ page }) => {
    await page.getByRole('link', { name: 'logo' }).click();
    await page
      .getByTestId('namespace-listbox')
      .getByRole('combobox', { name: 'Default' })
      .click();
    await page.getByText('staging').click();
    await expect(page.getByRole('heading', { name: 'Flags' })).toBeVisible();
  });

  test('can update namespace', async ({ page }) => {
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('link', { name: 'Namespaces' }).click();
    await expect(
      page.getByRole('heading', { name: 'Namespaces' })
    ).toBeVisible();
    await page.getByRole('link', { name: 'staging', exact: true }).click();
    await page.getByLabel('Name', { exact: true }).fill('test');
    await page.getByLabel('Description').fill('Test Namespace');
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(
      page.getByText('Successfully updated namespace')
    ).toBeVisible();
  });

  test('deleting current namespace switches to default namespace', async ({
    page
  }) => {
    await page.getByRole('link', { name: 'logo' }).click();
    await page
      .getByTestId('namespace-listbox')
      .getByRole('combobox', { name: 'Default' })
      .click();
    await page.getByText('test', { exact: true }).click();
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('link', { name: 'Namespaces' }).click();
    await expect(
      page.getByRole('heading', { name: 'Namespaces' })
    ).toBeVisible();
    await page.getByRole('link', { name: 'Delete , test' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await expect(
      page.getByTestId('namespace-listbox').getByRole('combobox', {
        name: 'Default'
      })
    ).toBeVisible();
  });

  test('cannot delete default namespace', async ({ page }) => {
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('link', { name: 'Namespaces' }).click();
    await expect(
      page.getByRole('heading', { name: 'Namespaces' })
    ).toBeVisible();
    await page.getByText('Delete, default').click();
    // assert that the default namespace is still there even after clicking 'delete'
    await page.getByRole('cell', { name: 'default' }).first().click();
  });
});
