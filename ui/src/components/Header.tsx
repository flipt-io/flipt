import { Bars3BottomLeftIcon } from '@heroicons/react/24/outline';
import { useSelector } from 'react-redux';
import { selectInfo, selectReadonly } from '~/app/meta/metaSlice';
import { useSession } from '~/data/hooks/session';
import Notifications from './header/Notifications';
import UserProfile from './header/UserProfile';

type HeaderProps = {
  setSidebarOpen: (sidebarOpen: boolean) => void;
};

export default function Header(props: HeaderProps) {
  const { setSidebarOpen } = props;

  const info = useSelector(selectInfo);
  const readOnly = useSelector(selectReadonly);

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
        <div className="ml-4 flex items-center space-x-1.5 md:ml-6">
          {/* read-only mode */}
          {readOnly && (
            <span className="nightwind-prevent inline-flex items-center gap-x-1.5 rounded-full px-3 py-1 text-xs font-medium text-violet-950 bg-violet-200">
              <svg
                className="h-1.5 w-1.5 fill-orange-400"
                viewBox="0 0 6 6"
                aria-hidden="true"
              >
                <circle cx={3} cy={3} r={3} />
              </svg>
              Read-Only
            </span>
          )}
          {/* notifications */}
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
