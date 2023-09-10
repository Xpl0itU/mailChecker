package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
	}

	filtersContents, err := os.ReadFile("filters.json")
	if err != nil {
		log.Fatalln(err)
	}
	var filters []MailFilter
	if err := json.Unmarshal(filtersContents, &filters); err != nil {
		log.Fatalln(err)
	}

	if err := checkEmailsWithFilters(filters, os.Getenv("SERVER"), os.Getenv("EMAIL"), os.Getenv("PASSWORD"), os.Getenv("MAIL_OK_FOLDER"), os.Getenv("MAIL_FAILED_FOLDER")); err != nil {
		log.Fatalln(err)
	}
}

type MailFilter struct {
	Mail           string `json:"mail"`
	Subject        string `json:"subject"`
	FailIfFound    bool   `json:"fail_if_found"`
	HourThreshold  int    `json:"hour_threshold"`
	Comment        string `json:"comment"`
	FailIfNotFound bool   `json:"fail_if_not_found"`
}

func checkEmailsWithFilters(filters []MailFilter, server, email, password, mailOkFolder, mailFailedFolder string) error {
	anyErrors := false
	c, err := connectToIMAP(server, email, password)
	if err != nil {
		return err
	}
	defer c.Logout()

	c.Create(mailOkFolder)
	c.Create(mailFailedFolder)

	for _, filter := range filters {
		messages, err := searchEmails(c, filter)
		if err != nil {
			return err
		}

		if messages.Empty() && filter.FailIfNotFound {
			log.Printf("Not found: %+v\n", filter)
			anyErrors = true
			continue
		}

		if filter.FailIfFound {
			log.Printf("Error: %+v\n", filter)
		}

		if filter.FailIfFound {
			if err := moveMessages(c, messages, mailFailedFolder); err != nil {
				log.Println(err)
			} else {
				log.Printf("Moved messages to %s\n", mailFailedFolder)
			}
			anyErrors = true
		} else {
			if err := moveMessages(c, messages, mailOkFolder); err != nil {
				log.Println(err)
			} else {
				log.Printf("Moved messages to %s\n", mailOkFolder)
			}
		}
	}

	if anyErrors {
		return errors.New("found errors")
	}
	return nil
}

func connectToIMAP(server, email, password string) (*client.Client, error) {
	c, err := client.DialTLS(server, nil)
	if err != nil {
		return nil, err
	}

	if err := c.Login(email, password); err != nil {
		return nil, err
	}

	return c, nil
}

func searchEmails(c *client.Client, filter MailFilter) (*imap.SeqSet, error) {
	_, err := c.Select("INBOX", false)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	rangeStart := now.Add(-time.Duration(filter.HourThreshold) * time.Hour)

	criteria := imap.NewSearchCriteria()
	criteria.Header.Add("FROM", filter.Mail)
	criteria.Header.Add("SUBJECT", filter.Subject)
	criteria.Since = rangeStart

	ids, err := c.Search(criteria)
	if err != nil {
		return nil, err
	}

	messages := new(imap.SeqSet)
	messages.AddNum(ids...)

	return messages, nil
}

func moveMessages(c *client.Client, messages *imap.SeqSet, folderName string) error {
	err := c.Move(messages, folderName)
	if err != nil {
		return err
	}
	return nil
}
