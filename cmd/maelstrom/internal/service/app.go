package service

import "github.com/jonny7/maelstrom/cmd/maelstrom/internal/application"

func NewApplication() application.App {
	return application.New()
}
