package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	govkapi "github.com/VitJRBOG/GoVkApi/v2"
	"io/ioutil"
	"log"
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
	AccessToken string       `json:"access_token"`
	Groups      []Groups     `json:"groups"`
	BanReasons  []BanReasons `json:"ban_reasons"`
}

// Groups хранит информацию о пабликах
type Groups struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// BanReasons хранит информацию о сроках и причинах блокировки
type BanReasons struct {
	Reason        string `json:"reason"`
	DurationTitle string `json:"duration_title"`
	Duration      int    `json:"duration"`
}

func main() {
	err := initialization()
	if err != nil {
		log.Fatalln(err)
	}

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
	var reasonIndex int
	reasonIndex, banInfo.ReasonTitle, err = selectReason()
	if err != nil {
		fmt.Println(fmt.Errorf("%v", err))
		os.Exit(0)
	}
	fmt.Println("> Gotcha reason title of ban.")

	// выбираем дату автоматической разблокировки
	banInfo.UnbanDate, err = getUnbanDate(reasonIndex)
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

// initialization проверяет наличие ресурсных файлов и создает их, если отсутствуют
func initialization() error {

	// проверяем наличие файла с путем к файлу с данными
	if _, err := os.Stat("path.txt"); os.IsNotExist(err) {

		// если отсутствует, то создаем новый
		valuesBytes := []byte("")
		err = ioutil.WriteFile("path.txt", valuesBytes, 0644)
		if err != nil {
			return err
		}
		fmt.Println("COMPUTER [Initialization]: File \"path.txt\" has been created.")
	}

	// получаем путь к файлу с данными
	path, err := readPath()
	if err != nil {
		return err
	}

	// проверяем наличие файла с данными
	if _, err := os.Stat(path + "data.json"); os.IsNotExist(err) {

		// если отсутствует, то создаем новый
		var data Data

		// заполняем список групп
		groupData, err := getGroupData()
		if err != nil {
			return err
		}
		data.Groups = append(data.Groups, groupData)

		// заполняем список типов банов
		banReason, err := getBanReasonsData()
		if err != nil {
			return err
		}
		data.BanReasons = append(data.BanReasons, banReason)

		// формируем массив байт с данными
		valuesBytes, err := json.Marshal(data)
		if err != nil {
			return err
		}

		// записываем файл
		err = writeJSON(path+"data.json", valuesBytes)
		if err != nil {
			return err
		}
		fmt.Println("COMPUTER [Initialization]: File \"data.json\" has been created.")
	}

	return nil
}

// getGroupData принимает данные о группе
func getGroupData() (Groups, error) {
	fmt.Println("COMPUTER [Initialization]: Data of groups not found. Need new data.")
	var groupData Groups

	fmt.Print("> [Group name]: ")
	// принимаем ввод названия группы
	var userAnswer string
	_, err := fmt.Scan(&userAnswer)
	if err != nil {
		return groupData, err
	}
	groupData.Name = userAnswer

	fmt.Print("> [Group ID]: ")
	// принимаем ввод идентификатора группы
	_, err = fmt.Scan(&userAnswer)
	if err != nil {
		return groupData, err
	}
	groupData.ID = userAnswer

	return groupData, nil
}

func getBanReasonsData() (BanReasons, error) {
	fmt.Println("COMPUTER [Initialization]: Data of bans not found. Need new data.")
	var banReason BanReasons

	fmt.Print("> [Reason]: ")
	// принимаем комментарий для бана
	var userAnswer string
	scnr := bufio.NewScanner(os.Stdin)
	scnr.Scan()
	userAnswer = scnr.Text()
	banReason.Reason = userAnswer

	durationTitles := [5]string{"Day", "Week", "Month", "Year", "End of the year"}
	// оглашаем весь список
	for i, durationTitle := range durationTitles {
		fmt.Printf("> [Duration titles]: %d == %v\n", i+1, durationTitle)
	}
	fmt.Print("> [Duration title]: ")

	// запрашиваем номер срока
	_, err := fmt.Scan(&userAnswer)
	if err != nil {
		return banReason, err
	}

	// конвертируем строку с номером срока в целочисленный тип
	intUserAnswer, err := strconv.Atoi(userAnswer)
	if err != nil {
		return banReason, err
	}

	// получаем название срока по его номеру
	banReason.DurationTitle = durationTitles[intUserAnswer-1]

	durations := [5]int{86400, 604800, 2629743, 31556926, 0}

	// получаем длительность срока по номеру его названия
	banReason.Duration = durations[intUserAnswer-1]

	return banReason, nil
}

// getAccessToken получает новый токен доступа
func getAccessToken() error {
	fmt.Print("> [New access token]: ")

	// принимаем ввод нового токена
	var userAnswer string
	_, err := fmt.Scan(&userAnswer)
	if err != nil {
		return err
	}

	// получаем путь к файлу с шаблонами
	pathToFile, err := readPath()
	if err != nil {
		return err
	}

	// получаем из json файла данные
	rawData, err := readJSON(pathToFile + "data.json")
	if err != nil {
		return err
	}

	// конвертируем их в читабельный вид
	var data Data
	err = json.Unmarshal(rawData, &data)

	// перезаписываем токен доступа
	data.AccessToken = userAnswer

	// формируем массив байт с шаблонами
	valuesBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatalln(err)
	}

	// записываем его в файл
	err = ioutil.WriteFile(pathToFile+"data.json", valuesBytes, 0644)

	return nil
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
	resp, err := govkapi.Method("users.get", data.AccessToken, params)
	if err != nil {

		switch true {
		// если сервер вернул ошибку "слишком много запросов в секунду", то повторяем запрос
		case strings.Contains(err.Error(), "many requests per second"):
			return getUserID(userURL)
		// если сервер вернул ошибку "токен устарел", то просим у пользователя новый токен
		case strings.Contains(err.Error(), "access_token has expired"):
			if err := getAccessToken(); err != nil {
				return "", "", err
			}
			return getUserID(userURL)
		// если сервер вернул ошибку "токен был получен для другого ip", то просим у пользователя новый токен
		case strings.Contains(err.Error(), "access_token was given to another ip address"):
			if err := getAccessToken(); err != nil {
				return "", "", err
			}
			return getUserID(userURL)
		// если сервер вернул ошибку "токен невалидный", то просим у пользователя новый токен
		case strings.Contains(err.Error(), "invalid access_token"):
			if err := getAccessToken(); err != nil {
				return "", "", err
			}
			return getUserID(userURL)
		// если сервер вернул ошибку "токен не введен", то просим у пользователя новый токен
		case strings.Contains(err.Error(), "no access_token passed"):
			if err := getAccessToken(); err != nil {
				return "", "", err
			}
			return getUserID(userURL)
		default:
			return "", "", err
		}

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
func selectReason() (int, string, error) {

	// получаем данные
	data, err := getData()
	if err != nil {
		return 0, "", err
	}

	// оглашаем весь список
	for i, reason := range data.BanReasons {
		fmt.Printf("> [Reasons]: %d == %v (%v)\n", i+1, reason.Reason, reason.DurationTitle)
	}
	fmt.Println("> [Reasons]: 00 == Quit")

	// запрашиваем номер причины
	var userAnswer string
	_, err = fmt.Scan(&userAnswer)
	if err != nil {
		return 0, "", err
	}

	// если пользователь ввел 00, то завершаем программу
	if userAnswer == "00" {
		fmt.Println("> Quit...")
		os.Exit(0)
	}

	// конвертируем строку с номером причины в целочисленный тип
	intUserAnswer, err := strconv.Atoi(userAnswer)
	if err != nil {
		return 0, "", err
	}

	reasonIndex := intUserAnswer - 1

	return reasonIndex, data.BanReasons[reasonIndex].Reason, nil
}

// getUnbanDate получаем дату разблокировки
func getUnbanDate(reasonIndex int) (string, error) {

	// получаем данные
	data, err := getData()
	if err != nil {
		return "", err
	}

	var unbanDate string

	// если блокировка "до конца года", то определяем дату последнего дня текущего года
	if data.BanReasons[reasonIndex].DurationTitle == "End of the year" {

		// определяем текущую дату и время
		nowDate := time.Now()

		// собираем дату последнего дня года
		lastYearsDay := time.Date(nowDate.Year(), 12, 31, 23, 59, 59, 0, nowDate.Location())

		// преобразуем полученную дату в unixtime
		unbanDate = strconv.Itoa(int(lastYearsDay.Unix()))
	} else {

		// получаем из списка значение срока для данной причины блокировки
		duration := data.BanReasons[reasonIndex].Duration

		// если срок блокировки равен 0, значит это вечный бан, поле параметра должно быть пустым
		if duration == 0 {
			return "", nil
		}

		// определяем текущую дату и время
		nowDate := int(time.Now().Unix())

		// прибавляем к текущей дате срок блокировки
		unbanDate = strconv.Itoa(nowDate + duration)
	}

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
	_, err = govkapi.Method("groups.ban", data.AccessToken, params)
	if err != nil {
		// если сервер вернул ошибку "слишком много запросов в секунду", то повторяем запрос
		if strings.Contains(err.Error(), "many requests per second") {
			return banUser(banInfo)
		}
		return err
	}

	return nil
}
