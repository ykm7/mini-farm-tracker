import { defineStore } from 'pinia'
import type { ObjectId } from '@/types/ObjectId'
import axios from 'axios'
import type { Sensor } from '@/models/Sensor'

const BASE_URL: string = import.meta.env.VITE_BASE_URL

interface SensorState {
  sensors: Sensor[] // could be a map.
  loading: boolean
}

export const useSensorStore = defineStore('sensor', {
  state: (): SensorState => ({
    sensors: [],
    loading: false,
  }),

  getters: {
    totalsensors: (state): number => {
      return state.sensors.length
    },
    getsensors: (state) => async (): Promise<Sensor[]> => {
      return state.sensors
    },
    getSensorById: (state) => (sensor: ObjectId) => {
      return state.sensors.find((a) => a.Id.toString() === sensor.toString())
    },
  },

  actions: {
    async fetchData() {
      try {
        const response = await axios.get<Sensor[]>(`${BASE_URL}/api/sensors`)
        this.sensors = response.data
      } catch (e) {
        console.log("🚀 ~ fetchData ~ e:", e)
        this.sensors = []
      }
    },
    /**
     * Will add if it doesn't exist, otherwise will update
     * @param sensor
     */
    addSensor(sensor: Sensor) {
      const foundIdx = this.sensors.findIndex((a) => a.Id === sensor.Id)
      if (foundIdx == -1) {
        this.sensors[foundIdx] = sensor
      } else {
        this.sensors.push(sensor)
      }
    },
    removeSensor(id: ObjectId) {
      const index = this.sensors.findIndex((sensor) => sensor.Id.toString() === id.toString())
      if (index !== -1) {
        this.sensors.splice(index, 1)
      }
    },
    setLoading(status: boolean) {
      this.loading = status
    },
  },
})
