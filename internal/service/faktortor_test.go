package service_test

// **Das Tor: wo der zweite Faktor verlangt wird** — Werkstück 7 von
// Phase 4.12
// (#1351).
//
// # Zwei Tore und ein drittes, das keines ist
//
// **Der Profilwechsel** und die **Anmeldung unter einer eigenen Kennung**
// fragen. Die Anmeldung über einen **Anbieter** fragt nicht — der hat seinen
// eigenen zweiten Faktor, und einer obendrauf wäre eine Hürde auf einen
// Nachweis, den dieses Programm nicht ausgestellt hat.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/kennwort"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Wechsel, den die Prüfung abweist, findet nicht statt.**
//
// **Am Zeiger und nicht am Rückgabewert** (§10): eine Abnahme ist
// beobachtbares Verhalten. „Es kam ein Fehler" sagt nichts darüber, ob der
// Zeiger schon gewandert war.
func TestEinAbgewiesenerWechselFindetNichtStatt(t *testing.T) {
	t.Parallel()

	ablage := zweiprofile()

	verwaltung := service.NeueNutzerverwaltung(ablage).
		MitFaktor(&pruefdoppel{fehler: domain.FehlenderFaktor(domain.FaktorTOTP)})

	_, fehler := verwaltung.Wechseln(t.Context(), 2, "")
	if !errors.Is(fehler, domain.ErrZweiterFaktor) {
		t.Fatalf("durchgelassen wurde: %v", fehler)
	}

	if ablage.zeiger != 1 {
		t.Errorf("der Zeiger wanderte auf %d", ablage.zeiger)
	}
}

// **Und einer, den sie durchlässt, findet statt.**
//
// **Die Gegenprobe zur Zeile darüber**: ohne sie prüfte der Lauf nur, dass
// ein Fehler abweist — nicht, dass die Prüfung überhaupt etwas durchlässt.
func TestEinBestaetigterWechselFindetStatt(t *testing.T) {
	t.Parallel()

	ablage := zweiprofile()
	pruefer := &pruefdoppel{}

	verwaltung := service.NeueNutzerverwaltung(ablage).MitFaktor(pruefer)

	if _, fehler := verwaltung.Wechseln(t.Context(), 2, "123456"); fehler != nil {
		t.Fatalf("abgewiesen: %v", fehler)
	}

	if ablage.zeiger != 2 {
		t.Errorf("der Zeiger steht auf %d", ablage.zeiger)
	}

	// **Und der Nachweis reist bis zum Prüfer — samt dem Nutzer, dem das
	// Ziel gehört.**
	if pruefer.nachweis != "123456" || pruefer.nutzerID != 1 {
		t.Errorf("geprüft wurde %+v", pruefer)
	}
}

// **Ohne Prüfer wird nicht gefragt.**
//
// **Eine Verwaltung ohne Prüfer bricht keine Zusage, sondern macht keine** —
// und ein Lauf, der nur den Wechsel meint, soll nicht an einer Naht scheitern,
// die zu seinem Fall nichts beiträgt.
func TestOhnePrueferWirdNichtGefragt(t *testing.T) {
	t.Parallel()

	ablage := zweiprofile()

	if _, fehler := service.NeueNutzerverwaltung(ablage).
		Wechseln(t.Context(), 2, ""); fehler != nil {
		t.Fatalf("abgewiesen: %v", fehler)
	}

	if ablage.zeiger != 2 {
		t.Errorf("der Zeiger steht auf %d", ablage.zeiger)
	}
}

// **Die Anmeldung unter einer eigenen Kennung fragt ebenso.**
func TestDieAnmeldungUnterEigenerKennungFragt(t *testing.T) {
	t.Parallel()

	gehuetet, fehler := kennwort.Huete("einlangesgeheimnis")
	if fehler != nil {
		t.Fatalf("hüten: %v", fehler)
	}

	ablage := &kennungsdoppel{bekannt: domain.Nutzer{
		ID: 1, Name: "Ich", Email: "wer@example.org", Kennwort: gehuetet,
	}}

	pruefer := &pruefdoppel{fehler: domain.FehlenderFaktor(domain.FaktorTOTP)}

	_, abgewiesen := service.NeueKontoanmeldung(ablage).MitFaktor(pruefer).
		Anmelden(t.Context(), "wer@example.org", "einlangesgeheimnis", "")
	if !errors.Is(abgewiesen, domain.ErrZweiterFaktor) {
		t.Fatalf("durchgelassen wurde: %v", abgewiesen)
	}

	// **Und zwar erst nach dem Kennwort** — sonst verriete die Frage, dass es
	// dieses Konto gibt.
	if pruefer.nutzerID != 1 {
		t.Errorf("geprüft wurde %+v", pruefer)
	}
}

// **Und mit dem Nachweis kommt sie durch.**
func TestDieAnmeldungKommtMitNachweisDurch(t *testing.T) {
	t.Parallel()

	gehuetet, fehler := kennwort.Huete("einlangesgeheimnis")
	if fehler != nil {
		t.Fatalf("hüten: %v", fehler)
	}

	ablage := &kennungsdoppel{bekannt: domain.Nutzer{
		ID: 1, Name: "Ich", Email: "wer@example.org", Kennwort: gehuetet,
	}}

	herein, abgewiesen := service.NeueKontoanmeldung(ablage).
		MitFaktor(&pruefdoppel{}).
		Anmelden(t.Context(), "wer@example.org", "einlangesgeheimnis", "123456")
	if abgewiesen != nil {
		t.Fatalf("abgewiesen: %v", abgewiesen)
	}

	if herein.ID != 1 {
		t.Errorf("herein kam %+v", herein)
	}
}

// zweiprofile gibt eine Ablage mit einem Nutzer und zwei Profilen.
//
// **Der Zeiger steht auf dem ersten** — sonst schriebe ein Wechsel auf das
// gezeigte gar nicht, und der Lauf prüfte den Weg nicht, den er meint.
func zweiprofile() *nutzerdoppel {
	return &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{
			1: {
				{ID: 1, NutzerID: 1, Name: "Erstes"},
				{ID: 2, NutzerID: 1, Name: "Zweites"},
			},
		},
		zeiger: 1,
	}
}

// pruefdoppel ist ein Prüfer, der sagt, was der Lauf will.
type pruefdoppel struct {
	fehler error

	// nutzerID und nachweis halten fest, wonach gefragt wurde.
	nutzerID int64
	nachweis string

	// verlangt sagt je Nutzer, ob ein Wechsel zu ihm nach dem Faktor fragte
	// — ID-03 (#1702); verlangtfehler lässt die Auskunft scheitern.
	verlangt       map[int64]bool
	verlangtfehler error
}

func (p *pruefdoppel) VerlangtBeimWechsel(_ context.Context, nutzerID int64) (bool, error) {
	return p.verlangt[nutzerID], p.verlangtfehler
}

func (p *pruefdoppel) Pruefe(_ context.Context, nutzerID int64, nachweis string) error {
	p.nutzerID = nutzerID
	p.nachweis = nachweis

	return p.fehler
}

// PruefeBeimWechsel ist derselbe Weg mit einem anderen Namen.
//
// **Im Betrieb sind es zwei**: der Nutzer darf den einen abschalten und den
// anderen nicht. **In diesem Doppel nicht**, weil die Läufe hier prüfen, ob
// **überhaupt** gefragt wird — welcher der beiden, prüft
// `TestDerWechselLaesstSichAbschaltenUndDasAnmeldenNicht`.
func (p *pruefdoppel) PruefeBeimWechsel(
	ctx context.Context, nutzerID int64, nachweis string,
) error {
	return p.Pruefe(ctx, nutzerID, nachweis)
}

// **Der Wechsel lässt sich abschalten, das Anmelden nicht.**
//
// # Warum das die eigentliche Aussage von Punkt 11 ist
//
// **Weil eine Übersteuerung, die alles abschaltet, keine Einstellung ist,
// sondern ein Ausschalter.** Auf einem Rechner mit einem Menschen ist ein
// Einmalkennwort bei jedem Profilwechsel Reibung ohne Gegenwert; beim
// Anmelden steht der Faktor für das ein, wofür es ihn gibt.
func TestDerWechselLaesstSichAbschaltenUndDasAnmeldenNicht(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
	}

	einstellungen := &einstellungsdoppel{
		werte: map[string]string{
			domain.NutzerFaktorbeimWechsel: domain.FaktorNurAnmelden,
		},
	}

	dienst := service.NeuerZweiterfaktor(ablage, ablage, &sensordoppel{}).
		MitEinstellungen(service.NeueNutzereinstellungen(einstellungen))

	// **Der Wechsel fragt nicht mehr.**
	if fehler := dienst.PruefeBeimWechsel(t.Context(), 1, ""); fehler != nil {
		t.Errorf("der Wechsel wurde trotzdem abgewiesen: %v", fehler)
	}

	// **Das Anmelden schon.**
	if fehler := dienst.Pruefe(t.Context(), 1, ""); !errors.Is(
		fehler, domain.ErrZweiterFaktor) {
		t.Errorf("das Anmelden ließ durch: %v", fehler)
	}
}

// **Und mit der Vorgabe fragt auch der Wechsel.**
//
// **Die Gegenprobe zur Zeile darüber**: ohne sie prüfte der Lauf nur, dass
// eine Einstellung etwas abschaltet — nicht, dass ohne sie geprüft wird.
func TestMitDerVorgabeFragtAuchDerWechsel(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
	}

	dienst := service.NeuerZweiterfaktor(ablage, ablage, &sensordoppel{}).
		MitEinstellungen(service.NeueNutzereinstellungen(&einstellungsdoppel{}))

	if fehler := dienst.PruefeBeimWechsel(t.Context(), 1, ""); !errors.Is(
		fehler, domain.ErrZweiterFaktor) {
		t.Errorf("der Wechsel ließ durch: %v", fehler)
	}
}

// **Ohne Einstellungen wird geprüft.**
//
// **Die strengere Antwort ist die richtige, wenn niemand gefragt hat.**
func TestOhneEinstellungenWirdGeprueft(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
	}

	dienst := service.NeuerZweiterfaktor(ablage, ablage, &sensordoppel{})

	if fehler := dienst.PruefeBeimWechsel(t.Context(), 1, ""); !errors.Is(
		fehler, domain.ErrZweiterFaktor) {
		t.Errorf("der Wechsel ließ durch: %v", fehler)
	}
}

// **Ein kaputter Bestand reißt auch hier ab.**
func TestEinKaputterBestandReisstBeimWechselAb(t *testing.T) {
	t.Parallel()

	ablage := &faktordoppel{
		faktoren: []domain.Faktor{{NutzerID: 1, Art: domain.FaktorTOTP}},
	}

	dienst := service.NeuerZweiterfaktor(ablage, ablage, &sensordoppel{}).
		MitEinstellungen(service.NeueNutzereinstellungen(
			&einstellungsdoppel{fehler: errStummerBestand}))

	if fehler := dienst.PruefeBeimWechsel(t.Context(), 1, ""); !errors.Is(
		fehler, errStummerBestand) {
		t.Errorf("durchgelassen wurde: %v", fehler)
	}
}

// einstellungsdoppel ist eine Ablage der Nutzereinstellungen im Speicher.
type einstellungsdoppel struct {
	werte  map[string]string
	fehler error
}

func (e *einstellungsdoppel) Nutzereinstellungen(context.Context, int64) (
	map[string]string, error,
) {
	if e.fehler != nil {
		return nil, e.fehler
	}

	if e.werte == nil {
		return map[string]string{}, nil
	}

	return e.werte, nil
}

func (e *einstellungsdoppel) NutzereinstellungSetzen(
	_ context.Context, _ int64, schluessel, wert string,
) error {
	if e.fehler != nil {
		return e.fehler
	}

	if e.werte == nil {
		e.werte = map[string]string{}
	}

	e.werte[schluessel] = wert

	return nil
}
