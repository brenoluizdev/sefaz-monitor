package ui

import (
	"time"

	"github.com/lxn/walk"

	"sefazmonitor/internal/updater"
	"sefazmonitor/internal/version"
)

// updateCheckInterval é o intervalo entre checagens automáticas em segundo
// plano, enquanto o app fica aberto por vários dias sem reiniciar.
const updateCheckInterval = 12 * time.Hour

// startUpdateChecker dispara uma checagem pouco depois do início (sem
// atrasar a abertura do app) e depois periodicamente.
func (a *App) startUpdateChecker() {
	go func() {
		time.Sleep(10 * time.Second)
		a.checkForUpdate(false)

		ticker := time.NewTicker(updateCheckInterval)
		defer ticker.Stop()
		for range ticker.C {
			a.checkForUpdate(false)
		}
	}()
}

// checkForUpdate verifica, baixa e instala uma atualização se houver.
// manual=true dá feedback mesmo quando não há nada novo ou a checagem
// falha (clique explícito no menu); checagens automáticas ficam quietas
// nesses dois casos, e só se manifestam quando de fato há algo para
// instalar.
func (a *App) checkForUpdate(manual bool) {
	found, err := updater.CheckAndInstall(func(newVersion string) {
		notify("SEFAZ Monitor: atualizando", "Instalando a versão "+newVersion+". O app vai reiniciar sozinho em alguns segundos.")
		time.Sleep(1500 * time.Millisecond)
	})

	if err != nil {
		if manual {
			notify("SEFAZ Monitor", "Não foi possível verificar atualizações: "+err.Error())
		}
		return
	}

	if !found {
		if manual {
			notify("SEFAZ Monitor", "Você já está na versão mais recente ("+version.Current+").")
		}
		return
	}

	// A instalação silenciosa já foi disparada como um processo separado.
	// Este processo precisa sair agora para liberar o próprio executável,
	// que o instalador vai sobrescrever.
	a.mon.Stop()
	a.mw.Synchronize(func() {
		walk.App().Exit(0)
	})
}
