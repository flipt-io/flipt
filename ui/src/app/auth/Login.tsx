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
      <div className="flex min-h-screen flex-col justify-center bg-white sm:px-6 lg:px-8">
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
            <div className="mt-8 max-w-sm sm:mx-auto sm:w-full md:max-w-lg">
              <div className="px-4 py-8 sm:px-10">
                {providers && providers.length > 0 && (
                  <div className="mt-6 flex flex-col space-y-5">
                    {providers.map((provider) => (
                      <div key={provider.name}>
                        <a
                          href="#"
                          className="inline-flex w-full justify-center rounded-md border px-4 py-2 text-sm font-medium shadow-sm bg-white text-gray-500 border-gray-300 hover:shadow-violet-300 hover:text-violet-500"
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
                )}
                {(!providers || providers.length === 0) && (
                  <div className="shadow bg-white sm:rounded-lg">
                    <div className="px-4 py-5 sm:p-6">
                      <h3 className="text-base font-semibold leading-6 text-gray-900">
                        No Providers
                      </h3>
                      <div className="mt-2 max-w-xl text-sm text-gray-500">
                        <p>
                          Authentication is set to{' '}
                          <span className="font-medium">required</span>,
                          however, there are no login providers configured.
                          Please see the documentation for more information.
                        </p>
                      </div>
                      <div className="mt-3 text-sm leading-6">
                        <a
                          href="https://www.flipt.io/docs/configuration/authentication#method-oidc"
                          className="font-semibold text-violet-600 hover:text-violet-500"
                        >
                          Configuring Authentication
                          <span aria-hidden="true"> &rarr;</span>
                        </a>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </main>
      </div>
    </>
  );
}
