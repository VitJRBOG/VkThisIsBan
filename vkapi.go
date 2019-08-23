package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Error хранит информацию об ошибке, которую вернул vk api
type Error struct {
	Code          int    `json:"error_code"`
	Message       string `json:"error_msg"`
	RequestParams []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"request_params"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("code %d: %s", e.Code, e.Message)
}

func sendRequestVKAPI(methodName string, params map[string]string, accessToken string) ([]byte, error) {

	// собираем ссылку к методу vk api
	apiurl := "https://api.vk.com/method/"
	queryURL, err := url.Parse(apiurl + methodName)
	if err != nil {
		return nil, err
	}

	// собираем параметры для запроса к vk api
	query := url.Values{}
	for key, value := range params {
		query.Set(key, value)
	}
	query.Set("access_token", accessToken)
	query.Set("lang", "0")

	// кодируем и добавляем параметры запроса к url запроса
	queryURL.RawQuery = query.Encode()

	// отправляем запрос к vk api
	resp, err := http.Get(queryURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// извлекаем тело ответа сервера
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// описываем структуру для данных из ответа сервера
	var handler struct {
		Error    *Error
		Response json.RawMessage
	}

	// заполняем структуру для ответа сервера
	err = json.Unmarshal(body, &handler)
	if err != nil {
		return nil, err
	}

	// проверяем ответ сервера на наличие сообщения об ошибке
	if handler.Error != nil {
		return nil, handler.Error
	}

	// если ошибки нет, то возвращаем результат запроса
	return handler.Response, nil
}
