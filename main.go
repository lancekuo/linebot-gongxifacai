// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/v7/linebot"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var bot *linebot.Client
var userList []string
var logger *zap.Logger

func main() {

	encoderCfg := zap.NewProductionEncoderConfig()
	atom := zap.NewAtomicLevel()
	logger := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))

	atom.SetLevel(zap.InfoLevel)
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	zap.L().Info("replaced zap's global loggers")
	logger.Info("Enable leader election", zap.Bool("enableLeaderElection", true))

	var err error
	userList = []string{
		"Ian",
		"Mark",
		"Lucas",
		"Ploking",
	}
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	zap.L().Info("ParseRequest", zap.Any("events", events))
	// logger.Info("Gongxifacai", zap.Any("WeekNumber", getWeekNumber()), zap.Any("Who", userList[getWeekUserIdx()]))

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			// Handle only on text message
			case *linebot.TextMessage:
				// GetMessageQuota: Get how many remain free tier push message quota you still have this month. (maximum 500)
				quota, err := bot.GetMessageQuota().Do()
				if err != nil {
					log.Println("Quota err:", err)
				}
				// message.ID: Msg unique ID
				// message.Text: Msg text
				responseMessage := processTextMessage(message.Text)
				if len(responseMessage) > 0 {
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(responseMessage)).Do(); err != nil {
						log.Print(err)
					}
				}
				zap.L().Info("msg ID:" + message.ID + ":" + "Get:" + message.Text + " , \n OK! remain message:" + strconv.FormatInt(quota.Value, 10))

			// Handle only on Sticker message
			case *linebot.StickerMessage:
				var kw string
				for _, k := range message.Keywords {
					kw = kw + "," + k
				}

				// outStickerResult := fmt.Sprintf("收到貼圖訊息: %s, pkg: %s kw: %s  text: %s", message.StickerID, message.PackageID, kw, message.Text)
				// if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(outStickerResult)).Do(); err != nil {
				// 	log.Print(err)
				// }
			}
		}
	}
}
func processTextMessage(text string) string {
	if strings.Index(text, "恭喜發財") == 0 {

		return fmt.Sprintf("這週是第%d週，應該是%s要買喔。%s上週的對獎了沒！？", getWeekNumber(), userList[getWeekUserIdx()], userList[getWeekUserIdx()-1])
	} else {
		return ""
	}
}
func getWeekNumber() int {
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		panic(err)
	}
	tn := time.Now().In(loc)
	fmt.Println(tn)
	_, week := tn.ISOWeek()
	return week
}
func getWeekUserIdx() int {
	var userIdx int
	userIdx = getWeekNumber() % len(userList)
	return userIdx
}
