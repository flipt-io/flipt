import { useField } from 'formik';

import { cls } from '~/utils/helpers';

type TextAreaProps = {
  id: string;
  name: string;
  autocomplete?: boolean;
  rows?: number;
  className?: string;
  placeholder?: string;
};

export default function TextArea(props: TextAreaProps) {
  const { id, rows = 3, className, placeholder, autocomplete = false } = props;
  const [field, meta] = useField(props);
  const hasError = meta.touched && meta.error;

  return (
    <>
      <textarea
        id={id}
        rows={rows}
        className={cls(
          'block w-full rounded-md border-gray-300 bg-gray-50 text-gray-900 shadow-xs dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 sm:text-sm',
          className,
          {
            'border-red-400 dark:border-red-500': hasError
          }
        )}
        placeholder={placeholder}
        autoComplete={autocomplete ? 'on' : 'off'}
        {...field}
      />
      {meta.touched && meta.error ? (
        <div className="mt-1 text-sm text-red-500 dark:text-red-400">
          {meta.error}
        </div>
      ) : null}
    </>
  );
}
