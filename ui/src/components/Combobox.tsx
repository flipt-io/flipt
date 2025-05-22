import { type VariantProps } from 'class-variance-authority';
import { Check, ChevronsUpDown, LucideIcon } from 'lucide-react';
import { useRef, useState } from 'react';

import { Button, buttonVariants } from '~/components/Button';
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
  onInputChange?: (value: string) => void;
  icon?: LucideIcon;
};

export default function Combobox<T extends ISelectable>(
  props: ComboboxProps<T> & VariantProps<typeof buttonVariants>
) {
  const {
    id,
    name,
    className,
    values,
    selected,
    setSelected,
    placeholder,
    disabled,
    onInputChange,
    icon
  } = props;

  const Icon = icon;
  const ref = useRef(null);
  const [openOptions, setOpenOptions] = useState(false);
  return (
    <Popover open={openOptions} onOpenChange={setOpenOptions}>
      <PopoverTrigger asChild disabled={disabled}>
        <Button
          ref={ref}
          name={name + '-select-button'}
          id={id + '-select-button'}
          data-testid={name + '-select-button'}
          variant={props.variant || 'secondaryline'}
          size={props.size}
          aria-expanded={openOptions}
          className={cls('justify-between gap-2', className, {
            'text-muted-foreground': !selected
          })}
        >
          {Icon && <Icon className="h-4 w-4 text-muted-foreground" />}
          {selected ? selected.displayValue : placeholder}
          <ChevronsUpDown className="text-muted-foreground h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent
        aria-labelledby={name + '-select-button'}
        className="p-0"
        style={{ minWidth: 'var(--radix-popover-trigger-width)' }}
        ref={ref}
      >
        <Command>
          <CommandInput
            placeholder="Search ..."
            className="h-8 border-0 ring-0"
            onValueChange={onInputChange}
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
