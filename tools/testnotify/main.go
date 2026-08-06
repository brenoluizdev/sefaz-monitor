// Programa auxiliar de demonstração (não faz parte do app): dispara os três
// estilos de notificação (verde/amarelo/vermelho) em sequência, usando o
// mesmo mecanismo (go-toast + AUMID registrado no atalho) que o app real
// usa, para visualizar como os alertas aparecem no Windows.
package main

import (
	"fmt"
	"time"

	"github.com/go-toast/toast"
)

const aumid = "SefazMonitor.App"

func send(title, msg string) {
	n := toast.Notification{AppID: aumid, Title: title, Message: msg, Audio: toast.Default}
	if err := n.Push(); err != nil {
		fmt.Println("erro:", err)
	}
}

func main() {
	send("SEFAZ SP: Operacional", "Todos os serviços operacionais (SEFAZ verificou às 10:32:00)")
	time.Sleep(5 * time.Second)
	send("SEFAZ SP: Instável", "Status Serviço: Amarelo (SEFAZ verificou às 10:33:00)")
	time.Sleep(5 * time.Second)
	send("SEFAZ SP: Indisponível", "Autorização: Vermelho; Status Serviço: Vermelho (SEFAZ verificou às 10:34:00)")
}
