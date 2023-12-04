import { Listbox as L, Transition } from '@headlessui/react';
import { CheckIcon, ChevronUpDownIcon } from '@heroicons/react/20/solid';
import { Fragment } from 'react';
import { ISelectable } from '~/types/Selectable';
import { cls } from '~/utils/helpers';

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
    <L
      as="div"
      name={name}
      className={className}
      value={selected}
      by="key"
      onChange={(v: T) => {
        setSelected && setSelected(v);
      }}
      disabled={disabled}
    >
      {({ open }) => (
        <>
          <div className="relative mt-2">
            <L.Button
              className={cls(
                'text-gray-900 relative w-full cursor-default rounded-md px-2 py-2 pl-3 pr-10 text-left focus:outline-none sm:text-sm sm:leading-6',
                {
                  'bg-gray-100': disabled,
                  'bg-gray-50 shadow-sm ring-1 ring-inset ring-gray-300 focus:ring-1 focus:ring-violet-600':
                    !disabled
                }
              )}
              id={`${id}-select-button`}
            >
              <div className="flex items-center">
                <span className="text-gray-600 block truncate font-medium">
                  {selected?.displayValue}
                </span>
                {!disabled && (
                  <span className="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-2">
                    <ChevronUpDownIcon
                      className="text-gray-400 h-5 w-5"
                      aria-hidden="true"
                    />
                  </span>
                )}
              </div>
            </L.Button>

            <Transition
              show={open}
              as={Fragment}
              leave="transition ease-in duration-100"
              leaveFrom="opacity-100"
              leaveTo="opacity-0"
            >
              <L.Options
                className="bg-gray-50 absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-md py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm"
                id={`${id}-select-options`}
              >
                {values?.map((v) => (
                  <L.Option
                    key={v.key}
                    className={({ active }) =>
                      cls(
                        'text-gray-900 relative cursor-default select-none py-2 pl-3 pr-9',
                        {
                          'text-white bg-violet-300': active
                        }
                      )
                    }
                    value={v}
                  >
                    {({ selected, active }) => (
                      <>
                        <span
                          className={cls('block truncate font-normal', {
                            'font-semibold': selected
                          })}
                        >
                          {v.displayValue}
                        </span>

                        {selected ? (
                          <span
                            className={cls(
                              'text-violet-600 absolute inset-y-0 right-0 flex items-center pr-4',
                              {
                                'text-white': active
                              }
                            )}
                          >
                            <CheckIcon className="h-5 w-5" aria-hidden="true" />
                          </span>
                        ) : null}
                      </>
                    )}
                  </L.Option>
                ))}
              </L.Options>
            </Transition>
          </div>
        </>
      )}
    </L>
  );
}
