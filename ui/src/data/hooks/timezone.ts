import moment from 'moment';
import { useContext } from 'react';
import { PreferencesContext } from '~/components/PreferencesProvider';
import { TimezoneType } from '~/types/Preferences';

export const useTimezone = () => {
  const { preferences, setPreferences } = useContext(PreferencesContext);
  const timezone = preferences.timezone;

  const setTimezone = (v: TimezoneType) => {
    setPreferences({
      ...preferences,
      timezone: v
    });
  };

  const inTimezone = (v: string) => {
    return timezone === TimezoneType.LOCAL
      ? moment(v).format('YYYY-MM-DD HH:mm:ss')
      : moment.utc(v).format('YYYY-MM-DD HH:mm:ss') + ' UTC';
  };

  return {
    timezone,
    setTimezone,
    inTimezone
  };
};
