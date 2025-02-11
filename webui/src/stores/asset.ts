import type { Asset } from "@/models/Asset"
import { defineStore } from "pinia"
import type { ObjectId } from "@/types/ObjectId"
import axios from "axios"

const BASE_URL: string = import.meta.env.VITE_BASE_URL

interface AssetState {
  assets: Asset[] // could be a map.
  loading: boolean
}

export const useAssetStore = defineStore("asset", {
  state: (): AssetState => ({
    assets: [],
    loading: false,
  }),

  getters: {
    totalAssets: (state): number => {
      return state.assets.length
    },
    getAssets: (state) => async (): Promise<Asset[]> => {
      return state.assets
    },
    getAssetById: (state) => (assetId: ObjectId) => {
      return state.assets.find((a) => a.Id.toString() === assetId.toString())
    },
  },

  actions: {
    async fetchData() {
      try {
        const response = await axios.get<Asset[]>(`${BASE_URL}/api/assets`)
        this.assets = response.data
      } catch (e) {
        console.log("🚀 ~ fetchData ~ e:", e)
        this.assets = []
      }
    },
    /**
     * Will add if it doesn't exist, otherwise will update
     * @param asset
     */
    addAsset(asset: Asset) {
      const foundIdx = this.assets.findIndex((a) => a.Id === asset.Id)
      if (foundIdx == -1) {
        this.assets[foundIdx] = asset
      } else {
        this.assets.push(asset)
      }
    },
    removeAsset(id: ObjectId) {
      const index = this.assets.findIndex((asset) => asset.Id.toString() === id.toString())
      if (index !== -1) {
        this.assets.splice(index, 1)
      }
    },
    setLoading(status: boolean) {
      this.loading = status
    },
  },
})
