package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
<<<<<<< HEAD
<<<<<<< HEAD
	"ResuMatch/internal/usecase"
	"ResuMatch/internal/utils"
<<<<<<< HEAD
=======
	"ResuMatch/internal/usecase"
>>>>>>> c773955 (Made vacansies usecases and handlers)
=======
>>>>>>> a6396a4 (Fix mistakes)
	l "ResuMatch/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
<<<<<<< HEAD
=======
	"context"
	"fmt"
>>>>>>> 8cdc676 (Add vacancy usecases and handlers)
)

type VacanciesService struct {
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
	vacanciesRepository repository.VacancyRepository
<<<<<<< HEAD
	//skillRepository          repository.SkillRepository
=======
	vacanciesRepository      repository.VacancyRepository
>>>>>>> 71cf6a4 (Made vacansies usecases and handlers)
=======
	vacanciesRepository      repository.VacancyRepository
=======
	vacanciesRepository repository.VacancyRepository
	//skillRepository          repository.SkillRepository
>>>>>>> a6396a4 (Fix mistakes)
>>>>>>> d7704b3 (Fix mistakes)
=======
	vacanciesRepository      repository.VacancyRepository
>>>>>>> bf6489c (Fix mistakes)
	cityRepository           repository.CityRepository
	applicantRepository      repository.ApplicantRepository
	specializationRepository repository.SpecializationRepository
}

func NewVacanciesService(vacancyRepo repository.VacancyRepository,
<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
	//skillRepo repository.SkillRepository,
=======
>>>>>>> e918c1a (Fix issues with conflicts)
=======
>>>>>>> d7704b3 (Fix mistakes)
=======
>>>>>>> bf6489c (Fix mistakes)
	cityRepo repository.CityRepository,
	applicantRepo repository.ApplicantRepository,
	specializationRepo repository.SpecializationRepository,
) usecase.Vacancy {
	return &VacanciesService{
		vacanciesRepository:      vacancyRepo,
		cityRepository:           cityRepo,
		applicantRepository:      applicantRepo,
		specializationRepository: specializationRepo,
	}
}

func (s *VacanciesService) CreateVacancy(ctx context.Context, request *dto.VacancyCreate) (*dto.VacancyResponse, error) {

	requestID := utils.GetRequestID(ctx)

	employerID, ok := ctx.Value("employerID").(int)
	if !ok {
		return nil, entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("не удалось получить ID работодателя из контекста"),
		)
	}

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"employerID": employerID,
	}).Info("Создание вакансии")

	vacancy := &entity.Vacancy{
		Title:                request.Title,
		IsActive:             true,
		EmployerID:           employerID,
		SpecializationID:     request.SpecializationID,
		WorkFormat:           request.WorkFormat,
		Employment:           request.Employment,
		Schedule:             request.Schedule,
		WorkingHours:         request.WorkingHours,
		SalaryFrom:           request.SalaryFrom,
		SalaryTo:             request.SalaryTo,
		TaxesIncluded:        request.TaxesIncluded,
		Experience:           request.Experience,
		Description:          request.Description,
		Tasks:                request.Tasks,
		Requirements:         request.Requirements,
		OptionalRequirements: request.OptionalRequirements,
	}

	if err := vacancy.Validate(); err != nil {
		return nil, err
	}

	createdVacancy, err := s.vacanciesRepository.Create(ctx, vacancy)
	if err != nil {
		return nil, err
	}

	if len(request.Skills) > 0 {
		skillIDs, err := s.vacanciesRepository.FindSkillIDsByNames(ctx, request.Skills)
		if err != nil {
			return nil, err
		}

		if len(skillIDs) > 0 {
			if err := s.vacanciesRepository.AddSkills(ctx, createdVacancy.ID, skillIDs); err != nil {
				return nil, err
			}
		}
	}

	if len(request.City) > 0 {
		cityIDs, err := s.vacanciesRepository.FindCityIDsByNames(ctx, request.City)
		if err != nil {
			return nil, err
		}

		if len(cityIDs) > 0 {
			if err := s.vacanciesRepository.AddCity(ctx, createdVacancy.ID, cityIDs); err != nil {
				return nil, err
			}
		}
	}

	skills, err := s.vacanciesRepository.GetSkillsByVacancyID(ctx, createdVacancy.ID)
	if err != nil {
		return nil, err
	}

	cities, err := s.vacanciesRepository.GetCityByVacancyID(ctx, createdVacancy.ID)
	if err != nil {
		return nil, err
	}

	var specializationName string
	if createdVacancy.SpecializationID != 0 {
		specialization, err := s.specializationRepository.GetByID(ctx, createdVacancy.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}

	experienceStr := fmt.Sprintf("%d+ лет", createdVacancy.Experience)

	response := &dto.VacancyResponse{
		ID:                   createdVacancy.ID,
		EmployerID:           createdVacancy.EmployerID,
		Title:                createdVacancy.Title,
		SpecializationID:     specializationName,
		WorkFormat:           createdVacancy.WorkFormat,
		Employment:           createdVacancy.Employment,
		Schedule:             createdVacancy.Schedule,
		WorkingHours:         createdVacancy.WorkingHours,
		SalaryFrom:           createdVacancy.SalaryFrom,
		SalaryTo:             createdVacancy.SalaryTo,
		TaxesIncluded:        createdVacancy.TaxesIncluded,
		Experience:           experienceStr,
		Description:          createdVacancy.Description,
		Tasks:                createdVacancy.Tasks,
		Requirements:         createdVacancy.Requirements,
		OptionalRequirements: createdVacancy.OptionalRequirements,
		CreatedAt:            createdVacancy.CreatedAt,
		UpdatedAt:            createdVacancy.UpdatedAt,
	}

	response.Skills = make([]string, 0, len(skills))
	for _, skill := range skills {
		response.Skills = append(response.Skills, skill.Name)
	}

	response.City = make([]string, 0, len(cities))
	for _, city := range cities {
		response.City = append(response.City, city.Name)
	}

	return response, nil
}

func (vs *VacanciesService) GetVacancy(ctx context.Context, id int) (*dto.VacancyResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"vacancyID": id,
	}).Info("Получение вакансии по ID")

	vacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var specializationName string
	if vacancy.SpecializationID != 0 {
		specialization, err := vs.specializationRepository.GetByID(ctx, vacancy.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}

	skills, err := vs.vacanciesRepository.GetSkillsByVacancyID(ctx, vacancy.ID)
	if err != nil {
		return nil, err
	}

	cities, err := vs.vacanciesRepository.GetCityByVacancyID(ctx, vacancy.ID)
	if err != nil {
		return nil, err
	}

	experienceStr := fmt.Sprintf("%d+ лет", vacancy.Experience)

	response := &dto.VacancyResponse{
		ID:                   vacancy.ID,
		EmployerID:           vacancy.EmployerID,
		Title:                vacancy.Title,
		SpecializationID:     specializationName,
		WorkFormat:           vacancy.WorkFormat,
		Employment:           vacancy.Employment,
		Schedule:             vacancy.Schedule,
		WorkingHours:         vacancy.WorkingHours,
		SalaryFrom:           vacancy.SalaryFrom,
		SalaryTo:             vacancy.SalaryTo,
		TaxesIncluded:        vacancy.TaxesIncluded,
		Experience:           experienceStr,
		Description:          vacancy.Description,
		Tasks:                vacancy.Tasks,
		Requirements:         vacancy.Requirements,
		OptionalRequirements: vacancy.OptionalRequirements,
		CreatedAt:            vacancy.CreatedAt,
		UpdatedAt:            vacancy.UpdatedAt,
		Skills:               make([]string, 0, len(skills)),
		City:                 make([]string, 0, len(cities)),
	}

	for _, skill := range skills {
		response.Skills = append(response.Skills, skill.Name)
	}

	for _, city := range cities {
		response.City = append(response.City, city.Name)
	}

	return response, nil
}

func (vs *VacanciesService) UpdateVacancy(ctx context.Context, id int, request *dto.VacancyUpdate) (*dto.VacancyResponse, error) {

	requestID := utils.GetRequestID(ctx)

	employerID, ok := ctx.Value("employerID").(int)
	if !ok {
		return nil, entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("не удалось получить ID работодателя из контекста"),
		)
	}

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"vacancyID":  id,
		"employerID": employerID,
	}).Info("Обновление вакансии")

	existingVacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingVacancy.EmployerID != employerID {
		return nil, entity.NewError(
			entity.ErrForbidden,
			fmt.Errorf("вакансия с id=%d не принадлежит работодателю с id=%d", id, employerID),
		)
	}

	vacancy := &entity.Vacancy{
		ID:                   id,
		EmployerID:           employerID,
		Title:                request.Title,
		SpecializationID:     request.SpecializationID,
		WorkFormat:           request.WorkFormat,
		Employment:           request.Employment,
		Schedule:             request.Schedule,
		WorkingHours:         request.WorkingHours,
		SalaryFrom:           request.SalaryFrom,
		SalaryTo:             request.SalaryTo,
		TaxesIncluded:        request.TaxesIncluded,
		Experience:           request.Experience,
		Description:          request.Description,
		Tasks:                request.Tasks,
		Requirements:         request.Requirements,
		OptionalRequirements: request.OptionalRequirements,
	}

	if err := vacancy.Validate(); err != nil {
		return nil, err
	}

	updatedVacancy, err := vs.vacanciesRepository.Update(ctx, vacancy)
	if err != nil {
		return nil, err
	}

	if err := vs.vacanciesRepository.DeleteSkills(ctx, id); err != nil {
		return nil, err
	}
	if len(request.Skills) > 0 {
		skillIDs, err := vs.vacanciesRepository.FindSkillIDsByNames(ctx, request.Skills)
		if err != nil {
			return nil, err
		}
		if len(skillIDs) > 0 {
			if err := vs.vacanciesRepository.AddSkills(ctx, id, skillIDs); err != nil {
				return nil, err
			}
		}
	}

	if err := vs.vacanciesRepository.DeleteCity(ctx, id); err != nil {
		return nil, err
	}
	if len(request.City) > 0 {
		cityIDs, err := vs.vacanciesRepository.FindCityIDsByNames(ctx, request.City)
		if err != nil {
			return nil, err
		}
		if len(cityIDs) > 0 {
			if err := vs.vacanciesRepository.AddCity(ctx, id, cityIDs); err != nil {
				return nil, err
			}
		}
	}
	var specializationName string
	if updatedVacancy.SpecializationID != 0 {
		specialization, err := vs.specializationRepository.GetByID(ctx, updatedVacancy.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}
	skills, err := vs.vacanciesRepository.GetSkillsByVacancyID(ctx, id)
	if err != nil {
		return nil, err
	}
	cities, err := vs.vacanciesRepository.GetCityByVacancyID(ctx, id)
	if err != nil {
		return nil, err
	}
	experienceStr := fmt.Sprintf("%d+ лет", updatedVacancy.Experience)
	response := &dto.VacancyResponse{
		ID:                   updatedVacancy.ID,
		EmployerID:           updatedVacancy.EmployerID,
		Title:                updatedVacancy.Title,
		SpecializationID:     specializationName,
		WorkFormat:           updatedVacancy.WorkFormat,
		Employment:           updatedVacancy.Employment,
		Schedule:             updatedVacancy.Schedule,
		WorkingHours:         updatedVacancy.WorkingHours,
		SalaryFrom:           updatedVacancy.SalaryFrom,
		SalaryTo:             updatedVacancy.SalaryTo,
		TaxesIncluded:        updatedVacancy.TaxesIncluded,
		Experience:           experienceStr,
		Description:          updatedVacancy.Description,
		Tasks:                updatedVacancy.Tasks,
		Requirements:         updatedVacancy.Requirements,
		OptionalRequirements: updatedVacancy.OptionalRequirements,
		CreatedAt:            updatedVacancy.CreatedAt,
		UpdatedAt:            updatedVacancy.UpdatedAt,
		Skills:               make([]string, 0, len(skills)),
		City:                 make([]string, 0, len(cities)),
	}

	for _, skill := range skills {
		response.Skills = append(response.Skills, skill.Name)
	}

	for _, city := range cities {
		response.City = append(response.City, city.Name)
	}
	return response, nil
}

func (vs *VacanciesService) DeleteVacancy(ctx context.Context, id int, employerID int) (*dto.DeleteVacancy, error) {

	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"vacancyID":  id,
		"employerID": employerID,
	}).Info("Удаление вакансии")

	existingVacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingVacancy.EmployerID != employerID {
		return nil, entity.NewError(
			entity.ErrForbidden,
			fmt.Errorf("вакансия с id=%d не принадлежит работодателю с id=%d", id, employerID),
		)
	}

	if err := vs.vacanciesRepository.DeleteSkills(ctx, id); err != nil {
		return nil, err
	}

	if err := vs.vacanciesRepository.DeleteCity(ctx, id); err != nil {
		return nil, err
	}

	if err := vs.vacanciesRepository.Delete(ctx, id); err != nil {
		return nil, err
	}

	return &dto.DeleteVacancy{
		Success: true,
		Message: fmt.Sprintf("Вакансия с id=%d успешно удалена", id),
	}, nil
}

func (s *VacanciesService) GetAll(ctx context.Context) ([]dto.VacancyShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("Получение списка всех резюме")

	vacancies, err := s.vacanciesRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]dto.VacancyShortResponse, 0, len(vacancies))
	for _, vacancy := range vacancies {
		var specializationName string
		if vacancy.SpecializationID != 0 {
			specialization, err := s.specializationRepository.GetByID(ctx, vacancy.SpecializationID)
			if err != nil {
				l.Log.WithFields(logrus.Fields{
					"requestID":        requestID,
					"vacancyID":        vacancy.ID,
					"specializationID": vacancy.SpecializationID,
					"error":            err,
				}).Error("ошибка при получении специализации")
				continue
			}
			specializationName = specialization.Name
		}

		shortVacancy := dto.VacancyShortResponse{
			ID:             vacancy.ID,
			Title:          vacancy.Title,
			EmployerID:     vacancy.EmployerID,
			Specialization: specializationName,
			WorkFormat:     vacancy.WorkFormat,
			Employment:     vacancy.Employment,
			WorkingHours:   vacancy.WorkingHours,
			SalaryFrom:     vacancy.SalaryFrom,
			SalaryTo:       vacancy.SalaryTo,
			CreatedAt:      vacancy.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      vacancy.UpdatedAt.Format(time.RFC3339),
		}

		response = append(response, shortVacancy)
	}

	return response, nil
<<<<<<< HEAD
<<<<<<< HEAD

// func (s *VacanciesService) ApplyToVacancy(ctx context.Context, vacancyID, applicantID, resumeID int) error {
// 	if _, err := s.vacanciesRepository.GetByID(ctx, vacancyID); err != nil {
// 		return err
// 	}
// 	// Проверяем, не откликался ли уже
// 	hasResponded, err := s.vacanciesRepository.ResponseExists(ctx, vacancyID, applicantID)
// 	if err != nil {
// 		return err
// 	}
// 	if hasResponded {
// 		return entity.NewError(entity.ErrAlreadyExists,
// 			fmt.Errorf("you have already applied to this vacancy"))
// 	}

<<<<<<< HEAD
	return s.vacanciesRepository.CreateResponse(ctx, vacancyID, applicantID, resumeID)
=======
=======
)

type VacanciesService struct {
	vacanciesRepository repository.VacancyRepository
	//skillRepository          repository.SkillRepository
	cityRepository           repository.CityRepository
	applicantRepository      repository.ApplicantRepository
	specializationRepository repository.SpecializationRepository
>>>>>>> c773955 (Made vacansies usecases and handlers)
}

func NewVacanciesService(vacancyRepo repository.VacancyRepository,
	//skillRepo repository.SkillRepository,
	//cityRepo repository.CityRepository,
	applicantRepo repository.ApplicantRepository,
	specializationRepo repository.SpecializationRepository,
) usecase.Vacancy {
	return &VacanciesService{
		vacanciesRepository: vacancyRepo,
		//skillRepository:          skillRepo,
		cityRepository:           cityRepo,
		applicantRepository:      applicantRepo,
		specializationRepository: specializationRepo,
	}
}

func (s *VacanciesService) CreateVacancy(ctx context.Context, request *dto.VacancyCreate) (*dto.VacancyResponse, error) {
	requestID := utils.GetRequestID(ctx)

	employerID, ok := ctx.Value("employerID").(int)
	if !ok {
		return nil, entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("не удалось получить ID работодателя из контекста"),
		)
	}

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"employerID": employerID,
	}).Info("Создание вакансии")

	vacancy := &entity.Vacancy{
		Title:                request.Title,
		IsActive:             true,
		EmployerID:           employerID,
		SpecializationID:     request.SpecializationID,
		WorkFormat:           request.WorkFormat,
		Employment:           request.Employment,
		Schedule:             request.Schedule,
		WorkingHours:         request.WorkingHours,
		SalaryFrom:           request.SalaryFrom,
		SalaryTo:             request.SalaryTo,
		TaxesIncluded:        request.TaxesIncluded,
		Experience:           request.Experience,
		Description:          request.Description,
		Tasks:                request.Tasks,
		Requirements:         request.Requirements,
		OptionalRequirements: request.OptionalRequirements,
	}

	if err := vacancy.Validate(); err != nil {
		return nil, err
	}

	createdVacancy, err := s.vacanciesRepository.Create(ctx, vacancy)
	if err != nil {
		return nil, err
	}

	if len(request.Skills) > 0 {
		skillIDs, err := s.vacanciesRepository.FindSkillIDsByNames(ctx, request.Skills)
		if err != nil {
			return nil, err
		}

		if len(skillIDs) > 0 {
			if err := s.vacanciesRepository.AddSkills(ctx, createdVacancy.ID, skillIDs); err != nil {
				return nil, err
			}
		}
	}

	if len(request.City) > 0 {
		cityIDs, err := s.vacanciesRepository.FindCityIDsByNames(ctx, request.City)
		if err != nil {
			return nil, err
		}

		if len(cityIDs) > 0 {
			if err := s.vacanciesRepository.AddCity(ctx, createdVacancy.ID, cityIDs); err != nil {
				return nil, err
			}
		}
	}

	skills, err := s.vacanciesRepository.GetSkillsByVacancyID(ctx, createdVacancy.ID)
	if err != nil {
		return nil, err
	}

	cities, err := s.vacanciesRepository.GetCityByVacancyID(ctx, createdVacancy.ID)
	if err != nil {
		return nil, err
	}

	var specializationName string
	if createdVacancy.SpecializationID != 0 {
		specialization, err := s.specializationRepository.GetByID(ctx, createdVacancy.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}

	experienceStr := fmt.Sprintf("%d+ лет", createdVacancy.Experience)

	response := &dto.VacancyResponse{
		ID:                   createdVacancy.ID,
		EmployerID:           createdVacancy.EmployerID,
		Title:                createdVacancy.Title,
		SpecializationID:     specializationName,
		WorkFormat:           createdVacancy.WorkFormat,
		Employment:           createdVacancy.Employment,
		Schedule:             createdVacancy.Schedule,
		WorkingHours:         createdVacancy.WorkingHours,
		SalaryFrom:           createdVacancy.SalaryFrom,
		SalaryTo:             createdVacancy.SalaryTo,
		TaxesIncluded:        createdVacancy.TaxesIncluded,
		Experience:           experienceStr,
		Description:          createdVacancy.Description,
		Tasks:                createdVacancy.Tasks,
		Requirements:         createdVacancy.Requirements,
		OptionalRequirements: createdVacancy.OptionalRequirements,
		CreatedAt:            createdVacancy.CreatedAt,
		UpdatedAt:            createdVacancy.UpdatedAt,
	}

	response.Skills = make([]string, 0, len(skills))
	for _, skill := range skills {
		response.Skills = append(response.Skills, skill.Name)
	}

	response.City = make([]string, 0, len(cities))
	for _, city := range cities {
		response.City = append(response.City, city.Name)
	}

	return response, nil
}

func (vs *VacanciesService) GetVacancy(ctx context.Context, id int) (*dto.VacancyResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
		"vacancyID": id,
	}).Info("Получение вакансии по ID")

	vacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var specializationName string
	if vacancy.SpecializationID != 0 {
		specialization, err := vs.specializationRepository.GetByID(ctx, vacancy.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}

	skills, err := vs.vacanciesRepository.GetSkillsByVacancyID(ctx, vacancy.ID)
	if err != nil {
		return nil, err
	}

	cities, err := vs.vacanciesRepository.GetCityByVacancyID(ctx, vacancy.ID)
	if err != nil {
		return nil, err
	}

	experienceStr := fmt.Sprintf("%d+ лет", vacancy.Experience)

	response := &dto.VacancyResponse{
		ID:                   vacancy.ID,
		EmployerID:           vacancy.EmployerID,
		Title:                vacancy.Title,
		SpecializationID:     specializationName,
		WorkFormat:           vacancy.WorkFormat,
		Employment:           vacancy.Employment,
		Schedule:             vacancy.Schedule,
		WorkingHours:         vacancy.WorkingHours,
		SalaryFrom:           vacancy.SalaryFrom,
		SalaryTo:             vacancy.SalaryTo,
		TaxesIncluded:        vacancy.TaxesIncluded,
		Experience:           experienceStr,
		Description:          vacancy.Description,
		Tasks:                vacancy.Tasks,
		Requirements:         vacancy.Requirements,
		OptionalRequirements: vacancy.OptionalRequirements,
		CreatedAt:            vacancy.CreatedAt,
		UpdatedAt:            vacancy.UpdatedAt,
		Skills:               make([]string, 0, len(skills)),
		City:                 make([]string, 0, len(cities)),
	}

	for _, skill := range skills {
		response.Skills = append(response.Skills, skill.Name)
	}

	for _, city := range cities {
		response.City = append(response.City, city.Name)
	}

	return response, nil
}

func (vs *VacanciesService) UpdateVacancy(ctx context.Context, id int, request *dto.VacancyUpdate) (*dto.VacancyResponse, error) {
	requestID := utils.GetRequestID(ctx)

	employerID, ok := ctx.Value("employerID").(int)
	if !ok {
		return nil, entity.NewError(
			entity.ErrBadRequest,
			fmt.Errorf("не удалось получить ID работодателя из контекста"),
		)
	}

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"vacancyID":  id,
		"employerID": employerID,
	}).Info("Обновление вакансии")

	existingVacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingVacancy.EmployerID != employerID {
		return nil, entity.NewError(
			entity.ErrForbidden,
			fmt.Errorf("вакансия с id=%d не принадлежит работодателю с id=%d", id, employerID),
		)
	}

	vacancy := &entity.Vacancy{
		ID:                   id,
		EmployerID:           employerID,
		Title:                request.Title,
		SpecializationID:     request.SpecializationID,
		WorkFormat:           request.WorkFormat,
		Employment:           request.Employment,
		Schedule:             request.Schedule,
		WorkingHours:         request.WorkingHours,
		SalaryFrom:           request.SalaryFrom,
		SalaryTo:             request.SalaryTo,
		TaxesIncluded:        request.TaxesIncluded,
		Experience:           request.Experience,
		Description:          request.Description,
		Tasks:                request.Tasks,
		Requirements:         request.Requirements,
		OptionalRequirements: request.OptionalRequirements,
	}

	if err := vacancy.Validate(); err != nil {
		return nil, err
	}

	updatedVacancy, err := vs.vacanciesRepository.Update(ctx, vacancy)
	if err != nil {
		return nil, err
	}

	if err := vs.vacanciesRepository.DeleteSkills(ctx, id); err != nil {
		return nil, err
	}
	if len(request.Skills) > 0 {
		skillIDs, err := vs.vacanciesRepository.FindSkillIDsByNames(ctx, request.Skills)
		if err != nil {
			return nil, err
		}
		if len(skillIDs) > 0 {
			if err := vs.vacanciesRepository.AddSkills(ctx, id, skillIDs); err != nil {
				return nil, err
			}
		}
	}

	if err := vs.vacanciesRepository.DeleteCity(ctx, id); err != nil {
		return nil, err
	}
	if len(request.City) > 0 {
		cityIDs, err := vs.vacanciesRepository.FindCityIDsByNames(ctx, request.City)
		if err != nil {
			return nil, err
		}
		if len(cityIDs) > 0 {
			if err := vs.vacanciesRepository.AddCity(ctx, id, cityIDs); err != nil {
				return nil, err
			}
		}
	}
	var specializationName string
	if updatedVacancy.SpecializationID != 0 {
		specialization, err := vs.specializationRepository.GetByID(ctx, updatedVacancy.SpecializationID)
		if err != nil {
			return nil, err
		}
		specializationName = specialization.Name
	}
	skills, err := vs.vacanciesRepository.GetSkillsByVacancyID(ctx, id)
	if err != nil {
		return nil, err
	}
	cities, err := vs.vacanciesRepository.GetCityByVacancyID(ctx, id)
	if err != nil {
		return nil, err
	}
	experienceStr := fmt.Sprintf("%d+ лет", updatedVacancy.Experience)
	response := &dto.VacancyResponse{
		ID:                   updatedVacancy.ID,
		EmployerID:           updatedVacancy.EmployerID,
		Title:                updatedVacancy.Title,
		SpecializationID:     specializationName,
		WorkFormat:           updatedVacancy.WorkFormat,
		Employment:           updatedVacancy.Employment,
		Schedule:             updatedVacancy.Schedule,
		WorkingHours:         updatedVacancy.WorkingHours,
		SalaryFrom:           updatedVacancy.SalaryFrom,
		SalaryTo:             updatedVacancy.SalaryTo,
		TaxesIncluded:        updatedVacancy.TaxesIncluded,
		Experience:           experienceStr,
		Description:          updatedVacancy.Description,
		Tasks:                updatedVacancy.Tasks,
		Requirements:         updatedVacancy.Requirements,
		OptionalRequirements: updatedVacancy.OptionalRequirements,
		CreatedAt:            updatedVacancy.CreatedAt,
		UpdatedAt:            updatedVacancy.UpdatedAt,
		Skills:               make([]string, 0, len(skills)),
		City:                 make([]string, 0, len(cities)),
	}

	for _, skill := range skills {
		response.Skills = append(response.Skills, skill.Name)
	}

	for _, city := range cities {
		response.City = append(response.City, city.Name)
	}
	return response, nil
}

func (vs *VacanciesService) DeleteVacancy(ctx context.Context, id int, employerID int) (*dto.DeleteVacancy, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID":  requestID,
		"vacancyID":  id,
		"employerID": employerID,
	}).Info("Удаление вакансии")

	existingVacancy, err := vs.vacanciesRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existingVacancy.EmployerID != employerID {
		return nil, entity.NewError(
			entity.ErrForbidden,
			fmt.Errorf("вакансия с id=%d не принадлежит работодателю с id=%d", id, employerID),
		)
	}

	if err := vs.vacanciesRepository.DeleteSkills(ctx, id); err != nil {
		return nil, err
	}

	if err := vs.vacanciesRepository.DeleteCity(ctx, id); err != nil {
		return nil, err
	}

	if err := vs.vacanciesRepository.Delete(ctx, id); err != nil {
		return nil, err
	}

	return &dto.DeleteVacancy{
		Success: true,
		Message: fmt.Sprintf("Вакансия с id=%d успешно удалена", id),
	}, nil
}

func (s *VacanciesService) GetAll(ctx context.Context) ([]dto.VacancyShortResponse, error) {
	requestID := utils.GetRequestID(ctx)

	l.Log.WithFields(logrus.Fields{
		"requestID": requestID,
	}).Info("Получение списка всех резюме")

	vacancies, err := s.vacanciesRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]dto.VacancyShortResponse, 0, len(vacancies))
	for _, vacancy := range vacancies {
		var specializationName string
		if vacancy.SpecializationID != 0 {
			specialization, err := s.specializationRepository.GetByID(ctx, vacancy.SpecializationID)
			if err != nil {
				l.Log.WithFields(logrus.Fields{
					"requestID":        requestID,
					"vacancyID":        vacancy.ID,
					"specializationID": vacancy.SpecializationID,
					"error":            err,
				}).Error("ошибка при получении специализации")
				continue
			}
			specializationName = specialization.Name
		}

		shortVacancy := dto.VacancyShortResponse{
			ID:             vacancy.ID,
			Title:          vacancy.Title,
			EmployerID:     vacancy.EmployerID,
			Specialization: specializationName,
			WorkFormat:     vacancy.WorkFormat,
			Employment:     vacancy.Employment,
			WorkingHours:   vacancy.WorkingHours,
			SalaryFrom:     vacancy.SalaryFrom,
			SalaryTo:       vacancy.SalaryTo,
			CreatedAt:      vacancy.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      vacancy.UpdatedAt.Format(time.RFC3339),
		}

		response = append(response, shortVacancy)
	}

<<<<<<< HEAD
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

func (s *VacanciesService) ApplyToVacancy(ctx context.Context, vacancyID, applicantID, resumeID int) error {
	if _, err := s.vacanciesRepository.GetByID(ctx, vacancyID); err != nil {
		return err
	}
	// Проверяем, не откликался ли уже
	hasResponded, err := s.vacanciesRepository.ResponseExists(ctx, vacancyID, applicantID)
	if err != nil {
		return err
	}
	if hasResponded {
		return entity.NewError(entity.ErrAlreadyExists,
			fmt.Errorf("you have already applied to this vacancy"))
	}

	return s.vacanciesRepository.CreateResponse(ctx, vacancyID, applicantID, resumeID)
}
=======
// 	return s.vacanciesRepository.CreateResponse(ctx, vacancyID, applicantID, resumeID)
// }
<<<<<<< HEAD
>>>>>>> 71cf6a4 (Made vacansies usecases and handlers)
=======
=======
=======
>>>>>>> bf6489c (Fix mistakes)
}
func (s *VacanciesService) ApplyToVacancy(ctx context.Context, vacancyID, applicantID, resumeID int) error {
	if _, err := s.vacanciesRepository.GetByID(ctx, vacancyID); err != nil {
		return err
	}
	// Проверяем, не откликался ли уже
	hasResponded, err := s.vacanciesRepository.ResponseExists(ctx, vacancyID, applicantID)
	if err != nil {
		return err
	}
	if hasResponded {
		return entity.NewError(entity.ErrAlreadyExists,
			fmt.Errorf("you have already applied to this vacancy"))
	}

	return s.vacanciesRepository.CreateResponse(ctx, vacancyID, applicantID, resumeID)
}
<<<<<<< HEAD
>>>>>>> a6396a4 (Fix mistakes)
>>>>>>> d7704b3 (Fix mistakes)
=======
>>>>>>> bf6489c (Fix mistakes)
