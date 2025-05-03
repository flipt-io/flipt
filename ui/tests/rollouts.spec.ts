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

  test('can create threshold rollout', async ({ page }) => {
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
  });

  test('can quick edit threshold rollout', async ({ page }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();
    await page.getByRole('list').getByLabel('Percentage').first().fill('70');
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
  });

  test('can edit threshold rollout', async ({ page }) => {
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
      await page
        .getByLabel('New Rollout')
        .getByText('Segment', { exact: true })
        .click();

      await page
        .getByRole('dialog', { name: 'New Rollout' })
        .getByTestId('segmentKey-0-select-button')
        .click();
      await page.getByRole('option', { name: 'Test Rollout' }).click();

      await page
        .getByLabel('New Rollout')
        .getByRole('button', { name: 'Add' })
        .click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });
  });

  test('can create segment rollout with multiple segments', async ({
    page
  }) => {
    await test.step('create additional segment', async () => {
      await page.getByRole('link', { name: 'Segments' }).click();
      await page.getByRole('button', { name: 'New Segment' }).click();
      await page.getByLabel('Name').fill('Second Segment');
      await page.getByLabel('Description').click();
      await page.getByRole('button', { name: 'Create' }).click();
    });

    await test.step('create multi-segment rollout', async () => {
      await page.reload();
      await page.getByRole('link', { name: 'Flags' }).click();
      await page.getByRole('link', { name: 'test-boolean' }).click();
      await page.getByRole('button', { name: 'New Rollout' }).click();
      await page
        .getByLabel('New Rollout')
        .getByLabel('Segment', { exact: true })
        .check();

      // Add first segment
      await page
        .getByRole('dialog', { name: 'New Rollout' })
        .getByTestId('segmentKey-0-select-button')
        .click();
      await page.getByRole('option', { name: 'Test Rollout' }).click();

      await page
        .getByRole('dialog', { name: 'New Rollout' })
        .getByTestId('add-segment-button-0')
        .click();

      // Add second segment
      await page
        .getByRole('dialog', { name: 'New Rollout' })
        .locator('#segmentKey-1-select-input')
        .click();
      await page.getByRole('option', { name: 'Second Segment' }).click();

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

    // Wait for rollouts to be visible before attempting drag
    await page.getByTestId('rollout-0').waitFor();
    await page.getByTestId('rollout-1').waitFor();

    await page
      .getByTestId('rollout-0')
      .getByRole('button', { name: 'Rollout' })
      .first()
      .dragTo(
        page
          .getByTestId('rollout-1')
          .getByRole('button', { name: 'Rollout' })
          .first()
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

    // Wait for rollout to be visible before attempting to click
    await page.getByTestId('rollout-1').waitFor();

    await page
      .getByTestId('rollout-1')
      .getByTestId('rollout-menu-button')
      .first()
      .click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
    await expect(page.getByTestId('rollout-1')).toBeHidden();
  });
});
