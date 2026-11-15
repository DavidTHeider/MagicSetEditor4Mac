// Die Blätter, die das Netz erreichen — abgetrennt von `regeln_test.go`, weil
// es damit über die Zeilengrenze aus §2 ging (A8, #1447).
//
// **Der Schnitt folgt der Frage.** Dort steht, wie die Schichten zueinander
// liegen; hier, welche Dienstklienten es gibt und warum jeder genau `domain`
// und `netz` kennt. Es sind acht Zeilen mit derselben Begründung, und sie
// wachsen mit jedem Anbieter — A9 und A10 stehen im Plan.
package schichten

// netzblaetter sind die Pakete, die einen fremden Dienst über `netz` lesen.
//
// **Sie kennen einander nicht.** Zwei Anbieter derselben Fähigkeit sind zwei
// Blätter; was sie gemeinsam haben, steht in `domain`.
var netzblaetter = map[string][]string{
	// **`mtgapi` ist ein Blatt wie `scryfall`** — Phase 4.8
	// (#1321).
	//
	// Es liest Karten von magicthegathering.io und kennt dafür `domain` und
	// `netz` — **denselben** Transport wie Scryfall. Das ist der Grund, warum
	// die Netzgrenze bei ihren fünf Einträgen bleibt: „jeder weitere fremde
	// Dienst hängt an ihm und kommt **nicht** in diese Aufzählung"
	// ([0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)).
	"mtgapi": {"domain", "netz"},

	// **`mtggraphql` ist ein Blatt wie `mtgapi`** — Phase 4.8, derselbe
	// Transport und dieselbe Begründung.
	"mtggraphql": {"domain", "netz"},

	// **`mtgjsonpreise` ist ein Blatt wie `mtgapi`** — Werkstück A3 von Phase
	// 5.3 (#1442).
	//
	// **Nicht dasselbe wie `mtgjson`**: dort ein **Dateiformat**, das jemand
	// mitbringt, hier ein **Dienst** mit eigenem Vertrag, eigenen Grenzen und
	// eigenem Transport. Das Issue verlangt die Trennung ausdrücklich — „statt
	// eine vorhandene Produktionsanbindung zu behaupten".
	//
	// **Und es kennt `mtgjson` nicht**: die Preisdatei hat mit dem
	// Kartenformat nichts gemein außer dem Namen des Anbieters.
	"mtgjsonpreise": {"domain", "netz"},

	// **`cardtrader` ist ein Blatt wie `mtgjsonpreise`** — Werkstück A4 von
	// Phase 5.3 (#1443): Transport über `netz`, Zuordnung auf `domain`. Es
	// kennt keinen Bestand; welcher Druck zu welchem Blueprint gehört, weiß
	// `service`.
	"cardtrader": {"domain", "netz"},

	// **`ollama` ist ein Blatt wie `cardtrader`** — A7 (#1446). Dass sein
	// Dienst auf der Schleife liegt, ändert daran nichts: „Loopback allein
	// beweist keine lokale Inferenz", und wer ihn deshalb nicht führte,
	// hätte einen Netzverbraucher ohne Eintrag.
	"ollama": {"domain", "netz"},

	// **`anthropic` ist ein Blatt wie `cardtrader`** — A8 (#1447): Transport
	// über `netz`, Zuordnung auf `domain`.
	//
	// **Und es kennt `internal/ollama` nicht.** Zwei Anbieter derselben
	// Fähigkeit sind zwei Blätter; was sie gemeinsam haben, steht in
	// `domain` (die sieben Fähigkeiten aus A6) und nicht in einem von beiden.
	"anthropic": {"domain", "netz"},

	// **`openai` ist ein Blatt wie `anthropic`** — A9 (#1448), und es kennt
	// `internal/anthropic` **nicht**: zwei Anbieter derselben Fähigkeit sind
	// zwei Blätter, und was sie gemeinsam haben, steht in `domain`.
	"openai": {"domain", "netz"},

	// **`xai` ist ein Blatt wie die beiden darüber** — A10 (#1449). Es
	// kennt weder `openai` noch `anthropic`, und das ist hier mehr als
	// Ordnung: xAI hat an **zwei** eigenen Endpunkten **zwei verschiedene**
	// Bildformen, und die Modellliste hat eine Gestalt, die es bei keinem
	// anderen gibt. Ein gemeinsames Paket hätte genau das verwischt.
	"xai": {"domain", "netz"},

	// **`applefm` ist ein Blatt wie die drei darüber** — AI-05 (#2042). Es
	// kennt keinen anderen Anbieter, und sein Transport ist ein Unix-Socket
	// aus `netz`: kein Port, kein Proxy.
	"applefm": {"domain", "netz"},

	// **`scrydex` ist ein Blatt wie `mtgapi`** — Phase 4.8, derselbe Transport
	// und dieselbe Begründung. **Nur Magic**: Scrydex führt sechs Spiele, hier
	// zählt eines.
	"scrydex": {"domain", "netz"},
}

// mitNetzblaettern legt sie zur Tabelle der Schichten — samt Wurzelknoten.
//
// **Eine Tabelle und nicht zwei.** Die Prüfungen schlagen in `erlaubt` nach;
// zwei Karten nebeneinander wären die zweite Aufzählung, vor der §2 warnt.
func mitNetzblaettern(kern map[string][]string) map[string][]string {
	for paket, ziele := range netzblaetter {
		kern[paket] = ziele
	}

	for paket, ziele := range wurzelknoten {
		kern[paket] = ziele
	}

	for paket, ziele := range naehte {
		kern[paket] = ziele
	}

	return kern
}
