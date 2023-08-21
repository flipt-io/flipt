import { Menu, Transition } from '@headlessui/react';
import { UserCircleIcon } from '@heroicons/react/24/solid';
import { Fragment } from 'react';
import { expireAuthSelf } from '~/data/api';
import { useError } from '~/data/hooks/error';
import { useSession } from '~/data/hooks/session';
import { IAuthMethodGithubMetadata } from '~/types/auth/Github';
import { IAuthMethodOIDCMetadata } from '~/types/auth/OIDC';
import { classNames } from '~/utils/helpers';

type UserProfileProps = {
  metadata?: IAuthMethodOIDCMetadata | IAuthMethodGithubMetadata;
};

export default function UserProfile(props: UserProfileProps) {
  const { metadata } = props;

  const { setError } = useError();
  const { clearSession } = useSession();

  let name: string | undefined;
  let imgURL: string | undefined;

  if (metadata) {
    if ('io.flipt.auth.github.name' in metadata) {
      name = metadata['io.flipt.auth.github.name'] ?? 'User';
      if (metadata['io.flipt.auth.github.picture']) {
        imgURL = metadata['io.flipt.auth.github.picture'];
      }
    } else if ('io.flipt.auth.oidc.name' in metadata) {
      name = metadata['io.flipt.auth.oidc.name'] ?? 'User';
      if (metadata['io.flipt.auth.oidc.picture']) {
        imgURL = metadata['io.flipt.auth.oidc.picture'];
      }
    }
  }

  const logout = async () => {
    expireAuthSelf()
      .then(() => {
        clearSession();
        window.location.href = '/';
      })
      .catch((err) => {
        setError(err);
      });
  };

  return (
    <Menu as="div" className="relative ml-3">
      <div>
        <Menu.Button className="nightwind-prevent bg-white flex max-w-xs items-center rounded-full text-sm hover:ring-2 hover:ring-white/80 focus:outline-none focus:ring-2 focus:ring-violet-500 focus:ring-offset-2">
          <span className="sr-only">Open user menu</span>
          {imgURL && (
            <img
              className="h-8 w-8 rounded-full"
              src={imgURL}
              alt={name}
              title={name}
              referrerPolicy="no-referrer"
            />
          )}
          {!imgURL && (
            <UserCircleIcon
              className="text-violet-300 h-8 w-8 rounded-full"
              aria-hidden="true"
            />
          )}
        </Menu.Button>
      </div>
      <Transition
        as={Fragment}
        enter="transition ease-out duration-100"
        enterFrom="transform opacity-0 scale-95"
        enterTo="transform opacity-100 scale-100"
        leave="transition ease-in duration-75"
        leaveFrom="transform opacity-100 scale-100"
        leaveTo="transform opacity-0 scale-95"
      >
        <Menu.Items className="bg-white absolute right-0 z-10 mt-2 w-48 origin-top-right rounded-md py-1 shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
          <Menu.Item key="logout">
            {({ active }) => (
              <a
                href="#"
                onClick={(e) => {
                  e.preventDefault();
                  logout();
                }}
                className={classNames(
                  active ? 'bg-gray-100' : '',
                  'text-gray-700 block px-4 py-2 text-sm'
                )}
              >
                Logout
              </a>
            )}
          </Menu.Item>
        </Menu.Items>
      </Transition>
    </Menu>
  );
}
