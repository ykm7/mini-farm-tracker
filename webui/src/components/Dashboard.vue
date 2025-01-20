<script setup lang="ts">
import type { RawData } from '@/models/Data'
import type { Sensor } from '@/models/Sensor'
import axios from 'axios'
import type { ChartData, ChartOptions } from 'chart.js'
import { computed, ref } from 'vue'
import { Line } from 'vue-chartjs'
import { Chart, TimeScale, LinearScale, PointElement, LineElement, Title, Tooltip, Legend } from 'chart.js'
import 'chartjs-adapter-moment'

// Register necessary Chart.js components
Chart.register(
  TimeScale, 
  LinearScale, 
  PointElement, 
  LineElement, 
  Title, 
  Tooltip, 
  Legend
)

interface TimePoint {
  x: Date;  // Date as string
  y: number;
}

const BASE_URL: string = import.meta.env.VITE_BASE_URL

const message = ref('')
const availableSensors = ref<Sensor[]>([])
const rawData = ref<RawData[]>([])
  const chartOptions = ref<ChartOptions<'line'>>({
  responsive: true,
  scales: {
    x: {
      type: 'time',
      time: {
        unit: undefined, // 'day', // Adjust granularity as needed
        displayFormats: {
          day: 'MMM DD' // Customize date display format
        }
      },
      title: {
        display: true,
        text: 'Timestamp'
      }
    },
    y: {
      title: {
        display: true,
        text: 'Value'
      }
    }
  },
  plugins: {
    title: {
      display: true,
      text: 'Time Series Chart'
    }
  }
})

// const rawDataGraph = computed<ChartData<'line', TimePoint[]>>(() => {
const rawDataGraph = computed<ChartData<'line'>>(() => {
  return {
    datasets: [
      {
        label: `Raw data for: ${rawData.value[0].Sensor}`,
        data: rawData.value.map((v) => {
          return {
            x: v.Timestamp as unknown as number, // TODO: FIX! I should be able to use the explicit casting above but this causes the 'Line' component to have issues
            y: v.Data,
          }
        })
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
  try {
    const response = await axios.get<Sensor[]>(`${BASE_URL}/api/sensors`)
    availableSensors.value = response.data
    console.log('🚀 ~ pullSensorsFn ~ availableSensors.value:', availableSensors.value)

    if (availableSensors.value?.length > 0) {
      pullSensorsRawDataFn(availableSensors.value[0].Id)
    }
  } catch (e) {
    console.warn(e)
  }
}

const pullSensorsRawDataFn = async (sensorId: string) => {
  try {
    const response = await axios.get<RawData[]>(`${BASE_URL}/api/sensors/${sensorId}/data/raw_data`)
    rawData.value = response.data
    console.log('🚀 ~ pullSensorsRawDataFn ~ rawData.value:', rawData.value)
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
    <table>
      <thead>
        <tr>
          <th>Id</th>
          <th>Description</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="sensor in availableSensors" :key="sensor.Id">
          <td>{{ sensor.Id }}</td>
          <td>{{ sensor.Description }}</td>
        </tr>
      </tbody>
    </table>
  </div>

  <br />

  <div v-if="rawData.length > 0 && rawDataGraph?.datasets.length > 0">
    <a>Raw data for the first sensor</a>
    <!-- {{  rawDataGraph?.datasets }} -->
    <div class="container">
      <Line :options="chartOptions" :data="rawDataGraph" />
    </div>
  </div>
</template>
