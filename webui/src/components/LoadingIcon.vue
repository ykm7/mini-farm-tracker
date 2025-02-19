<template>
  <div class="loading-container">
    <img src="/crop-growth.svg" alt="Loading" class="loading-icon" />
  </div>
</template>

<script lang="ts">
  import { defineComponent } from "vue"

  export default defineComponent({
    name: "LoadingIcon",
    props: {},
  })
</script>

<style scoped>
  .loading-container {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100vh;
    width: 100vw;
  }

  .loading-icon {
    width: 48px; /* Increased size for visibility */
    height: 48px;
    animation: grow 1.5s ease-in-out infinite;
    display: block; /* Ensure it's not hidden */
    margin: auto; /* Center it */
    color: #000000; /* Adjust color for visibility */
  }

  @keyframes grow {
    0%,
    100% {
      transform: scale(1);
    }
    50% {
      transform: scale(1.2);
    }
  }
</style>
