package sndotfiles

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/jonhadfield/gosn"
	"github.com/ryanuber/columnize"
)

func Sync(session gosn.Session, home string, quiet bool) (noPushed, noPulled int, err error) {
	var remote tagsWithNotes
	remote, err = get(session)
	if err != nil {
		return
	}
	err = preflight(remote)
	if err != nil {
		return
	}
	return sync(session, remote, home, quiet)
}

func sync(session gosn.Session, twn tagsWithNotes, home string, quiet bool) (noPushed, noPulled int, err error) {
	var itemDiffs []ItemDiff
	itemDiffs, err = diff(twn, home, nil)
	if err != nil {
		if strings.Contains(err.Error(), "tags with notes not supplied") {
			err = errors.New("no remote dotfiles found")
		}
		return
	}

	var itemsToPush, itemsToPull []ItemDiff
	var itemsToSync bool
	for _, itemDiff := range itemDiffs {
		switch itemDiff.diff {
		case localNewer:
			//push
			itemDiff.remote.Content.SetText(itemDiff.local)
			itemsToPush = append(itemsToPush, itemDiff)
			itemsToSync = true
		case localMissing:
			// pull
			itemsToPull = append(itemsToPull, itemDiff)
			itemsToSync = true
		case remoteNewer:
			// pull
			itemsToPull = append(itemsToPull, itemDiff)
			itemsToSync = true
		}
	}
	bold := color.New(color.Bold).SprintFunc()

	// check items to sync
	if !itemsToSync {
		if !quiet {
			fmt.Println(bold("nothing to do"))
		}
		return
	}

	// push
	if len(itemsToPush) > 0 {
		_, err = push(session, itemsToPush)
		noPushed = len(itemsToPush)
		if err != nil {
			return
		}
	}

	res := make([]string, len(itemsToPush))
	green := color.New(color.FgGreen).SprintFunc()
	strPushed := green("pushed")
	strPulled := green("pulled")

	for _, pushItem := range itemsToPush {
		line := fmt.Sprintf("%s | %s", bold(addDot(pushItem.homeRelPath)), strPushed)
		res = append(res, line)
	}

	// pull
	if err = pull(itemsToPull); err != nil {
		return
	}
	noPulled = len(itemsToPull)

	for _, pullItem := range itemsToPull {
		line := fmt.Sprintf("%s | %s\n", bold(addDot(pullItem.homeRelPath)), strPulled)
		res = append(res, line)
	}
	if !quiet {
		fmt.Println(columnize.SimpleFormat(res))
	}
	return noPushed, noPulled, err
}
