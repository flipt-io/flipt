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
          <h3 className="text-xl font-semibold text-gray-700">Preferences</h3>
          <p className="mt-2 text-sm text-gray-500">
            Manage how information is displayed in the UI
          </p>
        </div>
        <div className="mt-6 max-w-4xl">
          <div className="divide-y divide-gray-200">
            <div className="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:pt-5">
              <label
                htmlFor="location"
                className="text-sm font-bold text-gray-500"
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
                className="text-sm font-bold text-gray-500"
                passive
              >
                UTC Timezone
                <p className="mt-2 text-xs font-normal">
                  Display dates and times in UTC timezone
                </p>
                <p className="mt-2 text-xs font-semibold">
                  {inTimezone(new Date().toISOString())}
                </p>
              </Switch.Label>
              <dd className="mt-1 flex text-sm text-gray-900 sm:col-span-2 sm:mt-0">
                <Switch
                  checked={isUTC}
                  onChange={() => {
                    dispatch(
                      timezoneChanged(isUTC ? Timezone.LOCAL : Timezone.UTC)
                    );
                  }}
                  className={cls(
                    'relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent bg-gray-200 transition-colors duration-200 ease-in-out focus:outline-none sm:ml-auto',
                    { 'bg-violet-400': isUTC }
                  )}
                >
                  <span
                    aria-hidden="true"
                    className={cls(
                      'inline-block h-5 w-5 translate-x-0 transform rounded-full bg-background shadow ring-0 transition duration-200 ease-in-out',
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
