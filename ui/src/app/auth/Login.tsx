import {
  IconDefinition,
  faGithub,
  faGitlab,
  faGoogle,
  faOpenid
} from '@fortawesome/free-brands-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { useEffect, useMemo } from 'react';
import { Navigate } from 'react-router';

import { useListAuthProvidersQuery } from '~/app/auth/authApi';

import Loading from '~/components/Loading';
import { Toaster } from '~/components/Sonner';

import { IAuthMethod } from '~/types/Auth';

import logoFlag from '~/assets/logo-flag.png';
import { useError } from '~/data/hooks/error';
import { useSession } from '~/data/hooks/session';
import { upperFirst } from '~/utils/helpers';
import { Button } from '~/components/ui/button';

interface ILoginProvider {
  displayName: string;
  icon: IconDefinition;
}

interface IAuthDisplay {
  name: string;
  authorize_url: string;
  icon: IconDefinition;
}

const knownProviders: Record<string, ILoginProvider> = {
  google: {
    displayName: 'Google',
    icon: faGoogle
  },
  gitlab: {
    displayName: 'GitLab',
    icon: faGitlab
  },
  auth0: {
    displayName: 'Auth0',
    icon: faOpenid
  },
  github: {
    displayName: 'GitHub',
    icon: faGithub
  }
};

function InnerLoginButtons() {
  const { setError, clearError } = useError();

  const authorize = async (uri: string) => {
    const res = await fetch(uri, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!res.ok || res.status !== 200) {
      const { message } = await res.json();
      setError('Unable to authenticate: ' + message);
      return;
    }
    clearError();
    const body = await res.json();
    window.location.href = body.authorizeUrl;
  };
  const {
    data: listAuthProviders,
    isLoading,
    error
  } = useListAuthProvidersQuery();

  useEffect(() => {
    if (error) {
      setError(error);
    }
  }, [error, setError]);

  const providers = useMemo(() => {
    return (listAuthProviders?.methods || [])
      .filter(
        (m: IAuthMethod) =>
          (m.method === 'METHOD_OIDC' || m.method === 'METHOD_GITHUB') &&
          m.enabled
      )
      .flatMap<any, IAuthDisplay>((m: IAuthMethod) => {
        if (m.method === 'METHOD_GITHUB') {
          return {
            name: 'GitHub',
            authorize_url: m.metadata.authorize_url,
            icon: faGithub
          };
        }
        if (m.method === 'METHOD_OIDC') {
          return Object.entries(m.metadata.providers).map(([k, value]) => {
            k = k.toLowerCase();
            const v = value as {
              authorize_url: string;
            };
            return {
              name: knownProviders[k]?.displayName || upperFirst(k),
              authorize_url: v.authorize_url,
              icon: knownProviders[k]?.icon || faOpenid
            };
          });
        }
      });
  }, [listAuthProviders]);

  if (isLoading) {
    return <Loading />;
  }

  return (
    <>
      {providers.length > 0 && (
        <div className="mt-6 flex flex-col space-y-3">
          {providers.map((provider) => (
            <Button
              key={provider.name}
              variant="outline"
              className="flex items-center gap-2 w-full justify-center"
              onClick={(e) => {
                e.preventDefault();
                authorize(provider.authorize_url);
              }}
            >
              <FontAwesomeIcon
                icon={provider.icon}
                className="h-5 w-5 text-muted-foreground"
                aria-hidden={true}
              />
              <span className="font-medium">With {provider.name}</span>
            </Button>
          ))}
        </div>
      )}
      {providers.length === 0 && (
        <div className="bg-background border border-muted shadow-sm rounded-lg mt-6">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-base font-semibold leading-6 text-foreground">
              No Providers
            </h3>
            <div className="mt-2 max-w-xl text-sm text-muted-foreground">
              <p>
                Authentication is set to{' '}
                <span className="font-medium">required</span>
                , however, there are no login providers configured. Please see
                the documentation for more information.
              </p>
            </div>
            <div className="mt-3 text-sm leading-6">
              <a
                href="https://www.flipt.io/docs/configuration/authentication"
                className="font-semibold text-violet-600 hover:text-violet-500 dark:text-violet-400 dark:hover:text-violet-300"
              >
                Configuring Authentication
                <span aria-hidden="true"> &rarr;</span>
              </a>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

function InnerLogin() {
  const { session } = useSession();

  if (session && (!session.required || session.authenticated)) {
    return <Navigate to="/" />;
  }

  return (
    <div className="flex min-h-screen flex-col justify-center items-center bg-background">
      <img
        src={logoFlag}
        alt="logo"
        width={64}
        height={64}
        className="h-16 w-16 mb-4 rounded-lg"
      />
      <h2 className="text-2xl font-bold text-foreground mb-2">Login to Flipt</h2>
      <InnerLoginButtons />
    </div>
  );
}

export default function Login() {
  return (
    <>
      <InnerLogin />
      <Toaster />
    </>
  );
}
