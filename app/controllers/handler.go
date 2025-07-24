package controllers

import (
	
	"backendtku/app/usecase"
	"backendtku/config"

	"gorm.io/gorm"
)

type Handler struct {
	DB          *gorm.DB
	Config      *config.Config
	
	AdminUseCase usecase.AdminUseCase
	AuthUseCase usecase.AuthUseCase
	ClaimUseCase usecase.ClaimUseCase
	EnrollmentUseCase usecase.EnrollmentUseCase
	HomeUseCase usecase.HomeUseCase
	ProductUseCase usecase.ProductUseCase
	ProfileUseCase usecase.ProfileUseCase
}

func NewHandler(db *gorm.DB, cfg *config.Config, adminUseCase usecase.AdminUseCase, authUseCase usecase.AuthUseCase, claimUseCase usecase.ClaimUseCase, enrollmentUseCase usecase.EnrollmentUseCase, homeUseCase usecase.HomeUseCase, productUseCase usecase.ProductUseCase, profileUseCase usecase.ProfileUseCase) *Handler {
	return &Handler{
		DB:          db,
		Config:      cfg,
		
		AdminUseCase: adminUseCase,
		AuthUseCase: authUseCase,
		ClaimUseCase: claimUseCase,
		EnrollmentUseCase: enrollmentUseCase,
		HomeUseCase: homeUseCase,
		ProductUseCase: productUseCase,
		ProfileUseCase: profileUseCase,
	}
}
