package core

import "testing"

// func Test_parseLDDS45(t *testing.T) {
// 	type args struct {
// 		bs []byte
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    *LDDS45RawData
// 		wantErr bool
// 	}{
// 		{
// 			args: args{
// 				bs: []byte{
// 					0x0B, 0x49, // - 2889mV
// 					0x00, 0x00, // 0B05(H) = 2821 (D) = 0 mm.
// 					0x01,       // Normal uplink packet.
// 					0xFF, 0x3F, // (FF3F & FC00 == 1) , temp = (FF3FH - 65536)/10 = -19.3 degrees.
// 					0x00, // Detect Ultrasonic Sensor
// 				},
// 			},
// 			want: &LDDS45RawData{
// 				Battery:      2889,
// 				Distance:     0,
// 				InterruptPin: 1,
// 				Temperature:  -19.3,
// 				SensorFlag:   0,
// 			},
// 			wantErr: false,
// 		},
// 		{
// 			args: args{
// 				bs: []byte{
// 					0x0B, 0x45, // - 2885mV
// 					0x0B, 0x05, // 0B05(H) = 2821 (D) = 2821 mm.
// 					0x00,       // Normal uplink packet.
// 					0x01, 0x05, // (0105 & FC00 == 0), temp = 0105H /10 = 26.1 degree
// 					0x01, // Detect Ultrasonic Sensor
// 				},
// 			},
// 			want: &LDDS45RawData{
// 				Battery:      2885,
// 				Distance:     2821,
// 				InterruptPin: 0,
// 				Temperature:  26.1,
// 				SensorFlag:   1,
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := parseLDDS45(tt.args.bs)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("parseLDDS45() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("parseLDDS45() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestLDDS45RawData_DetermineValid(t *testing.T) {
	tests := []struct {
		name       string
		sensorData SensorData
		want       struct {
			valid bool
			err   error
		}
	}{
		{
			name: "Testing raw LDDS45RawData data metric which should be valid",
			sensorData: SensorData{
				LDDS45: &LDDS45RawData{
					Distance: "2205 mm",
				},
			},
			want: struct {
				valid bool
				err   error
			}{
				valid: true,
				err:   nil,
			},
		},
		{
			name: "Testing raw LDDS45RawData data metric which should be not valid",
			sensorData: SensorData{
				LDDS45: &LDDS45RawData{
					Distance: "no value",
				},
			},
			want: struct {
				valid bool
				err   error
			}{
				valid: false,
				err:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := tt.sensorData.DetermineValid()

			if err != tt.want.err {
				t.Errorf("LDDS45RawData.DetermineValid() = %v, want %v", err, tt.want.err)
			}

			if valid != tt.want.valid {
				t.Errorf("LDDS45RawData.DetermineValid() = %v, want %v", valid, tt.want.valid)
			}
		})
	}
}
