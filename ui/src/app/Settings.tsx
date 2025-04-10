import { Outlet, useOutletContext } from 'react-router';

import { PageHeader } from '~/components/Page';
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
    <div className="space-y-6">
      <PageHeader title="Settings" />
      <TabBar tabs={tabs} />
      <Outlet context={useOutletContext()} />
    </div>
  );
}
