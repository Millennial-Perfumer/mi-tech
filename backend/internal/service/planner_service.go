package service

import (
	"mi-tech/internal/dto"
	"mi-tech/internal/entity"
	"mi-tech/internal/repository"
)

type PlannerService struct {
	repo repository.PlannerRepository
}

func NewPlannerService(repo repository.PlannerRepository) *PlannerService {
	return &PlannerService{repo: repo}
}

// Boards
func (s *PlannerService) ListBoards() ([]dto.PlannerBoardResponse, error) {
	entities, err := s.repo.ListBoards()
	if err != nil {
		return nil, err
	}
	responses := make([]dto.PlannerBoardResponse, len(entities))
	for i, e := range entities {
		responses[i] = dto.PlannerBoardToResponse(e)
	}
	return responses, nil
}

func (s *PlannerService) GetBoard(id uint) (dto.PlannerBoardResponse, error) {
	e, err := s.repo.GetBoardByID(id)
	if err != nil {
		return dto.PlannerBoardResponse{}, err
	}
	return dto.PlannerBoardToResponse(e), nil
}

// Sprints
func (s *PlannerService) ListSprints(status string) ([]dto.PlannerSprintResponse, error) {
	entities, err := s.repo.ListSprints(status)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.PlannerSprintResponse, len(entities))
	for i, e := range entities {
		responses[i] = dto.PlannerSprintResponse{
			ID:        e.ID,
			Name:      e.Name,
			Goal:      e.Goal,
			StartDate: e.StartDate,
			EndDate:   e.EndDate,
			Status:    e.Status,
		}
	}
	return responses, nil
}

func (s *PlannerService) CreateSprint(req dto.CreateSprintRequest) (dto.PlannerSprintResponse, error) {
	sprint := &entity.PlannerSprint{
		Name:      req.Name,
		Goal:      req.Goal,
		StartDate: &req.StartDate,
		EndDate:   &req.EndDate,
		Status:    "planned", // default status
	}

	if err := s.repo.CreateSprint(sprint); err != nil {
		return dto.PlannerSprintResponse{}, err
	}

	return dto.PlannerSprintResponse{
		ID:        sprint.ID,
		Name:      sprint.Name,
		Goal:      sprint.Goal,
		StartDate: sprint.StartDate,
		EndDate:   sprint.EndDate,
		Status:    sprint.Status,
	}, nil
}

func (s *PlannerService) UpdateSprint(id uint, req dto.UpdateSprintRequest) (dto.PlannerSprintResponse, error) {
	sprint, err := s.repo.GetSprintByID(id)
	if err != nil {
		return dto.PlannerSprintResponse{}, err
	}

	if req.Name != nil {
		sprint.Name = *req.Name
	}
	if req.Goal != nil {
		sprint.Goal = *req.Goal
	}
	if req.StartDate != nil {
		sprint.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		sprint.EndDate = req.EndDate
	}
	if req.Status != nil {
		sprint.Status = *req.Status
	}

	if err := s.repo.UpdateSprint(&sprint); err != nil {
		return dto.PlannerSprintResponse{}, err
	}

	return dto.PlannerSprintResponse{
		ID:        sprint.ID,
		Name:      sprint.Name,
		Goal:      sprint.Goal,
		StartDate: sprint.StartDate,
		EndDate:   sprint.EndDate,
		Status:    sprint.Status,
	}, nil
}

func (s *PlannerService) DeleteSprint(id uint) error {
	return s.repo.DeleteSprint(id)
}

func (s *PlannerService) GetTaskByID(id uint) (dto.PlannerTaskResponse, error) {
	e, err := s.repo.GetTaskByID(id)
	if err != nil {
		return dto.PlannerTaskResponse{}, err
	}
	return dto.PlannerTaskToResponse(e), nil
}

// Tasks
func (s *PlannerService) ListTasks(filter repository.PlannerFilter) ([]dto.PlannerTaskResponse, error) {
	entities, err := s.repo.ListTasks(filter)
	if err != nil {
		return nil, err
	}
	responses := make([]dto.PlannerTaskResponse, len(entities))
	for i, e := range entities {
		responses[i] = dto.PlannerTaskToResponse(e)
	}
	return responses, nil
}

func (s *PlannerService) CreateTask(req dto.CreateTaskRequest) (dto.PlannerTaskResponse, error) {
	task := &entity.PlannerTask{
		BoardID:     req.BoardID,
		ColumnID:    req.ColumnID,
		SprintID:    req.SprintID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Metadata:    req.Metadata,
	}

	if err := s.repo.CreateTask(task); err != nil {
		return dto.PlannerTaskResponse{}, err
	}
	return dto.PlannerTaskToResponse(*task), nil
}

func (s *PlannerService) MoveTask(taskID uint, toColumnID uint, newOrder int) error {
	return s.repo.MoveTask(taskID, toColumnID, newOrder)
}

func (s *PlannerService) UpdateTask(id uint, req dto.UpdateTaskRequest) (dto.PlannerTaskResponse, error) {
	task, err := s.repo.GetTaskByID(id)
	if err != nil {
		return dto.PlannerTaskResponse{}, err
	}

	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.SprintID != nil {
		task.SprintID = req.SprintID
	}
	if req.Metadata != nil {
		task.Metadata = req.Metadata
	}

	if err := s.repo.UpdateTask(&task); err != nil {
		return dto.PlannerTaskResponse{}, err
	}
	return dto.PlannerTaskToResponse(task), nil
}

func (s *PlannerService) DeleteTask(id uint) error {
	return s.repo.DeleteTask(id)
}

func (s *PlannerService) GetAnalytics(sprintID, taskID uint) (dto.PlannerAnalyticsResponse, error) {
	velocity, err := s.repo.GetSprintVelocity(sprintID)
	if err != nil {
		return dto.PlannerAnalyticsResponse{}, err
	}
	leadTime, err := s.repo.GetTaskLeadTime(taskID)
	if err != nil {
		return dto.PlannerAnalyticsResponse{}, err
	}
	return dto.PlannerAnalyticsResponse{
		SprintVelocity: velocity,
		TaskLeadTime:   leadTime,
	}, nil
}
