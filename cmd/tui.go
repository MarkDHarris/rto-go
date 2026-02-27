package cmd

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"rto/ui/app"
)

func RunBubbleteaTUI() error {
	model, err := app.New()
	if err != nil {
		return fmt.Errorf("could not create app model: %w", err)
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	// After the TUI exits, save the data
	if err := model.BadgeData().Save(); err != nil {
		return fmt.Errorf("saving badge data: %w", err)
	}
	if err := model.EventData().Save(); err != nil {
		return fmt.Errorf("saving events: %w", err)
	}
	if err := model.VacationData().Save(); err != nil {
		return fmt.Errorf("saving vacations: %w", err)
	}
	if err := model.HolidayData().Save(); err != nil {
		return fmt.Errorf("saving holidays: %w", err)
	}
	if err := model.GetSettings().SaveTo(model.DataDir()); err != nil {
		return fmt.Errorf("saving settings: %w", err)
	}

	return nil
}
