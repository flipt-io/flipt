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
        'block rounded-md border-gray-300 bg-gray-50 py-2 pl-3 pr-10 text-base text-gray-900 focus:border-violet-300 focus:outline-hidden focus:ring-violet-300 sm:text-sm',
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
