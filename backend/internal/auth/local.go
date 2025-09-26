package auth

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/tanaymehhta/self/backend/internal/models"
)

type LocalAuthService struct {
	db         *gorm.DB
	jwtManager *JWTManager
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name"`
}

type AuthResponse struct {
	User   models.User `json:"user"`
	Tokens *TokenPair  `json:"tokens"`
}

// AuthUser represents the auth.users table
type AuthUser struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key"`
	Email              string    `gorm:"unique;not null"`
	EncryptedPassword  string    `gorm:"column:encrypted_password"`
	EmailConfirmedAt   *sql.NullTime
	CreatedAt          sql.NullTime `gorm:"default:now()"`
	UpdatedAt          sql.NullTime `gorm:"default:now()"`
	RawUserMetaData    []byte       `gorm:"type:jsonb;default:'{}'"`
}

func (AuthUser) TableName() string {
	return "auth.users"
}

func NewLocalAuthService(db *gorm.DB, jwtManager *JWTManager) *LocalAuthService {
	return &LocalAuthService{
		db:         db,
		jwtManager: jwtManager,
	}
}

func (s *LocalAuthService) Register(req RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	var existingAuthUser AuthUser
	if err := s.db.Where("email = ?", req.Email).First(&existingAuthUser).Error; err == nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create auth user
	authUser := AuthUser{
		ID:                uuid.New(),
		Email:             req.Email,
		EncryptedPassword: string(hashedPassword),
	}

	// Start transaction
	tx := s.db.Begin()
	if err := tx.Create(&authUser).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create auth user: %w", err)
	}

	// The trigger should create the public.users record automatically
	// Wait a moment and fetch the user
	var user models.User
	if err := tx.Where("id = ?", authUser.ID).First(&user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	tx.Commit()

	// Generate JWT tokens
	tokens, err := s.jwtManager.GenerateTokenPair(authUser.ID.String(), authUser.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		User:   user,
		Tokens: tokens,
	}, nil
}

func (s *LocalAuthService) Login(req LoginRequest) (*AuthResponse, error) {
	// Find auth user
	var authUser AuthUser
	if err := s.db.Where("email = ?", req.Email).First(&authUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(authUser.EncryptedPassword), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Get user profile
	var user models.User
	if err := s.db.Where("id = ?", authUser.ID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}

	// Generate JWT tokens
	tokens, err := s.jwtManager.GenerateTokenPair(authUser.ID.String(), authUser.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &AuthResponse{
		User:   user,
		Tokens: tokens,
	}, nil
}

func (s *LocalAuthService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}
	return &user, nil
}