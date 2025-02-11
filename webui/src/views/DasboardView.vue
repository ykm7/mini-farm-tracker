<script setup lang="ts">
import Dashboard from '../components/DashboardMain.vue'
import DashboardHeader from '../components/DashboardHeader.vue'
</script>

<template>
  <main class="main-content">
    <div class="dashboard-header">
      <DashboardHeader />
      <!-- <div class="dashboard-content"></div> -->
    </div>
    <div class="dashboard">
      <Dashboard />
    </div>
  </main>
</template>

<style scoped>
.main-content {
  padding: 2rem;
  min-height: 100vh;
  min-width: 100vw;
  display: flex;
  flex-direction: column;
}

.dashboard-header {
  height: 50px;
  flex-shrink: 0;
}

.dashboard {
  flex-grow: 1;
}
</style>