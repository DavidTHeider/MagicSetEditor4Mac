// **Preise holen, ablegen und zusammenzählen** — Werkstück 3 von Phase 3.4
// (#1165).
//
// # Was diese Schicht dazutut
//
// Der Klient bringt Zahlen, der Speicher legt Zeilen ab. Hier steht, **was
// daraus wird**: eine Momentaufnahme bekommt ihren Zeitpunkt, und eine
// Sammlung bekommt ihre Summe — je Währung, mit dem Alter des ältesten Preises,
// der darin steckt.
//
// # Und was sie ausdrücklich nicht tut
//
// **Sie holt nicht von allein.** Ein Programm, das beim Öffnen Netzverkehr für
// dreihundert Karten auslöst, überrascht seine Nutzenden — dasselbe Argument
// wie beim Paketbezug
// ([0052](../../docs/entscheidungen/0052-die-sammlung-ist-nicht-das-set.md)).

package service

import (
	"context"
	"slices"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Preisablage ist, was die Pflege zum Lesen und Schreiben braucht.
type Preisablage interface {
	Eintragen(ctx context.Context, staende ...domain.Preisstand) error
	Neueste(ctx context.Context, printing string) ([]domain.Preisstand, error)

	// NeuesteAlle gibt die jüngsten Stände **aller** Drucke.
	//
	// **Für das Portfolio und nur dafür** (#1292). Wer einen einzelnen Druck
	// ansieht, nimmt [Preisablage.Neueste]; wer über die ganze Sammlung
	// rechnet, würde damit einmal je Exemplar fragen.
	NeuesteAlle(ctx context.Context) ([]domain.Preisstand, error)
	// Verlauf gibt eine Reihe in der Zeit.
	//
	// **Die Reihe kommt als [domain.Preisstand]** und nicht als vier
	// Argumente (A3, #1442): sie hat seither sieben Teile, und sieben
	// Zeichenketten nebeneinander verwechselt irgendwann jemand.
	Verlauf(ctx context.Context, printing string, reihe domain.Preisstand) (
		[]domain.Preisstand, error)
}

// Preisquellenklient holt die Preise eines Drucks.
//
// **Als Schnittstelle und nicht als `*scryfall.Klient`** (§2, §16): so braucht
// kein Lauf das Netz, und eine zweite Quelle passte daneben.
//
// **Und er gibt neutrale Stände** (A3, #1442): eine zweite Quelle passt so
// daneben, statt diesen Code zu ändern. Was ein Feld der Quelle **bedeutet**,
// weiß ihr Paket (§16, dritte Schicht).
//
// **Was der Klient nicht setzt, ist [domain.Preisstand.Erhoben]**: wann
// gefragt wurde, weiß der Dienst — er hält die Uhr.
type Preisquellenklient interface {
	PreiseFuerDruck(ctx context.Context, kennung string) (
		[]domain.Preisstand, time.Duration, []domain.Befund, error)

	// Name sagt, wer antwortet.
	//
	// **Und er wird gebraucht, wenn die Antwort leer ist** (A3, #1442): „die
	// Quelle nennt hier nichts mehr" ist erst dann eine Aussage, wenn
	// feststeht, **welche** Quelle geantwortet hat. Aus den Ständen ist das
	// nicht zu lesen — es sind keine da.
	Name() domain.Preisquelle
}

// Preispflege ist der Anwendungsfall „was ist das gerade wert".
type Preispflege struct {
	haushaltswahl

	ablage Preisablage

	// sammlung ist, woraus die Übersicht ihre Exemplare nimmt.
	//
	// **Sie steht hier und nicht im Aufrufer**: „was ist meine Sammlung wert"
	// ist ein Anwendungsfall, und ein Anwendungsfall gehört in eine Schicht.
	sammlung Sammlungsablage

	quelle Preisquellenklient
	jetzt  func() time.Time
}

// NeuePreispflege baut sie.
//
// **Die Uhr kommt herein.** Ein Zeitstempel, den nur `time.Now` liefern kann,
// macht jeden Lauf darüber zu einer Wette auf die Ausführungsdauer.
func NeuePreispflege(
	ablage Preisablage, sammlung Sammlungsablage, quelle Preisquellenklient,
	jetzt func() time.Time,
) *Preispflege {
	return &Preispflege{
		ablage: ablage, sammlung: sammlung, quelle: quelle, jetzt: jetzt,
	}
}

// Holen fragt die Quelle nach den Preisen eines Drucks und legt sie ab.
//
// **Der zweite Rückgabewert ist das Alter der Antwort** — sie kann aus dem
// Zwischenspeicher kommen, und dann ist sie älter als jetzt (§16, Satz 5).
func (p *Preispflege) Holen(ctx context.Context, printing string) (
	staende []domain.Preisstand, alter time.Duration,
	befunde []domain.Befund, fehler error,
) {
	preise, alter, befunde, fehler := p.quelle.PreiseFuerDruck(ctx, printing)
	if fehler != nil {
		return nil, 0, befunde, fehler
	}

	// **Der Zeitpunkt ist der des Abrufs, nicht der des Eintrags.** Kam die
	// Antwort aus dem Zwischenspeicher, ist sie so alt wie ihr Alter sagt —
	// und ein Preisstand, der sich frischer gäbe, als er ist, wäre eine
	// Behauptung.
	erhoben := p.jetzt().Add(-alter)

	staende = make([]domain.Preisstand, 0, len(preise))

	// **Eine ungeprüfte Reihe ist kein Wegfall und kein Wert** (F04, #2246):
	// die Quelle nannte sie, nur nicht lesbar. Sie schützt den letzten
	// bekannten Stand vor dem Widerruf und wird selbst nicht abgelegt — sonst
	// stünde sie als neueste da und verdeckte den Wert, der weiter gilt.
	var ungeprueft []domain.Preisstand

	// **Der Zeitpunkt kommt von hier, alles andere von der Quelle** (A3,
	// #1442): wann gefragt wurde, weiß der Dienst; **was** die Antwort
	// bedeutet, weiß das Paket der Quelle.
	for _, preis := range preise {
		preis.Erhoben = erhoben

		if preis.Lage == domain.PreisUngeprueft {
			ungeprueft = append(ungeprueft, preis)

			continue
		}

		staende = append(staende, preis)
	}

	// **Und was die Quelle nicht mehr nennt, wird widerrufen** (#1201).
	//
	// Ohne diesen Zug blieb der letzte bekannte Wert der „neueste": ein
	// erfolgreicher Abruf konnte die alte Aussage nicht zurücknehmen, und
	// „der letzte gute Stand bleibt bei Netzausfall" war nicht von „die
	// Quelle sagt heute keinen Preis" zu unterscheiden (§16, Sätze 5 und 6).
	//
	// **Nur nach einem geglückten Abruf.** Der gescheiterte ist oben schon
	// zurückgekehrt.
	//
	// **Und nur für die Quelle, die geantwortet hat** (A3, #1442) — siehe
	// [Preispflege.wegfaelle].
	wegfaelle, wegfallfehler := p.wegfaelle(
		ctx, printing, p.quelle.Name(), slices.Concat(staende, ungeprueft), erhoben)
	if wegfallfehler != nil {
		return nil, 0, befunde, wegfallfehler
	}

	// **Alle oder keiner.** Ein halber Abruf wäre ein Stand, den es so nie
	// gab — und niemand könnte ihn als halb erkennen.
	if err := p.ablage.Eintragen(ctx, append(staende, wegfaelle...)...); err != nil {
		return nil, 0, befunde, err
	}

	return staende, alter, befunde, nil
}

// Neueste gibt den jüngsten Stand je Reihe eines Drucks — ohne zu fragen.
func (p *Preispflege) Neueste(ctx context.Context, printing string) (
	[]domain.Preisstand, error,
) {
	return p.ablage.Neueste(ctx, printing)
}

// Verlauf gibt eine Reihe in der Zeit.
//
// **Die Reihe wird genannt und nicht erraten** (A3, #1442): wer den Verlauf
// einer Karte sehen will, sagt, welchen — den Verkaufspreis bei Cardmarket in
// Euro für die Folienausführung, und nicht „den Preis".
func (p *Preispflege) Verlauf(
	ctx context.Context, printing string, reihe domain.Preisstand,
) ([]domain.Preisstand, error) {
	return p.ablage.Verlauf(ctx, printing, reihe)
}

// wegfaelle gibt einen Vermerk je Reihe, die **diese Quelle** nicht mehr nennt.
//
// # Was eine Reihe ist
//
// Quelle, Markt, Preisart, Währung, Ausführung, Zustand und Sprache
// ([domain.Preisstand.Reihe]) — dieselbe Einteilung, nach der `Neueste`
// gruppiert.
//
// # Warum nur die eigenen
//
// **Ein Wegfall bei einer Quelle widerruft keine Reihe einer anderen** (A3,
// #1442). **Die Quelle ist dabei eine Familie**: MTGJSON trägt die Preise von
// fünf Händlern, und was es für einen Druck nicht mehr nennt, ist bei **allen**
// weggefallen ([domain.Preisquelle.GehoertZu]). Sonst ließe ein Abruf bei
// Scryfall die Reihen von MTGJSON verschwinden — mit einem Wert von null, der
// nichts bedeutet, und einem Fehler, den niemand sieht.
//
// **Welche Quelle geantwortet hat, steht in der Antwort.** Eine leere Antwort
// nennt keine, und dann ist auch nichts zu widerrufen: „die Quelle hat
// geantwortet und nennt nichts" ist erst dann eine Aussage, wenn feststeht,
// **welche** Quelle das war — deshalb bekommt [Preispflege.Holen] sie als
// Argument.
//
// # Warum der Zeitpunkt derselbe ist
//
// Der Wegfall ist **Teil derselben Antwort** wie die Preise daneben. Ein
// späterer Zeitpunkt behauptete einen zweiten Abruf, den es nicht gab.
//
// # Was nicht vermerkt wird
//
// **Eine Reihe, die schon nicht mehr genannt ist.** Sonst wüchse der Verlauf
// bei jedem Abruf um eine Zeile, die dasselbe sagt wie die davor — und der
// Verlauf ist die Ansicht, in der jemand eine Bewegung sucht.
func (p *Preispflege) wegfaelle(
	ctx context.Context, printing string, quelle domain.Preisquelle,
	frisch []domain.Preisstand, erhoben time.Time,
) ([]domain.Preisstand, error) {
	bekannt, err := p.ablage.Neueste(ctx, printing)
	if err != nil {
		return nil, err
	}

	genannt := make(map[string]bool, len(frisch))
	for i := range frisch {
		genannt[frisch[i].Reihe()] = true
	}

	var vermerke []domain.Preisstand

	for i := range bekannt {
		// **Über die Familie und nicht über Gleichheit** (A3, #1442): MTGJSON
		// ist ein Sammler, und seine Stände heißen `mtgjson:cardkingdom`. Ein
		// Vergleich auf Gleichheit fände sie ungleich zu `mtgjson` und
		// widerriefe nichts — die Zusage stand im Kommentar und nicht im Code.
		if !bekannt[i].Quelle.GehoertZu(quelle) ||
			genannt[bekannt[i].Reihe()] || bekannt[i].Lage != domain.PreisGenannt {
			continue
		}

		widerruf := bekannt[i]
		widerruf.ID = 0
		widerruf.Lage = domain.PreisWeggefallen
		widerruf.Wert = 0
		widerruf.Beobachtet = time.Time{}
		widerruf.Erhoben = erhoben

		vermerke = append(vermerke, widerruf)
	}

	return vermerke, nil
}
