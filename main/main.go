package main

import (
	"strings"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	Actor struct {
		Name string `json:"display_name"`
	} `json:"actor"`

	Approval struct {
		User struct {
			Name string `json:"display_name"`
		} `json:"user"`
	} `json:"approval"`

	Repository struct {
		Name string `json:"name"`
		Project struct {
			Name string `json:"name"`
		} `json:"project"`
	} `json:"repository"`

	PullRequest struct {
		State string `json:"state"`
		Title string `json:"title"`
	} `json:"pullrequest"`
	
}

type SlackRequest struct {
	Text string `json:"text"`
} 

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(request events.APIGatewayProxyRequest) (Response, error) {
	var buf bytes.Buffer

	data := Request{}
	json.Unmarshal([]byte(request.Body), &data)
	fmt.Println("Body ",request.Body)
	
	url:=bytes.Buffer{}
	url.WriteString("https://docs.google.com/forms/d/e/1FAIpQLScThanFbQQ-uKYTzzXFnxyF8p30kcqlTSXekFCeGsmMDsdMtw/viewform?usp=pp_url")
	project := strings.Replace(data.Repository.Project.Name, " ", "%20", -1)
	repository := strings.Replace(data.Repository.Name," ","%20", -1)
	actorName := strings.Replace(data.Actor.Name, " ", "%20", -1)
	title := strings.Replace(data.PullRequest.Title, " ", "%20", -1)

	url.WriteString("&entry.375868032=")
	url.WriteString(project)
	url.WriteString("&entry.1818420032=")
	url.WriteString(repository)
	url.WriteString("&entry.497941682=")
	url.WriteString(actorName)
	url.WriteString("&entry.1697493684=")
	url.WriteString(title)
	url.WriteString("&entry.2015034005=")
	 
	if (data.PullRequest.State == "DECLINED") {
		url.WriteString("False")
	} else {
		url.WriteString("True")
	}
	fmt.Println("URL ", url.String())
	
	
	postToSlack(url.String())


	body, err := json.Marshal(map[string]interface{}{
		"message": data.PullRequest.State,
	})
	if err != nil {
		return Response{StatusCode: 404}, err
	}
	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "pr-handler",
		},
	}

	return resp, nil
}

func postToSlack(url string) (string) {
	slackRequestStruct := &SlackRequest{Text: url}
	requestJSON,_ := json.Marshal(slackRequestStruct)
	postURL := "https://hooks.slack.com/services/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"

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(requestJSON))
	
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
			panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	return "Done"
}

func main() {
	lambda.Start(Handler)
}
