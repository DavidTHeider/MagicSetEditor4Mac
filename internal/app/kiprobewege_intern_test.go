package app

// **Die Fehlerwege der Messung** — Werkstück A9 von Phase 5.3
// (#1448).
//
// # Warum im Paket
//
// **Weil einige dieser Wege von außen gar nicht zu erreichen sind**: ein
// Kontext ohne Profil, eine Ablage ohne Bestand, eine Quelle, die beim
// Vorschlagen scheitert. Sie fangen den **nächsten** Aufrufer ab — und ein
// Zweig, den niemand betreten hat, ist eine Behauptung (§10).

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ohne Profil im Kontext wird nicht gemessen.**
func TestOhneProfilWirdNichtGemessen(t *testing.T) {
	t.Parallel()

	anwendung := KIAnwendungUeber(openaiAn(t)).MitMarktbund(erfundenerBund{})

	if antwort := anwendung.KIModellPruefen(
		t.Context(), "openai", "gpt-probe", ""); antwort.Fehler == nil {
		t.Error("ohne Profil wurde gemessen")
	}

	if antwort := anwendung.KIModellVergessen(
		t.Context(), "openai", "gpt-probe"); antwort.Fehler == nil {
		t.Error("ohne Profil wurde vergessen")
	}
}

// **Ohne Bestand gibt es keine Ablage.**
//
// **Und beide Wege sagen es**: messen und vergessen brauchen dieselbe
// Ablage, und ohne sie ist beides kein leeres Ergebnis, sondern ein Fehler.
func TestOhneBestandGibtEsKeineProbenablage(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")

	if _, fehler := anwendung.probenablage(); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("der Fehler ist %v", fehler)
	}

	if antwort := anwendung.KIModellVergessen(
		riegelkontext(t), "openai", "gpt-probe"); antwort.Fehler == nil {
		t.Error("ohne Bestand wurde vergessen")
	}
}

// **Ein unbekannter Anbieter wird abgewiesen, bevor etwas hinausgeht.**
func TestEinUnbekannterAnbieterWirdNichtGemessen(t *testing.T) {
	t.Parallel()

	anwendung := KIAnwendungUeber(openaiAn(t)).MitMarktbund(erfundenerBund{})

	if antwort := anwendung.KIModellPruefen(riegelkontext(t),
		"gibtesnicht", "x", ""); antwort.Fehler == nil {
		t.Error("der unbekannte Anbieter kam durch")
	}
}

// **Ein unlesbarer Stand hält die Messung auf.**
func TestEinUnlesbarerStandHaeltDieMessungAuf(t *testing.T) {
	t.Parallel()

	anwendung := NeueApp("")
	anwendung.MitEinstellungen(&halbstolpernde{gutBis: 0})

	if antwort := anwendung.KIModellPruefen(riegelkontext(t),
		"openai", "gpt-probe", ""); antwort.Fehler == nil {
		t.Error("der Lesefehler kam nicht durch")
	}
}

// **Ein Fehler der Quelle kommt durch die Hülle.**
//
// **Und wird nicht zu einer leeren Antwort**: `probequelle` ist eine Hülle
// und keine zweite Fachlogik.
func TestEinFehlerDerQuelleKommtDurchDieHuelle(t *testing.T) {
	t.Parallel()

	huelle := probequelle{quelle: stolperndeQuelle{}}

	if _, fehler := huelle.Vorschlagen(t.Context(), "m", "x", nil); fehler == nil {
		t.Fatal("der Fehler der Quelle kam nicht durch")
	}

	if _, fehler := huelle.Fassung(t.Context()); fehler == nil {
		t.Error("der Fehler der Fassung kam nicht durch")
	}
}

// stolperndeQuelle scheitert an allem.
type stolperndeQuelle struct{}

func (stolperndeQuelle) Fassung(context.Context) (string, error) {
	return "", errors.New("nichts da")
}

func (stolperndeQuelle) Modelle(context.Context) (
	[]domain.KIModell, []domain.Befund, error,
) {
	return nil, nil, errors.New("nichts da")
}

func (stolperndeQuelle) Vorschlagen(
	context.Context, string, string, string, map[string]any,
) (KIRohantwort, error) {
	return KIRohantwort{}, errors.New("nichts da")
}

// **Eine geglückte Hülle reicht durch, was sie bekommt.**
//
// **Die Gegenprobe zur Zeile darüber**: ohne sie sagte der Lauf nur, dass
// ein Fehler durchkommt — nicht, dass sonst etwas Richtiges ankommt.
func TestEineGeglueckteHuelleReichtDurch(t *testing.T) {
	t.Parallel()

	huelle := probequelle{quelle: redselig{antwort: KIRohantwort{
		Inhalt: `{"probe":"bereit"}`, Modell: "m", Abgeschnitten: true,
	}}}

	antwort, fehler := huelle.Vorschlagen(t.Context(), "m", "x", nil)
	if fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	if antwort != (service.KIProbeantwort{
		Inhalt: `{"probe":"bereit"}`, Modell: "m", Abgeschnitten: true,
	}) {
		t.Errorf("die Antwort ist %+v", antwort)
	}
}

// **Ohne Profil im Kontext gibt es keine gemessenen Fähigkeiten.**
//
// **Und das ist kein Fehler, sondern eine engere Auskunft** (§16, Satz 5):
// ohne Messungen steht im Katalog, was die Anbieter selbst sagen.
func TestOhneProfilGibtEsKeineGemessenenFaehigkeiten(t *testing.T) {
	t.Parallel()

	if proben := KIAnwendungUeber(openaiAn(t)).
		gemesseneFaehigkeiten(t.Context()); proben != nil {
		t.Errorf("ohne Profil kamen %+v", proben)
	}

	if proben := NeueApp("").
		gemesseneFaehigkeiten(riegelkontext(t)); proben != nil {
		t.Errorf("ohne Bestand kamen %+v", proben)
	}
}

// **Und eine unlesbare Ablage leert den Katalog nicht.**
func TestEineUnlesbareAblageLeertDenKatalogNicht(t *testing.T) {
	t.Parallel()

	db := openaiAn(t)

	anwendung := KIAnwendungUeber(db)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if proben := anwendung.gemesseneFaehigkeiten(riegelkontext(t)); proben != nil {
		t.Errorf("aus der geschlossenen Ablage kamen %+v", proben)
	}
}

// **Ein Befund zu einem Modell, das im Katalog steht, macht es nicht
// messbar.**
//
// **Und ein zweiter Befund zu demselben Modell auch nicht zweimal.** Ohne
// beide Zweige stünde ein Modell unter »nicht beschrieben«, über das etwas
// feststeht — oder zweimal.
func TestEinBefundZuEinemBekanntenModellMachtEsNichtMessbar(t *testing.T) {
	t.Parallel()

	modelle := []domain.KIModell{{Alias: "bekannt", Anbieter: "openai"}}

	befunde := []domain.Befund{
		{Datei: "bekannt", Meldung: "x"},
		{Datei: "neu", Meldung: "y"},
		{Datei: "neu", Meldung: "noch einmal"},
		{Meldung: "ohne Modell"},
	}

	messbar := messbareAus(befunde, modelle)

	if len(messbar) != 1 || messbar[0] != "neu" {
		t.Errorf("messbar sind %v", messbar)
	}
}

// redselig antwortet mit dem, was ihm mitgegeben wurde.
type redselig struct{ antwort KIRohantwort }

func (redselig) Fassung(context.Context) (string, error) { return "x", nil }

func (redselig) Modelle(context.Context) (
	[]domain.KIModell, []domain.Befund, error,
) {
	return nil, nil, nil
}

func (r redselig) Vorschlagen(
	context.Context, string, string, string, map[string]any,
) (KIRohantwort, error) {
	return r.antwort, nil
}

// **Ohne Ablage wird nicht gemessen — obwohl die Quelle stünde.**
//
// # Warum dieser Zweig eigens steht
//
// **Weil er zwischen zwei anderen liegt.** Die Quelle kommt aus den
// Einstellungen, die Ablage aus dem Bestand — und beides ist nicht dasselbe.
// Wer nur »ohne Bestand« prüfte, träfe den Zweig davor; wer nur »unbekannter
// Anbieter« prüfte, den danach.
func TestOhneAblageWirdNichtGemessen(t *testing.T) {
	t.Parallel()

	quelle := redselig{antwort: KIRohantwort{Inhalt: `{"probe":"bereit"}`}}

	anwendung := NeueApp("").MitMarktbund(erfundenerBund{})
	anwendung.MitEinstellungen(einschalter{})
	anwendung.MitKIAnbietern(
		func(string) KILaufzeitquelle { return quelle },
		map[string]func(adresse string) KILaufzeitquelle{
			openaianbieter: func(string) KILaufzeitquelle { return quelle },
		})
	anwendung.MitKosten(stummeswerk{})
	anwendung.MitKIFreigabe("freigabe", domain.Betrag{Wert: 1, Waehrung: "EUR"})

	if antwort := anwendung.KIModellPruefen(riegelkontext(t),
		openaianbieter, "gpt-probe", ""); antwort.Fehler == nil {
		t.Fatal("ohne Ablage wurde gemessen")
	}
}

// einschalter gibt jeden Dienst als eingeschaltet.
type einschalter struct{}

func (einschalter) Alle(context.Context) (map[string]string, error) {
	return map[string]string{}, nil
}

func (einschalter) Setzen(context.Context, string, string) error { return nil }

func (einschalter) Stand(context.Context) (map[string]service.Wahl, error) {
	stand := map[string]service.Wahl{}

	for _, dienst := range domain.Dienste() {
		stand[domain.DienstEinstellung(dienst)] = service.Wahl{
			Gewaehlt: domain.DienstAn, Wirksam: domain.DienstAn,
		}
	}

	return stand, nil
}

// stummeswerk bucht alles ohne Widerspruch.
type stummeswerk struct{}

func (stummeswerk) Reservieren(
	_ context.Context, auftrag service.Kostenauftrag, _ string,
) (domain.Kostenposten, error) {
	return domain.Kostenposten{Kennung: auftrag.Kennung}, nil
}

func (stummeswerk) Stand(context.Context, string, domain.Zeitraum) (
	domain.Haushaltsstand, error,
) {
	return domain.Haushaltsstand{}, nil
}

func (stummeswerk) Posten(context.Context, string, domain.Zeitraum) (
	[]domain.Kostenposten, error,
) {
	return nil, nil
}

func (stummeswerk) BudgetSetzen(
	context.Context, string, domain.Zeitraum, domain.Betrag,
) error {
	return nil
}

func (stummeswerk) Abschliessen(context.Context, string, domain.Betrag) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, nil
}

func (stummeswerk) Freigeben(context.Context, string) (domain.Kostenposten, error) {
	return domain.Kostenposten{}, nil
}

func (stummeswerk) AlsUnklarMerken(context.Context, string, string) (
	domain.Kostenposten, error,
) {
	return domain.Kostenposten{}, nil
}

// **Ein Schreibfehler beim Vergessen kommt durch.**
func TestEinSchreibfehlerBeimVergessenKommtDurch(t *testing.T) {
	t.Parallel()

	db := openaiAn(t)

	anwendung := KIAnwendungUeber(db)

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if antwort := anwendung.KIModellVergessen(riegelkontext(t),
		"openai", "gpt-probe"); antwort.Fehler == nil {
		t.Error("der Schreibfehler kam nicht durch")
	}
}
