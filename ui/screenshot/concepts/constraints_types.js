import { capture } from '../../screenshot.js';

(async () => {
  await capture('concepts', 'constraints_types', async (page) => {
    await page.getByRole('link', { name: 'Segments' }).click();
    await page.getByRole('link', { name: 'lunch-party' }).click();
    await page.getByRole('button', { name: 'New Constraint' }).click();
    await page.getByLabel('Property').fill('sale_date');
    await page.getByLabel('Type').selectOption('DATETIME_COMPARISON_TYPE');
    await page.getByLabel('Operator').selectOption('gt');
  });
})();
