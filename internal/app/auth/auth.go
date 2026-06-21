package auth

import (
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth/handler"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth/model"
	"github.com/Udang-Keju/shrimpy-discord-bot/internal/app/auth/repository"
	"gorm.io/gorm"
)

// Module wraps the layers of the auth business feature.
type Module struct {
	Repo    *repository.UserRepo
	Handler *handler.AuthHandler
}

// Build compiles all layers of the auth feature.
func Build(db *gorm.DB, jwtSecret []byte, tokenEncKey []byte) *Module {
	repo := repository.NewUserRepo(db)
	h := handler.NewAuthHandler(repo, jwtSecret, tokenEncKey)
	return &Module{
		Repo:    repo,
		Handler: h,
	}
}

// Models returns all GORM models utilized by the auth feature.
func (m *Module) Models() []any {
	return []any{&model.User{}}
}

