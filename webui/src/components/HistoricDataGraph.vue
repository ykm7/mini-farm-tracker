<template>
  <div class="graph-top-wrapper">
    <!-- <div class="graph-buttons"></div> -->
    <div class="graph-wrapper">
      <div class="graph-custom-wrapper-group graph-custom-wrapper">
        <Bar :options="chartOptions" :data="rawDataGraph" />
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
  import {
    type AggregatedDataPoint,
    type CalibratedDataNamesGrouping,
    type ExtendedDataPoint,
  } from "@/types/GraphRelated"
  import {
    BarElement,
    CategoryScale,
    Chart,
    Legend,
    LinearScale,
    TimeScale,
    Title,
    Tooltip,
    type ChartData,
    type ChartDataset,
    type ChartOptions,
  } from "chart.js"
  import "chartjs-adapter-moment"
  import { computed } from "vue"
  import { Bar } from "vue-chartjs"

  Chart.register(CategoryScale, LinearScale, BarElement, TimeScale, Title, Tooltip, Legend)

  const props = defineProps<{
    data: Partial<CalibratedDataNamesGrouping>
  }>()

  const chartOptions = computed<ChartOptions<"bar">>(() => {
    return {
      responsive: true,
      maintainAspectRatio: false,
      scales: {
        y: {
          beginAtZero: true,
          title: {
            display: true,
            text: "Rainfall (mm)",
          },
          ticks: {
            callback: (value) => `${value} mm`,
          },
        },
        x_day: {
          type: "time",
          time: {
            parser: "YYYY-MM-DD",
            unit: "day",
            displayFormats: {
              day: "MMM DD",
            },
          },
          ticks: {
            source: "data",
            align: "start",
            autoSkip: true,
            maxTicksLimit: 10,
          },
        },
        x_week: {
          type: "time",
          time: {
            parser: "YYYY-[W]WW",
            unit: "week",
            displayFormats: {
              week: "YYYY [W]WW",
            },
          },
          ticks: {
            source: "data",
            align: "start",
            autoSkip: true,
            maxTicksLimit: 10,
          },
        },
        x_month: {
          type: "time",
          time: {
            parser: "YYYY-MM-DD",
            unit: "month",
            displayFormats: {
              month: "MMM YYYY",
            },
          },
          ticks: {
            source: "data",
            align: "start",
            autoSkip: true,
            maxTicksLimit: 12,
          },
        },
        x_year: {
          type: "time",
          time: {
            parser: "YYYY-MM-DD",
            unit: "year",
            displayFormats: {
              year: "YYYY",
            },
          },
          ticks: {
            source: "data",
            align: "start",
            autoSkip: false,
          },
        },
      },
      plugins: {
        legend: {
          position: "top",
        },
        tooltip: {
          mode: "x",
          callbacks: {
            // title: (tooltipItems) => {
            //   const item = tooltipItems[0]
            //   return `${item.dataset.label} - ${item.label}`
            // },
            label: (context) => {
              return `Rainfall: ${context.parsed.y.toFixed(2)} mm`
            },
            // TODO: Expand on this in the future
            // footer: (tooltipItems) => {
            //   const value = tooltipItems[0].parsed.y
            //   if (value > 50) return "Heavy rainfall"
            //   if (value > 25) return "Moderate rainfall"
            //   return "Light rainfall"
            // },
          },
        },
      },
      elements: {
        bar: {
          borderWidth: 1,
          hoverBackgroundColor: "rgba(75, 192, 192, 0.8)",
          hoverBorderColor: "rgb(75, 192, 192)",
          hoverBorderWidth: 2,
        },
      },
      barPercentage: 1.0,
      categoryPercentage: 1.0,
    }
  })

  function getColor(value: number) {
  const intensity = value / 100; // Adjust based on your max rainfall
  return `rgba(0, 0, 255, ${intensity})`;
}

  const rawDataGraph = computed<ChartData<"bar", ExtendedDataPoint[]>>(() => {
    const data: ChartDataset<"bar", ExtendedDataPoint[]>[] = []

    const rain = props.data.RAIN_FALL_HOURLY

    if (!rain) {
      return {
        datasets: [],
      }
    }

    if (rain.data.DAILY) {
      data.push({
        xAxisID: "x_day",
        label: "DAILY",
        data: rain.data.DAILY.map((d: AggregatedDataPoint) => {
          return {
            x: d.date,
            y: d.value,
          }
        }),
        backgroundColor: "rgba(255, 99, 132, 0.2)",
        // TODO: maybe play with this when we have more data, looks weird currently.
        // backgroundColor: (context) => {
        //   console.log("🚀 ~ context:", context)
        //   const dataPoint = context.raw as ExtendedDataPoint;
        //   return getColor(dataPoint.y);
        // },
        borderColor: "rgb(255, 99, 132)",
        borderWidth: 1,
      })
    }

    if (rain.data.WEEKLY) {
      data.push({
        xAxisID: "x_week",
        label: "WEEKLY",
        data: rain.data.WEEKLY.map((d: AggregatedDataPoint) => {
          return {
            x: d.date,
            y: d.value,
          }
        }),
        backgroundColor: "rgba(255, 159, 64, 0.2)",
        borderColor: "rgb(255, 159, 64)",
        borderWidth: 1,
      })
    }

    if (rain.data.MONTHLY) {
      data.push({
        xAxisID: "x_month",
        label: "MONTHLY",
        data: rain.data.MONTHLY.map((d: AggregatedDataPoint) => {
          return {
            x: d.date,
            y: d.value,
          }
        }),
        backgroundColor: "rgba(255, 205, 86, 0.2)",
        borderColor: "rgb(255, 205, 86)",
        borderWidth: 1,
      })
    }

    if (rain.data.YEARLY) {
      data.push({
        xAxisID: "x_year",
        label: "YEARLY",
        data: rain.data.YEARLY.map((d: AggregatedDataPoint) => {
          return {
            x: d.date,
            y: d.value,
          }
        }),
        backgroundColor: "rgba(75, 192, 192, 0.2)",
        borderColor: "rgb(75, 192, 192)",
        borderWidth: 1,
      })
    }

    return {
      datasets: data,
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
        /* width: 100%; */
        display: flex;
        flex-direction: column;

        .available-graph-data-options {
          flex-grow: 0;
        }

        canvas {
          min-height: 0;
          /* width: 100%; */
        }
      }
    }
  }
</style>
