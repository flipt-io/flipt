import { addMinutes, format, parseISO } from 'date-fns';
import { useSelector } from 'react-redux';
import { selectTimezone } from '~/app/preferences/preferencesSlice';
import { Timezone } from '~/types/Preferences';

export const useTimezone = () => {
  const timezone = useSelector(selectTimezone);

  const inTimezone = (v: string) => {
    const d = parseISO(v);
    return timezone === Timezone.LOCAL
      ? format(d, 'yyyy-MM-dd HH:mm:ss')
      : format(addMinutes(d, d.getTimezoneOffset()), 'yyyy-MM-dd HH:mm:ss') +
          ' UTC';
  };

  return {
    inTimezone
  };
};
