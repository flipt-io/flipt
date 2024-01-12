import { capture } from '../../screenshot.js';

(async () => {
  await capture('getting_started', 'create_constraint', async (page) => {
    await page.getByRole('link', { name: 'Segments' }).click();
    await page.getByRole('link', { name: 'all-users' }).click();
    await page.getByRole('button', { name: 'New Constraint' }).click();
    await page.getByLabel('Property').fill('admin');
    await page
      .getByRole('combobox', { name: 'Type' })
      .selectOption('BOOLEAN_COMPARISON_TYPE');
    await page
      .getByRole('combobox', { name: 'Operator' })
      .selectOption('notpresent');
  });
})();
