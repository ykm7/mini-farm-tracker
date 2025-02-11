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
    <div>
      <!-- <h1>Weather Information</h1> -->
      <!-- <WeatherWidget :lat="'-33.6668553'" :long="'115.0810599'" :openWeatherApiKey="'<YOUR-OPEN-WEATHER-API-KEY>'" /> -->
      <!-- <vue-weather
        api-key="<your-open-weather-map-api-key>"
        units="uk"
        latitude="24.8864"
        longitude="91.8807"
      /> -->
    </div>
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
