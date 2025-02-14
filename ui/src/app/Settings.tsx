import { Outlet, useOutletContext } from 'react-router';

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
    }
  ];

  return (
    <>
      <div className="flex items-center justify-between">
        <div className="min-w-0 flex-1">
          <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight">
            Settings
          </h1>
        </div>
      </div>
      <TabBar tabs={tabs} />
      <Outlet context={useOutletContext()} />
    </>
  );
}
