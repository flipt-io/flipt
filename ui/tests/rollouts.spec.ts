import { expect, test } from '@playwright/test';

test.describe('Rollouts', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: 'Flags' }).click();
  });

  test('can create boolean flag', async ({ page }) => {
    await page.getByRole('button', { name: 'New Flag' }).click();
    await page.getByTestId('BOOLEAN_FLAG_TYPE').click();
    await page.getByLabel('Name').fill('test-boolean');
    await page.getByLabel('Description').click();
    // await page.getByLabel('Boolean').check();
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Successfully created flag')).toBeVisible();
  });

  test('can update default rollout', async ({ page }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();
    await page.locator('#defaultValue').click();
    await page.getByRole('button', { name: 'Update' }).last().click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
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
    await page.getByRole('button', { name: 'Add' }).click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
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
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
  });

  test('can edit rollout', async ({ page }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();
    await page.getByTestId('rollout-menu-button').click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();
    await page.getByRole('textbox').click();
    await page.getByRole('textbox').fill('test2');
    await page.getByRole('button', { name: 'Done' }).click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
  });

  test('can create segment rollout', async ({ page }) => {
    await test.step('create segment', async () => {
      await page.getByRole('link', { name: 'Segments' }).click();
      await page.getByRole('button', { name: 'New Segment' }).click();
      await page.getByLabel('Name').fill('Test Rollout');
      await page.getByLabel('Description').click();
      await page.getByRole('button', { name: 'Create' }).click();
    });

    await test.step('create rollout', async () => {
      await page.reload();
      await page.getByRole('link', { name: 'Flags' }).click();
      await page.getByRole('link', { name: 'test-boolean' }).click();
      await page.getByRole('button', { name: 'New Rollout' }).click();
      await page.getByLabel('New Rollout').getByLabel('Segment').check();

      await page
        .getByLabel('New Rollout')
        .locator('#segmentKey-0-select-button')
        .click();
      await page.getByLabel('New Rollout').getByText('Test Rollout').click();
      await page
        .getByLabel('New Rollout')
        .getByRole('button', { name: 'Add' })
        .click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });
  });

  test('can reorder rollouts', async ({ page }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();
    await page
      .getByTestId('rollout-0')
      .getByRole('button', { name: 'Rollout' })
      .dragTo(
        page.getByTestId('rollout-1').getByRole('button', { name: 'Rollout' })
      );
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
    await expect(
      page.getByTestId('rollout-0').getByRole('button', { name: '1' })
    ).toBeVisible();
    await expect(
      page.getByTestId('rollout-1').getByRole('button', { name: '2' })
    ).toBeVisible();
  });

  test('can delete rollout', async ({ page }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();
    await page
      .getByTestId('rollout-1')
      .getByTestId('rollout-menu-button')
      .click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
    await expect(page.getByTestId('rollout-1')).toBeHidden();
  });
});
