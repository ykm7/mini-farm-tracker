// Daily Aggregates Collection
db.createCollection("daily_aggregates", {
  timeseries: {
    timeField: "date",
    metaField: "metadata",
    granularity: "days",
  },
});

// Weekly Aggregates Collection
db.createCollection("weekly_aggregates", {
  timeseries: {
    timeField: "date",
    metaField: "metadata",
    granularity: "days", // Weekly aggregation still uses daily granularity
  },
});

// Monthly Aggregates Collection
db.createCollection("monthly_aggregates", {
  timeseries: {
    timeField: "date",
    metaField: "metadata",
    granularity: "days", // Monthly aggregation can use hourly or daily granularity
  },
});

// Yearly Aggregates Collection
db.createCollection("yearly_aggregates", {
  timeseries: {
    timeField: "date",
    metaField: "metadata",
    granularity: "days", // yearly aggregation can use hourly or daily granularity
  },
});

db.createCollection("aggregates", {
    timeseries: {
      timeField: "date",
      metaField: "metadata",
      granularity: "days"
    }
  });

// aggregations
db.calibrated_data.aggregate([
    {
      $match: {
        "dataPoints.rainfallHourly": { $exists: true }
      }
    },
    {
      $group: {
        _id: {
          date: { $dateToString: { format: "%Y-%m-%d", date: "$timestamp" } },
          sensor: "$sensor"
        },
        totalRainfall: { $sum: "$dataPoints.rainfallHourly.data" },
        // takes the unit of the first document within the group. 
        // This is counter to the purpose of storing units for each point.
        unit: { $first: "$dataPoints.rainfallHourly.units" }
      }
    },
    {
      $project: {
        _id: 0,
        date: { $dateFromString: { dateString: "$_id.date" } },
        metadata: { 
          sensor: "$_id.sensor",
          type: "daily"
        },
        totalRainfall: {
          value: "$totalRainfall",
          unit: "$unit"
        }
      }
    },
    {
      $merge: {
        into: "daily_aggregates",
        on: ["date", "metadata.sensor"],
        whenMatched: "replace",
        whenNotMatched: "insert"
      }
    }
  ]);