<script setup lang="ts">
import type { RawData } from '@/models/Data'
import type { Sensor } from '@/models/Sensor'
import axios from 'axios'
import { ref } from 'vue'

import TimeseriesGraph from './TimeseriesGraph.vue'
const BASE_URL: string = import.meta.env.VITE_BASE_URL

const message = ref('')
const availableSensors = ref<Sensor[]>([])
const selectedSensor = ref<Sensor | undefined>(undefined)
const dataPull = ref<boolean>(false)

const rawData = ref<RawData[]>([])

const pingServerFn = async () => {
  console.log('Attempting to ping server')
  try {
    await axios.get(`${BASE_URL}/ping`)
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
    <TimeseriesGraph :rawData="rawData" />
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
</style>
