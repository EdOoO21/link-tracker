package stackoverflow

import (
	"fmt"
	"strings"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

const (
	previewLimit          = 200
	itemsPerTypePageLimit = 5
)

type stackOverflowUpdatesResponse struct {
	Title    string
	Answers  []stackOverflowAnswer
	Comments []stackOverflowComment
}

func (r stackOverflowUpdatesResponse) toResourceUpdate(since time.Time) ports.ResourceUpdate {
	newAnswers := r.newAnswers(since)
	newComments := r.newComments(since)
	if len(newAnswers) == 0 && len(newComments) == 0 {
		return ports.ResourceUpdate{
			HasUpdate:  false,
			LastUpdate: since,
			Message:    "",
		}
	}

	return ports.ResourceUpdate{
		HasUpdate:  true,
		LastUpdate: r.latestCreatedAt(since, newAnswers, newComments),
		Message:    r.message(newAnswers, newComments),
	}
}

func (r stackOverflowUpdatesResponse) newAnswers(since time.Time) []stackOverflowAnswer {
	filtered := make([]stackOverflowAnswer, 0, len(r.Answers))
	for _, answer := range r.Answers {
		if answer.createdAt().After(since) {
			filtered = append(filtered, answer)
		}
	}

	return filtered
}

func (r stackOverflowUpdatesResponse) newComments(since time.Time) []stackOverflowComment {
	filtered := make([]stackOverflowComment, 0, len(r.Comments))
	for _, comment := range r.Comments {
		if comment.createdAt().After(since) {
			filtered = append(filtered, comment)
		}
	}

	return filtered
}

func (r stackOverflowUpdatesResponse) latestCreatedAt(
	since time.Time,
	answers []stackOverflowAnswer,
	comments []stackOverflowComment,
) time.Time {
	latest := since
	for _, answer := range answers {
		if answer.createdAt().After(latest) {
			latest = answer.createdAt()
		}
	}
	for _, comment := range comments {
		if comment.createdAt().After(latest) {
			latest = comment.createdAt()
		}
	}

	return latest
}

func (r stackOverflowUpdatesResponse) message(answers []stackOverflowAnswer, comments []stackOverflowComment) string {
	var builder strings.Builder
	builder.WriteString("Обнаружены новые StackOverflow обновления:\n\n")
	builder.WriteString("Тема вопроса: ")
	builder.WriteString(strings.TrimSpace(r.Title))

	if len(answers) > 0 {
		builder.WriteString("\n\nОтветы:\n")
		for i, answer := range answers {
			fmt.Fprintf(&builder, "%d. Автор: %s\n", i+1, strings.TrimSpace(answer.Owner.DisplayName))
			builder.WriteString("Создано: ")
			builder.WriteString(answer.createdAt().Format(time.RFC3339))
			builder.WriteString("\nПревью: ")
			builder.WriteString(answer.preview())
			if i != len(answers)-1 {
				builder.WriteString("\n\n")
			}
		}
	}

	if len(comments) > 0 {
		builder.WriteString("\n\nКомментарии:\n")
		for i, comment := range comments {
			fmt.Fprintf(&builder, "%d. Автор: %s\n", i+1, strings.TrimSpace(comment.Owner.DisplayName))
			builder.WriteString("Создано: ")
			builder.WriteString(comment.createdAt().Format(time.RFC3339))
			builder.WriteString("\nПревью: ")
			builder.WriteString(comment.preview())
			if i != len(comments)-1 {
				builder.WriteString("\n\n")
			}
		}
	}

	return builder.String()
}

func (a stackOverflowAnswer) createdAt() time.Time {
	return time.Unix(a.CreationDate, 0).UTC()
}

func (c stackOverflowComment) createdAt() time.Time {
	return time.Unix(c.CreationDate, 0).UTC()
}

func (a stackOverflowAnswer) preview() string {
	return trimPreview(a.Body)
}

func (c stackOverflowComment) preview() string {
	return trimPreview(c.Body)
}

func trimPreview(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "отсутствует"
	}

	runes := []rune(text)
	if len(runes) <= previewLimit {
		return text
	}

	return strings.TrimSpace(string(runes[:previewLimit]))
}
