import { createContext, useEffect, useMemo } from 'react';
import { getAuthSelf, getConfig, getInfo } from '~/data/api';
import { useLocalStorage } from '~/data/hooks/storage';
import { IAuthGithubInternal } from '~/types/auth/Github';
import { IAuthJWTInternal } from '~/types/auth/JWT';
import { IAuthOIDCInternal } from '~/types/auth/OIDC';

type Session = {
  required: boolean;
  authenticated: boolean;
  self?: IAuthOIDCInternal | IAuthGithubInternal | IAuthJWTInternal;
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
    const clearSessionIfNecessary = async () => {
      const config = await getConfig();
      if (session && session.required !== config.authentication.required) {
        clearSession();
      }
    };

    const loadSession = async () => {
      try {
        await getInfo();
      } catch (err) {
        // if we can't get the info, we're not logged in
        // or there was an error, either way, clear the session so we redirect
        // to the login page
        clearSession();
        return;
      }

      if (session) {
        clearSessionIfNecessary();
        if (session) {
          return;
        }
      }

      const newSession = {
        required: true,
        authenticated: false
      } as Session;

      try {
        const self: IAuthOIDCInternal | IAuthGithubInternal | IAuthJWTInternal =
          await getAuthSelf();
        newSession.authenticated = true;
        newSession.self = self;
      } catch (err) {
        // if we can't get the self info and we got here then auth is likely not enabled
        // so we can just return
        newSession.required = false;
      } finally {
        if (newSession) {
          setSession(newSession);
        }
      }
    };

    loadSession();
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
