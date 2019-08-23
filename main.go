package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// BanInfo хранит информацию о блокировке пользователя
type BanInfo struct {
	UserID      string
	GroupID     string
	ReasonTitle string
	UnbanDate   string
}

// Data хранит данные для параметров
type Data struct {
	AccessToken  string         `json:"access_token"`
	Groups       []Groups       `json:"groups"`
	BanReasons   []string       `json:"ban_reasons"`
	BanDurations []BanDurations `json:"ban_durations"`
}

// Groups хранит информацию о пабликах
type Groups struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// BanDurations хранит информацию о сроках блокировки
type BanDurations struct {
	Title    string `json:"title"`
	Duration int    `json:"duration"`
}

func main() {
	var banInfo BanInfo

	// получаем url к страничке пользователя
	userURL, err := getUserURL()
	if err != nil {
		fmt.Println(fmt.Errorf("%v", err))
		os.Exit(0)
	}
	fmt.Println("> Gotcha URL to user.")

	// получаем id и полное имя пользователя
	var userFullName string
	banInfo.UserID, userFullName, err = getUserID(userURL)
	if err != nil {
		fmt.Println(fmt.Errorf("%v", err))
		os.Exit(0)
	}
	fmt.Println("> Gotcha ID of user.")

	// выбираем id группы, где будем банить пользователя
	banInfo.GroupID, err = selectGroup()
	if err != nil {
		fmt.Println(fmt.Errorf("%v", err))
		os.Exit(0)
	}
	fmt.Println("> Gotcha ID of group.")

	// выбираем причину, которая будет записана в комментарии к блокировке
	banInfo.ReasonTitle, err = selectReason()
	if err != nil {
		fmt.Println(fmt.Errorf("%v", err))
		os.Exit(0)
	}
	fmt.Println("> Gotcha reason title of ban.")

	// выбираем дату автоматической разблокировки
	banInfo.UnbanDate, err = selectUnbanDate()
	if err != nil {
		fmt.Println(fmt.Errorf("%v", err))
		os.Exit(0)
	}
	fmt.Println("> Gotcha unban date.")

	// блокируем пользователя в соответствии с параметрами
	err = banUser(banInfo)
	if err != nil {
		fmt.Println(fmt.Errorf("%v", err))
		os.Exit(0)
	}

	fmt.Println(fmt.Sprintf("> %v is banned. Check this!", userFullName))
}

// getData получает данные для параметров из файла json
func getData() (Data, error) {
	var data Data

	// получаем путь к файлу с данными
	path, err := readPath()
	if err != nil {
		return data, err
	}
	fullPath := path + "data.json"

	// получаем из json файла данные
	rawData, err := readJSON(fullPath)
	if err != nil {
		return data, err
	}

	// конвертируем их в читабельный вид
	err = json.Unmarshal(rawData, &data)

	return data, nil
}

// getUserURL получает url на пользовательскую страничку
func getUserURL() (string, error) {
	fmt.Print("> [URL to user's profile]: ")
	var userAnswer string
	_, err := fmt.Scan(&userAnswer)
	if err != nil {
		return "", err
	}
	return userAnswer, nil
}

func getUserID(userURL string) (string, string, error) {
	pos := strings.LastIndex(userURL, "/")
	userScreenName := strings.Replace(userURL, userURL[0:pos+1], "", -1)

	// формируем карту с параметрами для запроса к vk api
	params := map[string]string{
		"user_ids": userScreenName,
		"v":        "5.101",
	}

	// получаем данные
	data, err := getData()
	if err != nil {
		return "", "", err
	}

	// отправляем запрос на получение данных о пользователе
	resp, err := sendRequestVKAPI("users.get", params, data.AccessToken)
	if err != nil {
		// если сервер вернул ошибку "слишком много запросов в секунду", то повторяем запрос
		if strings.Contains(err.Error(), "many requests per second") {
			return getUserID(userURL)
		}
		return "", "", err
	}

	// описываемт структуру для словаря с данными о пользователе
	type UserInfo struct {
		UserID          int    `json:"id"`
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		IsClosed        bool   `json:"is_closed"`
		CanAccessClosed bool   `json:"can_access_closed"`
	}
	var userInfo []UserInfo

	// парсим полученный массив байт в структуру с данными о пользователе
	err = json.Unmarshal(resp, &userInfo)
	if err != nil {
		return "", "", err
	}

	// извлекаем id пользователя
	userID := strconv.Itoa(userInfo[0].UserID)

	// извлекаем и собираем полное имя пользователя
	userFullName := fmt.Sprintf("%v %v", userInfo[0].FirstName, userInfo[0].LastName)

	return userID, userFullName, nil
}

// selectGroup дает указать группу, в которой жертва будет заблокирована
func selectGroup() (string, error) {

	// получаем данные
	data, err := getData()
	if err != nil {
		return "", err
	}

	// оглашаем весь список
	for i, groupData := range data.Groups {
		fmt.Printf("> [Groups names]: %d == %v\n", i+1, groupData.Name)
	}
	fmt.Println("> [Groups names]: 00 == Quit")

	// запрашиваем номер группы
	var userAnswer string
	_, err = fmt.Scan(&userAnswer)
	if err != nil {
		return "", err
	}

	// если пользователь ввел 00, то завершаем программу
	if userAnswer == "00" {
		fmt.Println("> Quit...")
		os.Exit(0)
	}

	// конвертируем строку с номером группы в целочисленный тип
	intUserAnswer, err := strconv.Atoi(userAnswer)
	if err != nil {
		return "", err
	}

	// с помощью номера, который ввел пользователь, получаем из списка id группы
	groupID := data.Groups[intUserAnswer-1].ID

	return groupID, nil
}

// selectReason дает указать причину блокировки
func selectReason() (string, error) {

	// получаем данные
	data, err := getData()
	if err != nil {
		return "", err
	}

	// оглашаем весь список
	for i, reason := range data.BanReasons {
		fmt.Printf("> [Reasons]: %d == %v\n", i+1, reason)
	}
	fmt.Println("> [Reasons]: 00 == Quit")

	// запрашиваем номер причины
	var userAnswer string
	_, err = fmt.Scan(&userAnswer)
	if err != nil {
		return "", err
	}

	// если пользователь ввел 00, то завершаем программу
	if userAnswer == "00" {
		fmt.Println("> Quit...")
		os.Exit(0)
	}

	// конвертируем строку с номером причины в целочисленный тип
	intUserAnswer, err := strconv.Atoi(userAnswer)
	if err != nil {
		return "", err
	}

	// с помощью номера, который ввел пользователь, получаем из списка текст причины
	reasonTitle := data.BanReasons[intUserAnswer-1]

	return reasonTitle, nil
}

// selectUnbanDate дает указать дату разблокировки
func selectUnbanDate() (string, error) {

	// получаем данные
	data, err := getData()
	if err != nil {
		return "", err
	}

	// оглашаем весь список
	for i, durationData := range data.BanDurations {
		fmt.Printf("> [Duration]: %d == %v\n", i+1, durationData.Title)
	}
	fmt.Println("> [Duration]: 00 == Quit")

	// запрашиваем номер срока
	var userAnswer string
	_, err = fmt.Scan(&userAnswer)
	if err != nil {
		return "", err
	}

	// если пользователь ввел 00, то завершаем программу
	if userAnswer == "00" {
		fmt.Println("> Quit...")
		os.Exit(0)
	}

	// конвертируем строку с номером срока в целочисленный тип
	intUserAnswer, err := strconv.Atoi(userAnswer)
	if err != nil {
		return "", err
	}

	// с помощью номера, который ввел пользователь, получаем из списка значение срока
	duration := data.BanDurations[intUserAnswer-1].Duration

	// если срок блокировки равен 0, значит это вечный бан, поле параметра должно быть пустым
	if duration == 0 {
		return "", nil
	}

	// определяем текущую дату и время
	nowDate := int(time.Now().Unix())

	// прибавляем к текущей дате срок блокировки
	unbanDate := strconv.Itoa(nowDate + duration)

	return unbanDate, nil
}

// banUser блокирует пользователя
func banUser(banInfo BanInfo) error {
	// собираем карту с параметрами для запроса к vk api
	params := map[string]string{
		"group_id": banInfo.GroupID,
		"owner_id": banInfo.UserID,
		"end_date": banInfo.UnbanDate,
		"comment":  banInfo.ReasonTitle,
		"v":        "5.101",
	}

	// получаем данные
	data, err := getData()
	if err != nil {
		return err
	}

	// отправляем запрос на блокировку пользователя в соответствии с параметрами
	_, err = sendRequestVKAPI("groups.ban", params, data.AccessToken)
	if err != nil {
		// если сервер вернул ошибку "слишком много запросов в секунду", то повторяем запрос
		if strings.Contains(err.Error(), "Too many requests per second") {
			banUser(banInfo)
		}
		return err
	}

	return nil
}
