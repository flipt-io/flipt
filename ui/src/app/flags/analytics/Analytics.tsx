import { useMemo, useState } from 'react';
import { useSelector } from 'react-redux';
import { useOutletContext } from 'react-router-dom';
import Combobox from '~/components/forms/Combobox';
import {
  Chart as ChartJS,
  BarElement,
  CategoryScale,
  LinearScale,
  ArcElement,
  TimeScale,
  Title,
  Tooltip,
  Legend
} from 'chart.js';
import 'chartjs-adapter-date-fns';
import { add, format } from 'date-fns';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { IFlag } from '~/types/Flag';
import { useGetFlagEvaluationCountQuery } from '../analyticsApi';
import { BarGraph } from '~/components/graphs/BarGraph';
import { IFilterable } from '~/types/Selectable';
import { Formik } from 'formik';

ChartJS.register(
  ArcElement,
  BarElement,
  CategoryScale,
  LinearScale,
  TimeScale,
  Title,
  Tooltip,
  Legend
);

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
  }
];

export default function Analytics() {
  const [selectedDuration, setSelectedDuration] =
    useState<FilterableDurations | null>(durations[0]);
  const { flag } = useOutletContext<AnalyticsProps>();
  const namespace = useSelector(selectCurrentNamespace);

  const now = Date.now();
  const getFlagEvaluationCount = useGetFlagEvaluationCountQuery({
    namespaceKey: namespace.key,
    flagKey: flag.key,
    from: format(
      add(now, {
        minutes: selectedDuration?.value ? selectedDuration.value * -1 : -60
      }),
      timeFormat
    ),
    to: format(now, timeFormat)
  });

  const flagEvaluationCount = useMemo(() => {
    return {
      timestamps: getFlagEvaluationCount.data?.timestamps,
      values: getFlagEvaluationCount.data?.values
    };
  }, [getFlagEvaluationCount, selectedDuration]);

  const initialValues = {
    durationValue: selectedDuration?.key
  };

  return (
    <div className="mx-12 my-12">
      <>
        <Formik initialValues={initialValues} onSubmit={async function () {}}>
          {(formik) => (
            <Combobox<FilterableDurations>
              id="durationValue"
              name="durationValue"
              placeholder="Select duration"
              className="absolute right-24 z-20"
              values={durations}
              selected={selectedDuration}
              setSelected={setSelectedDuration}
            />
          )}
        </Formik>
      </>
      <div className="relative top-8">
        <BarGraph
          timestamps={flagEvaluationCount.timestamps || []}
          values={flagEvaluationCount.values || []}
          flagKey={flag.key}
        />
      </div>
    </div>
  );
}
