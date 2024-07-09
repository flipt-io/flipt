import {
  CodeBracketIcon,
  Cog6ToothIcon,
  FlagIcon,
  QuestionMarkCircleIcon,
  UsersIcon
} from '@heroicons/react/24/outline';
import { useSelector } from 'react-redux';
import { NavLink, useMatches } from 'react-router-dom';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { cls } from '~/utils/helpers';
import NamespaceListbox from './namespaces/NamespaceListbox';
import { RouteMatches } from '~/types/Routes';

type Icon = (
  props: React.PropsWithoutRef<React.SVGProps<SVGSVGElement>>
) => any;

type NavItemProps = {
  to: string;
  name: string;
  Icon: Icon;
  onClick?: () => void;
};

function NavItem(props: NavItemProps) {
  const { to, name, Icon, onClick } = props;

  return (
    <NavLink
      key={name}
      to={to}
      className={({ isActive }) =>
        cls(
          'group nightwind-prevent text-white flex items-center rounded-md px-2 py-2 text-sm font-medium',
          {
            'bg-gray-800 md:text-gray-700 md:bg-gray-50 md:dark:text-gray-50 md:dark:bg-gray-700':
              isActive,
            'hover:bg-gray-700 md:text-gray-600 md:hover:text-gray-700 md:hover:bg-gray-50 md:dark:text-gray-400 md:dark:hover:text-gray-300 md:dark:hover:bg-gray-700':
              !isActive
          }
        )
      }
      onClick={onClick}
    >
      <Icon
        className="nightwind-prevent text-white mr-3 h-6 w-6 flex-shrink-0 md:text-gray-500 md:dark:text-gray-400"
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

export default function Nav(props: NavProps) {
  const { className, sidebarOpen, setSidebarOpen } = props;

  const matches = useMatches();
  let path = '';

  const namespace = useSelector(selectCurrentNamespace);
  path = `/namespaces/${namespace?.key}`;

  // if the current route is namespaced, we want to allow the namespace nav to be selectable
  const namespaceNavEnabled = matches.some((m) => {
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
      name: 'Support',
      to: 'support',
      Icon: QuestionMarkCircleIcon
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
        <div className="text-gray-400 flex space-x-1 px-3 pt-2 text-xs">
          <span className="flex-shrink-0">Command Mode:</span>
          <div className="flex-shrink-0">
            <kbd className="text-gray-400">ctrl</kbd> +{' '}
            <kbd className="text-gray-400">k</kbd>
          </div>
        </div>
      </div>
    </nav>
  );
}
