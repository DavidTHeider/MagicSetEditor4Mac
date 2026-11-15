package app

import (
	"path/filepath"
	"reflect"
	"testing"
)

// **Nach dem Start ist jede Naht gesetzt.**
//
// # Was ohne diesen Wächter geschieht
//
// Eine nicht gesetzte Naht ist ein `nil`-Zeiger, und der erste Aufruf darüber
// **stürzt ab**. Über die Wails-Bindung heißt das nicht „Fehler", sondern
// **nichts**: die Zusage wird nie erfüllt und nie abgelehnt, und die Seite
// steht für immer bei „Wird geladen …".
//
// Genau so ist es beim Bau der Setangaben passiert
// (#567): die
// Naht war beschrieben, die Bindung stand, jeder Test war grün — und im
// Fenster drehte sich das Rad. Dieselbe Klasse wie der fehlende Kontext in
// Phase 5, und derselbe Grund: kein Test kann es sehen, weil jeder seine Nähte
// selbst setzt.
//
// # Warum über Reflexion und nicht als Liste
//
// Eine Liste wäre eine zweite Aufzählung neben dem Verbund, und die vergisst
// beim nächsten Feld jemand. Reflexion zählt nach, was **dasteht**.
//
// # Warum er in eingebettete Verbünde hineingeht
//
// **Weil er sonst genau die verlöre, die man wegräumt.** Ein `struct` gilt
// hier als „gesetzt" — ein Nullwert ist bei `vorschaulauf` kein Mangel. Für
// einen **eingebetteten** Verbund heißt dieselbe Regel: alles darin ist
// unsichtbar, und wer vier Nähte hineinschiebt, um eine Datei unter 400 Zeilen
// zu bringen, nimmt ihnen den Wächter mit (§2, §10).
//
// Gefunden am 2026-09-06 beim Ausschneiden von [bestandsnaehte]. Dass er
// vorher schon zwei übersah — `leutenaehte` und `fremddienste` —, hat niemand
// bemerkt, denn ein Wächter, der weniger prüft, wird nicht rot.
//
// # Was er nicht prüft
//
// **Ob die Naht die richtige ist.** Er sieht, dass etwas gesetzt ist, nicht
// was. Ein `Kartenschreiber`, der versehentlich als Massenschreiber
// eingetragen wurde, kommt hier durch.
func TestNachDemStartIstJedeNahtGesetzt(t *testing.T) {
	anwendung := NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	// **Die Nähte, die das Fenster setzt**, stehen hier nicht: Menüleiste,
	// Dateidialog und Meldeweg brauchen ein echtes Fenster und werden in
	// `Bereit` gesetzt, nicht in `Starten`.
	ausserhalb := map[string]string{
		"fenster":     "kommt aus `Bereit`, nicht aus `Starten`",
		"menueSetzen": "braucht ein Fenster",
		"dateiwahl":   "braucht ein Fenster",
		"melden":      "braucht einen Wails-Kontext",
		"sprache":     "meldet die Oberfläche, nicht der Start",

		// **Keine Naht, sondern ein Befund.** `paketfehler` hält fest, warum
		// es kein Paketverzeichnis gibt — `nil` heißt: es gibt eins.
		"paketfehler": "ist ein aufgehobener Fehler, keine Naht",

		// **Dasselbe, und mit umgekehrtem Vorzeichen.** `startfehler` hält
		// fest, woran der Start scheiterte; nach einem geglückten ist er
		// `nil`, und genau das ist die gute Nachricht (#962).
		"startfehler": "ist ein aufgehobener Fehler, keine Naht",

		// **Und ein aufgehobener Satz.** `rueckspielrest` hält, was ein
		// abgebrochenes Zurückspielen beim Start liegen ließ; **leer ist die
		// gute Nachricht** (#1170, Befund 2 des zweiten Freigabeblicks).
		"rueckspielrest": "ist ein aufgehobener Satz, keine Naht",

		// **Und ein aufgehobener Rückweg** — Werkstück 5 von Phase 4.13
		// (#1352). `rueckweg` hält die Sicherung, die diese Fassung öffnen
		// kann; **`nil` heißt „es liegt keine da"** und ist selbst die
		// Auskunft. Gesucht wird er nur, wenn der Start scheitert — nach
		// einem geglückten steht hier richtigerweise nichts.
		"rueckweg": "ist ein aufgehobener Befund, keine Naht",

		// **Und eine zweite Verbindung, die es meistens nicht gibt** —
		// Werkstück 16 von Phase 4.12
		// ([0073](../../docs/entscheidungen/0073-ein-leser-darf-wirklich-lesen.md)).
		//
		// `nurLesen` steht **nur**, solange der Zeiger in einem fremden Profil
		// steht, in das jemand als `leser` gekommen ist. Nach einem
		// gewöhnlichen Start ist sie `nil`, und das ist der Normalfall — ein
		// Ruf darüber gibt es nicht, denn die Repositories hängen dann an
		// [App.db].
		//
		// **Der Lauf dafür steht in `nurlesen_test.go`**: er prüft, dass sie
		// entsteht, wenn sie soll — und wieder verschwindet.
		"nurLesen": "steht nur für einen Leser in einem fremden Profil",

		// **Die eingesetzte Wache** (N7, #1568): `nil` heißt hier „die
		// echte Autorisierung gilt", nicht „niemand fragt". Dieselbe
		// Gestalt wie bei `bund` darunter — [App.darf] fällt auf
		// [App.autorisierung] zurück, und die ist gesetzt.
		"wache": "nil heißt die echte Autorisierung, nicht keine",

		// **Und der eingesetzte Dienststarter** (N6, #1572): `nil` heißt
		// hier „das Programm neben diesem", nicht „keiner". Dieselbe
		// Gestalt wie bei `wache` darüber — [App.dienststarten] fällt auf
		// [starteDienstprogramm] zurück, und das ist der Betriebsfall.
		//
		// **Der Lauf dafür steht in `dienstposten_intern_test.go`**: er
		// setzt sie und prüft, dass der Posten darüber fährt.
		"dienststarter": "nil heißt das Programm neben diesem, nicht keines",

		// **Und aufgehobene Pfade.** `wartendeDateien` hält, was der Finder
		// vor dem Fenster schickte; leer heißt: es kam nichts (#891).
		//
		// **Mehrere seit dem 2026-08-25**: Wails ruft `OnFileOpen` je Datei,
		// und ein einzelner Pfad ließ alle außer dem letzten wortlos fallen
		// (Zweitblick über Phase Y, Befund 3).
		"wartendeDateien": "sind aufgehobene Pfade, keine Naht",

		// **Und ein gemerkter Text.** `letzteDiagnose` hält, was
		// [App.Diagnose] zuletzt gezeigt hat, damit [App.DiagnoseSichern]
		// **genau diesen** Text sichert (#1211, Befund 5). Leer heißt: es
		// wurde noch nichts gezeigt — und dann gibt es nichts zu sichern.
		"letzteDiagnose": "ist ein gemerkter Text, keine Naht",

		// **Und drei gemerkte Stücke desselben Schlags** (O10, #1434): die
		// Vorschau und die Teile des zuletzt gezeigten Pakets und der Ort des
		// zuletzt geschriebenen. Alle sind vor der ersten Handlung leer — das ist
		// die Zusage „ohne Vorschau wird nicht geschrieben", nicht ein
		// vergessener Wurzelknoten.
		"letzteVorschau":  "ist die letzte Vorschau, keine Naht",
		"letzteTeile":     "sind die Teile der letzten Vorschau, keine Naht",
		"letzterPaketort": "ist der Ort des letzten Pakets, keine Naht",

		// **`bund` ist `nil`, und das ist seine Einstellung** — dann gilt der
		// Anmeldebund von macOS. Ihn beim Start zu setzen hieße, dass jeder
		// Lauf, der die Anwendung startet, in den **echten** Schlüsselbund
		// dieses Rechners schriebe (§3); der Grund steht bei
		// [fremddienste.bund].
		//
		// **Dieselbe Gestalt wie [App.sicherungspruefer]**: nil heißt dort
		// „die echte Prüfung", hier „der echte Bund" — nicht „keiner".
		"bund": "nil heißt der Anmeldebund von macOS, nicht keiner",

		// **Und derselbe Grund beim Marktzugang** (A4, #1443): nil heißt
		// [App.bundFuerDenMarkt] nimmt den echten.
		"marktbund": "nil heißt der Anmeldebund von macOS, nicht keiner",

		// **Und der Bund der eigenen Identität** (N1, #1454) — aus
		// demselben Grund: `bundFuerDenPeer` nimmt den echten, wenn keiner
		// eingesetzt ist. Gesetzt wird er nur in Läufen, damit keiner in
		// den Anmeldebund dieses Rechners schreibt.
		"peerbund": "nil heißt der Anmeldebund von macOS, nicht keiner",

		// **Der Wähler** (N1, #1454): nil heißt, dass es noch keinen
		// Transport gibt — er kommt mit N2 (libp2p) und N3 (Pion). Die
		// **Verwaltung** daneben steht seit `NeueApp` und wird geprüft.
		"waehler": "nil heißt: es gibt noch keinen Transport (N2, N3)",

		// **Die Kommandozentrale** (U12, #1423): die Quelle setzt der
		// Wurzelknoten, die Befehle kommen aus ihr, und der Abbruch steht nur,
		// solange ein Befehl läuft.
		"befehlsquelle":   "setzt der Wurzelknoten des Fensters — Tests haben keinen",
		"befehle":         "kommen aus der Quelle, und ohne Quelle ist die Zentrale leer und sagt es",
		"befehlAbbrechen": "steht nur, solange ein Befehl läuft",
		"skriptAbbrueche": "steht je Karte nur, solange ein Ausdruck läuft (#1424)",

		// **Die Adresse des Bulkbezugs** (A3, #1442): leer heißt die Vorgabe
		// des Pakets — `https://mtgjson.com/api/v5`. Sie steht nur, wo ein
		// Lauf gegen eine eigene Gegenstelle geht.
		"bulkadresse": "leer heißt die Vorgabe von internal/mtgjsonpreise",

		// **Und der Modellkatalog ist eine Aufzählung, keine Naht** (A6,
		// #1445). Er ist heute **leer**, weil kein Anbieter angebunden ist;
		// A7 (#1446) bringt den ersten. Eine leere Liste ist die richtige
		// Auskunft — „es steht kein Modell fest" —, und ein Ruf darüber
		// stürzt nicht ab, sondern gibt sie.
		"kimodelle": "ist die Aufzählung der feststehenden Modelle — heute keines",

		// **Und das Zuhause der Serverdatei ist leer im Betrieb** (A7,
		// #1446): leer heißt das echte Zuhause dieses Nutzers. Es steht nur,
		// wo ein Lauf nicht den Stand dieses Rechners lesen darf (§3).
		"kizuhause": "leer heißt das echte Zuhause, nicht keines",
		"geraet":    "nil heißt die Naht des Prozesses — eine je Prozess, nicht je Anwendung (AI-05)",

		// **Und die Freigabe für einen kostenpflichtigen Auftrag steht nur,
		// solange einer läuft** (A8, #1447). Sie gehört zu **einem** Ruf und
		// nicht zu einer Sitzung; nach dem Start gibt es keinen.
		"kinachweis": "steht nur für einen kostenpflichtigen Auftrag",
		"kigrenze":   "steht nur neben dem Nachweis, zu dem sie gehört",

		// **Und die zwei KI-Zwischenstände sind keine Nähte** (A7, #1446):
		// das Auftragswerk entsteht beim ersten Vorschlag, der Nahtfehler
		// steht nur, wo einer aufgetreten ist.
		"kiwerk":       "entsteht beim ersten Vorschlag, nicht beim Start",
		"kinahtfehler": "ist ein aufgehobener Fehler, keine Naht",

		// **Und die Protokollweiche setzt `main`** (OPS-04, #1721): ohne sie
		// bleibt das Protokoll, wie `main` es gebaut hat, und die Diagnose
		// nennt keine Kanäle — ein Aufruf darüber stürzt nicht ab.
		"protokollweiche": "setzt main — ohne sie bleibt das Protokoll, wie es gebaut wurde",

		// **Die Liste der jüngsten Sets ist ein gemerkter Stand, keine Naht**
		// (UXR-F14, #2029): sie wird gefüllt, wenn die Leiste das erste Mal mit
		// dem Bestand entsteht, und ohne Sets bleibt sie leer — und leer ist
		// die richtige Auskunft (`menuenaht.go`).
		"zuletztliste": "ist die zuletzt gelesene Liste der jüngsten Sets, keine Naht",

		// **Und die Liste der Schnittstellen liest im Betrieb den Rechner**
		// (G08, #2272): nil heißt `entdeckung.Schnittstellenliste`, und nur
		// ein Lauf setzt eine eigene, damit er nicht die Netze dieses Macs
		// zeigt (`netzschnittstellen.go`).
		"netzliste": "nil heißt die Schnittstellen dieses Rechners, nicht keine",
	}

	naehtePruefen(t, reflect.ValueOf(anwendung).Elem(), ausserhalb)
}

// naehtePruefen geht einen Verbund durch und steigt in eingebettete hinab.
func naehtePruefen(t *testing.T, wert reflect.Value, ausserhalb map[string]string) {
	t.Helper()

	typ := wert.Type()

	for i := range typ.NumField() {
		feld := typ.Field(i)

		if feld.Anonymous && feld.Type.Kind() == reflect.Struct {
			naehtePruefen(t, wert.Field(i), ausserhalb)

			continue
		}

		if grund, egal := ausserhalb[feld.Name]; egal {
			t.Logf("%s: übergangen — %s", feld.Name, grund)

			continue
		}

		if !istGesetzt(wert.Field(i)) {
			t.Errorf("die Naht %q ist nach dem Start nicht gesetzt — "+
				"ein Aufruf darüber stürzt ab, und das Fenster wartet für immer",
				feld.Name)
		}
	}
}

// istGesetzt sagt, ob ein Feld mehr als seinen Nullwert trägt.
func istGesetzt(feld reflect.Value) bool {
	switch feld.Kind() {
	case reflect.Interface, reflect.Pointer, reflect.Func, reflect.Map, reflect.Slice:
		return !feld.IsNil()

	case reflect.String:
		return feld.String() != ""

	// **Alles andere gilt als gesetzt.** Ein `struct` wie `vorschaulauf` ist
	// von Anfang an brauchbar; ein Nullwert ist dort kein Mangel.
	default:
		return true
	}
}
