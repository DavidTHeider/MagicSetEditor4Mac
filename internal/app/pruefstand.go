package app

// **Was geprüft ist und was nur da** — OPS-07.AK-02
// (#1724).
//
// # Vier Stände, und jeder hat einen Beleg
//
// **Geprüft** heißt: ein Lauf oder eine Probe hat es belegt, und der Beleg
// liegt im Repositorium. **Vorhanden** heißt: gebaut, aber nie gegen das
// Original oder die Plattform gelaufen. **Beschrieben** heißt: nicht gebaut.
// **Blockiert** heißt: es fehlt, was nur ein Mensch oder ein Gerät belegen
// kann.
//
// # Und ein Wächter hält die Liste gegen ihre Quellen
//
// `internal/werkzeuge/pruefstand_test.go`: jeder Beleg liegt da, und die
// fremden Dienste stehen so, wie `docs/HAERTUNG.md` sie führt. **Was er nicht
// hält**, steht in seinem Kopf: ob ein Beleg trägt.

// Die vier Stände.
const (
	StandGeprueft    = "geprueft"
	StandVorhanden   = "vorhanden"
	StandBeschrieben = "beschrieben"
	StandBlockiert   = "blockiert"
)

// Die Bereiche.
const (
	BereichPlattform        = "plattform"
	BereichSprache          = "sprache"
	BereichBarrierefreiheit = "barrierefreiheit"
	BereichRegeln           = "regeln"
	BereichIntegration      = "integration"
)

// PruefpunktDTO ist ein Gegenstand mit seinem Stand und seinem Beleg.
type PruefpunktDTO struct {
	Bereich    string `json:"bereich"`
	Gegenstand string `json:"gegenstand"`
	Stand      string `json:"stand"`

	// Beleg ist eine Datei im Repositorium oder die Adresse eines Issues.
	Beleg string `json:"beleg"`
}

const haertung = "docs/HAERTUNG.md"

// Pruefstand gibt die Liste.
//
// **Der Gegenstand ist zugleich der Schlüssel des Katalogs**: das Fenster
// übersetzt, was eine Beschreibung ist, und lässt stehen, was ein Eigenname
// ist — ein Dienst heißt in jeder Sprache so, wie er heißt (§9).
func Pruefstand() []PruefpunktDTO {
	return []PruefpunktDTO{
		{BereichPlattform, "macOS 27 auf Apple Silicon", StandGeprueft, "docs/belege/huellenprobe/ops08.md"},
		// **Belegt ist 15.2, erlaubt ist ab 15.2** (Vermerk 0184): was dazwischen
		// liegt, lässt das Bündel zu, und kein Lauf hat es gesehen.
		{BereichPlattform, "macOS 15.2 auf Apple Silicon", StandGeprueft, "docs/belege/huellenprobe/untergrenze-15.2.md"},
		{BereichPlattform, "macOS 15.3 bis 26", StandVorhanden, "build/darwin/Info.plist"},
		{BereichPlattform, "Windows, Linux", StandBeschrieben, "docs/PORTIERUNG.md"},
		{BereichSprache, "Deutsch, English", StandGeprueft, "frontend/src/sprache/sprache.test.tsx"},
		// **Sechs Kataloge, maschinell übersetzt und ohne Redaktion** (H04,
		// #2258) — vorhanden, nicht geprüft. Hier stand bis zum 2026-10-11
		// „Weitere Sprachen: beschrieben" mit #1687, das seit LOC-01 zu ist (J01).
		{BereichSprache, "Español, Français, Italiano, 日本語, Русский, 中文", StandVorhanden, "docs/LOKALISIERUNG.md"},
		{BereichBarrierefreiheit, "WCAG 2.2 AA (axe-core)", StandGeprueft, "frontend/src/barrierefreiheit.test.tsx"},
		{BereichBarrierefreiheit, "Tastatur und Vorleser im WebKit", StandGeprueft, "docs/belege/webkitprobe/ux10-vorleser.md"},
		// **Die Handprüfung und nicht ihr Issue** (Vermerk 0205): das Issue liegt
		// im privaten Entwicklungsarchiv, die Tafel der offenen Handprüfungen im
		// Repositorium — dort steht, was ein Mensch mit VoiceOver prüft.
		{BereichBarrierefreiheit, "VoiceOver", StandBlockiert, "docs/HANDPRUEFUNGEN.md"},
		{BereichRegeln, "Comprehensive Rules", StandVorhanden, "docs/REGELABDECKUNG.md"},
		{BereichIntegration, "Scryfall", StandGeprueft, haertung},
		{BereichIntegration, "magicthegathering.io", StandGeprueft, haertung},
		{BereichIntegration, "MTGJSON (Preise)", StandGeprueft, haertung},
		{BereichIntegration, "MTGJSON (Karten)", StandGeprueft, haertung},
		{BereichIntegration, "MTGJSON (GraphQL)", StandVorhanden, haertung},
		{BereichIntegration, "CardTrader", StandVorhanden, haertung},
		{BereichIntegration, "Scrydex", StandVorhanden, haertung},
		{BereichIntegration, "Ollama", StandGeprueft, haertung},
		{BereichIntegration, "Anthropic", StandVorhanden, haertung},
		{BereichIntegration, "OpenAI", StandVorhanden, haertung},
		{BereichIntegration, "xAI", StandVorhanden, haertung},
		{BereichIntegration, "Anmeldeanbieter (OAuth)", StandVorhanden, haertung},
	}
}
