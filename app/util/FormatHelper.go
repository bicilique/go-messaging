package util

import (
	"fmt"
	"go-messaging/model"
)

func FormatIocMessage(ioc model.IocPayload) string {
	return fmt.Sprintf(
		"*🔔 New IOC Received*\n\n"+
			"*ID:* `%s`\n"+
			"*Value:* `%s`\n"+
			"*Type:* `%s`\n"+
			"*Description:* %s\n"+
			"*Case ID:* `%s`\n"+
			"*Link:* [Open in IRIS](%s)\n",
		ioc.ID,
		ioc.Value,
		ioc.Type,
		ioc.Description,
		ioc.CaseID,
		ioc.Link,
	)
}
