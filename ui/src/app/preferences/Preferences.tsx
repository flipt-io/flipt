import { Formik } from 'formik';
import { Clock, Moon, Sun } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Switch } from '~/components/Switch';
import Select from '~/components/forms/Select';

import { Theme, Timezone } from '~/types/Preferences';

import { useNotification } from '~/data/hooks/notification';
import { useTimezone } from '~/data/hooks/timezone';

import {
  resetLastSaved,
  selectLastSaved,
  selectTheme,
  selectTimezone,
  themeChanged,
  timezoneChanged
} from './preferencesSlice';

export default function Preferences() {
  const timezone = useSelector(selectTimezone);
  const theme = useSelector(selectTheme);
  const lastSaved = useSelector(selectLastSaved);

  const dispatch = useDispatch();
  const { setNotification } = useNotification();

  const initialValues = {
    timezone: timezone,
    theme: theme
  };

  const { inTimezone } = useTimezone();
  const isUTC = useMemo(() => timezone === Timezone.UTC, [timezone]);

  // Debounce timer for notifications
  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null);
  const DEBOUNCE_DELAY = 1000; // 1 second

  // Track if this is the initial load
  const [isInitialLoad, setIsInitialLoad] = useState(true);

  // Set initial load state to false after component mounts
  useEffect(() => {
    setIsInitialLoad(false);
  }, []);

  // Show toast when preferences are saved, with debounce
  useEffect(() => {
    if (lastSaved && !isInitialLoad) {
      // Clear any existing timer
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }

      // Set new timer
      debounceTimerRef.current = setTimeout(() => {
        setNotification('Preferences saved', {
          description: 'Your preferences have been automatically saved.',
          duration: 2000
        });
        dispatch(resetLastSaved());
        debounceTimerRef.current = null;
      }, DEBOUNCE_DELAY);
    } else if (lastSaved && isInitialLoad) {
      // Just reset the lastSaved without showing notification on initial load
      dispatch(resetLastSaved());
    }

    // Cleanup timer on unmount
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, [lastSaved, dispatch, setNotification, isInitialLoad]);

  return (
    <Formik initialValues={initialValues} onSubmit={() => {}}>
      <div className="my-6">
        <div className="max-w-2xl">
          {/* Appearance Section */}
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6 mb-6">
            <h3 className="text-lg font-medium text-gray-800 dark:text-gray-100 mb-4">
              Appearance
            </h3>

            <div className="space-y-6">
              {/* Theme Preference */}
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between">
                <div className="mb-3 sm:mb-0">
                  <div className="flex items-center">
                    {theme === Theme.DARK ? (
                      <Moon className="h-5 w-5 text-violet-500 dark:text-violet-400 mr-2" />
                    ) : (
                      <Sun className="h-5 w-5 text-violet-500 dark:text-violet-400 mr-2" />
                    )}
                    <label
                      htmlFor="theme"
                      className="text-sm font-medium text-gray-700 dark:text-gray-300"
                    >
                      Theme
                    </label>
                  </div>
                  <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    Choose your preferred theme for the application
                  </p>
                </div>
                <Select
                  id="theme"
                  name="theme"
                  value={theme || Theme.SYSTEM}
                  options={[
                    { value: Theme.LIGHT, label: 'Light' },
                    { value: Theme.DARK, label: 'Dark' },
                    { value: Theme.SYSTEM, label: 'System' }
                  ]}
                  data-testid="select-theme"
                  onChange={(e) => {
                    dispatch(themeChanged(e.target.value as Theme));
                  }}
                  className="w-full sm:w-48"
                />
              </div>
            </div>
          </div>

          {/* DateTime Section */}
          <div className="bg-white dark:bg-gray-900 rounded-lg shadow-sm border border-gray-200 dark:border-gray-700 p-6">
            <h3 className="text-lg font-medium text-gray-800 dark:text-gray-100 mb-4">
              Date & Time
            </h3>

            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between">
              <div className="mb-3 sm:mb-0">
                <div className="flex items-center">
                  <Clock className="h-5 w-5 text-violet-500 dark:text-violet-400 mr-2" />
                  <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
                    UTC Timezone
                  </span>
                </div>
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                  Display dates and times in UTC timezone
                </p>
                <p className="mt-2 text-xs font-medium text-gray-600 dark:text-gray-300 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded inline-block">
                  {inTimezone(new Date().toISOString())}
                </p>
              </div>
              <Switch
                checked={isUTC}
                aria-labelledby="label-switch-tmz"
                data-testid="switch-timezone"
                onCheckedChange={() => {
                  dispatch(
                    timezoneChanged(isUTC ? Timezone.LOCAL : Timezone.UTC)
                  );
                }}
              />
            </div>
          </div>
        </div>
      </div>
    </Formik>
  );
}
