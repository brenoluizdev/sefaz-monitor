package ui

import (
	"syscall"
	"unsafe"

	"github.com/go-toast/toast"

	"sefazmonitor/internal/monitor"
)

// aumid identifica este app para o Windows. Precisa ser IDÊNTICO ao valor
// gravado no atalho do Menu Iniciar pelo instalador (ver
// installer/sefaz-monitor.iss e installer/set-aumid.ps1) — sem essa
// correspondência, o Windows aceita a notificação mas nunca a exibe. Ver
// https://learn.microsoft.com/windows/win32/shell/enable-desktop-toast-with-appusermodelid
const aumid = "SefazMonitor.App"

var setCurrentProcessExplicitAppUserModelID = syscall.NewLazyDLL("shell32.dll").NewProc("SetCurrentProcessExplicitAppUserModelID")

// registerAppUserModelID associa o processo em execução à mesma identidade
// gravada no atalho, para que os toasts disparados por este processo sejam
// atribuídos corretamente pelo Windows.
func registerAppUserModelID() {
	name, err := syscall.UTF16PtrFromString(aumid)
	if err != nil {
		return
	}
	setCurrentProcessExplicitAppUserModelID.Call(uintptr(unsafe.Pointer(name)))
}

// notify dispara uma notificação nativa do Windows (toast) para uma
// transição de status. Silenciosamente ignora erros: uma notificação que
// falha não deveria derrubar o monitoramento em segundo plano.
func notify(title, message string) {
	n := toast.Notification{
		AppID:   aumid,
		Title:   title,
		Message: message,
		Audio:   toast.Default,
	}
	_ = n.Push()
}

func severityTitle(uf string, s monitor.Status) string {
	return "SEFAZ " + uf + ": " + s.String()
}
