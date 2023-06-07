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
      await page.getByRole('button', { name: 'Create' }).click();
      await expect(
        page.getByText('Successfully created constraint')
      ).toBeVisible();
    });

    await test.step('edit constraint properties', async () => {
      await page.getByRole('link', { name: 'Edit , foo' }).click();
      await page
        .getByRole('combobox', { name: 'Type' })
        .selectOption('STRING_COMPARISON_TYPE');
      await page.getByLabel('Value').dblclick();
      await page.getByLabel('Value').fill('bar');
      await page
        .getByRole('dialog', { name: 'Edit Constraint' })
        .getByRole('button', { name: 'Update' })
        .click();
      await expect(
        page.getByText('Successfully updated constraint')
      ).toBeVisible();
    });
  });
});

test.describe('Segments - Read Only', () => {
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
    await page.getByRole('link', { name: 'Segments' }).click();
  });

  test('can not create segment', async ({ page }) => {
    await expect(
      page.getByRole('button', { name: 'New Segment' })
    ).toBeDisabled();
  });

  test('can not update segment', async ({ page }) => {
    await page.getByRole('link', { name: 'test-segment' }).click();
    await page.getByLabel('Description').click();
    await page.getByLabel('Description').fill("i'm a test 2");
    await expect(page.getByRole('button', { name: 'Update' })).toBeDisabled();
  });

  test('can not add constraints to segment', async ({ page }) => {
    await page.getByRole('link', { name: 'test-segment' }).click();

    await expect(
      page.getByRole('button', { name: 'New Constraint' })
    ).toBeDisabled();
  });

  test('can not delete segment', async ({ page }) => {
    await page.getByRole('link', { name: 'test-segment' }).click();
    await expect(page.getByRole('button', { name: 'Delete' })).toBeDisabled();
  });
});
