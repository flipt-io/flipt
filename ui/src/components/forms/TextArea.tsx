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
          'border-input bg-input/60 focus:border-brand focus:ring-brand block w-full rounded-md shadow-xs sm:text-sm',
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
        <div className="text-destructive mt-1 text-sm">{meta.error}</div>
      ) : null}
    </>
  );
}
