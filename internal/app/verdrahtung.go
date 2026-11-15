// **Die Verdrahtung des Fensters** — aus `start.go` herausgelöst am
// 2026-09-06, Werkstück 16 von Phase 4.12
// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)).
//
// # Warum sie eine eigene Funktion ist
//
// **Weil sie zweimal laufen können muss.** Wer als [domain.RolleLeser] in
// einem fremden Profil steht, arbeitet über eine **zweite**, nur lesende
// Verbindung — und die Repositories hängen an der Verbindung, über der sie
// gebaut wurden. Ein Umhängen zur Laufzeit gibt es nicht; was es gibt, ist
// ein zweiter Bau.
//
// **Der Schnitt folgt der Frage** (§2): `start.go` sagt, **wann** was
// geschieht — Rest eines Zurückspielens, öffnen, migrieren, auffrischen;
// hier steht, **woraus** die Anwendung besteht.

package app

import (
	"database/sql"
	"log/slog"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// verdrahten baut jede Naht über einer Verbindung.
//
// **Ohne die Nutzerverwaltung** — die steht in `starten` und bleibt an der
// **schreibenden** Verbindung. Der Grund steht bei
// [App.nurLesendVerdrahtenWennNoetig]: wer nur lesen darf, muss trotzdem
// wieder hinausfinden, und das Hinausfinden schreibt den Zeiger.
func (a *App) verdrahten(db *sql.DB) {
	leser := storage.NeuerSetLeser(db)
	a.bestand = leser
	a.benutzung = storage.NeuerSetSpeicher(db)
	// **Die Übersicht** — Werkstück U11 von Phase 5.1 (#1422). Über die
	// Anwendung, weil ihre Nähte erst in späteren Startschritten stehen.
	a.MitWeiterarbeiten(a.weiterarbeitenUeber(storage.NeuerArbeitsverlauf(db)))

	// **Die Kommandozentrale** — Werkstück U12 von Phase 5.1 (#1423): die
	// Befehle kommen aus der Quelle, die der Wurzelknoten gesetzt hat.
	a.befehleAus(db)

	// Das Auftragsbuch steht vom Start an — der Wächter über die Nähte fragt
	// nach jedem Feld, und ein Buch, das erst der erste Ruf anlegt, sähe für
	// ihn aus wie eine vergessene Naht. Der Riegel in `auftragsbuch()` bleibt
	// für eine App ohne `Starten`, wie Tests sie bauen.
	a.auftraege = service.NeuesAuftragsbuch(a.uhr)
	a.zeilen = leser
	a.einzelkarte = leser
	a.feldwerte = leser
	a.kartenleser = leser
	a.sucher = leser
	a.statistiker = leser
	// **Mit der Kopfquelle**, damit die Regel „eine Quelle, ein Set" aus dem
	// Fachkern kommt und nicht aus dieser Fassade (#886).
	a.importeur = service.NeuerImporteur(storage.NeuerSetSpeicher(db)).MitKoepfen(leser)
	// **Die eigenen Dateien der Anwendung** stehen in `ablagennaht.go`.
	ablage := a.ablagen()

	// **Mit Setwache** (#1336).
	a.verlauf = service.NeuerVerlaufsdienst(
		storage.NeuerVerlaufsleser(db)).MitSetwache(storage.NeuerSetLeser(db))
	a.stand = storage.NeuerStandleser(db)

	// **Ein Paketverzeichnis, das nicht da ist, hält den Start nicht auf.**
	// Wer nie ein Paket bezieht, braucht keines; und ein Start, der an einem
	// Verzeichnis scheitert, das erst Phase 7 füllt, wäre eine Zusage, die
	// niemand angefordert hat. Der Fehler kommt beim ersten Bezug.
	//
	// **Der Fehler wird aber aufgehoben, nicht verworfen.** Der erste Anlauf
	// warf ihn weg und baute den Bezieher mit leerer Wurzel — daraus wurde ein
	// relativer Pfad, den ein späterer Bezug geräumt hätte. Der Zweitblick über
	// Phase 9 hat es gefunden; `Paketbezieher` weist eine solche Wurzel jetzt
	// selbst ab, und hier steht der Grund, warum sie leer sein kann.
	pakete, fehler := a.paketort()
	if fehler != nil {
		a.paketfehler = fehler

		// **Und er steht ab jetzt auch im Protokoll** (#857). Gemeldet wird er
		// erst beim ersten Bezug — bis dahin weiß niemand, dass die Anwendung
		// mit einer leeren Paketwurzel läuft.
		slog.Default().Error("das Paketverzeichnis ließ sich nicht bestimmen",
			"grund", fehler.Error())
	}

	bezieher := service.NeuerPaketbezieher(pakete, Kenner())
	a.paketbezieher = bezieher

	// **Die Pakete liegen nicht unmittelbar unter dem Bezugsverzeichnis.**
	// Bezogen wird ein ganzes Repositorium, und `M15-Magic-Pack` legt seine 190
	// Stile unter `M15-Magic-Pack/data/` ab. `Bildquelle` und `Paketleser`
	// fügen Wurzel und Paketnamen zusammen — bis #206 fanden sie deshalb
	// **nichts**, und niemand sah es: die Anwendung zeigte ohnehin den
	// Behelfsplan.
	//
	// **Einmal beim Start**, nicht bei jedem Zugriff (siehe
	// [msepaket.Paketwurzeln]).
	//
	// **Die Saat kommt hinten dran** (#58),
	// und die Reihenfolge ist die Aussage: wer ein Paket bezogen hat, meint das
	// bezogene. Bis hierher stand die Saat im Repository und nicht im Bündel —
	// wer die gebaute Anwendung startete, hatte kein Spiel, keinen Stil und
	// keine Karte. Ein leerer Pfad fällt in `Paketwurzeln` weg; das ist der
	// Fall „ohne Bündel gestartet", und er ist benannt statt behauptet.
	// **Der verwaltete Ort gilt immer — auch leer.**
	//
	// `Paketwurzeln` meldet, **wo Pakete liegen**; ein frisch angelegtes,
	// leeres Paketverzeichnis trägt keines und fällt heraus. Wer danach ein
	// Paket installiert, legt es an einen Ort, den weder Bildquelle noch
	// Stilspeicher kennen — sichtbar erst beim nächsten Start. Der Zweitblick
	// über Phase 7 hat es angestrichen.
	//
	// **Behoben wird es hier und nicht in `Paketwurzeln`.** Der erste Anlauf
	// ließ die genannte Wurzel dort immer mitzählen — und machte zwei
	// bestehende Tests rot, die sagen, was die Funktion bedeutet: „nur, was
	// Pakete trägt". Die ältere Regel ist die bessere (§10); was hier gilt, ist
	// eine Aussage des Composition Roots über **seinen** Ort, nicht über die
	// Bedeutung einer Suche.
	// **Und die mitgelieferten Vorlagen zwischen bezogenen und Saat**
	// (#702):
	// wer ein Paket bezogen hat, meint das bezogene; wer nichts bezogen hat,
	// bekommt das mitgelieferte statt der Probevorlagen aus der Saat.
	wurzeln := mitVerwaltetemOrt(pakete,
		msepaket.Paketwurzeln(pakete, a.vorlagenort(), a.saatort()))

	// **Der Katalog der installierten Pakete** (#54). Er liest dieselben
	// Wurzeln, die auch Stil, Bild und Schrift benutzen — eine zweite Liste
	// wäre eine zweite Wahrheit darüber, was installiert ist.
	//
	// **Verwaltet wird nur das Bezugsverzeichnis.** Die Saat liegt im Bündel;
	// ein Entfernen dort löschte einen Teil der Anwendung (#58). Der Katalog
	// **zeigt** sie trotzdem — er soll ja sagen, was da ist.
	katalog := storage.NeuerPaketkatalog(db)
	verwalter := service.NeuerPaketverwalter(katalog, pakete, wurzeln...).
		MitSetkoepfen(storage.NeuerSetLeser(db))
	a.pakete = verwalter

	// **Und die Stände der bezogenen Bäume** (G04, #2251) — über denselben
	// Verwalter und denselben Bezieher, damit Katalog und Platte einer sind.
	a.paketbezuege = service.NeuePaketaktualisierung(verwalter, bezieher, bezieher)

	// **Wer hier arbeitet und wie er hereinkommt** — Phase 4.12 (#1351).
	// Der Aufbau steht in `leutestart.go`; er ist der einzige, der sechs
	// Nähte auf einmal setzt.
	// **Die Leute bleiben an der schreibenden Verbindung** — Werkstück 16 von
	// Phase 4.12
	// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)).
	//
	// **Weil man aus einem nur lesenden Profil wieder hinausfinden muss**, und
	// das Hinausfinden schreibt den Zeiger. Hinge die Nutzerverwaltung an der
	// nur lesenden Verbindung, säße ein `leser` fest — in einem fremden
	// Profil, ohne Weg zurück.
	//
	// **Das ist die eine benannte Ausnahme**, und sie ist eng: ein Zeiger in
	// der Tafel `bestand`. Alles, was zum **Inhalt** eines Profils gehört —
	// Sets, Karten, Decks, Sammlung, Turniere, Einstellungen — hängt an der
	// Verbindung, die dieser Ruf bekommt.
	a.leuteStarten(a.db, db)

	// **Nach den Leuten, weil der Riegel von ihnen kommt** — Werkstück O1
	// von Phase 5.2 (#1425): jeder Weg, der eine Datei schreibt, fragt die
	// Autorisierung, und die entsteht in `leuteStarten`. `query_only` allein
	// hielte einen Leser hier nicht auf — die Platte fragt den Bestand nicht.
	a.exporteur = service.NeuerExporteur(leser).MitPruefer(a.autorisierung)
	a.speicherer = service.NeuerSetschreiber(leser, leser).
		MitBildablage(ablage).MitPruefer(a.autorisierung).
		MitStilaufloesung(service.StilaufloesungUeber(wurzeln...))

	// **Die Einstellungen liegen im Bestand** (#245): Sprache,
	// Erscheinungsbild, Schriftgröße. Bis dahin lag die Sprache im
	// Browserspeicher — ein Ort, den der MCP-Server nicht hat und den keine
	// Sicherung mitnimmt.
	// **Mit der Naht zum Menschen** — Werkstück 9 (#1351), Punkt 9: steht bei
	// einem Profil „wie der Nutzer", gilt, was beim Menschen steht.
	a.einstellungen = einstellungsverwalterUeber(db)

	a.kiVerdrahten()

	// **Und der Austausch darüber** — Werkstück O3 von Phase 5.2
	// (#1427).
	// **Nach `leuteStarten`**, wie der Exporteur darüber: die Übernahme ist
	// `lokalSchreiben` und fragt die Autorisierung, bevor sie schreibt.
	a.austausch = service.NeuerEinstellungsaustausch(
		storage.NeueEinstellungen(db), storage.NeueNutzerablage(db),
		storage.NeueEinstellungen(db), a.autorisierung)

	// **Die Optionen einer Exportvorlage werden je Vorlage behalten** (#492),
	// wie MSE sie hält.
	a.exportoptionen = storage.NeueExportoptionen(db)

	// **Die Stilauswahl über denselben Katalog** (#41): welche Stile es gibt,
	// weiß der Paketkatalog — eine zweite Liste wäre eine zweite Wahrheit.
	stilwaehler := service.NeuerStilwaehler(
		leser, katalog, storage.NeueStilablage(db), leser, wurzeln...)
	a.stile = stilwaehler

	// **Und die Auto-Replace-Regeln über denselben Wähler** (#897): sie kommen
	// aus dem Spielpaket, das er ohnehin öffnet.
	a.autoersetzung = stilwaehler

	schreiber := storage.NeuerKartenschreiber(db)

	// **Und dazwischen steht der Feldweg** (#696): ein Feld mit
	// `save value: false` gehört nicht in den Bestand, sondern in die Felder,
	// aus denen es entsteht. Wer hier den rohen Schreiber einsetzte, legte
	// `text:` ab — und die Vorlage läse es nie.
	//
	// **Nur an dieser Naht.** Der Massenschreiber, der Anleger und die
	// Setangaben gehen weiter unmittelbar an den Bestand; sie schreiben keine
	// Feldwerte, die eine Vorlage ausrechnet.
	//
	// **Und um beides herum steht die Kartenaufsicht** — seit
	// #808.
	// Sie hält die Karte gegen das Set, das der Aufrufer nennt
	// ([0041](../../docs/entscheidungen/0041-die-setgrenze-liegt-im-dienst.md)).
	//
	// **Auch für das Fenster, und nicht nur für MCP.** Der naheliegende
	// Einwand — die Kennung komme aus der Liste des offenen Sets, also könne
	// sie nicht fremd sein — widerlegt sich an §7 Regel 1: die steht in der
	// Leitlinie, **weil** eine stehengebliebene Kennung nach einem Setwechsel
	// vorgekommen ist. Genau die zeigt auf eine Karte des vorigen Sets.
	//
	// Der Zweitblick über Phase N hat es angestrichen: eine Regel, die für
	// jede Kartenänderung zugesagt und für vier von sieben Wegen gehalten
	// wird, ist keine Regel.
	// **Der Feldweg rechnet, was gespeichert wird** (#2396) — den Nachzug
	// bekommt er unten, denn er braucht das Zeichenwerk.
	feldweg := service.NeuerFeldweg(leser, stilwaehler, schreiber)

	aufsicht := service.NeueKartenaufsicht(leser, schreiber).
		MitFeldweg(feldnaht{Kartenschreiber: schreiber, weg: feldweg}, schreiber).
		// **Damit eine Karte, die in einem Deck liegt, nicht unbemerkt
		// verschwindet** (#1369).
		MitDeckwache(storage.NeueDeckablage(db))

	a.schreiber = aufsicht

	// **Die Aufsicht bekommt die Vorlage** (#1211, Befund 1). Ohne sie reichte
	// die Massenänderung ein nicht editierbares oder berechnetes Feld
	// ungeprüft bis in den Speicher durch.
	a.massenschreiber = aufsicht.MitFeldformen(stilwaehler)

	// **Der Zugang zu Scryfall steht in `scryfallstart.go`** (#1167): er war
	// der längste zusammenhängende Abschnitt dieser Datei, und mit der
	// Kartensuche ging sie über die Zeilengrenze (§2). Der Schnitt folgt einer
	// Frage: hier steht, **was** beim Start entsteht, dort, wie der fremde
	// Dienst verdrahtet wird.
	a.scryfallStarten(db, aufsicht, leser)

	// **Decks, Formate und ihre beiden Auskünfte** — Phase 4.6
	// (#1293).
	//
	// **Über dieselben Ablagen wie alles andere.** Ein zweiter Deckspeicher
	// wären zwei Wahrheiten über dieselbe Tabelle.
	a.deckwegeStarten(db, leser)

	// **Marken an Karten**
	// (#1208).
	//
	// **Über denselben Leser wie alles andere**: dass eine Karte zu dem
	// genannten Set gehört, entscheidet der Bestand und nicht der Aufrufer.
	a.marken = service.NeueMarkenpflege(storage.NeueMarken(db), leser, leser)

	// **Und an jedem anderen Fachobjekt** (UX-12, #1684) — über dieselbe
	// Naht, die Kommandozeile, Dienst und Werkzeugserver bekommen.
	a.objektmarken = ObjektmarkenUeber(db)

	// **Karten kopieren, ausschneiden und einfügen** (#1168). Der Löscher ist
	// die **Aufsicht** und nicht der Speicher: sie führt den Verlauf, und ein
	// ausgeschnittenes Set soll sich zurücknehmen lassen.
	//
	// **Und mit der Bildablage** — ohne sie reisen Bildfelder ohne ihre Datei
	// (#1257).
	// Sie hat hier **gefehlt**: die Fähigkeit war gebaut, geprüft und über
	// diesen Weg nicht erreichbar; `docs/PARITAET.md` sagte trotzdem, das
	// Fenster trage `image_data`. Ein Paket ohne Aufrufer ist keine
	// Anwendungsfunktion (§10, erste Dimension) — angestrichen vom Nachlauf
	// über die Nacharbeit
	// (#1282,
	// Befund 1).
	a.ablage = service.NeueKartenablage(leser, leser, aufsicht).
		MitLoescher(aufsicht).
		MitBildern(ablage).
		MitFeldern(stilwaehler)

	// **Suchen und Ersetzen über ein Set** (#1168). Der Schreiber ist der
	// **Kartenschreiber** und nicht die Aufsicht: `WerteSetzen` legt alle
	// Werte selbst in eine Verlaufsgruppe, und ⌘Z nimmt sie in einem Zug
	// zurück.
	a.ersetzung = service.NeuerErsetzer(leser, stilwaehler, schreiber)

	// **Die Spalten der Kartenliste und ihre Ordnung** (#1168). Die Wahl
	// gehört dem **Spiel** und nicht dem Set — deshalb liegt sie in einer
	// eigenen Ablage und nicht bei den Setangaben.
	a.listenbau = service.NeuerKartenlistenbau(
		leser, leser, stilwaehler, storage.NeueListenablage(db), stilwaehler).
		// **Und die Marken je Karte** (#1208) — über dieselbe Ablage wie
		// [App.Marken]: zwei Ablagen über derselben Tabelle wären zwei
		// Wahrheiten über dieselbe Marke.
		MitMarken(storage.NeueMarken(db))

	// **Auch das Anlegen geht über die Aufsicht** (#2516): sie rechnet die
	// Werte der neuen Karte. Bis dahin legte hier der Bestand selbst an, und
	// eine Karte aus dem Fenster trug anderes als eine von der Kommandozeile.
	a.anleger = aufsicht
	a.loescher = aufsicht
	a.kopierer = aufsicht
	a.tauscher = aufsicht

	// **Und schreibt die Setangaben** (#567) — dieselbe Begründung, und gelesen
	// wird aus demselben Leser wie alles andere.
	a.setwerte = setwertnaht{leser: leser, schreiber: schreiber}

	// **Und die Symbolbibliothek** (DOM-02, #1697) — über dieselbe Ablage wie
	// die gezeichneten Symbole, denn eine eingesetzte Kopie ist eines, und
	// über dieselben Setangaben.
	a.symbolbibliothek = service.NeueSymbolbibliothek(storage.NeueSymbolbibliothek(db),
		service.NeueSymbolablage(ablage), setwertnaht{leser: leser, schreiber: schreiber}, leser).
		MitFassungen(stilwaehler).MitPruefer(a.autorisierung)

	// **Und die eigenen Übersetzungen** (LOC-02, #1688) — geprüft gegen den
	// deutschen Katalog, den das Programm mitbringt.
	a.sprachpakete = NeueSprachpaketeUeber(db).MitPruefer(a.autorisierung)

	// **Und der Bestand kann sich selbst wegtragen** (#799): `storage.Sichern`
	// schreibt über `VACUUM INTO` eine vollständige Datei, während gearbeitet
	// wird. Bis hierher konnte sie niemand aufrufen.
	a.bestandssicherer = bestandsnaht{db: db}

	// **Und er kann sich selbst verschlüsseln** (Werkstück 14 von Phase
	// 4.12). Dieselbe Stelle und derselbe Grund: die Naht hält die
	// **laufende** Verbindung, und nach jedem Tausch ist es eine andere.
	a.bestandschiffre = chiffrenaht{holen: a.laufendeVerbindung}

	// **Und er weiß, was eine Rückkehr kostet** (O7, #1431). Dieselbe Stelle
	// und derselbe Grund: beide Leser und das Buch hängen an der **laufenden**
	// Verbindung, und nach einem Zurückspielen ist es eine andere — das Buch
	// liegt dann im zurückgespielten Bestand, und genau dorthin gehört die
	// Epoche.
	a.rueckkehrnaht = rueckkehrnaht{holen: a.laufendeVerbindung}

	// **Und die Fassade fragt darüber** (O7, #1431). Sie geht über
	// `bestandsweg()`, weil der bei jedem Ruf neu entsteht — ein festgehaltener
	// Weg trüge nach einem Zurückspielen die alte Verbindung.
	a.MitRuecksprung(ruecksprungnaht{weg: a.bestandsweg})

	// **Und die Datenpflege** (O8, #1432). Dieselbe Stelle und derselbe
	// Grund: sie liest, sichert und schreibt über die **laufende**
	// Verbindung, und nach einem Zurückspielen ist es eine andere.
	a.MitPflege(PflegeUeber(a.laufendeVerbindung, a.datenbankpfad))

	// **Und die abgestufte Rücksetzung** (O4, #1428) und **das Kostenjournal**
	// (O12, #1439) — die Gründe stehen bei den beiden Nähten.
	a.ruecksetzungVerdrahten()
	a.kostenVerdrahten()

	// **Und der lokale Suchindex** (A15, #1514) — aus demselben Grund über
	// die laufende Verbindung: ein Zurückspielen tauscht den Bestand, und
	// der Index liegt darin.
	a.retrievalVerdrahten()

	// **Und die Recherche darüber** (A16, #1515): sie hängt an derselben
	// Verwaltung — A16 baut keinen zweiten Retrievalkern.
	a.rechercheVerdrahten()

	// **Und die Bilderkennung** (A13.3, #1452) — sie hängt an derselben
	// Drucksuche wie eine getippte Kartenliste, damit ein Foto keinen Druck
	// nennen kann, den die Suche nicht auch gefunden hätte.
	a.bilderkennungVerdrahten()

	// **Und das Zeichenwerk** — es steht in `verdrahtungzeichnen.go`, weil
	// diese Datei sonst über die Zeilengrenze aus §2 ginge. Der Schnitt
	// folgt der Frage: hier hängt, **was der Bestand kann**, dort, **was
	// eine Karte zu einem Bild macht**.
	werk := a.zeichenwerkVerdrahten(zeichenteile{
		db: db, leser: leser, wurzeln: wurzeln, ablage: ablage,
		stilwaehler: stilwaehler,
	})

	// **Gespeichert wird, was gerechnet ist** (#2396) — über **dieselben**
	// Stile, Schriften und Bilder wie die Vorschau: zwei Umgebungen wären zwei
	// Antworten darauf, was ein Feld rechnet.
	feldweg.MitNachzug(werk.Stile, werk.Schriften, werk.Bildquelle)
}
