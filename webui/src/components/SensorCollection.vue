<template>
  <div>
    <h4>Sensor Collection</h4>
    <div>
      <a>Available sensors</a>
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
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const BASE_URL: string = import.meta.env.VITE_BASE_URL

import TimeseriesGraph from './TimeseriesGraph.vue'
import axios from 'axios'
import type { RawData } from '@/models/Data'
import type { Sensor } from '@/models/Sensor'
const message = ref('')
const availableSensors = ref<Sensor[]>([])
const selectedSensor = ref<Sensor | undefined>(undefined)
const dataPull = ref<boolean>(false)
const rawData = ref<RawData[]>([])

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
