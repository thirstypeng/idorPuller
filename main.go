package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DbData struct {
	gorm.Model
	ResponseStr string
}

// I used global variables so that the variables are not copied repeatedly
var (
	url, cookie, rangeStr   string
	formData                string
	range1, range2, indexId int
	output                  string
	protocol                string
	dbMode                  bool = false
	db                      *gorm.DB
)

//go run main.go -u https://httpbin.org/get?id=[ID] -c example:123  -r 1-100
//go run main.go -u "https://httpbin.org/get?id=[ID]&arg1=1337" -c example:123  -r 1-100 -db 1
//go run main.go -u https://httpbin.org/post -c example:123 -p POST -d id=[ID] -r 1-100
//go run main.go -u https://httpbin.org/post -c example:123 -p POST -d "id=[ID]&arg=1337" -r 1-100 -db 1

func init() {
	flag.StringVar(&url, "u", "", "``Enter your url. (-u \"http://example.com?id=[ID]&arg1=xyz\")")
	flag.StringVar(&cookie, "c", "", "``Enter your cookie as raw. (-c \"token=some_token; clicked=true\")")
	flag.StringVar(&formData, "d", "", "``enter your form data in raw format (-d \"example=xyz&id=[ID]\")")
	flag.StringVar(&rangeStr, "r", "", "``Enter the id range to be attacked by idor. (-r 1-500)")
	flag.StringVar(&output, "o", "output.txt", "``Enter the name of the output file (default: output.txt)")
	flag.StringVar(&protocol, "x", "get", "``Enter get or post method (-x post Default:GET )")
	flag.StringVar(&protocol, "p", "get", "``Enter get or post method (-p post Default:GET )")

	flag.BoolVar(&dbMode, "db", false, "Will it be written to the database? Variable type is bool (-db true)")
}

/***************************************************************
*					 Parse FUNCTION						   *
***************************************************************/
func ParseRange(rangeStr string) {
	var err error

	index := strings.Index(rangeStr, "-")

	range1, err = strconv.Atoi(rangeStr[:index])
	if err != nil {
		print("Please enter the specified value range correctly!")
	}

	range2, err = strconv.Atoi(rangeStr[index+1:])
	if err != nil {
		print("Please enter the specified value range correctly!")
	}
}

func ParseUrl(baseUrl string, liveId int) string {
	returnedUrl := baseUrl[:indexId] + strconv.Itoa(liveId) + baseUrl[indexId+4:]
	return returnedUrl
}

func CalcUrlIndex(idMarker string) {
	indexId = strings.Index(idMarker, "[ID]")

	if indexId == -1 {
		print("Enter [ID] pointer to specify the id part (\"arg1=example&arg1=[ID]\")")
		os.Exit(1)
	}
}

func FormDataParser(formData string, liveId int) string {
	returnedUrl := formData[:indexId] + strconv.Itoa(liveId) + formData[indexId+4:]
	return returnedUrl
}

/***************************************************************
*					 Request FUNCTION						   *
***************************************************************/

func GetRequest(id int) string {
	req, err := http.NewRequest("GET", ParseUrl(url, id), nil)
	if err != nil {
		print("An error occurred while creating the GET request. Please make sure you have entered the information correctly. Then try again.")
		os.Exit(1)
	}

	req.Header.Add("Cookie", cookie)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		print("An error occurred in the request to the server. Please make sure you have entered the information correctly. Then try again.")
		os.Exit(1)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		print("Server gave us 200 OK external replies\n")
		print("status code:", resp.Status)
		os.Exit(1)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body)
}

func FormPostRequest(formData string) string {

	bodyReader := strings.NewReader(FormDataParser(formData, range1))

	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		print("An error occurred while creating the GET request. Please make sure you have entered the information correctly. Then try again.")
		os.Exit(1)
	}
	req.Header.Add("Cookie", cookie)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		print("An error occurred in the request to the server. Please make sure you have entered the information correctly. Then try again.")
		os.Exit(1)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		print("Server gave us 200 OK external replies\n")
		print("status code:", resp.Status)
		os.Exit(1)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body)
}

/***************************************************************
*					 Output FUNCTION						   *
***************************************************************/

// write the output in txt file
func WriteText(response string) {
	f, err := os.OpenFile(output, // predefined variable output
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}

	defer f.Close()

	f.WriteString(response + "\n=====================================================================================================\n\n")
}

// Connect to DB for write to DB
func ConnectDB() *gorm.DB {
	/*
		Check this address. You can make the appropriate database connection -
		by changing the necessary places.

		https://gorm.io/docs/connecting_to_the_database.html
	*/

	/*	PostgreSQL
		import (
			"gorm.io/driver/postgres"
		  	"gorm.io/gorm"
		)

		dsn := "host=localhost user=gorm password=gorm dbname=go	rm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	*/

	//DON'T FORGET TO ENTER DATABASE INFORMATION IN `DBNAME`
	dsn := "root:@tcp(127.0.0.1:3306)/DBNAME?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		print("DON'T FORGET TO ENTER DATABASE INFORMATION IN `DBNAME`")
		print("Make sure you make the DB connection correctly! \n You can check the ConnectDB() function in the source code.")
	}

	db.AutoMigrate(&DbData{})
	return db
}

// write the output in the DB
func writeDB(response string) {
	db.Create(&DbData{ResponseStr: response})
}

func main() {

	flag.Parse()
	ParseRange(rangeStr)
	protocol = strings.ToUpper(protocol)

	if protocol == "GET" {
		CalcUrlIndex(url)

		if dbMode == false {
			for range1 < range2 {
				WriteText(GetRequest(range1))
				range1++
			}

		} else {
			db = ConnectDB()

			for range1 < range2 {
				writeDB(GetRequest(range1))
				range1++
			}
		}
	}

	if protocol == "POST" {
		CalcUrlIndex(formData)

		if dbMode == false {
			for range1 < range2 {
				WriteText(FormPostRequest(formData))
				range1++
			}
		} else {
			db = ConnectDB()

			for range1 < range2 {
				writeDB(FormPostRequest(formData))
				range1++
			}
		}
	}

}
