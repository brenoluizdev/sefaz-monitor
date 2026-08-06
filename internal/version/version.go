// Package version expõe a versão do app embutida no binário no momento da
// build.
package version

// Current é sobrescrita no build de release via:
//
//	go build -ldflags "-X sefazmonitor/internal/version.Current=1.2.3"
//
// (o workflow .github/workflows/release.yml faz isso automaticamente a
// partir da tag git). Builds locais/de desenvolvimento ficam com "dev", o
// que desativa a verificação de atualização — não faz sentido comparar
// versão contra uma build que não corresponde a nenhum release.
var Current = "dev"

// IsDev indica uma build local, sem versão de release conhecida.
func IsDev() bool {
	return Current == "dev"
}
