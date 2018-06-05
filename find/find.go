package find

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"regexp"
)

const (
	// gmail credentials
	gmailUsername string = "yfekiopt@gmail.com"
	gmailPassword string = "FfF2w3uf"
	gmailHost     string = "smtp.gmail.com"
	gmailPort     int    = 587

	//REGEX
	regexPattern string = `(?s:<div class="ad">.+?title="(.*?)".+?href="(.+?)".+?<div class="price"><span>(\d+))`
)

type (
	searchStruct struct {
		link  string
		name  string
		email string
	}
	foundStruct struct {
		Link  string
		Title string
		Price string
	}
	foundMap       map[string][]foundStruct
	templateStruct struct {
		FromAddress string
		ToAddress   string
		NumFound    int
		StructAdded []foundStruct
	}
)

func connectToDB() (*sql.DB, error) {
	conn, err := sql.Open(driver, database)
	if err != nil {
		return nil, fmt.Errorf("cannot conenct to the database")
	}

	return conn, nil
}

func fetchFromDB(conn *sql.DB) ([]searchStruct, error) {
	var dbData []searchStruct

	rows, err := conn.Query("SELECT link, name, email FROM search")
	if err != nil {
		return nil, fmt.Errorf("query failed")
	}
	defer rows.Close()

	for rows.Next() {
		tmp := searchStruct{}
		if err := rows.Scan(&tmp.link, &tmp.name, &tmp.email); err != nil {
			return nil, fmt.Errorf("query scan failed")
		}

		dbData = append(dbData, tmp)
	}

	return dbData, nil
}

func scrapeWebsite(rows []searchStruct) (foundMap, error) {
	r := regexp.MustCompile(regexPattern)

	found := make(foundMap)
	for _, row := range rows {
		req, err := http.NewRequest("GET", row.link, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot make a new request")
		}

		resp, err := http.DefaultTransport.RoundTrip(req)
		if err != nil {
			return nil, fmt.Errorf("could not scrape the link")
		}
		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read the content")
		}

		matches := r.FindAllSubmatch(respBody, -1)
		for _, match := range matches {

			if len(match) < 4 {
				return nil, fmt.Errorf("not enough matches")
			}

			found[row.email] = append(found[row.email], foundStruct{
				Title: string(match[1]),
				Link:  string(match[2]),
				Price: string(match[3]),
			})
		}
	}

	return found, nil
}

func addToDB(conn *sql.DB, found foundMap) (foundMap, error) {
	added := make(foundMap)

	stmt, err := conn.Prepare("INSERT INTO found(link, title, price) values(?,?,?)")
	if err != nil {
		return nil, fmt.Errorf("could not prepare the statement")
	}

	trans, err := conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("sql begin failed")
	}

	for email, items := range found {
		for _, item := range items {
			if _, err := trans.Stmt(stmt).Exec(item.Link, item.Title, item.Price); err != nil {
				continue
			}
			added[email] = append(added[email], item)
		}
	}

	if err := trans.Commit(); err != nil {
		return nil, err
	}

	return added, nil
}

func send(added foundMap) error {
	templateStr :=
		`From: Bolha Scraper <{{ .FromAddress }}>
To: {{ .ToAddress }}
MIME-Version: 1.0
Content-type: text/html; charset=UTF-8
Subject: {{ .NumFound }} new item(s) found!

{{ range $key, $item := .StructAdded }}({{ $item.Price }}€) <a href="http://www.bolha.com/{{ $item.Link }}">{{ $item.Title }}</a><br>{{ end }}
`

	templateParsed, err := template.New("emailTemplate").Parse(templateStr)
	if err != nil {
		return fmt.Errorf("could not parse template")
	}

	auth := smtp.PlainAuth("", gmailUsername, gmailPassword, gmailHost)

	for email, items := range added {
		if len(items) == 0 {
			continue
		}

		emailTemplate := templateStruct{
			FromAddress: gmailUsername,
			ToAddress:   email,
			NumFound:    len(items),
			StructAdded: items,
		}

		var html bytes.Buffer
		templateParsed.Execute(&html, emailTemplate)

		if err := smtp.SendMail(fmt.Sprintf("%s:%d", gmailHost, gmailPort), auth, gmailUsername, []string{emailTemplate.ToAddress}, html.Bytes()); err != nil {
			return fmt.Errorf("error sending to %s: %s\n", emailTemplate.ToAddress, err)
		}
	}

	return nil
}

func main() {
	conn, err := connectToDB()
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	dbData, err := fetchFromDB(conn)
	if err != nil {
		log.Fatalln(err)
	}

	found, err := scrapeWebsite(dbData)
	if err != nil {
		log.Fatalln(err)
	}

	added, err := addToDB(conn, found)
	if err != nil {
		log.Fatalln(err)
	}

	if err := send(added); err != nil {
		log.Fatalln(err)
	}
}
