import { Code, CircleHelp, Flag, Settings, Users } from 'lucide-react';
import { useSelector } from 'react-redux';
import { NavLink, useMatches } from 'react-router';
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
      aria-label={name}
      className={({ isActive }) =>
        cls(
          'flex items-center rounded-md px-2 py-2 text-sm font-medium text-white',
          {
            'md:bg-secondary md:text-secondary-foreground md: md:dark:text-gray-950':
              isActive,
            'hover: md:text-secondary-foreground md:hover:bg-secondary md:hover:text-secondary-foreground dark:hover: dark:hover:text-secondary-foreground md:dark:text-muted-foreground':
              !isActive
          }
        )
      }
      onClick={onClick}
    >
      <Icon
        className="md:text-muted-foreground md:dark:text-muted-foreground mr-3 h-6 w-6 shrink-0 text-white"
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
      Icon: Flag
    },
    {
      name: 'Segments',
      to: `${path}/segments`,
      Icon: Users
    },
    {
      name: 'Console',
      to: `${path}/console`,
      Icon: Code
    }
  ];

  const secondaryNavigation = [
    {
      name: 'Settings',
      to: 'settings',
      Icon: Settings
    },
    {
      name: 'Support',
      to: 'support',
      Icon: CircleHelp
    }
  ];

  return (
    <nav
      className={`${className} flex grow flex-col overflow-y-auto`}
      aria-label="Sidebar"
    >
      <div className="mb-4 flex shrink-0 flex-col px-2">
        <NamespaceListbox disabled={!namespaceNavEnabled} />
      </div>
      <div className="flex grow flex-col space-y-1 px-2">
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
      <div className="shrink-0 space-y-1 px-2">
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
        <div className="text-muted-foreground flex space-x-1 px-3 pt-2 text-xs">
          <span className="shrink-0">Command Mode:</span>
          <div className="shrink-0">
            <kbd className="text-muted-foreground">ctrl</kbd> +{' '}
            <kbd className="text-muted-foreground">k</kbd>
          </div>
        </div>
      </div>
    </nav>
  );
}
