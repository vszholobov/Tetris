package main

import (
	"fmt"
	"strconv"
	"tetrisClient/keyboard"
)

type Menu struct {
	currentSessionIndex int
	sessionsList        []SessionDto
	isEnded             bool
	isCreateSession     bool
	isExit              bool
	keyboard            *keyboard.Keyboard
}

func MakeMenu(keyboard *keyboard.Keyboard) Menu {
	sessionsList := getSessionsList()
	return Menu{
		currentSessionIndex: 0,
		sessionsList:        sessionsList,
		isEnded:             false,
		isCreateSession:     false,
		isExit:              false,
		keyboard:            keyboard,
	}
}

func (menu *Menu) getSelectedSessionId() string {
	return strconv.FormatInt(menu.sessionsList[menu.currentSessionIndex].SessionId, 10)
}

func (menu *Menu) showMenu() {
	menu.keyboard.Clear()
	fmt.Println(" Tetris🕹️ " + Version)
	fmt.Println("----------")
	for index, session := range menu.sessionsList {
		currentItem := ""
		if index == menu.currentSessionIndex {
			currentItem += "\033[30;5;107m"
		}
		currentItem += strconv.FormatInt(session.SessionId, 10)
		currentItem += " "
		currentItem += strconv.FormatBool(session.Started)
		if index == menu.currentSessionIndex {
			currentItem += "\033[0m"
		}
		fmt.Println(currentItem)
	}
}

func (menu *Menu) handleMenu() {
	for !menu.isEnded {
		menu.showMenu()
		input := menu.keyboard.Read()
		switch input {
		case 115:
			// s
			if len(menu.sessionsList) == 0 {
				continue
			}
			menu.currentSessionIndex++
			menu.currentSessionIndex = menu.currentSessionIndex % len(menu.sessionsList)
		case 119:
			// w
			if len(menu.sessionsList) == 0 {
				continue
			}
			menu.currentSessionIndex--
			if menu.currentSessionIndex < 0 {
				menu.currentSessionIndex = len(menu.sessionsList) - 1
			}
		case 114:
			// r
			menu.sessionsList = getSessionsList()
		case 99:
			// c
			menu.isEnded = true
			menu.isCreateSession = true
			continue
		case 13:
			// enter
			if len(menu.sessionsList) == 0 {
				continue
			}
			menu.isEnded = true
			continue
		case 101:
			// e
			menu.isEnded = true
			menu.isExit = true
		default:
			// skip unknown input
			continue
		}
	}
}
