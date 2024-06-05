import { Menu, Transition } from '@headlessui/react';
import { UserCircleIcon } from '@heroicons/react/24/solid';
import { Fragment } from 'react';
import { useError } from '~/data/hooks/error';
import { useSession } from '~/data/hooks/session';
import { User } from '~/types/auth/User';
import { cls } from '~/utils/helpers';
import { useNavigate } from 'react-router-dom';
import { expireAuthSelf } from '~/data/api';

type UserProfileProps = {
  user: User;
};

export default function UserProfile(props: UserProfileProps) {
  const { user } = props;
  const { setError } = useError();
  const { clearSession } = useSession();

  const navigate = useNavigate();
  const logout = async () => {
    try {
      await expireAuthSelf();
      clearSession();
      if (user?.issuer) {
        window.location.href = `//${user.issuer}`;
      } else {
        navigate('/login');
      }
    } catch (err) {
      setError(err);
    }
  };

  return (
    <Menu as="div" className="relative ml-3">
      <div>
        <Menu.Button className="nightwind-prevent bg-white flex max-w-xs items-center rounded-full text-sm ring-1 ring-white hover:ring-2 hover:ring-violet-500/80 focus:outline-none focus:ring-1 focus:ring-violet-500 focus:ring-offset-2">
          <span className="sr-only">Open user menu</span>
          {user.imgURL && (
            <img
              className="h-7 w-7 rounded-full"
              src={user.imgURL}
              alt={user.name}
              title={user.name}
              referrerPolicy="no-referrer"
            />
          )}
          {!user.imgURL && (
            <UserCircleIcon
              className="nightwind-prevent text-gray-800 h-6 w-6 rounded-full"
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
          {(user.name || user.login) && (
            <Menu.Item disabled>
              {({ active }) => (
                <span
                  className={cls(
                    'border-gray-200 flex flex-col border-b px-4 py-2',
                    { 'bg-gray-100': active }
                  )}
                >
                  <span className="text-gray-600 flex-1 text-sm">
                    {user.name}
                  </span>
                  {user.login && (
                    <span className="text-gray-400 text-xs">{user.login}</span>
                  )}
                </span>
              )}
            </Menu.Item>
          )}
          <Menu.Item key="logout">
            {({ active }) => (
              <a
                href="#"
                onClick={(e) => {
                  e.preventDefault();
                  logout();
                }}
                className={cls('text-gray-700 block px-4 py-2 text-sm', {
                  'bg-gray-100': active
                })}
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
