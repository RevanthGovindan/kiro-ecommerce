package users

import (
	"errors"

	"ecommerce-website/internal/models"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

type UpdateProfileRequest struct {
	FirstName string  `json:"firstName" binding:"required"`
	LastName  string  `json:"lastName" binding:"required"`
	Phone     *string `json:"phone,omitempty"`
}

type CreateAddressRequest struct {
	Type       string  `json:"type" binding:"required,oneof=shipping billing"`
	FirstName  string  `json:"firstName" binding:"required"`
	LastName   string  `json:"lastName" binding:"required"`
	Company    *string `json:"company,omitempty"`
	Address1   string  `json:"address1" binding:"required"`
	Address2   *string `json:"address2,omitempty"`
	City       string  `json:"city" binding:"required"`
	State      string  `json:"state" binding:"required"`
	PostalCode string  `json:"postalCode" binding:"required"`
	Country    string  `json:"country" binding:"required"`
	Phone      *string `json:"phone,omitempty"`
	IsDefault  bool    `json:"isDefault"`
}

type UpdateAddressRequest struct {
	FirstName  string  `json:"firstName" binding:"required"`
	LastName   string  `json:"lastName" binding:"required"`
	Company    *string `json:"company,omitempty"`
	Address1   string  `json:"address1" binding:"required"`
	Address2   *string `json:"address2,omitempty"`
	City       string  `json:"city" binding:"required"`
	State      string  `json:"state" binding:"required"`
	PostalCode string  `json:"postalCode" binding:"required"`
	Country    string  `json:"country" binding:"required"`
	Phone      *string `json:"phone,omitempty"`
	IsDefault  bool    `json:"isDefault"`
}

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrAddressNotFound = errors.New("address not found")
	ErrUnauthorized    = errors.New("unauthorized access to address")
)

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}

// GetProfile retrieves user profile information
func (s *Service) GetProfile(userID string) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("Addresses").Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Remove password from response
	user.Password = ""
	return &user, nil
}

// UpdateProfile updates user profile information
func (s *Service) UpdateProfile(userID string, req UpdateProfileRequest) (*models.User, error) {
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Update user fields
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Phone = req.Phone

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}

	// Remove password from response
	user.Password = ""
	return &user, nil
}

// GetUserOrders retrieves user's order history
func (s *Service) GetUserOrders(userID string) ([]models.Order, error) {
	var orders []models.Order
	if err := s.db.Preload("Items").Preload("Items.Product").Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}

	return orders, nil
}

// CreateAddress creates a new address for the user
func (s *Service) CreateAddress(userID string, req CreateAddressRequest) (*models.Address, error) {
	// Verify user exists
	var user models.User
	if err := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// If this is set as default, unset other default addresses of the same type
	if req.IsDefault {
		if err := s.db.Model(&models.Address{}).Where("user_id = ? AND type = ?", userID, req.Type).Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}

	address := models.Address{
		UserID:     userID,
		Type:       req.Type,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Company:    req.Company,
		Address1:   req.Address1,
		Address2:   req.Address2,
		City:       req.City,
		State:      req.State,
		PostalCode: req.PostalCode,
		Country:    req.Country,
		Phone:      req.Phone,
		IsDefault:  req.IsDefault,
	}

	if err := s.db.Create(&address).Error; err != nil {
		return nil, err
	}

	return &address, nil
}

// GetAddresses retrieves all addresses for a user
func (s *Service) GetAddresses(userID string) ([]models.Address, error) {
	var addresses []models.Address
	if err := s.db.Where("user_id = ?", userID).Order("is_default DESC, created_at DESC").Find(&addresses).Error; err != nil {
		return nil, err
	}

	return addresses, nil
}

// GetAddress retrieves a specific address for a user
func (s *Service) GetAddress(userID, addressID string) (*models.Address, error) {
	var address models.Address
	if err := s.db.Where("id = ? AND user_id = ?", addressID, userID).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAddressNotFound
		}
		return nil, err
	}

	return &address, nil
}

// UpdateAddress updates an existing address
func (s *Service) UpdateAddress(userID, addressID string, req UpdateAddressRequest) (*models.Address, error) {
	var address models.Address
	if err := s.db.Where("id = ? AND user_id = ?", addressID, userID).First(&address).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAddressNotFound
		}
		return nil, err
	}

	// If this is set as default, unset other default addresses of the same type
	if req.IsDefault && !address.IsDefault {
		if err := s.db.Model(&models.Address{}).Where("user_id = ? AND type = ? AND id != ?", userID, address.Type, addressID).Update("is_default", false).Error; err != nil {
			return nil, err
		}
	}

	// Update address fields
	address.FirstName = req.FirstName
	address.LastName = req.LastName
	address.Company = req.Company
	address.Address1 = req.Address1
	address.Address2 = req.Address2
	address.City = req.City
	address.State = req.State
	address.PostalCode = req.PostalCode
	address.Country = req.Country
	address.Phone = req.Phone
	address.IsDefault = req.IsDefault

	if err := s.db.Save(&address).Error; err != nil {
		return nil, err
	}

	return &address, nil
}

// DeleteAddress deletes an address
func (s *Service) DeleteAddress(userID, addressID string) error {
	result := s.db.Where("id = ? AND user_id = ?", addressID, userID).Delete(&models.Address{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrAddressNotFound
	}

	return nil
}
