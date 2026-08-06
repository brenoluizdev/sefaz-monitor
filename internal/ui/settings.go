package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"sefazmonitor/internal/config"
)

func (a *App) buildSettingsWindow() error {
	var win *walk.MainWindow
	var tv *walk.TableView
	var intervalEdit *walk.NumberEdit

	selected := make(map[string]bool, len(a.cfg.SelectedUFs))
	for _, c := range a.cfg.SelectedUFs {
		selected[c] = true
	}
	a.model = newUFTableModel(selected)

	err := MainWindow{
		AssignTo: &win,
		Title:    "SEFAZ Monitor — Configurações",
		Size:     Size{Width: 780, Height: 540},
		Layout:   VBox{},
		Visible:  false,
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{Text: "Verificar a cada:"},
					NumberEdit{AssignTo: &intervalEdit, MinValue: 5, MaxValue: 120, Value: float64(a.cfg.IntervalMinutes), Suffix: " min"},
					HSpacer{},
				},
			},
			Label{Text: "Marque as UFs que deseja monitorar:"},
			TableView{
				AssignTo:         &tv,
				CheckBoxes:       true,
				AlternatingRowBG: true,
				ColumnsOrderable: true,
				Columns: []TableViewColumn{
					{Title: "UF", Width: 40},
					{Title: "Estado", Width: 170},
					{Title: "Status", Width: 110},
					{Title: "Mensagem", Width: 300},
					{Title: "Última verificação", Width: 140},
				},
				StyleCell: func(style *walk.CellStyle) {
					if style.Col() != 2 {
						return
					}
					row := a.model.rows[style.Row()]
					style.TextColor = statusWalkColor(row.Status)
				},
				Model: a.model,
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "Verificar agora",
						OnClicked: func() {
							go a.mon.CheckNow()
						},
					},
					PushButton{
						Text: "Salvar",
						OnClicked: func() {
							a.saveFromSettings(intervalEdit)
						},
					},
				},
			},
		},
	}.Create()
	if err != nil {
		return err
	}

	// Fechar a janela ("X") apenas a esconde: a bandeja continua ativa e o
	// monitoramento em segundo plano não é interrompido.
	win.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		*canceled = true
		win.Hide()
	})

	a.settingsWin = win
	return nil
}

func (a *App) openSettings() {
	if a.settingsWin == nil {
		if err := a.buildSettingsWindow(); err != nil {
			walk.MsgBox(nil, "SEFAZ Monitor", "Não foi possível abrir a janela de configurações: "+err.Error(), walk.MsgBoxIconError)
			return
		}
	}
	a.settingsWin.Show()
	a.settingsWin.Activate()
}

func (a *App) saveFromSettings(intervalEdit *walk.NumberEdit) {
	a.cfg.SelectedUFs = a.model.selectedCodes()
	a.cfg.IntervalMinutes = int(intervalEdit.Value())

	if err := config.Save(a.cfg); err != nil {
		walk.MsgBox(a.settingsWin, "SEFAZ Monitor", "Falha ao salvar configuração: "+err.Error(), walk.MsgBoxIconError)
		return
	}

	a.mon.SetConfig(a.cfg)
	go a.mon.CheckNow()

	a.ni.SetToolTip(a.trayTooltip())
	walk.MsgBox(a.settingsWin, "SEFAZ Monitor", "Configuração salva.", walk.MsgBoxIconInformation)
}
