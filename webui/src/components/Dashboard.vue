<script setup lang="ts">
import { onMounted } from 'vue'

const assetCollection = useAssetStore()

onMounted(async () => {
  await assetCollection.fetchData()
})

import { useAssetStore } from '@/stores/asset'
</script>

<template>
  <main>
    <slot>
      <router-view></router-view>
    </slot>
  </main>
</template>

<style scoped>
</style>
