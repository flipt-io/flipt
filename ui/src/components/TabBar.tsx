import { NavLink } from 'react-router-dom';
import { classNames } from '~/utils/helpers';

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
    <div className="mt-3 flex flex-row sm:mt-5">
      <div className="border-b-2 border-gray-200">
        <nav className="-mb-px flex space-x-8">
          {tabs.map((tab) => (
            <NavLink
              end
              key={tab.name}
              to={tab.to}
              className={({ isActive }) =>
                classNames(
                  isActive
                    ? 'border-violet-500 text-violet-600'
                    : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700',
                  'whitespace-nowrap border-b-2 px-1 py-3 text-sm font-medium'
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
