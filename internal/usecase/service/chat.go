package service

import (
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"context"
)

const ResponseMessage = "Вы откликнулись"

type ChatService struct {
	ApplicantUC usecase.Applicant
	EmployerUC  usecase.Employer
	ResumeUC    usecase.ResumeUsecase
	VacancyUC   usecase.Vacancy
	ChatRepo    repository.ChatRepository
	MessageRepo repository.MessageRepository
}

// TODO исправить возвращаемое значение на интерфейс
func NewChatService(
	applicantUC usecase.Applicant,
	employerUC usecase.Employer,
	resumeUC usecase.ResumeUsecase,
	vacancyUC usecase.Vacancy,
	chatRepository repository.ChatRepository,
	messageRepository repository.MessageRepository,
) *ChatService {
	return &ChatService{
		ApplicantUC: applicantUC,
		EmployerUC:  employerUC,
		ResumeUC:    resumeUC,
		VacancyUC:   vacancyUC,
		ChatRepo:    chatRepository,
		MessageRepo: messageRepository,
	}
}

func (s *ChatService) StartChat(ctx context.Context, vacancyID, resumeID, applicantID, employerID int) (int, error) {
	chat, err := s.ChatRepo.CreateChat(ctx, vacancyID, resumeID, employerID, applicantID)
	if err != nil {
		return -1, err
	}

	_, err = s.MessageRepo.CreateMessage(ctx, chat.ID, chat.ApplicantID, true, ResponseMessage)
	if err != nil {
		return -1, err
	}

	return chat.ID, nil
}

func (s *ChatService) GetChat(ctx context.Context, chatID int, userID int, role string) (*dto.ChatResponse, error) {
	resp, err := s.ChatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	vacancy, err := s.VacancyUC.GetVacancy(ctx, resp.VacancyID, userID, role)
	if err != nil {
		return nil, err
	}

	resume, err := s.ResumeUC.GetByID(ctx, resp.ResumeID)
	if err != nil {
		return nil, err
	}

	chat := &dto.ChatResponse{
		ID: resp.ID,
		Vacancy: &dto.VacancyChatResponse{
			ID:         vacancy.ID,
			EmployerID: vacancy.EmployerID,
			Title:      vacancy.Title,
		},
		Resume: &dto.ResumeChatResponse{
			ID:          resume.ID,
			ApplicantID: resume.ApplicantID,
			Profession:  resume.Profession,
		},
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
	return chat, nil
}

func (s *ChatService) SendMessage(ctx context.Context, chatID, senderID int, role string, payload string) (*dto.MessageResponse, error) {
	fromApplicant := isApplicant(role)
	resp, err := s.MessageRepo.CreateMessage(ctx, chatID, senderID, fromApplicant, payload)
	if err != nil {
		return nil, err
	}

	var avatarPath string
	if fromApplicant {
		applicant, err := s.ApplicantUC.GetUser(ctx, senderID)
		if err != nil {
			return nil, err
		}
		avatarPath = applicant.AvatarPath
	} else {
		employer, err := s.EmployerUC.GetUser(ctx, senderID)
		if err != nil {
			return nil, err
		}
		avatarPath = employer.LogoPath
	}

	message := &dto.MessageResponse{
		ID:            resp.ID,
		ChatID:        resp.ChatID,
		SenderID:      resp.SenderID,
		Avatar:        avatarPath,
		FromApplicant: resp.FromApplicant,
		Payload:       resp.Payload,
		SentAt:        resp.SentAt,
	}

	return message, nil
}

func (s *ChatService) GetUserChats(ctx context.Context, userID int, role string) ([]interface{}, error) {
	fromApplicant := isApplicant(role)
	resp, err := s.ChatRepo.GetForUser(ctx, userID, fromApplicant)
	if err != nil {
		return nil, err
	}

	var chats []interface{}
	for _, chat := range resp {
		vacancy, err := s.VacancyUC.GetVacancy(ctx, chat.VacancyID, userID, role)
		if err != nil {
			return nil, err
		}
		if role == "applicant" {
			employer, err := s.EmployerUC.GetUser(ctx, userID)
			if err != nil {
				return nil, err
			}
			chats = append(chats, &dto.ApplicantChatResponse{
				ID: chat.ID,
				Employer: &dto.ChatShortResponseEmployer{
					ID:          employer.ID,
					CompanyName: employer.CompanyName,
					LogoPath:    employer.LogoPath,
				},
				VacancyTitle: vacancy.Title,
			})
		} else if role == "employer" {
			applicant, err := s.ApplicantUC.GetUser(ctx, userID)
			if err != nil {
				return nil, err
			}
			chats = append(chats, &dto.EmployerChatResponse{
				ID: chat.ID,
				Applicant: &dto.ChatShortResponseApplicant{
					ID:         applicant.ID,
					FirstName:  applicant.FirstName,
					LastName:   applicant.LastName,
					MiddleName: applicant.MiddleName,
					AvatarPath: applicant.AvatarPath,
				},
				VacancyTitle: vacancy.Title,
			})
		}
	}
	return chats, nil
}

func (s *ChatService) GetChatMessages(ctx context.Context, chatID int) ([]*dto.MessageResponse, error) {
	messages, err := s.MessageRepo.GetMessagesForChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	var chatMessages []*dto.MessageResponse
	for _, msg := range messages {
		var avatarPath string
		if msg.FromApplicant {
			applicant, err := s.ApplicantUC.GetUser(ctx, msg.SenderID)
			if err != nil {
				return nil, err
			}
			avatarPath = applicant.AvatarPath
		} else {
			employer, err := s.EmployerUC.GetUser(ctx, msg.SenderID)
			if err != nil {
				return nil, err
			}
			avatarPath = employer.LogoPath
		}

		chatMessages = append(chatMessages, &dto.MessageResponse{
			ID:            msg.ID,
			ChatID:        msg.ChatID,
			SenderID:      msg.SenderID,
			Avatar:        avatarPath,
			FromApplicant: msg.FromApplicant,
			Payload:       msg.Payload,
			SentAt:        msg.SentAt,
		})
	}
	return chatMessages, nil
}

func isApplicant(role string) bool {
	return role == "applicant"
}
