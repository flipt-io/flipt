import { Switch } from '~/components/Switch';
import { Formik } from 'formik';
import { useMemo } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import Select from '~/components/forms/Select';
import { useTimezone } from '~/data/hooks/timezone';
import { Theme, Timezone } from '~/types/Preferences';
import {
  selectLoadSegmentFlagReferences,
  selectTheme,
  selectTimezone,
  loadSegmentFlagReferencesChanged,
  themeChanged,
  timezoneChanged
} from './preferencesSlice';

export default function Preferences() {
  const timezone = useSelector(selectTimezone);
  const theme = useSelector(selectTheme);
  const loadSegmentFlagReferences = useSelector(
    selectLoadSegmentFlagReferences
  );
  const dispatch = useDispatch();

  const initialValues = {
    timezone: timezone,
    theme: theme
  };

  const { inTimezone } = useTimezone();
  const isUTC = useMemo(() => timezone === Timezone.UTC, [timezone]);

  return (
    <Formik initialValues={initialValues} onSubmit={() => {}}>
      <div className="divide-border my-10 divide-y">
        <div className="space-y-1">
          <h3 className="text-secondary-foreground text-xl font-semibold">
            Preferences
          </h3>
          <p className="text-muted-foreground mt-2 text-sm">
            Manage how information is displayed in the UI
          </p>
        </div>
        <div className="mt-6 max-w-4xl">
          <div className="divide-border divide-y">
            <div className="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:pt-5">
              <label
                htmlFor="location"
                className="text-muted-foreground text-sm font-bold"
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
            <div className="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:pt-5">
              <span
                className="text-muted-foreground text-sm font-bold"
                id="label-switch-segment-flags"
              >
                Segment flag usage
                <p className="mt-2 text-xs font-normal">
                  Load which flags reference each segment on the segment page
                  and segments list. Turning this off avoids extra API requests
                  when you have many flags.
                </p>
              </span>
              <dd className="sm:col-span-2 sm:mt-0 sm:text-right">
                <Switch
                  checked={loadSegmentFlagReferences}
                  aria-labelledby="label-switch-segment-flags"
                  onCheckedChange={() => {
                    dispatch(
                      loadSegmentFlagReferencesChanged(
                        !loadSegmentFlagReferences
                      )
                    );
                  }}
                />
              </dd>
            </div>
            <div className="py-4 sm:grid sm:grid-cols-3 sm:gap-4 sm:py-5 sm:pt-5">
              <span
                className="text-muted-foreground text-sm font-bold"
                id="label-switch-tmz"
              >
                UTC Timezone
                <p className="mt-2 text-xs font-normal">
                  Display dates and times in UTC timezone
                </p>
                <p className="mt-2 text-xs font-semibold">
                  {inTimezone(new Date().toISOString())}
                </p>
              </span>
              <dd className="sm:col-span-2 sm:mt-0 sm:text-right">
                <Switch
                  checked={isUTC}
                  aria-labelledby="label-switch-tmz"
                  onCheckedChange={() => {
                    dispatch(
                      timezoneChanged(isUTC ? Timezone.LOCAL : Timezone.UTC)
                    );
                  }}
                />
              </dd>
            </div>
          </div>
        </div>
      </div>
    </Formik>
  );
}
