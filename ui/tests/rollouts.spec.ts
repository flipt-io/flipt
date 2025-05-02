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
      await page.getByLabel('New Rollout').getByLabel('Segment').check();

      // Add first segment
      await page
        .getByLabel('New Rollout')
        .locator('#segmentKey-0-select-button')
        .click();
      await page.getByLabel('New Rollout').getByText('Test Rollout').click();

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
        .getByLabel('New Rollout')
        .locator('#segmentKey-1-select-button')
        .click();
      await page.getByLabel('New Rollout').getByText('Second Segment').click();

      // Select the AND operator
      const operatorOptions = await page.getByRole('radio').all();
      // Find the AND operator radio button
      for (const radio of operatorOptions) {
        const text = await radio.evaluate((node) => {
          const label = document.querySelector(`label[for="${node.id}"]`);
          return label ? label.textContent : '';
        });
        if (text && text.includes('AND')) {
          await radio.click();
          break;
        }
      }

      await page
        .getByLabel('New Rollout')
        .getByRole('button', { name: 'Add' })
        .click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(page.getByText('Successfully updated flag')).toBeVisible();
    });
  });

  test('can edit segment rollout and add/remove segments', async ({ page }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();

    // Find a segment rollout and edit it
    const rolloutElements = await page.getByTestId(/rollout-\d+/).all();
    for (const rollout of rolloutElements) {
      const isSegmentRollout =
        (await rollout.getByText('Segment Rollout').count()) > 0;
      if (isSegmentRollout) {
        await rollout.getByTestId('rollout-menu-button').click();
        await page.getByRole('menuitem', { name: 'Edit' }).click();

        // Test removing a segment if there are multiple
        const minusButtons = await page.getByRole('button').all();
        const removeSegmentButton = minusButtons.find((button) =>
          button.evaluate((node) => node.innerHTML.includes('fa-minus'))
        );

        if (removeSegmentButton) {
          // Should have at least 2 segments to be able to remove one
          const segmentCount = await page
            .locator('[id^="segmentKey-"]')
            .count();
          if (segmentCount > 1) {
            await removeSegmentButton.click();
            // Verify segment was removed
            await expect(page.locator('[id^="segmentKey-"]')).toHaveCount(
              segmentCount - 1
            );
          }
        }

        // Save changes
        await page.getByRole('button', { name: 'Done' }).click();
        await page.getByRole('button', { name: 'Update' }).click();
        await expect(page.getByText('Successfully updated flag')).toBeVisible();
        break;
      }
    }
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

  test('cannot delete the last segment in a segment rollout', async ({
    page
  }) => {
    await page.getByRole('link', { name: 'test-boolean' }).click();

    // Find a segment rollout to edit
    const rolloutElements = await page.getByTestId(/rollout-\d+/).all();
    for (const rollout of rolloutElements) {
      const isSegmentRollout =
        (await rollout.getByText('Segment Rollout').count()) > 0;
      if (isSegmentRollout) {
        await rollout.getByTestId('rollout-menu-button').click();
        await page.getByRole('menuitem', { name: 'Edit' }).click();

        // Count current segments
        const segmentCount = await page.locator('[id^="segmentKey-"]').count();

        // If there's only one segment, verify no minus button is visible
        if (segmentCount === 1) {
          const minusButtons = await page.getByRole('button').all();
          const hasMinusButton = minusButtons.some((button) =>
            button.evaluate((node) => node.innerHTML.includes('fa-minus'))
          );

          // Should not have a minus button for the last segment
          expect(hasMinusButton).toBeFalsy();

          // Verify the plus button is shown instead (if more segments are available)
          const hasPlusButton = minusButtons.some((button) =>
            button.evaluate((node) => node.innerHTML.includes('fa-plus'))
          );

          // Cancel without making changes
          await page.getByRole('button', { name: 'Cancel' }).click();
          break;
        }
      }
    }
  });
});
