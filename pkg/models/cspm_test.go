package models

import (
	"encoding/json"
	"testing"
)

func TestFlexInt_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		want      int
		wantError bool
	}{
		{
			name: "文字列型の数値",
			json: `{"totalCount":"485"}`,
			want: 485,
		},
		{
			name: "整数型の数値",
			json: `{"totalCount":485}`,
			want: 485,
		},
		{
			name: "文字列型のゼロ",
			json: `{"totalCount":"0"}`,
			want: 0,
		},
		{
			name: "整数型のゼロ",
			json: `{"totalCount":0}`,
			want: 0,
		},
		{
			name:      "無効な文字列",
			json:      `{"totalCount":"invalid"}`,
			wantError: true,
		},
		{
			name:      "空文字列",
			json:      `{"totalCount":""}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response struct {
				TotalCount FlexInt `json:"totalCount"`
			}

			err := json.Unmarshal([]byte(tt.json), &response)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if got := response.TotalCount.Int(); got != tt.want {
				t.Errorf("FlexInt.Int() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFlexInt_Int(t *testing.T) {
	tests := []struct {
		name string
		fi   FlexInt
		want int
	}{
		{
			name: "正の数",
			fi:   FlexInt(123),
			want: 123,
		},
		{
			name: "ゼロ",
			fi:   FlexInt(0),
			want: 0,
		},
		{
			name: "負の数",
			fi:   FlexInt(-456),
			want: -456,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fi.Int(); got != tt.want {
				t.Errorf("FlexInt.Int() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestComplianceResponse_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantCount   int
		wantDataLen int
		wantError   bool
	}{
		{
			name: "totalCountが文字列のレスポンス",
			json: `{
				"data": [
					{
						"requirementId": "req-1",
						"name": "Test Requirement",
						"policyId": "policy-1",
						"policyName": "Test Policy",
						"platform": "AWS",
						"pass": false
					}
				],
				"totalCount": "1"
			}`,
			wantCount:   1,
			wantDataLen: 1,
		},
		{
			name: "totalCountが整数のレスポンス",
			json: `{
				"data": [
					{
						"requirementId": "req-1",
						"name": "Test Requirement",
						"policyId": "policy-1",
						"policyName": "Test Policy",
						"platform": "AWS",
						"pass": false
					}
				],
				"totalCount": 1
			}`,
			wantCount:   1,
			wantDataLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response ComplianceResponse

			err := json.Unmarshal([]byte(tt.json), &response)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if got := response.TotalCount.Int(); got != tt.wantCount {
				t.Errorf("TotalCount = %d, want %d", got, tt.wantCount)
			}

			if got := len(response.Data); got != tt.wantDataLen {
				t.Errorf("len(Data) = %d, want %d", got, tt.wantDataLen)
			}
		})
	}
}

