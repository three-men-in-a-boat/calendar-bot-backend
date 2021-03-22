package usecase

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/calendar-bot/cmd/config"
	"github.com/calendar-bot/pkg/types"
	"github.com/calendar-bot/pkg/users/repository"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"time"
)

const (
	nonceBitsLength = 256
	nonceBase       = 16
)

const stateExpire = 5 * time.Minute

type UserUseCase struct {
	userRepository repository.UserRepository
	conf           *config.App
}

func NewUserUseCase(userRepo repository.UserRepository, conf *config.App) UserUseCase {
	return UserUseCase{
		userRepository: userRepo,
		conf:           conf,
	}
}

type tokenGetResp struct {
	ExpiresIn    int64  `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

func (uuc *UserUseCase) GenOauthLinkForTelegramID(telegramID int64) (string, error) {
	bigInt, err := rand.Prime(rand.Reader, nonceBitsLength)
	if err != nil {
		return "", errors.WithStack(err)
	}

	state := bigInt.Text(nonceBase)

	if err := uuc.userRepository.SetTelegramIDByState(state, telegramID, stateExpire); err != nil {
		return "", errors.Wrapf(err, "cannot insert telegramID=%d in redis, state=%s", telegramID, state)
	}

	link := uuc.generateOAuthLink(state)

	return link, nil

}

func (uuc *UserUseCase) GetTelegramIDByState(state string) (int64, error) {
	return uuc.userRepository.GetTelegramIDByState(state)
}

func (uuc *UserUseCase) TelegramCreateUser(tgUserID int64, mailAuthCode string) (err error) {

	response, err := http.PostForm(
		"https://oauth.mail.ru/token",
		url.Values{
			"code":          []string{mailAuthCode},
			"grant_type":    []string{"authorization_code"},
			"redirect_uri":  []string{uuc.conf.OAuth.RedirectURI},
			"client_id":     []string{uuc.conf.OAuth.ClientID},
			"client_secret": []string{uuc.conf.OAuth.ClientSecret},
		},
	)

	if err != nil {
		return errors.Wrapf(err, "cannot get token for")
	}

	defer func() {
		if newErr := response.Body.Close(); newErr != nil {
			if err != nil {
				err = errors.Wrap(err, newErr.Error())
			} else {
				err = errors.WithStack(newErr)
			}
		}
	}()

	if response.StatusCode != http.StatusOK {
		return errors.New("something wrong with token recv: " + response.Status)
	}

	var tokenResp tokenGetResp

	if newErr := json.NewDecoder(response.Body).Decode(&tokenResp); err != nil {
		return newErr
	}

	accessToken := tokenResp.AccessToken

	response, err = http.Get(fmt.Sprintf("https://oauth.mail.ru/userinfo?access_token=%s", accessToken))

	if err != nil {
		return errors.Wrapf(err, "cannot get accessToken for tgUserID=%d", tgUserID)
	}

	defer func() {
		if newErr := response.Body.Close(); newErr != nil {
			if err != nil {
				err = errors.Wrap(err, newErr.Error())
			} else {
				err = errors.WithStack(newErr)
			}
		}
	}()

	if response.StatusCode != http.StatusOK {
		return errors.New("something wrong with user info recv: " + response.Status)
	}

	var userInfo map[string]string
	if newErr := json.NewDecoder(response.Body).Decode(&userInfo); err != nil {
		return newErr
	}

	if err, ok := userInfo["error"]; ok {
		return errors.New(err)
	}

	email := userInfo["email"]
	userID := userInfo["id"]

	user := types.User{
		ID:                 0,
		UserID:             userID,
		MailUserEmail:      email,
		MailAccessToken:    tokenResp.AccessToken,
		MailRefreshToken:   tokenResp.RefreshToken,
		MailTokenExpiresIn: time.Now().Add(time.Second * time.Duration(tokenResp.ExpiresIn)),
		TelegramUserId:     tgUserID,
		CreatedAt:          time.Time{},
	}

	return uuc.userRepository.CreateUser(user)
}

func (uuc *UserUseCase) generateOAuthLink(state string) string {
	return fmt.Sprintf(
		"https://oauth.mail.ru/login?client_id=%s&response_type=code&scope=%s&redirect_uri=%s&state=%s",
		uuc.conf.OAuth.ClientID,
		uuc.conf.OAuth.Scope,
		uuc.conf.OAuth.RedirectURI,
		state,
	)
}
