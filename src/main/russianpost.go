package main

import (
	"bytes"
	"crypto/tls"
	"os"
	"time"

	//	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"

	//"log"
	"net/http"
	"strings"
)

//https://www.onlinetool.io/xmltogo/
type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	S       string   `xml:"S,attr"`
	Body    struct {
		Text                        string `xml:",chardata"`
		GetOperationHistoryResponse struct {
			Text                 string `xml:",chardata"`
			Ns2                  string `xml:"ns2,attr"`
			Ns3                  string `xml:"ns3,attr"`
			Ns4                  string `xml:"ns4,attr"`
			Ns5                  string `xml:"ns5,attr"`
			Ns6                  string `xml:"ns6,attr"`
			Ns7                  string `xml:"ns7,attr"`
			OperationHistoryData struct {
				Text          string `xml:",chardata"`
				HistoryRecord []struct {
					Text              string `xml:",chardata"`
					AddressParameters struct {
						Text               string `xml:",chardata"`
						DestinationAddress struct {
							Text        string `xml:",chardata"`
							Index       string `xml:"Index"`
							Description string `xml:"Description"`
						} `xml:"DestinationAddress"`
						OperationAddress struct {
							Text        string `xml:",chardata"`
							Index       string `xml:"Index"`
							Description string `xml:"Description"`
						} `xml:"OperationAddress"`
						MailDirect struct {
							Text   string `xml:",chardata"`
							ID     string `xml:"Id"`
							Code2A string `xml:"Code2A"`
							Code3A string `xml:"Code3A"`
							NameRU string `xml:"NameRU"`
							NameEN string `xml:"NameEN"`
						} `xml:"MailDirect"`
						CountryOper struct {
							Text   string `xml:",chardata"`
							ID     string `xml:"Id"`
							Code2A string `xml:"Code2A"`
							Code3A string `xml:"Code3A"`
							NameRU string `xml:"NameRU"`
							NameEN string `xml:"NameEN"`
						} `xml:"CountryOper"`
						CountryFrom struct {
							Text   string `xml:",chardata"`
							ID     string `xml:"Id"`
							Code2A string `xml:"Code2A"`
							Code3A string `xml:"Code3A"`
							NameRU string `xml:"NameRU"`
							NameEN string `xml:"NameEN"`
						} `xml:"CountryFrom"`
					} `xml:"AddressParameters"`
					FinanceParameters struct {
						Text       string `xml:",chardata"`
						Payment    string `xml:"Payment"`
						Value      string `xml:"Value"`
						MassRate   string `xml:"MassRate"`
						InsrRate   string `xml:"InsrRate"`
						AirRate    string `xml:"AirRate"`
						Rate       string `xml:"Rate"`
						CustomDuty string `xml:"CustomDuty"`
					} `xml:"FinanceParameters"`
					ItemParameters struct {
						Text            string `xml:",chardata"`
						Barcode         string `xml:"Barcode"`
						ValidRuType     string `xml:"ValidRuType"`
						ValidEnType     string `xml:"ValidEnType"`
						ComplexItemName string `xml:"ComplexItemName"`
						MailRank        struct {
							Text string `xml:",chardata"`
							ID   string `xml:"Id"`
							Name string `xml:"Name"`
						} `xml:"MailRank"`
						PostMark struct {
							Text string `xml:",chardata"`
							ID   string `xml:"Id"`
							Name string `xml:"Name"`
						} `xml:"PostMark"`
						MailType struct {
							Text string `xml:",chardata"`
							ID   string `xml:"Id"`
							Name string `xml:"Name"`
						} `xml:"MailType"`
						MailCtg struct {
							Text string `xml:",chardata"`
							ID   string `xml:"Id"`
							Name string `xml:"Name"`
						} `xml:"MailCtg"`
						Mass string `xml:"Mass"`
					} `xml:"ItemParameters"`
					OperationParameters struct {
						Text     string `xml:",chardata"`
						OperType struct {
							Text string `xml:",chardata"`
							ID   string `xml:"Id"`
							Name string `xml:"Name"`
						} `xml:"OperType"`
						OperAttr struct {
							Text string `xml:",chardata"`
							ID   string `xml:"Id"`
							Name string `xml:"Name"`
						} `xml:"OperAttr"`
						OperDate string `xml:"OperDate"`
					} `xml:"OperationParameters"`
					UserParameters struct {
						Text    string `xml:",chardata"`
						SendCtg struct {
							Text string `xml:",chardata"`
							ID   string `xml:"Id"`
							Name string `xml:"Name"`
						} `xml:"SendCtg"`
						Sndr string `xml:"Sndr"`
						Rcpn string `xml:"Rcpn"`
					} `xml:"UserParameters"`
				} `xml:"historyRecord"`
			} `xml:"OperationHistoryData"`
		} `xml:"getOperationHistoryResponse"`
	} `xml:"Body"`
}

func req2russianpost(barcode string) string {

	url := fmt.Sprintf("%s%s",
		"https://tracking.russianpost.ru",
		"/rtm34",
	)

	reqload := []byte(strings.TrimSpace(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:oper="http://russianpost.org/operationhistory" xmlns:data="http://russianpost.org/operationhistory/data" xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
   <soap:Header/>
   <soap:Body>
      <oper:getOperationHistory>
         <data:OperationHistoryRequest>
            <data:Barcode>%s</data:Barcode>
            <data:MessageType>0</data:MessageType>
            <data:Language>RUS</data:Language>
         </data:OperationHistoryRequest>
         <data:AuthorizationHeader soapenv:mustUnderstand="1">
            <data:login>%s</data:login>
            <data:password>%s</data:password>
         </data:AuthorizationHeader>
      </oper:getOperationHistory>
   </soap:Body>
</soap:Envelope>`))

	username := os.Getenv("USER_RUSSIANPOST")
	password := os.Getenv("PASS_RUSSIANPOST")

	someXMLasBytes := fmt.Sprintf(string(reqload), barcode, username, password)

	//fmt.Println(string(someXMLasBytes))
	var dialTimeout = time.Duration(30 * time.Second)

	httpClient := &http.Client{
		Timeout: dialTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	//httpClient := new(http.Client)
	resp, err := httpClient.Post(url, `application/soap+xml; charset=utf-8`, bytes.NewBufferString(someXMLasBytes))
	if err != nil {
		//log.Fatal("Error on dispatching request. ", err.Error())
		return fmt.Sprintf("Извините, возникла ошибка:%v", err.Error())
	}

	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("Извините, возникла ошибка:%v", err.Error())
		// handle error
	}
	defer resp.Body.Close()

	envelope := Envelope{}
	err = xml.Unmarshal(htmlData, &envelope)
	if err != nil {
		return fmt.Sprintf("Извините, возникла ошибка:%v", err.Error())

	}

	//fmt.Printf("%+v \n", envelope.Body.GetOperationHistoryResponse.OperationHistoryData)
	//fmt.Println(len(string(htmlData)))

	var Delivstatus []string
	var sDelivstatus string
	for i, rec := range envelope.Body.GetOperationHistoryResponse.OperationHistoryData.HistoryRecord {
		if i == 0 {
			Delivstatus = append(Delivstatus, fmt.Sprintf("%v Масса=%vгр.", rec.ItemParameters.ComplexItemName, rec.ItemParameters.Mass))
			Delivstatus = append(Delivstatus, fmt.Sprintf("%v %v\t %v\t", string(rec.OperationParameters.OperDate)[:19],
				rec.AddressParameters.OperationAddress.Description,
				rec.OperationParameters.OperAttr.Name))

		} else {
			Delivstatus = append(Delivstatus, fmt.Sprintf("%v %v\t %v\t", string(rec.OperationParameters.OperDate)[:19],
				rec.AddressParameters.OperationAddress.Description,
				rec.OperationParameters.OperAttr.Name))
		}
	}

	// Если ничего не найдено
	if len(Delivstatus) == 0 {
		Delivstatus = append(Delivstatus, fmt.Sprintf("Уточните, пожалуйста, номер ШПИ и повторите запрос."))

	}
	sDelivstatus = strings.Join(Delivstatus, "\n")
	return string(sDelivstatus)

}

//
//func main() {
//	barcode := "614025290401601"
//	//_= req2russianpost(barcode)
//	fmt.Printf("%v\n", req2russianpost(barcode))
//	return
//}