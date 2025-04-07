import { Bars3BottomLeftIcon } from '@heroicons/react/24/outline';
import { useSelector } from 'react-redux';
import { Link } from 'react-router';

import { selectInfo } from '~/app/meta/metaSlice';

import logoLight from '~/assets/logo-light.png';
import { useSession } from '~/data/hooks/session';
import { getUser } from '~/data/user';

import EnvironmentListbox from './environments/EnvironmentListbox';
import Notifications from './header/Notifications';
import UserProfile from './header/UserProfile';

type HeaderProps = {
  setSidebarOpen: (sidebarOpen: boolean) => void;
};

export default function Header(props: HeaderProps) {
  const { setSidebarOpen } = props;
  const info = useSelector(selectInfo);
  const { session } = useSession();
  const user = getUser(session);

  return (
    <div className="fixed left-0 right-0 top-0 z-20 flex h-16 shrink-0 bg-black dark:border-b dark:border-b-white/20 dark:md:border-b-0">
      <button
        type="button"
        className="without-ring px-4 text-white md:hidden"
        onClick={() => setSidebarOpen(true)}
      >
        <span className="sr-only">Open sidebar</span>
        <Bars3BottomLeftIcon className="h-6 w-6" aria-hidden="true" />
      </button>

      <div className="flex flex-1 items-center justify-between px-4">
        <div className="flex items-center gap-4">
          <Link to="/">
            <img
              src={logoLight}
              alt="logo"
              width={549}
              height={191}
              className="h-10 w-auto"
            />
          </Link>
          <EnvironmentListbox className="w-48" />
        </div>

        <div className="flex items-center gap-2 pr-2">
          {/* notifications */}
          {info && info.updateAvailable && <Notifications info={info} />}

          {/* user profile */}
          {user && <UserProfile user={user} />}
        </div>
      </div>
    </div>
  );
}
