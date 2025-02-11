<template>
  <div>
    <h4>Asset Collection</h4>
    <div :key="asset.Id.toString()" v-for="asset in assets">
      <CCard class="card-holder" style="margin: 0.5rem 0">
        <div class="card-details">
          <CCardTitle>{{ asset.Name }}</CCardTitle>
          <CCardBody>
            <div>{{ asset.Description }}</div>
            <div>
              <a>Metrics:</a>
              <CListGroup
                flush
                :key="`${metric?.Height}-${metric?.Radius}-${metric?.Volume}`"
                v-for="metric in asset.Metrics"
              >
                <CListGroupItem v-if="metric?.Volume"
                  ><label>Volume:</label> {{ metric.Volume }} litres</CListGroupItem
                >
              </CListGroup>
            </div>

            <div>
              <a>Attached Sensors:</a>
              <CListGroup flush :key="sensor" v-for="sensor in asset.Sensors">
                <CListGroupItem>{{ sensor }}</CListGroupItem>
              </CListGroup>
            </div>
          </CCardBody>
        </div>

        <div class="card-graph">
          <div class="group-section">
            <AsyncWrapper :promise="assetToData.get(asset.Id)!">
              <template v-slot="{ data }">
                <div v-if="data">
                  <TimeseriesGraph
                    :item="asset"
                    @update-starting-date="handleUpdateStartingTimeEvent"
                    :displayData="data"
                    emptyLabel="No calibrated data available for this asset"
                  />
                </div>
              </template>
            </AsyncWrapper>
          </div>
        </div>
      </CCard>
    </div>
  </div>
</template>

<script setup lang="ts">
  import type { Asset } from "@/models/Asset"
  import type { CalibratedData } from "@/models/Data"
  import { useAssetStore } from "@/stores/asset"
  import type { GraphData, Unit } from "@/types/GraphRelated"
  import type { ObjectId } from "@/types/ObjectId"
  import { CCard, CCardBody, CCardTitle, CListGroup, CListGroupItem } from "@coreui/vue"
  import axios from "axios"
  import { computed, ref, watch } from "vue"
  import AsyncWrapper from "./AsyncWrapper.vue"
  import TimeseriesGraph from "./TimeseriesGraph.vue"

  const BASE_URL: string = import.meta.env.VITE_BASE_URL
  const assetCollection = useAssetStore()

  const assets = computed<Asset[]>(() => assetCollection.assets)
  const assetIdToStarting = ref<Map<ObjectId, number>>(new Map())
  const assetToData = ref<Map<ObjectId, Promise<GraphData>>>(new Map())

  function handleUpdateStartingTimeEvent(asset: Asset, startingOffset: number) {
    const newMap = new Map(assetIdToStarting.value)
    newMap.set(asset.Id, startingOffset)
    assetIdToStarting.value = newMap
  }

  watch(
    assets,
    (newAssets) => {
      newAssets.forEach((a) => {
        assetToData.value.set(a.Id, Promise.resolve({}))
      })
    },
    { immediate: true }
  )

  watch(
    assetIdToStarting,
    (newMap, oldMap) => {
      // Iterate through the new map to find changes
      newMap.forEach((newValue, key) => {
        const oldValue = oldMap.get(key)
        // TODO: Have to re-visit this otherwise each time ANY time trigger is updated ALL arrays are updated
        if (!oldValue || newValue !== oldValue) {
          assetToData.value.set(
            key,
            pullCalibratedDataFn(
              assets.value.find((a) => a.Id.toString() === key.toString())!,
              newValue
            )
          )
        }
      })
    },
    { deep: true }
  )

  const pullCalibratedDataFn = async (
    asset: Asset,
    startOffset: number,
    endOffset: number = 0
  ): Promise<GraphData> => {
    const now = new Date()
    const start = new Date(now.getTime() - startOffset)
    const end = new Date(now.getTime() - endOffset)

    // const params = new URLSearchParams({
    //   start: start.toISOString(),
    //   end: end.toISOString(),
    // })

    const graphData: GraphData = {}

    if (!(asset.Sensors && asset.Sensors?.length > 0)) {
      return graphData
    }

    const data: CalibratedData[] = []
    const params = new URLSearchParams({
      start: start.toISOString(),
      end: end.toISOString(),
    })

    while (true) {
      try {
        // TODO: This still is limiting a single sensor per asset.
        const response = await axios.get<CalibratedData[]>(
          `${BASE_URL}/api/sensors/${asset.Sensors[0]}/data/calibrated_data?${params.toString()}`
        )

        // efficient array adding
        Array.prototype.push.apply(data, response.data)

        const limitHeader = Number(response.headers["x-max-data-limit"])
        if (Number.isNaN(limitHeader)) {
          console.error("Unable to find the expected header 'x-max-data-limit'")
          break
        }

        const available = response.data.length

        if (available < limitHeader) {
          // We have the all within the specified limit limit
          break
        }

        params.set("start", response.data[available - 1].Timestamp)
      } catch (e) {
        console.warn(e)
        return {}
      }
    }

    data.forEach((d: CalibratedData) => {
      if (d.DataPoints.Volume) {
        if (graphData.Volume == null) {
          graphData.Volume = {
            data: [],
            unit: d.DataPoints.Volume.Units as Unit,
          }
        }

        graphData.Volume!.data.push({
          value: d.DataPoints.Volume.Data,
          timestamp: d.Timestamp,
        })
      }

      if (d.DataPoints.AirTemperature) {
        if (graphData.AirTemperature == null) {
          graphData.AirTemperature = {
            data: [],
            unit: d.DataPoints.AirTemperature.Units as Unit,
          }
        }

        graphData.AirTemperature.data.push({
          value: d.DataPoints.AirTemperature.Data,
          timestamp: d.Timestamp,
        })
      }

      if (d.DataPoints.AirHumidity) {
        if (graphData.AirHumidity == null) {
          graphData.AirHumidity = {
            data: [],
            unit: d.DataPoints.AirHumidity.Units as Unit,
          }
        }

        graphData.AirHumidity.data.push({
          value: d.DataPoints.AirHumidity.Data,
          timestamp: d.Timestamp,
        })
      }

      if (d.DataPoints.LightIntensity) {
        if (graphData.LightIntensity == null) {
          graphData.LightIntensity = {
            data: [],
            unit: d.DataPoints.LightIntensity.Units as Unit,
          }
        }

        graphData.LightIntensity.data.push({
          value: d.DataPoints.LightIntensity.Data,
          timestamp: d.Timestamp,
        })
      }

      if (d.DataPoints.UvIndex) {
        if (graphData.UvIndex == null) {
          graphData.UvIndex = {
            data: [],
            unit: d.DataPoints.UvIndex.Units as Unit,
          }
        }

        graphData.UvIndex.data.push({
          value: d.DataPoints.UvIndex.Data,
          timestamp: d.Timestamp,
        })
      }

      if (d.DataPoints.WindSpeed) {
        if (graphData.WindSpeed == null) {
          graphData.WindSpeed = {
            data: [],
            unit: d.DataPoints.WindSpeed.Units as Unit,
          }
        }

        graphData.WindSpeed.data.push({
          value: d.DataPoints.WindSpeed.Data,
          timestamp: d.Timestamp,
        })
      }

      if (d.DataPoints.WindDirection) {
        if (graphData.WindDirection == null) {
          graphData.WindDirection = {
            data: [],
            unit: d.DataPoints.WindDirection.Units as Unit,
          }
        }

        graphData.WindDirection.data.push({
          value: d.DataPoints.WindDirection.Data,
          timestamp: d.Timestamp,
        })
      }

      if (d.DataPoints.RainfallHourly) {
        if (graphData.RainfallHourly == null) {
          graphData.RainfallHourly = {
            data: [],
            unit: d.DataPoints.RainfallHourly.Units as Unit,
          }
        }

        graphData.RainfallHourly.data.push({
          value: d.DataPoints.RainfallHourly.Data,
          timestamp: d.Timestamp,
        })
      }

      if (d.DataPoints.BarometricPressure) {
        if (graphData.BarometricPressure == null) {
          graphData.BarometricPressure = {
            data: [],
            unit: d.DataPoints.BarometricPressure.Units as Unit,
          }
        }

        graphData.BarometricPressure.data.push({
          value: d.DataPoints.BarometricPressure.Data,
          timestamp: d.Timestamp,
        })
      }
    })
    return graphData
  }
</script>
<style scoped>
  .card-title {
    text-align: center;
    margin-top: var(--cui-card-spacer-y);
  }
</style>
