import { NavLink } from 'react-router';

import { cls } from '~/utils/helpers';

export interface Tab {
  name: string;
  to: string;
}

type TabBarProps = {
  tabs: Tab[];
};

export default function TabBar(props: TabBarProps) {
  const { tabs } = props;

  return (
    <div className="flex flex-row">
      <div className="border-b-2 border-gray-200 dark:border-gray-700">
        <nav className="-mb-px flex space-x-8">
          {tabs.map((tab) => (
            <NavLink
              end
              key={tab.name}
              to={tab.to}
              className={({ isActive }) =>
                cls(
                  'whitespace-nowrap border-b-2 px-1 py-2 text-sm font-medium',
                  {
                    'border-violet-500 text-violet-600 dark:text-violet-400':
                      isActive,
                    'border-transparent text-gray-600 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:border-gray-600 dark:hover:text-gray-300':
                      !isActive
                  }
                )
              }
            >
              {tab.name}
            </NavLink>
          ))}
        </nav>
      </div>
    </div>
  );
}
