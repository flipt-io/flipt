import moment from 'moment';
import { useSelector } from 'react-redux';
import { selectTimezone } from '~/app/preferences/preferencesSlice';
import { Timezone } from '~/types/Preferences';

export const useTimezone = () => {
  const timezone = useSelector(selectTimezone);

  const inTimezone = (v: string) => {
    return timezone === Timezone.LOCAL
      ? moment(v).format('YYYY-MM-DD HH:mm:ss')
      : moment.utc(v).format('YYYY-MM-DD HH:mm:ss') + ' UTC';
  };

  return {
    inTimezone
  };
};
