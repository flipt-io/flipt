import { createContext } from 'react';
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
  const [timezone, setTimezone] = useLocalStorage('timezone', 'local');

  return (
    <TimezoneContext.Provider
      value={{
        timezone,
        setTimezone
      }}
    >
      {children}
    </TimezoneContext.Provider>
  );
}
