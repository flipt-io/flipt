import 'chartjs-adapter-date-fns';
import { addMinutes } from 'date-fns';
import { PauseIcon, PlayIcon } from 'lucide-react';
import { ChartNoAxesCombinedIcon } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { useNavigate, useParams } from 'react-router';

import { selectCurrentEnvironment } from '~/app/environments/environmentsApi';
import { useGetFlagEvaluationCountQuery } from '~/app/flags/analyticsApi';
import { useListFlagsQuery } from '~/app/flags/flagsApi';
import { selectInfo } from '~/app/meta/metaSlice';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';

import { Button } from '~/components/Button';
import Combobox from '~/components/Combobox';
import { PageHeader } from '~/components/Page';
import Well from '~/components/Well';
import Listbox from '~/components/forms/Listbox';
import { Graph } from '~/components/graphs';

import { FlagType, IFlag, flagTypeToLabel } from '~/types/Flag';
import { ISelectable } from '~/types/Selectable';

import { useError } from '~/data/hooks/error';

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

// Helper for RTK skip
const skipToken = undefined as any;

export default function Analytics() {
  const environment = useSelector(selectCurrentEnvironment);
  const namespace = useSelector(selectCurrentNamespace);
  const info = useSelector(selectInfo);

  const { setError, clearError } = useError();
  const navigate = useNavigate();
  const { flagKey } = useParams();

  const {
    data: flagsData,
    error,
    isLoading: flagsLoading
  } = useListFlagsQuery({
    environmentKey: environment.key,
    namespaceKey: namespace.key
  });

  useEffect(() => {
    if (error) {
      setError(error);
      return;
    }
    clearError();
  }, [clearError, error, setError]);

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
  >(null);
  const [selectedDuration, setSelectedDuration] = useState<FilterableDurations>(
    durations[0]
  );
  const [pollingInterval, setPollingInterval] = useState<number>(0);

  // Select flag from path param if present
  useEffect(() => {
    if (flagKey && flagOptions.length > 0) {
      const found = flagOptions.find((f) => f.key === flagKey);
      if (found && (!selectedFlag || selectedFlag.key !== found.key)) {
        setSelectedFlag(found);
      }
    } else if (!flagKey) {
      setSelectedFlag(null);
    }
  }, [flagKey, flagOptions, selectedFlag]);

  // Keep selectedFlag in sync with flagOptions
  useEffect(() => {
    if (
      flagOptions.length > 0 &&
      selectedFlag &&
      !flagOptions.find((f) => f.key === selectedFlag.key)
    ) {
      setSelectedFlag(null);
    }
    if (flagOptions.length === 0) {
      setSelectedFlag(null);
    }
  }, [flagsData, flagOptions, selectedFlag]);

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
            {flagOptions.length > 0 && (
              <Combobox
                id="flagKey"
                name="flagKey"
                className="w-lg"
                placeholder="Select or search for a flag"
                values={flagOptions}
                selected={selectedFlag}
                setSelected={(flag) => {
                  setSelectedFlag(flag);
                  if (flag) {
                    navigate(
                      `/namespaces/${namespace.key}/analytics/${flag.key}`,
                      { replace: true }
                    );
                  } else {
                    navigate(`/namespaces/${namespace.key}/analytics`, {
                      replace: true
                    });
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
        {flagOptions.length === 0 ? (
          <div className="mt-12 w-full">
            <Well>
              <ChartNoAxesCombinedIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-2">
                No Flags Available
              </h3>
              <p className="text-sm text-muted-foreground mb-4">
                At least one flag must exist to view analytics
              </p>
              <Button
                aria-label="New Flag"
                variant="primary"
                onClick={() =>
                  navigate(`/namespaces/${namespace.key}/flags/new`)
                }
              >
                Create Your First Flag
              </Button>
            </Well>
          </div>
        ) : !selectedFlag ? (
          <div className="mt-12 w-full">
            <Well>
              <ChartNoAxesCombinedIcon className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <h3 className="text-lg font-medium text-muted-foreground mb-2">
                No Flag Selected
              </h3>
              <p className="text-sm text-muted-foreground mb-4">
                Select a flag to view analytics.
              </p>
            </Well>
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
