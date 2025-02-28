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
          <div
            v-if="asset.Sensors && sensorToAggregationData.get(asset.Sensors[0])"
            class="group-section"
          >
            <HistoricDataGraph :data="sensorToAggregationData.get(asset.Sensors[0])!" />
          </div>
        </div>
      </CCard>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { customMerge } from "@/helper"
  import type { Asset } from "@/models/Asset"
  import {
    AGGREGATION_TYPE,
    CalibratedDataNames,
    type AggregationData,
    type CalibratedData,
  } from "@/models/Data"
  import { useAssetStore } from "@/stores/asset"
  import type {
    AggregatedDataMapping,
    AggregatedDataPoint,
    CalibratedDataNamesGrouping,
    GraphData,
    Unit,
  } from "@/types/GraphRelated"
  import type { ObjectId } from "@/types/ObjectId"
  import { CCard, CCardBody, CCardTitle, CListGroup, CListGroupItem } from "@coreui/vue"
  import axios, { type CancelTokenSource } from "axios"

  import mergeWith from "lodash/mergeWith"
  import { computed, onMounted, ref, watch } from "vue"
  import AsyncWrapper from "./AsyncWrapper.vue"
  import HistoricDataGraph from "./HistoricDataGraph.vue"
  import TimeseriesGraph from "./TimeseriesGraph.vue"

  const BASE_URL: string = import.meta.env.VITE_BASE_URL
  const assetCollection = useAssetStore()

  const assets = computed<Asset[]>(() => assetCollection.assets)
  const assetIdToStarting = ref<Map<ObjectId, number>>(new Map())
  const assetToData = ref<Map<ObjectId, Promise<GraphData>>>(new Map())

  const sensorToAggregationData = ref<Map<string, Partial<CalibratedDataNamesGrouping>>>(new Map())
  // sensor id -> cancellation tokens for all network calls (for the sensor)
  const cancelTokens: Map<ObjectId, CancelTokenSource[]> = new Map()

  function handleUpdateStartingTimeEvent(asset: Asset, startingOffset: number) {
    const newMap = new Map(assetIdToStarting.value)
    newMap.set(asset.Id, startingOffset)
    assetIdToStarting.value = newMap
  }

  watch(
    assets,
    (newAssets) => {
      newAssets.forEach((a) => {
        cancelTokens.set(a.Id, [])
        assetToData.value.set(a.Id, Promise.resolve({}))
      })
    },
    { immediate: true }
  )

  watch(
    assetIdToStarting,
    (newMap, oldMap) => {
      // Iterate through the new map to find changes
      newMap.forEach(async (newValue, key) => {
        const oldValue = oldMap.get(key)
        // TODO: Have to re-visit this otherwise each time ANY time trigger is updated ALL arrays are updated
        if (!oldValue || newValue !== oldValue) {
          const generator = pullCalibratedDataFn(
            assets.value.find((a) => a.Id.toString() === key.toString())!,
            newValue
          )

          for await (const data of generator) {
            assetToData.value.set(key, Promise.resolve(data))
          }
        }
      })
    },
    { deep: true }
  )

  const pullCalibratedDataFn = async function* (
    asset: Asset,
    startOffset: number,
    endOffset: number = 0
  ): AsyncGenerator<GraphData> {
    cancelTokens.get(asset.Id)!.forEach((source) => source.cancel("Request cancelled"))
    cancelTokens.set(asset.Id, [])

    const now = new Date()
    const start = new Date(now.getTime() - startOffset)
    const end = new Date(now.getTime() - endOffset)

    let graphData: GraphData = {}

    if (!(asset.Sensors && asset.Sensors?.length > 0)) {
      return graphData
    }

    const params = new URLSearchParams({
      start: start.toISOString(),
      end: end.toISOString(),
    })

    while (true) {
      const newGraphData: GraphData = {}

      const source = axios.CancelToken.source()
      cancelTokens.get(asset.Id)!.push(source)

      try {
        // TODO: This still is limiting a single sensor per asset.
        const response = await axios.get<CalibratedData[]>(
          `${BASE_URL}/api/sensors/${asset.Sensors[0]}/data/calibrated_data?${params.toString()}`,
          { cancelToken: source.token }
        )

        response.data.forEach((d: CalibratedData) => {
          if (d.DataPoints.Volume) {
            if (newGraphData.Volume == null) {
              newGraphData.Volume = {
                data: [],
                unit: d.DataPoints.Volume.Units as Unit,
              }
            }

            newGraphData.Volume!.data.push({
              value: d.DataPoints.Volume.Data,
              timestamp: d.Timestamp,
            })
          }

          if (d.DataPoints.AirTemperature) {
            if (newGraphData.AirTemperature == null) {
              newGraphData.AirTemperature = {
                data: [],
                unit: d.DataPoints.AirTemperature.Units as Unit,
              }
            }

            newGraphData.AirTemperature.data.push({
              value: d.DataPoints.AirTemperature.Data,
              timestamp: d.Timestamp,
            })
          }

          if (d.DataPoints.AirHumidity) {
            if (newGraphData.AirHumidity == null) {
              newGraphData.AirHumidity = {
                data: [],
                unit: d.DataPoints.AirHumidity.Units as Unit,
              }
            }

            newGraphData.AirHumidity.data.push({
              value: d.DataPoints.AirHumidity.Data,
              timestamp: d.Timestamp,
            })
          }

          if (d.DataPoints.LightIntensity) {
            if (newGraphData.LightIntensity == null) {
              newGraphData.LightIntensity = {
                data: [],
                unit: d.DataPoints.LightIntensity.Units as Unit,
              }
            }

            newGraphData.LightIntensity.data.push({
              value: d.DataPoints.LightIntensity.Data,
              timestamp: d.Timestamp,
            })
          }

          if (d.DataPoints.UvIndex) {
            if (newGraphData.UvIndex == null) {
              newGraphData.UvIndex = {
                data: [],
                unit: d.DataPoints.UvIndex.Units as Unit,
              }
            }

            newGraphData.UvIndex.data.push({
              value: d.DataPoints.UvIndex.Data,
              timestamp: d.Timestamp,
            })
          }

          if (d.DataPoints.WindSpeed) {
            if (newGraphData.WindSpeed == null) {
              newGraphData.WindSpeed = {
                data: [],
                unit: d.DataPoints.WindSpeed.Units as Unit,
              }
            }

            newGraphData.WindSpeed.data.push({
              value: d.DataPoints.WindSpeed.Data,
              timestamp: d.Timestamp,
            })
          }

          if (d.DataPoints.WindDirection) {
            if (newGraphData.WindDirection == null) {
              newGraphData.WindDirection = {
                data: [],
                unit: d.DataPoints.WindDirection.Units as Unit,
              }
            }

            newGraphData.WindDirection.data.push({
              value: d.DataPoints.WindDirection.Data,
              timestamp: d.Timestamp,
            })
          }

          if (d.DataPoints.RainGauge) {
            if (newGraphData.RainGauge == null) {
              newGraphData.RainGauge = {
                data: [],
                unit: d.DataPoints.RainGauge.Units as Unit,
              }
            }

            newGraphData.RainGauge.data.push({
              value: d.DataPoints.RainGauge.Data,
              timestamp: d.Timestamp,
            })
          }

          if (d.DataPoints.BarometricPressure) {
            if (newGraphData.BarometricPressure == null) {
              newGraphData.BarometricPressure = {
                data: [],
                unit: d.DataPoints.BarometricPressure.Units as Unit,
              }
            }

            newGraphData.BarometricPressure.data.push({
              value: d.DataPoints.BarometricPressure.Data,
              timestamp: d.Timestamp,
            })
          }

          if (d.DataPoints.PeakWindGust) {
            if (newGraphData.PeakWindGust == null) {
              newGraphData.PeakWindGust = {
                data: [],
                unit: d.DataPoints.PeakWindGust.Units as Unit,
              }
            }

            newGraphData.PeakWindGust.data.push({
              value: d.DataPoints.PeakWindGust.Data,
              timestamp: d.Timestamp,
            })
          }

          if (d.DataPoints.RainAccumulation) {
            if (newGraphData.RainAccumulation == null) {
              newGraphData.RainAccumulation = {
                data: [],
                unit: d.DataPoints.RainAccumulation.Units as Unit,
              }
            }

            newGraphData.RainAccumulation.data.push({
              value: d.DataPoints.RainAccumulation.Data,
              timestamp: d.Timestamp,
            })
          }
        })

        graphData = mergeWith({}, graphData, newGraphData, customMerge)
        yield graphData

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
        if (axios.isCancel(e)) {
          // pass
        } else {
          // Handle other errors
          console.warn(e)
        }

        break
      } finally {
        const index = cancelTokens.get(asset.Id)!.indexOf(source)
        if (index > -1) {
          cancelTokens.get(asset.Id)!.splice(index, 1)
        }
      }
    }
  }

  /**
   * TODO: Give actually better name
   */

  /**
   * TODO:
   * Add sensor
   * Add data type
   *
   * Currently with the limited data not really required.
   */
  const pullAggregatedData = async function (sensorId: string = "2cf7f1c0613006fe"): Promise<void> {
    const now = new Date()

    const epoch = new Date()
    epoch.setFullYear(now.getFullYear() - 2)

    // TODO: fix
    const type: CalibratedDataNames = CalibratedDataNames.RAIN_ACCUMULATION
    const params = new URLSearchParams({
      start: epoch.toISOString(),
      end: now.toISOString(),
      dataType: type,
    })

    try {
      const response = await axios.get<AggregationData[]>(
        `${BASE_URL}/api/sensors/${sensorId}/data/aggregated_data?${params.toString()}`
      )

      if (response.data.length == 0) {
        return
      }

      const t: AggregatedDataMapping = {
        // type: type,
        data: {
          HOURLY: [],
          DAILY: [],
          WEEKLY: [],
          MONTHLY: [],
          YEARLY: [],
        },
      }

      response.data.forEach((d: AggregationData) => {
        const u: AggregatedDataPoint = {
          unit: d.totalValue?.unit!,
          value: d.totalValue?.value!,
          date: d.date,
        }

        switch (d.metadata.period) {
          case AGGREGATION_TYPE.HOURLY:
            t.data.HOURLY.push(u)
            break

          case AGGREGATION_TYPE.DAILY:
            t.data.DAILY.push(u)
            break

          case AGGREGATION_TYPE.WEEKLY:
            t.data.WEEKLY.push(u)
            break

          case AGGREGATION_TYPE.MONTHLY:
            t.data.MONTHLY.push(u)
            break

          case AGGREGATION_TYPE.YEARLY:
            t.data.YEARLY.push(u)
            break
        }
      })

      const x: Partial<CalibratedDataNamesGrouping> = {
        RAIN_GAUGE: t,
      }

      sensorToAggregationData.value.set(sensorId, x)
    } catch (e) {
      console.warn(e)
    }
  }

  onMounted(pullAggregatedData)
</script>
<style scoped>
  .card-title {
    text-align: center;
    margin-top: var(--cui-card-spacer-y);
  }
</style>
