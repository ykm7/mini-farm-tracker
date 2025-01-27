import { ObjectId } from 'mongodb'

interface AssetMetricsCylinderVolume {
    Volume: number;
    Radius: number;
    Height: number;
}

interface AssetMetrics {
    Volume?: AssetMetricsCylinderVolume;
}

export interface Asset {
    Id: ObjectId;
    Name: string;
    Description: string;
    Sensors?: string[];
    Metrics?: AssetMetrics;
}