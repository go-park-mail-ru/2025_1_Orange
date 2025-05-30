package service

import (
	"ResuMatch/internal/entity"
	"ResuMatch/internal/entity/dto"
	"ResuMatch/internal/repository"
	"ResuMatch/internal/usecase"
	"ResuMatch/pkg/sanitizer"
	"context"
	"errors"
	"fmt"
	"strings"
)

const ResponseMessage = "Отклик на вакансию"

type ChatService struct {
	ApplicantUC usecase.Applicant
	EmployerUC  usecase.Employer
	ResumeUC    usecase.ResumeUsecase
	VacancyUC   usecase.Vacancy
	ChatRepo    repository.ChatRepository
	MessageRepo repository.MessageRepository
}

func NewChatService(
	applicantUC usecase.Applicant,
	employerUC usecase.Employer,
	resumeUC usecase.ResumeUsecase,
	vacancyUC usecase.Vacancy,
	chatRepository repository.ChatRepository,
	messageRepository repository.MessageRepository,
) usecase.Chat {
	return &ChatService{
		ApplicantUC: applicantUC,
		EmployerUC:  employerUC,
		ResumeUC:    resumeUC,
		VacancyUC:   vacancyUC,
		ChatRepo:    chatRepository,
		MessageRepo: messageRepository,
	}
}

func (s *ChatService) GetVacancyChat(ctx context.Context, vacancyID, applicantID int, role string) (*dto.ChatResponse, error) {
	chat, err := s.ChatRepo.GetForVacancy(ctx, vacancyID, applicantID)
	if err != nil {
		return nil, err
	}

	if chat != nil {
		return s.GetChat(ctx, chat.ID, applicantID, role)
	}
	info, err := s.ChatRepo.GetVacancyChatInfo(ctx, vacancyID, applicantID)
	if err != nil {
		return nil, err
	}

	if info == nil {
		return nil, entity.NewError(entity.ErrNotFound, errors.New("отклик не найден"))
	}

	chatID, err := s.StartChat(ctx, info.VacancyID, info.ResumeID, applicantID, info.EmployerID)
	if err != nil {
		return nil, err
	}

	return s.GetChat(ctx, chatID, applicantID, role)
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

	if (role == "applicant" && resp.ApplicantID != userID) ||
		(role == "employer" && resp.EmployerID != userID) {
		return nil, entity.NewError(entity.ErrForbidden, errors.New("у вас нет доступа к этому чату"))
	}

	vacancy, err := s.VacancyUC.GetVacancy(ctx, resp.VacancyID, userID, role)
	if err != nil {
		return nil, err
	}

	resume, err := s.ResumeUC.GetByID(ctx, resp.ResumeID)
	if err != nil {
		return nil, err
	}

	applicant, err := s.ApplicantUC.GetUser(ctx, resume.ApplicantID)
	if err != nil {
		return nil, err
	}

	employer, err := s.EmployerUC.GetUser(ctx, vacancy.EmployerID)
	if err != nil {
		return nil, err
	}

	chat := &dto.ChatResponse{
		ID: resp.ID,
		Vacancy: &dto.VacancyChatResponse{
			ID:         vacancy.ID,
			EmployerID: vacancy.EmployerID,
			LogoPath:   employer.LogoPath,
			Title:      vacancy.Title,
		},
		Resume: &dto.ResumeChatResponse{
			ID:          resume.ID,
			ApplicantID: resume.ApplicantID,
			AvatarPath:  applicant.AvatarPath,
			Profession:  resume.Profession,
		},
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
	return chat, nil
}

func (s *ChatService) SendMessage(ctx context.Context, chatID, senderID int, role string, payload string) (*dto.MessageResponse, error) {
	fromApplicant := isApplicant(role)

	chat, err := s.ChatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	sanitizedPayload := sanitizer.StrictPolicy.Sanitize(payload)

	resp, err := s.MessageRepo.CreateMessage(ctx, chatID, senderID, fromApplicant, sanitizedPayload)
	if err != nil {
		return nil, err
	}

	var avatarPath string
	var receiverID int
	if fromApplicant {
		applicant, err := s.ApplicantUC.GetUser(ctx, senderID)
		if err != nil {
			return nil, err
		}
		avatarPath = applicant.AvatarPath
		receiverID = chat.EmployerID
	} else {
		employer, err := s.EmployerUC.GetUser(ctx, senderID)
		if err != nil {
			return nil, err
		}
		avatarPath = employer.LogoPath
		receiverID = chat.ApplicantID
	}

	message := &dto.MessageResponse{
		ID:            resp.ID,
		ChatID:        resp.ChatID,
		SenderID:      resp.SenderID,
		ReceiverID:    receiverID,
		Avatar:        avatarPath,
		FromApplicant: resp.FromApplicant,
		Payload:       resp.Payload,
		SentAt:        resp.SentAt,
	}

	return message, nil
}

func (s *ChatService) GetUserChats(ctx context.Context, userID int, role string) (dto.ChatResponseList, error) {
	fromApplicant := isApplicant(role)
	resp, err := s.ChatRepo.GetForUser(ctx, userID, fromApplicant)
	if err != nil {
		return nil, err
	}

	var chats dto.ChatResponseList
	for _, chat := range resp {
		vacancy, err := s.VacancyUC.GetVacancy(ctx, chat.VacancyID, userID, role)
		if err != nil {
			return nil, err
		}

		var otherUser dto.ChatUserPreview
		switch role {
		case "applicant":
			employer, err := s.EmployerUC.GetUser(ctx, chat.EmployerID)
			if err != nil {
				return nil, err
			}
			otherUser = dto.ChatUserPreview{
				ID:         employer.ID,
				Name:       employer.CompanyName,
				AvatarPath: employer.LogoPath,
			}
		case "employer":
			applicant, err := s.ApplicantUC.GetUser(ctx, chat.ApplicantID)
			if err != nil {
				return nil, err
			}
			fullName := strings.TrimSpace(applicant.LastName + " " + applicant.FirstName + " " + applicant.MiddleName)
			otherUser = dto.ChatUserPreview{
				ID:         applicant.ID,
				Name:       fullName,
				AvatarPath: applicant.AvatarPath,
			}
		default:
			return nil, entity.NewError(entity.ErrBadRequest, fmt.Errorf("неизвестная роль пользователя: %s", role))
		}

		chats = append(chats, &dto.ChatShortResponse{
			ID:           chat.ID,
			VacancyTitle: vacancy.Title,
			User:         otherUser,
		})
	}
	return chats, nil
}

func (s *ChatService) GetChatMessages(ctx context.Context, chatID int) (dto.MessagesResponseList, error) {
	chat, err := s.ChatRepo.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	messages, err := s.MessageRepo.GetMessagesForChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	var chatMessages []*dto.MessageResponse
	for _, msg := range messages {
		var avatarPath string
		var receiverID int
		if msg.FromApplicant {
			applicant, err := s.ApplicantUC.GetUser(ctx, msg.SenderID)
			if err != nil {
				return nil, err
			}
			avatarPath = applicant.AvatarPath
			receiverID = chat.EmployerID
		} else {
			employer, err := s.EmployerUC.GetUser(ctx, msg.SenderID)
			if err != nil {
				return nil, err
			}
			avatarPath = employer.LogoPath
			receiverID = chat.ApplicantID
		}

		chatMessages = append(chatMessages, &dto.MessageResponse{
			ID:            msg.ID,
			ChatID:        msg.ChatID,
			SenderID:      msg.SenderID,
			ReceiverID:    receiverID,
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
