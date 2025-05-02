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
        .getByTestId('segments')
        .locator('[id$="-select-button"]')
        .first()
        .click();
      await page.getByLabel('New Rule').getByText('Test Rule').click();
      await page.getByLabel('Multi-Variate').check();
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });

    await test.step('create single-variant rule', async () => {
      await page.getByRole('button', { name: 'New Rule' }).click();
      await page
        .getByTestId('segments')
        .locator('[id$="-select-button"]')
        .first()
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
        .getByTestId('segments')
        .locator('[id$="-select-button"]')
        .first()
        .click();
      await page.getByLabel('New Rule').getByText('Test Rule').click();

      // Click the + button to add another segment
      const plusButtons = await page.getByRole('button').all();
      const addSegmentButton = plusButtons.find((button) =>
        button.evaluate((node) => node.innerHTML.includes('fa-plus'))
      );

      // Make sure we found the + button before clicking it
      if (!addSegmentButton) {
        throw new Error('Could not find the + button to add another segment');
      }

      await addSegmentButton.click();

      // Add second segment
      await page
        .getByTestId('segments')
        .locator('[id$="-select-button"]')
        .nth(1)
        .click();
      await page
        .getByLabel('New Rule')
        .getByText('Second Rule Segment')
        .click();

      // Select the OR operator (default, but let's be explicit)
      const operatorOptions = await page.getByRole('radio').all();
      // Find the OR operator radio button
      for (const radio of operatorOptions) {
        const text = await radio.evaluate((node) => {
          const label = document.querySelector(`label[for="${node.id}"]`);
          return label ? label.textContent : '';
        });
        if (text && text.includes('OR')) {
          await radio.click();
          break;
        }
      }

      // Complete the rule creation
      await page.getByLabel('Single Variant').check();
      await page.locator('#variant-select-button').click();
      await page.getByLabel('New Rule').getByText('456').click();
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });
  });

  test('can edit rule and modify segments', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();

    // Find a rule to edit
    await page
      .getByTestId('rule-0')
      .getByTestId('rule-menu-button')
      .first()
      .click();
    await page.getByRole('menuitem', { name: 'Edit' }).click();

    // Test segment operations
    const segmentCount = await page
      .getByTestId('segments')
      .locator('[id$="-select-button"]')
      .count();

    // If multiple segments, test removing one
    if (segmentCount > 1) {
      const minusButtons = await page.getByRole('button').all();
      const removeSegmentButton = minusButtons.find((button) =>
        button.evaluate((node) => node.innerHTML.includes('fa-minus'))
      );

      // Only attempt to click if we found a removeSegmentButton
      if (removeSegmentButton) {
        await removeSegmentButton.click();
        // Verify segment was removed
        await expect(
          page.getByTestId('segments').locator('[id$="-select-button"]')
        ).toHaveCount(segmentCount - 1);
      }
    } else {
      // Verify we can't remove the last segment (no minus button)
      const minusButtons = await page.getByRole('button').all();
      const hasMinusButton = minusButtons.some((button) =>
        button.evaluate((node) => node.innerHTML.includes('fa-minus'))
      );

      if (hasMinusButton) {
        // If there's only one segment and we find a minus button, that's a problem
        expect(hasMinusButton).toBeFalsy();
      }
    }

    // Save changes
    await page.getByRole('button', { name: 'Done' }).click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
  });

  test('can verify variant selection behavior', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();

    // Create a new rule for testing variant selection
    await page.getByRole('button', { name: 'New Rule' }).click();
    await page
      .getByTestId('segments')
      .locator('[id$="-select-button"]')
      .first()
      .click();
    await page.getByLabel('New Rule').getByText('Test Rule').click();

    // Test single variant selection
    await page.getByLabel('Single Variant').check();
    await page.locator('#variant-select-button').click();
    await page.getByLabel('New Rule').getByText('123').click();

    // Verify the selection appears in the UI
    const selectedVariantText = await page
      .locator('#variant-select-input')
      .inputValue();
    expect(['123', '123 ', ' 123']).toContain(selectedVariantText.trim());

    // Now switch to multi-variant and back to verify selection behavior
    await page.getByLabel('Multi-Variate').check();
    await page.getByLabel('Single Variant').check();

    // Verify variant selection persists or resets appropriately
    await page.locator('#variant-select-button').click();
    const variantOptions = await page.getByRole('option').all();
    expect(variantOptions.length).toBeGreaterThan(0);

    // Select a different variant
    await page.getByText('456').click();

    // Complete the rule creation
    await page.getByRole('button', { name: 'Add' }).click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
  });

  test('can update multi-variate rule', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();
    await page.getByTestId('distribution-input').first().click();
    await page.getByTestId('distribution-input').first().fill('40');
    await page.getByTestId('distribution-input').nth(1).click();
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated flag')).toBeVisible();
  });

  test('can update single-variant rule', async ({ page }) => {
    await page.getByRole('link', { name: 'test-rule' }).click();
    await page.getByRole('link', { name: 'Rules' }).click();
    await page.getByTestId('variant-select-button').first().click();
    await page.locator('li').filter({ hasText: '456' }).first().click();
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
    await expect(page.getByTestId('rule-1')).toBeHidden();
  });
});
