<script setup lang="ts">
import type { RawData } from '@/models/Data'
import type { Sensor } from '@/models/Sensor'
import axios from 'axios'
import type { ChartData, ChartOptions } from 'chart.js'
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

// Register necessary Chart.js components
Chart.register(TimeScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend)

interface TimePoint {
  x: Date // Date as string
  y: number
}

const BASE_URL: string = import.meta.env.VITE_BASE_URL

const message = ref('')
const availableSensors = ref<Sensor[]>([])
const selectedSensor = ref<Sensor | undefined>(undefined)
const dataPull = ref<boolean>(false)

const rawData = ref<RawData[]>([])
const chartOptions = ref<ChartOptions<'line'>>({
  responsive: true,
  scales: {
    x: {
      type: 'time',
      time: {
        unit: undefined, // 'day', // Adjust granularity as needed
        displayFormats: {
          day: 'MMM DD', // Customize date display format
        },
      },
      ticks: {
        color: 'white',
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
        text: 'Value',
      },
      ticks: {
        color: 'white',
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
      borderColor: 'white',
      backgroundColor: 'white',
    },
    point: {
      borderColor: 'white',
      backgroundColor: 'white',
    },
  },
})

// const rawDataGraph = computed<ChartData<'line', TimePoint[]>>(() => {
const rawDataGraph = computed<ChartData<'line'>>(() => {
  return {
    datasets: [
      {
        label: `Raw data for: ${rawData.value[0].Sensor}`,
        data: rawData.value
          ? rawData.value.map((v) => {
              return {
                x: v.Timestamp as unknown as number, // TODO: FIX! I should be able to use the explicit casting above but this causes the 'Line' component to have issues
                y: v.Data,
              }
            })
          : [],
      },
    ],
  }
})

const pingServerFn = async () => {
  console.log('Attempting to ping server')
  try {
    const response = await axios.get(`${BASE_URL}/ping`)
    console.log(response)
    message.value = 'success'
  } catch (e) {
    console.warn(e)
    message.value = 'failure'
  }
}

const pullSensorsFn = async () => {
  dataPull.value = true
  try {
    const response = await axios.get<Sensor[]>(`${BASE_URL}/api/sensors`)
    availableSensors.value = response.data
  } catch (e) {
    console.warn(e)
  }
}

const pullSensorsRawDataFn = async (sensor: Sensor) => {
  try {
    selectedSensor.value = sensor
    const response = await axios.get<RawData[]>(
      `${BASE_URL}/api/sensors/${sensor.Id}/data/raw_data`,
    )
    rawData.value = response.data
    console.log(rawData.value)
  } catch (e) {
    console.warn(e)
  }
}
</script>

<template>
  <div>
    <div>
      <a>Basic PING test to the server.</a>
      <button class="button" @click="pingServerFn" title="ping server">ping server</button>
    </div>
    <a>Able to connect to server: {{ message }}</a>
  </div>

  <br />

  <div>
    <a>Available sensors <i>Note that these are currently mocked values (including the data)</i></a>
    <button class="button" @click="pullSensorsFn">Pull sensors</button>

    <br />

    <div v-if="dataPull">
      <i
        title="These are the raw values written by the installed device. Caliberated values to be added later"
        >Select a table row to pull 'raw' data</i
      >
      <table>
        <thead>
          <tr>
            <th>Id</th>
            <th>Description</th>
          </tr>
        </thead>
        <tbody>
          <tr
            class="cursor-pointer"
            @click="pullSensorsRawDataFn(sensor)"
            v-for="sensor in availableSensors"
            :key="sensor.Id"
          >
            <td>{{ sensor.Id }}</td>
            <td>{{ sensor.Description }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <br />

  <div v-if="selectedSensor">
    <a>Raw data</a>
    <div>
      <Line
        v-if="rawData.length > 0 && rawDataGraph?.datasets.length > 0"
        class="container"
        :options="chartOptions"
        :data="rawDataGraph"
      />
      <div v-else class="empty-chart-placeholder">No data available for this sensor</div>
    </div>
  </div>
</template>

<style scoped>
table {
  width: 100%;
  background-color: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(5px);
  border-radius: 0.5rem;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  border-spacing: 0;
}

thead {
  background-color: rgba(255, 255, 255, 0.2);
}

th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-weight: 600;
  border-bottom: 1px solid rgba(255, 255, 255, 0.3);
}

tr {
  cursor: pointer;
  transition: all 0.2s ease;
}

tr:hover {
  background-color: rgba(255, 255, 255, 0.2);
}

td {
  padding: 0.75rem 1rem;
}

.empty-chart-placeholder {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 300px;
  background-color: rgba(0, 0, 0, 0.05);
  color: gray;
  border-radius: 8px;
}
</style>
