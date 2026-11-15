// Die Schichtenregel aus AGENTS.md §2 — als Tabelle und als Prüfung darüber.
//
// Getrennt von `schichten_test.go`: hier steht die Regel samt einer Prüfung,
// die sie an **erfundenen** Importlisten vorführt; dort steht die Anwendung
// auf den wirklichen Baum. Der Grund ist die Gegenprobe (§10) — zu jeder Zeile
// der Tabelle gehört ein Fall, der den Verstoß **herbeiführt** und die Meldung
// festhält, statt ihn einmal von Hand einzubauen und wieder zurückzunehmen.
package schichten

// modulpfad ist das Präfix, an dem ein projekteigener Import erkennbar ist.
const modulpfad = "github.com/DavidTHeider/MagicSetEditor4Mac"

// erlaubt sagt je Paket unter internal/, welche anderen es importieren darf.
//
// Eine leere Liste heißt: **nichts** aus dem Projekt. Ein Paket, das hier fehlt,
// ist nicht etwa frei — es ist unbekannt, und `TestJedesPaketIstGefuehrt` weist
// das ab. Sonst entstünde eine Schicht, die niemand eingeordnet hat, und die
// Regel wäre lautlos umgangen.
// **Die Blätter am Netz stehen daneben** (`regelnnetzblaetter_test.go`):
// acht Zeilen mit derselben Begründung, und die Datei ging mit ihnen über
// die Zeilengrenze aus §2.
var erlaubt = mitNetzblaettern(map[string][]string{
	"domain":    {},
	"msescript": {},

	// **`laufbeleg` ist ein Blatt und kennt nichts aus dem Projekt** — QA-01
	// (#1733).
	// Es ist ein Werkzeug der Entwicklung, das Messläufe belegt, und kein
	// ausgeliefertes Programm kennt es: dass keine Schicht es einbindet,
	// hält diese Tabelle, denn keine nennt es als Ziel.
	"laufbeleg": {},
	// **`dateisperre` kam am 2026-08-27 dazu** — der Zweitblick über Phase 3.3.
	// Prüfsumme und Tausch waren nicht serialisiert, und zwei Fenster
	// überschrieben einander. Dieselbe Kante wie bei `storage`, `netz` und
	// `scryfall`, und aus demselben Grund: der Kniff steht einmal statt
	// viermal.
	"msepaket": {"domain", "dateisperre"},
	"render":   {"domain"},

	// **`bezug` ist ein Blatt und kennt das Netz.** Es holt Archive und packt
	// sie sicher aus; was ein MSE-Paket ist, weiß es nicht. Dieselbe Trennung
	// wie bei `schriften`: die Plattformkenntnis liegt außen, die Fachlichkeit
	// innen.
	//
	// **`netz` kam am 2026-08-23 dazu**, und es ist dieselbe Richtung wie bei
	// `schriften` → `render`: der Transport steht dort, wo er für **alle**
	// Dienste steht, und `bezug` benutzt ihn. Vorher stand er hier — mit einem
	// zweiten Verbraucher wäre er eine Kopie geworden
	// (#1002).
	"bezug": {"domain", "netz"},

	// **`mcp` ist ein Blatt und kennt nur die Domäne.** Es beschreibt seine
	// Naht zum Bestand als eigene Schnittstelle; wer sie erfüllt, entscheidet
	// der Composition Root des Servers. Ein Protokollpaket, das `storage`
	// kennte, ließe sich nur mit einer Datenbank prüfen.
	"mcp": {"domain"},

	// **Die Nähte des MCP-Servers als Paket** — Werkstück U12 von Phase 5.1
	// (#1423): dünne Adapter von `mcp.Naehte` auf `app`, die der Server
	// **und** die Kommandozentrale im Fenster füllen. Sie kennen `app`, weil
	// sie dessen Nähte rufen — nicht umgekehrt.
	"mcpnaehte": {"app", "domain", "mcp", "service", "storage"},

	// **`scryfall` ist ein Blatt und kennt `domain` und `netz`.**
	//
	// Es bringt die **zweite und dritte** Schicht aus §16 mit — das fremde DTO
	// und die Zuordnung —, der Transport steht in `netz`. Dieselbe Richtung
	// wie bei `bezug`: was das Netz kann, steht an einer Stelle; was eine
	// Scryfall-Karte ist, steht hier.
	//
	// **`service` steht ausdrücklich nicht dabei.** Ein Klient, der
	// Anwendungsfälle kennte, wäre kein Blatt mehr — und die Zuordnung ließe
	// sich nicht ohne Bestand prüfen.
	// **`dateisperre` kam am 2026-08-26 dazu** (#1163, Befund Z7-03): die
	// Sammeldatei wiegt 392 MB, und zwei Prozesse luden sie beide. Dieselbe
	// Kante wie bei `netz` und `storage`, und aus demselben Grund — der
	// Kniff steht einmal statt dreimal.
	"scryfall": {"domain", "netz", "dateisperre"},

	// **`aktualisierung` ist ein Blatt wie `scryfall`** — Werkstück 1 von
	// #1100.
	//
	// Dieselben drei Schichten aus §16: der Transport steht in `netz`, das
	// fremde DTO (`freigabeDTO`) und die Zuordnung (`zuFreigabe`) stehen hier.
	// Es weiß, was eine Freigabe von GitHub ist — und nichts davon, was diese
	// Anwendung damit tut.
	//
	// **`service` steht ausdrücklich nicht dabei**, aus demselben Grund wie
	// dort: ein Klient, der Anwendungsfälle kennte, wäre kein Blatt mehr.
	"aktualisierung": {"domain", "netz"},

	// **`netz` ist ein Blatt und kennt `domain` für seine Fehler.**
	//
	// Es ist die erste der drei Schichten aus §16 — Transport, fremdes DTO,
	// Zuordnung —, und die beiden anderen liegen im Paket des jeweiligen
	// Dienstes. Was eine Karte ist, weiß es nicht; was eine Frist ist, wissen
	// die Dienste nicht.
	"netz": {"domain", "dateisperre"},

	// **`dateisperre` ist ein Blatt ohne jede Kante** — auch `domain` kennt es
	// nicht (#1163).
	//
	// Was es tut, ist Betriebssystem und nicht Fachlichkeit: eine
	// Nachbardatei, ein `flock`, und auf Windows nichts.
	//
	// **Zwei Verbraucher, und das ist der Grund für das Paket**: `storage`
	// hält den Bestand über einen Tausch (Z6-02), `netz` teilt sich das
	// Hostbudget einer Gegenstelle (Z7-03). Zwei Kopien desselben Kniffs
	// liefen an dem Tag auseinander, an dem eine von ihnen etwas dazulernte —
	// dieselbe Überlegung, aus der `internal/netz` entstanden ist.
	"dateisperre": {},

	// **`rest` ist ein Blatt wie `mcp` — und aus demselben Grund.**
	//
	// Es beschreibt die Anwendungsfälle, die es anbietet, als **eigene**
	// Schnittstellen; wer sie erfüllt, entscheidet `cmd/mse-rest`. Ein
	// Protokollpaket, das `service` oder `storage` kennte, ließe sich nur mit
	// einer Datenbank prüfen — und es hätte einen Weg, eigene Fachlogik
	// anzusammeln.
	//
	// **Das zweite ist hier keine Stilfrage.**
	// [0046](../../docs/entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md)
	// nennt „er bekommt eigene Fachlogik" als eine der drei Bedingungen, unter
	// denen die Entscheidung für diesen Dienst wieder fiele. Diese Zeile ist
	// die strukturelle Seite davon.
	"rest": {"domain"},

	// **`mtgjson` ist ein Blatt wie `msepaket`.** Es liest und schreibt ein
	// fremdes Kartenformat und kennt dafür `domain` — mehr nicht. Wer ihm
	// `service` gäbe, machte aus einem Formatleser einen Anwendungsfall, und
	// die Zuordnung ließe sich nicht mehr ohne Bestand prüfen.
	"mtgjson": {"domain"},

	// **`anmeldung` ist ein Blatt** — Werkstück 4 von Phase 4.12
	// (#1351).
	//
	// Es kennt `domain` und sonst nichts aus dem Projekt: es weiß, wie ein
	// OAuth-Fluss geht, nicht, wer ein Nutzer ist und wo er liegt.
	//
	// **Und es kennt `netz` nicht**, anders als die Kartendienste: der Klient
	// wird ihm gereicht, und den Horcher auf der Schleifenadresse baut es
	// selbst — dafür gibt `netz` nichts her.
	"anmeldung": {"domain"},

	// **`kennwort` ist ein Blatt** — Werkstück 6 von Phase 4.12
	// (#1351).
	//
	// Es hütet Kennwörter mit Argon2id und kennt `domain` für die
	// Fehlerklassen. Was ein Nutzer ist, weiß es nicht — und deshalb lässt es
	// sich für sich prüfen, ohne Bestand und ohne Fenster.
	"kennwort": {"domain"},

	// **`biometrie` ist ein Blatt** — Werkstück 7 desselben Issues. Es fragt
	// `LocalAuthentication` und weiß nichts über Nutzer; `prozessspeicher`
	// fragt ebenso das System, nach dem Speicher, ohne jede Kante (OPS02).
	"biometrie":       {"domain"},
	"prozessspeicher": {},

	// **`totp` ist ein Blatt** — Werkstück 7 von Phase 4.12
	// (#1351).
	// Es rechnet RFC 6238 und weiß nichts über Nutzer; `domain` kennt es für
	// seine Fehlerklassen.
	"totp": {"domain"},

	// **`vcard` ist ein Blatt ohne Kante** (ID-04, #1703), wie `msescript`.
	"vcard": {},

	// **`peer` ist ein Blatt** — Werkstück N1 von Phase 5.4
	// (#1454).
	//
	// Es hält die Identität dieses Programms gegenüber anderen: ein
	// Ed25519-Paar, seine Kennung in der Gestalt, die libp2p festlegt, und
	// die Unterschrift darunter. **Es kennt keinen Schlüsselbund** — wo das
	// Geheimnis liegt, entscheidet `app`, so wie bei `kennwort` und `totp`.
	// **Und kein libp2p**: die Kennung ist gerechnet, nicht bezogen; der
	// Adapter kommt in N2.
	"peer": {"domain"},

	// **`schluesselbund` ist ein Blatt** — Phase 4.8
	// (#1321).
	//
	// Es legt Geheimnisse dort ab, wo macOS sie hütet, und kennt dafür
	// `domain` — für die Fehlerarten und sonst nichts. Dieselbe Trennung wie
	// bei `schriften`: die **Plattformkenntnis** liegt außen, die Fachlichkeit
	// innen.
	//
	// **Und es kennt `netz` ausdrücklich nicht.** Ein Schlüsselbund ruft
	// niemanden an; wer hier eine Netzkante sähe, hätte einen Fehler gefunden.
	"schluesselbund": {"domain"},

	// **`sammelliste` ist ein Blatt wie `mtgjson`** — Phase 4.5
	// (#1292).
	//
	// Es zerteilt eine eingefügte Kartenliste in Zeilen: Text hinein, Zeilen
	// heraus. Es kennt `domain` wegen [domain.Befund] und
	// [domain.ErrValidation] — und sonst nichts. **Die Zuordnung** einer Zeile
	// auf einen bestimmten Druck braucht Scryfall und steht in `service`; ein
	// Zerteiler, der ins Netz griffe, ließe sich ohne Netz nicht prüfen.
	"sammelliste": {"domain"},

	// **`schriften` kennt `render` — und das ist dieselbe Richtung wie bei
	// `storage` und `service`.**
	//
	// `render` sagt, **was** es braucht (`Schriftquelle`, `Glyphenschrift`);
	// `schriften` weiß, **wo** auf dieser Plattform Schriften liegen. Ein
	// Rasterer, der Verzeichnisse kennt, wäre einer, der sich nur auf einem
	// Rechner prüfen lässt.
	"schriften": {"render"},

	// **`leitung` kennt `domain` und `peer` — und `service` gerade nicht.**
	//
	// Es trägt Nachrichten über einen Strom und erfüllt `service.Verbindung`
	// **strukturell**: Go prüft Schnittstellen über die Gestalt, und ein
	// Import dorthin wäre eine Kante, die dieses Blatt nicht braucht. Was
	// den Strom aufbaut und die Leitung als `service.Verbindung` weiterreicht,
	// ist der Composition Root.
	//
	// **`peer` steht hier wegen zweier Dinge**: [peer.Kennung], weil die
	// Schnittstelle genau diesen Typ verlangt, und [peer.Hoechstlaenge],
	// weil die Grenze eines Rahmens dieselbe sein muss wie die einer Hülle —
	// zwei Zahlen liefen auseinander (N2,
	// #1455).
	"leitung": {"domain", "peer"},

	// **Die Blätter am Netz und die drei Ausnahmen stehen daneben** —
	// `regelnnetzblaetter_test.go` und `regelnnaehte_test.go`.
	// **`entdeckung` kennt `domain` und `peer` — wie `leitung`, und aus
	// demselben Grund** — Werkstück N2 von Phase 5.4
	// (#1455).
	//
	// Es findet Gegenüber im selben Netz und trägt dabei eine **Peerkennung**
	// hinaus und herein; was sie ist, weiß `peer`. Was ein Kontakt ist und
	// was er darf, weiß es nicht — das entscheidet `service` nach dem
	// Handschlag.
	"entdeckung": {"domain", "peer"},

	// **`svg` ist ein Blatt und kennt `render` — wie `schriften`, und aus
	// verwandtem Grund.**
	//
	// Es schreibt einen fertigen Zeichenplan als SVG und ist damit der
	// **zweite Verbraucher** desselben Plans
	// (#837).
	// Genau das ist die Naht, die Phase P prüft: trägt die Trennung aus §2 —
	// `render` bekommt einen Plan, keine Skripte — auch dann, wenn jemand
	// anderes als der Rasterer ihn liest?
	//
	// **Die Richtung stimmt**: `svg` kennt `render`, nicht umgekehrt. Ein
	// Rasterer, der von SVG wüsste, wäre die Vermischung, die §2 verhindert.
	"svg": {"render"},

	// **`pdf` kennt nichts aus dem Projekt.** Es setzt Bilder und Linien auf
	// Seiten; was eine Karte ist, weiß es nicht. Dieselbe Trennung wie bei
	// `msescript`: ein Formatschreiber, der eine Kartenentität kennte, ließe
	// sich nur mit einer Karte prüfen.
	"pdf": {},

	// **`vorschauweg` kennt nichts aus dem Projekt** —
	// #1270.
	//
	// Es nimmt Bytes entgegen, gibt eine Adresse zurück und liefert sie über
	// die Assetleitung des Fensters wieder aus. Was ein Kartenbild ist, weiß
	// es nicht — dieselbe Trennung wie bei `pdf`.
	//
	// **Nicht einmal `domain`**, und das ist die knappste Stelle in dieser
	// Liste: es hat keinen fachlichen Fehler zu melden. Ein Bild ist da oder
	// verdrängt, und das sagt ein 404.
	"vorschauweg": {},

	// **`sicherung` ist ein Blatt wie `pdf` — es kennt `domain` und sonst
	// nichts** (#1162).
	//
	// Es packt benannte Dateien in ein Archiv und holt sie wieder heraus,
	// samt Manifest und Prüfsummen. **Was** in eine Sicherung gehört, weiß es
	// nicht — das entscheidet `internal/app`, der einzige Ort, der Bestand
	// und Beiwerk zugleich kennt.
	//
	// `domain` steht dort für die Fehlerarten: eine beschädigte Sicherung ist
	// ein `ErrConflict`, eine fremde Datei ein `ErrValidation`, und beides
	// muss ein Aufrufer unterscheiden können (§8).
	"sicherung": {"domain"},

	// **`schluesselbund` seit 0072**: derselbe Grund wie bei `dateisperre`.
	"storage": {"domain", "dateisperre", "schluesselbund"},

	// **`svg` kam am 2026-08-22 zu `service` dazu** (#852), und der Grund
	// gehört danebengeschrieben: SVG ist ein **Ausgabeformat der Ausfuhr**,
	// genau wie PDF — `service.Bildausfuhr` entscheidet, welches gewünscht
	// ist, und reicht den Plan weiter. Ohne diese Kante stand das Paket da
	// und niemand konnte eine SVG erzeugen.
	//
	// **Die Richtung bleibt dieselbe wie bei `pdf`**: der Dienst kennt den
	// Formatschreiber, nicht umgekehrt.
	// **`scryfall` kam am 2026-08-23 dazu** (#1005), und aus demselben Grund
	// wie `bezug`: es ist eine **Quelle**, aus der ein Anwendungsfall schöpft.
	// `service.Scryfallabgleich` holt eine Karte, zeigt den Unterschied und
	// schreibt erst auf Anweisung.
	//
	// **Die Richtung bleibt dieselbe**: der Dienst kennt den Klienten, nicht
	// umgekehrt. Ein Klient, der Anwendungsfälle kennte, wäre kein Blatt mehr.
	//
	// **`netz` steht ausdrücklich nicht dabei.** Wer eine Frist oder eine
	// Drossel einstellt, tut das im Composition Root; ein Anwendungsfall, der
	// HTTP kennt, hätte den Transport wieder in der Fachlichkeit.
	//
	// **`sicherung` kam am 2026-08-28 dazu** (#1175) — dieselbe Richtung wie
	// `pdf` und `svg`: ein Blatt, das ein **Ausgabeformat** schreibt, und der
	// Dienst entscheidet, was hineingehört. Es stand bis dahin bei `app`,
	// und das war die Ausnahme, die diese Zeile ersetzt.
	// **`dateisperre` kam am 2026-08-28 dazu** (#1190), und der Grund steht im
	// Zuschnitt des Issues: die Schreibsperre der Paketwerkstatt liegt an
	// einer **Paketdatei**, der Bezug tauscht einen **Baum** aus. Wer die
	// Ebenen ordnet, ordnet sie hier — `msepaket` weiß, wie ein Paket
	// aussieht; **wo** die Bezugsbäume liegen, weiß nur der Dienst.
	// **`sammelliste` kam am 2026-08-30 dazu** (#1292) — dieselbe Richtung
	// wie `msepaket` und `mtgjson`: ein Blatt, das ein **Eingabeformat** liest,
	// und der Dienst entscheidet, was daraus wird. Zerteilen ist rein; die
	// Zuordnung einer Zeile auf einen Druck braucht Scryfall und gehört
	// deshalb hierher und nicht ins Blatt.
	"service": {
		"bezug", "dateisperre", "domain", "msepaket", "msescript", "mtgjson",
		"pdf", "render", "sammelliste", "scryfall", "sicherung", "svg",

		// **Und der zweite Faktor** — Werkstück 7 desselben Issues, aus
		// demselben Grund wie das Kennwort: **ob** ein Code stimmt, ist eine
		// fachliche Entscheidung, und die fällt in `service`.
		"totp",

		// **Und das Kennwort** — Werkstück 6 von Phase 4.12
		// (#1351).
		//
		// **Anders als `anmeldung` steht es hier und nicht bei `app`**: wer
		// hereinkommt, ist eine fachliche Entscheidung, und sie fällt in
		// `service`. Das Paket ist ein Blatt wie `pdf` — es weiß, was
		// Argon2id ist, und nichts über Nutzer.
		//
		// **Was dadurch nicht entsteht**: eine Kante von `domain` dorthin.
		// [domain.Nutzer.Kennwort] trägt das Ergebnis und prüft es nicht —
		// ein Hash hat keine fachliche Form.
		"kennwort",

		// **Und die vCard** — ID-04 (#1703), wie `sammelliste`: ein Blatt,
		// das ein Eingabeformat liest, und der Dienst gleicht ab.
		"vcard",

		// **Und die Peeridentität** — Werkstück N1 von Phase 5.4
		// (#1454),
		// aus demselben Grund wie `totp` und `kennwort`: **ob** eine
		// Identität gilt, ist eine fachliche Entscheidung.
		//
		// **Und sie ist keine, die eine Naht treffen darf.** Aus „der
		// Schlüsselbund hat nichts" werden drei verschiedene Dinge, je
		// nachdem, was im Bestand steht; das Issue verlangt, dass keines
		// davon **still** behoben wird. Eine Regel in der Naht stünde dort,
		// wo niemand sie sucht.
		"peer",
	},
	// **`netz` und `scryfall` kamen am 2026-08-23 dazu** (#1005), und beide
	// nur für den **Composition Root**: `app.Scryfallnaht` baut Klient,
	// Frist, Drossel und Zwischenspeicher und reicht sie dem Dienst.
	//
	// **Das ist dieselbe Aufzählung wie bei `storage`**: `app` darf sie
	// kennen, weil es verdrahtet — und die Fachlichkeit dahinter kennt sie
	// nicht.
	// **`aktualisierung` kam am 2026-08-25 dazu** (#1100), aus demselben Grund
	// wie `netz` und `scryfall`: der Composition Root baut den Klienten.
	//
	// **`sicherung` stand hier vom 2026-08-26 bis zum 2026-08-28** (#1162,
	// dann #1175). Die Begründung lautete: `app` sei der einzige Ort, der
	// **Datenbankpfad und Bilderverzeichnis zugleich** kennt.
	//
	// **Das stimmte und trug trotzdem nicht.** Wer die Orte kennt, muss
	// deshalb nicht die Reihenfolge entscheiden. Der ganze Rückspielweg —
	// welche Archivfassungen zulässig sind, in welcher Folge der zweiteilige
	// Tausch läuft, wann ein abgebrochener nachgeholt wird — lag in der
	// Fassade, und §2 verbietet dort fachliche Entscheidungen. Diese Ausnahme
	// hielt also nicht eine Kante offen, sondern eine ganze Schicht am
	// falschen Platz.
	//
	// Er liegt seither in `service.Bestandsweg`; die Orte und der
	// Lebenszyklus werden **hineingereicht** — dieselbe Gestalt wie beim
	// Ausschneiden der Zwischenablage (#1244).
	//
	// **`vorschauweg` kam am 2026-08-28 dazu** (#1270): die Fassade legt ein
	// gerastertes Bild dort ab und gibt seine **Adresse** in die Antwort,
	// statt 242 KB Base64 durch die Wails-Naht zu schicken. Sie kennt das
	// Blatt, wie sie `pdf` und `svg` kennt — und **nicht** `net/http`; der
	// Rückgabetyp von `app.Bildweg` ist `*vorschauweg.Weg`, damit weder hier
	// noch in `main.go` ein Netzimport entsteht (#54).
	"app": erlaubtInApp,

	// **Die Hülle kennt die Anwendung — und nicht umgekehrt** (ENG-02,
	// #1727).
	// `internal/app` beschreibt, was das Fenster tun soll, `fenster` setzt es
	// auf Wails 3 um. `vorschauweg` kennt sie für die Weiche vor dem
	// Assetserver: welche Wege Bilder sind, steht dort an **einer** Stelle.
	"fenster": {"app", "vorschauweg"},

	// **Die Wurzelknoten stehen daneben** (`regelnwurzeln_test.go`): vier Zeilen mit derselben
	// Begründung, und die Datei ging mit ihnen über die Zeilengrenze aus §2.
})
