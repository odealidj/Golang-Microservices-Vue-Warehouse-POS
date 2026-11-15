package usecase

import (
	"context"
	"errors"
	"micro-warehouse/user-service/model"
	"micro-warehouse/user-service/service"
	"sync"
	"testing"
	"time"
)

type stubUserRepo struct {
	users         map[uint]model.User
	lastCreated   model.User
	lastPage      int
	lastLimit     int
	deletedID     uint
	deleteCalled  bool
	getUserErr    error
}

func (s *stubUserRepo) CreateUser(ctx context.Context, user model.User) (*model.User, error) {
	s.lastCreated = user
	user.ID = 7
	return &user, nil
}
func (s *stubUserRepo) GetAllUsers(ctx context.Context, page, limit int, search, sortBy, sortOrder string) ([]model.User, int64, error) {
	s.lastPage = page
	s.lastLimit = limit
	return nil, 0, nil
}
func (s *stubUserRepo) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	if s.getUserErr != nil {
		return nil, s.getUserErr
	}
	u, ok := s.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &u, nil
}
func (s *stubUserRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, nil
}
func (s *stubUserRepo) UpdateUser(ctx context.Context, user model.User) error { return nil }
func (s *stubUserRepo) DeleteUser(ctx context.Context, id uint) error {
	s.deleteCalled = true
	s.deletedID = id
	return nil
}
func (s *stubUserRepo) GetUserByRoleName(ctx context.Context, roleName string) ([]model.User, error) {
	return nil, nil
}
func (s *stubUserRepo) AssignUserToRole(ctx context.Context, userID uint, roleID uint) error {
	return nil
}
func (s *stubUserRepo) EditAssignUserToRole(ctx context.Context, assignRoleID uint, userID uint, roleID uint) error {
	return nil
}
func (s *stubUserRepo) GetUserRoleByID(ctx context.Context, assignRoleID uint) (*model.UserRole, error) {
	return nil, nil
}
func (s *stubUserRepo) GetAllUserRoles(ctx context.Context, page, limit int, search, sortBy, sortOrder string) ([]model.UserRole, int64, error) {
	return nil, 0, nil
}

type stubRabbitMQ struct {
	mu      sync.Mutex
	wg      sync.WaitGroup
	payload service.EmailPayload
	called  bool
}

func (s *stubRabbitMQ) PublishEmail(ctx context.Context, payload service.EmailPayload) error {
	s.mu.Lock()
	s.payload = payload
	s.called = true
	s.mu.Unlock()
	s.wg.Done()
	return nil
}
func (s *stubRabbitMQ) Close() error { return nil }

func TestCreateUser_HashesPasswordAndPublishesWelcomeWithoutPlaintext(t *testing.T) {
	repo := &stubUserRepo{}
	rmq := &stubRabbitMQ{}
	rmq.wg.Add(1)
	uc := NewUserUsecase(repo, rmq)

	err := uc.CreateUser(context.Background(), model.User{
		Name:     "Siti",
		Email:    "siti@example.com",
		Password: "secret-password",
	})
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if repo.lastCreated.Password == "" || repo.lastCreated.Password == "secret-password" {
		t.Fatalf("expected hashed password, got %q", repo.lastCreated.Password)
	}

	done := make(chan struct{})
	go func() {
		rmq.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for PublishEmail goroutine")
	}

	rmq.mu.Lock()
	defer rmq.mu.Unlock()
	if !rmq.called || rmq.payload.Password != "" || rmq.payload.Type != "welcome_email" || rmq.payload.Email != "siti@example.com" {
		t.Fatalf("unexpected email payload published: %+v", rmq.payload)
	}
}

func TestGetAllUsers_NormalizesPagination(t *testing.T) {
	repo := &stubUserRepo{}
	uc := NewUserUsecase(repo, &stubRabbitMQ{})

	if _, _, err := uc.GetAllUsers(context.Background(), 0, 500, "", "", ""); err != nil {
		t.Fatalf("GetAllUsers failed: %v", err)
	}
	if repo.lastPage != 1 || repo.lastLimit != 10 {
		t.Fatalf("expected normalized page=1 limit=10, got page=%d limit=%d", repo.lastPage, repo.lastLimit)
	}
}

func TestDeleteUser_ChecksExistenceFirst(t *testing.T) {
	repo := &stubUserRepo{getUserErr: errors.New("user not found")}
	uc := NewUserUsecase(repo, &stubRabbitMQ{})

	err := uc.DeleteUser(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error when user does not exist")
	}
	if repo.deleteCalled {
		t.Fatal("expected repo.DeleteUser not to be called when GetUserByID fails")
	}
}
