<template>
  <div>
    <h4>Asset Collection</h4>
    <div v-for="asset in assets">
      <CCard class="card-holder" style="margin: 0.5rem 0">
        <div class="card-details">
          <CCardTitle>{{ asset.Name }}</CCardTitle>
          <CCardBody>
            <div>{{ asset.Description }}</div>
            <div>
              <a>Metrics:</a>
              <CListGroup flush v-for="metric in asset.Metrics">
                <CListGroupItem v-if="metric?.Volume"
                  ><label>Volume:</label> {{ metric.Volume }} litres</CListGroupItem
                >
              </CListGroup>
            </div>

            <div>
              <a>Attached Sensors:</a>
              <CListGroup flush v-for="sensor in asset.Sensors">
                <CListGroupItem>{{ sensor }}</CListGroupItem>
              </CListGroup>
            </div>
          </CCardBody>
        </div>

        <div class="card-graph">
          <div class="group-section">
            <AsyncWrapper :promise="assetToData.get(asset.Id)!">
              <template v-slot="{ data }">
                <div v-if="data">
                  <TimeseriesGraph
                    :item="asset"
                    @update-starting-date="handleUpdateStartingTimeEvent"
                    :displayData="data"
                    emptyLabel="No calibrated data available for this asset"
                  />
                </div>
              </template>
            </AsyncWrapper>
          </div>
        </div>
      </CCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Asset } from '@/models/Asset'
import { useAssetStore } from '@/stores/asset'
import TimeseriesGraph from './TimeseriesGraph.vue'
import AsyncWrapper from './AsyncWrapper.vue'
import { computed, ref, watch } from 'vue'
import { CCard, CCardBody, CCardTitle, CListGroup, CListGroupItem } from '@coreui/vue'
import axios from 'axios'
import type { CalibratedData } from '@/models/Data'
import type { DisplayPoint, GraphData, Unit } from '@/types/GraphRelated'
import type { ObjectId } from '@/types/ObjectId'

const BASE_URL: string = import.meta.env.VITE_BASE_URL
const assetCollection = useAssetStore()

const assets = computed<Asset[]>(() => assetCollection.assets)
const assetIdToStarting = ref<Map<ObjectId, number>>(new Map())
const assetToData = ref<Map<ObjectId, Promise<GraphData>>>(new Map())

function handleUpdateStartingTimeEvent(asset: Asset, startingOffset: number) {
  const newMap = new Map(assetIdToStarting.value)
  newMap.set(asset.Id, startingOffset)
  assetIdToStarting.value = newMap
}

watch(
  assets,
  (newAssets, _) => {
    newAssets.forEach((a) => {
      assetToData.value.set(a.Id, Promise.resolve({}))
    })
  },
  { immediate: true },
)

// const firstMapSet = ref<boolean>(true)
watch(
  assetIdToStarting,
  (newMap, oldMap) => {
    // Iterate through the new map to find changes
    newMap.forEach((newValue, key) => {
      const oldValue = oldMap.get(key)
      // TODO: Have to re-visit this otherwise each time ANY time trigger is updated ALL arrays are updated
      if (!oldValue || newValue !== oldValue) {
        assetToData.value.set(
          key,
          pullCalibratedDataFn(
            assets.value.find((a) => a.Id.toString() === key.toString())!,
            newValue,
          ),
        )
        // firstMapSet.value = false
      }
    })
  },
  { deep: true },
)

const pullCalibratedDataFn = async (
  asset: Asset,
  startOffset: number,
  endOffset: number = 0,
): Promise<GraphData> => {
  const now = new Date()
  const start = new Date(now.getTime() - startOffset)
  const end = new Date(now.getTime() - endOffset)

  const params = new URLSearchParams({
    start: start.toISOString(),
    end: end.toISOString(),
  })

  const graphData: GraphData = {}

  if (!(asset.Sensors && asset.Sensors?.length > 0)) {
    return graphData
  }
  try {
    // TODO: Handle multiple sensors on a asset
    const response = await axios.get<CalibratedData[]>(
      `${BASE_URL}/api/sensors/${asset.Sensors[0]}/data/calibrated_data?${params.toString()}`,
    )

    response.data.forEach((d: CalibratedData) => {
      if (d.DataPoints.Volume) {
        if (graphData.Volume == null) {
          graphData.Volume = {
            data: [],
            unit: d.DataPoints.Volume.Unit as Unit,
          }
        }

        graphData.Volume!.data.push({
          value: d.DataPoints.Volume.Data,
          timestamp: d.Timestamp,
        })
      }

      if (d.DataPoints.AirTemperature) {
        if (graphData.AirTemperature == null) {
          graphData.AirTemperature = {
            data: [],
            unit: d.DataPoints.AirTemperature.Unit as Unit,
          }
        }

        graphData.AirTemperature.data.push({
          value: d.DataPoints.AirTemperature.Data,
          timestamp: d.Timestamp,
        })
      }
    })
    console.log('🚀 ~ graphData:', graphData)
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
</style>
