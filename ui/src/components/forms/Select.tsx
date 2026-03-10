import { useField } from 'formik';
import { cls } from '~/utils/helpers';

type SelectProps = {
  id: string;
  name: string;
  options?: { value: string; label: string }[];
  children?: React.ReactNode;
  className?: string;
  value?: string;
  onChange?: (e: React.ChangeEvent<HTMLSelectElement>) => void;
  disabled?: boolean;
};

export default function Select(props: SelectProps) {
  const {
    id,
    name,
    options,
    children,
    className,
    value,
    onChange,
    disabled = false
  } = props;

  const [field] = useField({
    name,
    type: 'select'
  });

  return (
    <select
      {...field}
      id={id}
      name={name}
      className={cls(
        'border-input bg-input/60 text-secondary-foreground focus:border-brand focus:ring-brand block rounded-md py-2 pr-10 pl-3 text-base focus:outline-hidden sm:text-sm',
        className
      )}
      value={value}
      onChange={onChange || field.onChange}
      disabled={disabled}
    >
      {options &&
        options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      {!options && children}
    </select>
  );
}
