import { PauseIcon, PlayIcon } from '@heroicons/react/24/outline';
import 'chartjs-adapter-date-fns';
import { addMinutes } from 'date-fns';
import { ChartNoAxesCombinedIcon } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate, useSearchParams } from 'react-router';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { useGetFlagEvaluationCountQuery } from '~/app/flags/analyticsApi';
import { useListFlagsQuery } from '~/app/flags/flagsApi';
import { selectInfo } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';

import Combobox from '~/components/Combobox';
import { PageHeader } from '~/components/Page';
import Well from '~/components/Well';
import Listbox from '~/components/forms/Listbox';
import { Graph } from '~/components/graphs';

import { FlagType, IFlag, flagTypeToLabel } from '~/types/Flag';
import { ISelectable } from '~/types/Selectable';

interface IDuration {
  value: number;
}

type FilterableDurations = IDuration & ISelectable;

const durations: FilterableDurations[] = [
  {
    value: 30,
    key: '30 minutes',
    displayValue: '30 minutes'
  },
  {
    value: 60,
    key: '1 hour',
    displayValue: '1 hour'
  },
  {
    value: 60 * 4,
    key: '4 hours',
    displayValue: '4 hours'
  },
  {
    value: 60 * 12,
    key: '12 hours',
    displayValue: '12 hours'
  },
  {
    value: 60 * 24,
    key: '1 day',
    displayValue: '1 day'
  }
];

export default function Analytics() {
  const environment = useSelector(selectCurrentEnvironment);
  const namespace = useSelector(selectCurrentNamespace);
  const info = useSelector(selectInfo);
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const { data: flagsData, isLoading: flagsLoading } = useListFlagsQuery({
    environmentKey: environment.key,
    namespaceKey: namespace.key
  });

  // Map flags to ISelectable for Combobox, including status and type label
  const flagOptions = (flagsData?.flags || []).map((flag: IFlag) => {
    const status =
      flag.enabled || flag.type === FlagType.BOOLEAN ? 'active' : 'inactive';
    return {
      ...flag,
      status: status as 'active' | 'inactive',
      displayValue: `${flag.name} | ${flagTypeToLabel(flag.type)}`
    };
  });

  // Use null instead of undefined for selectedFlag
  const [selectedFlag, setSelectedFlag] = useState<
    (IFlag & ISelectable) | null
  >(flagOptions[0] || null);
  const [selectedDuration, setSelectedDuration] = useState<FilterableDurations>(
    durations[0]
  );
  const [pollingInterval, setPollingInterval] = useState<number>(0);

  // Select flag from query param if present
  useEffect(() => {
    const flagKey = searchParams.get('flag');
    if (flagKey && flagOptions.length > 0) {
      const found = flagOptions.find((f) => f.key === flagKey);
      if (found && (!selectedFlag || selectedFlag.key !== found.key)) {
        setSelectedFlag(found);
      }
    }
  }, [searchParams, flagOptions]);

  // Keep selectedFlag in sync with flagOptions
  useEffect(() => {
    if (
      flagOptions.length > 0 &&
      (!selectedFlag || !flagOptions.find((f) => f.key === selectedFlag.key))
    ) {
      setSelectedFlag(flagOptions[0]);
    }
    if (flagOptions.length === 0) {
      setSelectedFlag(null);
    }
  }, [flagsData]);

  // Analytics query
  const d = new Date();
  d.setSeconds(0);
  d.setMilliseconds(0);

  const getFlagEvaluationCount = useGetFlagEvaluationCountQuery(
    selectedFlag && info.analytics?.enabled
      ? {
          environmentKey: environment.key,
          namespaceKey: namespace.key,
          flagKey: selectedFlag.key,
          from: addMinutes(
            d,
            selectedDuration?.value ? selectedDuration.value * -1 : -60
          ).toISOString(),
          to: d.toISOString()
        }
      : skipToken,
    {
      pollingInterval,
      skip: !info.analytics?.enabled || !selectedFlag
    }
  );

  const flagEvaluationCount = useMemo(() => {
    return {
      timestamps: getFlagEvaluationCount.data?.timestamps || [],
      values: getFlagEvaluationCount.data?.values || []
    };
  }, [getFlagEvaluationCount]);

  // Set the polling interval to 0 every time we change durations or flag
  useEffect(() => {
    setPollingInterval(0);
  }, [selectedDuration, selectedFlag]);

  // Empty state if analytics disabled
  if (!info.analytics?.enabled) {
    return (
      <>
        <PageHeader title="Analytics" />
        <div className="mt-10">
          <Well>
            <div className="flex flex-col items-center text-center p-4">
              <ChartNoAxesCombinedIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-2">
                Analytics Disabled
              </h3>
              <p className="text-sm text-muted-foreground mb-4">
                See the configuration{' '}
                <a
                  className="text-violet-500 dark:text-violet-400 hover:text-violet-600 dark:hover:text-violet-300 transition-colors"
                  href="https://www.flipt.io/docs/configuration/analytics"
                >
                  documentation
                </a>{' '}
                for more information.
              </p>
            </div>
          </Well>
        </div>
      </>
    );
  }

  return (
    <>
      <PageHeader title="Analytics" />
      <div className="mt-2">
        <div className="sm:flex sm:items-center mb-6">
          <div className="sm:flex-auto">
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-300">
              Track and measure the impact of your flags in real-time.
            </p>
          </div>
          <div className="flex items-center gap-2">
            {flagOptions.length > 0 && selectedFlag && (
              <Combobox
                id="flagValue"
                name="flagValue"
                className="w-lg"
                placeholder="Select or search for a flag"
                values={flagOptions}
                selected={selectedFlag}
                setSelected={(flag) => {
                  setSelectedFlag(flag);
                  if (flag) {
                    setSearchParams({ flag: flag.key }, { replace: true });
                  }
                }}
                disabled={flagsLoading}
              />
            )}
            <Listbox<FilterableDurations>
              id="durationValue"
              name="durationValue"
              values={durations}
              selected={selectedDuration}
              setSelected={setSelectedDuration}
              className="w-32"
            />
            {pollingInterval !== 0 ? (
              <PauseIcon
                className="h-5 w-5 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 cursor-pointer"
                onClick={() => setPollingInterval(0)}
              />
            ) : (
              <PlayIcon
                className="h-5 w-5 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 cursor-pointer"
                onClick={() => setPollingInterval(3000)}
              />
            )}
          </div>
        </div>
        {flagsLoading ? (
          <div className="mt-10 text-center text-muted-foreground">
            Loading flags…
          </div>
        ) : flagOptions.length === 0 ? (
          <div className="mt-12 w-full">
            <Well>
              <ChartNoAxesCombinedIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-2">
                No Flags Available
              </h3>
              <p className="text-sm text-muted-foreground mb-4">
                At least one flag must exist to view analytics
              </p>
              <button
                aria-label="New Flag"
                onClick={() =>
                  navigate(`/namespaces/${namespace.key}/flags/new`)
                }
                className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-violet-500 text-white hover:bg-violet-600 h-9 px-4 py-2"
              >
                Create Your First Flag
              </button>
            </Well>
          </div>
        ) : !selectedFlag ? (
          <div className="mt-10 text-center text-muted-foreground">
            Select a flag to view analytics.
          </div>
        ) : (
          <div className="mt-10">
            <Graph
              timestamps={flagEvaluationCount.timestamps}
              values={flagEvaluationCount.values}
              flagKey={selectedFlag.key}
            />
          </div>
        )}
      </div>
    </>
  );
}

// Helper for RTK skip
const skipToken = undefined as any;
