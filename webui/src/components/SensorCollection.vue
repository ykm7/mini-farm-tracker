<template>
  <div>
    <h4>Sensor Collection</h4>
    <div>
      <a>Available sensors</a>

      <div v-for="sensor in sensors">
        <CCard style="margin: 0.5rem 0">
          <CCardTitle>{{ sensor.Id }}</CCardTitle>
          <!-- <CCardSubtitle class="mb-2 text-body-secondary">{{ asset.Id }}</CCardSubtitle> -->
          <CCardBody>{{ sensor.Description }}</CCardBody>

          <div class="group-section">
            <!-- {{  sensorToData.get(sensor.Id) }} -->

            <Suspense>
              <template #default>
                <div>
                <AsyncWrapper :promise="pullSensorsRawDataFn(sensor)">
                  <template v-slot="{ data }">
                    <div v-if="data">
                      <!-- {{ data[0] }} -->
                      <TimeseriesGraph :rawData="data" />
                    </div>
                  </template>
                </AsyncWrapper>
              </div>
              </template>
              <template #fallback>
                <div>Loading...</div>
              </template>
            </Suspense>
          </div>
        </CCard>
      </div>

      <!-- <button class="button" @click="pullSensorsFn">Pull sensors</button>

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
        </table> -->
      <!-- </div> -->
    </div>

    <!-- <br /> -->

    <!-- <div v-if="selectedSensor">
      <a>Raw data</a>
      <TimeseriesGraph :rawData="rawData" />
    </div> -->
  </div>
</template>

<script setup lang="ts">
import {
  CCard,
  CCardBody,
  CCardTitle,
  CCardSubtitle,
  CListGroup,
  CListGroupItem,
} from '@coreui/vue'
import { Suspense } from 'vue'
import AsyncWrapper from './AsyncWrapper.vue'
import { computed, ref, watch, watchEffect, type ComputedRef } from 'vue'
import { toRaw } from 'vue'

const BASE_URL: string = import.meta.env.VITE_BASE_URL

import TimeseriesGraph from './TimeseriesGraph.vue'
import axios from 'axios'
import type { RawData } from '@/models/Data'
import type { Sensor } from '@/models/Sensor'
import { useSensorStore } from '@/stores/sensor'
const availableSensors = ref<Sensor[]>([])
const selectedSensor = ref<Sensor | undefined>(undefined)
const dataPull = ref<boolean>(false)
const rawData = ref<RawData[]>([])

const sensorCollection = useSensorStore()

const sensors = computed<Sensor[]>(() => sensorCollection.sensors)

// const sensorToData = ref<Map<string, RawData[]>>(new Map<string, RawData[]>())

// watch(
//   sensors,
//   (newSensor, oldSensor) => {
//     console.log('🚀 ~ watch ~ oldSensor:', oldSensor)
//     console.log('🚀 ~ watch ~ newSensor:', newSensor)

//     newSensor.forEach(async (s) => {
//       const data = await pullSensorsRawDataFn(s)
//       console.log('🚀 ~ sensors.value.forEach ~ data:', data)
//       sensorToData.value.set(s.Id, data)
//     })
//   },
//   { deep: true },
// )

// const sensorToData: ComputedRef<Map<string, RawData[]>> = computed<Map<string, RawData[]>>(() => {
//   const dataMap = new Map<string, RawData[]>()

//   sensors.value.forEach(async (s) => {
//     const data = await pullSensorsRawDataFn(s)
//     console.log("🚀 ~ sensors.value.forEach ~ data:", data)
//     dataMap.set(s.Id, data)
//   })

//   return dataMap
// })

// const pullSensorsFn = async () => {
//   dataPull.value = true
//   try {
//     const response = await axios.get<Sensor[]>(`${BASE_URL}/api/sensors`)
//     availableSensors.value = response.data
//   } catch (e) {
//     console.warn(e)
//   }
// }

const pullSensorsRawDataFn = async (sensor: Sensor): Promise<RawData[]> => {
  try {
    // return []
    selectedSensor.value = sensor
    const response = await axios.get<RawData[]>(
      `${BASE_URL}/api/sensors/${sensor.Id}/data/raw_data`,
    )
    return response.data ? response.data : []
  } catch (e) {
    console.warn(e)
    return []
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
