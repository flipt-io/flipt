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
          'bg-input/60 border-input focus:border-brand focus:ring-brand disabled:bg-muted disabled:text-muted-foreground block w-full rounded-md shadow-xs disabled:cursor-not-allowed sm:text-sm',
          className,
          {
            'border-destructive': hasError
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
        <div className="text-destructive mt-1 text-sm">{meta.error}</div>
      ) : null}
    </>
  );
}
