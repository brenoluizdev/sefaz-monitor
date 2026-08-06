; Script do Inno Setup para o instalador do SEFAZ Monitor.
; Compilar com: ISCC.exe installer\sefaz-monitor.iss
; (gera installer\output\SefazMonitorSetup.exe)

#define MyAppName "SEFAZ Monitor"
; Sobrescrito no CI via "ISCC /DMyAppVersion=1.2.3 ..." a partir da tag git,
; para acompanhar exatamente a versão embutida no binário (ver
; internal/version). O valor abaixo só vale para builds locais manuais.
#ifndef MyAppVersion
  #define MyAppVersion "0.0.0-dev"
#endif
#define MyAppPublisher "SEFAZ Monitor"
#define MyAppExeName "SefazMonitor.exe"
; Identidade (AUMID) usada para registrar o atalho e para o app disparar
; notificações do Windows corretamente. Deve ser idêntica à constante usada
; em internal/ui/notify.go.
#define MyAppAUMID "SefazMonitor.App"

[Setup]
AppId={{697ABFD0-2175-4583-A5E7-2FB673661522}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
; Instalação por usuário (sem exigir privilégios de administrador): o
; público-alvo típico (departamentos fiscais/contábeis) muitas vezes usa
; máquinas corporativas sem direitos de admin local.
PrivilegesRequired=lowest
; Precisa ser IDÊNTICO ao nome do mutex criado pelo app (ver
; internal/ui/singleinstance.go). Permite ao instalador detectar e fechar o
; app em execução durante uma auto-atualização (/CLOSEAPPLICATIONS) e
; reabri-lo depois (/RESTARTAPPLICATIONS).
AppMutex=SefazMonitorAppMutex
DefaultDirName={localappdata}\Programs\SefazMonitor
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
UninstallDisplayIcon={app}\{#MyAppExeName}
OutputDir=output
OutputBaseFilename=SefazMonitorSetup
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
SetupIconFile=..\resources\app.ico

[Languages]
Name: "brazilianportuguese"; MessagesFile: "compiler:Languages\BrazilianPortuguese.isl"

[Tasks]
Name: "desktopicon"; Description: "Criar ícone na área de trabalho"; GroupDescription: "Ícones adicionais:"
Name: "startupicon"; Description: "Iniciar automaticamente com o Windows"; GroupDescription: "Inicialização:"

[Files]
Source: "..\SefazMonitor.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "set-aumid.ps1"; DestDir: "{tmp}"; Flags: nocompression

[Icons]
Name: "{group}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\Desinstalar {#MyAppName}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon
Name: "{userstartup}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: startupicon

[Run]
; Registra a identidade (AUMID) no atalho do Menu Iniciar — necessário para
; que o Windows exiba as notificações do app (ver set-aumid.ps1).
Filename: "powershell.exe"; Parameters: "-NoProfile -ExecutionPolicy Bypass -File ""{tmp}\set-aumid.ps1"" ""{group}\{#MyAppName}.lnk"" ""{app}\{#MyAppExeName}"" ""{#MyAppAUMID}"""; Flags: runhidden; StatusMsg: "Registrando notificações do Windows..."
Filename: "{app}\{#MyAppExeName}"; Description: "Executar {#MyAppName} agora"; Flags: nowait postinstall skipifsilent

[UninstallDelete]
; Remove o binário; a configuração do usuário em %APPDATA%\SefazMonitor é
; mantida propositalmente (permite reinstalar sem perder as UFs escolhidas).
Type: files; Name: "{app}\{#MyAppExeName}"
