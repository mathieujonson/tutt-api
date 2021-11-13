package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type TuttResponse struct {
	TheThingId string `json:"theThingId"`
}

type State struct {
	Values struct {
		Thing struct {
			ThingInput struct {
				Value string `json:"Value"`
			} `json:"thing_input"`
		} `json:"thing"`
		OuterDatePicker struct {
			InnerDatePicker struct {
				SelectedDate string `json:"selected_date"`
			} `json:"datepicker"`
		} `json:"date_picker"`
		Background struct {
			SelectImage struct {
				SelectedOption struct {
					Text struct {
						Text string `json:"text"`
					} `json:"text"`
					Value string `json:"value"`
				} `json:"selected_option"`
			} `json:"select_image"`
		} `json:"background"`
	} `json:"values"`
}

type InteractivePayload struct {
	Type        string 		`json:"type"`
	Token    	string 		`json:"token"`
	CallbackID  string      `json:"callback_id"`
	TriggerID   string      `json:"trigger_id"`
	View struct {
		Id		string		`json:"id"`
		State   State		`json:"state"`
	} `json:"view"`
}

func Interactive(c *gin.Context) {
	payloadAsBytes := []byte(c.PostForm("payload"))
	fmt.Println("payloadAsBytes: " + string(payloadAsBytes))

	var payload InteractivePayload
	err := json.Unmarshal(payloadAsBytes, &payload)
	if err != nil {
		panic(err.Error())
	}

	switch {
	case payload.Type == "shortcut":
		sendToSlack(getThingFormJson(payload.TriggerID))
		c.Data(200, "application/json", nil)
	case payload.Type == "view_submission":
		theThingId := postThing(payload.View.State)
		c.Data(200, "application/json", getThingLinkJson(theThingId))
	}
}

func sendToSlack(data []byte) {
	req, _ := http.NewRequest("POST", "https://slack.com/api/views.open", bytes.NewBuffer(data))

	req.Header.Add("Authorization", "Bearer " + os.Getenv("TUTT_SLACK_AUTHORIZATION_TOKEN"))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
    
	resp, err := client.Do(req)
    if err != nil {
        panic(err.Error())
    }

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	
	fmt.Println("sendToSlack response:")
	fmt.Println("statusCode: ", resp.StatusCode)
	fmt.Println("body: ", string(body))

    defer resp.Body.Close()
}

func postThing(state State) string {
	req, _ := http.NewRequest("POST", "https://api.timeuntilthething.com/thing", bytes.NewBuffer([]byte(`{
		"name": "` + state.Values.Thing.ThingInput.Value + `",
		"date": "` + state.Values.OuterDatePicker.InnerDatePicker.SelectedDate + `",
		"background": "` + state.Values.Background.SelectImage.SelectedOption.Value + `"
	}`)))

	client := &http.Client{}
    
	resp, err := client.Do(req)
    if err != nil {
		panic(err.Error())
    }

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	
	var tuttResponse TuttResponse
	err = json.Unmarshal(body, &tuttResponse)
    if err != nil {
        panic(err.Error())
    }
    
	fmt.Println("statusCode: ", resp.StatusCode)
	fmt.Println("theThingId: ", string(tuttResponse.TheThingId))

    defer resp.Body.Close()

	return tuttResponse.TheThingId
}

func getThingLinkJson(theThingId string) []byte {
	fmt.Println("theThingId: ", theThingId)
	return []byte(`{
		"response_action": "update",
		"view": {
			"type": "modal",
			"title": {
				"type": "plain_text",
				"text": "Got it!"
			},
			"blocks": [
				{
					"type": "section",
					"text": {
						"type": "mrkdwn",
						"text": "Watch the countdown for the <https://timeuntilthething.com/thing/` + theThingId + `|thing>. Don't forget to bookmark it!"
					}
				}
			]
		}
	}`)
}

func getThingFormJson(trigger_id string) []byte {
	return []byte(`{
		"trigger_id": "` + trigger_id + `",
		"view": {
			"type": "modal",
			"callback_id": "thing-form",
			"title": {
				"type": "plain_text",
				"text": "Add a thing"
			},
			"submit": {
				"type": "plain_text",
				"text": "Add the thing!"
			},
			"blocks": [
				{
					"type": "input",
					"block_id": "thing",
					"label": {
						"type": "plain_text",
						"text": "What's your thing?"
					},
					"element": {
						"type": "plain_text_input",
						"action_id": "thing_input",
						"placeholder": {
							"type": "plain_text",
							"text": "Title of your thing"
						}
					}
				},
				{
					"type": "section",
					"block_id": "date_picker",
					"text": {
						"type": "mrkdwn",
						"text": "When is the thing?"
					},
					"accessory": {
						"type": "datepicker",
						"action_id": "datepicker",
						"initial_date": "` + time.Now().Format("2006-01-02") + `",
						"placeholder": {
							"type": "plain_text",
							"text": "Select a date"
						}
					}
				},
				{
					"type": "section",
					"block_id": "background",
					"text": {
						"type": "mrkdwn",
						"text": "What's a good background for your thing?"
					},
					"accessory": {
						"action_id": "select_image",
						"type": "static_select",
						"placeholder": {
							"type": "plain_text",
							"text": "Select a background"
						},
						"options": [` + buildBackgroundImages() + `]
					}
				}
			]
		}
	}`)
}

func buildBackgroundImages() string {
	options := []string{}

	backgroundImages := map[string]string {
		"beach":"Sandy Beach",
		"city":"City Skyline",
		"concert":"Concert at Night",
		"dinner":"Fine Dining",
		"fireworks":"Fireworks Exploding",
		"flight":"Airplane in the Sky",
		"mountaina":"Snowy Mountains",
		"office":"Office Space",
		"sport":"Sports Arena",
	}

	for key, element := range backgroundImages {
		options = append(options, `{
			"text": {
				"type": "plain_text",
				"text": "` + element + `"
			},
			"value": "` + key + `"
		}`)
	}

	return strings.Join(options[:], ",")
}