import { useField } from 'formik';
import { Check, ChevronsUpDown } from 'lucide-react';
import { useState } from 'react';

import { Button } from '~/components/Button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList
} from '~/components/Command';
import { Popover, PopoverContent, PopoverTrigger } from '~/components/Popover';

import { ISelectable } from '~/types/Selectable';

import { cls } from '~/utils/helpers';

type ComboboxProps<T extends ISelectable> = {
  id: string;
  name: string;
  placeholder?: string;
  values?: T[];
  selected: T | null;
  setSelected?: (v: T | null) => void;
  disabled?: boolean;
  className?: string;
};

export default function Combobox<T extends ISelectable>(
  props: ComboboxProps<T>
) {
  const {
    id,
    name,
    className,
    values,
    selected,
    setSelected,
    placeholder,
    disabled
  } = props;

  const [field] = useField(props);
  const [openOptions, setOpenOptions] = useState(false);
  return (
    <Popover open={openOptions} onOpenChange={setOpenOptions}>
      <PopoverTrigger asChild disabled={disabled}>
        <Button
          name={name + '-select-button'}
          id={id + '-select-button'}
          data-testid={name + '-select-button'}
          variant="outline"
          aria-expanded={openOptions}
          className={cls(
            'w-full justify-between border-gray-300 shadow-xs border mt-1 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-100 dark:focus:ring-violet-500 px-3',
            className,
            { 'text-muted-foreground dark:text-muted-foreground': !selected }
          )}
        >
          {selected ? selected.displayValue : placeholder}
          <ChevronsUpDown className="text-muted-foreground h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        aria-labelledby={name + '-select-button'}
        className="p-0"
        style={{ width: 'var(--radix-popover-trigger-width)' }}
      >
        <Command>
          <CommandInput
            placeholder="Search ..."
            className="h-8 border-0 ring-0"
          />
          <CommandList>
            <CommandEmpty className="py-2 text-center text-sm text-muted-foreground">
              No results found.
            </CommandEmpty>

            <CommandGroup>
              {values?.map((item) => (
                <CommandItem
                  key={item.key}
                  value={item.key}
                  aria-role="option"
                  aria-label={item.displayValue}
                  onSelect={(key) => {
                    const v = values?.find((i) => i.key == key) || item;
                    if (v) {
                      setSelected && setSelected(v);
                      field.onChange({ target: { value: v?.key, id } });
                      setOpenOptions(false);
                    }
                  }}
                >
                  <div className="flex items-center">
                    {item?.status && (
                      <span
                        className={cls(
                          'mr-3 inline-block h-2 w-2 shrink-0 rounded-full bg-gray-400',
                          {
                            'bg-green-400 data-[selected=true]:bg-green-600':
                              item.status === 'active'
                          }
                        )}
                        aria-hidden="true"
                      />
                    )}
                    <span
                      className={cls('truncate text-muted-foreground', {
                        'text-foreground': selected?.key === item.key
                      })}
                    >
                      {item?.displayValue}
                    </span>
                  </div>
                  <Check
                    className={cls(
                      'ml-auto text-foreground',
                      selected?.key === item.key ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
