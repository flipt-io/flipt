import { Menu, Transition } from '@headlessui/react';
import { ChevronDownIcon } from '@heroicons/react/20/solid';
import { Fragment } from 'react';
import { Icon } from '~/types/Icon';
import { cls } from '~/utils/helpers';

type DropdownAction = {
  id: string;
  label: string;
  icon?: Icon;
  onClick: () => void;
  disabled?: boolean;
  title?: string;
  className?: string;
  activeClassName?: string;
  inActiveClassName?: string;
};

type DropdownProps = {
  label: string;
  actions: DropdownAction[];
};

export default function Dropdown(props: DropdownProps) {
  const { label, actions } = props;
  return (
    <Menu as="div" className="relative inline-block text-left">
      <div>
        <Menu.Button className="bg-white text-gray-700 mb-1 inline-flex w-full justify-center gap-x-1.5 rounded-md px-4 py-2 text-sm font-semibold shadow-sm ring-2 ring-gray-300 hover:bg-gray-50 focus:outline-none focus:ring-violet-300">
          {label}
          <ChevronDownIcon
            className="text-gray-400 -mr-1 h-5 w-5"
            aria-hidden="true"
          />
        </Menu.Button>
      </div>

      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items className="bg-white absolute right-0 z-10 mt-2 w-56 origin-top-right divide-y divide-gray-100 rounded-md shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
          {actions.map((action) => (
            <div className="py-1" key={action.id}>
              {!action.disabled && (
                <Menu.Item key={action.id}>
                  {({ active, close }) => (
                    <a
                      href="#"
                      className={cls(
                        'group flex items-center px-4 py-2 text-sm',
                        action.className,
                        active
                          ? (action.activeClassName ??
                              'text-gray-900 bg-gray-100')
                          : (action.inActiveClassName ?? 'text-gray-700')
                      )}
                      onClick={(e) => {
                        e.preventDefault();
                        action.onClick();
                        close();
                      }}
                    >
                      {action.icon && (
                        <action.icon
                          className="text-gray-400 mr-3 h-5 w-5 group-hover:text-gray-500"
                          aria-hidden="true"
                        />
                      )}
                      {action.label}
                    </a>
                  )}
                </Menu.Item>
              )}
              {action.disabled && (
                <Menu.Item key={action.id}>
                  {({ active }) => (
                    <span
                      className={cls(
                        'group flex items-center px-4 py-2 text-sm hover:cursor-not-allowed',
                        action.className,
                        active
                          ? (action.activeClassName ??
                              'text-gray-700 bg-gray-100')
                          : (action.inActiveClassName ?? 'text-gray-500')
                      )}
                    >
                      {action.icon && (
                        <action.icon
                          className="text-gray-300 mr-3 h-5 w-5 group-hover:text-gray-400"
                          aria-hidden="true"
                        />
                      )}
                      {action.label}
                    </span>
                  )}
                </Menu.Item>
              )}
            </div>
          ))}
        </Menu.Items>
      </Transition>
    </Menu>
  );
}
