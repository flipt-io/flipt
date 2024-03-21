import * as Yup from 'yup';

export const keyValidation = Yup.string()
  .required('Required')
  .matches(
    /^[-_,A-Za-z0-9]+$/,
    'Only letters, numbers, hypens and underscores allowed'
  );

export const keyWithDotValidation = Yup.string()
  .required('Required')
  .matches(
    /^[-_,A-Za-z0-9.]+$/,
    'Only letters, numbers, hypens, dots and underscores allowed'
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

// contextValidation is similar to json validation but it only allows object with strings
export const contextValidation = Yup.string()
  .nullable()
  .optional()
  .test('is-context', 'Must be valid context value', (value: any) => {
    if (value === undefined || value === null || value === '') {
      return true;
    }
    try {
      const data = JSON.parse(value);
      if (Array.isArray(data) || typeof data !== 'object') {
        return false;
      }
      for (const value of Object.values(data)) {
        if (typeof value !== 'string') {
          return false;
        }
      }
      return true;
    } catch {
      return false;
    }
  });

const MAX_JSON_ARRAY_ITEMS = 100;

const checkJsonArray =
  (checkItem: (v: any) => boolean) => (value: any, ctx: any) => {
    if (value === undefined || value === null || value === '') {
      return true;
    }

    try {
      const json = JSON.parse(value);
      if (!Array.isArray(json) || !json.every(checkItem)) {
        return false;
      }
      if (json.length > MAX_JSON_ARRAY_ITEMS) {
        return ctx.createError({
          message: `Too many items (maximum ${MAX_JSON_ARRAY_ITEMS})`
        });
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
