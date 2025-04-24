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

export type SelectableEnvironment = Pick<IEnvironment, 'key'> & ISelectable;

type EnvironmentListboxProps = {
  className?: string;
};

export default function EnvironmentListbox(props: EnvironmentListboxProps) {
  const { className } = props;
  const environment = useSelector(selectCurrentEnvironment);
  const dispatch = useAppDispatch();
  const environments = useSelector(selectEnvironments);

  // Sort environments to show default first
  const sortedEnvironments = [...environments].sort((a, b) => {
    if (a.default) return -1;
    if (b.default) return 1;
    return a.key.localeCompare(b.key);
  });

  return (
    <Listbox
      as="div"
      value={environment}
      by="key"
      onChange={(env) => dispatch(currentEnvironmentChanged(env))}
      disabled={environments.length <= 1}
      data-testid="environment-listbox"
    >
      <div className={cls('relative', className)}>
        <Listbox.Button className="group flex items-center gap-1 rounded px-2 py-1 text-sm text-white hover:bg-white/10 uppercase focus:outline-none">
          <span>{environment?.key || 'Unknown Environment'}</span>
          {environments.length > 1 && (
            <ChevronDownIcon
              className="h-4 w-4 text-gray-400 transition-transform duration-200 group-hover:text-white ui-open:rotate-180"
              aria-hidden="true"
            />
          )}
        </Listbox.Button>

        <Transition
          as={Fragment}
          leave="transition ease-in duration-100"
          leaveFrom="opacity-100"
          leaveTo="opacity-0"
        >
          <Listbox.Options className="absolute right-0 z-50 mt-1 max-h-60 w-full min-w-[160px] overflow-auto rounded-lg bg-white dark:bg-gray-100 py-1 text-sm shadow-lg focus:outline-none">
            {sortedEnvironments.map((env) => (
              <Listbox.Option
                key={env.key}
                className={({ active }) =>
                  cls(
                    'relative cursor-pointer select-none px-3 py-2',
                    active
                      ? 'bg-violet-500 text-white dark:bg-violet-400 dark:text-gray-100'
                      : 'text-gray-900 dark:text-gray-900'
                  )
                }
                value={env}
              >
                {({ selected }) => (
                  <span
                    className={cls(
                      'block truncate',
                      selected ? 'font-semibold' : 'font-normal'
                    )}
                  >
                    {env.key}
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
