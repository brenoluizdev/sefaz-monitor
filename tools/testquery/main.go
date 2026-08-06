// Programa auxiliar de teste manual (não faz parte do app): busca a página
// pública de disponibilidade da NFe e imprime o status de algumas UFs, para
// validar o scraper.
package main

import (
	"fmt"
	"os"
	"time"

	"sefazmonitor/internal/scraper"
	"sefazmonitor/internal/ufs"
)

func main() {
	snap, err := scraper.Fetch(20 * time.Second)
	if err != nil {
		fmt.Println("erro:", err)
		os.Exit(1)
	}
	fmt.Println("Última verificação (SEFAZ):", snap.ServerCheckedAt)
	fmt.Println("Linhas encontradas:", len(snap.Rows))
	for env := range snap.Rows {
		fmt.Println(" -", env)
	}

	fmt.Println()
	for _, code := range []string{"SP", "MG", "AC", "RJ", "MA", "AM"} {
		u, _ := ufs.ByCode(code)
		row, ok := snap.Rows[u.EnvKey()]
		if !ok {
			fmt.Printf("%s (%s): linha não encontrada\n", code, u.EnvKey())
			continue
		}
		fmt.Printf("%s (linha %s): pior=%s offenders=%v\n", code, u.EnvKey(), row.Worst(), row.Offenders())
	}
}
