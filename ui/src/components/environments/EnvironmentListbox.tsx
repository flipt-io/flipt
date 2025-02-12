import { Listbox, Transition } from '@headlessui/react';
import { ChevronDownIcon } from '@heroicons/react/20/solid';
import { Fragment } from 'react';
import { useSelector } from 'react-redux';

import {
  currentEnvironmentChanged,
  selectCurrentEnvironment,
  selectEnvironments
} from '~/app/environments/environmentsApi';

import { IEnvironment } from '~/types/Environment';
import { ISelectable } from '~/types/Selectable';

import { useAppDispatch } from '~/data/hooks/store';
import { cls } from '~/utils/helpers';

export type SelectableEnvironment = Pick<IEnvironment, 'name'> & ISelectable;

type EnvironmentListboxProps = {
  className?: string;
};

export default function EnvironmentListbox(props: EnvironmentListboxProps) {
  const { className } = props;
  const environment = useSelector(selectCurrentEnvironment);
  const dispatch = useAppDispatch();
  const environments = useSelector(selectEnvironments);

  // If there's only one environment, render a simple div instead of a listbox
  if (environments.length <= 1) {
    return (
      <div className={cls('px-2 uppercase py-1 text-sm text-white', className)}>
        {environment?.name || 'No Environment'}
      </div>
    );
  }

  return (
    <Listbox
      value={environment}
      onChange={(env) => dispatch(currentEnvironmentChanged(env))}
    >
      <div className={cls('relative', className)}>
        <Listbox.Button className="group flex w-full items-center gap-1 rounded px-2 py-1 text-sm text-white hover:bg-white/10 uppercase">
          <span>{environment?.name || 'Select Environment'}</span>
          <ChevronDownIcon
            className="h-4 w-4 text-gray-400 transition-transform duration-200 group-hover:text-white ui-open:rotate-180"
            aria-hidden="true"
          />
        </Listbox.Button>

        <Transition
          as={Fragment}
          leave="transition ease-in duration-100"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <Listbox.Options className="absolute right-0 z-50 mt-1 max-h-60 w-full min-w-[160px] overflow-auto rounded-md bg-gray-900 py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
            {environments.map((env) => (
              <Listbox.Option
                key={env.name}
                className={({ active }) =>
                  cls(
                    'relative cursor-default select-none px-4 py-2',
                    active ? 'bg-gray-800 text-white' : 'text-gray-300'
                  )
                }
                value={env}
              >
                {({ selected }) => (
                  <span
                    className={cls(
                      'block truncate',
                      selected ? 'font-medium text-white' : ''
                    )}
                  >
                    {env.name}
                  </span>
                )}
              </Listbox.Option>
            ))}
          </Listbox.Options>
        </Transition>
      </div>
    </Listbox>
  );
}
