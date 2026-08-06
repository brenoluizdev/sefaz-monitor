// Package ui monta a bandeja do sistema (NotifyIcon), a janela de
// configurações e liga tudo ao monitor.Monitor: intervalo de verificação,
// seleção de UFs e disparo de notificações quando o status de uma UF muda.
package ui

import (
	"fmt"

	"github.com/lxn/walk"

	"sefazmonitor/internal/config"
	"sefazmonitor/internal/monitor"
)

// App amarra a janela principal (oculta), o NotifyIcon da bandeja, a janela
// de configurações (criada sob demanda) e o Monitor de status.
type App struct {
	mw  *walk.MainWindow
	ni  *walk.NotifyIcon
	mon *monitor.Monitor
	cfg config.Config

	settingsWin *walk.MainWindow
	model       *ufTableModel
}

// Run inicializa a aplicação e bloqueia executando o loop de mensagens do
// Windows até que o usuário escolha "Sair" no menu da bandeja.
func Run() error {
	if !acquireSingleInstanceLock() {
		walk.MsgBox(nil, "SEFAZ Monitor", "O SEFAZ Monitor já está em execução — veja o ícone na bandeja do sistema.", walk.MsgBoxIconInformation)
		return nil
	}

	registerAppUserModelID()

	a := &App{cfg: config.Load()}

	mw, err := walk.NewMainWindow()
	if err != nil {
		return fmt.Errorf("criar janela principal: %w", err)
	}
	a.mw = mw

	a.mon = monitor.New(a.onTransition, a.onUpdate)
	a.mon.SetConfig(a.cfg)

	if err := a.setupTray(); err != nil {
		return fmt.Errorf("configurar bandeja: %w", err)
	}

	a.mon.Start()
	defer a.mon.Stop()

	a.startUpdateChecker()

	if len(a.cfg.SelectedUFs) == 0 {
		// Primeira execução: já abre a janela de configuração para o
		// usuário escolher quais UFs monitorar.
		a.openSettings()
	}

	mw.Run()
	return nil
}

func (a *App) setupTray() error {
	ni, err := walk.NewNotifyIcon(a.mw)
	if err != nil {
		return err
	}
	a.ni = ni

	icon, err := statusIcon(monitor.Unknown)
	if err != nil {
		return err
	}
	if err := ni.SetIcon(icon); err != nil {
		return err
	}
	if err := ni.SetToolTip(a.trayTooltip()); err != nil {
		return err
	}

	openAction := walk.NewAction()
	openAction.SetText("Configurações...")
	openAction.Triggered().Attach(func() { a.openSettings() })
	if err := ni.ContextMenu().Actions().Add(openAction); err != nil {
		return err
	}

	checkAction := walk.NewAction()
	checkAction.SetText("Verificar agora")
	checkAction.Triggered().Attach(func() { go a.mon.CheckNow() })
	if err := ni.ContextMenu().Actions().Add(checkAction); err != nil {
		return err
	}

	updateAction := walk.NewAction()
	updateAction.SetText("Verificar atualizações")
	updateAction.Triggered().Attach(func() { go a.checkForUpdate(true) })
	if err := ni.ContextMenu().Actions().Add(updateAction); err != nil {
		return err
	}

	if err := ni.ContextMenu().Actions().Add(walk.NewSeparatorAction()); err != nil {
		return err
	}

	exitAction := walk.NewAction()
	exitAction.SetText("Sair")
	exitAction.Triggered().Attach(func() {
		a.mon.Stop()
		walk.App().Exit(0)
	})
	if err := ni.ContextMenu().Actions().Add(exitAction); err != nil {
		return err
	}

	ni.MouseDown().Attach(func(x, y int, button walk.MouseButton) {
		if button == walk.LeftButton {
			a.openSettings()
		}
	})

	return ni.SetVisible(true)
}

func (a *App) trayTooltip() string {
	n := len(a.cfg.SelectedUFs)
	if n == 0 {
		return "SEFAZ Monitor — nenhuma UF configurada"
	}
	return fmt.Sprintf("SEFAZ Monitor — %d UF(s) — pior status: %s", n, a.mon.Worst())
}

// onTransition é chamado pelo Monitor (em goroutine de background) sempre
// que o status classificado de uma UF muda. Atualiza o ícone/tooltip da
// bandeja e dispara uma notificação nativa do Windows (toast).
func (a *App) onTransition(old, new monitor.UFState) {
	a.mw.Synchronize(func() {
		icon, err := statusIcon(a.mon.Worst())
		if err == nil {
			a.ni.SetIcon(icon)
		}
		a.ni.SetToolTip(a.trayTooltip())
	})

	if old.Status == monitor.Unknown {
		// Primeira classificação desde que o app iniciou: não é uma
		// "queda" nem uma "recuperação", só o estado inicial.
		return
	}

	msg := new.Message
	if msg == "" {
		msg = fmt.Sprintf("Mudou de %s para %s", old.Status, new.Status)
	}

	// notify() sobe um processo powershell.exe e pode levar 1-2s; roda fora
	// da goroutine de UI para não travar a janela/bandeja enquanto isso.
	go notify(severityTitle(new.UF, new.Status), msg)
}

// onUpdate é chamado ao final de cada ciclo completo de verificação, para
// atualizar a tabela da janela de configurações (se estiver aberta).
func (a *App) onUpdate() {
	a.mw.Synchronize(func() {
		if a.model != nil {
			a.model.applyStates(a.mon.States())
		}
	})
}
