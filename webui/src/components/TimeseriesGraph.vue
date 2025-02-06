<template>
  <div class="graph-top-wrapper">
    <div class="graph-buttons">
      <button
        :class="{ selected: selectedPeriod === ONE_HOUR }"
        @click="selectTimePeriod(ONE_HOUR)"
      >
        1 Hour
      </button>
      <button :class="{ selected: selectedPeriod === ONE_DAY }" @click="selectTimePeriod(ONE_DAY)">
        24 Hours
      </button>
      <button
        :class="{ selected: selectedPeriod === ONE_WEEK }"
        @click="selectTimePeriod(ONE_WEEK)"
      >
        7 Days
      </button>
      <button
        :class="{ selected: selectedPeriod === ONE_MONTH }"
        @click="selectTimePeriod(ONE_MONTH)"
      >
        1 Month
      </button>
      <button
        :class="{ selected: selectedPeriod === ONE_YEAR }"
        @click="selectTimePeriod(ONE_YEAR)"
      >
        1 Year
      </button>
      <button
        :class="{ selected: selectedPeriod === ALL_YEARS }"
        @click="selectTimePeriod(ALL_YEARS)"
      >
        ALL
      </button>
    </div>

    <div class="graph-wrapper">
      <Line
        v-if="displayData.length > 0 && rawDataGraph?.datasets.length > 0"
        class="graph-custom-wrapper"
        :options="chartOptions"
        :data="rawDataGraph"
      />
      <div v-else class="graph-custom-wrapper">{{ emptyLabel }}</div>
    </div>
  </div>
</template>

<script setup lang="ts" generic="T">
import type { ChartData, ChartOptions, Point, ChartDataset } from 'chart.js'
import { computed, onMounted, ref } from 'vue'
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
import { ONE_DAY, ONE_HOUR, ONE_MONTH, ONE_WEEK, ONE_YEAR, ALL_YEARS } from '@/helper'
import type { DisplayPoint, Unit } from '@/types/GraphRelated'

Chart.register(TimeScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

const props = defineProps<{
  item: T
  displayData: DisplayPoint[]
  lineLabel: string
  emptyLabel: string
  title: string
  yAxisUnit: Unit
}>()

const emit = defineEmits<{
  (e: 'update-starting-date', item: T, startingOffset: number): void
}>()

const selectedPeriod = ref(0)

onMounted(() => {
  // Simply tester
  // let counter = 1
  // setInterval(() => {
  //   emit('update-starting-date', props.item, ONE_DAY * counter)
  //   counter++
  // }, 5000)
  // emit('update-starting-date', props.item, ONE_WEEK)
  selectTimePeriod(ONE_WEEK)
})

const selectTimePeriod = (period: number) => {
  selectedPeriod.value = period
  emit('update-starting-date', props.item, period)
}

const rawDataGraph = computed<ChartData<'line', Point[]>>(() => {
  return {
    datasets: [
      {
        label: props.lineLabel,
        data: props.displayData
          ? props.displayData.map<Point>((v) => {
              return {
                x: v.timestamp as unknown as number, // TODO: FIX! I should be able to use the explicit casting above but this causes the 'Line' component to have issues
                y: v.value,
              }
            })
          : [],
      },
    ],
  }
})

const dynamicTimeUnit = (dataPoints: DisplayPoint[]) => {
  const oldest = dataPoints[0]
  const newest = dataPoints[dataPoints.length - 1]

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
          unit: rawDataGraph.value ? dynamicTimeUnit(props.displayData) : undefined,
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

<style scoped>
.graph-top-wrapper {
  display: flex;

  .graph-buttons {
    display: flex;
    flex-direction: column;
    justify-content: space-between;

    button {
      flex: auto;
      background-color: #42b883;
      color: #ffffff;
      border: none;
      padding: 10px 15px;
      cursor: pointer;
      border-radius: 4px;
      font-size: 14px;
      transition: background-color 0.3s ease;
    }

    @media (max-width: 1024px) {
      button {
        padding: 5px 7px;
      }
    }

    button:hover {
      background-color: #36495d;
    }

    button.selected {
      background-color: #36495d;
      box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.2);
      transform: translateY(1px);
    }
  }

  .graph-wrapper {
    flex-grow: 1;
    flex-shrink: 1;
    max-width: 100%;
    min-width: 0;

    .graph-custom-wrapper {
      display: flex;
      justify-content: center;
      align-items: center;
      height: 300px;
      background-color: rgba(0, 0, 0, 0.05);
      color: gray;
      border-radius: 8px;
    }
  }
}
</style>
