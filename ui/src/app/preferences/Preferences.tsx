import { Formik } from 'formik';
import { Clock, Moon, Radio, Sun } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { Switch } from '~/components/Switch';
import Select from '~/components/forms/Select';

import { Realtime, Theme, Timezone } from '~/types/Preferences';

import { useNotification } from '~/data/hooks/notification';
import { useTimezone } from '~/data/hooks/timezone';

import { PreferenceLabel, PreferenceSection } from './PreferenceSection';
import {
  realtimeChanged,
  resetLastSaved,
  selectLastSaved,
  selectRealtime,
  selectTheme,
  selectTimezone,
  themeChanged,
  timezoneChanged
} from './preferencesSlice';

export default function Preferences() {
  const timezone = useSelector(selectTimezone);
  const theme = useSelector(selectTheme);
  const realtime = useSelector(selectRealtime);
  const lastSaved = useSelector(selectLastSaved);

  const dispatch = useDispatch();
  const { setNotification } = useNotification();

  const initialValues = {
    timezone: timezone,
    theme: theme,
    realtime: realtime
  };

  const { inTimezone } = useTimezone();
  const isUTC = useMemo(() => timezone === Timezone.UTC, [timezone]);

  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null);
  const DEBOUNCE_DELAY = 1000;

  const [isInitialLoad, setIsInitialLoad] = useState(true);

  useEffect(() => {
    setIsInitialLoad(false);
  }, []);

  useEffect(() => {
    if (lastSaved && !isInitialLoad) {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }

      debounceTimerRef.current = setTimeout(() => {
        setNotification('Preferences saved', {
          description: 'Your preferences have been automatically saved.',
          duration: 2000
        });
        dispatch(resetLastSaved());
        debounceTimerRef.current = null;
      }, DEBOUNCE_DELAY);
    } else if (lastSaved && isInitialLoad) {
      dispatch(resetLastSaved());
    }

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
          <PreferenceSection title="Appearance">
            <PreferenceLabel
              icon={theme === Theme.DARK ? Moon : Sun}
              label="Theme"
              description="Choose your preferred theme for the application"
            >
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
            </PreferenceLabel>
          </PreferenceSection>

          <PreferenceSection title="Date & Time">
            <PreferenceLabel
              icon={Clock}
              label="UTC Timezone"
              description="Display dates and times in UTC timezone"
            >
              <>
                <p className="mt-2 text-xs font-medium text-gray-600 dark:text-gray-300 bg-gray-100 dark:bg-gray-800 px-2 py-1 rounded inline-block mr-4">
                  {inTimezone(new Date().toISOString())}
                </p>
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
              </>
            </PreferenceLabel>
          </PreferenceSection>

          <PreferenceSection title="Live Updates">
            <PreferenceLabel
              icon={Radio}
              label="Sync"
              description="Automatically keep flags and segments in sync with the server."
            >
              <Switch
                checked={realtime === Realtime.ON}
                aria-labelledby="label-switch-realtime"
                data-testid="switch-realtime"
                onCheckedChange={() => {
                  dispatch(
                    realtimeChanged(
                      realtime === Realtime.ON ? Realtime.OFF : Realtime.ON
                    )
                  );
                }}
              />
            </PreferenceLabel>
          </PreferenceSection>
        </div>
      </div>
    </Formik>
  );
}
