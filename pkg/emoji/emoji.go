package emoji

import "fmt"

const (
	Check           = "✅"
	Cross           = "❌"
	Warning         = "⚠️"
	Rocket          = "🚀"
	Bug             = "🐛"
	Gear            = "⚙️"
	Lock            = "🔒"
	Unlock          = "🔓"
	MagnifyingGlass = "🔍"
	Trash           = "🗑️"
	Pencil          = "📝"
	Clock           = "⏰"
	Calendar        = "📅"
	Bell            = "🔔"
	Star            = "⭐"
	Fire            = "🔥"
	Heart           = "❤️"

	// Status indicators
	Success = "✅"
	Error   = "❌"
	Info    = "ℹ️"

	// Todo app specific
	Todo          = "📋"
	StatusDone    = "✅"
	StatusPending = "⏳"
	StatusExpired = "⌛"
	StatusUrgent  = "🚨"
)

func AddEmoji(emoji, text string) string {
	return fmt.Sprintf("%s %s", emoji, text)
}
