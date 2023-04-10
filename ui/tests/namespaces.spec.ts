import { expect, test } from '@playwright/test';

test.beforeEach(async ({ page }) => {
  await page.goto('/');
  await page.reload();
});

test('can create namespace', async ({ page }) => {
  await page.getByRole('link', { name: 'Settings' }).click();
  await expect(page.getByRole('heading', { name: 'Namespaces' })).toBeVisible();
  await page.getByRole('button', { name: 'New Namespace' }).click();
  await page.getByLabel('Name', { exact: true }).fill('staging');
  await page.getByLabel('Description').fill('Staging Namespace');
  await page.getByRole('button', { name: 'Create' }).click();
  await expect(page.getByText('Successfully created namespace')).toBeVisible();
});

test('can switch to newly created namespace', async ({ page }) => {
  await page.getByRole('link', { name: 'logo' }).click();
  await page.getByRole('button', { name: 'Default' }).click();
  await page.getByText('staging').click();
  await expect(page.getByRole('heading', { name: 'Flags' })).toBeVisible();
});

test('can update namespace', async ({ page }) => {
  await page.getByRole('link', { name: 'Settings' }).click();
  await expect(page.getByRole('heading', { name: 'Namespaces' })).toBeVisible();
  await page.getByRole('link', { name: 'staging', exact: true }).click();
  await page.getByLabel('Name', { exact: true }).fill('test');
  await page.getByLabel('Description').fill('Test Namespace');
  await page.getByRole('button', { name: 'Update' }).click();
  await expect(page.getByText('Successfully updated namespace')).toBeVisible();
});

test('can delete namespace', async ({ page }) => {
  await page.getByRole('link', { name: 'Settings' }).click();
  await expect(page.getByRole('heading', { name: 'Namespaces' })).toBeVisible();
  await page.getByRole('link', { name: 'Delete , test' }).click();
  await page.getByRole('button', { name: 'Delete' }).click();
});

test('cannot delete default namespace', async ({ page }) => {
  await page.getByRole('link', { name: 'Settings' }).click();
  await expect(page.getByRole('heading', { name: 'Namespaces' })).toBeVisible();
  await page.getByText('Delete, default').click();
  // assert that the default namespace is still there even after clicking 'delete'
  await page.getByRole('cell', { name: 'default' }).first().click();
});

test('cannot switch namespace while on settings page', async ({ page }) => {
  await page.getByRole('button', { name: 'default' }).click();
  await page.getByRole('button', { name: 'default' }).click();
  await page.getByRole('link', { name: 'Settings' }).click();
});
