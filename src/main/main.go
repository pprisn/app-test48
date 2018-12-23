package main

import (
	"encoding/json"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)
// для вендоринга используется GB
// сборка проекта gb build
// установка зависимостей gb vendor fecth gopkg.in/telegram-bot-api.v4
// установка зависимостей из манифеста gb vendor restore
type Joke struct {
	ID   uint32 `json: "id"`
	Joke string `json: "joke"`
}
type JokeResponse struct {
	Type  string `json:"type"`
	Value Joke   `json:"value"`
}

var buttons = []tgbotapi.KeyboardButton{
	tgbotapi.KeyboardButton{Text: "Get Прикол"},
        tgbotapi.KeyboardButton{Text: "Прикол на русском"},	
}
//При старте приложения, оно скажет телеграму ходить с обновлениями по этому URL

const WebhookURL = "https://app-test48.herokuapp.com/"

//const Key = "trnsl.1.1.20181223T210433Z.361ff973b9abaaa1.4c419ee8e5989f4c18f5039d5049b5a5d7b398d7"
const WebTranslateURL = "https://translate.yandex.net/api/v1.5/tr.json/translate?lang=en-ru&key=trnsl.1.1.20181223T210433Z.361ff973b9abaaa1.4c419ee8e5989f4c18f5039d5049b5a5d7b398d7"

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
	sjoke := getJoke()
	c := http.Client{}
        transURL := WebTranslateURL+"&txt="+sjoke
	resp, err := c.Get(transURL)
	if err != nil {
		return "Переводчик API not responding"
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var tjoke string
	//err = json.Unmarshal(body, &tjoke)
        tjoke = resp.Body.String()
	if err != nil {
		return "Joke error"
	}
	return tjoke
        

}

func main() {
        // Heroku прокидывает порт для приложения в переменную окружения PORT
        port := os.Getenv("PORT")
	bot, err := tgbotapi.NewBotAPI("721794920:AAG6xnxtZHmCC-u6-55-LMAVnIakqOjUqv0")
        //721794920:AAG6xnxtZHmCC-u6-55-LMAVnIakqOjUqv0
        if err != nil {
		log.Fatal(err)
	}

	bot.Debug =true

	log.Printf("Authorized on account %s",bot.Self.UserName)
	// устанавливаем иубхук
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(WebhookURL))
        if err != nil {
		log.Fatal(err)
	}

        updates := bot.ListenForWebhook("/")
        go http.ListenAndServe(":"+port,nil)
        
	// получаем все обновления из канала updates
	for update := range updates {
		var message tgbotapi.MessageConfig
		log.Println("received text: ",update.Message.Text)
		switch update.Message.Text {
		case "Get Прикол" :
			//Если пользователь нажал на кнопку то придет сообщение Get Joke
		        message = tgbotapi.NewMessage(update.Message.Chat.ID, getJoke())
		case "Прикол на русском" :
			//Если пользователь нажал на кнопку то придет сообщение Get Joke
		        message = tgbotapi.NewMessage(update.Message.Chat.ID, getTranslate())

		default:
			message = tgbotapi.NewMessage(update.Message.Chat.ID, `Press "Get Joke" to receive joke`)
		}
		// В ответном сообщении бота просим показать клавиатуре
		message.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons)
		bot.Send(message)
	}

}
