package app

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Die zwei Nähte einer Partie** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum sie in einem Verbund stehen und nicht als zwei Felder in `App`
//
// **Weil `anwendung.go` damit über die Zeilengrenze aus §2 ging** — und weil
// der Schnitt trägt: beide gehören zu derselben Sache, und wer eine Partie
// einsetzt, setzt fast immer beide.
//
// **Zwei und nicht eine**, denn es sind zwei Fragen: die eine **führt** eine
// Partie, die andere **baut** sie. Eine Fläche, die nur liest, braucht den
// Aufbau nicht.
//
// **Nil heißt: keine.** Eine Anwendung ohne Bestand hat keine Partien, und
// das ist kein Fehler — es ist der Normalfall in einem Lauf (§3).
type partienaht struct {
	leitung Partieleitung
	aufbau  Partieaufbau

	// freigabe ist die Partie, die für Gäste offensteht (G5, #1462).
	//
	// **Sie steht hier und nicht als eigenes Feld in `App`** — aus demselben
	// Grund wie die beiden darüber, und weil sie zu derselben Sache gehört.
	// **Anders als sie ist sie aber keine Naht**: sie beschreibt einen
	// **Lauf**, hat ein Schloss und ändert sich, während gespielt wird.
	freigabe freigabefelder

	// beitritt ist der **fremde** Tisch, an dem dieses Programm sitzt (G5,
	// #1462) — die Gegenrichtung zur Freigabe.
	//
	// **Einer zur Zeit.** Zwei fremde Tische hießen: jede hereinkommende
	// Nachricht müsste sagen, welchen sie meint — und das wäre wieder eine
	// Angabe der Gegenseite.
	beitritt beitrittsfelder

	// runde sind die offenen Gastsitzungen dieses Wirts (G5, #1462).
	//
	// # Warum sie neben der Freigabe steht und nicht darin
	//
	// **Weil sie eine andere Lebensdauer hat.** Die Freigabe ist die
	// **Zusage**, dass ein neuer Gast einen Sitz bekommt; wer sie beendet,
	// schickt niemanden vom Tisch. Die Runde führt, wer **gerade sitzt** —
	// und das überlebt das Ende der Zusage.
	//
	// **Ihr Nullwert ist benutzbar**: eine Anwendung ohne offenen Tisch hat
	// eine leere Runde, und das ist der Normalfall.
	runde service.Partierunde
}

// beitrittsfelder ist, was eine Anwendung über den fremden Tisch hält.
type beitrittsfelder struct {
	schloss sync.Mutex

	// griff ist die Leitung zum Wirt — **die Schnittstelle und nicht der
	// Griff** (§2).
	//
	// **Weil hier nur zwei Dinge gebraucht werden**: schicken und leeren.
	// Der erste Anlauf hielt den `*service.Sitzungsgriff`, und damit ließ
	// sich kein Weg dieses Tisches prüfen, ohne zwei Identitäten und eine
	// echte Leitung aufzubauen — der Fehlerweg von [App.PartieWuenschen]
	// stand deshalb unbetreten da (§3).
	griff  service.Angenommene
	partie string

	// lauf zählt die Beitritte — jedes Setzen und jedes Räumen.
	//
	// **Er sagt einer Schleife, ob ihr Tisch noch ihrer ist.** Wer
	// aufgestanden ist oder sich woanders hingesetzt hat, bekommt von der
	// alten Schleife weder eine Partie nachgetragen noch ein „pausiert"
	// hinterher ([App.tischVerloren]). Ein Vergleich des Griffs täte es auch
	// — aber ein Doppel, das kein Zeiger ist, ließe ihn zur Laufzeit
	// abstürzen.
	lauf uint64

	// sitzung ist die **eine** Sendeseite dieses Tisches.
	//
	// # Warum sie gemerkt wird und nicht je Ruf entsteht
	//
	// **Weil sie die Folgenummern vergibt.** Hier stand „je Ruf eine neue und
	// keine gemerkte — der Wirt liest beides gar nicht", und das war zur
	// Hälfte richtig: der Wirt liest den **Absender** nicht, die **Folge**
	// aber sehr wohl. Jede frische Sitzung fing wieder bei eins an, und
	// `domain.Empfaengerstand` hielt alles ab der zweiten Nachricht für eine
	// Wiederholung — **still**.
	//
	// **Sichtbar wurde es erst mit der Rückkehr**: sie belegt die Folge eins,
	// und damit fiel der erste Satz eines Gastes weg statt seines zweiten.
	sitzung *service.Partiesitzung

	// post stellt die Bildantworten des Wirts zu — **eine je Beitritt**, denn
	// geschickt wird über einen Gast je Ruf und empfangen in der Schleife
	// (P9, #2377).
	post *service.Bildpost

	// mitgebracht ist das Deck, mit dem sich dieser Gast setzt, wenn sein
	// Sitz wartet (P9b, #2377) — `nil`, wenn er nichts mitbringt.
	mitgebracht *domain.Mitgebracht

	// absage ist der Grund, mit dem der Wirt zuletzt ablehnte (P9c).
	absage string

	// zeigerstand vergibt die Stände der **eigenen** Präsenz als Gast.
	//
	// # Warum er hier liegt und nicht im Gast
	//
	// **Weil er die einzelne Handlung überleben muss.** `service.Partiegast`
	// wird je Ruf neu gebaut — ein Zähler darin finge bei jedem Zeigen wieder
	// bei eins an, und das andere Ende wiese alles nach dem ersten ab.
	//
	// **Und warum nicht in der Sitzung**: eine Präsenz verbraucht keine
	// Folge (der Grund steht bei [domain.Praesenz]), sie hat also keine
	// Stelle in der Reihenfolge, die jene führt.
	//
	// **Er wird nicht zurückgesetzt, wenn dieses Programm den Tisch
	// wechselt** — und das ist richtig: der Wirt am anderen Ende baut für
	// jede Sitzung einen frischen Prüfer, und eine Zahl, die nur steigt,
	// kommt bei jedem durch.
	zeigerstand atomic.Uint64

	// folge und abdruck sind, wo dieser Gast im Strom der Partie steht.
	//
	// **Hier und nicht im `service.Partiegast`**, weil ein Gast je Ruf neu
	// gebaut wird und ein Stand den Abriss überleben muss — genau daran ist
	// der Zeigerstand daneben schon einmal gescheitert.
	folge   int64
	abdruck string

	// standpartie ist die Partie, zu der dieser Stand gehört.
	//
	// **Ohne sie wäre der Stand ein Zettel ohne Aufschrift.** Wer an einem
	// Tisch bei Folge 40 stand und sich danach an einen anderen setzt, käme
	// dort als Rückkehrer bei Folge 40 an — und bekäme einen Abbruch für eine
	// Partie, die er nie gesehen hat. Sie steht **neben** dem Stand und wird
	// beim Lesen verglichen; ein Rücksetzen beim Wechsel wäre eine zweite
	// Stelle, an der man es vergessen kann.
	standpartie string

	// gefragt ist der Augenblick, in dem der Rückkehrwunsch hinausging.
	//
	// **Null heißt: es wartet gerade keiner.** Sie wird beim ersten, was
	// zurückkommt, wieder auf null gesetzt — eine Messung, die zweimal
	// abschlösse, zählte beim zweiten Mal die Zeit seit dem ersten mit.
	gefragt time.Time

	// rueckkehr ist, wie lange die letzte Runde gedauert hat (G5, #1462).
	//
	// **Das ist der RTT, den dieses Protokoll wirklich dreht.** Ein eigener
	// Takt wäre eine zweite Wahrheit über dieselbe Leitung und kostete bei
	// jedem Schlag eine Nachricht.
	rueckkehr time.Duration

	// pausiert heißt: die Leitung riss, ohne dass jemand aufgestanden ist.
	//
	// **Der Unterschied ist der ganze Punkt** (#1462: „Hostverlust
	// pausiert"). Wer aufsteht, hat den Tisch verlassen; wem die Leitung
	// wegbricht, dem gehört sein Platz weiter — [App.tischVerloren].
	pausiert bool
}

// beitritt gibt den offenen fremden Tisch — nil heißt: keiner.
func (a *App) beitritt() (
	griff service.Angenommene, partie string,
	sitzung *service.Partiesitzung,
) {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	return a.partie.beitritt.griff, a.partie.beitritt.partie,
		a.partie.beitritt.sitzung
}

// bildpost gibt die Post des laufenden Beitritts — sie entsteht mit ihm in
// [App.beitrittMerken], und gefragt wird nur, wer an einem Tisch sitzt.
func (a *App) bildpost() *service.Bildpost {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	return a.partie.beitritt.post
}

// beitrittMerken trägt ihn ein — oder räumt ihn weg — und gibt den Lauf.
//
// **Ohne Partie gibt es noch keine Sendeseite**: jede Nachricht trägt ihre
// Partie, und eine ohne weist schon `domain.Partienachricht.Pruefe` ab. Sie
// entsteht, sobald der Wirt die Partie nennt ([App.partieErfahren]).
func (a *App) beitrittMerken(
	griff service.Angenommene, partie string,
) (sitzung *service.Partiesitzung, lauf uint64) {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	a.partie.beitritt.lauf++
	a.partie.beitritt.griff, a.partie.beitritt.partie = griff, partie
	a.partie.beitritt.sitzung = nil
	a.partie.beitritt.post = &service.Bildpost{}

	if griff != nil && partie != "" {
		a.partie.beitritt.sitzung = sendeseite(griff, partie)
	}

	return a.partie.beitritt.sitzung, a.partie.beitritt.lauf
}

// partieErfahren trägt die Partie ein, die der Wirt genannt hat (UX-08.AK-04,
// #1680).
//
// **Nur für denselben Beitritt** — der Grund steht bei
// [beitrittsfelder.lauf].
func (a *App) partieErfahren(
	lauf uint64, partie string,
) (sitzung *service.Partiesitzung, gilt bool) {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	if a.partie.beitritt.lauf != lauf || a.partie.beitritt.griff == nil {
		return nil, false
	}

	a.partie.beitritt.partie = partie
	a.partie.beitritt.sitzung = sendeseite(a.partie.beitritt.griff, partie)

	return a.partie.beitritt.sitzung, true
}

// sendeseite baut die **eine** Sendeseite eines Tisches.
//
// **Eine je Tisch und nicht eine je Ruf** — der Grund steht bei
// [beitrittsfelder.sitzung].
func sendeseite(
	griff service.Angenommene, partie string,
) *service.Partiesitzung {
	return service.NeuePartiesitzung(
		partie, gastplatzhalter, griff,
		service.NeueReihenfolge(erstePartieepoche))
}

// tischVerloren räumt einen Tisch ab, dessen Leitung endete — und merkt,
// dass der Platz pausiert.
//
// # Wenn der Wirt weg ist — G5 (#1462)
//
// „Hostverlust **pausiert**; **nur derselbe Host** setzt fort. **Kein
// automatisches Gewinnergebnis** und **keine automatische Hostmigration**."
// Alle drei sind Verzichte, und alle drei gelten **durch den Bau**: der Gast
// hält den Vollzustand nicht — er liegt beim Wirt —, also kann er weder
// weiterspielen noch ein Ergebnis setzen noch den Wirt ersetzen. Was er tut,
// ist dies hier: den Platz als pausiert merken und auf die Rückkehr warten.
// Fortsetzen kann nur der Rechner, der die Partie führt
// (`service.Partieleitung.Fortsetzen`).
//
// **Hier standen bis #2436 zwei Regeln in `domain`** (`BeiWirtsverlust`,
// `DarfFortsetzen`), auf die die Kommentare zeigten; gerufen hat sie niemand,
// denn es gibt keinen Weg, auf dem ein Gast den Stand des Wirts setzt.
//
// **Nur, wenn er noch der ist, zu dem die Schleife gehört.** Wer aufgestanden
// ist, hat den Tisch verlassen und ist nicht pausiert; wer sich woanders
// hingesetzt hat, sitzt dort weiter.
//
// **Und eine Absage des Wirts merkt es sich** (P9c, #2377) — samt Meldung
// ans Fenster, das sonst bis „Nachsehen“ an einem Tisch säße, den es nicht
// mehr gibt.
func (a *App) tischVerloren(lauf uint64, ende error) {
	a.partie.beitritt.schloss.Lock()

	if a.partie.beitritt.lauf != lauf {
		a.partie.beitritt.schloss.Unlock()

		return
	}

	a.partie.beitritt.lauf++
	a.partie.beitritt.griff, a.partie.beitritt.partie = nil, ""
	a.partie.beitritt.sitzung = nil
	a.partie.beitritt.pausiert = true
	a.partie.beitritt.absage = ""

	var absage service.Absage
	if errors.As(ende, &absage) {
		a.partie.beitritt.absage = absage.Grund
	}

	a.partie.beitritt.schloss.Unlock()

	if a.melden != nil {
		a.melden(a.fenster, EreignisPartiebeitritt)
	}
}

// absage gibt den Grund, mit dem der Wirt zuletzt ablehnte.
func (a *App) absage() string {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	return a.partie.beitritt.absage
}

// Merke trägt ein, wo dieser Gast zuletzt stand — [service.Standhalter].
func (a *App) Merke(folge int64, abdruck string) {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	a.partie.beitritt.folge = folge
	a.partie.beitritt.abdruck = abdruck
	a.partie.beitritt.standpartie = a.partie.beitritt.partie

	// **Hier schließt die Messung der Rückkehr** (G5, #1462).
	//
	// **Hier und nicht im Abzugsweg**: der Wirt antwortet auf einen Wunsch
	// je nach Lücke mit `weiter`, mit Ereignissen oder mit einem Abzug —
	// **jeder** dieser Wege kommt an dieser Stelle vorbei, und eine Messung
	// an einem von ihnen bliebe bei den anderen leer.
	if !a.partie.beitritt.gefragt.IsZero() {
		a.partie.beitritt.rueckkehr = time.Since(a.partie.beitritt.gefragt)
		a.partie.beitritt.gefragt = time.Time{}
	}
}

// gefragt merkt, dass ein Rückkehrwunsch hinausgegangen ist.
func (a *App) gefragt(wann time.Time) {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	a.partie.beitritt.gefragt = wann
}

// letzteRueckkehr gibt die gemessene Runde — und ob je eine gemessen wurde.
//
// **Zwei Rückgaben und keine Null als Auskunft**: „unter einer Millisekunde"
// und „nie gemessen" sind zwei Antworten, und die Abnahme des Issues verlangt
// ausdrücklich, dass fehlende Messgrößen als unbekannt erscheinen.
func (a *App) letzteRueckkehr() (time.Duration, bool) {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	return a.partie.beitritt.rueckkehr, a.partie.beitritt.rueckkehr > 0
}

// Stand gibt es zurück — null und leer, wenn er noch nie dabei war.
//
// **Und ebenso, wenn er zu einer anderen Partie gehört** — der Grund steht
// bei [beitrittsfelder.standpartie].
func (a *App) Stand() (folge int64, abdruck string) {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	if a.partie.beitritt.standpartie != a.partie.beitritt.partie {
		return 0, ""
	}

	return a.partie.beitritt.folge, a.partie.beitritt.abdruck
}

// istPausiert sagt, ob der letzte Tisch weggebrochen ist.
func (a *App) istPausiert() bool {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	return a.partie.beitritt.pausiert
}

// tischVergessen räumt den Stand weg — nur, wenn jemand wirklich aufsteht.
//
// **Und nicht beim Abriss.** Was ein Gast gesehen hat, ist die Grundlage
// seiner Rückkehr; wer es beim ersten Wackler wegräumte, ließe ihn jedes Mal
// von vorn anfangen — und der Wirt schickte jedes Mal den ganzen Abzug.
func (a *App) tischVergessen() {
	a.partie.beitritt.schloss.Lock()
	defer a.partie.beitritt.schloss.Unlock()

	a.partie.beitritt.folge, a.partie.beitritt.abdruck = 0, ""
	a.partie.beitritt.standpartie = ""
	a.partie.beitritt.pausiert = false
}
