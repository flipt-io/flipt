import { Switch } from '@headlessui/react';
import { Formik } from 'formik';
import { useMemo } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import Select from '~/components/forms/Select';
import { useTimezone } from '~/data/hooks/timezone';
import { Theme, Timezone } from '~/types/Preferences';
import { cls } from '~/utils/helpers';
import {
  selectTheme,
  selectTimezone,
  themeChanged,
  timezoneChanged
} from './preferencesSlice';

export default function Preferences() {
  const timezone = useSelector(selectTimezone);
  const theme = useSelector(selectTheme);

  const dispatch = useDispatch();

  const initialValues = {
    timezone: timezone,
    theme: theme
  };

  const { inTimezone } = useTimezone();
  const isUTC = useMemo(() => timezone === Timezone.UTC, [timezone]);

  return (
    <Formik initialValues={initialValues} onSubmit={() => {}}>
      <div className="my-10 divide-y divide-gray-200">
        <div className="space-y-1">
          <h3 className="text-gray-700 text-xl font-semibold">Preferences</h3>
          <p className="text-gray-500 mt-2 text-sm">
            Manage how information is displayed in the UI
          </p>
        </div>
        <div className="mt-6 max-w-4xl">
          <div className="divide-y divide-gray-200">
            <div className="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:pt-5">
              <label
                htmlFor="location"
                className="text-gray-500 text-sm font-medium"
              >
                Theme
              </label>
              <span></span>
              <Select
                id="location"
                name="location"
                value={theme || Theme.SYSTEM}
                options={[
                  { value: Theme.LIGHT, label: 'Light' },
                  { value: Theme.DARK, label: 'Dark' },
                  { value: Theme.SYSTEM, label: 'System' }
                ]}
                onChange={(e) => {
                  dispatch(themeChanged(e.target.value as Theme));
                }}
              />
            </div>
            <Switch.Group
              as="div"
              className="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:pt-5"
            >
              <Switch.Label
                as="span"
                className="text-gray-500 text-sm font-medium"
                passive
              >
                UTC Timezone
                <p className="text-gray-400 mt-2 text-sm">
                  Display dates and times in UTC timezone
                </p>
                <p className="text-gray-400 mt-2 text-sm">
                  {inTimezone(new Date().toISOString())}
                </p>
              </Switch.Label>
              <dd className="text-gray-900 mt-1 flex text-sm sm:col-span-2 sm:mt-0">
                <Switch
                  checked={isUTC}
                  onChange={() => {
                    dispatch(
                      timezoneChanged(isUTC ? Timezone.LOCAL : Timezone.UTC)
                    );
                  }}
                  className={cls(
                    'bg-gray-200 relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none sm:ml-auto',
                    { 'bg-violet-400': isUTC }
                  )}
                >
                  <span
                    aria-hidden="true"
                    className={cls(
                      'bg-white inline-block h-5 w-5 translate-x-0 transform rounded-full shadow ring-0 transition duration-200 ease-in-out',
                      {
                        'translate-x-5': isUTC
                      }
                    )}
                  />
                </Switch>
              </dd>
            </Switch.Group>
          </div>
        </div>
      </div>
    </Formik>
  );
}
