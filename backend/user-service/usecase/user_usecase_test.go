package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"micro-warehouse/user-service/mocks"
	"micro-warehouse/user-service/model"
	"micro-warehouse/user-service/service"

	"go.uber.org/mock/gomock"
)

func TestCreateUser_HashesPasswordAndPublishesWelcomeWithoutPlaintext(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepositoryInterface(ctrl)
	rmq := mocks.NewMockRabbitMQServiceInterface(ctrl)

	var (
		mu          sync.Mutex
		wg          sync.WaitGroup
		gotPayload  service.EmailPayload
		emailCalled bool
	)
	wg.Add(1)

	repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, user model.User) (*model.User, error) {
			if user.Password == "" || user.Password == "secret-password" {
				t.Fatalf("expected hashed password, got %q", user.Password)
			}
			user.ID = 7
			return &user, nil
		})

	rmq.EXPECT().PublishEmail(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, payload service.EmailPayload) error {
			mu.Lock()
			gotPayload = payload
			emailCalled = true
			mu.Unlock()
			wg.Done()
			return nil
		})

	uc := NewUserUsecase(repo, rmq)
	err := uc.CreateUser(context.Background(), model.User{
		Name:     "Siti",
		Email:    "siti@example.com",
		Password: "secret-password",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for PublishEmail goroutine")
	}

	mu.Lock()
	defer mu.Unlock()
	if !emailCalled || gotPayload.Password != "" || gotPayload.Type != "welcome_email" || gotPayload.Email != "siti@example.com" {
		t.Fatalf("unexpected email payload published: %+v", gotPayload)
	}
}

func TestGetAllUsers_NormalizesPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepositoryInterface(ctrl)
	rmq := mocks.NewMockRabbitMQServiceInterface(ctrl)

	repo.EXPECT().GetAllUsers(gomock.Any(), 1, 10, "", "", "").Return(nil, int64(0), nil)

	uc := NewUserUsecase(repo, rmq)
	if _, _, err := uc.GetAllUsers(context.Background(), 0, 500, "", "", ""); err != nil {
		t.Fatalf("GetAllUsers failed: %v", err)
	}
}

func TestDeleteUser_ChecksExistenceFirst(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockUserRepositoryInterface(ctrl)
	rmq := mocks.NewMockRabbitMQServiceInterface(ctrl)

	repo.EXPECT().GetUserByID(gomock.Any(), uint(99)).Return(nil, errors.New("user not found"))
	repo.EXPECT().DeleteUser(gomock.Any(), gomock.Any()).Times(0)

	uc := NewUserUsecase(repo, rmq)
	if err := uc.DeleteUser(context.Background(), 99); err == nil {
		t.Fatal("expected error when user does not exist")
	}
}
