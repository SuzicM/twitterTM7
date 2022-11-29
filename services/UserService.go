package services

import "registration/twitterTM7/models"

type UserService interface {
	FindUserById(string) (*models.DBResponse, error)
	FindUserByUsername(string) (*models.DBResponse, error)
}
