package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const ADVICE_UPDATE_INTERVAL = 7200
const apiUrl = "http://fucking-great-advice.ru/api/random"

type Advice struct {
	mut             sync.Mutex
	Id  int32 //0 - no rain, 1 - light possible rain, 2 - rain
	lastText  string
	updateTime      time.Time
}




func (advice *Advice) updateAdviceEveryNsec(N uint64 /*, b *tb.Bot,  chats *Chats*/) {

	for {
		advice.updateAdvice()

		//wake up every N minutes
		time.Sleep(time.Second * time.Duration(N))

	}
}

const TEXT_DEFAULT = "Падажжи, стартую..."
const TEXT_FAIL = "Отдохни, блять."


func (advice *Advice) isFresh() (fresh bool) {

	log.Printf("Is advice fresh: %t", time.Since(advice.updateTime).Hours() < 6)
	return time.Since(advice.updateTime).Hours() < 6

}


func (advice *Advice) getfreshAdvice() (text string) {
	text = advice.updateAdvice()
	if len(text) <2 {
		if advice.isFresh() {
			text = advice.lastText
		} else {
			text = TEXT_DEFAULT
		}
	}

	return
}

func (advice *Advice) getAdvice() (text string) {
	text = TEXT_DEFAULT
	if advice.isFresh() {
		text = advice.lastText
	}
	return
}


//{"id":26140,"text":"Как тратишь время, так блять и живёшь","sound":""}
type JSAdvice struct {

	Id	int `json:id`
	Text string `json:text`
	Sound string `json:sound`
}

func (advice *Advice) updateAdvice() (text string){
	var myClient = &http.Client{Timeout: 30 * time.Second}

	res, err := myClient.Get(apiUrl)
	if err == nil && res.StatusCode == 200 {
		dec := json.NewDecoder(res.Body)

		for dec.More() {
			var jval JSAdvice
			err := dec.Decode(&jval)
			if err != nil {
				fmt.Println(err)
				break
			}
			if len(jval.Text) > 0 {

				advice.mut.Lock()

				fmt.Printf("Advice: id:%u, text: %s\n", jval.Id, jval.Text)

				advice.lastText = jval.Text
				text = jval.Text
				advice.updateTime = time.Now()
				advice.mut.Unlock()

			}

			break
		}
		err = res.Body.Close()
	} else if res != nil && res.StatusCode != 200 {
		log.Printf("Fetch advice error: %u %s", res.StatusCode, res.Status)
	} else {
		log.Println(err)
	}

	return
}

func InintWeather() (forecast *Advice) {
	forecast = &Advice{Id: 0, lastText: ""}
	go forecast.updateAdviceEveryNsec(ADVICE_UPDATE_INTERVAL)
	return
}

