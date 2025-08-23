import { useField } from 'formik';

import { cls } from '~/utils/helpers';

type InputProps = {
  id: string;
  name: string;
  type?: string;
  className?: string;
  autoComplete?: boolean;
  forwardRef?: React.RefObject<HTMLInputElement | null>;
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
        aria-invalid={hasError}
        className={cls(
          'placeholder:text-muted-foreground dbg-secondary/20 dark:bg-secondary/80 border-input w-min-0 w-full rounded-md text-sm flex items-center px-3 py-1 shadow-xs outline-none',
          'disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50',
          'focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[1px]',
          'aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive',
          className
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
        <div className="pt-1 text-sm text-destructive">{meta.error}</div>
      ) : null}
    </>
  );
}
