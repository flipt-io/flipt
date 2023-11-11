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

const checkJsonArray = (checkItem: (v: any) => boolean) => (value: any) => {
  if (value === undefined || value === null || value === '') {
    return true;
  }

  try {
    const json = JSON.parse(value);
    if (!Array.isArray(json) || !json.every(checkItem)) {
      return false;
    }

    return true;
  } catch {
    return false;
  }
};

export const jsonStringArrayValidation = Yup.string()
  .optional()
  .test(
    'is-json-string-array',
    'Must be valid JSON string array',
    checkJsonArray((v: any) => typeof v === 'string')
  );

export const jsonNumberArrayValidation = Yup.string()
  .optional()
  .test(
    'is-json-number-array',
    'Must be valid JSON number array',
    checkJsonArray((v: any) => typeof v === 'number')
  );
