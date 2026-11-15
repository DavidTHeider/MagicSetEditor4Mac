package app_test

// **Das Kostenjournal externer Aufträge** — Werkstück O12 von Phase 5.2
// (#1439).

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kostendoppel spielt den Dienst.
type kostendoppel struct {
	stand  domain.Haushaltsstand
	posten []domain.Kostenposten
	fehler error

	konto    string
	zeitraum domain.Zeitraum
	grenze   domain.Betrag
}

func (k *kostendoppel) Stand(_ context.Context, konto string, zeitraum domain.Zeitraum) (
	domain.Haushaltsstand, error,
) {
	k.konto, k.zeitraum = konto, zeitraum

	return k.stand, k.fehler
}

func (k *kostendoppel) Posten(_ context.Context, konto string, zeitraum domain.Zeitraum) (
	[]domain.Kostenposten, error,
) {
	k.konto, k.zeitraum = konto, zeitraum

	return k.posten, k.fehler
}

func (k *kostendoppel) BudgetSetzen(
	_ context.Context, konto string, zeitraum domain.Zeitraum, betrag domain.Betrag,
) error {
	k.konto, k.zeitraum, k.grenze = konto, zeitraum, betrag

	return k.fehler
}

func (k *kostendoppel) Reservieren(
	_ context.Context, _ service.Kostenauftrag, _ string,
) (domain.Kostenposten, error) {
	return domain.Kostenposten{}, k.fehler
}

func (k *kostendoppel) Abschliessen(_ context.Context, _ string, _ domain.Betrag) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, k.fehler
}

func (k *kostendoppel) Freigeben(_ context.Context, _ string) (domain.Kostenposten, error) {
	return domain.Kostenposten{}, k.fehler
}

func (k *kostendoppel) AlsUnklarMerken(_ context.Context, _, _ string) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, k.fehler
}

func kostenapp(doppel *kostendoppel) *app.App {
	anwendung := app.NeueApp("/nirgends/bestand.db")
	anwendung.MitKosten(doppel)

	return anwendung
}

func geld(wert int64) domain.Betrag { return domain.Betrag{Wert: wert, Waehrung: "EUR"} }

// **Die vier Zahlen kommen vollständig durch** — samt der Grenze und dem
// Zeitraum, über den sie gelten.
func TestDieVierZahlenKommenVollstaendigDurch(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{stand: domain.Haushaltsstand{
		Konto: "anbieter", Zeitraum: domain.ZeitraumMonat,
		Grenze: geld(50_000), Reserviert: geld(3_000),
		Verbraucht: geld(12_000), Unklar: geld(500), Verfuegbar: geld(34_500),
	}}

	antwort := kostenapp(doppel).Kostenstand(t.Context(), "anbieter", "monat")
	if antwort.Fehler != nil {
		t.Fatalf("der Stand: %+v", antwort.Fehler)
	}

	if doppel.konto != "anbieter" || doppel.zeitraum != domain.ZeitraumMonat {
		t.Errorf("gefragt wurde nach %q über %q", doppel.konto, doppel.zeitraum)
	}

	if antwort.Verfuegbar.Wert != 34_500 || antwort.Verfuegbar.Waehrung != "EUR" {
		t.Errorf("verfügbar sind %+v", antwort.Verfuegbar)
	}

	if antwort.Unklar.Wert != 500 || antwort.Grenze.Wert != 50_000 {
		t.Errorf("die Antwort ist %+v", antwort)
	}

	if antwort.Zeitraum != "monat" {
		t.Errorf("der Zeitraum ist %q", antwort.Zeitraum)
	}
}

// **Die Posten kommen ohne Prompt und ohne Zugangsdaten** — die Aufgabe ist
// eine Marke.
func TestDiePostenKommenAlsMarken(t *testing.T) {
	t.Parallel()

	begonnen := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)

	doppel := &kostendoppel{posten: []domain.Kostenposten{{
		Kennung: "eins", Konto: "anbieter", Aufgabe: "karte.text",
		Hoechstbetrag: geld(200), Betrag: geld(120),
		Stand: domain.KostenAbgerechnet, Begonnen: begonnen,
	}}}

	antwort := kostenapp(doppel).Kostenposten(t.Context(), "anbieter", "tag")
	if antwort.Fehler != nil {
		t.Fatalf("die Posten: %+v", antwort.Fehler)
	}

	if len(antwort.Posten) != 1 {
		t.Fatalf("es sind %d Posten", len(antwort.Posten))
	}

	einer := antwort.Posten[0]
	if einer.Stand != "abgerechnet" || einer.Aufgabe != "karte.text" ||
		einer.Betrag.Wert != 120 || einer.Begonnen == "" {
		t.Errorf("der Posten ist %+v", einer)
	}

	// **Ein Auftrag ohne Ende trägt keinen Zeitpunkt** — und keine erfundene
	// Null.
	if einer.Beendet != "" {
		t.Errorf("das Ende ist %q", einer.Beendet)
	}
}

// **Eine Grenze lässt sich setzen und wegnehmen.**
func TestEineGrenzeLaesstSichSetzenUndWegnehmen(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{}
	anwendung := kostenapp(doppel)

	if antwort := anwendung.KostenbudgetSetzen(
		t.Context(), "anbieter", "monat", 50_000, "EUR"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	if doppel.grenze.Wert != 50_000 || doppel.zeitraum != domain.ZeitraumMonat {
		t.Errorf("gesetzt wurde %+v über %q", doppel.grenze, doppel.zeitraum)
	}

	// **Null nimmt sie weg** — und dann geht kein Auftrag mehr hinaus.
	if antwort := anwendung.KostenbudgetSetzen(
		t.Context(), "anbieter", "monat", 0, "EUR"); antwort.Fehler != nil {
		t.Fatalf("wegnehmen: %+v", antwort.Fehler)
	}

	if doppel.grenze.Wert != 0 {
		t.Errorf("weggenommen wurde %+v", doppel.grenze)
	}
}

// **Jeder Fehler kommt als Fehler und nicht als leere Antwort** — eine leere
// Liste hieße „nichts gebucht".
func TestJederKostenfehlerKommtDurch(t *testing.T) {
	t.Parallel()

	doppel := &kostendoppel{fehler: errors.New("das Journal ist zu")}
	anwendung := kostenapp(doppel)

	stand := anwendung.Kostenstand(t.Context(), "anbieter", "tag")
	if stand.Fehler == nil {
		t.Error("der Stand verschluckte den Fehler")
	}

	posten := anwendung.Kostenposten(t.Context(), "anbieter", "tag")
	if posten.Fehler == nil || posten.Posten == nil {
		t.Errorf("die Posten sind %+v", posten)
	}

	if antwort := anwendung.KostenbudgetSetzen(
		t.Context(), "anbieter", "tag", 100, "EUR"); antwort.Fehler == nil {
		t.Error("das Setzen verschluckte den Fehler")
	}
}
