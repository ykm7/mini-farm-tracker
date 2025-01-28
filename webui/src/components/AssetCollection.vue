<template>
  <div>
    <h4>Asset Collection</h4>
    <div v-for="asset in assets">
      <CCard class="card-holder" style="margin: 0.5rem 0">
        <div class="card-details">
          <CCardTitle>{{ asset.Name }}</CCardTitle>
          <CCardBody>{{ asset.Description }}</CCardBody>

          <div>
            <a>Metrics:</a>
            <CListGroup flush v-for="metric in asset.Metrics">
              <CListGroupItem><label>Volume:</label> {{ metric?.Volume }} litres</CListGroupItem>
            </CListGroup>
          </div>

          <div>
            <a>Attached Sensors:</a>
            <CListGroup flush v-for="sensor in asset.Sensors">
              <CListGroupItem>{{ sensor }}</CListGroupItem>
            </CListGroup>
          </div>
        </div>
        <div class="card-graph">
          <div class="group-section">
            <Suspense>
              <template #default>
                <div>
                  <AsyncWrapper :promise="pullCalibratedDataFn(asset)">
                    <template v-slot="{ data }">
                      <div v-if="data">
                        <TimeseriesGraph
                          :displayData="data"
                          emptyLabel="No calibrated data available for this asset"
                          yAxisUnit="L"
                          lineLabel="Litres"
                          title="Water in tank"
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
</template>

<script setup lang="ts">
import type { Asset } from '@/models/Asset'
import { useAssetStore } from '@/stores/asset'
import TimeseriesGraph, { type DisplayPoint } from './TimeseriesGraph.vue'
import AsyncWrapper from './AsyncWrapper.vue'
import { computed } from 'vue'
import {
  CCard,
  CCardBody,
  CCardTitle,
  CCardSubtitle,
  CListGroup,
  CListGroupItem,
} from '@coreui/vue'
import axios from 'axios'
import type { CalibratedData } from '@/models/Data'

const BASE_URL: string = import.meta.env.VITE_BASE_URL
const assetCollection = useAssetStore()

const assets = computed<Asset[]>(() => assetCollection.assets)

const pullCalibratedDataFn = async (asset: Asset): Promise<DisplayPoint[]> => {
  if (!(asset.Sensors && asset.Sensors?.length > 0)) {
    return []
  }
  try {
    const response = await axios.get<CalibratedData[]>(
      `${BASE_URL}/api/sensors/${asset.Sensors[0]}/data/calibrated_data`,
    )
    
    const convertedData: DisplayPoint[] = response.data
    .map<DisplayPoint>((c: CalibratedData) => {
      return {
        timestamp: c.Timestamp,
        value: c.Data
      }
    })

    return convertedData
  } catch (e) {
    console.warn(e)
    return []
  }
}
</script>
<style scoped></style>
