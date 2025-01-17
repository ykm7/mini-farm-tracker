<script setup lang="ts">
import axios from 'axios'
import { ref } from 'vue';

const BASE_URL: string = import.meta.env.VITE_BASE_URL

const message = ref('');

const pingServerFn = async () => {
  console.log('Attempting to ping server')
  try {
    const response = await axios.get(`${BASE_URL}/ping`);
    console.log(response)
    message.value = 'success'
  } catch (e) {
    console.warn(e)
    message.value = 'failure'
  }
}
</script>

<template>
  <div>
      <a>Basic PING test to the server.</a>
    <button class="button" @click="pingServerFn" title="ping server">
      ping server
    </button>
    </div>
      <a>Able to connect to server: {{ message }}</a>
</template>
