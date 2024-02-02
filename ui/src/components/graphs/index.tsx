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
import { Bar } from 'react-chartjs-2';

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
  return (
    <>
      <Bar
        data={{
          labels: timestamps, // (timestamps && timestamps.length > 0) ? [timestamps[0], timestamps[timestamps.length / 2], timestamps[timestamps.length - 1]] : [],
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
                text: 'Time'
              }
            },
            y: {
              title: {
                display: true,
                text: 'Number of evaluations'
              }
            }
          }
        }}
      />
    </>
  );
}
