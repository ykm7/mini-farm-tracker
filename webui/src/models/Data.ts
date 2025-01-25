import { ObjectId } from 'mongodb'

interface Data {
  Id?: ObjectId // no need to be provided
  Sensor?: string // no need to be provided
  Timestamp: string
  Data: any
}

interface LDDS45RawData {
  Bat: number
  Distance: string
  Interrupt_flag: number
  TempC_DS18B20: string
  Sensor_flag: number
}

export interface RawData extends Data {
  Data: LDDS45RawData // TODO: RawData should be able to take various
}

export interface CalibratedData extends Data {}
