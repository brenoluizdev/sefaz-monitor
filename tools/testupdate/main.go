// Programa auxiliar de teste manual (não faz parte do app): simula uma
// instalação antiga (versão 1.0.0) e roda o fluxo real de auto-update
// contra o GitHub Releases público do projeto, pra validar ponta a ponta:
// detecção de versão nova, download, verificação de checksum e disparo do
// instalador silencioso.
package main

import (
	"fmt"

	"sefazmonitor/internal/updater"
	"sefazmonitor/internal/version"
)

func main() {
	version.Current = "1.0.0"
	fmt.Println("Simulando versão instalada:", version.Current)

	found, err := updater.CheckAndInstall(func(newVersion string) {
		fmt.Println("Checksum OK. Disparando instalador silencioso para a versão", newVersion, "...")
	})

	if err != nil {
		fmt.Println("ERRO:", err)
		return
	}
	if !found {
		fmt.Println("Nenhuma atualização encontrada (inesperado neste teste).")
		return
	}
	fmt.Println("Instalador disparado com sucesso. Verifique o Gerenciador de Tarefas / a versão instalada em seguida.")
}
