package controllers

import (
	"backendtku/app/repositories"
	"backendtku/config"

	"gorm.io/gorm"
)

type Handler struct {
	DB          *gorm.DB
	Config      *config.Config
	UserRepo    repositories.UserRepository
	ProductRepo repositories.ProductRepository
	EnrollRepo  repositories.EnrollmentRepository
}

func NewHandler(db *gorm.DB, cfg *config.Config, userRepo repositories.UserRepository, productRepo repositories.ProductRepository, enrollRepo repositories.EnrollmentRepository) *Handler {
	return &Handler{
		DB:          db,
		Config:      cfg,
		UserRepo:    userRepo,
		ProductRepo: productRepo,
		EnrollRepo:  enrollRepo,
	}
}
