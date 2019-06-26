package internal

import (
	"testing"

	"github.com/solution9th/S3Adapter/internal/auth"
	"github.com/solution9th/S3Adapter/internal/gateway"
	"github.com/solution9th/S3Adapter/mocks/mock_gateway"

	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
)

func TestNewGateway(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	convey.Convey("NewGateway", t, func() {

		tests := []struct {
			desc        string
			gatewayName string
			f           func() gateway.Gateway
			wantErr     error
		}{
			{
				"success",
				"testGateeay",
				func() gateway.Gateway {
					gt := mock_gateway.NewMockGateway(ctrl)

					gt.EXPECT().Name().Return("succss").AnyTimes()
					gt.EXPECT().Production().Return(true).AnyTimes()

					gt.EXPECT().NewS3Protocol(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

					return gt
				},
				nil,
			},
			{
				"err: not Production",
				"testGateeayNotProduction",
				func() gateway.Gateway {
					gt := mock_gateway.NewMockGateway(ctrl)
					gt.EXPECT().Production().Return(false).Times(1)
					return gt
				},
				ErrGatewayNotFound,
			},
		}

		for _, test := range tests {

			convey.Convey(test.desc, func() {
				GatewayMap[test.gatewayName] = test.f

				_, got := NewGateway(test.gatewayName, auth.Credentials{}, "")

				convey.So(got, convey.ShouldEqual, test.wantErr)
			})
		}

	})
}
