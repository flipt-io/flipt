import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '~/components/Select';

import { ISelectable } from '~/types/Selectable';

type ListBoxProps<T extends ISelectable> = {
  id: string;
  name: string;
  values?: T[];
  selected: T;
  setSelected?: (v: T) => void;
  disabled?: boolean;
  className?: string;
};

export default function Listbox<T extends ISelectable>(props: ListBoxProps<T>) {
  const { id, name, values, selected, setSelected, disabled, className } =
    props;

  return (
    <Select
      name={name}
      disabled={disabled}
      defaultValue={selected?.key}
      onValueChange={(key) => {
        if (setSelected) {
          const value = values?.find((el) => el.key == key) as T;
          value && setSelected(value);
        }
      }}
    >
      <SelectTrigger id={`${id}-select-button`} className={className}>
        <SelectValue placeholder="Select an option" />
      </SelectTrigger>
      <SelectContent>
        <SelectGroup>
          {values?.map((v) => (
            <SelectItem key={v.key} value={v.key}>
              {v.displayValue}
            </SelectItem>
          ))}
        </SelectGroup>
      </SelectContent>
    </Select>
  );
}
