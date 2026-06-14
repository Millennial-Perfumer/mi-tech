package test

import (
	"testing"

	"mi-tech/internal/domain/gst/dto"
	"mi-tech/internal/domain/gst/repository"
	"mi-tech/internal/domain/gst/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockGSTRepository struct {
	mock.Mock
}

func (m *mockGSTRepository) GetGSTSummary(startDate, endDate string) (repository.GSTSummaryResult, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).(repository.GSTSummaryResult), args.Error(1)
}

func (m *mockGSTRepository) GetStateSummary(startDate, endDate string) ([]repository.StateSummaryResult, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).([]repository.StateSummaryResult), args.Error(1)
}

func (m *mockGSTRepository) GetHSNSummary(startDate, endDate string) ([]repository.HSNSummaryResult, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).([]repository.HSNSummaryResult), args.Error(1)
}

func (m *mockGSTRepository) GetShopifyDocumentsIssued(startDate, endDate string) (*int64, *int64, int, int, error) {
	args := m.Called(startDate, endDate)
	var minVal, maxVal *int64
	if args.Get(0) != nil {
		minVal = args.Get(0).(*int64)
	}
	if args.Get(1) != nil {
		maxVal = args.Get(1).(*int64)
	}
	return minVal, maxVal, args.Int(2), args.Int(3), args.Error(4)
}

func (m *mockGSTRepository) GetAmazonDocumentsIssued(startDate, endDate string) (*int64, *int64, int, int, error) {
	args := m.Called(startDate, endDate)
	var minVal, maxVal *int64
	if args.Get(0) != nil {
		minVal = args.Get(0).(*int64)
	}
	if args.Get(1) != nil {
		maxVal = args.Get(1).(*int64)
	}
	return minVal, maxVal, args.Int(2), args.Int(3), args.Error(4)
}

func (m *mockGSTRepository) GetGSTR1B2CS(startDate, endDate string) ([]dto.B2CSRow, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).([]dto.B2CSRow), args.Error(1)
}

func (m *mockGSTRepository) GetGSTR1HSN(startDate, endDate string) ([]dto.HSNRow, error) {
	args := m.Called(startDate, endDate)
	return args.Get(0).([]dto.HSNRow), args.Error(1)
}

func TestGSTService_GetGSTR1JSON(t *testing.T) {
	mockRepo := new(mockGSTRepository)
	srv := service.NewGSTService(mockRepo)

	b2csRows := []dto.B2CSRow{
		{SplyTy: "INTRA", POS: "33", Rt: 18.0, TxVal: 100.0, Camt: 9.0, Samt: 9.0, Typ: "OE"},
	}
	hsnRows := []dto.HSNRow{
		{Num: 1, HsnSc: "330290", Desc: "Perfumes", Uqc: "PCS", Qty: 1.0, Val: 118.0, TxVal: 100.0, Camt: 9.0, Samt: 9.0},
	}

	mockRepo.On("GetGSTR1B2CS", "2026-06-01T00:00:00Z", "2026-06-30T23:59:59Z").Return(b2csRows, nil)
	mockRepo.On("GetGSTR1HSN", "2026-06-01T00:00:00Z", "2026-06-30T23:59:59Z").Return(hsnRows, nil)

	var shMin int64 = 2853
	var shMax int64 = 2855
	var amzMin int64 = 1
	var amzMax int64 = 1

	mockRepo.On("GetShopifyDocumentsIssued", "2026-06-01T00:00:00Z", "2026-06-30T23:59:59Z").
		Return(&shMin, &shMax, 2, 0, nil)

	mockRepo.On("GetAmazonDocumentsIssued", "2026-06-01T00:00:00Z", "2026-06-30T23:59:59Z").
		Return(&amzMin, &amzMax, 1, 0, nil)

	gstin := "33AUSPR1909H1ZC"
	payload, err := srv.GetGSTR1JSON("2026-06-01T00:00:00Z", "2026-06-30T23:59:59Z", gstin)

	assert.NoError(t, err)
	assert.Equal(t, gstin, payload.GSTIN)
	assert.Equal(t, "062026", payload.FP)
	assert.Len(t, payload.B2CS, 1)
	assert.Equal(t, "INTRA", payload.B2CS[0].SplyTy)

	assert.Len(t, payload.HSN.Data, 1)
	assert.Equal(t, "330290", payload.HSN.Data[0].HsnSc)

	assert.Len(t, payload.DocIssue.DocDet, 1)
	assert.Len(t, payload.DocIssue.DocDet[0].Docs, 2)

	assert.Equal(t, "SY-2853", payload.DocIssue.DocDet[0].Docs[0].From)
	assert.Equal(t, "SY-2855", payload.DocIssue.DocDet[0].Docs[0].To)
	assert.Equal(t, "AMZ-1", payload.DocIssue.DocDet[0].Docs[1].From)
	assert.Equal(t, "AMZ-1", payload.DocIssue.DocDet[0].Docs[1].To)

	mockRepo.AssertExpectations(t)
}
