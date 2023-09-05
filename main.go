package main

import (
	"encoding/json"
	"fmt"
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
		log.Fatal("Error loading .env file")
	}

	filtersContents, err := os.ReadFile("filters.json")
	if err != nil {
		log.Fatal(err)
	}
	var filters []MailFilter
	if err := json.Unmarshal(filtersContents, &filters); err != nil {
		log.Fatal(err)
	}

	if err := checkEmailsWithFilters(filters, os.Getenv("SERVER"), os.Getenv("EMAIL"), os.Getenv("PASSWORD"), os.Getenv("MAIL_OK_FOLDER"), os.Getenv("MAIL_FAILED_FOLDER")); err != nil {
		log.Fatal(err)
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

		if len(messages) == 0 && filter.FailIfNotFound {
			fmt.Printf("Not found: %+v\n", filter)
			continue
		}

		for _, msg := range messages {
			if filter.FailIfFound {
				fmt.Printf("Error: %+v\n", filter)
				moveMessage(c, msg, mailFailedFolder)
			} else {
				moveMessage(c, msg, mailOkFolder)
			}
		}
	}

	return nil
}

func connectToIMAP(server, email, password string) (*client.Client, error) {
	c, err := client.Dial(server)
	if err != nil {
		return nil, err
	}

	if err := c.Login(email, password); err != nil {
		return nil, err
	}

	return c, nil
}

func searchEmails(c *client.Client, filter MailFilter) ([]*imap.Message, error) {
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

	messagesChan := make(chan *imap.Message)
	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	go func() {
		if err := c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messagesChan); err != nil {
			log.Println("Error fetching messages:", err)
		}
	}()

	var messages []*imap.Message
	for msg := range messagesChan {
		if msg != nil {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

func moveMessage(c *client.Client, msg *imap.Message, folderName string) error {
	set := new(imap.SeqSet)
	set.AddNum(msg.SeqNum)
	err := c.Move(set, folderName)
	if err != nil {
		return err
	}
	fmt.Printf("Moved message to %s: %s\n", folderName, msg.Envelope.Subject)
	return nil
}
