package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// リクエストからqualityのパラメータを取得する
	qualityStr := request.QueryStringParameters["quality"]

	// qualityのパラメータを数値に変換する
	quality, err := strconv.Atoi(qualityStr)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid quality parameter",
		}, nil
	}

	// URLのパラメータから画像URLを取得する
	url := request.QueryStringParameters["url"]

	// 画像をダウンロードする
	resp, err := http.Get(url)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Failed to download image: %s", err.Error()),
		}, nil
	}
	defer resp.Body.Close()

	// 画像デコード
	srcImage, _, err := image.Decode(resp.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Failed to decode image: %s", err.Error()),
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

	// Base64エンコードされた文字列としてレスポンスを返す
	resizedImageString := base64.StdEncoding.EncodeToString(buf.Bytes())
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       resizedImageString,
	}, nil
}

func main() {
	lambda.Start(handler)
}
