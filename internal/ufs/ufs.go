// Package ufs contém os metadados das 27 unidades federativas usados para
// resolver qual linha da tabela pública de disponibilidade da NFe
// corresponde a cada UF: seu próprio ambiente autorizador ("own") ou um dos
// ambientes compartilhados ("SVRS" ou "SVAN").
package ufs

// UF representa uma unidade federativa e seu ambiente autorizador.
type UF struct {
	Code string // sigla, ex: "SP"
	Name string // nome completo
	Env  string // ambiente autorizador: "own" (usa o próprio Code como linha), "SVRS" ou "SVAN"
}

// All lista as 27 UFs (26 estados + Distrito Federal) em ordem alfabética.
var All = []UF{
	{"AC", "Acre", "SVRS"},
	{"AL", "Alagoas", "SVRS"},
	{"AM", "Amazonas", "own"},
	{"AP", "Amapá", "SVRS"},
	{"BA", "Bahia", "own"},
	{"CE", "Ceará", "SVRS"},
	{"DF", "Distrito Federal", "SVRS"},
	{"ES", "Espírito Santo", "SVRS"},
	{"GO", "Goiás", "own"},
	{"MA", "Maranhão", "SVAN"},
	{"MG", "Minas Gerais", "own"},
	{"MS", "Mato Grosso do Sul", "own"},
	{"MT", "Mato Grosso", "own"},
	{"PA", "Pará", "SVRS"},
	{"PB", "Paraíba", "SVRS"},
	{"PE", "Pernambuco", "own"},
	{"PI", "Piauí", "SVRS"},
	{"PR", "Paraná", "own"},
	{"RJ", "Rio de Janeiro", "SVRS"},
	{"RN", "Rio Grande do Norte", "SVRS"},
	{"RO", "Rondônia", "SVRS"},
	{"RR", "Roraima", "SVRS"},
	{"RS", "Rio Grande do Sul", "own"},
	{"SC", "Santa Catarina", "SVRS"},
	{"SE", "Sergipe", "SVRS"},
	{"SP", "São Paulo", "own"},
	{"TO", "Tocantins", "SVRS"},
}

// EnvKey devolve a chave da linha na tabela de disponibilidade que
// corresponde a esta UF: seu próprio código, ou "SVRS"/"SVAN".
func (u UF) EnvKey() string {
	if u.Env == "own" {
		return u.Code
	}
	return u.Env
}

// ByCode indexa All pela sigla da UF.
func ByCode(code string) (UF, bool) {
	for _, u := range All {
		if u.Code == code {
			return u, true
		}
	}
	return UF{}, false
}
