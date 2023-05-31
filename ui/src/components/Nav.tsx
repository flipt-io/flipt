import {
  BookOpenIcon,
  CodeBracketIcon,
  Cog6ToothIcon,
  FlagIcon,
  UsersIcon
} from '@heroicons/react/24/outline';
import { useSelector } from 'react-redux';
import { NavLink, useMatches } from 'react-router-dom';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { classNames } from '~/utils/helpers';
import NamespaceListbox from './namespaces/NamespaceListbox';

type Icon = (
  props: React.PropsWithoutRef<React.SVGProps<SVGSVGElement>>
) => any;

type NavItemProps = {
  to: string;
  name: string;
  Icon: Icon;
  external?: boolean;
  onClick?: () => void;
};

function NavItem(props: NavItemProps) {
  const { to, name, Icon, external, onClick } = props;

  return external ? (
    <a
      key={name}
      href={to}
      target="_blank"
      rel="noreferrer"
      className="group flex items-center rounded-md px-2 py-2 text-sm font-medium text-white hover:bg-violet-400 md:text-gray-600 md:hover:bg-gray-50"
    >
      <Icon
        className="text-wite mr-3 h-6 w-6 flex-shrink-0 hover:bg-gray-50 md:text-gray-500"
        aria-hidden="true"
      />
      {name}
    </a>
  ) : (
    <NavLink
      key={name}
      to={to}
      className={({ isActive }) =>
        classNames(
          isActive
            ? 'bg-violet-100 text-gray-600 md:bg-gray-50'
            : 'text-white hover:bg-violet-400 md:text-gray-600 md:hover:bg-gray-50',
          'group flex items-center rounded-md px-2 py-2 text-sm font-medium'
        )
      }
      onClick={onClick}
    >
      <Icon
        className="text-wite mr-3 h-6 w-6 flex-shrink-0 hover:bg-gray-50 md:text-gray-500"
        aria-hidden="true"
      />
      {name}
    </NavLink>
  );
}

type NavProps = {
  className?: string;
  sidebarOpen?: boolean;
  setSidebarOpen?: (open: boolean) => void;
};

// allows us to add custom properties to the route object
interface RouteMatches {
  namespaced: boolean;
}

export default function Nav(props: NavProps) {
  const { className, sidebarOpen, setSidebarOpen } = props;

  let matches = useMatches();
  let path = '';

  const namespace = useSelector(selectCurrentNamespace);
  path = `/namespaces/${namespace?.key}`;

  // if the current route is namespaced, we want to allow the namespace nav to be selectable
  let namespaceNavEnabled = matches.some((m) => {
    let r = m.handle as RouteMatches;
    return r?.namespaced;
  });

  const navigation = [
    {
      name: 'Flags',
      to: `${path}/flags`,
      Icon: FlagIcon
    },
    {
      name: 'Segments',
      to: `${path}/segments`,
      Icon: UsersIcon
    },
    {
      name: 'Console',
      to: `${path}/console`,
      Icon: CodeBracketIcon
    }
  ];

  const secondaryNavigation = [
    {
      name: 'Settings',
      to: 'settings',
      Icon: Cog6ToothIcon
    },
    {
      name: 'Documentation',
      to: 'https://flipt.io/docs?utm_source=app',
      Icon: BookOpenIcon,
      external: true
    }
  ];

  return (
    <nav
      className={`${className} flex flex-grow flex-col overflow-y-auto`}
      aria-label="Sidebar"
    >
      <div className="mb-4 flex flex-shrink-0 flex-col px-2">
        <NamespaceListbox disabled={!namespaceNavEnabled} />
      </div>
      <div className="flex flex-grow flex-col space-y-1 px-2">
        {navigation.map((item) => (
          <NavItem
            key={item.name}
            {...item}
            onClick={() => {
              if (sidebarOpen && setSidebarOpen) {
                setSidebarOpen(false);
              }
            }}
          />
        ))}
      </div>
      <div className="flex-shrink-0 space-y-1 px-2">
        {secondaryNavigation.map((item) => (
          <NavItem
            key={item.name}
            {...item}
            onClick={() => {
              if (sidebarOpen && setSidebarOpen) {
                setSidebarOpen(false);
              }
            }}
          />
        ))}
      </div>
    </nav>
  );
}
