import { ObjectId } from 'mongodb'

interface Data {
  Id?: ObjectId
  Sensor: string
  Timestamp: Date
  Data: number
}

export interface RawData extends Data {}

export interface CalibratedData extends Data {}
