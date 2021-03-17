package repository

import (
	"database/sql"
	"github.com/calendar-bot/pkg/types"
)

type UserRepository struct {
	storage *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return UserRepository{storage: db}

}

func (us *UserRepository) CreateUser(user types.User) error {

	_, err := us.storage.Exec(`
			INSERT INTO users(
							  mail_user_id,
							  mail_user_email, 
							  mail_access_token,
							  mail_refresh_token,
							  mail_token_expires_in,
							  telegram_user_id)
			VALUES ($1, $2, $3, $4, $5, $6)`,
		user.UserID,
		user.MailUserEmail,
		user.MailAccessToken,
		user.MailRefreshToken,
		user.MailTokenExpiresIn,
		user.TelegramUserId,
	)

	if err != nil {
		println(err.Error())
		return err
	}

	return nil
}