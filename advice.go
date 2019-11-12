package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const ADVICES_DB = "alladvices.db"

const ADVICES_BUCKET = "alladvices"

type BoltAdvice struct {
	Text string
	Tags []string
}
type Advices struct {
	mut     sync.Mutex
	indices []int
	advices map[int]BoltAdvice
}

const TEXT_FAIL = "Отдохни, блять."

func (alladvices *Advices) getAdvice() (text string) {

	randKey := alladvices.indices[rand.Intn(len(alladvices.indices))]
	if advice, ok := alladvices.advices[randKey]; ok {
		return advice.Text
	}
	return TEXT_FAIL
}

func InitAdvice() (alladvices *Advices) {

	alladvices = &Advices{
		advices: make(map[int]BoltAdvice),
		indices: make([]int, 0),
	}

	db, err := bolt.Open(ADVICES_DB, 0600, nil)
	if db != nil && err == nil {
		err = db.View(func(tx *bolt.Tx) error {
			if b := tx.Bucket([]byte(ADVICES_BUCKET)); b != nil {
				c := b.Cursor()
				k, v := c.First()
				count := 0
				for ; k != nil; k, v = c.Next() {
					//fmt.Printf("key=%s, value=%s\n", k, v)
					count++
					var bAdvice BoltAdvice
					if err := json.Unmarshal(v, &bAdvice); err == nil {
						index, err := strconv.Atoi(string(k))
						if err == nil {
							alladvices.indices = append(alladvices.indices, index)
							alladvices.advices[index] = bAdvice
						}

					}
				}
				log.Printf("Loaded %d values ", len(alladvices.indices))
			}
			return nil
		})
	} else {
		log.Fatalf("Can't open database: %s", err) //dumb ipc sync
	}

	if len(alladvices.indices) == 0 {
		log.Fatal("error load database")
	}
	rand.Seed(time.Now().Unix())
	//go advice.updateAdviceEveryNsec(ADVICE_UPDATE_INTERVAL)
	//os.Exit(0)
	return
}
