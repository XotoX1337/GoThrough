package overlay

import "github.com/XotoX1337/GoThrough/engine"

// StepInfo is the data shape sent to the Wails frontend.
type StepInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Current     int    `json:"current"`
	Total       int    `json:"total"`
	IsFirst     bool   `json:"isFirst"`
	IsLast      bool   `json:"isLast"`
}

// App is the Go backend exposed to the frontend via Wails bindings.
type App struct {
	eng *engine.Engine
}

func (a *App) CurrentStep() StepInfo { return a.stepInfo() }

func (a *App) Next() StepInfo {
	_ = a.eng.Next()
	return a.stepInfo()
}

func (a *App) Prev() StepInfo {
	_ = a.eng.Prev()
	return a.stepInfo()
}

func (a *App) stepInfo() StepInfo {
	current, total := a.eng.Progress()
	step := a.eng.Current()
	return StepInfo{
		Title:       step.Title,
		Description: step.Description,
		Current:     current,
		Total:       total,
		IsFirst:     current == 1,
		IsLast:      a.eng.Done(),
	}
}
