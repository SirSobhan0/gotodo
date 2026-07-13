package todo

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/jalaali/go-jalaali"

	pkgtodo "github.com/SirSobhan0/gotodo/pkg/todo"
)

// English UI Strings
const (
	title                 = "Go Todo TUI - Time Tracker"
	newTaskPrompt         = "New Task:"
	newProjectPrompt      = "New Project Name:"
	inputPlaceholder      = "Type here..."
	noTasks               = "No tasks yet. Press 'a' to add one!"
	noProjects            = "No projects found. Press 'a' to create one!"
	helpAdd               = "add"
	helpDelete            = "delete"
	helpToggle            = "start/pause/resume"
	helpComplete          = "complete task"
	helpNav               = "nav"
	helpQuit              = "quit"
	helpConfirm           = "confirm"
	helpCancelBack        = "back"
	helpScrollUp          = "scroll up"
	helpScrollDown        = "scroll down"
	helpConfirmStay       = "confirm (stay)"
	helpToggleLineNumbers = "toggle line #s"
	helpToggleCalendar    = "toggle calendar (G/J)"
	savingTasks           = "Saving tasks..."
	bye                   = "Bye!"
	errorOnExit           = "Error on exit: %v\n"
	errorPrefix           = "Error: %v"
	errorSave             = "save error: %w"
	errorLoadingTasksLog  = "Error loading tasks: %v\n"
	inputAreaTaskTitle    = "📝 Add New Task"
	inputAreaProjTitle    = "📁 Create New Project"
	statsPending          = "Pending"
	statsInProgress       = "In Progress"
	statsCompleted        = "Completed"
	calendarGregorian     = "Gregorian (MM/DD)"
	calendarJalali        = "Jalali (MM/DD)"
)

type Model struct {
	projects          []string
	activeProject     string
	projectCursor     int
	tasks             []pkgtodo.Task
	cursor            int
	input             textinput.Model
	viewport          viewport.Model
	width, height     int
	mode              appMode
	helpMsg           string
	quitting          bool
	err               error
	keyMap            pkgtodo.KeyMap
	showLineNumbers   bool
	ready             bool
	useJalaliCalendar bool
}

type appMode int

const (
	modeSelectProject appMode = iota
	modeAddProject
	modeViewTasks
	modeAddTask
)

type TickMsg time.Time

const appHorizontalPadding = 2
const appVerticalPadding = 2

func (m *Model) initializeStyles() {
	pkgtodo.AppStyle = lipgloss.NewStyle().Padding(1)
	pkgtodo.TitleStyle = lipgloss.NewStyle().Bold(true).MarginBottom(1).Align(lipgloss.Center)
	pkgtodo.StatsStyle = lipgloss.NewStyle().Padding(0, 1).MarginBottom(1).Bold(true)
	pkgtodo.CalendarIndicatorStyle = lipgloss.NewStyle().Padding(0, 1).MarginBottom(1).Italic(true)

	pkgtodo.TaskViewportStyle = lipgloss.NewStyle().Border(lipgloss.DoubleBorder(), true)
	pkgtodo.ListItemStyle = lipgloss.NewStyle().Padding(0, 1)
	pkgtodo.SelectedListItemStyle = lipgloss.NewStyle().Reverse(true).Padding(0, 1)

	pkgtodo.StatusRenderWidth = lipgloss.Width(pkgtodo.StatusInProgress) + 1
	pkgtodo.TimeRenderWidth = lipgloss.Width("[00:00:00]") + 1
	pkgtodo.DateRenderWidth = lipgloss.Width("(00/00)") + 1
	pkgtodo.LineNumberWidth = lipgloss.Width("999. ")

	pkgtodo.StatusPendingStyle = lipgloss.NewStyle()
	pkgtodo.StatusInProgressStyle = lipgloss.NewStyle()
	pkgtodo.StatusPausedStyle = lipgloss.NewStyle()
	pkgtodo.StatusCompletedStyle = lipgloss.NewStyle()

	pkgtodo.DescriptionStyle = lipgloss.NewStyle().Align(lipgloss.Left)
	pkgtodo.TimeTextSyle = lipgloss.NewStyle()
	pkgtodo.DateTextSyle = lipgloss.NewStyle()
	pkgtodo.LineNumberStyle = lipgloss.NewStyle()

	pkgtodo.InputAreaStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).MarginBottom(1)
	pkgtodo.InputPromptStyle = lipgloss.NewStyle().PaddingRight(1)
	pkgtodo.FocusedInputStyle = lipgloss.NewStyle().Border(lipgloss.ThickBorder(), true).Padding(0, 1)
	pkgtodo.BlurredInputStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).Padding(0, 1)

	m.input.PromptStyle = pkgtodo.InputPromptStyle
	m.input.TextStyle = lipgloss.NewStyle()
	m.input.PlaceholderStyle = lipgloss.NewStyle()
	m.input.CursorStyle = lipgloss.NewStyle()

	pkgtodo.HelpStyle = lipgloss.NewStyle().Padding(0, 1).Bold(true)
	pkgtodo.ErrorStyle = lipgloss.NewStyle().Bold(true).Padding(0, 1).MarginBottom(1).Border(lipgloss.RoundedBorder()).Align(lipgloss.Center)

	m.keyMap = pkgtodo.KeyMap{
		Add:               key.NewBinding(key.WithKeys("a"), key.WithHelp("a", helpAdd)),
		Delete:            key.NewBinding(key.WithKeys("d"), key.WithHelp("d", helpDelete)),
		Toggle:            key.NewBinding(key.WithKeys("s"), key.WithHelp("s", helpToggle)),
		Complete:          key.NewBinding(key.WithKeys("c"), key.WithHelp("c", helpComplete)),
		Up:                key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", helpNav)),
		Down:              key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", helpNav)),
		Quit:              key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", helpQuit)),
		Enter:             key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", helpConfirm)),
		Esc:               key.NewBinding(key.WithKeys("esc", "b"), key.WithHelp("esc/b", helpCancelBack)),
		ScrollUp:          key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", helpScrollUp)),
		ScrollDown:        key.NewBinding(key.WithKeys("pgdown"), key.WithHelp("pgdown", helpScrollDown)),
		ToggleLineNumbers: key.NewBinding(key.WithKeys("n"), key.WithHelp("n", helpToggleLineNumbers)),
		ToggleCalendar:    key.NewBinding(key.WithKeys("t"), key.WithHelp("t", helpToggleCalendar)),
	}
	m.input.Placeholder = inputPlaceholder
}

func InitialModel() Model {
	m := Model{
		showLineNumbers:   false,
		useJalaliCalendar: false,
		mode:              modeSelectProject,
	}

	ti := textinput.New()
	ti.CharLimit = 156
	m.input = ti

	m.initializeStyles()

	vp := viewport.New(80, 20)
	m.viewport = vp
	m.viewport.Style = pkgtodo.TaskViewportStyle

	var err error
	m.projects, err = ListProjects()
	if err != nil {
		m.err = err
	}

	if len(m.projects) == 0 {
		m.mode = modeAddProject
		m.input.Prompt = newProjectPrompt + " "
		m.input.Focus()
	}

	m.helpMsg = generateHelp(m.keyMap, m.mode)
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, doTick())
}

func doTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return TickMsg(t) })
}

func (m *Model) ensureCursorVisible() {
	currentCursor := m.cursor
	if m.mode == modeSelectProject {
		currentCursor = m.projectCursor
	}

	if currentCursor < m.viewport.YOffset {
		m.viewport.SetYOffset(currentCursor)
	} else if currentCursor >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.SetYOffset(currentCursor - m.viewport.Height + 1)
	}
}

func (m *Model) saveCurrentProject() {
	if m.activeProject != "" {
		if err := SaveTasksToFile(m.activeProject, m.tasks); err != nil {
			m.err = fmt.Errorf(errorSave, err)
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.width = msg.Width
			m.height = msg.Height
			m.ready = true
		} else {
			m.width = msg.Width
			m.height = msg.Height
		}

		availableWidth := m.width - appHorizontalPadding
		currentAvailableHeight := m.height - appVerticalPadding

		titleViewHeight := lipgloss.Height(pkgtodo.TitleStyle.Render(title))
		currentAvailableHeight -= titleViewHeight

		statsBarContent := m.renderStatsBar()
		statsBarHeight := lipgloss.Height(pkgtodo.StatsStyle.Render(statsBarContent))
		currentAvailableHeight -= statsBarHeight

		calendarIndicatorText := calendarGregorian
		if m.useJalaliCalendar {
			calendarIndicatorText = calendarJalali
		}
		calendarIndicatorHeight := lipgloss.Height(pkgtodo.CalendarIndicatorStyle.Render(calendarIndicatorText))
		currentAvailableHeight -= calendarIndicatorHeight

		helpViewHeight := lipgloss.Height(pkgtodo.HelpStyle.Render(m.helpMsg))
		currentAvailableHeight -= helpViewHeight

		if m.err != nil {
			errorViewHeight := lipgloss.Height(pkgtodo.ErrorStyle.Render(fmt.Sprintf(errorPrefix, m.err)))
			currentAvailableHeight -= errorViewHeight
		}

		m.viewport.Width = max(1, availableWidth-pkgtodo.TaskViewportStyle.GetHorizontalFrameSize())

		if m.mode == modeAddTask || m.mode == modeAddProject {
			promptStr := newTaskPrompt
			titleStr := inputAreaTaskTitle
			if m.mode == modeAddProject {
				promptStr = newProjectPrompt
				titleStr = inputAreaProjTitle
			}
			m.input.Prompt = promptStr + " "

			inputCurrentStyle := pkgtodo.BlurredInputStyle
			if m.input.Focused() {
				inputCurrentStyle = pkgtodo.FocusedInputStyle
			}

			// Iteratively subtract frame sizes to prevent outer box expansion
			outerContentWidth := availableWidth - pkgtodo.InputAreaStyle.GetHorizontalFrameSize()
			innerContentWidth := outerContentWidth - inputCurrentStyle.GetHorizontalFrameSize()

			// Text input strictly bounded to the remaining inner area
			m.input.Width = max(10, innerContentWidth-lipgloss.Width(m.input.Prompt))

			inputFieldRender := inputCurrentStyle.Width(innerContentWidth).Render(m.input.View())
			inputBoxTitle := lipgloss.NewStyle().Bold(true).Width(outerContentWidth).Align(lipgloss.Center).Render(titleStr)
			inputContentForHeight := lipgloss.JoinVertical(lipgloss.Left, inputBoxTitle, inputFieldRender)

			inputAreaRenderedHeight := lipgloss.Height(pkgtodo.InputAreaStyle.Render(inputContentForHeight))
			currentAvailableHeight -= inputAreaRenderedHeight
		} else {
			m.viewport.Height = max(1, currentAvailableHeight-pkgtodo.TaskViewportStyle.GetVerticalFrameSize())
		}

		if m.mode == modeSelectProject {
			m.viewport.SetContent(m.renderProjectView())
		} else {
			m.viewport.SetContent(m.renderTasksView())
		}

	case TickMsg:
		return m, doTick()

	case tea.KeyMsg:
		if m.err != nil && msg.Type != tea.KeyCtrlC && msg.String() != "q" {
			m.err = nil
		}

		switch m.mode {
		case modeSelectProject:
			switch {
			case key.Matches(msg, m.keyMap.Quit):
				m.quitting = true
				return m, tea.Quit
			case key.Matches(msg, m.keyMap.Add):
				m.mode = modeAddProject
				m.input.Prompt = newProjectPrompt + " "
				m.input.SetValue("")
				m.input.Focus()
				m.helpMsg = generateHelp(m.keyMap, modeAddProject)
				return m, textinput.Blink
			case key.Matches(msg, m.keyMap.Up):
				if len(m.projects) > 0 && m.projectCursor > 0 {
					m.projectCursor--
					m.ensureCursorVisible()
				}
			case key.Matches(msg, m.keyMap.Down):
				if len(m.projects) > 0 && m.projectCursor < len(m.projects)-1 {
					m.projectCursor++
					m.ensureCursorVisible()
				}
			case key.Matches(msg, m.keyMap.Enter):
				if len(m.projects) > 0 {
					m.activeProject = m.projects[m.projectCursor]
					loadedTasks, _ := LoadTasksFromFile(m.activeProject)
					m.tasks = loadedTasks
					m.cursor = 0
					m.mode = modeViewTasks
					m.helpMsg = generateHelp(m.keyMap, modeViewTasks)
				}
			case key.Matches(msg, m.keyMap.Delete):
				if len(m.projects) > 0 {
					projToDelete := m.projects[m.projectCursor]
					os.Remove(GetProjectFile(projToDelete))
					m.projects, _ = ListProjects()
					if m.projectCursor >= len(m.projects) && len(m.projects) > 0 {
						m.projectCursor = len(m.projects) - 1
					} else if len(m.projects) == 0 {
						m.projectCursor = 0
					}
				}
			}

		case modeAddProject:
			switch {
			case key.Matches(msg, m.keyMap.Enter):
				val := strings.TrimSpace(m.input.Value())
				if val != "" {
					SaveTasksToFile(val, []pkgtodo.Task{})
					m.projects, _ = ListProjects()
					m.activeProject = val
					m.tasks = []pkgtodo.Task{}
					m.cursor = 0
					m.mode = modeViewTasks
					m.input.SetValue("")
					m.helpMsg = generateHelp(m.keyMap, modeViewTasks)
				}
			case key.Matches(msg, m.keyMap.Esc):
				m.mode = modeSelectProject
				m.input.Blur()
				m.input.SetValue("")
				m.helpMsg = generateHelp(m.keyMap, modeSelectProject)
			default:
				m.input, cmd = m.input.Update(msg)
				cmds = append(cmds, cmd)
			}

		case modeViewTasks:
			switch {
			case key.Matches(msg, m.keyMap.Esc):
				for i := range m.tasks {
					if m.tasks[i].Status == pkgtodo.InProgress {
						if !m.tasks[i].LastStartedAt.IsZero() {
							m.tasks[i].TimeSpent += time.Since(m.tasks[i].LastStartedAt)
							m.tasks[i].LastStartedAt = time.Now()
						}
					}
				}
				m.saveCurrentProject()
				m.mode = modeSelectProject
				m.activeProject = ""
				m.tasks = nil
				m.projects, _ = ListProjects()
				m.helpMsg = generateHelp(m.keyMap, modeSelectProject)

			case key.Matches(msg, m.keyMap.ToggleLineNumbers):
				m.showLineNumbers = !m.showLineNumbers
			case key.Matches(msg, m.keyMap.ToggleCalendar):
				m.useJalaliCalendar = !m.useJalaliCalendar
			case key.Matches(msg, m.keyMap.Quit):
				m.quitting = true
				for i := range m.tasks {
					if m.tasks[i].Status == pkgtodo.InProgress {
						if !m.tasks[i].LastStartedAt.IsZero() {
							m.tasks[i].TimeSpent += time.Since(m.tasks[i].LastStartedAt)
						}
						m.tasks[i].Status = pkgtodo.Paused
					}
				}
				m.saveCurrentProject()
				return m, tea.Quit
			case key.Matches(msg, m.keyMap.Add):
				m.mode = modeAddTask
				m.input.Prompt = newTaskPrompt + " "
				m.input.SetValue("")
				m.input.Focus()
				m.helpMsg = generateHelp(m.keyMap, modeAddTask)
				return m, textinput.Blink
			case key.Matches(msg, m.keyMap.Delete):
				if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
					m.tasks = append(m.tasks[:m.cursor], m.tasks[m.cursor+1:]...)
					if m.cursor >= len(m.tasks) && len(m.tasks) > 0 {
						m.cursor = len(m.tasks) - 1
					} else if len(m.tasks) == 0 {
						m.cursor = 0
					}
					m.saveCurrentProject()
				}
			case key.Matches(msg, m.keyMap.Up):
				if len(m.tasks) > 0 && m.cursor > 0 {
					m.cursor--
					m.ensureCursorVisible()
				}
			case key.Matches(msg, m.keyMap.Down):
				if len(m.tasks) > 0 && m.cursor < len(m.tasks)-1 {
					m.cursor++
					m.ensureCursorVisible()
				}
			case key.Matches(msg, m.keyMap.Toggle):
				if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
					task := &m.tasks[m.cursor]
					switch task.Status {
					case pkgtodo.Pending, pkgtodo.Paused:
						for i := range m.tasks {
							if m.tasks[i].Status == pkgtodo.InProgress && i != m.cursor {
								if !m.tasks[i].LastStartedAt.IsZero() {
									m.tasks[i].TimeSpent += time.Since(m.tasks[i].LastStartedAt)
								}
								m.tasks[i].Status = pkgtodo.Paused
							}
						}
						task.Status = pkgtodo.InProgress
						task.LastStartedAt = time.Now()
					case pkgtodo.InProgress:
						task.Status = pkgtodo.Paused
						if !task.LastStartedAt.IsZero() {
							task.TimeSpent += time.Since(task.LastStartedAt)
						}
					}
					m.saveCurrentProject()
				}
			case key.Matches(msg, m.keyMap.Complete):
				if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
					if m.tasks[m.cursor].Status == pkgtodo.InProgress {
						if !m.tasks[m.cursor].LastStartedAt.IsZero() {
							m.tasks[m.cursor].TimeSpent += time.Since(m.tasks[m.cursor].LastStartedAt)
						}
					}
					m.tasks[m.cursor].Status = pkgtodo.Completed
					m.saveCurrentProject()
				}
			}

		case modeAddTask:
			switch {
			case key.Matches(msg, m.keyMap.Enter):
				if strings.TrimSpace(m.input.Value()) != "" {
					newTask := pkgtodo.Task{ID: uuid.New(), Description: m.input.Value(), Status: pkgtodo.Pending, CreatedAt: time.Now()}
					m.tasks = append([]pkgtodo.Task{newTask}, m.tasks...)
					m.input.SetValue("")
					m.cursor = 0
					m.saveCurrentProject()

					// Ensure we flip back to the task list on Enter
					m.mode = modeViewTasks
					m.input.Blur()
					m.helpMsg = generateHelp(m.keyMap, modeViewTasks)
				}
			case key.Matches(msg, m.keyMap.Esc):
				m.mode = modeViewTasks
				m.input.Blur()
				m.input.SetValue("")
				m.helpMsg = generateHelp(m.keyMap, modeViewTasks)
			default:
				m.input, cmd = m.input.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	if m.mode == modeViewTasks && len(m.tasks) > 0 {
		if m.cursor >= len(m.tasks) {
			m.cursor = len(m.tasks) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
	} else if m.mode == modeSelectProject && len(m.projects) > 0 {
		if m.projectCursor >= len(m.projects) {
			m.projectCursor = len(m.projects) - 1
		}
		if m.projectCursor < 0 {
			m.projectCursor = 0
		}
	}

	if m.mode == modeSelectProject {
		m.viewport.SetContent(m.renderProjectView())
	} else {
		m.viewport.SetContent(m.renderTasksView())
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) renderStatsBar() string {
	if m.mode == modeSelectProject || m.mode == modeAddProject {
		return fmt.Sprintf("Projects: %d", len(m.projects))
	}

	pendingCount, inProgressCount, completedCount := 0, 0, 0
	for _, task := range m.tasks {
		switch task.Status {
		case pkgtodo.Pending:
			pendingCount++
		case pkgtodo.InProgress:
			inProgressCount++
		case pkgtodo.Completed:
			completedCount++
		}
	}
	return fmt.Sprintf("%s: %d | %s: %d | %s: %d",
		statsPending, pendingCount,
		statsInProgress, inProgressCount,
		statsCompleted, completedCount,
	)
}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	if m.quitting {
		finalMsg := savingTasks + "\n" + bye + "\n"
		if m.err != nil && !os.IsNotExist(m.err) {
			finalMsg = fmt.Sprintf(errorOnExit, m.err) + finalMsg
		}
		return pkgtodo.AppStyle.Render(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, finalMsg))
	}

	var viewParts []string

	headerTitle := title
	if m.activeProject != "" {
		headerTitle += fmt.Sprintf(" [%s]", m.activeProject)
	}
	viewParts = append(viewParts, pkgtodo.TitleStyle.Render(headerTitle))

	if m.err != nil && !os.IsNotExist(m.err) {
		viewParts = append(viewParts, pkgtodo.ErrorStyle.Render(fmt.Sprintf(errorPrefix, m.err)))
	}

	availableWidth := m.width - appHorizontalPadding

	statsWidth := availableWidth - pkgtodo.StatsStyle.GetHorizontalFrameSize()
	viewParts = append(viewParts, pkgtodo.StatsStyle.Width(statsWidth).Render(m.renderStatsBar()))

	if m.mode == modeViewTasks || m.mode == modeAddTask {
		calendarIndicatorText := "Calendar: " + calendarGregorian
		if m.useJalaliCalendar {
			calendarIndicatorText = "Calendar: " + calendarJalali
		}
		calWidth := availableWidth - pkgtodo.CalendarIndicatorStyle.GetHorizontalFrameSize()
		viewParts = append(viewParts, pkgtodo.CalendarIndicatorStyle.Width(calWidth).Render(calendarIndicatorText))
	}

	if m.mode == modeAddTask || m.mode == modeAddProject {
		inputCurrentStyle := pkgtodo.BlurredInputStyle
		if m.input.Focused() {
			inputCurrentStyle = pkgtodo.FocusedInputStyle
		}

		titleStr := inputAreaTaskTitle
		if m.mode == modeAddProject {
			titleStr = inputAreaProjTitle
		}

		outerContentWidth := availableWidth - pkgtodo.InputAreaStyle.GetHorizontalFrameSize()
		innerContentWidth := outerContentWidth - inputCurrentStyle.GetHorizontalFrameSize()

		// Explicitly constrain the input block rendering
		inputFieldRender := inputCurrentStyle.Width(innerContentWidth).Render(m.input.View())

		inputBoxTitle := lipgloss.NewStyle().Bold(true).Width(outerContentWidth).Align(lipgloss.Center).Render(titleStr)
		inputBoxContent := lipgloss.JoinVertical(lipgloss.Left, inputBoxTitle, inputFieldRender)

		viewParts = append(viewParts, pkgtodo.InputAreaStyle.Width(outerContentWidth).Render(inputBoxContent))

	} else {
		isEmpty := false
		emptyMsg := ""
		if m.mode == modeSelectProject && len(m.projects) == 0 {
			isEmpty = true
			emptyMsg = noProjects
		} else if m.mode == modeViewTasks && len(m.tasks) == 0 {
			isEmpty = true
			emptyMsg = noTasks
		}

		if isEmpty {
			noItemsRendered := lipgloss.Place(
				m.viewport.Width, m.viewport.Height,
				lipgloss.Center, lipgloss.Center,
				emptyMsg,
				lipgloss.WithWhitespaceChars(" "),
			)
			viewportBoxWidth := availableWidth - pkgtodo.TaskViewportStyle.GetHorizontalFrameSize()
			viewParts = append(viewParts, pkgtodo.TaskViewportStyle.Width(viewportBoxWidth).Height(m.viewport.Height+pkgtodo.TaskViewportStyle.GetVerticalFrameSize()).Render(noItemsRendered))
		} else {
			viewParts = append(viewParts, m.viewport.View())
		}
	}

	allContentAboveHelp := lipgloss.JoinVertical(lipgloss.Left, viewParts...)

	helpWidth := availableWidth - pkgtodo.HelpStyle.GetHorizontalFrameSize()
	helpBar := pkgtodo.HelpStyle.Width(helpWidth).Render(m.helpMsg)

	contentHeight := lipgloss.Height(allContentAboveHelp)
	helpHeight := lipgloss.Height(helpBar)
	totalContentHeight := contentHeight + helpHeight

	availableInnerHeight := m.height - appVerticalPadding
	var finalView string
	if totalContentHeight < availableInnerHeight {
		spacerHeight := availableInnerHeight - totalContentHeight
		spacer := lipgloss.NewStyle().Height(spacerHeight).Render("")
		finalView = lipgloss.JoinVertical(lipgloss.Left, allContentAboveHelp, spacer, helpBar)
	} else {
		finalView = lipgloss.JoinVertical(lipgloss.Left, allContentAboveHelp, helpBar)
	}

	return pkgtodo.AppStyle.Render(finalView)
}

func (m *Model) renderProjectView() string {
	var lines []string
	contentWidth := m.viewport.Width

	for i, proj := range m.projects {
		indentStr := "  "
		cursorStr := "❯ "
		if m.projectCursor == i {
			indentStr = cursorStr
		}

		projText := pkgtodo.DescriptionStyle.Render("📁 " + proj)

		itemStyleToUse := pkgtodo.ListItemStyle
		if m.cursor == i {
			itemStyleToUse = pkgtodo.SelectedListItemStyle
		}

		finalLineStyle := itemStyleToUse.Width(contentWidth - itemStyleToUse.GetHorizontalFrameSize())
		renderedLine := finalLineStyle.Render(indentStr + projText)
		lines = append(lines, renderedLine)
	}

	if len(lines) == 0 {
		return " "
	}
	return strings.Join(lines, "\n")
}

func (m *Model) renderTasksView() string {
	var taskLines []string
	contentWidth := m.viewport.Width

	for i, task := range m.tasks {
		var currentStatusStyle lipgloss.Style
		switch task.Status {
		case pkgtodo.Pending:
			currentStatusStyle = pkgtodo.StatusPendingStyle
		case pkgtodo.InProgress:
			currentStatusStyle = pkgtodo.StatusInProgressStyle
		case pkgtodo.Paused:
			currentStatusStyle = pkgtodo.StatusPausedStyle
		case pkgtodo.Completed:
			currentStatusStyle = pkgtodo.StatusCompletedStyle
		}
		statusText := currentStatusStyle.Render(task.Status.String())
		statusPart := statusText

		timeDisplay := task.TimeSpent
		if task.Status == pkgtodo.InProgress {
			if !task.LastStartedAt.IsZero() {
				timeDisplay += time.Since(task.LastStartedAt)
			}
		}
		formattedTime := pkgtodo.TimeTextSyle.Render("[" + formatDuration(timeDisplay) + "]")
		timePart := lipgloss.NewStyle().Align(lipgloss.Right).Width(pkgtodo.TimeRenderWidth).Render(formattedTime)

		var formattedDate string
		if m.useJalaliCalendar {
			utcTime := task.CreatedAt.In(time.UTC)
			gy, gm, gd := utcTime.Date()
			_, jm, jd, _ := jalaali.ToJalaali(gy, gm, gd)
			formattedDate = pkgtodo.DateTextSyle.Render(fmt.Sprintf("(%02d/%02d)", jm, jd))
		} else {
			formattedDate = pkgtodo.DateTextSyle.Render(fmt.Sprintf("(%02d/%02d)", task.CreatedAt.Month(), task.CreatedAt.Day()))
		}
		datePart := formattedDate

		lineNumStr := ""
		if m.showLineNumbers {
			lineNumStr = pkgtodo.LineNumberStyle.Render(fmt.Sprintf("%3d. ", i+1))
		}

		indentStr := "  "
		cursorStr := "❯ "
		if m.cursor == i {
			indentStr = cursorStr
		}

		currentLineNumberWidth := 0
		if m.showLineNumbers {
			currentLineNumberWidth = pkgtodo.LineNumberWidth
		}
		descAvailableWidth := contentWidth - lipgloss.Width(indentStr) - currentLineNumberWidth - pkgtodo.StatusRenderWidth - pkgtodo.DateRenderWidth - pkgtodo.TimeRenderWidth - lipgloss.Width("   ")
		if descAvailableWidth < 5 {
			descAvailableWidth = 5
		}

		descText := task.Description
		if lipgloss.Width(descText) > descAvailableWidth {
			runes := []rune(descText)
			truncatedRunes := []rune{}
			currentW := 0
			for _, r := range runes {
				runeW := lipgloss.Width(string(r))
				if currentW+runeW > descAvailableWidth-lipgloss.Width("...") {
					break
				}
				truncatedRunes = append(truncatedRunes, r)
				currentW += runeW
			}
			descText = string(truncatedRunes) + "..."
		}
		descriptionPart := pkgtodo.DescriptionStyle.Render(descText)

		statusPartRender := lipgloss.NewStyle().Width(pkgtodo.StatusRenderWidth).Render(statusPart)
		datePartRender := lipgloss.NewStyle().Width(pkgtodo.DateRenderWidth).Align(lipgloss.Left).Render(datePart)

		lineContent := lipgloss.JoinHorizontal(lipgloss.Top, lineNumStr, statusPartRender, " ", datePartRender, " ", descriptionPart, " ", timePart)

		itemStyleToUse := pkgtodo.ListItemStyle
		if m.cursor == i {
			itemStyleToUse = pkgtodo.SelectedListItemStyle
		}

		finalLineStyle := itemStyleToUse.Width(contentWidth - itemStyleToUse.GetHorizontalFrameSize())
		renderedLine := finalLineStyle.Render(indentStr + lineContent)
		taskLines = append(taskLines, renderedLine)
	}

	if len(taskLines) == 0 {
		return " "
	}
	return strings.Join(taskLines, "\n")
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func generateHelp(km pkgtodo.KeyMap, mode appMode) string {
	var parts []string
	switch mode {
	case modeSelectProject:
		parts = []string{
			km.Enter.Help().Key + " " + "select",
			km.Add.Help().Key + " " + "new project",
			km.Delete.Help().Key + " " + "delete project",
			km.Up.Help().Key + "/" + km.Down.Help().Key + " " + helpNav,
			km.Quit.Help().Key + " " + km.Quit.Help().Desc,
		}
	case modeViewTasks:
		parts = []string{
			km.Esc.Help().Key + " " + "back to projects",
			km.Add.Help().Key + " " + km.Add.Help().Desc,
			km.Delete.Help().Key + " " + km.Delete.Help().Desc,
			km.Up.Help().Key + "/" + km.Down.Help().Key + " " + helpNav,
			km.ToggleCalendar.Help().Key + " " + km.ToggleCalendar.Help().Desc,
			km.Toggle.Help().Key + " " + km.Toggle.Help().Desc,
			km.Complete.Help().Key + " " + km.Complete.Help().Desc,
			km.Quit.Help().Key + " " + km.Quit.Help().Desc,
		}
	case modeAddTask, modeAddProject:
		parts = []string{
			km.Enter.Help().Key + " " + helpConfirmStay,
			km.Esc.Help().Key + " " + km.Esc.Help().Desc,
		}
	}
	return strings.Join(parts, " │ ")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
