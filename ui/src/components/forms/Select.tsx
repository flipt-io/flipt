import { useField } from 'formik';
import { ComponentPropsWithoutRef } from 'react';

import { cls } from '~/utils/helpers';

interface SelectOption {
  value: string;
  label: string;
}

type SelectProps = Omit<ComponentPropsWithoutRef<'select'>, 'className'> & {
  options?: SelectOption[];
  className?: string;
  name: string; // Keep name required for Formik
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
    disabled = false,
    ...restProps
  } = props;

  const [field] = useField({
    name,
    type: 'select'
  });

  return (
    <select
      {...field}
      {...restProps}
      id={id}
      name={name}
      className={cls(
        'block rounded border-input bg-secondary py-2 pl-3 pr-10 min-h-9 text-base focus:outline-hidden text-secondary-foreground sm:text-sm',
        'focus-visible:border-ring focus-visible:ring-ring/50 mb-1',
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
