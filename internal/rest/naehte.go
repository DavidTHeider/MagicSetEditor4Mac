package rest

// **Die Nähte des Dienstes und ihre Gestalten** — abgetrennt am 2026-08-26,
// weil `dienst.go` über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Hier steht, **was** der Dienst von seiner Umgebung braucht und in welcher
// Gestalt er es weitergibt — Schnittstellen und DTOs. Daneben in `dienst.go`
// steht, **wie** er es tut: Marke, Drossel, Wachen.
//
// Die beiden wachsen aus verschiedenen Gründen: diese Datei mit jeder
// Fähigkeit, jene mit jeder Überlegung zur Sicherheit.

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Schreiber ist, was der schreibende Endpunkt braucht.
//
// **Mit Vorwert**, aus demselben Grund wie bei MCP: eine Änderung, die den
// vorigen Wert zurückgibt, lässt sich zurücknehmen. Lesen und Schreiben sind
// dabei **ein** Zug — sonst änderte ein zweiter Prozess dazwischen denselben
// Wert, und die Rücknahme führte auf einen Stand, den es nie gab.
type Schreiber interface {
	FeldSetzenMitVorwert(
		ctx context.Context, setID, karteID int64, feld, wert string,
	) (vorher string, hatte bool, fehler error)

	// **Und der Lebenszyklus** — seit
	// #1107.
	//
	// **Mit dem Set in jeder Signatur**, aus demselben Grund wie bei MCP
	// (#801):
	// eine verwechselte Kartenkennung löschte sonst eine Karte in einem Set,
	// das der Aufrufer nie genannt hat. Geprüft wird die Zugehörigkeit in
	// `service`, nicht hier
	// ([0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md)).
	KarteAnlegen(ctx context.Context, setID int64) (int64, error)
	KarteKopieren(ctx context.Context, setID, karteID int64) (int64, error)
	KarteLoeschen(ctx context.Context, setID, karteID int64, trotzdem bool) error
	KartenTauschen(ctx context.Context, setID, eineID, andereID int64) error

	// FelderSetzen setzt **ein** Feld auf **mehreren** Karten, in einer
	// Transaktion (#1390).
	//
	// **Alles oder nichts**: eine Karte, die nicht zu diesem Set gehört,
	// lässt den ganzen Aufruf scheitern, und keine ist danach geändert.
	//
	// **Sie gibt den Vorgang zurück** (#1393): mit ihm nimmt
	// `POST /sets/{setID}/zuruecknehmen` genau diese Änderung zurück.
	FelderSetzen(ctx context.Context, setID int64, karteIDs []int64, feld, wert string) (
		vorgang string, fehler error)

	// FelderVorschau sagt, was der Wert auf den Karten hieße — ohne zu
	// schreiben (Werkstück U6, #1417).
	FelderVorschau(ctx context.Context, setID int64, karteIDs []int64, feld, wert string) (
		domain.Massenvorschau, error)

	// FelderSetzenMitErwartung schreibt nur, wenn jede Karte noch auf dem
	// Altwert steht, den die Vorschau gezeigt hat.
	FelderSetzenMitErwartung(ctx context.Context, setID int64, erwartet []domain.Feldvorwert, feld, wert string) (
		vorgang string, fehler error)
}

// Naehte sind die Anwendungsfälle, die der Dienst bedient.
//
// **Als Verbund und nicht als fünfzehn Parameter**
// (#1107).
// Der Konstruktor war auf fünfzehn Stellen gewachsen, und eine
// Nahtsammlung war die einzige Aufzählung im Projekt, in der ein
// vertauschtes Paar noch übersetzte: `williger{}` erfüllt jede von ihnen.
// Ein benanntes Feld lässt sich nicht vertauschen.
type Naehte struct {
	Leser     Leser
	Schreiber Schreiber
	Ausfuhr   Ausfuhr
	Abgleich  Fremdabgleich
	Bilder    Fremdbilder
	Auskunft  Fremdauskuenfte

	// Suche ist die Kartensuche bei Scryfall (#1167).
	Suche Fremdkartensuche

	// Katalog ist die Auskunft über jede Fähigkeit der fremden Dienste —
	// Werkstück A1 von Phase 5.3 (#1440).
	Katalog Fremdkatalog

	// Peer ist die eigene Identität gegenüber anderen — Werkstück N1 von
	// Phase 5.4 (#1454).
	Peer Peernaht

	// Kontakte sind die bekannten Gegenüber — Werkstück N1 von Phase 5.4
	// (#1454).
	Kontakte Kontaktnaht

	// Sichtfreigaben sind die Sammlungssichten, die ein Profil einem
	// **Gegenüber** erteilt hat — Werkstück T3 von Phase 5.6
	// (#1468).
	//
	// **Nicht `Freigaben`** — den Namen trägt schon die Freigabe eines
	// **Profils** an einen zweiten Nutzer dieses Rechners (Phase 4.12). Zwei
	// Freigaben unter einem Namen wären die Verwechslung, die man später
	// nicht mehr findet.
	Sichtfreigaben Sichtfreigabenaht

	// Tausch ist der Abgleich gegen einen freigegebenen fremden Stand —
	// Werkstück T4 von Phase 5.6
	// (#1469).
	Tausch Tauschnaht

	// Post ist, was zu einem Gegenüber noch aussteht — Werkstück T5 von
	// Phase 5.6
	// (#1470).
	Post Postnaht

	// Verbuchung ist, was ein Handel im eigenen Bestand bewirkt — Werkstück
	// T6 von Phase 5.6
	// (#1471).
	Verbuchung Verbuchungsnaht

	// Leitung sind die Schritte vor dem Abschluss — Werkstück T1 von Phase
	// 5.6 (#1466), angeschlossen mit #1637.
	Leitung Leitungsnaht

	// Links sind die Wege nach draußen zu einer Karte — Werkstück T8 von
	// Phase 5.6
	// (#1473).
	Links Linknaht

	// Konto ist das eigene Konto bei einem Marktplatz — Werkstück T11 von
	// Phase 5.6 (#1476).
	//
	// **Nur lesend**: es gibt kein Verfahren, das dort schreibt.
	Konto Kontonaht

	// Deckeinfuhr liest eine Deckliste in ein Deck — Werkstück T8 von Phase
	// 5.6 (#1473).
	//
	// **Neben [Naehte.Sammlung] und nicht darin**: jene schreibt Besitz,
	// diese Plätze.
	Deckeinfuhr Deckeinfuhrnaht

	// Nachbarn sind die, die sich im selben Netz melden — Werkstück N2 von
	// Phase 5.4
	// (#1455).
	Nachbarn Nachbarnaht

	// Grenzen sagt, was dieser Prozess an Rufen und Parallelität zulässt (#1429).
	Grenzen Grenzenauskunft

	// Gesundheit sagt, ob dieser Prozess lebt und arbeiten kann (O9, #1433).
	Gesundheit Gesundheitsauskunft

	// Metriken nennt, was dieser Prozess gewirkt hat (O9, #1433).
	Metriken Metrikauskunft

	// Meta sagt, welches Programm antwortet (O9, #1433).
	Meta Metaauskunft

	// Vertraeglichkeit sagt, ob die neueste Freigabe diesen Bestand öffnen
	// kann (O6, #1430).
	Vertraeglichkeit Vertraeglichkeitsauskunft

	// Ruecksprung sagt, was zurückgehen kann und was eine Rückkehr kostet —
	// Werkstück O7 von Phase 5.2
	// (#1431).
	Ruecksprung Ruecksprungauskunft

	// Pflege sagt, was im Bestand herumliegt, und räumt es weg — Werkstück O8
	// von Phase 5.2
	// (#1432).
	Pflege Pflegeauskunft

	// Ruecksetzung setzt abgestuft zurück — Werkstück O4 von Phase 5.2
	// (#1428).
	Ruecksetzung Ruecksetzungsauskunft

	// Index ist der lokale Suchindex — Werkstück A15 von Phase 5.3
	// (#1514).
	Index Indexauskunft

	// Recherche fragt den Wissensbestand und stellt Kontextpakete zusammen —
	// Werkstück A16 derselben Phase
	// (#1515).
	Recherche Rechercheauskunft

	// Kosten führt das Journal externer Aufträge — Werkstück O12 von Phase
	// 5.2 (#1439).
	Kosten Kostenauskunft

	// Verweise bezieht entworfene Karten auf echte Drucke (#1167).
	Verweise Fremdverweise

	// Marken hängt Wörter an Karten
	// (#1208).
	Marken Fremdmarken

	// Ablage kopiert Karten zwischen Sets, Ersetzung sucht und ersetzt über
	// ein Set (#1168).
	Ablage    Fremdablage
	Ersetzung Fremdersetzung

	// Liste gibt die Spalten der Kartenliste und ihre Ordnung (#1168).
	Liste Fremdliste

	Sammeldaten Fremdsammeldaten
	Anlage      Fremdanlage
	Saetze      Fremdsaetze
	Konsole     Skriptkonsole

	// **Und die Paketwerkstatt** — seit
	// #1164.
	Werkstatt Paketwerkstatt

	// **Und die Sammlung samt Preisen** — seit
	// #1165.
	// Sie hängen an keinem Set, und deshalb steht `/sammlung` neben `/sets`.
	Sammlung Sammlungswerkstatt
	Preise   Preiswerkstatt

	// **Und die Listeneinfuhr** — seit Phase 4.5
	// (#1292).
	// Ohne sie braucht jedes Exemplar eine UUID von Hand.
	Listeneinfuhr Listeneinfuhrwerkstatt

	// Sammlungsausfuhr gibt die Sammlung aus und liest ein CSV zurück (#1292).
	Sammlungsausfuhr Sammlungsausfuhrwerkstatt

	// Sammlungssichten beantwortet Besitz, Lagerorte, Dubletten und Wünsche
	// (#1292, Vermerk 0057).
	Sammlungssichten Sammlungssichtenwerkstatt

	// Kartenskripte nennt und führt die `add cards script:`-Blöcke des
	// Spielpakets (#1337).
	Kartenskripte Kartenskriptwerkstatt

	// Auszaehlung zählt ein Set nach den `statistics dimension:`-Blöcken
	// seines Spielpakets aus (#1338).
	Auszaehlung Auszaehlwerkstatt

	// Parameterhilfen nennt beide Menüs zu einem Schlüsselwort (#1342).
	Parameterhilfen Parameterhilfenwerkstatt

	// Autoersetzungen zeigt und pflegt die `auto replace:`-Regeln eines
	// Spiels (#1341, Vermerk 0060).
	Autoersetzungen Autoersetzungswerkstatt

	// Decks bedient Decks und Formate (#1293, Phase 4.6), Turniere führt ein
	// Turnier über die Zeit (#1294, Phase 4.7).
	Decks    Deckwerkstatt
	Turniere Turnierwerkstatt

	// Symbolbibliothek führt die Symbole eines Profils (DOM-02, #1697),
	// Sprachpakete die eigenen Übersetzungen (LOC-02, #1688).
	Symbolbibliothek Symbolbibliothekswerkstatt
	Sprachpakete     Sprachwerkstatt

	// Symbol ist alles am Symbol eines Sets: nachzeichnen (#1340), lesen,
	// ablegen und die acht Rechnungen des Editors
	// (#1391).
	Symbol Symbolwerkstatt

	// Masse ändert und löscht über eine Auswahl und holt Preise in Schüben
	// (#1292).
	Masse Massenwerkstatt

	// **Und das Löschen eines Sets** — seit #1218.
	Loescher Setloescher

	// **Und die Sicherung des ganzen Bestands** — seit
	// #1169.
	// Sie hängt an keinem Set, und deshalb steht `/bestand` neben `/sets`.
	Sicherung    Bestandssicherung
	Dateien      Dateiansicht // was in einer Datei liegt — UX-13 (#1685)
	Objektmarken Objektmarken // Marken an jedem Fachobjekt — UX-12 (#1684)

	// **Und der portable Umschlag** — seit
	// #1457.
	Umschlag Umschlagwerkstatt

	// **Und die Partien** — seit Werkstück G1 von Phase 5.5
	// (#1458).
	Partien Partienwerkstatt

	// **Und der Verlauf einer Partie** — seit Werkstück G5 von Phase 5.5
	// (#1462).
	//
	// **Eine eigene Naht und kein fünftes Feld an [Partienwerkstatt]**: der
	// Verlauf beantwortet, was eine Partie **war**, jene, was sie **tut**.
	Partienverlauf Verlaufswerkstatt

	// **Und der Netzweg** — seit #1571.
	Netzweg        Netzweg
	Schnittstellen Schnittstellenauskunft // was Shellprofil und Codex kennen — OPS-09 (#1907)
	Schalter       Schalterauskunft       // was abgeschaltet ist — OPS-10 (#1908)

	// **Und die Regelfreigabe** — seit G4 (#1461).
	Regelfreigabe Regelfreigabe

	Wortwahl Wortlisten
	Woerter  Schluesselwoerter
	Pack     Pack
	Anleger  Setanleger

	// Auftraege führt die langen Läufe, um die jemand mit `Prefer:
	// respond-async` bat — FMT-03 (#1692). `nil` heißt: jeder Weg antwortet
	// gleich.
	Auftraege Auftragsnaht

	// Setvorschau sagt, was das Einfahren einer Datei täte — FMT-03.
	Setvorschau Setvorschau

	// Dateiablage nimmt übertragene Dateien an und hält Ausgaben zum
	// Abholen — FMT-03.AK-03. `nil` heißt: nur Pfade.
	Dateiablage Dateiablage

	// **Und die Setangaben, die Stile, die Pakete, die Kennzahlen, der
	// Verlauf und die Einstellungen** — seit
	// #1107.
	Angaben Setangaben
	Stile   Stilwahl
	Pakete  Paketkatalog

	// Probleme prüft die Karten eines Sets (PERF02, #2352).
	Probleme Probleme

	// Zusatz liest und setzt die Zusatzfelder einer Karte (#2298).
	Zusatz Zusatzwahl

	// Token führt die Zugriffstoken technischer Profile (#1351, Werkstück 10).
	Token Fremdtoken

	// Vorgaben führt die Profilvorgaben, die auf Zeilen zeigen (#1351,
	// Werkstück 15).
	Vorgaben Fremdvorgaben

	// Freigaben sagt, wer in welches fremde Profil darf (#1351, Werkstück 11).
	Freigaben Fremdfreigaben

	// Platz sagt, wie viel Platz was belegt — Werkstück 4 von Phase 4.13
	// (#1352).
	//
	// **Sie hängt an keiner Verbindung**: gezählt werden Dateien.
	Platz Fremdplatz

	// Deckuebernahme holt ein Deck aus einem fremden Profil (#1351,
	// Werkstück 12).
	Deckuebernahme Fremduebernahme

	// Nutzer führt die Nutzer dieses Bestands und ihre Profile (#1351).
	Nutzer Fremdnutzer

	// Kartenbilder setzt ein Bild von der Platte in ein Kartenfeld (#1291).
	Kartenbilder Kartenbilder

	// Probelauf rechnet den Druckbogen und die Exportvorlage durch, ohne zu
	// schreiben (UXR-F7, #2015).
	Probelauf Probelauf

	// Vorlagen führt die Exportvorlagen aus (#1291).
	Vorlagen      Vorlagenausgabe
	Zahlen        Kennzahlen
	Verlauf       Verlauf
	Einstellungen Einstellungen

	// Austausch führt die Einstellungen aus und ein — Werkstück O3 von
	// Phase 5.2
	// (#1427).
	Austausch Einstellungsaustausch

	// **Und die Aktualisierung** — seit
	// #1100.
	Fassungen Fassungen
}
