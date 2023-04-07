import { Bars3BottomLeftIcon } from '@heroicons/react/24/outline';
import { useEffect, useState } from 'react';
import { getInfo } from '~/data/api';
import { useSession } from '~/data/hooks/session';
import { Info } from '~/types/Meta';
import Notifications from './Notifications';
import UserProfile from './UserProfile';

type HeaderProps = {
  setSidebarOpen: (sidebarOpen: boolean) => void;
};

export default function Header(props: HeaderProps) {
  const { setSidebarOpen } = props;
  const [info, setInfo] = useState<Info | null>(null);

  useEffect(() => {
    getInfo()
      .then((info: Info) => {
        setInfo(info);
      })
      .catch(() => {
        // nothing to do, component will degrade gracefully
      });
  }, []);

  const { session } = useSession();

  return (
    <div className="sticky top-0 z-10 flex h-16 flex-shrink-0 bg-violet-400">
      <button
        type="button"
        className="without-ring px-4 text-white md:hidden"
        onClick={() => setSidebarOpen(true)}
      >
        <span className="sr-only">Open sidebar</span>
        <Bars3BottomLeftIcon className="h-6 w-6" aria-hidden="true" />
      </button>
      <div className="flex flex-1 justify-between px-4">
        <div className="flex flex-1" />
        <div className="ml-4 flex items-center md:ml-6">
          {/* notifications */}

          {/* TODO: currently we only show the update available notification, 
          this will need to be re-worked if we support other notifications */}
          {info && info.updateAvailable && <Notifications info={info} />}

          {/* user profile */}
          {session && session.self && (
            <UserProfile
              name={session.self.metadata['io.flipt.auth.oidc.name']}
              imgURL={session.self.metadata['io.flipt.auth.oidc.picture']}
            />
          )}
        </div>
      </div>
    </div>
  );
}
