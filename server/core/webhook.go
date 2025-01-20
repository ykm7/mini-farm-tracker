package core

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
https://www.thethingsindustries.com/docs/the-things-stack/concepts/data-formats/#uplink-messages
*/
// type UplinkMessage struct {
// 	EndDeviceIDs struct {
// 		DeviceID       string `json:"device_id"`
// 		ApplicationIDs struct {
// 			ApplicationID string `json:"application_id"`
// 		} `json:"application_ids"`
// 		// // DevEUI of the end device (eg: 0004A30B001C0530)
// 		DevEui string `json:"dev_eui"`
// 	} `json:"end_device_ids"`
// 	// // ISO 8601 UTC timestamp at which the message has been received by the Application Server (eg: "2020-02-12T15:15...")
// 	ReceivedAt    string `json:"received_at"`
// 	UplinkMessage struct {
// 		FPort uint32 `json:"f_port"`
// 		// // Frame payload (Base64)
// 		FrmPayload []byte `json:"frm_payload"`
// 		// Decoded payload object, decoded by the device payload formatter
// 		DecodedPayload map[string]interface{} `json:"decoded_payload"`
// 		RxMetadata     []struct {
// 			GatewayIDs struct {
// 				GatewayID string `json:"gateway_id"`
// 				EUI       string `json:"eui"`
// 			} `json:"gateway_ids"`
// 			// ISO 8601 UTC timestamp at which the uplink has been received by the gateway (et: "2020-02-12T15:15:45.787Z")
// 			Time         string  `json:"time"`
// 			Timestamp    int64   `json:"timestamp"`
// 			RSSI         int     `json:"rssi"`
// 			ChannelRSSI  int     `json:"channel_rssi"`
// 			SNR          float64 `json:"snr"`
// 			UplinkToken  string  `json:"uplink_token"`
// 			ChannelIndex int     `json:"channel_index"`
// 			Location     struct {
// 				Latitude  float64 `json:"latitude"`
// 				Longitude float64 `json:"longitude"`
// 				Altitude  int     `json:"altitude"`
// 				Source    string  `json:"source"`
// 			} `json:"location"`
// 		} `json:"rx_metadata"`
// 	} `json:"uplink_message"`
// }

type UplinkMessage struct {
	EndDeviceIDs struct {
		DeviceID       *string `json:"device_id,omitempty"`
		ApplicationIDs struct {
			ApplicationID *string `json:"application_id,omitempty"`
		} `json:"application_ids"`
		DevEUI  *string `json:"dev_eui,omitempty"`
		JoinEUI *string `json:"join_eui,omitempty"`
		DevAddr *string `json:"dev_addr,omitempty"`
	} `json:"end_device_ids"`
	CorrelationIDs *[]string `json:"correlation_ids,omitempty"`
	ReceivedAt     *string   `json:"received_at,omitempty"`
	UplinkMessage  struct {
		SessionKeyID   *string                `json:"session_key_id,omitempty"`
		FCount         *int                   `json:"f_cnt,omitempty"`
		FPort          *int                   `json:"f_port,omitempty"`
		FrmPayload     *string                `json:"frm_payload,omitempty"`
		DecodedPayload map[string]interface{} `json:"decoded_payload,omitempty"`
		RxMetadata     []struct {
			GatewayIDs struct {
				GatewayID *string `json:"gateway_id,omitempty"`
				EUI       *string `json:"eui,omitempty"`
			} `json:"gateway_ids"`
			Time         *string  `json:"time,omitempty"`
			Timestamp    *int64   `json:"timestamp,omitempty"`
			RSSI         *int     `json:"rssi,omitempty"`
			ChannelRSSI  *int     `json:"channel_rssi,omitempty"`
			SNR          *float64 `json:"snr,omitempty"`
			UplinkToken  *string  `json:"uplink_token,omitempty"`
			ChannelIndex *int     `json:"channel_index,omitempty"`
			Location     struct {
				Latitude  *float64 `json:"latitude,omitempty"`
				Longitude *float64 `json:"longitude,omitempty"`
				Altitude  *int     `json:"altitude,omitempty"`
				Source    *string  `json:"source,omitempty"`
			} `json:"location"`
		} `json:"rx_metadata,omitempty"`
		Settings struct {
			DataRate struct {
				Lora struct {
					Bandwidth       *int `json:"bandwidth,omitempty"`
					SpreadingFactor *int `json:"spreading_factor,omitempty"`
				} `json:"lora"`
			} `json:"data_rate"`
			CodingRate *string `json:"coding_rate,omitempty"`
			Frequency  *string `json:"frequency,omitempty"`
			Timestamp  *int64  `json:"timestamp,omitempty"`
			Time       *string `json:"time,omitempty"`
		} `json:"settings"`
		ReceivedAt      *string `json:"received_at,omitempty"`
		ConsumedAirtime *string `json:"consumed_airtime,omitempty"`
		Locations       map[string]struct {
			Latitude  *float64 `json:"latitude,omitempty"`
			Longitude *float64 `json:"longitude,omitempty"`
			Altitude  *int     `json:"altitude,omitempty"`
			Source    *string  `json:"source,omitempty"`
		} `json:"locations,omitempty"`
		VersionIDs struct {
			BrandID         *string `json:"brand_id,omitempty"`
			ModelID         *string `json:"model_id,omitempty"`
			HardwareVersion *string `json:"hardware_version,omitempty"`
			FirmwareVersion *string `json:"firmware_version,omitempty"`
			BandID          *string `json:"band_id,omitempty"`
		} `json:"version_ids"`
		NetworkIDs struct {
			NetID     *string `json:"net_id,omitempty"`
			TenantID  *string `json:"tenant_id,omitempty"`
			ClusterID *string `json:"cluster_id,omitempty"`
		} `json:"network_ids"`
		Simulated bool `json:"simulated"` // Keep as is since bool can't be nil
	} `json:"uplink_message"`
}

func handleWebhook(c *gin.Context, envs *environmentVariables) {
	apiKey := c.GetHeader("X-Downlink-Apikey")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-Downlink-Apikey header"})
		return
	}

	// Verify API Sign
	if apiKey != envs.ttn_webhhook_api {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Webhook env is invalid"})
		return
	}

	var uplinkMessage UplinkMessage
	if err := c.ShouldBindJSON(&uplinkMessage); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("'Raw' payload: %v\n", uplinkMessage.UplinkMessage.FrmPayload)

	log.Printf("'Raw' payload: %v\n", uplinkMessage.UplinkMessage.FrmPayload)
	log.Printf("'Decoded' payload: %v\n", uplinkMessage.UplinkMessage.DecodedPayload)

	log.Printf("device id is: %s\n", *uplinkMessage.EndDeviceIDs.DeviceID)
	log.Printf("device 'dev_eui' is: %s\n", *uplinkMessage.EndDeviceIDs.DevEUI)

	log.Printf("device 'rx_metadata' is: %v\n", uplinkMessage.UplinkMessage.RxMetadata)
	// TODO: Check its actually a device I care about

	// TODO: Store data point within Mongo

	// Respond with a success status
	c.JSON(http.StatusOK, gin.H{"message": "Webhook received successfully"})
}