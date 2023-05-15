import moment from 'moment';
import { useContext } from 'react';
import { TimezoneContext, TimezoneType } from '~/components/TimezoneProvider';

export const useTimezone = () => {
  const { timezone, setTimezone } = useContext(TimezoneContext);
  const inTimezone = (v: string) => {
    return timezone === TimezoneType.LOCAL
      ? moment(v).format('YYYY-MM-DD HH:mm:ss')
      : moment.utc(v).format('YYYY-MM-DD HH:mm:ss');
  };
  return {
    timezone,
    setTimezone,
    inTimezone
  };
};
