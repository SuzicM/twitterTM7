package services

import "registration/twitterTM7/models"

type AuthService interface {
	SignInUser(*models.SignInInput) (*models.DBResponse, error)
}
