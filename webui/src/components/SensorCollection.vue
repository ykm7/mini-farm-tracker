<template>
  <div>
    <h4>Sensor Collection</h4>
    <div>
      <a>Available sensors</a>

      <div v-for="sensor in sensors">
        <CCard class="card-holder" style="margin: 0.5rem 0">
          <div class="card-details">
            <CCardTitle>{{ sensor.Id }}</CCardTitle>
            <!-- <CCardSubtitle class="mb-2 text-body-secondary">{{ asset.Id }}</CCardSubtitle> -->
            <CCardBody>{{ sensor.Description }}</CCardBody>
          </div>
          <div class="card-graph">
            <div class="group-section">
              <Suspense>
                <template #default>
                  <div>
                    <AsyncWrapper :promise="pullSensorsRawDataFn(sensor)">
                      <template v-slot="{ data }">
                        <div v-if="data">
                          <TimeseriesGraph
                            :displayData="data"
                            emptyLabel="No data available for this sensor"
                            yAxisUnit="mm"
                            lineLabel="Distance"
                            title="Distance measured by sensor"
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
          </div>
        </CCard>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { CCard, CCardBody, CCardTitle } from '@coreui/vue'
import AsyncWrapper from './AsyncWrapper.vue'
import { computed, ref } from 'vue'

import TimeseriesGraph, { type DisplayPoint } from './TimeseriesGraph.vue'
import axios from 'axios'
import type { RawData } from '@/models/Data'
import type { Sensor } from '@/models/Sensor'
import { useSensorStore } from '@/stores/sensor'

const BASE_URL: string = import.meta.env.VITE_BASE_URL
const sensorCollection = useSensorStore()

const sensors = computed<Sensor[]>(() => sensorCollection.sensors)

const pullSensorsRawDataFn = async (sensor: Sensor): Promise<DisplayPoint[]> => {
  try {
    const response = await axios.get<RawData[]>(
      `${BASE_URL}/api/sensors/${sensor.Id}/data/raw_data`,
    )

    const convertedData: DisplayPoint[] = response.data
      .filter((d: RawData) => {
        return d?.Valid === false
      })
      .map<DisplayPoint>((d: RawData) => {
        return {
          timestamp: d.Timestamp,
          value: d.Data.Distance.split(' ')[0] as unknown as number,
        }
      })

    return convertedData
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
