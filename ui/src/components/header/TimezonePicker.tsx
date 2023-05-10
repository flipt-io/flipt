import { Menu, Transition } from '@headlessui/react';
import { GlobeAltIcon } from '@heroicons/react/24/outline';
import { Fragment } from 'react';
import { useTimezone } from '~/data/hooks/timezone';
import { classNames } from '~/utils/helpers';

export default function TimezonePicker() {
  const timezones = [
    { name: 'UTC', value: 'utc' },
    { name: 'Local', value: 'local' }
  ];

  const { timezone, setTimezone } = useTimezone();

  return (
    <Menu as="div" className="relative inline-block text-left">
      <div>
        <Menu.Button className="without-ring flex items-center rounded-full">
          <span className="sr-only">Open options</span>
          <GlobeAltIcon
            className="h-5 w-5 text-violet-200 hover:text-violet-100"
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
        <Menu.Items className="absolute right-0 z-10 mt-2 w-56 origin-top-right divide-y divide-gray-100 rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
          <div className="px-4 py-3">
            <p className="text-sm">Timezone</p>
          </div>
          <div className="py-1">
            {timezones.map((timezone) => (
              <Menu.Item key={timezone.value}>
                {({ active }) => (
                  <a
                    href="#"
                    className={classNames(
                      active ? 'bg-violet-100 text-gray-900' : 'text-gray-700',
                      'block px-4 py-2 text-sm'
                    )}
                    onClick={(e) => {
                      e.preventDefault();
                      setTimezone(timezone.value);
                    }}
                  >
                    {timezone.name}
                  </a>
                )}
              </Menu.Item>
            ))}
          </div>
        </Menu.Items>
      </Transition>
    </Menu>
  );
}
