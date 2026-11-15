package usecase

import (
	"context"
	"errors"
	"micro-warehouse/notificaiton-service/controller/request"
	"micro-warehouse/notificaiton-service/pkg/email"
	"testing"
)

type stubEmailService struct {
	customTo      string
	customSubject string
	customBody    string
	welcomeCalled bool
	welcomeInput  email.EmailPayload
	err           error
}

func (s *stubEmailService) SendWelcomeEmail(ctx context.Context, payload email.EmailPayload) error {
	s.welcomeCalled = true
	s.welcomeInput = payload
	return s.err
}

func (s *stubEmailService) SendCustomEmail(ctx context.Context, to, subject, body string) error {
	s.customTo = to
	s.customSubject = subject
	s.customBody = body
	return s.err
}

func TestSendEmail_ForwardsFields(t *testing.T) {
	svc := &stubEmailService{}
	uc := NewEmailUsecase(svc)

	err := uc.SendEmail(context.Background(), request.SendEmailRequest{
		To:      "kasir@example.com",
		Subject: "Laporan Harian",
		Body:    "Omzet hari ini tercatat.",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if svc.customTo != "kasir@example.com" || svc.customSubject != "Laporan Harian" || svc.customBody != "Omzet hari ini tercatat." {
		t.Fatalf("unexpected forwarded email fields: %+v", svc)
	}
}

func TestSendWelcomeEmail_BuildsWelcomePayloadAndPropagatesError(t *testing.T) {
	svc := &stubEmailService{}
	uc := NewEmailUsecase(svc)

	err := uc.SendWelcomeEmail(context.Background(), request.SendWelcomeEmailRequest{
		Email:  "keeper@example.com",
		UserID: 42,
		Name:   "Andi",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !svc.welcomeCalled {
		t.Fatal("expected SendWelcomeEmail to be called")
	}
	if svc.welcomeInput.Type != "welcome" || svc.welcomeInput.Email != "keeper@example.com" || svc.welcomeInput.UserID != 42 || svc.welcomeInput.Name != "Andi" {
		t.Fatalf("unexpected welcome payload: %+v", svc.welcomeInput)
	}

	wantErr := errors.New("smtp unavailable")
	svc.err = wantErr
	if err := uc.SendWelcomeEmail(context.Background(), request.SendWelcomeEmailRequest{Email: "a@b.com"}); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
