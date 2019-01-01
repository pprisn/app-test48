package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// для вендоринга используется GB
// сборка проекта gb build
// установка зависимостей gb vendor fecth gopkg.in/telegram-bot-api.v4
// установка зависимостей из манифеста gb vendor restore
//структура данных ответа от api.icndb.com
// { "type": "success", "value": { "id": 563, "joke": "Chuck Norris causes the Windows Blue Screen of Death.", "categories": ["nerdy"] } }
type Joke struct {
	ID   uint32 `json: "id"`
	Joke string `json: "joke"`
}
type JokeResponse struct {
	Type  string `json:"type"`
	Value Joke   `json:"value"`
}

//Структура данных ответа от https://translate.yandex.net/api/v1.5/tr.json/translate
// Пример на запрос Hellow Jack, ответ: {"code":200,"lang":"en-ru","text":["\"Хеллоу Джек\""]}
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

var Wd = map[string]string{
	"Sunday":    "Воскресенье",
	"Monday":    "Понедельник",
	"Tuesday":   "Вторник",
	"Wednesday": "Среда",
	"Thursday":  "Четверг",
	"Friday":    "Пятница",
	"Saturday":  "Суббота",
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

	//Регулярное выражение для запроса данных по объекту индекс ОПС
	var validCASE = regexp.MustCompile(`(?m)(^ops[0-9]{6})|(^OPS[0-9]{6})|(^Ops[0-9]{6})|(^Опс[0-9]{6})|(^ОПС[0-9]{6})$`)
	//Регулярное выражение для запроса данных трек номера Регион курьер Липецк 15 или 17 символов 000020004000085
	var validRKLIP = regexp.MustCompile(`(?m)(?m)^(([0-9]{15})|([0-9]{17}))$`)

	//var keywd string
	var sWd string

	// Читаем данные из канала updates и выполняем соответсвующие им действия
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
		default:
			if validCASE.MatchString(update.Message.Text) {
				//Если пользователь выполнил запрос opsINDEX

				//keywd =fmt.Sprintf("%s", time.Now().Weekday())
				sWd = fmt.Sprintf(Wd[fmt.Sprintf("%s", time.Now().Weekday())])
				message = tgbotapi.NewMessage(update.Message.Chat.ID, "Сегодня: "+sWd+" Вы запросили данные о почтовом отделении "+update.Message.Text)
				log.Printf("Запрос данных %s", update.Message.Text)
			} else if validRKLIP.MatchString(update.Message.Text) {
				// Поступил запрос трэк номера РегионКурьер Липецк
				message = tgbotapi.NewMessage(update.Message.Chat.ID, req2rkLip(string(update.Message.Text)))
			} else {
				message = tgbotapi.NewMessage(update.Message.Chat.ID, `Уточните ШКИ отправления.`)
			}
		}
		// В ответном сообщении бота просим показать клавиатуре
		message.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons)
		bot.Send(message)
	}

}
