<template>
  <div>
    <h4>Asset Collection</h4>
    <div>
      {{ assets }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAssetStore } from '@/stores/asset'
import { computed } from 'vue'
// import { storeToRefs } from 'pinia'

const assetCollection = useAssetStore()

const assets = computed(() => assetCollection.assets)
// const { assets } = storeToRefs(assetCollection)
</script>
<style scoped></style>
