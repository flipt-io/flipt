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
          <h2 className="text-2xl leading-7 font-bold text-gray-900 sm:truncate sm:text-3xl sm:tracking-tight">
            Settings
          </h2>
        </div>
      </div>
      <TabBar tabs={tabs} />
      <Outlet context={useOutletContext()} />
    </>
  );
}
