package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// quality は元の oyaki と合わせて決めうち
	quality := 90

	// Origin は 環境変数からとる
	originHost := os.Getenv("OYAKI_ORIGIN_HOST")

	url := "https://" + originHost + request.RequestContext.Path

	req, _ := http.NewRequest("GET", url, nil)

	// prepare request headers
	req.Header.Set("User-Agent", "oyaki-lambda")
	if request.Headers["If-Modified-Since"] != "" {
		req.Header.Set("If-Modified-Since", request.Headers["If-Modified-Since"])
	}
	xff := request.Headers["X-Forwarded-For"]
	if len(xff) > 1 {
		req.Header.Set("X-Forwarded-For", request.Headers["X-Forwarded-For"])
	}

	// 画像をダウンロードする
	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Failed to download image: %s", err.Error()),
		}, nil
	}
	defer resp.Body.Close()

	// 画像デコード
	var srcImage image.Image
	contentType := resp.Header.Get("Content-Type")
	// jpeg だけ
	switch {
	case strings.Contains(contentType, "jpeg"):
		srcImage, err = jpeg.Decode(resp.Body)
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Sprintf("Failed to decode image: %s", err.Error()),
			}, nil
		}
	default:
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Not supported image: %s", err.Error()),
		}, nil
	}

	// 画像最適化
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, srcImage, &jpeg.Options{Quality: quality})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Failed to encode image: %s", err.Error()),
		}, nil
	}

	// prepare response header
	rh := map[string]string{}
	rh["Content-Type"] = "image/jpeg"
	rh["Access-Control-Allow-Origin"] = "*/*"
	if resp.Header.Get("Last-Modified") != "" {
		rh["Last-Modified"] = resp.Header.Get("Last-Modified")
	} else {
		rh["Last-Modified"] = time.Now().UTC().Format(http.TimeFormat)
	}

	// Base64エンコードされた文字列としてレスポンスを返す
	resizedImageString := base64.StdEncoding.EncodeToString(buf.Bytes())
	return events.APIGatewayProxyResponse{
		StatusCode:      http.StatusOK,
		Headers:         rh,
		Body:            resizedImageString,
		IsBase64Encoded: true,
	}, nil
}

func main() {
	lambda.Start(handler)
}
