import * as Yup from 'yup';

export const keyValidation = Yup.string()
  .required('Required')
  .matches(
    /^[-_,A-Za-z0-9]+$/,
    'Only letters, numbers, hypens and underscores allowed'
  );

export const requiredValidation = Yup.string().required('Required');

export const jsonValidation = Yup.string()
  .nullable()
  .optional()
  .test('is-json', 'Must be valid JSON', (value: any) => {
    if (value === undefined || value === null || value === '') {
      return true;
    }
    try {
      JSON.parse(value);
      return true;
    } catch {
      return false;
    }
  });
