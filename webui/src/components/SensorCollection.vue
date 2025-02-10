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
            <CCardBody
              ><div>{{ sensor.Description }}</div></CCardBody
            >
          </div>
          <div class="card-graph">
            <div class="group-section">
              <AsyncWrapper :promise="sensorToData.get(sensor.Id)!">
                <template v-slot="{ data }">
                  <TimeseriesGraph
                    :item="sensor"
                    @update-starting-date="handleUpdateStartingTimeEvent"
                    :displayData="data ? data : {}"
                  />
                </template>
              </AsyncWrapper>
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
import { computed, ref, watch } from 'vue'

import TimeseriesGraph from './TimeseriesGraph.vue'
import axios from 'axios'
import type { RawData } from '@/models/Data'
import type { Sensor } from '@/models/Sensor'
import { useSensorStore } from '@/stores/sensor'
import type { GraphData } from '@/types/GraphRelated'

const BASE_URL: string = import.meta.env.VITE_BASE_URL
const sensorCollection = useSensorStore()
const sensors = computed<Sensor[]>(() => sensorCollection.sensors)
const sensorIdToStarting = ref<Map<string, number>>(new Map())
const sensorToData = ref<Map<string, Promise<GraphData>>>(new Map())

function handleUpdateStartingTimeEvent(sensor: Sensor, startingOffset: number) {
  const newMap = new Map(sensorIdToStarting.value)
  newMap.set(sensor.Id, startingOffset)
  sensorIdToStarting.value = newMap
}

watch(
  sensors,
  (newSensors, _) => {
    newSensors.forEach((s) => {
      sensorToData.value.set(s.Id, Promise.resolve({}))
    })
  },
  { immediate: true },
)

// const firstMapSet = ref<boolean>(true)
watch(
  sensorIdToStarting,
  (newMap, oldMap) => {
    // Iterate through the new map to find changes
    newMap.forEach((newValue, key) => {
      const oldValue = oldMap.get(key)
      // TODO: Have to re-visit this otherwise each time ANY time trigger is updated ALL arrays are updated
      if (!oldValue || newValue !== oldValue) {
        sensorToData.value.set(
          key,
          pullSensorData(sensors.value.find((s) => s.Id === key)!, newValue),
        )

        // firstMapSet.value = false
      }
      // }
    })
  },
  { deep: true },
)

const pullSensorData = async (
  sensor: Sensor,
  startOffset: number,
  endOffset: number = 0,
): Promise<GraphData> => {
  const now = new Date()
  const start = new Date(now.getTime() - startOffset)
  const end = new Date(now.getTime() - endOffset)

  const graphData: GraphData = {}

  const params = new URLSearchParams({
    start: start.toISOString(),
    end: end.toISOString(),
  })

  try {
    const response = await axios.get<RawData[]>(
      `${BASE_URL}/api/sensors/${sensor.Id}/data/raw_data?${params.toString()}`,
    )

    response.data.forEach((d) => {
      if (d.Data.LDDS45) {
        if (graphData.Raw == null) {
          graphData.Raw = {
            unit: 'mm',
            data: [],
          }
        }

        graphData.Raw?.data.push({
          value: d.Data.LDDS45.Distance.split(' ')[0] as unknown as number,
          timestamp: d.Timestamp,
        })
      }
    })

    return graphData
  } catch (e) {
    console.warn(e)
    return {}
  }
}
</script>

<style scoped>
.card-title {
  text-align: center;
  margin-top: var(--cui-card-spacer-y);
}

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
