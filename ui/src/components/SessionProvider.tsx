import { createContext, useEffect, useMemo } from 'react';
import { getAuthSelf, getInfo } from '~/data/api';
import { useLocalStorage } from '~/data/hooks/storage';
import { IAuthOIDCInternal } from '~/types/auth/OIDC';

type Session = {
  required: boolean;
  authenticated: boolean;
  self?: IAuthOIDCInternal;
};

interface SessionContextType {
  session?: Session;
  setSession: (data: any) => void;
  clearSession: () => void;
}

export const SessionContext = createContext({} as SessionContextType);

export default function SessionProvider({
  children
}: {
  children: React.ReactNode;
}) {
  const [session, setSession, clearSession] = useLocalStorage('session', null);

  useEffect(() => {
    const loadSession = async () => {
      let session = {
        required: true,
        authenticated: false
      } as Session;

      try {
        await getInfo();
      } catch (err) {
        // if we can't get the info, we're not logged in
        // or there was an error, either way, clear the session so we redirect
        // to the login page
        clearSession();
        return;
      }

      try {
        const self = await getAuthSelf();
        session = {
          authenticated: true,
          required: true,
          self: self
        };
      } catch (err) {
        // if we can't get the self info and we got here then auth is likely not enabled
        // so we can just return
        session = {
          authenticated: false,
          required: false
        };
      } finally {
        if (session) {
          setSession(session);
        }
      }
    };
    if (!session) loadSession();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const value = useMemo(
    () => ({
      session,
      setSession,
      clearSession
    }),
    [session, setSession, clearSession]
  );

  return (
    <SessionContext.Provider value={value}>{children}</SessionContext.Provider>
  );
}
