<script setup lang="ts">
import type { ChartData, ChartOptions, Point, ChartDataset } from 'chart.js'
import type { RawData } from '@/models/Data'
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

interface TimePoint {
  x: Date // Date as string
  y: number
}

const props = defineProps<{
  //   dataSets: ChartDataset<"line", Point[]>[]
  rawData: RawData[]
  emptyLabel: string
  yAxisUnit: 'mm' | 'cm' | 'm' | 'm³' | 'L'
}>()

const rawDataGraph = computed<ChartData<'line', Point[]>>(() => {
  return {
    datasets: [
      // {
      //   label: 'Battery',
      //   data: props.rawData
      //     ? props.rawData.map<Point>((v) => {
      //         return {
      //           x: v.Timestamp as unknown as number, // TODO: FIX! I should be able to use the explicit casting above but this causes the 'Line' component to have issues
      //           y: v.Data.Bat,
      //         }
      //       })
      //     : [],
      // },
      {
        label: `Distance`,
        data: props.rawData
          ? props.rawData.map<Point>((v) => {
              return {
                x: v.Timestamp as unknown as number, // TODO: FIX! I should be able to use the explicit casting above but this causes the 'Line' component to have issues
                y: v.Data.Distance.split(' ')[0] as unknown as number,
              }
            })
          : [],
      },
      // {
      //   label: `Raw data for: ${props.rawData?.length > 0 ? props.rawData[0].Sensor : 'Unknown'}`,
      //   data: props.rawData
      //     ? props.rawData.map<Point>((v) => {
      //         return {
      //           x: v.Timestamp as unknown as number, // TODO: FIX! I should be able to use the explicit casting above but this causes the 'Line' component to have issues
      //           y: v.Data.Temperature,
      //         }
      //       })
      //     : [],
      // },
      // {
      //   label: `Raw data for: ${props.rawData?.length > 0 ? props.rawData[0].Sensor : 'Unknown'}`,
      //   data: props.rawData

      //     ? props.rawData.map<Point>((v) => {
      //         return {
      //           x: v.Timestamp as unknown as number, // TODO: FIX! I should be able to use the explicit casting above but this causes the 'Line' component to have issues
      //           y: v.Data.SensorFlag,
      //         }
      //       })
      //     : [],
      // },
    ],
  }
})

const determineTimeUnit = (dataPoints: Point[]) => {
  const timeDiffs = dataPoints.slice(1).map((point, index) => {
    const prevTime = new Date(dataPoints[index].x).getTime()
    const currTime = new Date(point.x).getTime()
    return currTime - prevTime
  })

  const avgTimeDiff = timeDiffs.reduce((a, b) => a + b, 0) / timeDiffs.length
  const hours = avgTimeDiff / (1000 * 60 * 60)

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
            ? determineTimeUnit(rawDataGraph.value.datasets[0].data)
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
        text: 'Time Series Chart',
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
    v-if="rawData.length > 0 && rawDataGraph?.datasets.length > 0"
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
