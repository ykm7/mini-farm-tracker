<template>
  <header class="dashboard-header">
    <nav class="navigation">
      <ul>
        <li>
          <router-link to="/">
            <font-awesome-icon :icon="['fas', 'house']" />
          </router-link>
        </li>
        <li><router-link to="/">Assets</router-link></li>
        <li><router-link to="/sensor">Sensors</router-link></li>
      </ul>
    </nav>
    <div class="ping-test">
      <button class="button" @click="pingServerFn" title="ping server">ping server</button>
      <a>{{ pingMsg }}</a>
    </div>
    <div class="user-info">
      <span class="username">John Doe</span>
    </div>
  </header>
</template>

<script setup lang="ts">
import axios from 'axios'
import { ref } from 'vue'

const BASE_URL: string = import.meta.env.VITE_BASE_URL
const pingMsg = ref('')

const pingServerFn = async () => {
  console.log('Attempting to ping server')
  try {
    await axios.get(`${BASE_URL}/ping`)
    pingMsg.value = 'success'
  } catch (e) {
    console.warn(e)
    pingMsg.value = 'failure'
  }
}
</script>

<style scoped>
.dashboard-header {
  display: flex;
  justify-content: space-between;
}

.navigation {
  ul {
    display: flex;
    list-style-type: none;
    margin: 0;
    padding: 0;
  }
  li {
    margin-right: 1.5rem;
  }
}
</style>
