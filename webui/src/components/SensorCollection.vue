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
                        <TimeseriesGraph
                          :rawData="data"
                          emptyLabel="No data available for this sensor"
                          yAxisUnit="mm"
                        />
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { CCard, CCardBody, CCardTitle } from '@coreui/vue'
import { Suspense } from 'vue'
import AsyncWrapper from './AsyncWrapper.vue'
import { computed, ref } from 'vue'

import TimeseriesGraph from './TimeseriesGraph.vue'
import axios from 'axios'
import type { RawData } from '@/models/Data'
import type { Sensor } from '@/models/Sensor'
import { useSensorStore } from '@/stores/sensor'

const BASE_URL: string = import.meta.env.VITE_BASE_URL
const sensorCollection = useSensorStore()

const sensors = computed<Sensor[]>(() => sensorCollection.sensors)

const pullSensorsRawDataFn = async (sensor: Sensor): Promise<RawData[]> => {
  try {
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
