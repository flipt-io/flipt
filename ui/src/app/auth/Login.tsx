import {
  faGitlab,
  faGoogle,
  faOpenid
} from '@fortawesome/free-brands-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { toLower, upperFirst } from 'lodash';
import { useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';
import logoFlag from '~/assets/logo-flag.png';
import { listAuthMethods } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSession } from '~/data/hooks/session';
import { IAuthMethod } from '~/types/Auth';
import { IAuthMethodOIDC } from '~/types/auth/OIDC';

interface ILoginProvider {
  displayName: string;
  icon?: any;
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
    displayName: 'Auth0'
  }
};

export default function Login() {
  const { session } = useSession();

  const [providers, setProviders] = useState<
    {
      name: string;
      authorize_url: string;
      callback_url: string;
      icon: any;
    }[]
  >([]);

  const { setError, clearError } = useError();

  const authorize = async (uri: string) => {
    const res = await fetch(uri, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!res.ok || res.status !== 200) {
      setError(new Error('Unable to authenticate: ' + res.text()));
      return;
    }

    clearError();
    const body = await res.json();
    window.location.href = body.authorizeUrl;
  };

  useEffect(() => {
    const loadProviders = async () => {
      try {
        const resp = await listAuthMethods();
        // TODO: support alternative auth methods
        const authOIDC = resp.methods.find(
          (m: IAuthMethod) => m.method === 'METHOD_OIDC' && m.enabled
        ) as IAuthMethodOIDC;

        if (!authOIDC) {
          return;
        }

        const loginProviders = Object.entries(authOIDC.metadata.providers).map(
          ([k, v]) => {
            k = toLower(k);
            return {
              name: knownProviders[k]?.displayName || upperFirst(k), // if we dont know the provider, just capitalize the first letter
              authorize_url: v.authorize_url,
              callback_url: v.callback_url,
              icon: knownProviders[k]?.icon || faOpenid // if we dont know the provider icon, use the openid icon
            };
          }
        );
        setProviders(loginProviders);
      } catch (err) {
        setError(err instanceof Error ? err : Error(String(err)));
      }
    };

    loadProviders();
  }, [setProviders, setError]);

  if (session && (!session.required || session.authenticated)) {
    return <Navigate to="/" />;
  }

  return (
    <>
      <div className="flex min-h-screen flex-col justify-center sm:px-6 lg:px-8">
        <main className="flex px-6 py-10">
          <div className="w-full overflow-x-auto px-4 sm:px-6 lg:px-8">
            <div className="sm:mx-auto sm:w-full sm:max-w-md">
              <img
                src={logoFlag}
                alt="logo"
                width={512}
                height={512}
                className="m-auto h-20 w-auto"
              />
              <h2 className="mt-6 text-center text-3xl font-bold tracking-tight text-gray-900">
                Login to Flipt
              </h2>
            </div>

            <div className="mt-8 sm:mx-auto sm:w-full sm:max-w-sm">
              <div className="px-4 py-8 sm:px-10">
                <div className="mt-6 flex flex-col space-y-5">
                  {providers.map((provider) => (
                    <div key={provider.name}>
                      <a
                        href="#"
                        className="inline-flex w-full justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-500 shadow-sm hover:text-violet-500 hover:shadow-violet-300"
                        onClick={(e) => {
                          e.preventDefault();
                          authorize(provider.authorize_url);
                        }}
                      >
                        <span className="sr-only">
                          Sign in with {provider.name}
                        </span>
                        <FontAwesomeIcon
                          icon={provider.icon}
                          className="text-gray h-5 w-5"
                          aria-hidden={true}
                        />
                        <span className="ml-2">With {provider.name}</span>
                      </a>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </main>
      </div>
    </>
  );
}
