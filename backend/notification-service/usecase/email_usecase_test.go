package usecase

import (
	"context"
	"errors"
	"testing"

	"micro-warehouse/notificaiton-service/controller/request"
	"micro-warehouse/notificaiton-service/mocks"
	"micro-warehouse/notificaiton-service/pkg/email"

	"go.uber.org/mock/gomock"
)

func TestSendEmail_ForwardsFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockEmailServiceInterface(ctrl)

	svc.EXPECT().SendCustomEmail(gomock.Any(), "kasir@example.com", "Laporan Harian", "Omzet hari ini tercatat.").Return(nil)

	uc := NewEmailUsecase(svc)
	err := uc.SendEmail(context.Background(), request.SendEmailRequest{
		To:      "kasir@example.com",
		Subject: "Laporan Harian",
		Body:    "Omzet hari ini tercatat.",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestSendWelcomeEmail_BuildsWelcomePayloadAndPropagatesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockEmailServiceInterface(ctrl)

	svc.EXPECT().SendWelcomeEmail(gomock.Any(), email.EmailPayload{
		Type:   "welcome",
		Email:  "keeper@example.com",
		UserID: 42,
		Name:   "Andi",
	}).Return(nil)

	wantErr := errors.New("smtp unavailable")
	svc.EXPECT().SendWelcomeEmail(gomock.Any(), email.EmailPayload{
		Type:  "welcome",
		Email: "a@b.com",
	}).Return(wantErr)

	uc := NewEmailUsecase(svc)
	if err := uc.SendWelcomeEmail(context.Background(), request.SendWelcomeEmailRequest{
		Email:  "keeper@example.com",
		UserID: 42,
		Name:   "Andi",
	}); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if err := uc.SendWelcomeEmail(context.Background(), request.SendWelcomeEmailRequest{Email: "a@b.com"}); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
