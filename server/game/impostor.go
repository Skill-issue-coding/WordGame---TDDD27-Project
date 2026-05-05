package game

type ImpostorSettings struct {
	InputDuration      int `json:"inputDuration"`      // seconds to submit word
	DiscussionDuration int `json:"discussionDuration"` // seconds for discussion/voting
	ImpostorCount      int `json:"impostorCount"`
}

func DefaultImpostorSettings() ImpostorSettings {
	return ImpostorSettings{
		InputDuration:      2,
		DiscussionDuration: 15,
		ImpostorCount:      1,
	}
}
