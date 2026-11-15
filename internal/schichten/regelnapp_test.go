// **Was `internal/app` kennen darf** — herausgelöst am 2026-09-07, als
// `regeln_test.go` mit der Dateisperre über die Zeilengrenze aus §2 kam.
//
// # Warum der Schnitt hier verläuft
//
// **Weil die Frage eine andere ist.** Nebenan steht, **welche** Regel für jede
// Schicht gilt — meist eine Zeile je Paket. Hier steht, **warum** die Fassade
// ausgerechnet diese Nachbarn kennen darf, und das ist bei fast jedem eine
// eigene Geschichte: drei Kartendienste, ein Schlüsselbund, ein Sensor, eine
// Sperre, ein Anmeldefluss.
//
// **Und sie ist die einzige Liste, die so wächst.** `internal/app` ist der
// Composition Root: jede neue Plattformnaht kommt hier an, und jede bringt
// ihren Grund mit (§2).
package schichten

// erlaubtInApp führt die Pakete, die `internal/app` kennen darf.
var erlaubtInApp = []string{
	"aktualisierung", "bezug", "domain", "msepaket", "msescript",
	"mtgjson", "netz", "pdf", "render", "schriften", "scryfall",
	"service", "storage", "svg", "vorschauweg",

	// **Die drei neuen Kartendienste** — Phase 4.8
	// (#1321).
	//
	// Sie stehen hier aus demselben Grund wie `scryfall`: **hier** werden
	// die Klienten gebaut, und Adresse, Frist, Drossel und Zugang sind
	// Fragen der Verdrahtung, nicht der Fachlichkeit (§2, §16).
	// `internal/service` kennt keinen von ihnen — es kennt eine
	// **Fähigkeit** (`service.Suchdienst`).
	"mtgapi", "mtggraphql", "mtgjsonpreise", "scrydex", "cardtrader",
	"ollama", "anthropic", "openai", "xai",

	// **Und Apples Gerätemodell** (AI-05, #2042) — aus demselben Grund:
	// Socket, Prozess und Frist sind Verdrahtung.
	"applefm",

	// **Und der Schlüsselbund**, aus demselben Grund: wo ein Zugang
	// liegt, ist eine Frage der Plattform.
	"schluesselbund",

	// **Und die Leitung** — Werkstück N2 von Phase 5.4
	// (#1455).
	//
	// **Weil hier der Transport ausgesucht wird.** `service` stellt
	// [service.Waehler] und [service.Verbindung] auf und kennt keinen
	// Transport; dass ein `*leitung.Leitung` beides erfüllt, weiß nur der
	// Composition Root. Dieser Eintrag ist am 2026-09-14 aus `nurInTests`
	// **hierher gewandert**, und der Kommentar dort hatte den Umzug
	// vorgesehen.
	"leitung",

	// **Und der zweite Transport** — Werkstück N3 von Phase 5.4
	// (#1456).
	//
	// Aus demselben Grund wie `leitung` darüber, und er ist die Probe auf
	// die Zeile: **der Wähler entscheidet hier**, welchen von beiden er
	// nimmt — an einer Einladung im Adressfeld erkannt. Wüsste `service`
	// davon, hätte die Fachschicht sich einen Transport ausgesucht.
	"fernleitung",

	// **Und der Umschlag** — Werkstück N4 von Phase 5.4
	// (#1457).
	//
	// **Weil hier zwei vorhandene Wege verschnürt werden.** `austausch`
	// schreibt und liest ein Paket und weiß nichts von Sets; dass ein Set
	// **entfaltet** hineingeht und beim Herauskommen wieder ein `.mse-set`
	// ist, entscheidet der Composition Root. Ein `service`, das den
	// Umschlag kennte, hätte ein Dateiformat in der Fachschicht.
	"austausch",

	// **Und die Entdeckung** — Werkstück N2 von Phase 5.4
	// (#1455).
	//
	// Aus demselben Grund wie `leitung` darüber: `service` stellt
	// [service.Nachbarschaft] auf und kennt keinen Transport; dass mDNS sie
	// erfüllt, weiß nur der Composition Root.
	"entdeckung",

	// **Und die Anmeldung** — Werkstück 4 von Phase 4.12
	// (#1351).
	//
	// Aus demselben Grund wie die Kartendienste: **hier** wird der Fluss
	// verdrahtet — welcher Anbieter, welche Kennung, welches Geheimnis,
	// wer den Browser aufmacht. `internal/service` kennt keinen Anbieter;
	// es kennt eine [domain.Fremdkennung] und fragt, wem sie gehört.
	"anmeldung",

	// **Und die Biometrie** — Werkstück 7 desselben Issues, aus
	// demselben Grund: **hier** wird verdrahtet, wo ein Geheimnis liegt
	// und wer den Sensor fragt. `internal/service` kennt zwei Nähte
	// ([service.Geheimnisse], [service.Sensor]) und keinen Rechner.
	"biometrie",

	// **Und die Peeridentität** — Werkstück N1 von Phase 5.4
	// (#1454).
	//
	// Ein Blatt wie die darüber, und aus demselben Grund hier: eine
	// [peer.Kennung] ist ein **Typ** und keine Zeichenkette, und die Naht
	// nimmt sie von der Fläche entgegen. Wer sie hier als `string`
	// weiterreichte, hätte beim ersten vertauschten Argument eine
	// Profilkennung an ihrer Stelle.
	//
	// **Fachliches entscheidet sie nicht**: ob ein Kontakt etwas darf,
	// steht in [service.Kontaktwerk] und in [domain.Peerkontakt].
	"peer",

	// **Und die Dateisperre** — Befund 5 des Zweitblicks über Phase 4.13
	// (#1352).
	//
	// `app.Protokoll` dreht ein volles `mse.log` nur, wenn kein anderer
	// Prozess hineinschreibt — und diese Frage beantwortet eine Sperre auf
	// einer Nachbardatei, so wie bei `bestand.db.sperre`. Sie ist eine
	// **Frage an das Betriebssystem**, nicht an die Fachlichkeit, und
	// gehört damit zu `schluesselbund` und `biometrie`.
	//
	// **Warum keine Naht**: das Protokoll entsteht in `main.go`, **bevor**
	// der Bestand offen ist und bevor irgendetwas verdrahtet werden
	// könnte. Eine Naht, die niemand setzen kann, wäre ein Feld, das im
	// Betrieb `nil` bleibt — genau das, wovor `naehte_intern_test.go`
	// warnt.
	"dateisperre",

	// **Und der Prozessspeicher** — OPS02
	// (#2348),
	// aus demselben Grund wie die Dateisperre: eine **Frage an das
	// Betriebssystem** — wie viel braucht dieser Prozess, und wie viel seine
	// WebKit-Prozesse —, die die Metriken des Fensters stellen. Fachliches
	// entscheidet sie nicht.
	"prozessspeicher",
}
