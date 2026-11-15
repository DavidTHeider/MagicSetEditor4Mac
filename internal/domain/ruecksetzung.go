package domain

// **Zurücksetzen benennt genau, was betroffen ist** — Werkstück O4 von Phase
// 5.2 (#1428).
//
// # Warum es sechs Arten sind und nicht einen Knopf
//
// „Zurücksetzen" ist kein Vorgang, sondern eine Familie: das Fenster wieder
// aufräumen und einen Nutzer samt allem löschen haben nichts miteinander zu
// tun außer dem Wort. Ein Knopf für beides ist die Gestalt, in der jemand das
// eine drückt und das andere bekommt.
//
// **Die Reihenfolge ist die Schwere.** Was oben steht, nimmt weniger weg.

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
)

// Ruecksetzungsart ist eine der sechs.
type Ruecksetzungsart string

const (
	// RuecksetzungLayout stellt das Aussehen des Fensters zurück — **und
	// sonst nichts**. Die Abnahme sagt es: „Layoutreset löscht keine Karten."
	RuecksetzungLayout Ruecksetzungsart = "layout"

	// RuecksetzungProfileinstellungen nimmt die Wahl des Profils zurück, so
	// dass wieder **geerbt** wird.
	RuecksetzungProfileinstellungen Ruecksetzungsart = "profileinstellungen"

	// RuecksetzungNutzereinstellungen nimmt die Wahl des Menschen zurück.
	RuecksetzungNutzereinstellungen Ruecksetzungsart = "nutzereinstellungen"

	// RuecksetzungProfil löscht ein Profil samt dem, was ihm gehört.
	RuecksetzungProfil Ruecksetzungsart = "profil"

	// RuecksetzungNutzer löscht einen Menschen samt seinen Profilen.
	RuecksetzungNutzer Ruecksetzungsart = "nutzer"

	// RuecksetzungAnwendung setzt alles zurück — und verlangt dafür einen
	// [Nachweis].
	RuecksetzungAnwendung Ruecksetzungsart = "anwendung"
)

// Ruecksetzungsarten gibt die Reihe, von der leichtesten zur schwersten.
func Ruecksetzungsarten() []Ruecksetzungsart {
	return []Ruecksetzungsart{
		RuecksetzungLayout, RuecksetzungProfileinstellungen,
		RuecksetzungNutzereinstellungen, RuecksetzungProfil,
		RuecksetzungNutzer, RuecksetzungAnwendung,
	}
}

// IstRuecksetzungsart sagt, ob eine Marke eine der sechs ist.
func IstRuecksetzungsart(art string) bool {
	for _, bekannt := range Ruecksetzungsarten() {
		if string(bekannt) == art {
			return true
		}
	}

	return false
}

// RuecksetzungsZweck ist der Zweck des Nachweises für die ganze Anwendung.
const RuecksetzungsZweck = "ruecksetzung.anwendung"

// BrauchtNachweis sagt, ob diese Art einen [Nachweis] verlangt.
//
// # Warum nur die letzte
//
// **Weil nur sie über fremdes Eigentum geht.** Ein Profilverwalter darf sein
// Profil zurücksetzen und seinen Nutzer löschen; die ganze Anwendung
// zurückzusetzen nähme **anderen Eigentümern** ihre Daten weg, und dafür
// reicht keine Rolle. Das Issue verlangt es so: „Kein Profilverwalter erhält
// durch umfassenden Reset Rechte über andere Eigentümer; Gesamtreset verlangt
// passende lokale Bedien-/Administratorautorisierung."
//
// **Ein Nachweis ist die Gestalt davon** ([0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md)):
// er nennt Zweck, Ziel und Stand, läuft ab und wird einmal eingelöst. Ein
// Feld `bestaetigt`, das der Aufrufer selbst setzt, wäre keiner.
func (a Ruecksetzungsart) BrauchtNachweis() bool { return a == RuecksetzungAnwendung }

// BrauchtNamen sagt, ob diese Art den Namen ihres Ziels als Bindung verlangt.
//
// **Genau die beiden, die etwas Benanntes löschen.** `storage.Nutzerablage`
// hält den Namen seit #1351 gegen den gespeicherten; was keinen Namen hat —
// eine Einstellung, der ganze Bestand —, lässt sich nicht so bestätigen.
func (a Ruecksetzungsart) BrauchtNamen() bool {
	return a == RuecksetzungProfil || a == RuecksetzungNutzer
}

// Layouteinstellungen sind die Schlüssel, die eine Layoutrücksetzung anfasst.
//
// # Zehn, und die Liste ist die Zusage
//
// **Was hier nicht steht, wird nicht angefasst** — und das ist der Grund, dass
// es diese Liste gibt statt einer Filterregel über Namen. `sprache` gehört
// nicht dazu (eine Sprache ist kein Layout), `aktualisierung` und
// `bildkonflikt` auch nicht (das ist Verhalten), und `ersteschritte` ist eine
// Erinnerung und keine Anordnung.
//
// **Ein Wächter hält sie gegen [Einstellungsreihe]**: was hier steht, gibt es
// dort — sonst setzte die Rücksetzung einen Schlüssel zurück, den niemand
// kennt.
func Layouteinstellungen() []string {
	return []string{
		EinstellungErscheinungsbild, EinstellungSchriftgroesse,
		EinstellungSchriftart, EinstellungSeitenleiste, EinstellungLeistenlage,

		// **Auch der Fensterstart** (UX-14.AK-01, #1686): wer „das Fenster
		// sieht falsch aus" sagt, meint auch seine Größe.
		EinstellungFensterstart,

		// **Und die Gestaltung aus UX-09** (#1681, AK-03): Farbwelt,
		// Reichweite der Schrift, Fensterbild und eigene Merkmale. Das ist
		// der sichere Stand, den auch die Menüleiste erreicht, wenn im
		// Fenster nichts mehr zu lesen ist.
		EinstellungFarbwelt, EinstellungSchriftbereich,
		EinstellungHintergrund, EinstellungGestaltung,
	}
}

// Betroffenes ist eine Gruppe, die eine Rücksetzung anfasst.
type Betroffenes struct {
	// Was ist die Marke — vier Flächen bauen daraus ihren eigenen Satz.
	Was string `json:"was"`

	Anzahl int `json:"anzahl"`

	// Namen nennen ein paar davon — **damit jemand entscheiden kann**, nicht
	// damit etwas geprüft wäre.
	Namen []string `json:"namen"`
}

// Die Gruppen, die eine Rücksetzung nennen kann — **Marken und keine Sätze**.
const (
	BetroffenEinstellungen = "einstellungen"
	BetroffenProfile       = "profile"
	BetroffenSets          = "sets"
	BetroffenKarten        = "karten"
	BetroffenDecks         = "decks"
	BetroffenExemplare     = "exemplare"
	BetroffenToken         = "token"
	BetroffenFreigaben     = "freigaben"
	BetroffenBilder        = "bilder"

	// BetroffenZugaenge sind Profilfreigaben, die mit einem Profil oder
	// einem Menschen gehen (#2287) — `ON DELETE CASCADE`, Wanderung 0044.
	BetroffenZugaenge = "zugaenge"
)

// Die Hindernisse einer Rücksetzung.
//
// **Dieselbe Gestalt wie bei der Verträglichkeit** ([Hindernis]): eine Marke
// und die Stelle, die sie genauer benennt. Ein zweiter Typ für dieselbe Sache
// liefe an dem Tag auseinander, an dem einer von beiden ein Feld dazubekommt.
const (
	// RuecksetzungOffeneRueckkehr: über Mengen dieses Profils steht nichts
	// fest ([0088](../../docs/entscheidungen/0088-die-epoche-einer-rueckkehr.md)).
	//
	// **Das ist die Stelle, an der T2, T5 und T6 andocken** (#1428): „Bei
	// ungeklärten Handelsverbindlichkeiten Profillöschung sperren oder
	// belegwahrenden expliziten Klärungs-/Wiederaufnahmeweg verlangen."
	// Heute gibt es eine solche Verbindlichkeit — die unklare Menge einer
	// Rückkehr —, und sie sperrt. Wer quittiert hat, kann löschen; das ist
	// der belegwahrende Weg.
	RuecksetzungOffeneRueckkehr = "offeneRueckkehr"

	// RuecksetzungFremdesZiel: das Ziel gehört dem Akteur nicht.
	//
	// **Das ist die Zusage aus der Abnahme**: „Kein Profilverwalter erhält
	// durch umfassenden Reset Rechte über andere Eigentümer." Wer ein Profil
	// zurücksetzen will, das nicht in seiner Liste steht, oder einen anderen
	// Menschen löschen will, kommt hier heraus — **vor** jeder Löschung.
	RuecksetzungFremdesZiel = "fremdesZiel"

	// RuecksetzungLaufenderAuftrag: in diesem Prozess läuft noch etwas.
	//
	// **„Geordnet stoppen" heißt hier: nicht mitten hineinlöschen** (#1428).
	// Ein Auftrag lässt sich nicht abbrechen — das Auftragsbuch kennt keinen
	// Abbruch (U11) —, und ihn zu übergehen hieße, einem laufenden Schreiber
	// den Bestand unter den Händen wegzunehmen. Die Rücksetzung wartet, und
	// die Vorschau **nennt**, worauf.
	//
	// **Was in einem anderen Fenster läuft, sieht dieses nicht**: das Buch
	// steht im Arbeitsspeicher je Prozess. Die Zusage reicht so weit wie die
	// Messung (§10).
	RuecksetzungLaufenderAuftrag = "laufenderAuftrag"

	// RuecksetzungLetzterNutzer: der letzte Mensch lässt sich nicht löschen.
	RuecksetzungLetzterNutzer = "letzterNutzer"

	// RuecksetzungNachweisFehlt: die ganze Anwendung braucht einen.
	RuecksetzungNachweisFehlt = "nachweisFehlt"

	// RuecksetzungOhneTausch: diese Fläche tauscht den Bestand nicht — G09
	// (#2273).
	//
	// **Der Dienst hält eine Verbindung über seine ganze Laufzeit**, und ein
	// Neuanfang ist ein Tausch: danach hielte jede Naht einen Griff auf eine
	// geschlossene Verbindung. Neu angefangen wird am Fenster oder mit der
	// Kommandozeile, die beide ihren Lebenszyklus danach neu aufbauen.
	RuecksetzungOhneTausch = "ohneTausch"
)

// Ruecksetzungsvorschau ist, was eine Rücksetzung anfassen würde.
//
// **Sie ändert nichts.** Wer wissen will, was geschieht, bekommt die Auskunft,
// ohne dass etwas geschieht — dieselbe Zusage wie bei der Datenpflege
// ([0089](../../docs/entscheidungen/0089-ein-befund-ist-kein-auftrag.md)).
type Ruecksetzungsvorschau struct {
	Art Ruecksetzungsart `json:"art"`

	Betroffen []Betroffenes `json:"betroffen"`

	// Hindernisse sind die Gründe, aus denen es nicht geht.
	Hindernisse []Hindernis `json:"hindernisse"`

	// NameNoetig sagt, dass diese Art den **Namen** des Ziels als Bestätigung
	// verlangt — abgetippt, nicht angekreuzt.
	//
	// **Nicht dasselbe wie [Ruecksetzungsart.LoeschtDaten]**: die ganze
	// Anwendung löscht am meisten und verlangt **keinen** Namen, sondern
	// einen Nachweis. Ein Name wäre dort auch keine Bindung — es gibt
	// keinen, den man abtippen könnte.
	NameNoetig bool `json:"nameNoetig"`

	// NachweisNoetig sagt, dass diese Art einen [Nachweis] verlangt.
	//
	// **Kein Hindernis, sondern eine Ansage.** Ein fehlender Nachweis hält
	// die Ausführung an — aber erst dort: die Vorschau soll zeigen, was
	// geschähe, und nicht so aussehen, als ginge es grundsätzlich nicht.
	NachweisNoetig bool `json:"nachweisNoetig"`

	// Revision ist die Marke über das, was diese Vorschau zeigt.
	//
	// **Sie bindet die Ausführung an das Gesehene** — Befund 12 des
	// Zweitblicks über Phase 5.2. Vorher rechnete `Zuruecksetzen` eine neue
	// Vorschau und löschte auf **deren** Stand: was zwischen dem Hinsehen und
	// dem Klicken dazukam, ging ohne Zustimmung mit. Dieselbe Bindung wie bei
	// der Datenpflege ([0089](../../docs/entscheidungen/0089-ein-befund-ist-kein-auftrag.md))
	// und beim Einstellungsaustausch (0079).
	//
	// **Der Name bleibt daneben** und ersetzt sie nicht: er bindet an das
	// **Ziel** („dieses Profil, nicht das daneben"), die Revision an den
	// **Umfang** („diese drei Sets, nicht die vier von jetzt").
	Revision string `json:"revision"`

	// SicherungEmpfohlen sagt, dass vorher gesichert werden sollte.
	//
	// **Nicht „wird gesichert".** Das Issue verlangt: „Sicherung anbieten,
	// Verzicht nur explizit." Angeboten wird, was hier steht; verzichtet wird
	// im Aufruf.
	SicherungEmpfohlen bool `json:"sicherungEmpfohlen"`
}

// Ruecksetzungsmarke gibt die Marke über das, was eine Vorschau zeigt.
//
// **Über Art, Umfang und die genannten Namen.** Hier stand „nicht über die
// Namen" — und das war zu wenig: ein Set »Alt« in der Vorschau und ein Set
// »Neu« bei der Ausführung ergaben dieselbe Marke, weil die Zahl stimmte.
// Angestrichen vom Nachlauf zum Zweitblick über Phase 5.2 (Befund 9).
//
// **Die Namen sind das, was ein Mensch gesehen hat** — mehr gibt die Vorschau
// nicht her, und genau daran soll die Ausführung gebunden sein. **Sortiert**,
// weil eine Reihenfolge aus dem Bestand keine Aussage ist.
func Ruecksetzungsmarke(art Ruecksetzungsart, betroffen []Betroffenes) string {
	bau := &strings.Builder{}

	fmt.Fprintf(bau, "art=%s\n", string(art))

	gruppen := make([]string, 0, len(betroffen))

	for _, eines := range betroffen {
		namen := slices.Clone(eines.Namen)
		slices.Sort(namen)

		gruppen = append(gruppen, fmt.Sprintf("%s=%d:%s",
			eines.Was, eines.Anzahl, strings.Join(namen, ",")))
	}

	slices.Sort(gruppen)

	for _, gruppe := range gruppen {
		fmt.Fprintf(bau, "%s\n", gruppe)
	}

	return fmt.Sprintf("%x", sha256.Sum256([]byte(bau.String())))[:MarkeLaenge]
}

// Bereit sagt, ob diese Rücksetzung laufen darf.
func (v Ruecksetzungsvorschau) Bereit() bool { return len(v.Hindernisse) == 0 }

// LoeschtDaten sagt, ob diese Art Daten wegnimmt — und nicht nur eine Wahl.
//
// **Daran hängt das Angebot der Sicherung.** Eine Layoutrücksetzung braucht
// keine; ein gelöschter Nutzer ist ohne sie fort.
func (a Ruecksetzungsart) LoeschtDaten() bool {
	switch a {
	case RuecksetzungProfil, RuecksetzungNutzer, RuecksetzungAnwendung:
		return true
	case RuecksetzungLayout, RuecksetzungProfileinstellungen,
		RuecksetzungNutzereinstellungen:
		return false
	}

	return false
}
