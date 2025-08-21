package util

import (
	"fmt"
	"go-messaging/model"
	"strings"
)

func escapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

func FormatIocMessage(payload model.IocPayload) string {
	return fmt.Sprintf(
		"*🔔 New IOC Received*\n\n"+
			"*ID:* `%s`\n"+
			"*Value:* `%s`\n"+
			"*Type:* `%s`\n"+
			"*Description:* %s\n"+
			"*Case ID:* `%s`\n"+
			"*Link:* [Open in IRIS](%s)",
		escapeMarkdownV2(payload.ID),
		escapeMarkdownV2(payload.Value),
		escapeMarkdownV2(payload.Type),
		escapeMarkdownV2(payload.Description),
		escapeMarkdownV2(payload.CaseID),
		escapeMarkdownV2(payload.Link),
	)
}
