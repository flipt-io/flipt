import { useField } from 'formik';

type SelectProps = {
  id: string;
  name: string;
  options?: { value: string; label: string }[];
  children?: React.ReactNode;
  className?: string;
  value?: string;
  defaultValue?: string;
  handleChange?: (e: React.ChangeEvent<HTMLSelectElement>) => void;
};

export default function Select(props: SelectProps) {
  const { id, name, options, children, className, defaultValue, handleChange } =
    props;

  const [field] = useField({
    name,
    type: 'select'
  });

  return (
    <select
      id={id}
      className={`${className} block rounded-md border-gray-300 py-2 pl-3 pr-10 text-base focus:border-violet-300 focus:outline-none focus:ring-violet-300 sm:text-sm`}
      defaultValue={defaultValue}
      {...field}
      onChange={handleChange || field.onChange}
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
