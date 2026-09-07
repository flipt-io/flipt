import { expect, test } from '@playwright/test';

const exactText = (value: string) =>
  new RegExp(`^${value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}$`, 'i');

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
    test.setTimeout(30_000);

    await page.getByRole('link', { name: 'Settings' }).click();
    await page.getByRole('link', { name: 'Namespaces' }).click();
    await page.getByRole('button', { name: 'New Namespace' }).click();
    await page.getByLabel('Name', { exact: true }).fill('copy segment');
    await page.getByLabel('Description').fill('Copy Namespace');
    await page.getByRole('button', { name: 'Create' }).click();
    await expect(
      page.getByText('Successfully created namespace')
    ).toBeVisible();
    await page.getByRole('link', { name: 'Segments' }).click();
    await page.getByRole('link', { name: 'test-segment' }).click();

    // perform copy to new namespace
    await page.getByRole('button', { name: 'Actions' }).click();
    const copyAction = page.getByRole('menuitem', {
      name: 'Copy to Environment / Namespace'
    });
    await expect(copyAction).toBeEnabled({ timeout: 30000 });
    await copyAction.click();
    const namespaceSelect = page.getByRole('combobox', {
      name: 'Namespace'
    });
    await expect(namespaceSelect).toBeEnabled({ timeout: 30000 });
    await namespaceSelect.click();
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

  test('can copy segment to another environment', async ({ page }) => {
    test.setTimeout(30_000);

    await page.getByRole('link', { name: 'test-segment' }).click();

    await page.getByRole('button', { name: 'Actions' }).click();
    const copyAction = page.getByRole('menuitem', {
      name: 'Copy to Environment / Namespace'
    });
    await expect(copyAction).toBeEnabled({ timeout: 30000 });
    await copyAction.click();
    const environmentSelect = page.getByRole('combobox', {
      name: 'Environment'
    });
    await expect(environmentSelect).toBeEnabled({ timeout: 30000 });
    const currentEnvironmentName = (
      (await environmentSelect.textContent()) || ''
    )
      .trim()
      .toLowerCase();

    await environmentSelect.click();
    const environmentOptions = page.getByRole('option');
    await expect(environmentOptions.first()).toBeVisible({ timeout: 30000 });

    let targetEnvironmentName = '';
    const environmentOptionCount = await environmentOptions.count();
    for (let i = 0; i < environmentOptionCount; i++) {
      const option = environmentOptions.nth(i);
      const optionName = ((await option.textContent()) || '').trim();
      if (optionName && optionName.toLowerCase() !== currentEnvironmentName) {
        targetEnvironmentName = optionName;
        await option.click();
        break;
      }
    }

    test.skip(
      !targetEnvironmentName,
      'requires an alternate writable environment'
    );

    const namespaceSelect = page.getByRole('combobox', {
      name: 'Namespace'
    });
    await expect(namespaceSelect).toBeEnabled({ timeout: 30000 });
    await namespaceSelect.click();

    const namespaceOptions = page.getByRole('option');
    await expect(namespaceOptions.first()).toBeVisible({ timeout: 30000 });
    const targetNamespaceName = (
      (await namespaceOptions.first().textContent()) || ''
    ).trim();
    expect(targetNamespaceName).not.toBe('');
    await namespaceOptions.first().click();

    await page.getByRole('button', { name: 'Copy', exact: true }).click();
    await expect(page.getByText('Successfully copied segment')).toBeVisible();

    await page
      .getByTestId('environment-namespace-switcher')
      .getByRole('button')
      .click();
    await page
      .getByTestId('environment-listbox')
      .getByRole('button', { name: exactText(targetEnvironmentName) })
      .click();
    await page
      .getByTestId('namespace-listbox')
      .getByRole('button', { name: exactText(targetNamespaceName) })
      .click();

    // verify segment and its constraints were copied
    await page.getByRole('link', { name: 'Segments' }).click();
    await page.getByRole('link', { name: 'test-segment' }).click();
    await expect(page.getByText('Test Segment')).toBeVisible();
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
