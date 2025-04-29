import { expect, test } from '@playwright/test';

test.describe('Rules', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: 'Flags' }).click();
  });

  test('can create rule', async ({ page }) => {
    await test.step('create flag', async () => {
      await page.getByRole('button', { name: 'New Flag' }).click();
      await page.getByTestId('VARIANT_FLAG_TYPE').click();
      await page.getByLabel('Name').fill('Test Rule');
      await page.getByRole('switch', { name: 'Enabled' }).click();
      await page.getByRole('button', { name: 'Create' }).click();
      await page.getByRole('button', { name: 'New Variant' }).click();
      await page
        .getByRole('dialog', { name: 'New Variant' })
        .locator('#key')
        .fill('123');
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'New Variant' }).click();
      await page
        .getByRole('dialog', { name: 'New Variant' })
        .locator('#key')
        .fill('456');
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'Update' }).click();
    });

    await test.step('create segment', async () => {
      await page.getByRole('link', { name: 'Segments' }).click();
      await page.getByRole('button', { name: 'New Segment' }).click();
      await page.getByLabel('Name').fill('Test Rule');
      await page.getByLabel('Description').click();
      await page.getByRole('button', { name: 'Create' }).click();
    });

    await test.step('create multi-variate rule', async () => {
      await page.reload();
      await page.getByRole('link', { name: 'Flags' }).click();
      await page.getByRole('link', { name: 'test-rule' }).click();
      await page.getByRole('link', { name: 'Rules' }).click();
      await page.getByRole('button', { name: 'New Rule' }).click();
      await page.locator('#segmentKey-0-select-button').click();
      await page.getByLabel('New Rule').getByText('Test Rule').click();
      await page.getByLabel('Multi-Variate').check();
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });

    await test.step('create single-variant rule', async () => {
      await page.getByRole('button', { name: 'New Rule' }).click();
      await page
        .getByLabel('New Rule')
        .locator('#segmentKey-0-select-button')
        .click();
      await page.getByLabel('New Rule').getByText('Test Rule').click();
      await page.getByLabel('Single Variant').check();
      await page.locator('#variant-select-button').click();
      await page.getByLabel('New Rule').getByText('123').click();
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });
  });

  test('can update multi-variate rule', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();
    await page
      .locator(
        'input[name="rules\\.\\[0\\]\\.distributions\\.\\[0\\]\\.rollout"]'
      )
      .click();
    await page
      .locator(
        'input[name="rules\\.\\[0\\]\\.distributions\\.\\[0\\]\\.rollout"]'
      )
      .fill('40');
    await page
      .locator(
        'input[name="rules\\.\\[0\\]\\.distributions\\.\\[1\\]\\.rollout"]'
      )
      .click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
  });

  test('can update single-variant rule', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();
    await page
      .locator(
        '[id="rules\\.\\[1\\]\\.distributions\\.\\[0\\]\\.variant-select-button"]'
      )
      .click();
    await page
      .locator('li')
      .filter({ hasText: 'Single Variant' })
      .locator('li')
      .filter({ hasText: '456' })
      .click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
  });

  test('can reorder rules', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();
    await page
      .getByTestId('rule-1')
      .getByRole('button', { name: 'Rule' })
      .dragTo(page.getByTestId('rule-0').getByRole('button', { name: 'Rule' }));
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
    await expect(
      page.getByTestId('rule-0').getByRole('button', { name: '1' })
    ).toBeVisible();
    await expect(
      page.getByTestId('rule-1').getByRole('button', { name: '2' })
    ).toBeVisible();
  });

  test('can update default rule', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();
    await expect(
      page.getByRole('heading', { name: 'Default Rule' })
    ).toBeVisible();
    await page.locator('#defaultVariant-select-button').click();
    await page.getByRole('option', { name: '456' }).locator('div').click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
  });

  test('can delete rule', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();
    await page.getByTestId('rule-1').getByTestId('rule-menu-button').click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
    await expect(page.getByTestId('rule-1')).toBeHidden();
  });
});
