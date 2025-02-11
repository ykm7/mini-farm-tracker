<script setup lang="ts">
  import DashboardHeader from "@/components/DashboardHeader.vue" /* PartiallyEnd: #3632/scriptSetup.vue */
  import DashboardMain from "@/components/DashboardMain.vue"
</script>

<template>
  <main class="main-content">
    <div class="dashboard-header">
      <DashboardHeader />
      <!-- <div class="dashboard-content"></div> -->
    </div>
    <div class="dashboard">
      <DashboardMain />
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
