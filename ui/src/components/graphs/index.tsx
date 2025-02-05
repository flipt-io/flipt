import {
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LineElement,
  LinearScale,
  PointElement,
  TimeScale,
  TimeSeriesScale,
  Title,
  Tooltip
} from 'chart.js';
import { useMemo } from 'react';
import { Line } from 'react-chartjs-2';
import { useSelector } from 'react-redux';

import { selectTimezone } from '~/app/preferences/preferencesSlice';

import { Timezone } from '~/types/Preferences';

import { useTimezone } from '~/data/hooks/timezone';

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

export function Graph({ flagKey, timestamps, values }: GraphProps) {
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
              backgroundColor: 'rgba(167,139,250,0.5)',
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
                  millisecond: 'HH:mm',
                  second: 'HH:mm',
                  minute: 'HH:mm',
                  hour: 'HH:mm',
                  day: 'HH:mm'
                }
              },
              ticks: {
                autoSkip: true,
                maxTicksLimit: 7
              },
              title: {
                display: true,
                text: xLabel
              },
              grid: {
                display: false
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
              },
              grid: {
                color: 'rgba(167,142,247,0.3)'
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
