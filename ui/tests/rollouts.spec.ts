import { expect, test } from '@playwright/test';

test.describe('Rollouts', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: 'Flags' }).click();
  });

  test('can create boolean flag', async ({ page }) => {
    await page.getByRole('button', { name: 'New Flag' }).click();
    await page.getByLabel('Name').fill('test-boolean');
    await page.getByLabel('Boolean').check();
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Successfully created flag')).toBeVisible();
    await expect(
      page.getByRole('heading', { name: 'Default Rollout' })
    ).toBeVisible();
  });

  test('has default rollout', async ({ page }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();
    await expect(
      page.getByRole('heading', { name: 'Default Rollout' })
    ).toBeVisible();
    await page.locator('#defaultValue').selectOption('true');
    await page.getByRole('button', { name: 'Update' }).last().click();
    await expect(
      page.getByText('Successfully updated default rollout')
    ).toBeVisible();
  });

  test('can create rollout', async ({ page }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();
    await page.getByRole('button', { name: 'New Rollout' }).click();
    await page.getByLabel('Percentage').fill('100');
    await page
      .getByLabel('New Rollout')
      .getByLabel('Value')
      .selectOption('false');
    await page.getByRole('textbox').fill('test'); // TODO: should get description by label
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Successfully created rollout')).toBeVisible();
    await expect(
      page.getByRole('button', { name: 'Threshold Rollout' })
    ).toBeVisible();
    await page.getByRole('list').getByLabel('Percentage').first().fill('70');
    await page.getByRole('list').getByLabel('Percentage').click();
    await page.getByRole('button', { name: 'Reset' }).first().click();
    await expect(page.getByLabel('Percentage')).toHaveValue('100');
  });

  test('can quick edit rollout', async ({ page }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();
    await page.getByRole('list').getByLabel('Percentage').first().fill('70');
    await page
      .getByRole('list')
      .getByRole('button', { name: 'Update' })
      .first()
      .click();
    await expect(page.getByText('Successfully updated rollout')).toBeVisible();
  });

  test('can edit rollout', async ({ page }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();
    await page.getByTestId('rollout-menu-button').click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await page.getByRole('textbox').click();
    await page.getByRole('textbox').fill('test2');
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated rollout')).toBeVisible();
  });

  test('can delete rollout', async ({ page }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();
    await page.getByTestId('rollout-menu-button').click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await expect(
      page.getByRole('button', { name: 'Threshold Rollout' })
    ).toBeHidden();
  });
});

test.describe('Rollouts - Read Only', () => {
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
    await page.getByRole('link', { name: 'test-boolean' }).click();
  });

  test('cannot create rollout', async ({ page }) => {
    await expect(
      page.getByRole('button', { name: 'New Rollout' })
    ).toBeDisabled();
  });
});
