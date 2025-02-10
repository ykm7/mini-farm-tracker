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
      <!-- <div>
        {{  availableOptions }}
      </div> -->
      <Line
        v-if="rawDataGraph?.datasets.length > 0"
        class="graph-custom-wrapper"
        :options="chartOptions"
        :data="rawDataGraph"
      />
      <div v-else class="graph-custom-wrapper">{{ computedChartVisualSettings.emptyLabel }}</div>
    </div>
  </div>
</template>

<script setup lang="ts" generic="T">
import type { ChartData, ChartOptions, Point, ChartDataset } from 'chart.js'
import { computed, onMounted, ref, toRaw, watch } from 'vue'
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
import type { DisplayPoint, GraphData, GraphDataType, KeyOf, Unit } from '@/types/GraphRelated'

Chart.register(TimeScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

interface ChartVisualSettings {
  title: string
  lineLabel: string
  emptyLabel: string
}

const props = defineProps<{
  item: T
  displayData: GraphData
}>()

const computedDisplayData = computed(() => toRaw(props.displayData))

const computedChartVisualSettings = computed<ChartVisualSettings>(() => {
  const currentData = computedDisplayData.value
  
  if (currentData.Raw) {
    return {
      title: 'Distance measured by sensor',
      emptyLabel: 'No data available for this sensor',
      lineLabel: 'Distance',
    }
  } else if (currentData.Volume) {
    return {
      title: 'Water in tank',
      emptyLabel: 'No calibrated data available for this sensor',
      lineLabel: 'Litres',
    }
  } else if (currentData.AirTemperature) {
    return {
      title: 'Current air temperature',
      emptyLabel: 'No calibrated data available for this sensor',
      lineLabel: '℃',
    }
  }
  else {
    return {
      emptyLabel: 'LABEL UNKNOWN',
      title: 'TITLE UNKNOWN',
      lineLabel: 'LINE LABEL UNKNOWN',
    }
  }
})

const emit = defineEmits<{
  (e: 'update-starting-date', item: T, startingOffset: number): void
}>()

const selectedPeriod = ref(0)

interface SelectedGraph {
  key: KeyOf<GraphData>
  value: GraphDataType
}

const availableOptions = computed<KeyOf<GraphData>[]>(() => {
  const keys = Object.keys(props.displayData)
  if (keys.length === 0) {
    return []
  } else {
    return keys as KeyOf<GraphData>[]
  }
})

const selectedGraphType = ref<SelectedGraph | undefined>()

onMounted(() => {
  selectTimePeriod(ONE_WEEK)
  // setDefaultGraph(props.displayData)
})

watch(
  computedDisplayData,
  (newMap, oldMap) => {
    // console.log("🚀 ~ oldMap:", oldMap)
    // // console.log("🚀 ~ newMap, oldMap:", newMap, oldMap)

    setDefaultGraph(newMap)
  },
  { deep: true },
)

const selectTimePeriod = (period: number) => {
  selectedPeriod.value = period
  emit('update-starting-date', props.item, period)
}

const setDefaultGraph = (displayData: GraphData) => {
  // console.log("🚀 ~ setDefaultGraph ~ displayData:", displayData)
  // If "raw" is set, that should be the only entry.

  // TODO: Very brittle currently... only allowing for a single value to be available.
  const singleOption = displayData[availableOptions.value[0]]

  if (singleOption != null) {
    selectedGraphType.value = {
      key: 'Raw',
      value: singleOption,
    }
  }
}

const rawDataGraph = computed<ChartData<'line', Point[]>>(() => {
  const current = selectedGraphType.value
  // console.log("🚀 ~ current:", current)
  if (current == null) {
    return {
      datasets: [],
    }
  }

  return {
    datasets: [
      {
        label: computedChartVisualSettings.value.lineLabel,
        data: props.displayData
          ? current.value.data.map<Point>((v) => {
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
  const current = selectedGraphType.value
  if (current == null) {
    return {}
  }

  return {
    responsive: true,
    maintainAspectRatio: false,
    aspectRatio: 2,
    scales: {
      x: {
        type: 'time',
        time: {
          unit: rawDataGraph.value ? dynamicTimeUnit(current.value.data) : undefined,
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
          text: `Value (${current.value.unit})`,
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
        text: computedChartVisualSettings.value.title,
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
