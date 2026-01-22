package pg

import (
	"github.com/golang/mock/gomock"
	models "github.com/iliaonishchenko/aggreg8/internal/model"
	"github.com/iliaonishchenko/aggreg8/internal/service/pg/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetMetric(t *testing.T) {
	type want struct {
		metric *models.Metrics
		err    error
	}
	tests := []struct {
		name      string
		metricID  string
		mockSetup func(m *mocks.MockRepository)
		want      want
	}{
		{
			name:     "successfully retrieve existing metric",
			metricID: "id",
			mockSetup: func(m *mocks.MockRepository) {
				value := models.Metrics{ID: "id", MType: models.Gauge, Delta: nil, Value: fltPtr(42.42), Hash: ""}
				m.EXPECT().Get("id").Return(&value, nil)
			},
			want: want{
				metric: &models.Metrics{ID: "id", MType: models.Gauge, Delta: nil, Value: fltPtr(42.42), Hash: ""},
				err:    nil,
			},
		},
		{
			name:     "metric not found",
			metricID: "missing_id",
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().Get("missing_id").Return(nil, nil)
			},
			want: want{
				metric: nil,
				err:    nil,
			},
		},
		{
			name:     "database error",
			metricID: "error_id",
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().Get("error_id").Return(nil, assert.AnError)
			},
			want: want{
				metric: nil,
				err:    assert.AnError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockRepository(ctrl)
			tt.mockSetup(m)

			storage := NewPostgresStorage(m)

			res, err := storage.GetMetric(tt.metricID)
			require.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.metric, res)
		})
	}
}

func TestGetAllMetrics(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(m *mocks.MockRepository)
		want      []*models.Metrics
	}{
		{
			name: "successfully retrieve all metrics",
			mockSetup: func(m *mocks.MockRepository) {
				metrics := []*models.Metrics{
					{ID: "id1", MType: models.Gauge, Delta: nil, Value: fltPtr(10.0), Hash: ""},
					{ID: "id2", MType: models.Counter, Delta: intPtr(5), Value: nil, Hash: ""},
				}
				m.EXPECT().GetAll().Return(metrics, nil)
			},
			want: []*models.Metrics{
				{ID: "id1", MType: models.Gauge, Delta: nil, Value: fltPtr(10.0), Hash: ""},
				{ID: "id2", MType: models.Counter, Delta: intPtr(5), Value: nil, Hash: ""},
			},
		},
		{
			name: "no metrics found",
			mockSetup: func(m *mocks.MockRepository) {
				metrics := []*models.Metrics{}
				m.EXPECT().GetAll().Return(metrics, nil)
			},
			want: []*models.Metrics{},
		},
		{
			name: "database error",
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().GetAll().Return(nil, assert.AnError)
			},
			want: []*models.Metrics{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockRepository(ctrl)
			tt.mockSetup(m)

			storage := NewPostgresStorage(m)

			res := storage.GetAllMetrics()
			assert.Equal(t, tt.want, res)
		})
	}
}

func TestUpdateMetric(t *testing.T) {
	tests := []struct {
		name      string
		newMetric *models.Metrics
		mockSetup func(m *mocks.MockRepository)
		want      bool
	}{
		{
			name:      "successfully update metric",
			newMetric: &models.Metrics{ID: "id", MType: models.Gauge, Delta: nil, Value: fltPtr(99.9), Hash: ""},
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().Update(&models.Metrics{ID: "id", MType: models.Gauge, Delta: nil, Value: fltPtr(99.9), Hash: ""}).Return(nil)
			},
			want: true,
		},
		{
			name:      "failed to update metric",
			newMetric: &models.Metrics{ID: "id", MType: models.Counter, Delta: intPtr(10), Value: nil, Hash: ""},
			mockSetup: func(m *mocks.MockRepository) {
				m.EXPECT().Update(&models.Metrics{ID: "id", MType: models.Counter, Delta: intPtr(10), Value: nil, Hash: ""}).Return(assert.AnError)
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockRepository(ctrl)
			tt.mockSetup(m)

			storage := NewPostgresStorage(m)

			result := storage.UpdateMetric(tt.newMetric)
			assert.Equal(t, tt.want, result)
		})
	}
}

func fltPtr(f float64) *float64 {
	return &f
}

func intPtr(f int64) *int64 {
	return &f
}
