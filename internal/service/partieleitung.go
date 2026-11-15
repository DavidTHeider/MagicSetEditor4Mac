// **Die Partieleitung** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was diese Schicht dazutut
//
// Der Speicher schreibt Zeilen, `domain` rechnet, der Regelausleger sagt, was
// ein Befehl bewirkt — hier steht, **was gilt**, und das sind vier Dinge, die
// keine der anderen halten kann (§2):
//
//   - **Ein Befehl kommt von einem Sitz dieser Partie.** Ohne die Frage
//     schriebe eine erfundene Kennung in eine fremde Partie.
//   - **Derselbe Befehl erzeugt keinen zweiten Zug.** Die Kennung kommt vom
//     Absender, denn nur er weiß, ob er noch einmal dasselbe meint.
//   - **Eine erwartete Fassung, die nicht mehr gilt, ist ein Konflikt** —
//     und keine stille Anwendung auf einen Stand, den der Absender nicht
//     kannte.
//   - **Die Folge zählt hier**, und jedes Ereignis trägt den Abdruck seines
//     Vorgängers. Eine gebrochene Kette wird abgewiesen und **nie still
//     repariert** (#1458).
//
// # Der Wirt ist diese Leitung
//
// **In einer lokalen Partie ist der Wirt dieses Programm.** Er serialisiert,
// weil er der Einzige ist, der schreibt; wer über ein Netz dazukommt, schickt
// seinen Befehl hierher, und das ist Werkstück G5
// (#1462).
//
// **Deshalb steht hier keine Wirtsprüfung gegen den Absender**: jeder Sitz
// darf Befehle **stellen**, angewendet werden sie an einer Stelle. Wer daraus
// machte „nur der Wirt darf spielen", hätte eine Partie für einen.

package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Regelausleger sagt, was ein Befehl bewirkt — und was gerade geht.
//
// **Eine Schnittstelle dort, wo sie gebraucht wird** (§2). G1 liefert den
// Sandkasten; **welche** Regeln ein Format wirklich trägt, ist Werkstück G4
// (#1461).
type Regelausleger interface {
	// Auslegen gibt die Ereignisse und den Zustand danach.
	//
	// **Beides zusammen**, denn das eine ohne das andere ließe sich nicht
	// prüfen: wer nur Ereignisse bekäme, müsste sie hier noch einmal
	// anwenden — und hätte damit zwei Auslegungen desselben Befehls.
	Auslegen(partie domain.Partie, befehl domain.Spielbefehl) (
		[]domain.Ereignis, domain.Partie, error)

	// Erlaubt sagt, welche Befehlsarten `fuer` gerade stellen darf.
	Erlaubt(partie domain.Partie, fuer string) []string
}

// Befehlsausgang ist, was ein angenommener Befehl hinterlässt.
type Befehlsausgang struct {
	// Fassung ist die der Partie **danach**.
	Fassung int64

	Ereignisse []domain.Ereignis
}

// Partieleitung ist der Anwendungsfall „eine Partie führen".
type Partieleitung struct {
	partien  Partiespeicher
	ausleger Regelausleger

	// zug hält Lesen, Auslegen und Schreiben **eines** Zuges
	// zusammen (P9b, #2377). Das Fenster des Wirts und die Schleife jedes
	// Gastes schreiben in demselben Prozess; lasen zwei dieselbe Fassung,
	// scheiterte der zweite an der Prüfung des Speichers — und dieser
	// Konflikt beendete die Sitzung des Gastes. **Zwischen Prozessen** bleibt
	// die Prüfung des Speichers die Sperre.
	zug sync.Mutex

	// reihe hält das Verkünden in der Reihenfolge, in der `zug` annahm
	// (#2481) — sie übernimmt, bevor `zug` fällt. Der Grund steht in
	// `partieveroeffentlichung.go`.
	reihe sync.Mutex

	// geheimnis gibt das Geheimnis einer neuen Partie (#2480) — eingesetzt
	// wie Uhr und Zufall anderswo, damit ein Lauf denselben Anfang zweimal
	// bauen kann.
	geheimnis func() string
}

// NeuePartieleitung baut sie.
func NeuePartieleitung(
	partien Partiespeicher, ausleger Regelausleger,
) *Partieleitung {
	return &Partieleitung{partien: partien, ausleger: ausleger, geheimnis: rand.Text}
}

// Anlegen legt eine Partie an.
func (l *Partieleitung) Anlegen(ctx context.Context, partie domain.Partie) (
	domain.Partie, error,
) {
	if fehler := partie.Pruefe(); fehler != nil {
		return domain.Partie{}, fehler
	}

	// **Der Stand steht hier und nicht beim Aufrufer.** Eine Partie fängt
	// laufend an; wer sie pausiert anlegen wollte, hätte eine, auf die
	// niemand wartet.
	partie.Stand = domain.StandLaeuft

	// **Ein eigenes Geheimnis je Partie** (#2480, Vermerk 0201): mit ihm
	// sind die Sichtkennungen nicht zurückzurechnen.
	partie.Geheimnis = l.geheimnis()

	return l.partien.Anlegen(ctx, partie)
}

// Partie gibt eine Partie, wie `fuer` sie sehen darf.
//
// **Über die Projektion und nicht über den Zustand** — der Grund steht bei
// [domain.Partie.Sicht]: sie ist die Regel, nicht ihre Anwendung.
func (l *Partieleitung) Partie(
	ctx context.Context, kennung, fuer string,
) (domain.Partiesicht, error) {
	// **Ein an einen Sitz gebundenes Token sieht nur ihn** (#1819).
	if fehler := SitzbindungAus(ctx).Erlaubt(kennung, fuer); fehler != nil {
		return domain.Partiesicht{}, fehler
	}

	partie, fehler := l.partien.Partie(ctx, kennung)
	if fehler != nil {
		return domain.Partiesicht{}, fehler
	}

	// **Für den Bestand, der liest** (P9, #2377) — am Gastplatz der des
	// Gastes.
	return partie.SichtVon(fuer, BestandssichtAus(ctx)), nil
}

// Befehl wendet einen Befehl an — oder sagt, warum nicht.
//
// **Ohne Verkündung**: wer den Zug an einen Tisch trägt, nimmt
// [Partieleitung.BefehlVerkuendet].
func (l *Partieleitung) Befehl(
	ctx context.Context, befehl domain.Spielbefehl,
) (Befehlsausgang, error) {
	return l.inReihe(func() (Befehlsausgang, error) {
		return l.befehl(ctx, befehl)
	}, nil)
}

// befehl ist [Partieleitung.Befehl] unter dem Schloss des Zuges.
func (l *Partieleitung) befehl(
	ctx context.Context, befehl domain.Spielbefehl,
) (Befehlsausgang, error) {
	if fehler := befehl.Pruefe(); fehler != nil {
		return Befehlsausgang{}, fehler
	}

	// **Und es stellt nur für seinen Sitz** (#1819).
	if fehler := SitzbindungAus(ctx).Erlaubt(befehl.PartieID, befehl.Absender); fehler != nil {
		return Befehlsausgang{}, fehler
	}

	partie, fehler := l.partien.Partie(ctx, befehl.PartieID)
	if fehler != nil {
		return Befehlsausgang{}, fehler
	}

	// **Zuerst die Wiederholung.** Ein zweiter Ruf mit derselben Kennung ist
	// kein Fehler — er ist derselbe Wunsch und bekommt dieselbe Antwort.
	// Stünde diese Frage nach der Fassungsprüfung, bekäme er einen Konflikt,
	// obwohl er längst gewirkt hat.
	schon, gab, gefragt := l.partien.Befehlsergebnis(
		ctx, befehl.PartieID, befehl.Kennung)
	if gefragt != nil {
		return Befehlsausgang{}, gefragt
	}

	if gab {
		return Befehlsausgang{Fassung: partie.Fassung, Ereignisse: schon}, nil
	}

	// **Eine Absage nennt keine Kennung** (Vermerk 0201, Punkt 7) — die
	// Sätze der Domäne zitieren sie, und der Grund geht über die Leitung.
	ausgang, fehler := l.anwenden(ctx, partie, befehl)

	return ausgang, ohneKennungen(partie, fehler)
}

// anwenden legt den Befehl aus und schreibt, was dabei herauskommt.
//
// **Getrennt vom Ruf darüber**, weil dort eine andere Frage steht: ob dieser
// Befehl überhaupt neu ist. Die Wiederholung ist kein Sonderfall der
// Anwendung, sondern ihr Gegenteil — sie wendet nichts an.
func (l *Partieleitung) anwenden(
	ctx context.Context, partie domain.Partie, befehl domain.Spielbefehl,
) (Befehlsausgang, error) {
	// **Wer ein Objekt nennt, nennt seine Sichtkennung** (#2480, Vermerk
	// 0201, Punkt 6) — hier wird sie wieder die Kennung, die es jetzt trägt.
	// **Ein Gast über die Leitung nennt nie eine Kennung.**
	if fehler := keineKennungVonDerLeitung(ctx, partie, befehl); fehler != nil {
		return Befehlsausgang{}, fehler
	}

	befehl.Nutzlast = partie.Echtnamen(befehl.Nutzlast)

	if fehler := l.pruefeAnnahme(partie, befehl); fehler != nil {
		return Befehlsausgang{}, fehler
	}

	// **Die Rücknahme meint den Strom, nicht nur den Stand** — P8 (#2376):
	// welcher Befehl zuletzt kam, weiß nur der Speicher.
	if befehl.Art == BefehlZuruecknehmen {
		return l.zuruecknehmen(ctx, partie, befehl)
	}

	ereignisse, danach, ausgelegt := l.ausleger.Auslegen(partie, befehl)
	if ausgelegt != nil {
		return Befehlsausgang{}, ausgelegt
	}

	return l.schreiben(ctx, partie, befehl.Kennung, danach, ereignisse)
}

// pruefeAnnahme sagt, ob dieser Befehl überhaupt angenommen wird.
func (l *Partieleitung) pruefeAnnahme(
	partie domain.Partie, befehl domain.Spielbefehl,
) error {
	if partie.Stand != domain.StandLaeuft {
		return fmt.Errorf(
			"%w: diese Partie ist %s — sie nimmt keine Befehle an",
			domain.ErrConflict, partie.Stand)
	}

	if partie.SitzVon(befehl.Absender) == nil {
		return fmt.Errorf(
			"%w: %q sitzt nicht an diesem Tisch", domain.ErrValidation,
			befehl.Absender)
	}

	// **Null heißt: gegen die, die gerade gilt.** Der bequeme Fall für eine
	// Fläche ohne Nebenläufigkeit; wer sicher gehen will, nennt eine.
	if befehl.Erwartet != 0 && befehl.Erwartet != partie.Fassung {
		return fmt.Errorf(
			"%w: der Befehl rechnet gegen die Fassung %d, es gilt %d — "+
				"dazwischen ist etwas geschehen, das er nicht kennt",
			domain.ErrConflict, befehl.Erwartet, partie.Fassung)
	}

	return nil
}

// Verlauf gibt den Ereignisstrom, wie ein Sitz ihn sehen darf.
//
// # Die Zusage
//
// **#1462**: „kein Replay exportiert unbemerkt fremde private Informationen"
// und „Replay/Chat/Logs eines Teilnehmers enthalten nur seine freigegebenen
// Informationen."
//
// **Der Speicher liefert ungefiltert**, und diese Stelle filtert. Die
// umgekehrte Aufteilung — jeder Speicher filtert selbst — hätte dieselbe
// Regel an so vielen Stellen, wie es Speicher gibt, und wenn die erste sie
// vergäße, merkte es niemand.
//
// **Gerechnet wird in `domain`** ([domain.RueckschauFuer]), denn wer was sehen
// darf, ist eine fachliche Aussage und keine Frage der Ablage.
//
// # Warum zuerst die Partie geholt wird
//
// **Weil „noch nichts geschehen" und „gibt es nicht" zwei Antworten sind.**
// [Partiespeicher.Ereignisse] fragt eine Tabelle nach Zeilen, und keine ist
// keine — eine erfundene Kennung bekäme von dort eine leere Liste. Wer die
// durchreichte, antwortete auf einen Tippfehler mit „in dieser Partie ist
// noch nichts geschehen", und der Fragende suchte danach am falschen Ende
// (§8). **Ein Ruf mehr je Rückschau**, und er ist ihn wert.
func (l *Partieleitung) Verlauf(
	ctx context.Context, kennung, fuer string,
) ([]domain.Ereignis, error) {
	if fehler := SitzbindungAus(ctx).Erlaubt(kennung, fuer); fehler != nil {
		return nil, fehler
	}

	if _, fehler := l.partien.Partie(ctx, kennung); fehler != nil {
		return nil, fehler
	}

	strom, fehler := l.partien.Ereignisse(ctx, kennung, 0, 0)
	if fehler != nil {
		return nil, fehler
	}

	// **Und hier kommen die Einwilligungen dazu** (G5, #1462: „Consent für
	// gegebenenfalls spätere Offenlegung"). Sie wirken **nur** in der
	// Rückschau; warum, steht bei [domain.RueckschauFuer].
	freigaben, fehler := l.partien.Freigaben(ctx, kennung)
	if fehler != nil {
		return nil, fehler
	}

	return domain.RueckschauFuer(strom, fuer, freigaben), nil
}

// Pausieren hält eine Partie an — sie wartet auf ihren Wirt.
func (l *Partieleitung) Pausieren(ctx context.Context, kennung string) error {
	if fehler := SitzbindungAus(ctx).ErlaubtPartie(kennung); fehler != nil {
		return fehler
	}

	_, fehler := l.partien.Standsetzen(ctx, kennung, domain.StandPausiert)

	return fehler
}

// Fortsetzen nimmt sie wieder auf.
func (l *Partieleitung) Fortsetzen(ctx context.Context, kennung string) error {
	if fehler := SitzbindungAus(ctx).ErlaubtPartie(kennung); fehler != nil {
		return fehler
	}

	partie, fehler := l.partien.Partie(ctx, kennung)
	if fehler != nil {
		return fehler
	}

	// **Eine beendete Partie kommt nicht zurück.** Der Grund steht in
	// #1458: „dauerhafter Verlust führt zu explizitem Abbruch ohne
	// erfundenes Ergebnis" — und ein Abbruch, den man zurücknehmen kann,
	// ist keiner.
	if partie.Stand == domain.StandBeendet {
		return fmt.Errorf(
			"%w: diese Partie ist beendet — ein Abbruch, den man zurücknehmen "+
				"kann, ist keiner", domain.ErrConflict)
	}

	_, fehler = l.partien.Standsetzen(ctx, kennung, domain.StandLaeuft)

	return fehler
}

// Beenden schließt eine Partie ab.
func (l *Partieleitung) Beenden(ctx context.Context, kennung string) error {
	if fehler := SitzbindungAus(ctx).ErlaubtPartie(kennung); fehler != nil {
		return fehler
	}

	_, fehler := l.partien.Standsetzen(ctx, kennung, domain.StandBeendet)

	return fehler
}
