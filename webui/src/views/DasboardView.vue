<script setup lang="ts">
import Dashboard from '../components/Dashboard.vue'
import DashboardHeader from '../components/DashboardHeader.vue'
</script>

<template>
  <main>
    <div class="dashboard">
      <DashboardHeader />
      <div class="dashboard-content"></div>
    </div>
    <Dashboard />

    <router-view></router-view>
  </main>
</template>
