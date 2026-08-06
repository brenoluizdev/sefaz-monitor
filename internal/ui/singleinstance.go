package ui

import (
	"syscall"
	"unsafe"
)

// appMutexName precisa ser IDÊNTICO ao AppMutex configurado em
// installer/sefaz-monitor.iss. Serve a dois propósitos: impede duas
// instâncias do app rodando ao mesmo tempo, e permite que o instalador
// (rodado com /CLOSEAPPLICATIONS durante uma auto-atualização) detecte e
// encerre o app antes de sobrescrever o executável.
const appMutexName = "SefazMonitorAppMutex"

const errorAlreadyExists = 183

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutexW = kernel32.NewProc("CreateMutexW")
)

// acquireSingleInstanceLock devolve true se este processo é a única
// instância (o mutex acabou de ser criado por ele). Devolve false se outra
// instância já está rodando. O handle do mutex nunca é liberado
// explicitamente — fica associado ao processo até ele terminar, que é
// exatamente o comportamento desejado.
func acquireSingleInstanceLock() bool {
	namePtr, err := syscall.UTF16PtrFromString(appMutexName)
	if err != nil {
		return true
	}

	ret, _, lastErr := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(namePtr)))
	if ret == 0 {
		return true
	}
	return lastErr != syscall.Errno(errorAlreadyExists)
}
