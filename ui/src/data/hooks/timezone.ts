import { addMinutes, format, isValid, parseISO } from 'date-fns';
import { useSelector } from 'react-redux';

import { selectTimezone } from '~/app/preferences/preferencesSlice';

import { Timezone } from '~/types/Preferences';

export const useTimezone = () => {
  const timezone = useSelector(selectTimezone);

  const inTimezone = (v: string) => {
    // Handle empty or invalid strings
    if (!v || typeof v !== 'string') {
      throw new Error('Invalid datetime value');
    }
    
    try {
      const d = parseISO(v);
      
      // Check if the parsed date is valid
      if (!isValid(d)) {
        throw new Error('Invalid datetime format');
      }
      
      return timezone === Timezone.LOCAL
        ? format(d, 'yyyy-MM-dd HH:mm:ss')
        : format(addMinutes(d, d.getTimezoneOffset()), 'yyyy-MM-dd HH:mm:ss') +
            ' UTC';
    } catch (err) {
      throw new Error('Invalid datetime format');
    }
  };

  return {
    inTimezone
  };
};
