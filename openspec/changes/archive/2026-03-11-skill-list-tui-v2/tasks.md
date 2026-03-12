## 1. Model and State Machine

- [x] 1.1 Define `viewState` type with `stateRegPicker`, `stateLoading`, `stateSkillList` constants
- [x] 1.2 Define `skillListModel` struct with `state`, `regList`, `skillList`, `cfg`, `installed`, `installing`, `spinnerFrame`, `status`, `statusErr`, and `multiRegistry` fields
- [x] 1.3 Define `skillInstalledMsg` struct with `name string`, `registry string`, `err error`, and `backends int` fields
- [x] 1.4 Define `skillsLoadedMsg` struct with `skills []skill.Info` and `err error` fields
- [x] 1.5 Define `tickMsg` struct for spinner tick events

## 2. Custom Item Delegate

- [x] 2.1 Define `skillDelegate` struct with `installed`, `installing map[string]bool` and `spinnerFrame *int` fields
- [x] 2.2 Implement `skillDelegate.Height()` returning 2 (title + description line)
- [x] 2.3 Implement `skillDelegate.Spacing()` returning 1
- [x] 2.4 Implement `skillDelegate.Update()` returning the item and nil cmd (no delegate-level updates needed)
- [x] 2.5 Implement `skillDelegate.Render()`: show green ✓ for installed, spinner glyph for installing, blank for neither; render title and description lines with appropriate lipgloss styles for selected/normal states

## 3. tea.Cmd Functions

- [x] 3.1 Implement `loadSkillsCmd(cfg *config.Config, registryName string) tea.Cmd` that calls `buildCatalog` in a goroutine and returns `skillsLoadedMsg`
- [x] 3.2 Implement `installSkillCmd(cfg *config.Config, info skill.Info) tea.Cmd` that performs symlink creation, config save, and `syncer.Run` in a goroutine and returns `skillInstalledMsg`
- [x] 3.3 Implement `tickCmd() tea.Cmd` that returns `tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg{} })`

## 4. Model Init

- [x] 4.1 Implement `newSkillListModel(cfg *config.Config, fromRegistry string) skillListModel` that initializes `installed` map from `cfg.InstalledSkills`, sets `multiRegistry` based on registry count and `fromRegistry`, and sets initial state
- [x] 4.2 When `multiRegistry` is true: build `regList` from `cfg.Registries` names using `list.New`, set `state = stateRegPicker`
- [x] 4.3 When `multiRegistry` is false (single registry or `--from`): set `state = stateLoading` and return `loadSkillsCmd` from `Init()`
- [x] 4.4 Implement `Init() tea.Cmd`: return `loadSkillsCmd` if `state == stateLoading`, else nil

## 5. Update Handler

- [x] 5.1 Handle `tea.WindowSizeMsg`: resize whichever list is active
- [x] 5.2 Handle `tea.KeyMsg` in `stateRegPicker`: `q`/`ctrl+c` → quit; `enter` → set `state = stateLoading`, return `loadSkillsCmd` for selected registry
- [x] 5.3 Handle `tea.KeyMsg` in `stateLoading`: ignore all input (prevent accidental navigation during load)
- [x] 5.4 Handle `tea.KeyMsg` in `stateSkillList`: `q`/`ctrl+c` → quit; `esc` → back to picker if `multiRegistry`, else quit; `enter` → invoke install or show "already installed" status
- [x] 5.5 Handle `skillsLoadedMsg`: build `skillList` list.Model with custom delegate, set `state = stateSkillList`; on error set `statusErr = true`, `status = err.Error()`
- [x] 5.6 Handle `skillInstalledMsg`: on success update `installed[name] = true`, delete from `installing`, set green status; on error delete from `installing`, set red status; if `len(installing) == 0` stop ticking
- [x] 5.7 Handle `tickMsg`: increment `spinnerFrame`; if `len(installing) > 0` return another `tickCmd()`
- [x] 5.8 Delegate unhandled key and other messages to the active `list.Model` via its `Update` method

## 6. View

- [x] 6.1 Implement `View()`: dispatch to `viewRegPicker()`, `viewLoading()`, or `viewSkillList()` based on `state`
- [x] 6.2 Implement `viewRegPicker()`: render `regList.View()`
- [x] 6.3 Implement `viewLoading()`: render a simple "Loading…" message centered in the terminal
- [x] 6.4 Implement `viewSkillList()`: render `skillList.View()` plus a status line at the bottom styled green (success) or red (error) based on `statusErr`

## 7. Wire into skillListCmd

- [x] 7.1 Replace the existing `skillListModel` instantiation in `skillListCmd.RunE` with `newSkillListModel(cfg, skillListFrom)`
- [x] 7.2 Remove the old flat `buildCatalog` call and item-slice construction that previously happened before the TUI started
- [x] 7.3 Ensure `tea.WithAltScreen()` is still passed to `tea.NewProgram`

## 8. Styles and Polish

- [x] 8.1 Define `spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}` at package level
- [x] 8.2 Define lipgloss styles for ✓ glyph (green), spinner glyph (yellow/dim), selected item title (bold), normal item title, and status line (green / red)
- [x] 8.3 Set `skillList.Title` to `"<registry> · N skills"` when transitioning to `stateSkillList`
- [x] 8.4 Set `regList.Title` to `"Select a registry"` and apply the existing `titleStyle`
- [x] 8.5 Add help text to the skill list: `"↵ install  esc back  q quit"` (or `"↵ install  q quit"` for single-registry mode) — use `list.SetShowHelp(false)` and render manually in `viewSkillList` to keep it simple, or configure `list.AdditionalShortHelpKeys`

## 9. Verification

- [x] 9.1 Build the binary (`go build ./...`) and confirm no compilation errors
- [x] 9.2 Run with zero registries configured and verify the existing "no registries" message appears unchanged
- [ ] 9.3 Run with one registry and verify the picker is skipped and skill list opens directly with "registry · N skills" title
- [ ] 9.4 Run with two or more registries and verify the registry picker appears, Enter transitions to skill list, Esc returns to picker
- [ ] 9.5 Press Enter on an uninstalled skill and verify spinner appears, then ✓ on success with status message
- [ ] 9.6 Press Enter on an already-installed skill and verify "already installed" status message with no change to ✓ state
- [ ] 9.7 Press Enter rapidly on two skills and verify both show spinners and both complete independently
- [ ] 9.8 Run `sk skill list --from <registry>` and verify picker is skipped, Esc quits rather than going back
