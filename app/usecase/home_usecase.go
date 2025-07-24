package usecase

import (
	"backendtku/app/dto"
)

type HomeUseCase interface {
	GetHomeMessage() dto.BaseResponse
}

type homeUseCase struct {
}

func NewHomeUseCase() HomeUseCase {
	return &homeUseCase{}
}

func (uc *homeUseCase) GetHomeMessage() dto.BaseResponse {
	var Response dto.BaseResponse
	Response.Status = true
	Response.Message = "Welcome to Takaful Umum Back End"
	Response.Data = nil
	return Response
}
