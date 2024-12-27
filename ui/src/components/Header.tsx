import { Bars3BottomLeftIcon } from '@heroicons/react/24/outline';
import { useSelector } from 'react-redux';
import { selectConfig, selectInfo, selectReadonly } from '~/app/meta/metaSlice';
import { useSession } from '~/data/hooks/session';
import { getUser } from '~/data/user';
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
  const { ui } = useSelector(selectConfig);
  const topbarStyle = { backgroundColor: ui.topbar?.color };
  const user = getUser(session);

  return (
    <div className="sticky top-0 z-20 flex h-16 flex-shrink-0 bg-black dark:border-b dark:border-b-white/20 dark:md:border-b-0">
      <button
        type="button"
        className="without-ring px-4 text-white md:hidden"
        style={topbarStyle}
        onClick={() => setSidebarOpen(true)}
      >
        <span className="sr-only">Open sidebar</span>
        <Bars3BottomLeftIcon className="h-6 w-6" aria-hidden="true" />
      </button>

      <div
        className="top-0 flex flex-1 justify-between px-4"
        style={topbarStyle}
      >
        <div className="flex flex-1" />
        <div className="flex items-center gap-2 pr-2">
          {/* read-only mode */}
          {readOnly && <ReadOnly />}

          {/* notifications */}
          {info && info.updateAvailable && <Notifications info={info} />}

          {/* user profile */}
          {user && <UserProfile user={user} />}
        </div>
      </div>
    </div>
  );
}
