import { useField } from 'formik';

import { cls } from '~/utils/helpers';

type InputProps = {
  id: string;
  name: string;
  type?: string;
  className?: string;
  autoComplete?: boolean;
  forwardRef?: React.RefObject<HTMLInputElement>;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
} & React.InputHTMLAttributes<HTMLInputElement>;

export default function Input(props: InputProps) {
  const {
    id,
    type = 'text',
    className = '',
    onChange,
    autoComplete = false,
    forwardRef,
    ...rest
  } = props;

  const [field, meta] = useField(props);
  const hasError = !!(meta.touched && meta.error);

  return (
    <>
      <input
        ref={forwardRef}
        className={cls(
          'block w-full rounded-md border-gray-300 bg-gray-50 text-gray-900 shadow-xs dark:border-gray-600 dark:bg-gray-900 dark:text-gray-100 disabled:cursor-not-allowed disabled:border-gray-200 disabled:bg-gray-100 disabled:text-gray-500 dark:disabled:border-gray-700 dark:disabled:bg-gray-800 dark:disabled:text-gray-400 sm:text-sm',
          className,
          {
            'border-red-400 dark:border-red-500': hasError
          }
        )}
        id={id}
        type={type}
        {...field}
        onChange={(e) => {
          field.onChange(e);
          onChange && onChange(e);
        }}
        autoComplete={autoComplete ? 'on' : 'off'}
        {...rest}
      />
      {hasError && meta.error?.length && meta.error.length > 0 ? (
        <div className="mt-1 text-sm text-red-500 dark:text-red-400">
          {meta.error}
        </div>
      ) : null}
    </>
  );
}
