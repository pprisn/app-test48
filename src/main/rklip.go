package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// Статусы доставки отправлений Регион Курьер
var Delivstatnames = map[string]string{
	"dsLoaded":     "Загружено из файла",
	"dsNew":        "Новое",
	"dsToSend":     "К отправке",
	"dsToDelivery": "К доставке",
	"dsInDelivery": "Доставляется",
	"!dsDelivered": "Доставлено",
	"!dsRetired":   "Отсутствие адресата по адресу",
	"!dsDenied":    "Отказ от получения",
	"!dsUnclaimed": "Истечение срока",
}

//[{"barcode":"000020004000085","attachment":"39800075522535","postoffice":"399205","delivery_site":"39920501","receipt_date":"2014-10-02","delivery_status":"!dsDelivered","delivery_status_name":"\u0414\u043e\u0441\u0442\u0430\u0432\u043b\u0435\u043d\u043e","delivery_date":"2014-10-14 11:57:25"}]
type RKResp []struct {
	Barcode            string `json:"barcode"`
	Attachment         string `json:"attachment"`
	Postoffice         string `json:"postoffice"`
	DeliverySite       string `json:"delivery_site"`
	ReceiptDate        string `json:"receipt_date"`
	DeliveryStatus     string `json:"delivery_status"`
	DeliveryStatusName string `json:"delivery_status_name"`
	DeliveryDate       string `json:"delivery_date"`
}

func RKResp2nilbyte() []byte {
	data := []byte(`[{"barcode":"","attachment":"","postoffice":"","delivery_site":"","receipt_date":"","delivery_status":"","delivery_status_name":"","delivery_date":""}]`)
	return data
}

func req2rkLip(barcode string) string {

	var Delivstatus []string
	var sDelivstatus string
	sDelivstatus = ""
	sudkey := os.Getenv("SUDKEY")
	sudcrt := os.Getenv("SUDCRT")
	cacrt := []byte(os.Getenv("CACRT"))

	//	caCert, err := ioutil.ReadFile("ca.crt")
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	caCertPool := x509.NewCertPool()
	//	caCertPool.AppendCertsFromPEM(caCert)
	caCertPool.AppendCertsFromPEM(cacrt)
	//cert, err := tls.LoadX509KeyPair("sud.crt", "sud.key")
	cert, err := tls.X509KeyPair([]byte(sudcrt), []byte(sudkey))
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true,
			},
		},
	}

	//resp, err := client.Get("https://d01rkweblb.main.russianpost.ru/depeche/?r=service/status&attachment=000020004000085")
	//resp, err := client.Get("https://d01rkweblb.main.russianpost.ru/depeche/?r=service/status&barcode=000020004000085")
	urlRK := "https://d01rkweblb.main.russianpost.ru/depeche/?r=service/status&barcode="
	resp, err := client.Get(urlRK + barcode)
	if err != nil {
		Delivstatus = append(Delivstatus, fmt.Sprintf("Извините, сервис %v не доступен \n", urlRK))
		sDelivstatus = strings.Join(Delivstatus, ";")
		//log.Fatal(err)
		return sDelivstatus
	}

	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Delivstatus = append(Delivstatus, fmt.Sprintf("Извините, что-то пошло не так, повторите пожалуйста попытку. \n"))
		sDelivstatus = strings.Join(Delivstatus, ";")
		return sDelivstatus
		//log.Fatal(err)
	}

	defer resp.Body.Close()

	trk := RKResp{}
	// Если содержимое htmlDtat не будет соответствовать структуре RKResp будет panic
	// выполним проверку на соответствие htmlData структе RKResp
	// Проверка на валидность структуры htmlData, если не валидна - заполняем пустыми данными
	var validRKLip = regexp.MustCompile(`(?)(^\[\{"barcode":.*"attachment":.*"postoffice":.*"delivery_site":.*"receipt_date":.*"delivery_status":.*"delivery_status_name":.*"delivery_date":.*\}\]$)`)
	if !validRKLip.MatchString(string(htmlData)) {
		htmlData = RKResp2nilbyte()
	}
	err_trk := json.Unmarshal(htmlData, &trk)
	if err_trk != nil {
		Delivstatus = append(Delivstatus, fmt.Sprintf("Извините, API РегионКурьера изменилось, а pprisn_bot не в курсе, вы можете сообщить о проблеме по адресу pprisn@yandex.ru."))
		sDelivstatus = strings.Join(Delivstatus, "\n")
		return sDelivstatus
		//log.Fatal(err_trk)
	}

	if trk[0].Barcode == "" {

		Delivstatus = append(Delivstatus, fmt.Sprintf("Отправление с ШПИ %v не найдено\t", barcode))
		Delivstatus = append(Delivstatus, fmt.Sprintf("Уточните ШПИ, пожалуйста, и повторите запрос."))
		sDelivstatus = strings.Join(Delivstatus, "\n")

	} else {
		Delivstatus = append(Delivstatus, fmt.Sprintf("Штрих код           %v\t", trk[0].Barcode))
		Delivstatus = append(Delivstatus, fmt.Sprintf("Вложение            %v\t", trk[0].Attachment))
		Delivstatus = append(Delivstatus, fmt.Sprintf("Почтовое отделение  %v\t", trk[0].Postoffice))
		Delivstatus = append(Delivstatus, fmt.Sprintf("Доставочный участок %v\t", trk[0].DeliverySite))
		Delivstatus = append(Delivstatus, fmt.Sprintf("Дата приема         %v\t", trk[0].ReceiptDate))
		Delivstatus = append(Delivstatus, fmt.Sprintf("Статус доставки     %v\t", Delivstatnames[trk[0].DeliveryStatus]))
		Delivstatus = append(Delivstatus, fmt.Sprintf("Дата доставки       %v\t", trk[0].DeliveryDate))
		sDelivstatus = strings.Join(Delivstatus, "\n")
		//fmt.Printf(string(htmlData))
	}

	return sDelivstatus
}

//func main() {
//	barcode := "000020004000085"
//	status := req2rkLip(barcode)
//	fmt.Println(status)
//
//	barcode = "00069611513249351"
//	status = req2rkLip(barcode)
//	fmt.Println(status)
//
//	barcode = "100069611513249351"
//	status = req2rkLip(barcode)
//	fmt.Println(status)
//}
