#define MyAppName "NeuralPath Tactical Guard"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "NeuralPath"
#define MyAppExeName "NeuralPathTacticalGuard.exe"
#define MyAppId "{{D4F65E58-8D8B-4B9E-91C3-2C0E6E4B8F41}}"

[Setup]
AppId={#MyAppId}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={autopf}\NeuralPath\Tactical Guard
DefaultGroupName={#MyAppName}
OutputDir=..\artifacts\installer
OutputBaseFilename=NeuralPath-Tactical-Guard-{#MyAppVersion}-Setup-x64
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
Compression=lzma2
SolidCompression=yes
WizardStyle=modern
PrivilegesRequired=admin
DisableProgramGroupPage=yes
UninstallDisplayIcon={app}\{#MyAppExeName}
LicenseFile=..\relase\EULA.txt

[Languages]
Name: "italian"; MessagesFile: "compiler:Languages\Italian.isl"
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "Crea collegamento sul desktop"; Flags: unchecked

[Files]
Source: "..\artifacts\stage\NeuralPath Tactical Guard\*"; DestDir: "{app}"; Flags: ignoreversion recursesubdirs createallsubdirs

[Icons]
Name: "{autoprograms}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; IconFilename: "{app}\Satellite.ico"
Name: "{autodesktop}\{#MyAppName}"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon; IconFilename: "{app}\Satellite.ico"

[Run]
Filename: "{app}\{#MyAppExeName}"; Description: "Avvia {#MyAppName}"; Flags: nowait postinstall skipifsilent