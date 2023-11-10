import { Bars3BottomLeftIcon } from '@heroicons/react/24/outline';
import { useSelector } from 'react-redux';
import { selectInfo, selectReadonly } from '~/app/meta/metaSlice';
import { useSession } from '~/data/hooks/session';
import Notifications from './header/Notifications';
import ReadOnly from './header/ReadOnly';
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
    <div className="sticky top-0 z-10 flex h-16 flex-shrink-0 bg-gray-950 dark:border-b dark:border-b-white/20">
      <button
        type="button"
        className="without-ring nightwind-prevent text-white px-4 md:hidden"
        onClick={() => setSidebarOpen(true)}
      >
        <span className="sr-only">Open sidebar</span>
        <Bars3BottomLeftIcon className="h-6 w-6" aria-hidden="true" />
      </button>

      <div className="flex flex-1 justify-between px-4">
        <div className="flex flex-1" />
        <div className="flex items-center">
          {/* read-only mode */}
          {readOnly && <ReadOnly />}

          {/* notifications */}
          {info && info.updateAvailable && <Notifications info={info} />}

          {/* user profile */}
          {session && session.self && (
            <UserProfile metadata={session.self.metadata} />
          )}
        </div>
      </div>
    </div>
  );
}
