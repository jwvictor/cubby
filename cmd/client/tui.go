package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"log"
	"os"
)

type Page int

const (
	PageMainMenu = iota
	PageSearch
	PageNew
	PageExplore
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	currentPage Page
	mainMenu    tea.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, nil
		}
	}

	if m.currentPage == PageMainMenu {
		var cmd tea.Cmd
		m.mainMenu, cmd = m.mainMenu.Update(msg)
		return m, cmd
	}

	return m, nil
}
func (m model) View() string {
	if m.currentPage == PageMainMenu {
		return m.mainMenu.View()
	}
	return "Invalid view.\n\n"
}

type mainMenuModel struct {
	list list.Model
}

func (m mainMenuModel) Init() tea.Cmd {
	return nil
}

func (m mainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, nil
		}
	case tea.WindowSizeMsg:
		top, right, bottom, left := docStyle.GetMargin()
		m.list.SetSize(msg.Width-left-right, msg.Height-top-bottom)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m mainMenuModel) View() string {
	return docStyle.Render(m.list.View())
}

var mainMenuItems []list.Item = []list.Item{
	item{title: "Explore", desc: "List and view"},
	item{title: "Search", desc: "Basically sticky fabric"},
	item{title: "Terrycloth", desc: "In other words, towel fabric"},
}

func initialMainMenuModel() mainMenuModel {
	m := mainMenuModel{list: list.NewModel(mainMenuItems, list.NewDefaultDelegate(), 0, 0)}
	m.list.Title = "My Fave Things"
	return m
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Run Cubby TUI",
	Long:  `Runs Cubby in TUI mode, i.e. in a terminal but using text-based UI elements.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getClient()
		err := client.Authenticate()
		if err != nil {
			log.Printf("Error: Authentication - %s\n", err.Error())
		}

		p := tea.NewProgram(model{
			currentPage: PageMainMenu,
			mainMenu:    initialMainMenuModel(),
		}, tea.WithAltScreen())

		if err := p.Start(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}
