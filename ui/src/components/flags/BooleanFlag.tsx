import { NavLink } from 'react-router-dom';
import { IFlag } from '~/types/Flag';
import { cls } from '~/utils/helpers';
import Rollouts from '~/components/rollouts/Rollouts';

type BooleanFlagProps = {
  flag: IFlag;
};

export default function BooleanFlag({ flag }: BooleanFlagProps) {
  const booleanFlagTabs = [
    { name: 'Rollouts', to: '' },
    { name: 'Analytics', to: 'analytics' }
  ];

  return (
    <>
      <div className="mt-3 flex flex-row sm:mt-5">
        <div className="border-gray-200 border-b-2">
          <nav className="-mb-px flex space-x-8">
            {booleanFlagTabs.map((tab) => (
              <NavLink
                end
                key={tab.name}
                to={tab.to}
                className={({ isActive }) =>
                  cls('whitespace-nowrap border-b-2 px-1 py-2 font-medium', {
                    'text-violet-600 border-violet-500': isActive,
                    'text-gray-600 border-transparent hover:text-gray-700 hover:border-gray-300':
                      !isActive
                  })
                }
              >
                {tab.name}
              </NavLink>
            ))}
          </nav>
        </div>
      </div>
      <Rollouts flag={flag} />
    </>
  );
}
