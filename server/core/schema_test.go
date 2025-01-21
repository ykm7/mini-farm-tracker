package core

import (
	"reflect"
	"testing"
)

func Test_parseLDDS45(t *testing.T) {
	type args struct {
		bs []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *LDDS45RawData
		wantErr bool
	}{
		{
			args: args{
				bs: []byte{
					0x0B, 0x45, // - 2885mV
					0x0B, 0x05, // 0B05(H) = 2821 (D) = 2821 mm.
					0x00,       // Normal uplink packet.
					0x01, 0x05, // (0105 & FC00 == 0), temp = 0105H /10 = 26.1 degree
					0x01, // Detect Ultrasonic Sensor
				},
			},
			want: &LDDS45RawData{
				Battery:      2885,
				Distance:     2821,
				InterruptPin: 0,
				Temperature:  26.1,
				SensorFlag:   1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLDDS45(tt.args.bs)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLDDS45() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseLDDS45() = %v, want %v", got, tt.want)
			}
		})
	}
}
