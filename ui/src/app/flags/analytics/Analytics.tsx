import { useMemo } from 'react';
import { useSelector } from 'react-redux';
import { useOutletContext } from 'react-router-dom';
import {
  Chart as ChartJS,
  BarElement,
  CategoryScale,
  LinearScale,
  ArcElement,
  TimeScale,
  Title,
  Tooltip,
  Legend,
} from "chart.js";
import { Bar } from "react-chartjs-2";
import "chartjs-adapter-date-fns";
import { add, format } from 'date-fns';
import { selectCurrentNamespace } from '~/app/namespaces/namespacesSlice';
import { IFlag } from '~/types/Flag';
import { useGetFlagEvaluationCountQuery } from '../analyticsApi';

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

const timeFormat = "yyyy-MM-dd HH:mm:ss";

export default function Analytics() {
  const { flag } = useOutletContext<AnalyticsProps>();
  const namespace = useSelector(selectCurrentNamespace);

  const now = Date.now();
  const getFlagEvaluationCount = useGetFlagEvaluationCountQuery({
    namespaceKey: namespace.key,
    flagKey: flag.key,
    from: format(add(now, { hours: -1 }), timeFormat),
    to: format(now, timeFormat),
  });
  
  const flagEvaluationCount = useMemo(() => {
    return {
      timestamps: getFlagEvaluationCount.data?.timestamps,
      values: getFlagEvaluationCount.data?.values,
    };
  }, [getFlagEvaluationCount]);

  return (
    <div className="mx-12 my-12">
      <Bar data={{
        labels: flagEvaluationCount.timestamps,
        datasets: [
          {
            label: flag.key,
            data: flagEvaluationCount.values,
            backgroundColor: "rgba(167,139,250,0.6)",
            borderColor: "rgba(167,139,250,1)",
            borderWidth: 1,
          }
        ]
      }} options={{
        scales: {
          x: {
            type: "time",
            time: {
              unit: "minute",
              displayFormats: {
                minute: timeFormat,
              }
            },
            title: {
              display: true,
              text: "Time",
            },
          },
          y: {
            title: {
              display: true,
              text:  "Number of evaluations",
            },
          },
        }
      }} />
    </div>
  );
}
