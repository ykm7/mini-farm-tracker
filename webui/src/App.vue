<template>
  <Suspense @pending="onPending" @resolve="onResolve">
    <template #default>
      <AsyncApp />
    </template>
    <template #fallback>
      <LoadingIcon v-if="showLoading" />
    </template>
  </Suspense>
</template>

<script setup lang="ts">
  import LoadingIcon from "@/components/LoadingIcon.vue"
  import { defineAsyncComponent, onMounted, ref } from "vue"

  const AsyncApp = defineAsyncComponent(() => import("./AsyncApp.vue"))

  const showLoading = ref(false)
  let loadingTimeout: number | null = null

  const onPending = () => {
    loadingTimeout = setTimeout(() => {
      showLoading.value = true
      // diplay displaying a loading indicator. If we don't we "flicker" the loading icon which is poor UX. 
    }, 200)
  }

  const onResolve = () => {
    if (loadingTimeout) {
      clearTimeout(loadingTimeout)
    }
    showLoading.value = false
  }

  onMounted(() => {
    // Clean up the timeout if the component is unmounted
    return () => {
      if (loadingTimeout) {
        clearTimeout(loadingTimeout)
      }
    }
  })
</script>
