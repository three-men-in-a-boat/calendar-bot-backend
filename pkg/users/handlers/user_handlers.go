package handlers

import (
	"encoding/json"
	"github.com/calendar-bot/pkg/statesDict"
	"github.com/calendar-bot/pkg/types"
	"github.com/calendar-bot/pkg/users/usecase"
	"github.com/labstack/echo"
	"math/rand"
	"net/http"
	"strconv"
)

const RedirectUrlProd = "https://t.me/three_man_in_boat_bot"

type UserHandlers struct {
	userUseCase usecase.UserUseCase
	statesDict  statesDict.StatesDictionary
}

func NewUserHandlers(eventUseCase usecase.UserUseCase, states statesDict.StatesDictionary) UserHandlers {
	return UserHandlers{userUseCase: eventUseCase, statesDict: states}
}

func (e *UserHandlers) changeStateInLink(c echo.Context) error {
	name := c.Param("name")
	seed, err := strconv.Atoi(name)
	if err != nil {
		return err
	}
	rand.Seed(int64(seed))
	state := rand.Int()

	e.statesDict.States[int64(state)] = name

	link := "https://oauth.mail.ru/xlogin?client_id=885a013d102b40c7a46a994bc49e68f1&response_type=code&scope=&redirect_uri=https://calendarbot.xyz/api/v1/oauth&state=" + strconv.Itoa(state)

	return c.String(http.StatusOK, link)
}

func (e *UserHandlers) getEvents(c echo.Context) error {
	var events []types.Event
	event1 := types.Event{Name: "Meeting in Zoom", Participants: []string{"Nikolay, Alexey, Alexandr"}, Time: "Today 23:00"}
	event2 := types.Event{Name: "Meeting in university", Participants: []string{"Mike, Alex, Gabe"}, Time: "Tomorrow 23:00"}
	events = append(events, event1, event2)
	b, err := json.Marshal(events)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, string(b))
}

func (e *UserHandlers) InitHandlers(server *echo.Echo) {
	server.GET("/api/v1/oauth/telegram/:name/ref", e.changeStateInLink)
	server.GET("/api/v1/oauth/telegram/events", e.getEvents)
	server.GET("/api/v1/oauth", e.TelegramOauth)
}

func (uh *UserHandlers) TelegramOauth(rwContext echo.Context) error {
	values := rwContext.Request().URL.Query()

	code := values.Get("code")
	state := values.Get("state")

	stateInt, err := strconv.Atoi(state)
	if err != nil {
		println(err.Error())
		return err
	}

	tgId, err := strconv.Atoi(uh.statesDict.States[int64(stateInt)])
	if err != nil {
		println(err.Error())
		return err
	}

	if err := uh.userUseCase.TelegramCreateUser(int64(tgId), code); err != nil {
		return err
	}

	if err := rwContext.Redirect(http.StatusTemporaryRedirect, RedirectUrlProd); err != nil {
		return err
	}

	return nil
}
