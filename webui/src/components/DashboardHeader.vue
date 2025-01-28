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
    <div class="user-info">
      <span class="username">Default User (Reader)</span>
    </div>
  </header>
</template>

<script setup lang="ts"></script>

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
