package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	twitch "github.com/gempir/go-twitch-irc/v4"
)

type msg = twitch.PrivateMessage
type msghandler interface {
	handle(msg) bool
}

type handlerf func(msg) (handled bool)

func (f handlerf) handle(m msg) bool {
	return f(m)
}

type handlers struct {
	hf []msghandler
}

func (h *handlers) add(f msghandler) {
	h.hf = append(h.hf, f)
}
func (h *handlers) handle(m msg) bool {
	for _, f := range h.hf {
		if f.handle(m) {
			return true
		}
	}
	return false
}

func conds(h msghandler, conds ...handlerf) msghandler {
	return handlerf(func(m msg) bool {
		for _, f := range conds {
			if !f(m) {
				return false
			}
		}
		return h.handle(m)
	})
}

var gusm twitch.GlobalUserStateMessage

func match(exp string, h msghandler) msghandler {
	return conds(h,
		func(m msg) bool {
			mch, err := regexp.MatchString(exp, m.Message)
			if err != nil {
				log.Println(err)
				return false
			}

			return mch
		},
	)
}

func command(cmd string, h msghandler) msghandler {
	return match(fmt.Sprintf("(?i)^%s\\b", cmd), h)
}

func itemcommand(cmd string, h func(msg, MLBCard) bool) msghandler {
	return command(cmd, handlerf(func(m msg) bool {
		args := strings.TrimSpace(strings.TrimPrefix(m.Message, cmd))
		cards, _ := items.Search(args)
		t := false
		for _, card := range cards {
			res := h(m, card)
			if res == true {
				t = true
			}
		}

		return t
	}))
}

func listingcommand(cmd string, h func(msg, Listing) bool) msghandler {
	return command(cmd, handlerf(func(m msg) bool {
		args := strings.TrimSpace(strings.TrimPrefix(m.Message, cmd))
		cards, _ := listings.Search(args)
		t := false
		for _, card := range cards {
			res := h(m, card)
			if res == true {
				t = true
			}
		}

		return t
	}))
}

// func itemcommand(cmd string, h func(msg, []Item) bool) msghandler {
// 	return command(fmt.Sprintf("(?i)^!%s\\b", handlerf(func(m msg) bool {
// 		cards := items.Search(
// 		return h(m, item)
// 	}))
// }
