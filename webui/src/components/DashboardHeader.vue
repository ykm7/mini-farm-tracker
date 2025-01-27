<script setup lang="ts"></script>

<template>
  <header class="dashboard-header">
    <nav class="navigation">
      <ul>
        <li><a href="#">Assets</a></li>
        <li><a href="#">Sensors</a></li>
      </ul>
    </nav>
    <div class="user-info">
      <span class="username">John Doe</span>
    </div>
  </header>
</template>

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
