package main

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func init() {
	if os.Getenv(apiKey) == "" {
		log.Fatalf("you need to set the var %s=<your api key>\n", apiKey)
	}
}

const (
	configFile = "config.json"
	apiKey     = "SENDGRID_API_KEY" // #nosec G101
)

type Attachment struct {
	Content  string `json:"content"`
	Type     string `json:"type"`
	FileName string `json:"fileName"`
}

type Config struct {
	From           string       `json:"from"`
	FromName       string       `json:"fromName"`
	Subject        string       `json:"subject"`
	Attachments    []Attachment `json:"attachments,omitempty"`
	UseBCC         bool         `json:"useBCC,omitempty"`
	BCCPerEmail    int          `json:"bccPerEmail,omitempty"`
	RecipientsFile string       `json:"recipientsFile"`
	HtmlEmailFile  string       `json:"htmlEmailFile"`
}

func readConfig() *Config {
	b, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("could not read %s\n", configFile)
	}
	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatalln(err.Error())
	}

	if cfg.FromName == "" {
		log.Fatalln("fromName can not be empty")
	} else if cfg.From == "" {
		log.Fatalln("from can not be empty")
	} else if cfg.Subject == "" {
		log.Fatalln("subject can not be empty")
	} else if cfg.RecipientsFile == "" {
		log.Fatalln("recipientsFile can not be empty")
	} else if cfg.HtmlEmailFile == "" {
		log.Fatalln("htmlEmailFile can not be empty")
	}

	if _, err := os.Stat(cfg.RecipientsFile); err != nil {
		log.Fatalf("cant open %s\n", cfg.RecipientsFile)
	}
	if _, err := os.Stat(cfg.HtmlEmailFile); err != nil {
		log.Fatalf("can't open %s\n", cfg.HtmlEmailFile)
	}

	if cfg.Attachments != nil {
		for i, att := range cfg.Attachments {
			if att.FileName == "" || att.Type == "" {
				log.Fatalln("attachment is not configured correctly")
			}
			content, err := os.ReadFile(att.FileName)
			if err != nil {
				log.Fatalf("can not open %s for attachment\n", att.FileName)
			}
			cfg.Attachments[i].Content = base64.StdEncoding.EncodeToString(content)
		}
	}

	// don't allow multiple recipients in one email unless bcc is used, gdpr etc.
	if !cfg.UseBCC || cfg.BCCPerEmail == 0 {
		cfg.BCCPerEmail = 1
	}

	return &cfg
}

func readRecipients(recipientsFile string) []string {
	b, err := os.ReadFile(recipientsFile)
	if err != nil {
		log.Fatalf("could not read %s\n", recipientsFile)
	}

	s := string(b)
	if s == "" {
		log.Fatalln("there must be at least one recipient")
	}
	return strings.Split(s, "\n")
}

func readHtml(htmlFile string) string {
	b, err := os.ReadFile(htmlFile)
	if err != nil {
		log.Fatalf("could not read %s\n", htmlFile)
	}
	return string(b)
}

func main() {
	cfg := readConfig()
	recipients := readRecipients(cfg.RecipientsFile)
	emailHtml := readHtml(cfg.HtmlEmailFile)

	var attachments []*mail.Attachment
	if cfg.Attachments != nil {
		for _, att := range cfg.Attachments {
			a := mail.NewAttachment()
			a.Content = att.Content
			a.Type = att.Type
			a.Filename = att.FileName
			attachments = append(attachments, a)
			log.Printf("attaching: %s", att.FileName)
		}
	}

	for recipientsProcessed := 0; recipientsProcessed < len(recipients); {

		m := mail.NewV3Mail()
		m.From = mail.NewEmail(cfg.FromName, cfg.From)
		m.Subject = cfg.Subject

		mail.NewBCCSetting()

		m.AddAttachment(attachments...)

		p := mail.NewPersonalization()

		if !cfg.UseBCC {
			tos := []*mail.Email{
				mail.NewEmail("", recipients[recipientsProcessed]),
			}
			p.AddTos(tos...)
			log.Printf("sending mail to: %s", recipients[recipientsProcessed])
			recipientsProcessed++
		} else {
			// SendGrid (and I guess the SMTP protocol requires at least one to-address
			p.AddTos(mail.NewEmail(cfg.FromName, cfg.From))
			var bccs []*mail.Email
			batch := 0
			batchRecipients := strings.Builder{}
			for batch < cfg.BCCPerEmail {
				if recipientsProcessed+batch == len(recipients) {
					break
				}
				bccs = append(bccs, mail.NewEmail("", recipients[recipientsProcessed+batch]))
				batchRecipients.WriteString(recipients[recipientsProcessed+batch] + ", ")
				batch++
			}
			recipientsProcessed += batch
			log.Printf("sending mail to: %s", batchRecipients.String())
			p.AddBCCs(bccs...)
		}

		m.AddPersonalizations(p)

		m.AddContent(mail.NewContent("text/html", emailHtml))

		request := sendgrid.GetRequest(os.Getenv(apiKey), "/v3/mail/send", "https://api.sendgrid.com")
		request.Method = http.MethodPost
		request.Body = mail.GetRequestBody(m)
		response, err := sendgrid.API(request)
		if err != nil {
			log.Fatalln(err.Error())
		}
		if response.StatusCode != http.StatusAccepted {
			log.Fatalf("response from SendGrid was http %d: %s\n", response.StatusCode, response.Body)
		}

		log.Println("mail sent!")
	}
}
