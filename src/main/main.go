package main

import (
	"fmt"
	//	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// для вендоринга используется GB
// сборка проекта gb build
// установка зависимостей gb vendor fecth gopkg.in/telegram-bot-api.v4
// установка зависимостей из манифеста gb vendor restore
//структура данных ответа от api.icndb.com
type Joke struct {
	ID   uint32 `json: "id"`
	Joke string `json: "joke"`
}
type JokeResponse struct {
	Type  string `json:"type"`
	Value Joke   `json:"value"`
}
//Структура данных ответа от https://translate.yandex.net/api/v1.5/tr.json/translate
type TranslateJoke struct {
	CODE uint32   `json: "code"`
	Lang string   `json: "lang"`
	Text []string `json: "text"`
}

//Объявляем клавиатурные кнопки для tg
var buttons = []tgbotapi.KeyboardButton{
	tgbotapi.KeyboardButton{Text: "Get Прикол"},
	tgbotapi.KeyboardButton{Text: "Прикол на русском"},
}

var Keytg string
var Keyyandex string

func init() {
	Keytg = os.Getenv("KEYTG")
	Keyyandex = os.Getenv("KEYYANDEX")
}

//При старте приложения, оно скажет телеграму ходить с обновлениями по этому URL

const WebhookURL = "https://app-test48.herokuapp.com/"
const WebTranslateURL = "https://translate.yandex.net/api/v1.5/tr.json/translate"

func getJoke() string {
	c := http.Client{}
	resp, err := c.Get("http://api.icndb.com/jokes/random?limitTo=[nerdy]")
	if err != nil {
		return "jokes API not responding"
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	joke := JokeResponse{}
	err = json.Unmarshal(body, &joke)
	if err != nil {
		return "Joke error"
	}
	return joke.Value.Joke
}

func getTranslate() string {
	//Получим очередную шутку
	sjoke := getJoke()
	c := http.Client{}
	lang := "en-ru"
	// подготовим параметры для POST запроса
	builtParams := url.Values{"key": {Keyyandex}, "lang": {lang}, "text": {sjoke}, "options": {"1"}}
	resp, err := c.PostForm(WebTranslateURL, builtParams)
	if err != nil {
		return "Переводчик yandex API not responding..."
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	tjoke := TranslateJoke{}
	err = json.Unmarshal(body, &tjoke)
	if err != nil {
		serr := fmt.Sprintf("%v", err)
		return "Unmarshal error " + serr
	}
	return strings.Join(tjoke.Text[:], ",")

}

func main() {
	// Heroku прокидывает порт для приложения в переменную окружения PORT
	port := os.Getenv("PORT")

	bot, err := tgbotapi.NewBotAPI(Keytg)
	//Privat key telegram XXXXxxxx:kqOjUqv0
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)
	
	// устанавливаем привязку бота к сервису WebhookURL
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
	if err != nil {
		log.Fatal(err)
	}
        // Открываем канал для получения данных при обращении к сервису WebhookURL '/'
	updates := bot.ListenForWebhook("/")
	//Запускаем сервер через горутину 
	go http.ListenAndServe(":"+port, nil)

	// Читаем данные из канала updates
	for update := range updates {
		var message tgbotapi.MessageConfig
		log.Println("received text: ", update.Message.Text)
		switch update.Message.Text {
		case "Get Прикол":
			//Если пользователь нажал на кнопку то придет сообщение Get Joke
			message = tgbotapi.NewMessage(update.Message.Chat.ID, getJoke())
		case "Прикол на русском":
			//Если пользователь нажал на кнопку то придет сообщение Get Joke
			message = tgbotapi.NewMessage(update.Message.Chat.ID, getTranslate())
		case "^ops[0-9]{6}$":
			//Если пользователь нажал на кнопку то придет сообщение Get Joke
			message = tgbotapi.NewMessage(update.Message.Chat.ID, "Это почтовый индекс "+update.Message.Text)
		default:
			message = tgbotapi.NewMessage(update.Message.Chat.ID, `Press "Get Joke" to receive joke`)
		}
		// В ответном сообщении бота просим показать клавиатуре
		message.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons)
		bot.Send(message)
	}

}
