import { Outlet, useOutletContext } from 'react-router';

import { PageHeader } from '~/components/Page';
import TabBar from '~/components/TabBar';

export default function Settings() {
  const tabs = [
    {
      name: 'Preferences',
      to: '/settings'
    },
    {
      name: 'Namespaces',
      to: '/settings/namespaces'
    }
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center space-x-3">
        <PageHeader title="Settings" />
      </div>
      <TabBar tabs={tabs} />
      <Outlet context={useOutletContext()} />
    </div>
  );
}
