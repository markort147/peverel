package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
	"time"

	dt "github.com/markor147/peverel/internal/data"
	"github.com/markor147/peverel/internal/log"
	gomail "gopkg.in/mail.v2"
)

//go:embed email.tmpl
var emailTmpl string

func main() {
	// Log initialisation
	logLevel := os.Getenv("LOG_LEVEL")
	logOutput := os.Getenv("LOG_OUTPUT")
	logHeader := "${time_rfc3339} ${short_file}:${line} ${level} ${message}"
	closer, err := log.Init(logLevel, logOutput, logHeader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v", err)
		os.Exit(1)
	}
	if closer != nil {
		defer closer.Close()
	}

	// Init data service
	connStr := os.Getenv("DB_CONN_STRING")
	if err := dt.Init(connStr); err != nil {
		log.Logger.Fatal(err)
	}

	// Init template
	tmpl, err := template.New("templates").Parse(emailTmpl)
	if err != nil {
		log.Logger.Fatalf("parse templates: %v", err)
	}

	// Set up the scheduler
	emailSender := os.Getenv("EMAIL_SENDER")
	emailRecipients := strings.Split(os.Getenv("EMAIL_RECIPIENTS"), ",")
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	scheduledTime := os.Getenv("SCHEDULED_TIME")
	smtpPort, err := strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		log.Logger.Fatalf("parse SMTP_PORT: %v", err)
	}
	scheduledHours, err := strconv.Atoi(os.Getenv("SCHEDULED_HOURS"))
	if err != nil {
		log.Logger.Fatalf("parse SCHEDULED_HOURS: %v", err)
	}
	if scheduledTime != "" {
		// Parse the scheduled time
		parsedTime, _ := time.Parse("15:04-07", scheduledTime)
		now := time.Now()
		schedule := time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, now.Location())

		log.Logger.Infof("Service started with scheduled time: %s. Now is %s.", schedule.Format("15:04"), now.Format("15:04"))

		// Ff the scheduled time is today, add 24 hours to it
		initialDuration := schedule.Sub(now)
		if initialDuration < 0 {
			initialDuration += time.Duration(scheduledHours) * time.Hour
		}
		// Send the email after the initial duration
		log.Logger.Infof("Waiting for next tick: %f mins", initialDuration.Minutes())
		time.Sleep(initialDuration)
		sendEmail(tmpl, emailSender, emailRecipients, smtpServer, smtpPort, smtpUsername, smtpPassword)
		log.Logger.Infof("Waiting for next tick")

		// Set up a ticker to send the email every scheduledHours hours
		ticker := time.NewTicker(time.Duration(scheduledHours) * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			sendEmail(tmpl, emailSender, emailRecipients, smtpServer, smtpPort, smtpUsername, smtpPassword)
			log.Logger.Infof("Waiting for next tick")
		}
	} else {
		// If the scheduled time is not set, send the email immediately
		sendEmail(tmpl, emailSender, emailRecipients, smtpServer, smtpPort, smtpUsername, smtpPassword)
	}
}

func sendEmail(tmpl *template.Template, emailSender string, emailRecipients []string, smtpServer string, smtpPort int, smtpUsername string, smtpPassword string) {
	// Fet the expired tasks
	expiredTasks, err := dt.Tasks("", "0", true)
	if err != nil {
		log.Logger.Errorf("get expired tasks: %v", err)
		return
	}

	if len(expiredTasks) == 0 {
		// No expired tasks, do not send an email
		log.Logger.Infof("No expired tasks found")
		return
	} else {
		// Build the tasks list
		tasks := make([]map[string]string, 0)
		for _, task := range expiredTasks {
			group, _ := dt.GetTaskGroupName(task.Id)
			tasks = append(tasks, map[string]string{
				"Name":        task.Name,
				"Description": task.Description,
				"Group":       group,
			})
		}

		// Execute the email body template
		emailBodyBuilder := &strings.Builder{}
		if err := tmpl.ExecuteTemplate(emailBodyBuilder, "email", map[string]any{
			"Tasks": tasks,
			"Count": len(expiredTasks),
		}); err != nil {
			log.Logger.Errorf("Error executing template: %v", err)
			return
		}

		// Set up the email message
		message := gomail.NewMessage()
		message.SetHeader("From", emailSender)
		message.SetHeader("To", emailRecipients...)
		message.SetHeader("Subject", fmt.Sprintf("Peverel has something for you: %d expired tasks", len(expiredTasks)))
		message.SetBody("text/html", emailBodyBuilder.String())

		// Send the email
		dialer := gomail.NewDialer(smtpServer, smtpPort, smtpUsername, smtpPassword)
		if err := dialer.DialAndSend(message); err != nil {
			log.Logger.Errorf("Error sending email: %v", err)
		} else {
			log.Logger.Infof("Email sent succesfully")
		}
	}
}
