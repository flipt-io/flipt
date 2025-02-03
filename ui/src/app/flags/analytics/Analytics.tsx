import { useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { useOutletContext } from 'react-router';
import { PauseIcon, PlayIcon } from '@heroicons/react/24/outline';
import 'chartjs-adapter-date-fns';
import { addMinutes } from 'date-fns';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesApi';
import { IFlag } from '~/types/Flag';
import { Graph } from '~/components/graphs';
import { IFilterable } from '~/types/Selectable';
import Well from '~/components/Well';
import { useGetFlagEvaluationCountQuery } from '~/app/flags/analyticsApi';
import { selectInfo } from '~/app/meta/metaSlice';
import Listbox from '~/components/forms/Listbox';

type AnalyticsProps = {
  flag: IFlag;
};

interface IDuration {
  value: number;
}

type FilterableDurations = IDuration & IFilterable;

const durations: FilterableDurations[] = [
  {
    value: 30,
    key: '30 minutes',
    displayValue: '30 minutes',
    filterValue: '30 minutes'
  },
  {
    value: 60,
    key: '1 hour',
    displayValue: '1 hour',
    filterValue: '1 hour'
  },
  {
    value: 60 * 4,
    key: '4 hours',
    displayValue: '4 hours',
    filterValue: '4 hours'
  },
  {
    value: 60 * 12,
    key: '12 hours',
    displayValue: '12 hours',
    filterValue: '12 hours'
  },
  {
    value: 60 * 24,
    key: '1 day',
    displayValue: '1 day',
    filterValue: '1 day'
  }
];

export default function Analytics() {
  const [selectedDuration, setSelectedDuration] = useState<FilterableDurations>(
    durations[0]
  );
  const [pollingInterval, setPollingInterval] = useState<number>(0);

  const { flag } = useOutletContext<AnalyticsProps>();
  const namespace = useSelector(selectCurrentNamespace);

  const info = useSelector(selectInfo);

  const d = new Date();
  d.setSeconds(0);
  d.setMilliseconds(0);

  const getFlagEvaluationCount = useGetFlagEvaluationCountQuery(
    {
      namespaceKey: namespace.key,
      flagKey: flag.key,
      from: addMinutes(
        d,
        selectedDuration?.value ? selectedDuration.value * -1 : -60
      ).toISOString(),
      to: d.toISOString()
    },
    {
      pollingInterval,
      skip: !info.analyticsEnabled
    }
  );

  const flagEvaluationCount = useMemo(() => {
    return {
      timestamps: getFlagEvaluationCount.data?.timestamps || [],
      values: getFlagEvaluationCount.data?.values || []
    };
  }, [getFlagEvaluationCount]);

  // Set the polling interval to 0 every time we change
  // durations.
  useEffect(() => {
    setPollingInterval(0);
  }, [selectedDuration]);

  return (
    <div className="mt-2 max-w-screen-lg">
      {info.analyticsEnabled ? (
        <>
          <div className="sm:flex sm:items-center">
            <div className="sm:flex-auto">
              <p className="mt-1 text-sm text-gray-500">
                Track and measure the impact in real-time.
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Listbox<FilterableDurations>
                id="durationValue"
                name="durationValue"
                values={durations}
                selected={selectedDuration}
                setSelected={setSelectedDuration}
                className="-mt-2 w-32"
              />
              {pollingInterval !== 0 ? (
                <PauseIcon
                  className="h-5 w-5 text-gray-500"
                  onClick={() => setPollingInterval(0)}
                />
              ) : (
                <PlayIcon
                  className="h-5 w-5 text-gray-500"
                  onClick={() => setPollingInterval(3000)}
                />
              )}
            </div>
          </div>
          <div className="mt-10">
            <Graph
              timestamps={flagEvaluationCount.timestamps}
              values={flagEvaluationCount.values}
              flagKey={flag.key}
            />
          </div>
        </>
      ) : (
        <div className="mt-10">
          <Well>
            <p>Analytics Disabled</p>
            <p className="mt-4">
              See the configuration{' '}
              <a
                className="text-violet-500"
                href="https://www.flipt.io/docs/configuration/analytics"
              >
                documentation
              </a>{' '}
              for more information.
            </p>
          </Well>
        </div>
      )}
    </div>
  );
}
