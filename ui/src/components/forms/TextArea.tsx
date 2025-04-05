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
          'block w-full rounded-md border-gray-300 bg-gray-50 text-gray-900 shadow-xs focus:border-violet-500 focus:ring-violet-500 sm:text-sm',
          className,
          {
            'border-red-400': hasError
          }
        )}
        placeholder={placeholder}
        autoComplete={autocomplete ? 'on' : 'off'}
        {...field}
      />
      {meta.touched && meta.error ? (
        <div className="mt-1 text-sm text-red-500">{meta.error}</div>
      ) : null}
    </>
  );
}
