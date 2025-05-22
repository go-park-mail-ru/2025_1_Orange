package service

import (
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"ResuMatch/internal/utils"
	l "ResuMatch/pkg/logger"
	"context"

	// "fmt"

	"github.com/sirupsen/logrus"
)

type SpecializationService struct {
	specializationRepo repository.SpecializationRepository
}

func NewSpecializationService(specializationRepo repository.SpecializationRepository) usecase.SpecializationUsecase {
	return &SpecializationService{
		specializationRepo: specializationRepo,
	}
}

func (s *SpecializationService) GetAllSpecializationNames(ctx context.Context) (*dto.SpecializationNamesResponse, error) {
	requestID := utils.GetRequestID(ctx)

	specializations, err := s.specializationRepo.GetAll(ctx)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при получении списка специализаций")
		return nil, err
	}

	// Извлекаем только имена специализаций
	names := make([]string, 0, len(specializations))
	for _, spec := range specializations {
		names = append(names, spec.Name)
	}

	return &dto.SpecializationNamesResponse{
		Names: names,
	}, nil
}

func (s *SpecializationService) GetSpecializationSalaries(ctx context.Context) (*dto.SpecializationSalaryRangesResponse, error) {
	requestID := utils.GetRequestID(ctx)

	salaryRanges, err := s.specializationRepo.GetSpecializationSalaries(ctx)
	if err != nil {
		l.Log.WithFields(logrus.Fields{
			"requestID": requestID,
			"error":     err,
		}).Error("ошибка при получении вилок зарплат по специализациям")
		return nil, err
	}

	response := &dto.SpecializationSalaryRangesResponse{
		Specializations: make([]dto.SpecializationSalaryRange, 0, len(salaryRanges)),
	}

	for _, sr := range salaryRanges {
		response.Specializations = append(response.Specializations, dto.SpecializationSalaryRange{
			ID:        sr.ID,
			Name:      sr.Name,
			MinSalary: sr.MinSalary,
			MaxSalary: sr.MaxSalary,
			AvgSalary: sr.AvgSalary,
		})
	}

	return response, nil
}
