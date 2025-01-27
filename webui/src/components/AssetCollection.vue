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
          <TimeseriesGraph :rawData="[]" />
        </div>
      </CCard>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Asset } from '@/models/Asset'
import { useAssetStore } from '@/stores/asset'
import TimeseriesGraph from './TimeseriesGraph.vue'

import { computed } from 'vue'
import {
  CCard,
  CCardBody,
  CCardTitle,
  CCardSubtitle,
  CListGroup,
  CListGroupItem,
} from '@coreui/vue'

const assetCollection = useAssetStore()

const assets = computed<Asset[]>(() => assetCollection.assets)
</script>
<style scoped></style>
