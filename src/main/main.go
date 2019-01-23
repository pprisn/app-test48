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
// Описание вложенной структуры ответа ..{ "id": 563, "joke": "Chuck
type Joke struct {
	ID   uint32 `json: "id"`
	Joke string `json: "joke"`
}

// Описание начала структуры ответа  { "type": "success", "value": {..
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

// Карта соответствия английских и русских наименований дней недели
var Wd = map[string]string{
	"Sunday":    "Воскресенье",
	"Monday":    "Понедельник",
	"Tuesday":   "Вторник",
	"Wednesday": "Среда",
	"Thursday":  "Четверг",
	"Friday":    "Пятница",
	"Saturday":  "Суббота",
}

// Keytg ключ бота Telegramm
var Keytg string

// Keytg ключ сервиса на yandex
var Keyyandex string

//Проинициализируем ключи Keytg Keyyandex
func init() {
	Keytg = os.Getenv("KEYTG")
	Keyyandex = os.Getenv("KEYYANDEX")
}

//При старте приложения, оно скажет телеграму ходить с обновлениями по этому URL

//  WebhookURL url сервера бота
const WebhookURL = "https://app-test48.herokuapp.com/"

// WebTranslateURL url сервиса переводчика на русский с английского
const WebTranslateURL = "https://translate.yandex.net/api/v1.5/tr.json/translate"

// Функция getJoke() string , возвращает строку с шуткой, полученной от сервиса  http://api.icndb.com/jokes/random?limitTo=[nerdy]
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

// Функция getTranslate(mytext string) string, переводит полученный английский текст на русский или если
// текст отсутствует, получает очередную шутку и возвращает ее перевод на русском
func getTranslate(mytext string) string {
	var sjoke string
	if mytext == "" {
		//Получим очередную шутку на английском
		sjoke = getJoke()
	} else {
		// если поступил англ. текст, принимаем его для перевода
		sjoke = mytext
	}
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
	var validRKLIP = regexp.MustCompile(`(?m)^(([0-9]{15})|([0-9]{17}))$`)
	var validTranslate = regexp.MustCompile(`(?m)(^[a-z-A-Z].*$)`)
	var validRUSSIANPOST = regexp.MustCompile(`(?m)^(([0-9]{14})|([0-9A-Z]{13}))$`)
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
			message = tgbotapi.NewMessage(update.Message.Chat.ID, getTranslate(""))
		default:
			if validCASE.MatchString(update.Message.Text) == true {
				//Если пользователь выполнил запрос opsINDEX
				sWd = fmt.Sprintf(Wd[fmt.Sprintf("%s", time.Now().Weekday())])
				message = tgbotapi.NewMessage(update.Message.Chat.ID, "Сегодня: "+sWd+" Вы запросили данные о почтовом отделении "+update.Message.Text)
				log.Printf("Запрос данных %s", update.Message.Text)
			} else if validRKLIP.MatchString(update.Message.Text) == true {
				// Поступил запрос трэк номера РегионКурьер Липецк
				message = tgbotapi.NewMessage(update.Message.Chat.ID, req2rkLip(string(update.Message.Text)))
			} else if validRUSSIANPOST.MatchString(update.Message.Text) == true {
				// Поступил запрос трэк номера RUSSIANPOST
				//mystr = strings.ToUpper(string(update.Message.Text))
				message = tgbotapi.NewMessage(update.Message.Chat.ID, req2russianpost(string(update.Message.Text)))
				// Если в ОАСУ РПО не найдено отправление, ищем в РК
				if strings.Contains(message.Text, "Уточните") {
					message = tgbotapi.NewMessage(update.Message.Chat.ID, req2rkLip(string(update.Message.Text)))
				}

			} else if validTranslate.MatchString(update.Message.Text) == true {
				// Поступил запрос текста на английском - переведем его.
				message = tgbotapi.NewMessage(update.Message.Chat.ID, getTranslate(update.Message.Text))

			} else {
				message = tgbotapi.NewMessage(update.Message.Chat.ID, `Уточните Штриховой Почтовый Идентификатор, пожалуйста. И повторите запрос.`)
			}
		}
		// В ответном сообщении бота просим показать клавиатуру
		message.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons)
		bot.Send(message)
	}

}
