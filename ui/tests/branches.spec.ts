import { expect, test } from '@playwright/test';

test.describe('Branches', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('can create branch', async ({ page }) => {
    await page
      .getByTestId('environment-namespace-switcher')
      .getByRole('button')
      .click();
    await page
      .getByTestId('environment-listbox')
      .getByRole('button', { name: 'local' })
      .click();
    await page
      .getByTestId('namespace-listbox')
      .getByRole('button', { name: 'default' })
      .click();
    await page.getByTestId('create-branch-button').click();
    await page.getByRole('textbox', { name: 'New branch name' }).click();
    await page.getByRole('textbox', { name: 'New branch name' }).fill('foo');
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Branch created successfully')).toBeVisible();
    await page
      .getByTestId('environment-namespace-switcher')
      .getByRole('button')
      .click();
    await page
      .getByTestId('environment-listbox')
      .getByRole('button', { name: 'foo' })
      .click();
    await page
      .getByTestId('namespace-listbox')
      .getByRole('button', { name: 'default' })
      .click();
  });

  test('cannot delete branch if not confirmed', async ({ page }) => {
    await page
      .getByTestId('environment-namespace-switcher')
      .getByRole('button')
      .click();
    await page
      .getByTestId('environment-listbox')
      .getByRole('button', { name: 'foo' })
      .click();
    await page
      .getByTestId('namespace-listbox')
      .getByRole('button', { name: 'default' })
      .click();
    await page.getByText('Branched from: local').click();
    await page.getByRole('menuitem', { name: 'Delete branch' }).click();
    await page.getByRole('textbox', { name: 'foo' }).fill('bar');
    await expect(
      page.getByRole('button', { name: 'Delete branch' })
    ).toBeDisabled();
  });

  test('can delete branch', async ({ page }) => {
    await page
      .getByTestId('environment-namespace-switcher')
      .getByRole('button')
      .click();
    await page
      .getByTestId('environment-listbox')
      .getByRole('button', { name: 'foo' })
      .click();
    await page
      .getByTestId('namespace-listbox')
      .getByRole('button', { name: 'default' })
      .click();
    await page.getByText('Branched from: local').click();
    await page.getByRole('menuitem', { name: 'Delete branch' }).click();
    await page.getByRole('textbox', { name: 'foo' }).fill('foo');
    await page.getByRole('button', { name: 'Delete branch' }).click();
    await expect(page.getByText('Branch deleted successfully')).toBeVisible();
  });
});
