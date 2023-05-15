import { Switch } from '@headlessui/react';
import { useState } from 'react';
import { TimezoneType } from '~/components/TimezoneProvider';
import { useTimezone } from '~/data/hooks/timezone';
import { classNames } from '~/utils/helpers';

export default function Preferences() {
  const { timezone, setTimezone } = useTimezone();
  const [utcTimezoneEnabled, setUtcTimezoneEnabled] = useState(
    timezone === TimezoneType.UTC
  );

  return (
    <div className="my-10 divide-y divide-gray-200">
      <div className="space-y-1">
        <h1 className="text-xl font-semibold text-gray-700">Preferences</h1>
        <p className="mt-2 text-sm text-gray-500">
          Manage how information is displayed in the UI
        </p>
      </div>
      <div className="mt-6">
        <dl className="divide-y divide-gray-200">
          <Switch.Group
            as="div"
            className="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:pt-5"
          >
            <Switch.Label
              as="dt"
              className="text-sm font-medium text-gray-500"
              passive
            >
              UTC Timezone
              <p className="mt-2 text-sm text-gray-400">
                Display dates and times in UTC timezone
              </p>
            </Switch.Label>
            <dd className="mt-1 flex text-sm text-gray-900 sm:col-span-2 sm:mt-0">
              <Switch
                checked={utcTimezoneEnabled}
                onChange={() => {
                  setUtcTimezoneEnabled(!utcTimezoneEnabled);
                  setTimezone(
                    utcTimezoneEnabled ? TimezoneType.LOCAL : TimezoneType.UTC
                  );
                }}
                className={classNames(
                  utcTimezoneEnabled ? 'bg-purple-600' : 'bg-gray-200',
                  'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none sm:ml-auto'
                )}
              >
                <span
                  aria-hidden="true"
                  className={classNames(
                    utcTimezoneEnabled ? 'translate-x-5' : 'translate-x-0',
                    'inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out'
                  )}
                />
              </Switch>
            </dd>
          </Switch.Group>
        </dl>
      </div>
    </div>
  );
}
