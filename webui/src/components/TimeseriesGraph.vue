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
            :class="{ selected: selectedGraphType && selectedGraphType.key === option }"
            v-for="option in availableOptions"
            :key="option"
          >
            {{ option }}
          </button>
        </div>
        <div class="canvas-wrapper">
          <Line :options="chartOptions" :data="rawDataGraph" />
        </div>
      </div>

      <div v-else class="graph-custom-wrapper">{{ computedChartVisualSettings.emptyLabel }}</div>
    </div>
  </div>
</template>

<script setup lang="ts" generic="T">
  import { ALL_YEARS, ONE_DAY, ONE_HOUR, ONE_MONTH, ONE_WEEK, ONE_YEAR } from "@/helper"
  import {
    dynamicTimeUnit,
    type GraphData,
    type GraphDataType,
    type KeyOf,
  } from "@/types/GraphRelated"
  import type { ChartData, ChartOptions, Point } from "chart.js"
  import {
    Chart,
    Legend,
    LinearScale,
    LineElement,
    PointElement,
    TimeScale,
    Title,
    Tooltip,
  } from "chart.js"
  import "chartjs-adapter-moment"
  import { computed, onMounted, ref, toRaw, watch } from "vue"
  import { Line } from "vue-chartjs"

  Chart.register(TimeScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

  interface SelectedGraph {
    key: KeyOf<GraphData>
    // This is undefined possible for cases where historically we might not have a data type over a particular time period.
    // Noticed when a "new" data type was add but historically they weren't present.
    value?: GraphDataType
  }

  interface ChartVisualSettings {
    title: string
    lineLabel: string
    emptyLabel: string
  }

  const props = defineProps<{
    item: T
    displayData: GraphData
  }>()

  const selectedGraphType = ref<SelectedGraph | undefined>(undefined)
  const selectedPeriod = ref(0)

  const computedDisplayData = computed(() => toRaw(props.displayData))

  const computedChartVisualSettings = computed<ChartVisualSettings>(() => {
    const currentData = toRaw(selectedGraphType.value)

    const EMPTY = {
      emptyLabel: "LABEL UNKNOWN",
      title: "TITLE UNKNOWN",
      lineLabel: "LINE LABEL UNKNOWN",
    }

    if (currentData == null) {
      return EMPTY
    }

    const key = currentData.key
    if (key == null) {
      return EMPTY
    }

    const lineLabel = currentData.value != null ? currentData.value.unit : "unknown"

    switch (key) {
      case "Raw":
        return {
          title: "Distance measured by sensor",
          emptyLabel: "No data available for this sensor",
          lineLabel: "Distance",
        }

      case "Volume":
        return {
          title: "Water in tank",
          emptyLabel: "No calibrated data available for this sensor",
          lineLabel: lineLabel,
        }

      case "AirTemperature":
        return {
          title: "Current air temperature",
          emptyLabel: "No calibrated data available for this sensor",
          lineLabel: lineLabel,
        }

      case "AirHumidity":
        return {
          title: "Current air humidity",
          emptyLabel: "No calibrated data available for this sensor",
          lineLabel: lineLabel,
        }

      case "LightIntensity":
        return {
          title: "Current light intensity",
          emptyLabel: "No calibrated data available for this sensor",
          lineLabel: lineLabel,
        }

      case "UvIndex":
        return {
          title: "Current UV index",
          emptyLabel: "No calibrated data available for this sensor",
          lineLabel: lineLabel,
        }

      case "WindSpeed":
        return {
          title: "Current wind speed",
          emptyLabel: "No calibrated data available for this sensor",
          lineLabel: lineLabel,
        }

      case "WindDirection":
        return {
          title: "Current wind direction",
          emptyLabel: "No calibrated data available for this sensor",
          lineLabel: lineLabel,
        }

      case "RainGauge":
        return {
          title: "Rainfall intensity (hourly)",
          emptyLabel: "No calibrated data available for this sensor",
          lineLabel: lineLabel,
        }

      case "BarometricPressure":
        return {
          title: "Current barometric pressure",
          emptyLabel: "No calibrated data available for this sensor",
          lineLabel: lineLabel,
        }

      case "PeakWindGust":
        return {
          title: "Peak Wind Gust",
          emptyLabel: "No calibrated data available for this sensor",
          lineLabel: lineLabel,
        }

      case "RainAccumulation":
        return {
          title: "Rain Accumulation",
          emptyLabel: "No calibrated data available for this sensor",
          lineLabel: lineLabel,
        }
    }

    return EMPTY
  })

  const emit = defineEmits<{
    (e: "update-starting-date", item: T, startingOffset: number): void
  }>()

  onMounted(() => {
    selectTimePeriod(ONE_WEEK)
  })

  watch(
    computedDisplayData,
    (newMap) => {
      setDefaultGraph(newMap)
    },
    { deep: true }
  )

  const availableOptions = computed<KeyOf<GraphData>[]>(() => {
    const keys = Object.keys(props.displayData)
    if (keys.length === 0) {
      return []
    } else {
      return keys as KeyOf<GraphData>[]
    }
  })

  const selectTimePeriod = (period: number) => {
    selectedPeriod.value = period
    emit("update-starting-date", props.item, period)
  }

  const selectGraphOption = (option: keyof GraphData) => {
    selectedGraphType.value = {
      key: option,
      value: toRaw(props.displayData[option]!),
    }
  }

  const rawDataGraph = computed<ChartData<"line", Point[]>>(() => {
    const current = toRaw(selectedGraphType.value)

    // TODO: Figure out why I need the .key check
    if (current == null || current.key == null) {
      return {
        datasets: [],
      }
    }

    if (props.displayData == null) {
      return {
        datasets: [],
      }
    }

    // TODO: Add https://www.chartjs.org/docs/latest/samples/advanced/data-decimation.html for significant data points
    return {
      datasets: [
        {
          label: computedChartVisualSettings.value.lineLabel,
          data:
            props.displayData && current.value
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

  const chartOptions = computed<ChartOptions<"line">>(() => {
    const current = toRaw(selectedGraphType.value)
    if (current == null || current.value == null) {
      return {}
    }

    return {
      responsive: true,
      maintainAspectRatio: false,
      aspectRatio: 2,
      scales: {
        x: {
          type: "time",
          time: {
            unit: rawDataGraph.value != null ? dynamicTimeUnit(current.value.data) : undefined,
            displayFormats: {
              minute: "HH:mm",
              hour: "DD MMM HH:mm",
              day: "DD MMM YYYY",
              month: "MMM YYYY",
              year: "YYYY",
            },
          },
          ticks: {
            color: "black",
            autoSkip: true,
          },
          grid: {
            color: "rgba(255,255,255,0.2)",
          },
          title: {
            display: true,
            text: "Timestamp",
          },
        },
        y: {
          beginAtZero: true,
          title: {
            display: true,
            text: `Value (${current.value.unit})`,
          },
          ticks: {
            color: "black",
          },
          grid: {
            color: "rgba(255,255,255,0.2)",
          },
        },
      },
      plugins: {
        title: {
          display: true,
          text: titleValue.value,
        },
      },
      elements: {
        line: {
          borderColor: "black",
          backgroundColor: "black",
        },
        point: {
          borderColor: "black",
          backgroundColor: "black",
        },
      },
    }
  })

  /**
   * Very rough function to provide a accumulated value for the Rain Accumulation graph
   */
  const titleValue = computed(() => {
    if (computedChartVisualSettings.value == null) {
      return "No data available"
    }

    if (computedChartVisualSettings.value.title== "Rain Accumulation") {

      const current = toRaw(selectedGraphType.value)
      const accumulated = current?.value?.data.reduce((acc, v) => {
        acc += v.value;
        return acc
      }, 0);

      if (accumulated != null) {
        return computedChartVisualSettings.value.title + ": " + Number(accumulated.toFixed(2)) + " mm";
      }
    }

    return computedChartVisualSettings.value.title
  })
  

  const setDefaultGraph = (displayData: GraphData) => {
    var key: keyof GraphData
    if (selectedGraphType.value != null && selectedGraphType.value.key) {
      // a subgraph type has already been selected
      key = selectedGraphType.value.key
    } else {
      // on initial local there isn't a selected sub graph type
      key = availableOptions.value[0]
    }

    selectedGraphType.value = {
      key: key,
      value: displayData[key]!,
    }
  }
</script>

<style scoped>
  /* .graph-top-wrapper {
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
      display: flex;
      flex-direction: column;

      .graph-custom-wrapper {
        flex-grow: 1;
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
        width: 100%;
        height: 100%;

        .available-graph-data-options {
          flex: 0;
        }

        .canvas-wrapper {
          min-height: 0;
          flex-grow: 1;
          width: 100%;
        }

        canvas {
          min-height: 0;
        }
      }
    }
  } */
</style>
