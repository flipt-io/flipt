import {
  Chart as ChartJS,
  BarElement,
  CategoryScale,
  LinearScale,
  TimeScale,
  TimeSeriesScale,
  Title,
  Tooltip,
  Legend,
  PointElement,
  LineElement,
  Filler
} from 'chart.js';
import { useMemo } from 'react';
import { Line } from 'react-chartjs-2';
import { useSelector } from 'react-redux';
import { selectTimezone } from '~/app/preferences/preferencesSlice';
import { useTimezone } from '~/data/hooks/timezone';
import { Timezone } from '~/types/Preferences';

ChartJS.register(
  CategoryScale,
  LinearScale,
  TimeScale,
  TimeSeriesScale,
  Title,
  Tooltip,
  Legend,
  PointElement,
  LineElement,
  Filler
);

type GraphProps = {
  flagKey: string;
  timestamps: string[];
  values: number[];
};

export function BarGraph({ flagKey, timestamps, values }: GraphProps) {
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
    <div className="h-[40vh]">
      <Line
        data={{
          labels: formattedTimestamps,
          datasets: [
            {
              label: flagKey,
              data: values,
              backgroundColor: 'rgba(167,139,250,0.6)',
              borderColor: 'rgba(167,139,250,1)',
              borderWidth: 1,
              fill: true
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
                  minute: 'HH:mm'
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
              beginAtZero: true,
              ticks: {
                autoSkip: true,
                maxTicksLimit: 5
              },
              title: {
                display: true,
                text: 'Evaluations'
              }
            }
          },
          plugins: {
            title: {
              display: false
            },
            legend: {
              display: false
            },
            tooltip: {
              displayColors: false,
              callbacks: {
                title: function (tooltipItem) {
                  return formattedTimestamps[tooltipItem[0].dataIndex];
                },
                label: function (tooltipItem) {
                  return tooltipItem.formattedValue;
                }
              }
            }
          }
        }}
      />
    </div>
  );
}
