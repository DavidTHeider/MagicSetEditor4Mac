// **Ein weggefallener Preis nimmt den alten zurück** —
// #1201.
//
// # Der Befund
//
// `null` heißt nach eigenem Vertrag „kein Preis bekannt" und wird beim Lesen
// weggelassen; `Holen` schrieb nur die nichtleeren Reihen. `Neueste` wählt
// aber je Reihe den jüngsten **jemals** gespeicherten Stand.
//
// Wechselte eine Quelle erfolgreich von `"4.17"` auf `null`, blieb 4,17 der
// „neueste" Stand und ging weiter ins Portfolio ein. **Der erfolgreiche Abruf
// konnte die alte Aussage nicht widerrufen.**
//
// Damit war „der letzte gute Stand bleibt bei Netzausfall" nicht von „die
// Quelle sagt heute keinen Preis" zu unterscheiden — und die Zusage „kein
// Preis ist besser als ein falscher" (§16, Satz 5 und 6) an dieser Naht nicht
// eingelöst.
//
// # Was jetzt geschieht
//
// Ein geglückter Abruf, der eine bekannte Reihe **nicht** mehr nennt, schreibt
// dorthin einen **Wegfall**: einen Stand mit demselben Zeitpunkt, der sagt
// „hier nennt die Quelle nichts mehr". `Neueste` gibt ihn zurück, und wer
// rechnet, überspringt ihn.
//
// # Was ausdrücklich **nicht** geschieht
//
// **Kein Wegfall nach einem gescheiterten Abruf.** Dann weiß niemand etwas,
// und der letzte gute Stand bleibt — das ist §16, Satz 5, und es ist der
// Unterschied, um den es geht.
package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/scryfall"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// preisvorrichtung ist eine Pflege über einer Ablage, die **wie der Bestand**
// antwortet.
//
// # Warum ein zweites Doppel neben `preisdoppel`
//
// Jenes gibt bei `Neueste` **alle** Stände zurück, nicht den jüngsten je
// Reihe. Für die Läufe dort reicht das; für diesen hier wäre es eine
// Vorrichtung, die großzügiger ist als der Betrieb — und §3 nennt genau das
// als die Stelle, an der ein Test den falschen Weg prüft.
type preisvorrichtung struct {
	ablage *reihenablage
	quelle *steuerbarequelle
	pflege *service.Preispflege
	jetzt  time.Time
}

// reihenablage hält Stände und beantwortet `Neueste` je Reihe.
type reihenablage struct {
	staende []domain.Preisstand

	// lesefehler scheitert **nur** beim Lesen, eintragefehler **nur** beim
	// Schreiben. Getrennt, weil `Holen` seit #1201 erst liest und dann
	// schreibt: ein gemeinsamer Fehler erreichte den zweiten Zweig nie.
	lesefehler     error
	eintragefehler error
}

func (r *reihenablage) Eintragen(_ context.Context, staende ...domain.Preisstand) error {
	if r.eintragefehler != nil {
		return r.eintragefehler
	}

	r.staende = append(r.staende, staende...)

	return nil
}

func (r *reihenablage) Neueste(_ context.Context, printing string) (
	[]domain.Preisstand, error,
) {
	if r.lesefehler != nil {
		return nil, r.lesefehler
	}

	juengste := map[string]domain.Preisstand{}

	for _, stand := range r.staende {
		if stand.Printing != printing {
			continue
		}

		reihe := string(stand.Quelle) + "|" + stand.Waehrung + "|" + string(stand.Finish)
		if da, gefuehrt := juengste[reihe]; !gefuehrt || !stand.Erhoben.Before(da.Erhoben) {
			juengste[reihe] = stand
		}
	}

	gefunden := make([]domain.Preisstand, 0, len(juengste))
	for _, stand := range juengste {
		gefunden = append(gefunden, stand)
	}

	return gefunden, nil
}

// NeuesteAlle ist dasselbe über alle Drucke — die Auswahl je Reihe inbegriffen.
//
// **Über [reihenablage.Neueste] und nicht daneben** (#1292): diese Vorrichtung
// hält die Regel „je Quelle, Währung und Ausführung der jüngste", und sie an
// zwei Stellen zu schreiben hieße, sie an einer zu ändern.
func (r *reihenablage) NeuesteAlle(ctx context.Context) (
	[]domain.Preisstand, error,
) {
	gesehen := map[string]bool{}
	alle := []domain.Preisstand{}

	for _, stand := range r.staende {
		if gesehen[stand.Printing] {
			continue
		}

		gesehen[stand.Printing] = true

		juengste, err := r.Neueste(ctx, stand.Printing)
		if err != nil {
			return nil, err
		}

		alle = append(alle, juengste...)
	}

	// **Der Lesefehler gilt auch für einen leeren Bestand.** Ohne diese Zeile
	// käme er nur heraus, wenn zufällig ein Stand abgelegt ist — und der
	// Fehlerweg des Portfolios bliebe unbetreten (§10).
	if r.lesefehler != nil {
		return nil, r.lesefehler
	}

	return alle, nil
}

func (r *reihenablage) Verlauf(
	ctx context.Context, printing string, _ domain.Preisstand,
) ([]domain.Preisstand, error) {
	return r.Neueste(ctx, printing)
}

// steuerbarequelle gibt, was der Lauf ihr sagt.
type steuerbarequelle struct {
	preise []scryfall.Preis

	// ungeprueft sind Reihen, die die Quelle nannte und die sich nicht lesen
	// ließen (F04, #2246).
	ungeprueft []scryfall.Preis
	fehler     error
}

func (q *steuerbarequelle) PreiseFuerDruck(_ context.Context, kennung string) (
	[]domain.Preisstand, time.Duration, []domain.Befund, error,
) {
	staende := make([]domain.Preisstand, 0, len(q.preise)+len(q.ungeprueft))
	for _, preis := range q.preise {
		staende = append(staende, preis.AlsStand(kennung))
	}

	for _, preis := range q.ungeprueft {
		stand := preis.AlsStand(kennung)
		stand.Lage, stand.Wert = domain.PreisUngeprueft, 0
		staende = append(staende, stand)
	}

	return staende, 0, nil, q.fehler
}

// Name gibt die Quelle, in deren Namen dieses Doppel antwortet.
func (*steuerbarequelle) Name() domain.Preisquelle { return domain.PreisquelleScryfall }

func neuePreisvorrichtung(t *testing.T) *preisvorrichtung {
	t.Helper()

	vor := &preisvorrichtung{
		ablage: &reihenablage{},
		quelle: &steuerbarequelle{},
		jetzt:  time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC),
	}

	vor.pflege = service.NeuePreispflege(vor.ablage, nil, vor.quelle,
		func() time.Time { return vor.jetzt })

	return vor
}

// quelleGibt lässt die Quelle **eine** Reihe nennen.
//
// **Immer `eur`/`nonfoil`**: welche Reihe es ist, entscheidet hier nichts —
// gefragt ist, ob ein Wegfall vermerkt wird. Ein Parameter dafür wäre eine
// Wahl, die es nicht gibt.
func (v *preisvorrichtung) quelleGibt(betrag int64) {
	v.quelle.preise = []scryfall.Preis{{
		Waehrung: "eur", Finish: domain.FinishNonfoil, Betrag: betrag,
	}}
	v.quelle.fehler = nil
}

func (v *preisvorrichtung) quelleGibtNichts() {
	v.quelle.preise = nil
	v.quelle.fehler = nil
}

func (v *preisvorrichtung) quelleScheitert() {
	v.quelle.preise = nil
	v.quelle.fehler = errors.New("das Netz schweigt")
}

// mussHolen holt und lässt den Lauf scheitern, wenn es nicht geht.
func mussHolen(t *testing.T, vor *preisvorrichtung) {
	t.Helper()

	if _, _, _, fehler := vor.pflege.Holen(t.Context(), "druck-1"); fehler != nil {
		t.Fatalf("holen: %v", fehler)
	}
}

func TestEinWeggefallenerPreisWirdVermerkt(t *testing.T) {
	t.Parallel()

	vor := neuePreisvorrichtung(t)

	vor.quelleGibt(417)

	if _, _, _, fehler := vor.pflege.Holen(t.Context(), "druck-1"); fehler != nil {
		t.Fatalf("erster Abruf: %v", fehler)
	}

	// **Der zweite Abruf gelingt und nennt die Reihe nicht mehr.**
	vor.quelleGibtNichts()

	if _, _, _, fehler := vor.pflege.Holen(t.Context(), "druck-1"); fehler != nil {
		t.Fatalf("zweiter Abruf: %v", fehler)
	}

	staende, err := vor.pflege.Neueste(t.Context(), "druck-1")
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	if len(staende) != 1 {
		t.Fatalf("%d Stände, erwartet einen", len(staende))
	}

	if staende[0].Lage != domain.PreisWeggefallen {
		t.Errorf("der neueste Stand ist %d und gilt weiter — der erfolgreiche "+
			"Abruf konnte die alte Aussage nicht widerrufen", staende[0].Wert)
	}
}

// **Ein gescheiterter Abruf vermerkt nichts.**
//
// Dann weiß niemand etwas, und der letzte gute Stand bleibt (§16, Satz 5).
func TestEinGescheiterterAbrufVermerktKeinenWegfall(t *testing.T) {
	t.Parallel()

	vor := neuePreisvorrichtung(t)
	vor.quelleGibt(417)

	if _, _, _, fehler := vor.pflege.Holen(t.Context(), "druck-1"); fehler != nil {
		t.Fatalf("erster Abruf: %v", fehler)
	}

	vor.quelleScheitert()

	if _, _, _, fehler := vor.pflege.Holen(t.Context(), "druck-1"); fehler == nil {
		t.Fatal("der Abruf ist nicht gescheitert")
	}

	staende, err := vor.pflege.Neueste(t.Context(), "druck-1")
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	if len(staende) != 1 || staende[0].Lage == domain.PreisWeggefallen {
		t.Errorf("Stände %+v — nach einem gescheiterten Abruf bleibt der "+
			"letzte gute Stand", staende)
	}
}

// **Und eine Reihe, die wiederkommt, gilt wieder.**
func TestEineWiederkehrendeReiheGiltWieder(t *testing.T) {
	t.Parallel()

	vor := neuePreisvorrichtung(t)

	vor.quelleGibt(417)
	mussHolen(t, vor)

	vor.quelleGibtNichts()
	mussHolen(t, vor)

	vor.jetzt = vor.jetzt.Add(time.Hour)
	vor.quelleGibt(500)
	mussHolen(t, vor)

	staende, err := vor.pflege.Neueste(t.Context(), "druck-1")
	if err != nil {
		t.Fatalf("Neueste: %v", err)
	}

	if len(staende) != 1 || staende[0].Lage == domain.PreisWeggefallen || staende[0].Wert != 500 {
		t.Errorf("Stände %+v, erwartet einen mit 500", staende)
	}
}

// **Ein weggefallener Stand zählt nicht ins Portfolio.**
//
// Sein Wert ist null und bedeutet nichts. Ihn mitzurechnen hieße, eine Karte
// für wertlos zu erklären, deren Preis niemand kennt.
func TestEinWeggefallenerStandZaehltNichtInsPortfolio(t *testing.T) {
	t.Parallel()

	if service.ZaehltFuerPortfolio(domain.Preisstand{
		Waehrung: "eur", Finish: domain.FinishNonfoil, Lage: domain.PreisWeggefallen,
	}, &domain.Exemplar{Menge: 1}) {
		t.Error("ein weggefallener Stand wurde mitgerechnet")
	}
}

// **Ein Lesefehler beim Suchen der Wegfälle kommt durch.**
//
// Er steht **vor** dem Schreiben: wer die bekannten Reihen nicht lesen kann,
// weiß auch nicht, welche fehlen — und schriebe sonst Preise ohne ihren
// Widerruf.
func TestEinLesefehlerBeimWegfallsuchenKommtDurch(t *testing.T) {
	t.Parallel()

	vor := neuePreisvorrichtung(t)
	vor.quelleGibt(417)
	vor.ablage.lesefehler = errors.New("der Bestand schweigt")

	if _, _, _, fehler := vor.pflege.Holen(t.Context(), "druck-1"); !errors.Is(
		fehler, vor.ablage.lesefehler,
	) {
		t.Errorf("Fehler %v", fehler)
	}
}

// **Ein Schreibfehler kommt durch.**
//
// Er steht **nach** dem Suchen der Wegfälle; ein gemeinsamer Fehler für beides
// erreichte diesen Zweig nie, weil der erste zuerst greift.
func TestEinSchreibfehlerBeimPreisholenKommtDurch(t *testing.T) {
	t.Parallel()

	vor := neuePreisvorrichtung(t)
	vor.quelleGibt(417)
	vor.ablage.eintragefehler = errors.New("der Bestand nimmt nichts an")

	if _, _, _, fehler := vor.pflege.Holen(t.Context(), "druck-1"); !errors.Is(
		fehler, vor.ablage.eintragefehler,
	) {
		t.Errorf("Fehler %v", fehler)
	}
}
