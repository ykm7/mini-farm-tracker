<script setup lang="ts">
import type { ChartData, ChartOptions, Point, ChartDataset } from 'chart.js'
import type { CalibratedData, RawData } from '@/models/Data'
import { computed, ref } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart,
  TimeScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'
import 'chartjs-adapter-moment'
Chart.register(TimeScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

export interface DisplayPoint {
  value: number
  timestamp: string 
}

const props = defineProps<{
  //   dataSets: ChartDataset<"line", Point[]>[]
  displayData: DisplayPoint[]
  lineLabel: string
  emptyLabel: string
  title: string
  yAxisUnit: 'mm' | 'cm' | 'm' | 'm³' | 'L'
}>()

const rawDataGraph = computed<ChartData<'line', Point[]>>(() => {
  return {
    datasets: [
      {
        label: props.lineLabel,
        data: props.displayData
          ? props.displayData.map<Point>((v) => {
              return {
                x: v.timestamp as unknown as number, // TODO: FIX! I should be able to use the explicit casting above but this causes the 'Line' component to have issues
                y: v.value
              }
            })
          : [],
      },
    ],
  }
})

const dynamicTimeUnit = (dataPoints: DisplayPoint[]) => {
  const oldest = dataPoints[0]
  const newest = dataPoints[dataPoints.length-1]

  const diff = Date.parse(newest.timestamp) - Date.parse(oldest.timestamp)

  const hours = diff / (1000 * 60 * 60)

  if (hours < 1) return 'minute'
  if (hours < 24) return 'hour'
  if (hours < 30 * 24) return 'day'
  if (hours < 365 * 24) return 'month'
  return 'year'
}

const chartOptions = computed<ChartOptions<'line'>>(() => {
  return {
    responsive: true,
    maintainAspectRatio: false,
    aspectRatio: 2,
    scales: {
      x: {
        type: 'time',
        time: {
          unit: rawDataGraph.value
            ? dynamicTimeUnit(props.displayData)
            : undefined,
          displayFormats: {
            minute: 'HH:mm',
            hour: 'DD MMM HH:mm',
            day: 'DD MMM YYYY',
            month: 'MMM YYYY',
            year: 'YYYY',
          },
        },
        ticks: {
          color: 'black',
        },
        grid: {
          color: 'rgba(255,255,255,0.2)',
        },
        title: {
          display: true,
          text: 'Timestamp',
        },
      },
      y: {
        title: {
          display: true,
          text: `Value (${props.yAxisUnit})`,
        },
        ticks: {
          color: 'black',
        },
        grid: {
          color: 'rgba(255,255,255,0.2)',
        },
      },
    },
    plugins: {
      title: {
        display: true,
        text: props.title,
      },
    },
    elements: {
      line: {
        borderColor: 'black',
        backgroundColor: 'black',
      },
      point: {
        borderColor: 'black',
        backgroundColor: 'black',
      },
    },
  }
})
</script>

<template>
  <Line
    v-if="displayData.length > 0 && rawDataGraph?.datasets.length > 0"
    class="graph-custom-wrapper"
    :options="chartOptions"
    :data="rawDataGraph"
  />
  <div v-else class="graph-custom-wrapper">{{ emptyLabel }}</div>
</template>

<style scoped>
.graph-custom-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 300px;
  background-color: rgba(0, 0, 0, 0.05);
  color: gray;
  border-radius: 8px;
}
</style>
