import { ChartNoAxesCombinedIcon } from 'lucide-react';
import { useSelector } from 'react-redux';

import { selectInfo } from '~/app/meta/metaSlice';

import { PageHeader } from '~/components/Page';
import Well from '~/components/Well';

export default function Analytics() {
  const info = useSelector(selectInfo);

  return (
    <>
      <PageHeader title="Analytics" />

      <p className="mt-2 text-sm text-gray-500 dark:text-gray-300">
        Track and measure the impact in real-time.
      </p>

      {!info.analytics?.enabled && (
        <div className="mt-10">
          <Well>
            <ChartNoAxesCombinedIcon className="h-12 w-12 text-secondary-foreground/30 mb-4" />
            <h3 className="text-lg font-medium text-secondary-foreground mb-2">
              Analytics Disabled
            </h3>
            <p className="text-sm text-muted-foreground dark:text-gray-300">
              See the configuration{' '}
              <a
                className="text-violet-500 dark:text-violet-400 hover:text-violet-600 dark:hover:text-violet-300 transition-colors"
                href="https://www.flipt.io/docs/configuration/analytics"
              >
                documentation
              </a>{' '}
              for more information.
            </p>
          </Well>
        </div>
      )}
    </>
  );
}
