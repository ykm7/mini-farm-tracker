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
      <div
        class="graph-custom-wrapper-group graph-custom-wrapper"
        v-if="rawDataGraph?.datasets.length > 0"
      >
        <div class="available-graph-data-options">
          <button
            @click="selectGraphOption(option)"
            :class="{ selected: selectedGraphType?.key === option }"
            v-for="option in availableOptions"
          >
            {{ option }}
          </button>
        </div>
        <Line :options="chartOptions" :data="rawDataGraph" />
      </div>

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
  const currentData = toRaw(selectedGraphType.value)
  if (currentData == null) {
    return {
      emptyLabel: 'LABEL UNKNOWN',
      title: 'TITLE UNKNOWN',
      lineLabel: 'LINE LABEL UNKNOWN',
    }
  }

  const key = currentData.key
  const lineLabel = currentData.value.unit

  switch (key) {
    case 'Raw':
      return {
        title: 'Distance measured by sensor',
        emptyLabel: 'No data available for this sensor',
        lineLabel: 'Distance',
      }

    case 'Volume':
      return {
        title: 'Water in tank',
        emptyLabel: 'No calibrated data available for this sensor',
        lineLabel: lineLabel,
      }

    case 'AirTemperature':
      return {
        title: 'Current air temperature',
        emptyLabel: 'No calibrated data available for this sensor',
        lineLabel: lineLabel,
      }

    case 'AirHumidity':
      return {
        title: 'Current air humidity',
        emptyLabel: 'No calibrated data available for this sensor',
        lineLabel: lineLabel,
      }

    case 'LightIntensity':
      return {
        title: 'Current light intensity',
        emptyLabel: 'No calibrated data available for this sensor',
        lineLabel: lineLabel,
      }

    case 'UvIndex':
      return {
        title: 'Current UV index',
        emptyLabel: 'No calibrated data available for this sensor',
        lineLabel: lineLabel,
      }

    case 'WindSpeed':
      return {
        title: 'Current wind speed',
        emptyLabel: 'No calibrated data available for this sensor',
        lineLabel: lineLabel,
      }

    case 'WindDirection':
      return {
        title: 'Current wind direction',
        emptyLabel: 'No calibrated data available for this sensor',
        lineLabel: lineLabel,
      }

    case 'RainfallHourly':
      return {
        title: 'Current hourly rainfall',
        emptyLabel: 'No calibrated data available for this sensor',
        lineLabel: lineLabel,
      }

    case 'BarometricPressure':
      return {
        title: 'Current barometric pressure',
        emptyLabel: 'No calibrated data available for this sensor',
        lineLabel: lineLabel,
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
  console.log('🚀 ~ keys:', keys)
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
    setDefaultGraph(newMap)
  },
  { deep: true },
)

const selectTimePeriod = (period: number) => {
  selectedPeriod.value = period
  emit('update-starting-date', props.item, period)
}

const selectGraphOption = (option: keyof GraphData) => {
  selectedGraphType.value = {
    key: option,
    value: props.displayData[option]!,
  }
}

const setDefaultGraph = (displayData: GraphData) => {
  selectedGraphType.value = {
    key: availableOptions.value[0],
    value: displayData[availableOptions.value[0]]!,
  }
}

const rawDataGraph = computed<ChartData<'line', Point[]>>(() => {
  const current = toRaw(selectedGraphType.value)
  // console.log("🚀 ~ current:", current)
  if (current == null) {
    return {
      datasets: [],
    }
  }

  if (props.displayData == null) {
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
  console.log('🚀 ~ current:', current)
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
      font-size: x-small;
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

  .graph-buttons {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
  }

  .graph-wrapper {
    flex-grow: 1;
    flex-shrink: 1;
    max-width: 100%;
    min-width: 0;

    .graph-custom-wrapper {
      /* flex-grow: 1; */
      display: flex;
      justify-content: center;
      align-items: center;
      height: 325px;
      background-color: rgba(0, 0, 0, 0.05);
      color: gray;
      border-radius: 8px;
    }

    .graph-custom-wrapper-group {
      display: flex;
      flex-direction: column;

      .available-graph-data-options {
        flex-grow: 0;
      }

      canvas {
        min-height: 0;
      }
    }
  }
}
</style>
