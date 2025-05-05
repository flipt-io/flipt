import { Field, FieldProps } from 'formik';
import { useState } from 'react';

import { cls } from '~/utils/helpers';

function JsonInput({
  field,
  form,
  meta,
  className
}: FieldProps & { className?: string }): React.ReactElement {
  const [inputValue, setInputValue] = useState(
    JSON.stringify(field.value, null, 2)
  );

  const handleChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    const { value } = event.target;
    setInputValue(value);

    try {
      const parsedValue = JSON.parse(value);
      form.setFieldValue(field.name, parsedValue);
    } catch (e) {
      form.setFieldValue(field.name, value); // keep it as a string if not valid JSON
    }
  };

  return (
    <div>
      <textarea
        {...field}
        value={inputValue}
        onChange={handleChange}
        className={cls(
          'block w-full rounded-md border-gray-300 bg-gray-50 text-gray-900 shadow-xs focus:border-violet-500 focus:ring-violet-500 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 dark:focus:border-violet-500 dark:focus:ring-violet-500 sm:text-sm',
          className,
          {
            'border-red-400 dark:border-red-500': meta?.touched && meta?.error
          }
        )}
        placeholder="Enter JSON"
        rows={10}
        cols={50}
      />
      {meta?.touched && meta?.error ? (
        <div className="mt-1 text-sm text-red-500 dark:text-red-400">
          {meta.error}
        </div>
      ) : null}
    </div>
  );
}

const validateJson = (value: string): string | undefined => {
  if (typeof value === 'string') {
    try {
      JSON.parse(value);
    } catch (e) {
      return 'Invalid JSON';
    }
  }
  return undefined;
};

export function JsonTextArea({
  name,
  id,
  className
}: {
  name: string;
  id: string;
  className?: string;
}) {
  return (
    <Field
      name={name}
      id={id}
      validate={validateJson}
      component={JsonInput}
      className={className}
    />
  );
}
