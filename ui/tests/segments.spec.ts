import { expect, test } from '@playwright/test';

test.describe('Segments', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.getByRole('link', { name: 'Segments' }).click();
  });

  test('can create segment', async ({ page }) => {
    await page.getByRole('button', { name: 'New Segment' }).click();
    await page.getByLabel('Name').fill('Test Segment');
    await page.getByLabel('Description').click();
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(page.getByText('Successfully created segment')).toBeVisible();
  });

  test('can update segment', async ({ page }) => {
    await page.getByRole('link', { name: 'test-segment' }).click();
    await page.getByLabel('Description').click();
    await page.getByLabel('Description').fill("i'm a test");
    await page.getByRole('button', { name: 'Update' }).click();
    await expect(page.getByText('Successfully updated segment')).toBeVisible();
  });

  test('can add constraints to segment', async ({ page }) => {
    await page.getByRole('link', { name: 'test-segment' }).click();

    await test.step('add constraint', async () => {
      await page.getByRole('button', { name: 'New Constraint' }).click();
      await page.getByLabel('Property').fill('foo');
      await page
        .getByRole('combobox', { name: 'Type' })
        .selectOption('BOOLEAN_COMPARISON_TYPE');
      await page
        .getByRole('combobox', { name: 'Operator' })
        .selectOption('notpresent');
      await page.getByRole('button', { name: 'Add' }).click();
      await page.getByRole('button', { name: 'Update' }).click();
      await expect(
        page.getByText('Successfully updated segment')
      ).toBeVisible();
    });

    await test.step('edit constraint properties', async () => {
      await page.getByText('foo', { exact: true }).click();
      await page
        .getByRole('combobox', { name: 'Type' })
        .selectOption('STRING_COMPARISON_TYPE');
      await page.getByLabel('Value').dblclick();
      await page.getByLabel('Value').fill('bar');
      await page
        .getByRole('dialog', { name: 'Edit Constraint' })
        .getByRole('button', { name: 'Done' })
        .click();
      await page.getByRole('button', { name: 'Update' }).click();

      await expect(
        page.getByText('Successfully updated segment').last()
      ).toBeVisible();
    });
  });

  test('can copy segment to new namespace', async ({ page }) => {
    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('link', { name: 'Namespaces' }).click();
    await page.getByRole('button', { name: 'New Namespace' }).click();
    await page.getByLabel('Name', { exact: true }).fill('copy segment');
    await page.getByLabel('Description').fill('Copy Namespace');
    await page.getByRole('button', { name: 'Create' }).click();
    await page.getByRole('link', { name: 'Segments' }).click();
    await page.getByRole('link', { name: 'test-segment' }).click();

    // perform copy to new namespace
    await page.getByRole('button', { name: 'Actions' }).click();
    await page.getByRole('menuitem', { name: 'Copy to Namespace' }).click();
    await page.locator('#copyToNamespace-select-button').click();
    await page
      .getByRole('option', { name: 'copy segment', exact: true })
      .click();
    await page.getByRole('button', { name: 'Copy', exact: true }).click();
    await expect(page.getByText('Successfully copied segment')).toBeVisible();

    // switch to new namespace
    await page.getByRole('link', { name: 'Segments', exact: true }).click();
    await page
      .getByTestId('environment-namespace-switcher')
      .getByRole('button')
      .click();
    await page
      .getByTestId('namespace-listbox')
      .getByRole('button', { name: 'copy segment' })
      .click();

    // verify segment was copied
    await page.getByRole('link', { name: 'Segments' }).click();
    await page.getByRole('link', { name: 'test-segment' }).click();
    await expect(page.getByText('Test Segment')).toBeVisible();

    // verify constraints were copied
    await expect(page.getByText('foo', { exact: true })).toBeVisible();
  });

  test('can delete segment', async ({ page }) => {
    await page.getByRole('link', { name: 'test-segment' }).click();
    await page.getByRole('button', { name: 'Actions' }).click();
    await page.getByRole('menuitem', { name: 'Delete' }).click();
    await page.getByRole('button', { name: 'Delete' }).click();
    await expect(page.getByText('Successfully deleted segment')).toBeVisible();
  });
});
