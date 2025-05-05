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
      await page
        .getByRole('dialog', { name: 'New Rule' })
        .getByTestId('segmentKey-0-select-button')
        .click();
      await page.getByRole('option', { name: 'Test Rule' }).click();

      await page.getByLabel('Multi-Variate').check();
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });

    await test.step('create single-variant rule', async () => {
      await page.getByRole('button', { name: 'New Rule' }).click();
      await page
        .getByRole('dialog', { name: 'New Rule' })
        .getByTestId('segmentKey-0-select-button')
        .click();
      await page.getByRole('option', { name: 'Test Rule' }).click();
      await page.getByLabel('Single Variant').check();
      await page.locator('#variant-select-button').click();
      await page.getByLabel('New Rule').getByText('123').click();
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });
  });

  test('create rule with multiple segments', async ({ page }) => {
    await test.step('create additional segment', async () => {
      await page.getByRole('link', { name: 'Segments' }).click();
      await page.getByRole('button', { name: 'New Segment' }).click();
      await page.getByLabel('Name').fill('Second Rule Segment');
      await page.getByLabel('Description').click();
      await page.getByRole('button', { name: 'Create' }).click();
    });

    await test.step('create rule with multiple segments', async () => {
      await page.reload();
      await page.getByRole('link', { name: 'Flags' }).click();
      await page.getByRole('link', { name: 'test-rule' }).click();
      await page.getByRole('link', { name: 'Rules' }).click();
      await page.getByRole('button', { name: 'New Rule' }).click();

      // Add first segment
      await page
        .getByRole('dialog', { name: 'New Rule' })
        .getByTestId('segmentKey-0-select-button')
        .click();
      await page.getByRole('option', { name: 'Test Rule' }).click();

      await page
        .getByRole('dialog', { name: 'New Rule' })
        .getByTestId('add-segment-button-0')
        .click();

      // Add second segment
      await page
        .getByRole('dialog', { name: 'New Rule' })
        .getByTestId('segmentKey-1-select-button')
        .click();
      await page.getByRole('option', { name: 'Second Rule Segment' }).click();

      // Complete the rule creation
      await page.getByLabel('Single Variant').check();
      await page.locator('#variant-select-button').click();
      await page.getByLabel('New Rule').getByText('456').click();
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
      .fill('40');
    await page
      .locator(
        'input[name="rules\\.\\[0\\]\\.distributions\\.\\[1\\]\\.rollout"]'
      )
      .fill('60');
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
  });

  test('can edit single-variant rule via quick edit', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();

    // Find a single-variant rule
    const ruleElements = await page.getByTestId(/rule-\d+/).all();
    for (const rule of ruleElements) {
      const isSingleVariant =
        (await rule.getByText('Single Variant').count()) > 0;
      if (isSingleVariant) {
        // Open quick edit
        await rule.getByTestId('rule-menu-button').click();
        await page.getByRole('menuitem', { name: 'Quick Edit' }).click();

        // Change variant
        await page.locator('[id$="variant-select-button"]').click();

        // Select any variant that's not currently selected
        const variantOptions = await page.getByRole('option').all();
        if (variantOptions.length > 0) {
          await variantOptions[0].click();
        }

        // Save changes
        await page.getByRole('button', { name: 'Update' }).click();
        await expect(page.getByText('Successfully updated flag')).toBeVisible();
        break;
      }
    }
  });

  test('can reorder rules', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();

    // Wait for rules to be visible before attempting drag
    await page.getByTestId('rule-0').waitFor();
    await page.getByTestId('rule-1').waitFor();

    await page
      .getByTestId('rule-1')
      .getByRole('button', { name: 'Rule' })
      .first()
      .dragTo(
        page.getByTestId('rule-0').getByRole('button', { name: 'Rule' }).first()
      );
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

    // Wait for rule to be visible before attempting to click
    await page.getByTestId('rule-1').waitFor();

    await page
      .getByTestId('rule-1')
      .getByTestId('rule-menu-button')
      .first()
      .click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
  });
});
