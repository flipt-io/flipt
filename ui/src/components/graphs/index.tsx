import {
  Chart as ChartJS,
  BarElement,
  CategoryScale,
  LinearScale,
  ArcElement,
  TimeScale,
  TimeSeriesScale,
  Title,
  Tooltip,
  Legend
} from 'chart.js';
import { useMemo } from 'react';
import { Bar } from 'react-chartjs-2';
import { useSelector } from 'react-redux';
import { selectTimezone } from '~/app/preferences/preferencesSlice';
import { useTimezone } from '~/data/hooks/timezone';
import { Timezone } from '~/types/Preferences';

ChartJS.register(
  ArcElement,
  BarElement,
  CategoryScale,
  LinearScale,
  TimeScale,
  TimeSeriesScale,
  Title,
  Tooltip,
  Legend
);

const timeFormat = 'yyyy-MM-dd HH:mm:ss';

type BarGraphProps = {
  flagKey: string;
  timestamps: string[];
  values: number[];
};

export function BarGraph({ flagKey, timestamps, values }: BarGraphProps) {
  const timezone = useSelector(selectTimezone);
  const { inTimezone } = useTimezone();

  const isUTC = useMemo(() => timezone === Timezone.UTC, [timezone]);
  const formattedTimestamps = timestamps.map((timestamp) => {
    if (isUTC) {
      return inTimezone(timestamp).slice(0, -4);
    }
    return inTimezone(timestamp);
  });

  const xLabel = isUTC ? 'Time (UTC)' : 'Time (Local)';

  return (
    <div className="h-[90vh]">
      <Bar
        data={{
          labels: formattedTimestamps,
          datasets: [
            {
              label: flagKey,
              data: values,
              backgroundColor: 'rgba(167,139,250,0.6)',
              borderColor: 'rgba(167,139,250,1)',
              borderWidth: 1
            }
          ]
        }}
        options={{
          maintainAspectRatio: false,
          scales: {
            x: {
              type: 'timeseries',
              time: {
                displayFormats: {
                  minute: timeFormat
                }
              },
              ticks: {
                autoSkip: true,
                maxTicksLimit: 7
              },
              title: {
                display: true,
                text: xLabel
              }
            },
            y: {
              title: {
                display: true,
                text: 'Evaluations'
              }
            }
          }
        }}
      />
    </div>
  );
}
