import { createContext } from 'react';
import { useLocalStorage } from '~/data/hooks/storage';
import { Preferences, TimezoneType } from '~/types/Preferences';

interface PreferencesContextType {
  preferences: Preferences;
  setPreferences: (data: Preferences) => void;
}

export const PreferencesContext = createContext({} as PreferencesContextType);

export default function PreferencesProvider({
  children
}: {
  children: React.ReactNode;
}) {
  const [preferences, setPreferences] = useLocalStorage('preferences', {
    timezone: TimezoneType.LOCAL
  } as Preferences);

  return (
    <PreferencesContext.Provider
      value={{
        preferences,
        setPreferences
      }}
    >
      {children}
    </PreferencesContext.Provider>
  );
}
