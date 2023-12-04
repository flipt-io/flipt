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
          'text-gray-900 bg-gray-50 border-gray-300 block w-full rounded-md shadow-sm focus:border-violet-300 disabled:text-gray-500 disabled:bg-gray-100 disabled:border-gray-200 focus:ring-violet-300 disabled:cursor-not-allowed sm:text-sm',
          className,
          {
            'border-red-400': hasError
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
        <div className="text-red-500 mt-1 text-sm">{meta.error}</div>
      ) : null}
    </>
  );
}
