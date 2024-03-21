import { Formik } from 'formik';
import { useEffect, useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { useOutletContext } from 'react-router-dom';
import { PauseIcon, PlayIcon } from '@heroicons/react/24/outline';
import Combobox from '~/components/forms/Combobox';
import 'chartjs-adapter-date-fns';
import { addMinutes, format, parseISO } from 'date-fns';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { IFlag } from '~/types/Flag';
import { BarGraph } from '~/components/graphs';
import { IFilterable } from '~/types/Selectable';
import Well from '~/components/Well';
import { useGetFlagEvaluationCountQuery } from '~/app/flags/analyticsApi';
import { selectConfig } from '~/app/meta/metaSlice';

type AnalyticsProps = {
  flag: IFlag;
};

const timeFormat = 'yyyy-MM-dd HH:mm:ss';

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
  const [selectedDuration, setSelectedDuration] =
    useState<FilterableDurations | null>(durations[0]);
  const [pollingInterval, setPollingInterval] = useState<number>(0);

  const { flag } = useOutletContext<AnalyticsProps>();
  const namespace = useSelector(selectCurrentNamespace);

  const config = useSelector(selectConfig);

  const d = new Date();
  d.setSeconds(0);
  const nowISO = parseISO(d.toISOString());

  const getFlagEvaluationCount = useGetFlagEvaluationCountQuery(
    {
      namespaceKey: namespace.key,
      flagKey: flag.key,
      from: format(
        addMinutes(
          addMinutes(
            nowISO,
            selectedDuration?.value ? selectedDuration.value * -1 : -60
          ),
          nowISO.getTimezoneOffset()
        ),
        timeFormat
      ),
      to: format(addMinutes(nowISO, nowISO.getTimezoneOffset()), timeFormat)
    },
    {
      pollingInterval,
      skip: !config.analyticsEnabled
    }
  );

  const flagEvaluationCount = useMemo(() => {
    return {
      timestamps: getFlagEvaluationCount.data?.timestamps,
      values: getFlagEvaluationCount.data?.values
    };
  }, [getFlagEvaluationCount]);

  const initialValues = {
    durationValue: selectedDuration?.key
  };

  // Set the polling interval to 0 every time we change
  // durations.
  useEffect(() => {
    setPollingInterval(0);
  }, [selectedDuration]);

  return (
    <div className="mx-12 my-12">
      {config.analyticsEnabled ? (
        <>
          <>
            <Formik
              initialValues={initialValues}
              onSubmit={async function () {
                console.error('not implemented');
              }}
            >
              {() => (
                <div className="z-10 flex justify-center space-x-2 md:absolute md:right-24 md:justify-end">
                  <Combobox<FilterableDurations>
                    id="durationValue"
                    name="durationValue"
                    placeholder="Select duration"
                    values={durations}
                    selected={selectedDuration}
                    setSelected={setSelectedDuration}
                  />
                  <div className="mt-2">
                    {pollingInterval !== 0 ? (
                      <>
                        <PauseIcon
                          className="h-5 w-5"
                          onClick={() => setPollingInterval(0)}
                        />
                      </>
                    ) : (
                      <>
                        <PlayIcon
                          className="h-5 w-5"
                          onClick={() => setPollingInterval(3000)}
                        />
                      </>
                    )}
                  </div>
                </div>
              )}
            </Formik>
          </>
          <div className="md:relative md:top-12">
            <BarGraph
              timestamps={flagEvaluationCount.timestamps || []}
              values={flagEvaluationCount.values || []}
              flagKey={flag.key}
            />
          </div>
        </>
      ) : (
        <div className="flex flex-col text-center">
          <Well>
            <p className="text-gray-600 text-sm">Analytics Disabled</p>
            <p className="text-gray-500 mt-4 text-sm">
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
