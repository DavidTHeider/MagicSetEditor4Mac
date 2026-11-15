package app

// **Woraus die Anwendung besteht** — der Typ und seine Nähte.
//
// Eigene Datei, weil `app.go` mit ihr über 400 Zeilen ginge (§2). Der Schnitt
// folgt der Frage: dort steht, **was** ein Start tut und **wie** die Teile
// zusammenkommen, hier, **welche** es sind.
//
// **Der Composition Root bleibt `app.go`.** Diese Datei erzeugt nichts; sie
// erklärt nur. Der Schichtenwächter verlangt das ausdrücklich — die Erzeugung
// der Repositories ist eine Ausnahme für **eine** Datei, und eine Ausnahme,
// die auf zwei wächst, wächst weiter.

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// App ist der Composition Root (AGENTS.md §2) — hier und nur hier entstehen
// Datenbank, Repositories und Services und werden einander gereicht. Irgendwo
// muss die Verdrahtung stehen; sie an einer Stelle zu halten ist der Grund,
// warum die übrigen Schichten einander nicht kennen müssen.
type App struct {
	// protokollweiche ist die Weiche aus `main`; `nil` ohne sie (OPS-04).
	protokollweiche *Protokollweiche

	datenbankpfad string
	db            *sql.DB

	// lauf ist, was der Schluss braucht (BK-07a, `lebenszyklus.go`).
	lauf schlusslage

	// planlauf ist der freiwillige Plan der Sicherung (OPS01, #2347).
	planlauf sicherungsplanlauf

	// bulkadresse ist der Ort, an dem MTGJSON seine Preisdateien führt —
	// leer heißt: die Vorgabe des Pakets (A3, #1442). Sie steht hier, damit
	// ein Lauf gegen eine eigene Gegenstelle ohne Netz auskommt (§3).
	bulkadresse string

	// marktbund ist der Schlüsselbund, aus dem die Marktzugänge kommen —
	// nil heißt: der echte (A4, #1443).
	//
	// **Als Feld, damit ein Lauf ohne Schlüsselbund auskommt** (§3): ein
	// Test, der den echten fragte, öffnete auf diesem Rechner einen Dialog.
	marktbund Geheimnisleser

	// peerbund hält die eigene Peeridentität — nil heißt: der echte; was
	// sonst zu Peerverbindungen gehört, steht in `peersitzungnaht.go`.
	peerbund Peerbund
	peerlage peerfelder

	// fernlage steht neben `peerlage` und nicht darin — der Grund steht bei
	// [fernfelder] (N3, #1456).
	fernlage fernfelder

	// **Was sie über KI-Laufzeiten weiß**, steht in `kilaufzeit.go`.
	kinaehte

	betrieb betriebsgedaechtnis // das Betriebsband (OPS02), `betriebsband.go`

	// profil ist das Profil, dessen Daten dieses Fenster zeigt; warum atomar
	// und was die Null heißt, steht an [App.imProfil].
	profil atomic.Int64

	// umschaltung hält jede Bindung an, solange ein Profilwechsel umbaut.
	// Warum und wo genommen: [App.ProfilWechseln], [App.imProfil].
	umschaltung sync.RWMutex

	// **Wer hier arbeitet und wie er hereinkommt**, steht in
	// `anwendungleute.go`.
	leutenaehte

	// **Was sie über fremde Dienste weiß**, steht in `fremddienste.go`.
	fremddienste

	// **Was der Bestand als Datei ist** — und wer an ihn darf. Steht in
	// `anwendungbestand.go`.
	bestandsnaehte

	// oeffnen ist die Naht zur Datenbank.
	//
	// Als Feld und nicht als fester Aufruf, damit die Fehlerpfade darunter
	// prüfbar sind: eine Datenbank, die sich nicht schließen lässt, kommt mit
	// SQLite auf einer echten Platte nicht zustande — im Betrieb aber schon,
	// und dann entscheidet dieser Zweig, ob die Anwendung es meldet oder still
	// darüber hinweggeht.
	oeffnen func(context.Context, string) (*sql.DB, error)

	// **Wo etwas auf dieser Platte liegt** — sieben Nähte, und alle
	// beantworten dieselbe Frage. Sie stehen in `orte.go`.
	ortsnaehte

	// rueckspielGeprueft sorgt dafür, dass ein abgebrochenes Zurückspielen
	// **einmal je Prozess** nachgeholt wird (#1162).
	rueckspielGeprueft sync.Once

	// bestand ist die Naht zum Leser.
	//
	// Als Feld und aus demselben Grund wie `oeffnen`: die Fehlerpfade der
	// Bindungen sind sonst nur über eine kaputte Datenbank zu erreichen, und
	// dann prüft der Test die Datenbank statt die Übersetzung.
	bestand         Bestandsleser
	schreiber       Kartenschreiber
	zusatzwerte     Zusatzwertpflege
	massenschreiber Massenschreiber
	kartenleser     Kartenleser

	// marken sind die Marken an Karten (#1208), objektmarken die an jedem
	// Fachobjekt (UX-12, #1684, Vermerk 0135).
	marken       Markenpflege
	objektmarken *Objektmarkennaht

	// vorschaubilder macht die kleinen Bilder der Kartenliste (#1166).
	//
	// **`nil` ist erlaubt und heißt: die Liste bleibt Text.** Eine Fläche ohne
	// Zeichenwerk — die Kommandozeile etwa — hat keine Bildquelle, und eine
	// Liste ohne Bilder ist dort die richtige Antwort.
	vorschaubilder *service.Vorschaubildwerk

	// Die Zwischenspeicher des Zeichenwerks — OPS-08 (#1725).
	werkspeicherfelder

	// letzteDiagnose ist der Text, den [App.Diagnose] zuletzt gezeigt hat.
	//
	// **Er wird gesichert und kein zweiter erzeugt** — die Zustimmung gilt dem
	// Text, den jemand gelesen hat (#1211, Befund 5).
	letzteDiagnose string

	// letzteVorschau und letzteTeile schreibt [App.DiagnosepaketNach] (O10),
	// abzüglich abgewählter Teile (OPS-04.AK-04). `nil`: noch keine Vorschau.
	letzteVorschau *DiagnosepaketAntwort
	letzteTeile    []paketteil

	// letzterPaketort ist der Ort des zuletzt geschriebenen Pakets — der
	// einzige, den [App.DiagnosepaketLoeschen] anfassen darf.
	letzterPaketort string

	// schemaquelle sagt der Diagnose, welche Migrationen verbucht sind.
	schemaquelle Schemastand

	// gesundheit sagt, ob dieser Prozess lebt und arbeiten kann (O9, #1433).
	gesundheit Gesundheitsnaht

	// ablage kopiert Karten zwischen Sets (#1168).
	ablage Ablagequelle

	// ersetzung sucht und ersetzt über ein Set (#1168).
	ersetzung Ersetzungsquelle

	// listenbau baut die Spalten und die Ordnung der Kartenliste (#1168).
	listenbau Listenquelle

	// ablageLesen und ablageSchreiben sind die Naht zur Zwischenablage.
	ablageLesen     func(context.Context) (string, error)
	ablageSchreiben func(context.Context, string) error

	auskunft  Auskunftsquelle
	planbauer Planbauer
	vorlagen  Vorlagenquelle
	rasterer  Rasterer

	// vorschaulauf hält den letzten Vorschauauftrag — §7, der Abbruch.
	vorschaulauf abloesenderLauf

	// packlauf hält den letzten Packartenauftrag. **Ein eigener und nicht
	// derselbe**: eine Vorschau und eine Packartenliste sind zwei
	// Bezugsobjekte, und wer sie zusammenlegte, bräche die Vorschau, sobald
	// jemand die Packlasche öffnet.
	packlauf     abloesenderLauf
	sucher       Sucher
	statistiker  Statistiker
	importeur    Importeur
	setloeschung Setloescher
	exporteur    Exporteur
	speicherer   Speicherer
	bildausfuhr  Bildausfuehrer

	// Die drei Nähte, die eine Datei ohne Set schreiben: der Tischabzug
	// (#1713), die Partiedatei (#1711) und die Turnierdatei samt Bericht (#1699).
	tischabzug   Tischabzugschreiber
	partiedatei  Partiedatei
	turnierdatei Partiedatei

	bogenleger    Bogenleger
	paketbezieher Paketbezieher
	pakete        Paketfuehrer
	paketbezuege  Paketaktualisierer

	// einstellungen ist die Naht zu den Anwendungseinstellungen (#245) —
	// Sprache, Erscheinungsbild, Schriftgröße; `austausch` führt sie aus und
	// ein (Werkstück O3 von Phase 5.2, #1427).
	einstellungen Einstellungsfuehrer

	// ruecksprung führt die Rückkehren — Werkstück O7 von Phase 5.2
	// (#1431).
	ruecksprung Ruecksprungfuehrer

	// pflege führt die Datenpflege — Werkstück O8 von Phase 5.2
	// (#1432).
	pflege Pflegefuehrer

	// ruecksetzung führt die abgestufte Rücksetzung — Werkstück O4 von Phase
	// 5.2 (#1428).
	ruecksetzung Ruecksetzungsfuehrer

	// laufbuch hält die Bildaufträge je Karte und die Generation je Lauf.
	laufbuch

	// **Die Wissensnähte als Verbund** — eingebettet, damit diese Datei unter
	// der Zeilengrenze aus §2 bleibt (`wissensnaehte.go`), und zusammen, weil
	// sie eine Frage teilen.
	wissensnaehte

	// kosten führt das Journal externer Aufträge — Werkstück O12 von Phase
	// 5.2 (#1439).
	kosten Kostenfuehrer

	austausch Einstellungsaustauscher

	// fassungen sagt, welche Freigabe die neueste ist — `nil`, wenn kein
	// Zugang eingerichtet ist (#1100).
	fassungen Fassungsanleger

	// vorpruefer sagt, ob die neueste Freigabe diesen Bestand öffnen kann —
	// `nil`, solange kein Bestand offen ist (O6, #1430).
	vorpruefer *service.Vorpruefer

	// exportoptionen hält die Optionswerte je Exportvorlage (#492).
	exportoptionen Optionsablage
	stile          Stilwaehler
	autoersetzung  Autoersetzungsquelle

	// decks, formate, deckpruefung und deckabgleich sind die Nähte der
	// Phase 4.6 (#1293) — Decks bauen, Formate führen, prüfen, abgleichen;
	// turniere die der Phase 4.7 (#1294).
	decks        Decknaht
	formate      Formatnaht
	deckpruefung Deckpruefnaht
	deckabgleich Deckabgleichnaht

	// deckzuordnung ist die Naht von U9 (#1420) — welche Exemplare in einem
	// Deck liegen sollen, und was davon noch frei ist.
	deckzuordnung Deckzuordnungsnaht
	turniere      Turnierleitung

	// matchleitung bindet Turnierpaarungen an Partien (G7,
	// #1464).
	//
	// **Neben der Turnierleitung und nicht darin**: ein Turnier ohne Partien
	// braucht sie nicht, und eine Fläche, die nur Runden legt, soll keinen
	// Partiespeicher mitbringen müssen (§2).
	matchleitung Matchleitung

	// profilkenner sagt, welche Profile es gibt (G8,
	// #1465).
	//
	// **Nil heißt: keine Auskunft.** Eine Fläche ohne Bestand hat keine
	// Profile, und ein Befund von dort wäre eine Aussage über nichts.
	profilkenner Profilkenner

	// partie ist alles, was eine Partie braucht: die zwei Nähte aus G1
	// (#1458)
	// und die Freigabe samt Beitritt aus G5
	// (#1462).
	//
	// **In `partienaht.go`**, damit diese Datei unter der Zeilengrenze
	// bleibt — und weil dort steht, warum eine Naht und ein Lauf in
	// demselben Verbund liegen dürfen.
	partie partienaht

	// Die Nähte der Spieltexturen — `texturnaht.go`, eingebettet wegen §2.
	texturnaht

	// anleger legt neue Karten an (#562), loescher entfernt sie wieder (#621).
	anleger  Kartenanleger
	loescher Kartenloescher

	// kopierer legt eine zweite mit denselben Werten an (#625).
	kopierer Kartenkopierer

	// tauscher ordnet sie um (#628).
	tauscher Kartentauscher
	verlauf  Verlaufsfuehrer
	standnaht

	// paketfehler hält fest, warum es kein Paketverzeichnis gibt — falls es
	// keines gibt. Er kommt beim ersten Bezug ans Licht, nicht beim Start.
	paketfehler error

	// fenster ist der Wails-Kontext; er kommt mit `Bereit`, nicht mit `Starten`.
	fenster context.Context //nolint:containedctx // Wails gibt ihn nur so her

	// Die Nähte der Menüleiste und der Stand, den das Fenster ihr meldet —
	// `menuenaht.go`, eingebettet wegen §2 (UXR-F14, #2029).
	menuenaht

	// dateiwahl ist die Naht zum Dateidialog — aus demselben Grund wie
	// `menueSetzen`: er braucht ein Fenster, und ein Test hat keines.
	dateiwahl func(context.Context, Dateiauswahl) (string, error)

	// einzelkarte gibt eine einzelne Karte (#891).
	einzelkarte Einzelkartenleser

	// zeilen gibt die Karten eines Sets als Liste (#891).
	zeilen Kartenzeilenleser

	// feldwerte gibt ein Feld über ein ganzes Set (#891).
	feldwerte Kartenfeldwerteleser

	// Die Nähte der Übersicht — `uebersichtsnaehte.go`, eingebettet wegen §2.
	uebersichtsnaehte

	// wartendeDateien hält die Pfade, die der Finder vor dem Fenster schickte.
	//
	// **Sie können vor `Bereit` eintreffen** — das Apple-Event `odoc` kommt
	// beim Start, und ohne Kontext gibt es niemanden, dem man es melden könnte
	// (#891).
	//
	// **Mehrere, und das ist der Punkt.** Wails meldet jede Datei einzeln;
	// wer im Finder zwei Sets markiert und öffnet, löst zwei Aufrufe aus.
	wartendeDateien []string

	// startfehler hält, woran der letzte Startversuch scheiterte.
	//
	// **Damit das Fenster es zeigen kann** ([App.Startbefund]). Ohne ihn
	// beendete sich die Anwendung wortlos, und der Rückweg über eine
	// Sicherung war nicht erreichbar (#962).
	startfehler error

	// rueckspielrest hält, was ein abgebrochenes Zurückspielen beim Start
	// liegen ließ — für [App.Startbefund].
	rueckspielrest string

	// rueckweg hält die Sicherung, die nach einem gescheiterten Start
	// **wirklich** aufgeht — gesucht in `wegzurueck.go`, gezeigt vom Fenster.
	rueckweg *RueckwegDTO

	// dateiEntfernen ist die Naht zum Wegräumen einer Datei.
	//
	// **Als Feld, damit der halbe Weg prüfbar ist** (Befund 4 des Zweitblicks
	// über Phase 4.13): `PlatzFreigeben` löscht nacheinander, und was
	// geschieht, wenn das **dritte** scheitert, lässt sich mit Dateirechten
	// nicht herstellen — alle liegen im selben Verzeichnis, und ein
	// unschreibbares Verzeichnis lässt schon das erste scheitern.
	dateiEntfernen func(string) error

	// sprache ist die zuletzt gemeldete Sprache der Oberfläche.
	//
	// **Die nativen Teile brauchen sie**: Menüleiste und Dateidialog entstehen
	// in Go, und ohne diesen Wert stünden sie in der Sprache, mit der die
	// Anwendung gestartet ist.
	sprache string

	// paketquellen sind die Nähte zu dem, was in einem Paket steht —
	// Symbolzeichnungen, Setwerte, Schlüsselwörter, Packarten. Der Verbund
	// steht in `paketquellen.go`; sie beantworten dieselbe Frage, und
	// `anwendung.go` ginge mit ihnen über die Zeilengrenze aus §2.
	paketquellen

	// skriptkonsole wertet einen Ausdruck im Kontext einer Karte aus
	// (#1070).
	skriptkonsole Skriptkonsolenquelle

	// paketdateien liest und schreibt die Dateien installierter Pakete
	// (#1164).
	//
	// **Nicht `vorlagen`** — das ist die **Ausfuhr**vorlage weiter oben. Zwei
	// Dinge, die im Deutschen gleich heißen und nichts miteinander zu tun
	// haben: dort ein `.mse-export-template`, das ein Set nach HTML schreibt,
	// hier jede Textdatei eines installierten Pakets.
	paketdateien Paketdateiquelle

	// browseroeffner macht eine Adresse im Standardbrowser auf (T8, #1473).
	//
	// **Einsetzbar, damit ein Lauf ihn prüfen kann**: ein Test, der einen
	// Browser öffnet, ist keiner.
	browseroeffner service.Oeffner

	// vorlagenzeichner zeichnet eine Karte mit einer **ungespeicherten**
	// Vorlage (#1164).
	vorlagenzeichner Vorlagenzeichner

	// Die Nähte um die Sammlung — `anwendungsammlung.go`, eingebettet wegen §2.
	papiernaehte

	// uhr sagt, wie spät es ist — für das **Alter** eines Preises (#1165).
	//
	// **Sie ist wahlfrei**: ohne sie gilt `time.Now`. Ein Lauf, der ein Alter
	// prüft, setzt sie über [App.MitUhr]; alle anderen merken nichts davon.
	uhr func() time.Time

	// Die Nähte um Bilder — `bildnaehte.go`, eingebettet wegen §2.
	bildnaehte
}
