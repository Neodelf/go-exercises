package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type UserXML struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

var usersXML []UserXML
var users []User
var filePath = "dataset.xml"

func SearchServer(w http.ResponseWriter, r *http.Request) {
	file, _ := os.Open(filePath)

	orderField := r.FormValue("order_field")
	if orderField == "" {
		orderField = "Name"
	} else if orderField != "Id" && orderField != "Name" && orderField != "Age" {
		w.WriteHeader(http.StatusInternalServerError)
	}

	query := r.FormValue("query")
	offset, _ := strconv.Atoi(r.FormValue("offset"))
	limit, _ := strconv.Atoi(r.FormValue("limit"))
	orderBy, _ := strconv.Atoi(r.FormValue("order_by"))

	decoder := xml.NewDecoder(file)

	if len(usersXML) == 0 {
		for {
			var userXML UserXML
			tok, tokenErr := decoder.Token()
			if tokenErr != nil && tokenErr != io.EOF {
				fmt.Println("error happend", tokenErr)
				break
			} else if tokenErr == io.EOF {
				break
			}
			if tok == nil {
				fmt.Println("t is nil break")
			}
			switch tok := tok.(type) {
			case xml.StartElement:
				if tok.Name.Local == "row" {
					if err := decoder.DecodeElement(&userXML, &tok); err != nil {
						fmt.Println("error happend", err)
					}
					usersXML = append(usersXML, userXML)
				}
			}
		}
	}

	var shortUsersXML []UserXML

	for _, userXML := range usersXML {
		if len(query) > 0 {
			if strings.Contains(userXML.FirstName, query) || strings.Contains(userXML.LastName, query) || strings.Contains(userXML.About, query) {
				shortUsersXML = append(shortUsersXML, userXML)
			}
		} else {
			shortUsersXML = append(shortUsersXML, userXML)
		}
	}

	sort.Slice(shortUsersXML, func(i, j int) bool {
		if orderField == "Name" {
			if shortUsersXML[i].FirstName < shortUsersXML[j].FirstName {
				return true
			}
			if shortUsersXML[i].FirstName > shortUsersXML[j].FirstName {
				return false
			}
			return shortUsersXML[i].LastName < shortUsersXML[j].LastName
		}
		iValue := reflect.Indirect(reflect.ValueOf(&shortUsersXML[i])).FieldByName(orderField)
		jValue := reflect.Indirect(reflect.ValueOf(&shortUsersXML[j])).FieldByName(orderField)
		if orderBy == -1 {
			return int(iValue.Int()) > int(jValue.Int())
		} else if orderBy == 1 {
			return int(iValue.Int()) < int(jValue.Int())
		}
		return true
	})

	for _, userXML := range shortUsersXML {
		users = append(users, User{
			Id:     userXML.ID,
			Name:   userXML.FirstName + " " + userXML.LastName,
			Age:    userXML.Age,
			About:  userXML.About,
			Gender: userXML.Gender,
		})
	}

	limitedOffsettedUsers := users[offset : offset+limit]

	w.WriteHeader(http.StatusOK)
	json, _ := json.Marshal(limitedOffsettedUsers)
	io.WriteString(w, string(json[:]))
}

type TestCase struct {
	SearchRequest *SearchRequest
	Result        *SearchResponse
	IsError       bool
}

func TestSearchClientFindUsers(t *testing.T) {
	cases := []TestCase{
		TestCase{
			SearchRequest: &SearchRequest{
				Limit:      1,
				Offset:     1,    // Можно учесть после сортировки
				Query:      "tt", // подстрока в 1 из полей
				OrderField: "Name",
				OrderBy:    1, // -1 по убыванию, 0 как встретилось, 1 по возрастанию
			},
			Result: &SearchResponse{
				Users: []User{
					User{
						Id:   3,
						Name: "Everett Dillard",
						Age:  27,
						About: "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris." +
							" Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			IsError: false,
		},
		TestCase{
			SearchRequest: &SearchRequest{
				OrderField: "asd",
			},
			Result:  nil,
			IsError: true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	for caseNum, caseItem := range cases {
		client := &SearchClient{
			URL:         ts.URL,
			AccessToken: "accessToken",
		}
		result, err := client.FindUsers(*caseItem.SearchRequest)

		if err != nil && !caseItem.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && caseItem.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !reflect.DeepEqual(caseItem.Result, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, caseItem.Result, result)
		}
	}
	ts.Close()
}
