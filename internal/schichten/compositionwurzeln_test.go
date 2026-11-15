// **Die Composition Roots des Transports**
// (#1107).
//
// **Eine eigene Datei**, weil `composition_test.go` mit der Naht des
// Suchindex über die Zeilengrenze aus §2 wuchs. Der Schnitt folgt der Frage:
// dort, welche Datei den **Bestand** erzeugen darf, hier, welche einen
// **Klienten** bauen darf.
package schichten

// transportwurzeln verengt die Kanten auf `netz`, `scryfall` und die
// Fassungsprüfung.
//
// **Sie werden in `nurInDatei` eingesetzt**, und zwar in denselben Eintrag
// `"app"` — zwei Maps nebeneinander hießen zwei Regeln für dasselbe Paket.
func transportwurzeln() map[string][]string {
	return map[string][]string{
		// **`netz` und `scryfall` nur in zwei Dateien** (#1005, #888). Sie
		// stehen in der Liste darüber, weil `app` verdrahtet — aber nur an
		// **benannten** Stellen. Ohne diese Zeile dürfte jede Bindung einen
		// Klienten bauen, und der Transport wäre wieder über die Fassade
		// verteilt.
		//
		// **Die zweite kam mit den Sammeldateien.** `scryfall.go` stand bei
		// 425 Zeilen und ging damit über die Grenze aus §2; der Klient für
		// die Sammeldateien ist herausgezogen. Das ist dieselbe Erweiterung
		// wie `mcpnaht.go` in der Zeile darüber — zwei Namen halten die Kante
		// so eng wie einer, und was nicht geht, bleibt: sie auf das Paket als
		// Ganzes zu öffnen.
		"netz": {
			"scryfall.go", "sammelquelle.go", "scryfalldrossel.go",
			"scryfallverdrahtung.go",
			"scryfallnaehte.go",

			// **Und der Einbettungsklient des Suchindex** (A15, #1514): eine
			// eigene Bauzeile mit eigener Stillstandsfrist, weil ein kaltes
			// Modell seine Gewichte lädt, bevor es den ersten Vektor gibt.
			"retrievalnaht.go",

			// **Und die Zwischenablage für die Flächen ohne Fenster** (#1168).
			// Eine eigene Datei und keine Kante an `kartenablage.go`: dort
			// steht die **Bindung ans Fenster**, hier die Naht für die drei
			// Flächen ohne — dieselbe Trennung wie bei `setloeschnaht.go`.
			"ablagenahtflaechen.go",
			"ersetzungnahtflaechen.go",
			"listennahtflaechen.go",

			// **Und der Freigabezugang** (#1100), nach demselben Muster: eine
			// benannte Datei, kein offenes Paket.
			"aktualisierungverdrahtung.go",

			// **Und die drei neuen Kartendienste** — Werkstück 8 von
			// Phase 4.8 (#1321).
			//
			// **Eine Datei für alle drei**, und das ist der Punkt: sie
			// beantworten dieselbe Frage („wer sucht?"), und sie in drei
			// Dateien zu verteilen hieße, den Verteiler an drei Stellen zu
			// bauen. Was nicht geht, bleibt: die Kante auf das Paket als
			// Ganzes zu öffnen.
			"dienstsuchnaht.go",

			// **Der volle CardTrader-Bestandsexport braucht eine eigene
			// Frist** (Befund 6 des Zweitblicks über Phase 5.6): der
			// Anbieter empfiehlt 120 bis 180 Sekunden, die üblichen
			// dreißig brachen ihn ab.
			"kontoschaunaht.go",

			// **Und die Preise** (#1165) — derselbe Klient wie Auskunft,
			// Abgleich und Bildweg, samt seinen beiden Drosseln. Ein zweiter
			// daneben hielte die Grenze der Gegenstelle nicht ein.
			//
			// **Zwei Dateien**, weil es zwei Aufbauwege gibt: das Fenster
			// verdrahtet beim Start, die Flächen ohne Fenster bauen je Aufruf
			// (§16). Dieselbe Erweiterung wie bei `mcpnaht.go`.
			"sammlungsnaht.go", "sammlungfuerflaechen.go",

			// **Und die Kartensuche** (#1167) — derselbe Klient wie Auskunft
			// und Abgleich, aus demselben Grund: er bringt beide Drosseln mit,
			// und zwei unter derselben Grenze halten sie nicht ein.
			"scryfallsuchnaht.go",
			"listeneinfuhrnaht.go",
			"massennaht.go",

			// **Und die Sammlungssichten** (#1292, Vermerk 0057): sie lesen
			// den Bestand und brauchen **kein** Netz — deshalb stehen sie in
			// dieser Liste und in keiner der beiden darunter.
			"sammlungssichtennaht.go",

			// **Und die Massenvorgänge** (#1292): der Bestand für die
			// Änderungen, Scryfall für den Preisschub — deshalb steht sie in
			// dieser Liste **und** in den beiden darunter.
			"massennaht.go",

			// **Und der Bulkbezug** (A3, #1442): er baut den Netzklienten für
			// MTGJSON. **Nicht** in der Liste für `scryfall` darunter — es ist
			// eine andere Quelle mit einem anderen Ablauf.
			"preisbezugnaht.go",
			"marktbindung.go",
			"anbieterbindung.go",
			"kinaht.go",

			// **Und der Composition Root für die Flächen ohne Fenster**
			// (A13.3, #1452): [app.KIAnwendungUeber] stand bis dahin in
			// `kinaht.go` und ist von dort ausgezogen, als die Datei an die
			// Zeilengrenze aus §2 kam. Die Kante ist dieselbe, nur die Datei
			// ist neu.
			"kianwendung.go",

			// **Und die lokale KI-Laufzeit** (A7, #1446): sie baut den
			// Netzklienten samt Riegel. Dass ihre Adresse auf der Schleife
			// liegt, nimmt sie hier nicht heraus — „Loopback allein beweist
			// keine lokale Inferenz".
			"kilaufzeit.go",

			// **Und Anthropic** (A8, #1447): dieselbe Gestalt, ein anderer
			// Anbieter — Netzklient samt Riegel, und der Riegel fragt hier
			// den Dienstschalter und »nur lokal«.
			"kianthropicnaht.go", "kiopenainaht.go", "kixainaht.go",

			// **Und die zwei Stellen, an denen der Netzweg gelesen wird**
			// (N5, #1571):
			// `netzvorgaben.go` deutet die Einstellungen, `dienstwahl.go`
			// reicht sie weiter. **Beide sind Composition Root** — sie
			// entscheiden, welchen Transport ein Dienst bekommt, und genau
			// das ist die Aufgabe dieser Schicht. Die Fachlogik bleibt
			// draußen: *wie* eine Zahl zu deuten ist, steht in `domain`.
			"netzvorgaben.go", "dienstwahl.go",

			// **Und der Anmeldefluss** — er ist der **zweite** HTTP-Klient
			// dieser Anwendung und bekommt seit N5 dieselbe Proxywahl wie
			// der Rest. `internal/anmeldung` bleibt dabei ein Blatt: es
			// nimmt die Gestalt aus der Standardbibliothek entgegen, und
			// gefüllt wird sie hier.
			"anmeldestart.go",

			// **Und die Auskunft über den Netzweg** (N5, #1571): sie
			// rechnet den wirksamen Stand aus Einstellung und Umgebung —
			// derselbe Weg wie der Transport, und deshalb dieselbe Kante.
			"netzstand.go",
		},
		"scryfall": {
			"scryfall.go", "sammelquelle.go", "scryfallverdrahtung.go",
			"sammlungsnaht.go", "sammlungfuerflaechen.go",
			"scryfallsuchnaht.go",
			"listeneinfuhrnaht.go",
			"massennaht.go",

			// **Und die Nähte, die aus `scryfall.go` herausgezogen wurden**,
			// als es über 400 Zeilen ging (#1241). Dieselbe Erweiterung wie
			// oben bei `netz`, und aus demselben Grund: zwei Namen halten die
			// Kante so eng wie einer.
			"scryfallnaehte.go",

			// **Und der Korpus des Suchindex** (A15, #1514): er liest eine
			// Sammeldatei, und wer sie öffnet, kennt `scryfall.OeffneBulk`.
			// Der Fachkern kennt nur eine [service.Korpusquelle].
			"retrievalnaht.go",
		},
		// **Und drei für den Klienten selbst** (#1100). `aktualisierung.go`
		// meldet, `aktualisierungholen.go` holt — die Grenze zwischen Stufe 1
		// und Stufe 2 ist eine fachliche Entscheidung und gehört deshalb nicht
		// ins Fenster (§2). Beide brauchen `Freigabe` und `IstNeuer`; die
		// Aufzählung bleibt eine Aufzählung.
		"aktualisierung": {
			"aktualisierungverdrahtung.go", "aktualisierung.go",
			"aktualisierungholen.go",
		},
	}
}
