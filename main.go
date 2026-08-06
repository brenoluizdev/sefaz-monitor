// SEFAZ Monitor: aplicativo de bandeja que consulta periodicamente a página
// pública de disponibilidade da NFe e alerta o usuário quando o status de
// uma UF selecionada muda (operacional/instável/indisponível).
package main

import (
	"log"

	"sefazmonitor/internal/ui"
)

func main() {
	if err := ui.Run(); err != nil {
		log.Fatalf("erro fatal: %v", err)
	}
}
