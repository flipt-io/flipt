import { Listbox, Transition } from '@headlessui/react';
import { ChevronDownIcon } from '@heroicons/react/20/solid';
import { FolderIcon } from '@heroicons/react/24/outline';
import { Fragment } from 'react';
import { useSelector } from 'react-redux';
import { useLocation, useNavigate } from 'react-router';

import {
  currentNamespaceChanged,
  selectCurrentNamespace,
  selectNamespaces
} from '~/app/namespaces/namespacesApi';

import { INamespace } from '~/types/Namespace';
import { ISelectable } from '~/types/Selectable';

import { useAppDispatch } from '~/data/hooks/store';
import { addNamespaceToPath } from '~/utils/helpers';
import { cls } from '~/utils/helpers';

export type SelectableNamespace = Pick<INamespace, 'key' | 'name'> &
  ISelectable;

type NamespaceListboxProps = {
  disabled: boolean;
  className?: string;
};

export default function NamespaceListbox(props: NamespaceListboxProps) {
  const { disabled, className } = props;
  const namespace = useSelector(selectCurrentNamespace);
  const dispatch = useAppDispatch();
  const namespaces = useSelector(selectNamespaces);
  const location = useLocation();
  const navigate = useNavigate();

  const setCurrentNamespace = (namespace: INamespace) => {
    dispatch(currentNamespaceChanged(namespace));
    const newPath = addNamespaceToPath(location.pathname, namespace.key);
    navigate(newPath);
  };

  return (
    <Listbox
      as="div"
      value={namespace}
      by="key"
      onChange={setCurrentNamespace}
      disabled={disabled || namespaces.length <= 1}
      className={className}
    >
      {({ open }) => (
        <>
          <div className="relative">
            <Listbox.Button className="group font-medium flex w-full items-center rounded p-2 text-sm text-gray-900 bg-white hover:bg-gray-100 md:bg-transparent dark:text-white dark:bg-gray-800 md:dark:bg-transparent dark:hover:bg-gray-300">
              <FolderIcon
                className="mr-3 h-6 w-6 flex-shrink-0 text-white md:text-gray-500"
                aria-hidden="true"
              />
              <span className="flex-1 text-left">
                {namespace?.name || 'Select Namespace'}
              </span>
              {!disabled && namespaces.length > 1 && (
                <ChevronDownIcon
                  className={cls(
                    'h-4 w-4 text-gray-400 transition-transform duration-200 group-hover:text-gray-600 dark:group-hover:text-gray-300',
                    open ? 'rotate-180' : ''
                  )}
                  aria-hidden="true"
                />
              )}
            </Listbox.Button>

            <Transition
              show={open}
              as={Fragment}
              leave="transition ease-in duration-100"
              leaveFrom="opacity-100"
              leaveTo="opacity-0"
            >
              <Listbox.Options className="absolute right-0 z-50 mt-1 max-h-60 w-full min-w-[160px] overflow-auto rounded-md bg-white dark:bg-gray-800 py-1 text-base shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                {namespaces.map((ns) => (
                  <Listbox.Option
                    key={ns.key}
                    className={({ active }) =>
                      cls(
                        'relative cursor-default select-none px-4 py-2',
                        active
                          ? 'bg-gray-100 text-gray-900 dark:bg-violet-600 dark:text-white'
                          : 'text-gray-700 dark:text-gray-100'
                      )
                    }
                    value={ns}
                  >
                    {({ selected }) => (
                      <span
                        className={cls(
                          'block truncate',
                          selected ? 'font-medium' : ''
                        )}
                      >
                        {ns.name}
                      </span>
                    )}
                  </Listbox.Option>
                ))}
              </Listbox.Options>
            </Transition>
          </div>
        </>
      )}
    </Listbox>
  );
}
