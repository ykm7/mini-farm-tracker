<script setup lang="ts">
import { useAssetStore } from '@/stores/asset'
import { useSensorStore } from '@/stores/sensor'
import { onMounted } from 'vue'

const assetCollection = useAssetStore()
const sensorCollection = useSensorStore()

onMounted(async () => {
  const result = await Promise.allSettled([
    assetCollection.fetchData(),
    sensorCollection.fetchData(),
  ])
  result.forEach((r) => {
    if (r.status == 'rejected') {
      console.warn(r.reason)
    }
  })
})
</script>

<template>
  <main>
    <slot>
      <router-view></router-view>
    </slot>
  </main>
</template>

<style scoped></style>
