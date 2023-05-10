import { createContext, useState } from 'react';
import { useLocalStorage } from '~/data/hooks/storage';

interface TimezoneContextType {
  timezone?: string;
  setTimezone: (data: string) => void;
}

export const TimezoneContext = createContext({} as TimezoneContextType);

export default function TimezoneProvider({
  children
}: {
  children: React.ReactNode;
}) {
  const [_, setTimezoneStorage] = useLocalStorage('timezone', 'local');
  const [tz, setTz] = useState('local');

  const setTimezone = (data: string) => {
    setTimezoneStorage(data);
    setTz(data);
  };

  const value = { timezone: tz, setTimezone };

  return (
    <TimezoneContext.Provider value={value}>
      {children}
    </TimezoneContext.Provider>
  );
}
