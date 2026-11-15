package app

// **Die letzten Fehlerwege der KI** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Warum im Paket
//
// **Weil einige dieser Wege von außen gar nicht zu erreichen sind**: der
// gemerkte Nahtfehler, das Budget für ein Cloudmodell und die zwei
// Nullprüfungen des Schreibwegs. Sie fangen den **nächsten** Aufrufer ab —
// und ein Zweig, den niemand betreten hat, ist eine Behauptung (§10).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Cloudmodell ohne Journal und ohne Freigabe bekommt kein Budget.**
//
// **Zwei Gründe, und beide halten für sich.** Ohne Journal gibt es nichts zu
// buchen; ohne Nachweis fehlt die Freigabe eines Menschen, und
// `Kostenwache.Reservieren` verlangt sie ausdrücklich. Das Issue sagt es von
// der anderen Seite: „Unter Local-only, fehlendem Scope oder **Budget**
// entstehen null Anbieteranfragen."
func TestEinCloudmodellOhneBudgetBekommtKeines(t *testing.T) {
	t.Parallel()

	ohneJournal := kibudget{ort: domain.KIOrtCloud, anbieter: "anthropic"}

	if _, fehler := ohneJournal.Reservieren(
		t.Context(), "ki-4-1"); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	ohneNachweis := kibudget{
		ort: domain.KIOrtCloud, anbieter: "anthropic", wache: stummesJournal{},
	}

	if _, fehler := ohneNachweis.Reservieren(
		t.Context(), "ki-4-1"); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	// **Mit beidem geht es** — die Gegenprobe: ohne sie sagte der Lauf nur,
	// dass etwas abgewiesen wird.
	mitBeidem := kibudget{
		ort: domain.KIOrtCloud, anbieter: "anthropic", wache: stummesJournal{},
		nachweis: "der Mensch hat zugestimmt",
		grenze:   domain.Betrag{Wert: 100, Waehrung: "EUR"},
	}

	posten, fehler := mitBeidem.Reservieren(t.Context(), "ki-4-1")
	if fehler != nil {
		t.Fatalf("mit Freigabe kam kein Budget: %v", fehler)
	}

	if posten != "ki-4-1" {
		t.Errorf("der Posten ist %q", posten)
	}

	// **Und ein lokales bekommt eines ohne alles.**
	lokal, fehler := kibudget{ort: domain.KIOrtLokal}.Reservieren(
		t.Context(), "ki-4-1")
	if fehler != nil {
		t.Fatalf("das lokale Modell bekam kein Budget: %v", fehler)
	}

	if lokal != KostenpostenOhneKosten {
		t.Errorf("der Posten ist %q", lokal)
	}
}

// **Ohne Schreibweg wird nichts gesetzt — und die Meldung sagt es.**
func TestOhneSchreibwegWirdNichtsGesetzt(t *testing.T) {
	t.Parallel()

	karten := kikarten{app: NeueApp("")}

	if fehler := karten.FelderSetzen(riegelkontext(t), 4,
		map[string][]string{"name": {"x"}}, nil, ""); fehler == nil {
		t.Fatal("ohne Schreibweg wurde gesetzt")
	}

	if _, fehler := karten.Revision(riegelkontext(t), 4); fehler == nil {
		t.Fatal("ohne Leser kam eine Revision")
	}
}

// **Und eine Karte, die sich nicht lesen lässt, wird nicht geschrieben.**
func TestEineUnlesbareKarteWirdNichtGeschrieben(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEditor(stummerSchreiber{}, stummerLeser{}, nil, nil)

	karten := kikarten{app: anwendung}

	if fehler := karten.FelderSetzen(riegelkontext(t), 4,
		map[string][]string{"name": {"x"}}, nil, ""); fehler == nil {
		t.Fatal("die unlesbare Karte wurde geschrieben")
	}
}

// **Der gemerkte Nahtfehler kommt vor „kein Bestand offen".**
func TestDerGemerkteNahtfehlerKommtZuerst(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.kinahtfehler = errors.New("der Schreibweg ließ sich nicht bauen")

	if _, _, fehler := anwendung.kiwerkzeug(riegelkontext(t), ollamaanbieter, kideckung{}); fehler == nil ||
		fehler.Error() != "der Schreibweg ließ sich nicht bauen" {
		t.Fatalf("der Fehler ist %v", fehler)
	}
}

// **Ohne Bauplan gibt es kein Werkzeug.**
//
// **Auch mit eingeschaltetem Schalter**: der Schalter sagt, dass gefragt
// werden **darf**, nicht dass eine Naht dasteht.
func TestOhneBauplanGibtEsKeinWerkzeug(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(riegelkontext(t),
		domain.KILaufzeit, domain.KILaufzeitAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	anwendung := KIAnwendungUeber(db).MitKIZuhause(t.TempDir())
	anwendung.MitKILaufzeit(nil)

	if _, _, fehler := anwendung.kiwerkzeug(riegelkontext(t), ollamaanbieter, kideckung{}); fehler == nil {
		t.Fatal("ohne Bauplan kam ein Werkzeug")
	}
}

// **Ein Katalog, der sich nicht bauen lässt, kommt als Fehler.**
//
// **Zwei Modelle mit demselben Alias** sind ein Konflikt — und der gehört
// gemeldet, nicht stillschweigend auf eines reduziert.
func TestEinUnbaubarerKatalogKommtAlsFehler(t *testing.T) {
	t.Parallel()

	db := bestandFuerDenRiegel(t)

	if fehler := storage.NeueEinstellungen(db).Setzen(riegelkontext(t),
		domain.KILaufzeit, domain.KILaufzeitAn); fehler != nil {
		t.Fatalf("einschalten: %v", fehler)
	}

	doppelt := zweimalDasselbeModell()

	anwendung := KIAnwendungUeber(db).MitKIZuhause(t.TempDir()).
		MitKILaufzeit(func(string) KILaufzeitquelle {
			return festeLaufzeit{modelle: doppelt}
		})

	if _, _, fehler := anwendung.kiwerkzeug(riegelkontext(t), ollamaanbieter, kideckung{}); fehler == nil {
		t.Fatal("der doppelte Alias kam durch")
	}
}

// zweimalDasselbeModell gibt zwei Modelle mit demselben Alias.
func zweimalDasselbeModell() []domain.KIModell {
	eines := domain.KIModell{
		Alias: "gemma4:8b", Anbieter: "ollama", Laufzeit: "ollama 0.12.3",
		Ort: domain.KIOrtLokal,
		Geprueft: []domain.KIGeprueft{{
			Faehigkeit: domain.KITextgenerierung, Datum: "2026-09-11",
			Beleg: "Ollama meldet completion",
		}},
	}

	return []domain.KIModell{eines, eines}
}

var _ service.KIKarten = kikarten{}

// **Die Schemaquelle weist ohne Naht ab — und meldet einen Lesefehler.**
//
// **Von innen geprüft**, weil ein Ruf von außen vorher am Auftragswerk
// scheitert: es liest die Revision der Karte, bevor es das Schema holt.
func TestDieSchemaquelleWeistOhneNahtAb(t *testing.T) {
	t.Parallel()

	if _, fehler := (kischemaquelle{app: NeueApp("")}).Schema(
		riegelkontext(t), 4); fehler == nil {
		t.Fatal("ohne Naht kam ein Schema")
	}

	anwendung := NeueApp("")
	anwendung.MitEditor(stummerSchreiber{}, stummerLeser{}, nil, nil)
	anwendung.MitStilwahl(nil)

	if _, fehler := (kischemaquelle{app: anwendung}).Schema(
		riegelkontext(t), 4); fehler == nil {
		t.Fatal("ohne Stilwahl kam ein Schema")
	}

	anwendung.MitStilwahl(leererStilwaehler{})

	if _, fehler := (kischemaquelle{app: anwendung}).Schema(
		riegelkontext(t), 4); fehler == nil {
		t.Fatal("die unlesbare Karte gab ein Schema")
	}
}

// leererStilwaehler nennt nie ein Feld.
type leererStilwaehler struct{ Stilwaehler }

func (leererStilwaehler) Kartenfelder(context.Context, int64) (
	[]domain.Field, error,
) {
	return nil, nil
}

// **Der Riegel weist ab, wenn die Richtlinie nach dem Schalter scheitert.**
//
// **Zwei Lesezugriffe, und der zweite geht schief**: `pruefeKILaufzeit` liest
// erst Schalter und Adresse, dann die Richtlinie. Ein Doppel, das beim ersten
// scheitert, erreicht den zweiten Weg nie.
func TestDerRiegelWeistBeiUnlesbarerRichtlinieNachDemSchalterAb(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&eingeschalteteDannStolpernde{})

	if fehler := anwendung.pruefeKILaufzeit(
		riegelkontext(t), ""); fehler == nil {
		t.Fatal("der Riegel ließ durch")
	}
}

// eingeschalteteDannStolpernde melden den Schalter und scheitern danach.
type eingeschalteteDannStolpernde struct{ gelesen int }

func (e *eingeschalteteDannStolpernde) Stand(context.Context) (
	map[string]service.Wahl, error,
) {
	e.gelesen++

	if e.gelesen > 1 {
		return nil, errors.New("die Richtlinie ist unlesbar")
	}

	return map[string]service.Wahl{
		domain.KILaufzeit: {Wirksam: domain.KILaufzeitAn},
	}, nil
}

func (e *eingeschalteteDannStolpernde) Alle(context.Context) (
	map[string]string, error,
) {
	return map[string]string{}, nil
}

func (e *eingeschalteteDannStolpernde) Setzen(
	context.Context, string, string,
) error {
	return nil
}

// **Und das Werkzeug ebenso, wenn die Laufzeitfrage danach scheitert.**
func TestDasWerkzeugMeldetEinenSpaetenLesefehler(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&halbstolpernde{gutBis: 1})

	if _, _, fehler := anwendung.kiwerkzeug(riegelkontext(t), ollamaanbieter, kideckung{}); fehler == nil {
		t.Fatal("der späte Lesefehler kam nicht durch")
	}
}

// **Ohne Bauplan, aber mit Schalter, gibt es kein Werkzeug.**
//
// **Nicht dasselbe wie »der Schalter ist aus«**: hier darf gefragt werden,
// und es steht keine Naht da.
func TestMitSchalterOhneBauplanGibtEsKeinWerkzeug(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&nurEingeschaltete{})
	anwendung.MitKILaufzeit(func(string) KILaufzeitquelle {
		return festeLaufzeit{}
	})

	// **Der Bauplan wird nach der Frage entfernt** — `frageAnbieter` hat ihn
	// gebraucht, `kiwerkzeug` findet ihn nicht mehr. Genau diesen Zustand
	// fängt die Prüfung ab.
	werk, _, fehler := anwendung.kiwerkzeug(riegelkontext(t), ollamaanbieter, kideckung{})
	if fehler != nil {
		t.Fatalf("mit Bauplan kam kein Werkzeug: %v", fehler)
	}

	if werk == nil {
		t.Fatal("das Werkzeug ist leer")
	}

	// **Der Bauplan verschwindet, der Schalter bleibt** — dann kommt
	// `frageAnbieter` mit `aus` zurück, und `kiwerkzeug` sagt es.
	anwendung.kilaufzeit = nil

	if _, _, fehler := anwendung.kiwerkzeug(riegelkontext(t), ollamaanbieter, kideckung{}); fehler == nil {
		t.Fatal("ohne Bauplan kam ein Werkzeug")
	}

	// **Und wenn der Stand trotzdem »bereit« meldet**, hält die zweite
	// Nullprüfung: sie fängt den Fall ab, dass jemand die Naht zwischen
	// beiden Rufen entfernt.
	anwendung.MitEinstellungen(&nurEingeschaltete{})
	anwendung.MitKILaufzeit(func(string) KILaufzeitquelle {
		return festeLaufzeit{modelle: einModellFuerDenBauplan()}
	})

	staende, _, standfehler := anwendung.frageAnbieter(riegelkontext(t))
	if standfehler != nil {
		t.Fatalf("frageAnbieter: %v", standfehler)
	}

	anwendung.kilaufzeit = nil

	if _, _, fehler := anwendung.werkzeugAus(
		riegelkontext(t), staende, ollamaanbieter, kideckung{}); fehler == nil {
		t.Fatal("ohne Bauplan kam ein Werkzeug aus fertigen Ständen")
	}
}

// nurEingeschaltete melden den Schalter und sonst nichts.
type nurEingeschaltete struct{}

func (nurEingeschaltete) Stand(context.Context) (map[string]service.Wahl, error) {
	return map[string]service.Wahl{
		domain.KILaufzeit: {Wirksam: domain.KILaufzeitAn},
	}, nil
}

func (nurEingeschaltete) Alle(context.Context) (map[string]string, error) {
	return map[string]string{}, nil
}

func (nurEingeschaltete) Setzen(context.Context, string, string) error { return nil }

// GruppeVonAussenMitHerkunft gibt es an diesem Doppel **nur der
// Schnittstelle wegen** (A13.1b, #1452).
//
// **Sie scheitert laut.** Ein Doppel, das hier still gutginge, ließe einen
// Lauf über Gruppen grün werden, der gar keine geschrieben hat — §3: eine
// Vorrichtung, die großzügiger ist als der Betrieb.
func (stummerSchreiber) GruppeVonAussenMitHerkunft(
	context.Context, []domain.Kartenschritt,
) error {
	return errors.New("dieses Doppel schreibt keine Gruppen")
}
