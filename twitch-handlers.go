package main

import (
	"fmt"
	"log"
	"strings"

	twitch "github.com/gempir/go-twitch-irc/v4"
)

func twitchlistinghandlers(client *twitch.Client) handlerf {
	lh := listinghandlers{
		Client: client,
	}

	// ih.add allows you to just say what the command is and what to show
	lh.add("!price", func(c Listing) string {
		return fmt.Sprintf(
			"Buy now: %v | Sell now: %v",
			c.BestSellPrice, c.BestBuyPrice,
		)
	})

	return lh.handle
}

func twitchitemhandlers(client *twitch.Client) handlerf {
	ih := itemhandlers{
		Client: client,
	}

	// ih.add allows you to just say what the contact is and what to show
	ih.add("!contact", func(c MLBCard) string {
		return fmt.Sprintf(
			"Con. vs left: %v | Con. vs right: %v",
			c.ContactLeft, c.ContactRight,
		)
	})

	ih.add("!power", func(c MLBCard) string {
		return fmt.Sprintf(
			"Pow. vs left: %v | Pow. vs right: %v",
			c.PowerLeft, c.PowerRight,
		)
	})

	ih.add("!theme", func(c MLBCard) string {
		capts := findcaptains(c)
		names := make([]string, len(capts))
		for i, v := range capts {
			names[i] = v.Name
		}

		return fmt.Sprintf(
			"%v captains: %s",
			len(names), strings.Join(names, ", "),
		)
	})

	// same as theme
	// add a generic one (not a 1 card : 1 response)
	ih.handlers.add(itemcommand("!captains", func(m msg, c MLBCard) bool {
		capts := findcaptains(c)
		names := make([]string, len(capts))
		for i, v := range capts {
			names[i] = v.Name
		}

		client.Say(m.Channel, fmt.Sprintf(
			"%s | %v captains: %s",
			descr(c), len(names), strings.Join(names, ", "),
		))

		return true
	}))

	return ih.handle
}

func pmhandler(client *twitch.Client) func(m msg) {
	mh := &handlers{}

	// handler. leave these lines alone
	mh.add(twitchlistinghandlers(client))
	mh.add(twitchitemhandlers(client))
	return func(m msg) {
		if mh.handle(m) {
			log.Println("handled")
		}
	}
}

type itemhandlers struct {
	*twitch.Client
	handlers
}

func (ih *itemhandlers) add(prefix string, fmtr func(MLBCard) string) {
	msgtype := strings.ToUpper(strings.TrimPrefix(prefix, "!"))
	ih.handlers.add(command(prefix, handlerf(func(m msg) bool {
		args := strings.TrimSpace(strings.TrimPrefix(m.Message, prefix))
		cards, _ := items.Search(args)
		for _, card := range cards {
			ih.Client.Say(m.Channel, fmt.Sprintf("[%s] %s | %s",
				msgtype, descr(card), fmtr(card),
			))
		}

		if len(cards) == 0 {
			ih.Client.Say(m.Channel, fmt.Sprintf("[%s] data not found", msgtype))
		}

		return true
	})))

}

// create a description for a MLB card
// this cannot be used on listings
func descr(c MLBCard) string {
	return fmt.Sprintf("%s (%v) | %s %v", c.Name, c.Ovr, c.TeamShortName, c.Series)
}

type listinghandlers struct {
	*twitch.Client
	handlers
}

func (ih *listinghandlers) add(prefix string, fmtr func(Listing) string) {
	msgtype := strings.ToUpper(strings.TrimPrefix(prefix, "!"))

	ih.handlers.add(command(prefix, handlerf(func(m msg) bool {
		args := strings.TrimSpace(strings.TrimPrefix(m.Message, prefix))
		cards, _ := listings.Search(args)
		for _, card := range cards {
			prefix := fmt.Sprintf("%s (%v) | %s %v", card.Item.Name, card.Item.Ovr, card.Item.TeamShortName, card.Item.Series)
			ih.Client.Say(m.Channel, fmt.Sprintf("[%s] %s | %s",
				msgtype, prefix, fmtr(card),
			))
		}

		if len(cards) == 0 {
			ih.Client.Say(m.Channel, fmt.Sprintf("[%s] data not found", msgtype))
		}

		return true
	})))

}
