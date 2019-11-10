package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"net/http"
	"sync"
	"time"
)

const ADVICE_UPDATE_INTERVAL = 60
const apiUrl = "http://fucking-great-advice.ru/api/random"

const ADVICES_DB = "advices.db"
const ADVICES_BUCKET = "advices"

type Advice struct {
	mut        sync.Mutex
	Id         int32 //0 - no rain, 1 - light possible rain, 2 - rain
	lastText   string
	updateTime time.Time
	db         *bolt.DB
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
	if len(text) < 2 {
		if advice.isFresh() {
			text = advice.lastText
		} else {
			text = TEXT_FAIL
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
	Id    int    `json:"id"`
	Text  string `json:"text"`
	Sound string `json:"sound"`
}

func (advice *Advice) updateAdvice() (text string) {
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
			if len(jval.Text) > 2 {

				advice.mut.Lock()

				defer advice.mut.Unlock()

				fmt.Printf("Advice: id:%d, text: %s\n", jval.Id, jval.Text)

				advice.lastText = jval.Text
				text = jval.Text
				advice.updateTime = time.Now()

				if advice.db != nil {
					err := advice.db.Update(func(tx *bolt.Tx) error {
						b, err := tx.CreateBucketIfNotExists([]byte(ADVICES_BUCKET))
						if err != nil {
							return fmt.Errorf("Can't create a bucket: %s", err)
						}
						err = b.Put([]byte(fmt.Sprintf("%d", jval.Id)), []byte(jval.Text))
						//log.Printf("Saved: %s -> %.2f", now.UTC().Format(time.UnixDate), f)
						return err
					})
					if err != nil {
						log.Printf("Error add new advice: %s", err)
					}

				}

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

func InitAdvice() (advice *Advice) {
	advice = &Advice{Id: 0, lastText: ""}

	db, err := bolt.Open(ADVICES_DB, 0600, nil)
	if err == nil {
		advice.db = db
	}
	go advice.updateAdviceEveryNsec(ADVICE_UPDATE_INTERVAL)
	return
}
