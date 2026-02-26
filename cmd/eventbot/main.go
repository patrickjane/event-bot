package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var botToken string
var channelID string
var pollRate time.Duration = 30 * time.Second
var version = ""

var (
	// Load CET Location globally
	cetLocation     *time.Location
	reminderOffsets = []time.Duration{
		24 * time.Hour,
		2 * time.Hour,
		15 * time.Minute,
	}
)

type Reminder struct {
	EventID   string
	EventName string
	EventURL  string
	StartTime time.Time // The actual 24h start time
	RemindAt  time.Time // When the bot should post the message
}

type ReminderStore struct {
	sync.Mutex
	Pending []Reminder
}

var store = &ReminderStore{Pending: []Reminder{}}

func init() {
	// Initialize the timezone during startup
	var err error

	cetLocation, err = time.LoadLocation("Europe/Berlin") // "Europe/Berlin" is the standard TZ database name for CET/CEST

	if err != nil {
		log.Printf("Warning: Could not load CET location, falling back to local: %v", err)
		cetLocation = time.Local
	}
}

func main() {
	var tme int

	readString("DISCORD_BOT_TOKEN", &botToken, "")
	readString("DISCORD_CHANNEL_ID", &channelID, "")
	readInt("DISCORD_POLL_RATE", &tme, "30")

	pollRate = time.Duration(tme) * time.Second

	fmt.Printf("EventBot %s\n", version)
	fmt.Println("https://github.com/patrickjane/event-bot")

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatal(err)
	}

	dg.AddHandler(guildScheduledEventCreate)
	dg.Identify.Intents = discordgo.IntentsGuildScheduledEvents | discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		log.Fatal(err)
	}

	syncExistingEvents(dg)
	go reminderWorker(dg)

	fmt.Println("Bot is running. Monitoring events...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func syncExistingEvents(s *discordgo.Session) {
	for _, guild := range s.State.Guilds {
		events, err := s.GuildScheduledEvents(guild.ID, false)

		if err != nil {
			continue
		}

		for _, event := range events {
			cetTime := event.ScheduledStartTime.In(cetLocation)

			fmt.Printf("Found pending event '%s' at %s\n", event.Name, cetTime.Format("02.01. 15:04"))

			queueReminders(event)
		}
	}

	fmt.Printf("Sync complete. %d reminders in queue.\n", len(store.Pending))
}

func guildScheduledEventCreate(s *discordgo.Session, e *discordgo.GuildScheduledEventCreate) {
	event := e.GuildScheduledEvent
	eventURL := fmt.Sprintf("https://discord.com/events/%s/%s", event.GuildID, event.ID)

	cetTime := event.ScheduledStartTime.In(cetLocation)

	fmt.Printf("New event '%s' at %s has been created in discord, scheduling reminders and posting notification\n",
		event.Name, cetTime.Format("02.01. 15:04"))

	// Post the creation announcement
	msg := fmt.Sprintf("📢 **Neues Event wurde erstellt!** @everyone\n\n%s", eventURL)
	s.ChannelMessageSend(channelID, msg)

	queueReminders(event)
}

func queueReminders(event *discordgo.GuildScheduledEvent) {
	store.Lock()
	defer store.Unlock()

	eventURL := fmt.Sprintf("https://discord.com/events/%s/%s", event.GuildID, event.ID)

	for _, offset := range reminderOffsets {
		remindTime := event.ScheduledStartTime.Add(-offset)

		if time.Now().Before(remindTime) {
			store.Pending = append(store.Pending, Reminder{
				EventID:   event.ID,
				EventName: event.Name,
				EventURL:  eventURL,
				StartTime: event.ScheduledStartTime, // Store the fixed start time
				RemindAt:  remindTime,
			})

			cetTime := remindTime.In(cetLocation)

			fmt.Printf("   Scheduling reminder for event '%s' at %s\n", event.Name, cetTime.Format("02.01. 15:04"))
		}
	}
}

func reminderWorker(s *discordgo.Session) {
	ticker := time.NewTicker(time.Duration(pollRate))
	for range ticker.C {
		now := time.Now()
		store.Lock()

		var remaining []Reminder

		fmt.Printf("Checking %d reminders:\n", len(store.Pending))

		for _, r := range store.Pending {
			cetTime := r.RemindAt.In(cetLocation)

			fmt.Printf("   Event '%s' reminder due at: %s\n", r.EventName, cetTime.Format("02.01. 15:04"))

			if now.After(r.RemindAt) {
				cetTime := r.StartTime.In(cetLocation)
				timeStr := cetTime.Format("15:04")
				dateStr := cetTime.Format("02.01.")

				msg := fmt.Sprintf("⏰ **Reminder!** @everyone\n\nEvent '%s' startet am %s um %s!\n\n%s",
					r.EventName, dateStr, timeStr, r.EventURL)

				s.ChannelMessageSend(channelID, msg)

			} else {
				remaining = append(remaining, r)
			}
		}
		store.Pending = remaining
		store.Unlock()
	}
}

func readString(name string, target *string, defaultVal string) {
	value := os.Getenv(name)

	if value == "" {
		if defaultVal != "" {
			value = defaultVal
		} else {
			slog.Error(fmt.Sprintf("Missing env variable %s", name))
			os.Exit(1)
		}
	}

	if target == nil {
		slog.Error(fmt.Sprintf("Target for env variable %s is nil", name))
		os.Exit(1)
	}

	*target = value
}

func readInt(name string, target *int, defaultVal string) {
	var strVal string

	readString(name, &strVal, defaultVal)

	i, err := strconv.Atoi(strVal)

	if err != nil {
		slog.Error(fmt.Sprintf("Value for env variable %s is not a valid number: %s", name, strVal))
		os.Exit(1)
	}

	if target == nil {
		slog.Error(fmt.Sprintf("Target for env variable %s is nil", name))
		os.Exit(1)
	}

	*target = i
}
