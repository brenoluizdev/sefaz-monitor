// Programa auxiliar de teste manual (não faz parte do app): confirma que
// trocar o intervalo via SetConfig com o Monitor já rodando passa a valer
// imediatamente, em vez de só depois que a espera antiga (mais longa)
// terminar sozinha.
package main

import (
	"fmt"
	"time"

	"sefazmonitor/internal/config"
	"sefazmonitor/internal/monitor"
)

func main() {
	var checks []time.Time

	m := monitor.New(nil, func() {
		checks = append(checks, time.Now())
		fmt.Printf("[%s] onUpdate #%d\n", time.Now().Format("15:04:05.000"), len(checks))
	})

	m.SetConfig(config.Config{SelectedUFs: []string{"SP"}, IntervalSeconds: 600})
	m.Start()
	defer m.Stop()

	fmt.Println("Monitor iniciado com intervalo de 600s (10 min). Esperando 2s...")
	time.Sleep(2 * time.Second)

	fmt.Println("Trocando para intervalo de 3s. Se o fix funcionar, a próxima verificação deve vir em ~3s, não em ~600s.")
	before := time.Now()
	m.SetConfig(config.Config{SelectedUFs: []string{"SP"}, IntervalSeconds: 3})

	time.Sleep(8 * time.Second)

	var afterChange int
	for _, c := range checks {
		if c.After(before) {
			afterChange++
		}
	}
	fmt.Printf("Verificações após a troca de intervalo (em 8s): %d\n", afterChange)
	if afterChange >= 2 {
		fmt.Println("OK: o novo intervalo passou a valer imediatamente.")
	} else {
		fmt.Println("FALHOU: o novo intervalo não entrou em vigor a tempo.")
	}
}
