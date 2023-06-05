import { useField } from 'formik';

type SelectProps = {
  id: string;
  name: string;
  options?: { value: string; label: string }[];
  children?: React.ReactNode;
  className?: string;
  value?: string;
  onChange?: (e: React.ChangeEvent<HTMLSelectElement>) => void;
};

export default function Select(props: SelectProps) {
  const { id, name, options, children, className, value, onChange } = props;

  const [field] = useField({
    name,
    type: 'select'
  });

  return (
    <select
      {...field}
      id={id}
      className={`${className} block rounded-md py-2 pl-3 pr-10 text-base border-gray-300 focus:outline-none focus:ring-violet-300 focus:border-violet-300 sm:text-sm`}
      value={value}
      onChange={onChange || field.onChange}
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
