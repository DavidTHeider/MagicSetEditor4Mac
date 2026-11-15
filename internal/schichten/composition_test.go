// **Die Composition Roots, Datei für Datei**
// (#1107).
//
// **Eine eigene Datei**, weil `regeln_test.go` mit den Nähten dieser Änderung
// über die Zeilengrenze aus §2 wuchs. Der Schnitt folgt der Frage: dort,
// **welche Schicht** welche kennen darf, hier, **welche Datei** eine erlaubte
// Kante führen darf.
package schichten

// nurInDatei verengt eine erlaubte Kante auf benannte Dateien.
//
// AGENTS.md §2 erlaubt die Erzeugung der Repositories ausdrücklich **im
// Composition Root** — nicht überall in `internal/app`. Der Wächter kannte bis
// zum Zweitblick am Ende von Phase 1 nur Pakete; damit wäre derselbe Import in
// `menue.go` oder `fehler.go` grün geblieben, und die Ausnahme hätte sich
// still über das ganze Paket ausgebreitet.
//
// **Zwei Dateien, nicht eine** — seit dem 2026-08-22. Es gab immer **zwei**
// Composition Roots: den des Fensters und `MCPSchreiber`, den des zweiten
// Prozesses; sie standen nur in derselben Datei, und der Kommentar an
// `MCPSchreiber` nannte den zweiten seit jeher beim Namen. Getrennt hat sie
// die Zeilengrenze aus §2, und damit stießen zwei Regeln desselben Abschnitts
// zusammen: die eine erzwang den Schnitt, die andere nannte eine einzige
// Datei.
//
// **Aufgelöst wurde es zugunsten der Zahl, nicht des Namens.** Die Ausnahme
// ist eng zu halten, weil sie sich sonst über das Paket ausbreitet — und eine
// Aufzählung von zwei Namen hält sie genauso eng wie einer. Was **nicht** geht,
// ist die Kante auf `internal/app` als Ganzes zu öffnen.
var nurInDatei = map[string]map[string][]string{
	"app": {
		"storage": {
			// **`start.go` gehört zu `app.go`** und ist kein neuer Eintrag im
			// Sinne dieser Aufzählung: es ist dieselbe Verdrahtung, die am
			// 2026-08-26 an der Zeilengrenze aus §2 abgetrennt wurde. Wer sie
			// zusammenschöbe, bekäme wieder eine Datei über 400 Zeilen.
			"app.go", "start.go", "mcpnaht.go", "kartenstand.go", "scryfallnaht.go",
			"skriptnaht.go", "wortlistennaht.go", "schluesselwortnaht.go",
			"werkzeugnaehte.go",
			"packnaht.go",
			"bestandsnaehte.go",
			// **Und der eine Einstellungsverwalter** (OPS-02, #1719): hier
			// standen sechs gleichlautende Zeilen in sechs dieser Dateien;
			// jetzt baut ihn eine, damit die Ordnerprüfung an keiner fehlt.
			"einstellungsverwalter.go",
			"paketnaht.go",
			"stilnaht.go",
			// **Und die Setangaben** (#1107): derselbe Fall wie `packnaht.go`
			// — eine Naht, die Kommandozeile, Dienst und MCP-Server benutzen,
			// weil es dort kein gestartetes `App` gibt.
			"setangabennaht.go",

			// **Und der Vorlagenzeichner** (#1164): dieselbe Gestalt — eine
			// Naht, die eine Karte sucht, mit der eine **ungespeicherte**
			// Vorlage zu zeigen ist. Sie braucht dafür beides, den Bestand und
			// das Paketverzeichnis, und ist damit eine Verdrahtung.
			"vorlagenzeichnernaht.go",

			// **Und die Sammlung** (#1165): sie verdrahtet den Bestand und —
			// wenn es ein Zuhause gibt — den Klienten für die Preise. Zwei
			// Verbraucher in einer Datei, weil beide **denselben** Bestand
			// brauchen und die Preise ohne die Sammlung keine Übersicht
			// ergeben.
			"sammlungsnaht.go",
			"sammlungfuerflaechen.go",

			// **Und der Bulkbezug der Preise** (A3, #1442): er baut den
			// Klienten für MTGJSON und liest den Crosswalk aus dem Bestand.
			//
			// **Eine eigene Datei neben den beiden darüber**, weil er einen
			// **anderen** Ablauf hat: die Preise je Druck kommen von Scryfall,
			// je Abruf einer; hier kommt eine Datei mit allem, und was daraus
			// gebraucht wird, sagt der Bestand.
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

			// **Und der KI-Vorschlag** (A7, #1446): er liest das Profil aus
			// dem Kontext, den `storage` gesetzt hat.
			"kivorschlag.go", "kianbieter.go", "kimodellprobe.go",

			// **Und der Leseweg der Feldherkunft** (A11, #1450): er baut die
			// Ablage über der **laufenden** Verbindung, wie `kimodellprobe.go`
			// die Probenablage. Ein Feld an der Anwendung zeigte nach einem
			// Bestandswechsel auf die alte Datei.
			"kifeldherkunftnaht.go",
			"kibildnaht.go",
			"kibildeinsetzen.go",

			// **Und der Schemastand für die Diagnose** (#1211, Befund 2). Sie
			// nennt, welche Migrationen verbucht sind — ein Fehler nach einer
			// Aktualisierung ist ohne das nicht einzuordnen.
			//
			// **Eine eigene Datei und keine Kante an `diagnose.go`**: dort
			// steht, **was** in der Diagnose steht; hier, **woher** eine der
			// Angaben kommt. Die Aufzählung bleibt eine Aufzählung (§2).
			"diagnosenaht.go",
			// **Der Betriebszustand** (O9, #1433): eine Zeile, die
			// `service` und `storage` verbindet, damit `cmd/` es nicht muss.
			"gesundheitsnaht.go", "vertraeglichkeitsnaht.go",

			// **Und der Scryfall-Start** (#1167). Er stand in `start.go` und
			// ist herausgelöst, weil die Datei mit der Kartensuche über die
			// Zeilengrenze ging — dieselbe Verdrahtung, ein anderer Ort.
			"scryfallstart.go",

			// **Und das Löschen eines Sets** (#1218). Eine eigene Datei und
			// keine Kante an `setloeschen.go`: dort steht die **Bindung ans
			// Fenster**, hier die Naht für die drei Flächen ohne Fenster.
			"setloeschnaht.go",

			// **Und die Scryfall-Verdrahtung** (#1167). Sie baut seit dem
			// Verweis auf einen echten Druck auch die **Ablage** dazu:
			// `Verweisnaht` legt Klient und Bestandsablage zusammen. Ein
			// Aufrufer, der die Ablage selbst mitbrächte, könnte eine zweite
			// über derselben Tabelle bauen — zwei Wahrheiten über denselben
			// Verweis.
			"scryfallverdrahtung.go",
			"scryfallnaehte.go",

			// **Und die Zwischenablage für die Flächen ohne Fenster** (#1168).
			// Eine eigene Datei und keine Kante an `kartenablage.go`: dort
			// steht die **Bindung ans Fenster**, hier die Naht für die drei
			// Flächen ohne — dieselbe Trennung wie bei `setloeschnaht.go`.
			"ablagenahtflaechen.go",
			"ersetzungnahtflaechen.go",
			"listennahtflaechen.go",

			// **Und die Bestandssicherung für die Flächen ohne Fenster**
			// (#1169, Werkstück 2). Sie ist die vierte Wurzel: sie hält den
			// Lebenszyklus einer Verbindung, die ein **Befehl** mitbringt,
			// und nicht den des Fensters.
			//
			// **Ohne sie ginge es nicht**: der Weg selbst liegt seit #1175 in
			// `service.Bestandsweg` und bekommt Orte und Lebenszyklus
			// hineingereicht — irgendwo muss beides für einen Prozess ohne
			// Fenster gebaut werden, und das ist Verdrahtung.
			"bestandssicherungnaht.go",
			"markennaht.go",

			// **Und das Buch über die Rückkehren** (O7, #1431). Es baut eine
			// Ablage über der **laufenden** Verbindung und muss sie deshalb
			// bei jedem Ruf neu holen: nach einem Zurückspielen ist es eine
			// andere, und die Epoche gehört in den zurückgespielten Bestand.
			// Dieselbe Gestalt wie `chiffrenaht`, nur mit eigener Datei —
			// `bestandssicherungnaht.go` ginge mit ihr über die Zeilengrenze
			// aus §2.
			"rueckkehrnaht.go",

			// **Und die Datenpflege** (O8, #1432): sie baut ihre drei Wege
			// über der **laufenden** Verbindung und muss sie deshalb bei
			// jedem Ruf neu holen — dieselbe Gestalt wie `rueckkehrnaht.go`.
			"pflegenaht.go",
			"pflegeflaechen.go",

			// **Und die abgestufte Rücksetzung** (O4, #1428): dieselbe
			// Gestalt und ein Grund mehr — der Neuanfang **tauscht** den
			// Bestand, während die Naht steht.
			"ruecksetzungsnaht.go",

			// **Und das Kostenjournal** (O12, #1439): dieselbe Gestalt — es
			// bucht über die **laufende** Verbindung, und ein Zurückspielen
			// tauscht sie.
			"kostennaht.go",

			// **Und MTG JSON ohne Fenster** (#1291): der Exporteur wird hier
			// aus dem Bestand gebaut, damit Kommandozeile, MCP und Dienst
			// dieselbe Naht nehmen wie das Fenster.
			"mtgjsonnaht.go",

			// **Und das Kartenbild von der Platte** (#1291): dieselben Teile
			// wie `scryfallbildnaht.go`, nur ohne den Weg nach draußen.
			"bilddateinaht.go",

			// **Und die Exportvorlage ohne Fenster** (#1291): dieselben Teile
			// wie im Start, nur für eine Fläche ohne Dialog.
			"vorlagennaht.go",

			// **Und die Listeneinfuhr** (#1292): sie braucht **zwei** Quellen
			// — den Bestand zum Schreiben und Scryfall zum Deuten. Deshalb
			// steht sie in dieser Liste und in den beiden darunter.
			//
			// **Und nicht an `sammlungfuerflaechen.go` angehängt**: die
			// Sammlung zu führen braucht kein Netz, und wer die Einfuhr dort
			// einbaute, machte das Führen von einem Dienst abhängig, den es
			// zum Führen nicht braucht (§16, Regel 5).
			"listeneinfuhrnaht.go",

			// **Und die Sammlungssichten** (#1292, Vermerk 0057): sie lesen
			// den Bestand und brauchen **kein** Netz — deshalb stehen sie in
			// dieser Liste und in keiner der beiden darunter.
			"sammlungssichtennaht.go",

			// **Und die Massenvorgänge** (#1292): der Bestand für die
			// Änderungen, Scryfall für den Preisschub — deshalb steht sie in
			// dieser Liste **und** in den beiden darunter.
			"massennaht.go",

			// **Und die Setwache** (#1336): sie baut den Setleser für die
			// Kommandozeile, die `service` nicht kennen darf.
			"setwache.go",

			// **Und die Kartenskripte** (#1337): sie baut den Kartenschreiber
			// **und** den Stilspeicher — der eine legt die Karten an, der
			// andere kennt den Weg vom Set zum Spielpaket.
			"kartenskriptnaht.go",

			// **Und die Marken an jedem Fachobjekt** (UX-12, #1684) — eine
			// Naht für alle vier Flächen, wie `markennaht.go`.
			"objektmarkennaht.go",
			"auszaehlnaht.go",
			"autoersetzungspflege.go",

			// **Und die Deckwege für die drei Flächen ohne Fenster**
			// (#1293): dieselben Ablagen wie am Fenster, damit alle vier
			// dieselbe Fachlogik benutzen (§16).
			"decknahtflaechen.go",

			// **Und die Deckwege** (#1293): eine Datei baut Deckablage,
			// Formatablage, Verweise und Sammlungsspeicher — vier Ablagen für
			// vier Nähte, und alle über denselben Bestand.
			"deckstart.go",

			// **Und die Turnierwege für die drei Flächen ohne Fenster**
			// (#1294, Phase 4.7): Turnierablage, Deckablage und Setleser —
			// dieselben Ablagen wie am Fenster, damit alle vier dieselbe
			// Fachlogik benutzen (§16).
			//
			// **Die Verdrahtung fürs Fenster steht in `deckstart.go`** und
			// nicht in einer eigenen Datei: Turniere lesen Decks, und beide
			// gehen über dieselbe Deckablage. Eine zweite hieße zwei
			// Verbindungen auf dasselbe.
			"turnierflaeche.go",

			// **Nutzer und Profile für die drei Flächen ohne Fenster**
			// (#1351)
			// — dieselbe Gestalt wie `decknahtflaechen.go`.
			"nutzernaht.go",

			// **Das Profil an jedem Kontext** (#1351): `profilnaht.go` liest
			// den Zeiger und hängt ihn an, `startmigration.go` ruft es nach
			// der Wanderung. Beides ist Verdrahtung und keine Fachfrage.
			"profilnaht.go",
			"startmigration.go",

			// **Die Anmeldung über einen fremden Anbieter** (#1351,
			// Werkstück 4): sie baut den Fluss, liest die Kennung aus dem
			// Bestand und das Geheimnis aus dem Schlüsselbund. Alles davon ist
			// Verdrahtung; wer ein Mensch ist, entscheidet `internal/service`.
			"anmeldestart.go",

			// **Die Wache über vorgezeigte Zugriffstoken** (#1351,
			// Werkstück 10): sie führt `service` und `storage` zusammen, weil
			// der REST-Zugang nur `domain` kennt und `cmd/mse-rest` kein
			// `service`. Sie **entscheidet nichts** — sie ruft die Prüfung und
			// hängt das Profil an den Kontext, den `storage` liest.
			"tokenwache.go",

			// **Der Aufbau um Nutzer, Profile und Konten** (#1351): sechs
			// Nähte auf einmal, herausgelöst aus `start.go`, als die sechste
			// dazukam. Er entscheidet nichts — er setzt zusammen.
			"leutestart.go",

			// **Und das Angebot der Profilvorgaben** (#1351, Werkstück 15):
			// es führt fünf Quellen zusammen — Formate, Decks, Pakete,
			// Lagerorte —, und jede davon hat ihren eigenen Speicher.
			"vorgabennaht.go",

			// **Und die Verschlüsselung des Bestands** (#1351, Werkstück 14):
			// sie hängt `storage.Verschluesseln`, `…Entschluesseln` und den
			// Schlüsselbundweg an die Schnittstelle, die `service` aufstellt.
			// **Eine Naht und keine Entscheidung** — welche Richtung
			// geschaltet wird, steht im Dienst.
			"chiffrenaht.go",

			// **Und die zweite Verbindung samt der Verdrahtung darüber**
			// (#1351, Werkstück 16): `verdrahtung.go` baut jede Naht über
			// **einer** Verbindung — und `nurlesen.go` entscheidet, welche
			// das ist. Beides ist Verdrahtung und keine Fachlogik; **ob**
			// jemand nur lesen darf, sagt `internal/service`.
			"verdrahtung.go",
			"nurlesen.go",

			// **Und der Suchindex** (A15,
			// #1514):
			// eine Datei baut Ablage **und** Schreiber aus derselben
			// Verbindung und die Einbetterquelle darüber — der Fachkern darf
			// keine Laufzeit kennen (§2), und welche gemeint ist, entscheidet
			// die Einrichtung.
			"retrievalnaht.go",

			// **Und das Zeichenwerk** (#886): es steht seit dem 2026-09-12 in
			// einer eigenen Datei, weil `verdrahtung.go` über die
			// Zeilengrenze aus §2 ging. Dieselbe Wurzel, zwei Dateien — die
			// Aufzählung hält die Kante so eng wie eine.
			"verdrahtungzeichnen.go",

			// **Und die Recherche** (A16, #1515): sie liest die Ausgangskarte
			// und die Profilkennung — beides weiß der Bestand, und der
			// Fachkern kennt nur eine [service.Ausgangsquelle].
			"recherchenaht.go",

			// **Und die Peeridentität** (N1, #1454): sie hängt den Bestand
			// und den Schlüsselbund an die zwei Schnittstellen, die
			// [service.Peerwerk] aufgestellt hat. **Die Entscheidung liegt
			// dort**: aus „der Bund hat nichts" werden drei verschiedene
			// Dinge, und keines davon darf eine Naht treffen.
			"peernaht.go",
			"kontaktnaht.go",

			// **Die freigegebenen Sichten** (T3, #1468).
			"freigabenaht.go",
			"tauschnaht.go",
			"handelspostnaht.go", "handelsverbuchungsnaht.go",

			// **Und die zwei Übernahmewege** (Befund 1 des Nachlaufs über
			// Phase 5.3): ein KI-Vorschlag gehört dem Profil, das ihn
			// angelegt hat, und das **wirksame** Profil eines Rufs steht im
			// Kontext, den `imProfil` gefüllt hat — nicht am Zeiger des
			// Fensters. Beide Nähte holen es dort und reichen es an
			// `service` weiter; ein `storage`-Typ überquert die Kante nicht.
			"kiuebernahmenaht.go",
			"kigruppe.go",

			// **Und der Umschlag an den Flächen ohne Fenster** (N4, #1457):
			// dieselbe Gestalt wie `peerflaeche.go` — eine kopflose
			// Anwendung, in die genau die Nähte kommen, die der Umschlag
			// braucht. Ein zweiter Weg daneben wäre eine zweite Antwort auf
			// dieselbe Frage.
			"umschlagflaeche.go",

			// **Und der Netzweg ebenso** (N5, #1571) — dieselbe Gestalt,
			// und sie braucht nur die Einstellungen.
			"netzflaeche.go",

			// **Und die Partien** (G1, #1458) — dieselbe Gestalt wie
			// `turnierflaeche.go` darüber: eine kopflose Anwendung, die ihre
			// Ablage selbst setzt, damit alle vier Flächen dieselbe haben.
			"partieflaeche.go",

			// **Und die zwei Nähte von T8** (#1473): die Marktlinks brauchen
			// die Druckkennungen aus dem Bestand, die Decklisteneinfuhr die
			// Deckablage. Dieselbe Gestalt wie `listeneinfuhrnaht.go` —
			// **eine Naht je Fähigkeit und nicht eine große**, damit ein Ruf
			// nur das öffnet, was er braucht.
			"marktlinkenaht.go",
			"deckeinfuhrnaht.go",

			// **Und die Naht der Kontoschau** (T11, #1476): sie baut
			// denselben Klienten wie der Marktplatz daneben und liest
			// dieselbe Einstellungstabelle.
			"kontoschaunaht.go",
			"handelsleitungsnaht.go",

			// **Und die Symbolbibliothek** (DOM-02, #1697) — dieselbe Gestalt
			// wie `turnierflaeche.go`.
			"symbolbibliotheksflaeche.go",

			// **Und die Zusatzfelder einer Karte** (#2298) — die Naht für
			// Kommandozeile, MCP und REST, mit demselben Dienst wie im Fenster
			// und dem Kartenschreiber, der auf den gelesenen Stand schreibt.
			"zusatznaht.go",
		},
	},
}

// init hängt die Kanten des Transports an denselben Eintrag.
//
// **In `init` und nicht im Literal**, weil das Literal sonst über die
// Zeilengrenze aus §2 ginge (`compositionwurzeln_test.go`). Der Schnitt folgt
// der Frage: dort, welche Datei den **Bestand** erzeugen darf, hier, welche
// einen **Klienten** bauen darf.
func init() {
	for paket, dateien := range transportwurzeln() {
		nurInDatei["app"][paket] = dateien
	}
}

// Befund ist ein einzelner Verstoß.
type Befund struct {
	Paket  string
	Datei  string
	Import string
	Grund  string
}
