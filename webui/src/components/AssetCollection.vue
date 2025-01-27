<template>
  <div>
    <h4>Asset Collection</h4>
    <div v-for="asset in assets">
      <CCard style="margin: 0.5rem 0">
        <CCardTitle>{{ asset.Name }}</CCardTitle>
        <!-- <CCardSubtitle class="mb-2 text-body-secondary">{{ asset.Id }}</CCardSubtitle> -->
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

        <div class="group-section">
          <Suspense>
            <template #default>
              <div>
                <AsyncWrapper :promise="pullCalibratedDataFn(asset)">
                  <template v-slot="{ data }">
                    <div v-if="data">
                      <!-- {{ data[0] }} -->
                      <TimeseriesGraph
                        :rawData="data"
                        emptyLabel="No calibrated data available for this asset"
                        yAxisUnit="L"
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
</template>

<script setup lang="ts">
import type { Asset } from '@/models/Asset'
import { useAssetStore } from '@/stores/asset'
import TimeseriesGraph from './TimeseriesGraph.vue'
import { Suspense } from 'vue'
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

const pullCalibratedDataFn = async (asset: Asset): Promise<CalibratedData[]> => {
  if (!(asset.Sensors && asset.Sensors?.length > 0)) {
    return []
  }
  try {
    const response = await axios.get<CalibratedData[]>(
      `${BASE_URL}/api/sensors/${asset.Sensors[0]}/data/calibrated_data`,
    )
    return response.data ? response.data : []
  } catch (e) {
    console.warn(e)
    return []
  }
}
</script>
<style scoped></style>
