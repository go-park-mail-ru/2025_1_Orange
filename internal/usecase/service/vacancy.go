package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"context"
	"fmt"
)

type VacanciesService struct {
	vacanciesRepository repository.VacancyRepository
}

func NewVacanciesService(vacancyRepository repository.VacancyRepository) *VacanciesService {
	return &VacanciesService{
		vacanciesRepository: vacancyRepository,
	}
}

func (vs *VacanciesService) SearchVacancies(ctx context.Context) ([]*entity.Vacancy, error) {
	vacancies, err := vs.vacanciesRepository.GetAll(ctx)
	if err != nil {
		return nil, entity.NewError(entity.ErrInternal, err)
	}

	return vacancies, nil
}

func (vs *VacanciesService) GetVacanciesByEmpID(ctx context.Context, employerID uint64) ([]*dto.Vacancy, error) {
	vacancies, err := vs.vacanciesRepository.GetVacanciesByEmpID(ctx, employerID)
	if err != nil {
		return nil, entity.NewError(entity.ErrInternal, err)
	}

	result := make([]*dto.Vacancy, 0, len(vacancies))
	for _, vacancy := range vacancies {
		result = append(result, &dto.Vacancy{
			ID:                   vacancy.ID,
			Title:                vacancy.Title,
			IsActive:             vacancy.IsActive,
			EmployerID:           vacancy.EmployerID,
			SpecializationID:     vacancy.SpecializationID,
			WorkFormat:           vacancy.WorkFormat,
			Employment:           vacancy.Employment,
			Schedule:             vacancy.Schedule,
			WorkingHours:         vacancy.WorkingHours,
			SalaryFrom:           vacancy.SalaryFrom,
			SalaryTo:             vacancy.SalaryTo,
			TaxesIncluded:        vacancy.TaxesIncluded,
			Experience:           vacancy.Experience,
			Description:          vacancy.Description,
			Tasks:                vacancy.Tasks,
			Requirements:         vacancy.Requirements,
			OptionalRequirements: vacancy.OptionalRequirements,
		})
	}

	return result, nil
}

func (vs *VacanciesService) CreateVacancy(ctx context.Context, user *dto.UserFromSession) (*entity.Vacancy, error) {
	vacancy := new(entity.Vacancy)
	if user.UserType != entity.UserTypeEmployer {
		return nil, entity.NewError(entity.ErrForbidden, fmt.Errorf("Только работодатели могут создавать вакансии"))
	}

	vacancy.EmployerID = int(user.ID)

	id, err := vs.vacanciesRepository.Create(ctx, vacancy)
	if err != nil {
		return nil, entity.NewError(entity.ErrInternal, err)
	}

	createdVacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, entity.NewError(entity.ErrInternal, err)
	}

	return createdVacancy, nil
}

func (vs *VacanciesService) GetVacancy(ctx context.Context, id int) (int, error) {
	vacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return -1, entity.NewError(entity.ErrNotFound, err)
	}

	return vacancy.ID, nil
}

func (vs *VacanciesService) UpdateVacancy(ctx context.Context, id int, update *entity.Vacancy, user *dto.UserFromSession) (*entity.Vacancy, error) {
	existing, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, entity.NewError(entity.ErrNotFound, err)
	}

	if existing.EmployerID != user.ID {
		return nil, entity.NewError(entity.ErrForbidden, fmt.Errorf("Пользователь не владелец вакансии"))
	}

	err = vs.vacanciesRepository.Update(ctx, update)
	if err != nil {
		return nil, entity.NewError(entity.ErrInternal, err)
	}

	return update, nil
}

func (vs *VacanciesService) DeleteVacancy(ctx context.Context, id int, user *dto.UserFromSession) error {
	vacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return entity.NewError(entity.ErrNotFound, err)
	}

	if vacancy.EmployerID != user.ID {
		return entity.NewError(entity.ErrForbidden, fmt.Errorf("Пользователь не владелец вакансии"))
	}

	if err := vs.vacanciesRepository.Delete(ctx, user.ID, vacancy.ID); err != nil {
		return entity.NewError(entity.ErrInternal, err)
	}

	return nil
}
