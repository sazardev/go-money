package gmail

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/sazardev/go-money/internal/auth"
	"github.com/sazardev/go-money/internal/models"
	"golang.org/x/oauth2"
	gmail "google.golang.org/api/gmail/v1"
)

type GmailService struct {
	service *gmail.Service
}

// NewGmailService creates a new Gmail service instance
func NewGmailService(ctx context.Context, token *oauth2.Token) (*GmailService, error) {
	authenticator := auth.NewAuthenticator()
	client := authenticator.GetHTTPClient(ctx, token)

	service, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to create Gmail service: %v", err)
		return nil, err
	}

	return &GmailService{service: service}, nil
}

// GetMessages retrieves messages from Gmail with optional query
func (gs *GmailService) GetMessages(ctx context.Context, query string) ([]*models.Message, error) {
	var messages []*models.Message

	// List messages
	call := gs.service.Users.Messages.List("me")
	if query != "" {
		call = call.Q(query)
	}
	call = call.MaxResults(100)

	results, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve messages: %v", err)
	}

	if len(results.Messages) == 0 {
		log.Println("No messages found.")
		return messages, nil
	}

	// Get full message details
	for _, message := range results.Messages {
		msg, err := gs.GetMessage(ctx, message.Id)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetMessage retrieves a single message with full details
func (gs *GmailService) GetMessage(ctx context.Context, msgID string) (*models.Message, error) {
	message, err := gs.service.Users.Messages.Get("me", msgID).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve message: %v", err)
	}

	msg := &models.Message{
		ID:       message.Id,
		ThreadID: message.ThreadId,
		Date:     time.Now(),
	}

	// Parse headers
	for _, header := range message.Payload.Headers {
		switch header.Name {
		case "From":
			msg.From = header.Value
		case "To":
			msg.To = header.Value
		case "Subject":
			msg.Subject = header.Value
		case "Date":
			msg.Date = parseDate(header.Value)
		}
	}

	// Get body
	if message.Payload.Parts != nil {
		for _, part := range message.Payload.Parts {
			if part.MimeType == "text/plain" || part.MimeType == "text/html" {
				if part.Body != nil && part.Body.Data != "" {
					msg.Body = decodeBase64(part.Body.Data)
					break
				}
			}
		}
	} else if message.Payload.Body != nil && message.Payload.Body.Data != "" {
		msg.Body = decodeBase64(message.Payload.Body.Data)
	}

	// Get labels
	msg.Labels = message.LabelIds

	return msg, nil
}

// SearchMessages searches for messages using a query
func (gs *GmailService) SearchMessages(ctx context.Context, query string) ([]*models.Message, error) {
	return gs.GetMessages(ctx, query)
}

// parseDate parses email date header
func parseDate(dateStr string) time.Time {
	// Try RFC822 format first
	t, err := time.Parse(time.RFC822, dateStr)
	if err == nil {
		return t
	}
	return time.Now()
}

// decodeBase64 decodes base64 encoded data
func decodeBase64(encoded string) string {
	// TODO: Implement proper base64 decoding
	return encoded
}

// GetMessagesFromSender retrieves messages from a specific sender
func (gs *GmailService) GetMessagesFromSender(ctx context.Context, sender string) ([]*models.Message, error) {
	query := fmt.Sprintf("from:%s", sender)
	return gs.GetMessages(ctx, query)
}

// GetMessagesWithLabel retrieves messages with a specific label
func (gs *GmailService) GetMessagesWithLabel(ctx context.Context, label string) ([]*models.Message, error) {
	var messages []*models.Message

	// Get label ID
	labels, err := gs.service.Users.Labels.List("me").Do()
	if err != nil {
		return nil, err
	}

	var labelID string
	for _, l := range labels.Labels {
		if strings.EqualFold(l.Name, label) {
			labelID = l.Id
			break
		}
	}

	if labelID == "" {
		return messages, nil
	}

	// List messages with label
	call := gs.service.Users.Messages.List("me")
	call = call.LabelIds(labelID)
	call = call.MaxResults(100)

	results, err := call.Do()
	if err != nil {
		return nil, err
	}

	// Get full message details
	for _, message := range results.Messages {
		msg, err := gs.GetMessage(ctx, message.Id)
		if err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
