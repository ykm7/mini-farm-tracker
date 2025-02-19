<template>
  <Suspense>
    <template #default>
      <AsyncApp />
    </template>
    <template #fallback>
      <LoadingIcon />
    </template>
  </Suspense>
</template>

<script setup lang="ts">
import { defineAsyncComponent } from 'vue'
import LoadingIcon from "@/components/LoadingIcon.vue"

const AsyncApp = defineAsyncComponent(() => import('./AsyncApp.vue'))
</script>
