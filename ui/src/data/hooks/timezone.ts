import { useContext } from 'react';
import { TimezoneContext } from '~/components/TimezoneProvider';

export const useTimezone = () => {
  const { timezone, setTimezone } = useContext(TimezoneContext);
  return { timezone, setTimezone };
};
