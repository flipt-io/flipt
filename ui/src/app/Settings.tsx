import { Outlet, useOutletContext } from 'react-router-dom';
import TabBar from '~/components/TabBar';

export default function Settings() {
  const tabs = [
    {
      name: 'General',
      to: '/settings'
    },
    {
      name: 'Namespaces',
      to: '/settings/namespaces'
    },
    {
      name: 'API Tokens',
      to: '/settings/tokens'
    }
  ];

  return (
    <>
      <div className="flex items-center justify-between">
        <div className="min-w-0 flex-1">
          <h2 className="text-gray-900 text-2xl font-bold leading-7 sm:truncate sm:text-3xl sm:tracking-tight">
            Settings
          </h2>
        </div>
      </div>
      <TabBar tabs={tabs} />
      <Outlet context={useOutletContext()} />
    </>
  );
}
