<script setup lang="ts" generic="T">
import { ref, watchEffect } from 'vue'

const props = defineProps<{
  promise: Promise<T>
}>()

const data = ref<T | null>(null)
const error = ref<Error | null>(null)
const loading = ref(true)

watchEffect(async () => {
  loading.value = true
  try {
    data.value = await props.promise
  } catch (e) {
    error.value = e as Error
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <slot :data="data"></slot>
</template>
