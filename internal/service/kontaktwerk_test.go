package service_test

// **Wen dieses Profil kennt und was es ihm erlaubt** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Dass eine Bestätigung ein Vergleich ist und kein Klick.** Das Issue
// verlangt: „Gegenüber per Fingerprint/Bestätigung prüfen; Anzeigenamen allein
// sind kein Identitätsnachweis." Wer bestätigt, sagt **welchen** Abdruck er
// verglichen hat — und stimmt der nicht, endet es hier.

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// kontaktdoppel hält Kontakte im Arbeitsspeicher.
type kontaktdoppel struct {
	kontakte  map[string]*domain.Peerkontakt
	scheitert error
}

func neuesKontaktdoppel() *kontaktdoppel {
	return &kontaktdoppel{kontakte: map[string]*domain.Peerkontakt{}}
}

func (d *kontaktdoppel) Liste(
	context.Context,
) ([]domain.Peerkontakt, error) {
	if d.scheitert != nil {
		return nil, d.scheitert
	}

	liste := make([]domain.Peerkontakt, 0, len(d.kontakte))
	for _, kontakt := range d.kontakte {
		liste = append(liste, *kontakt)
	}

	return liste, nil
}

func (d *kontaktdoppel) Lesen(
	_ context.Context, kennung string,
) (domain.Peerkontakt, error) {
	if d.scheitert != nil {
		return domain.Peerkontakt{}, d.scheitert
	}

	kontakt, gibts := d.kontakte[kennung]
	if !gibts {
		return domain.Peerkontakt{}, domain.MitGrund(
			domain.ErrNotFound, nil, "einen Kontakt lesen")
	}

	return *kontakt, nil
}

func (d *kontaktdoppel) Begegnet(
	_ context.Context, kennung, fingerabdruck, name string, jetzt time.Time,
) error {
	if d.scheitert != nil {
		return d.scheitert
	}

	kontakt, gibts := d.kontakte[kennung]
	if !gibts {
		d.kontakte[kennung] = &domain.Peerkontakt{
			Kennung: kennung, Fingerabdruck: fingerabdruck,
			SelbstgewaehlterName: name,
			Stand:                domain.KontaktUnbestaetigt, Gesehen: jetzt,
		}

		return nil
	}

	kontakt.Fingerabdruck = fingerabdruck
	kontakt.SelbstgewaehlterName = name
	kontakt.Gesehen = jetzt

	return nil
}

func (d *kontaktdoppel) Stand(
	_ context.Context, kennung string, stand domain.Kontaktstand,
) error {
	if d.scheitert != nil {
		return d.scheitert
	}

	kontakt, gibts := d.kontakte[kennung]
	if !gibts {
		return domain.MitGrund(domain.ErrNotFound, nil, "einen Stand setzen")
	}

	kontakt.Stand = stand

	if stand != domain.KontaktBestaetigt {
		kontakt.Rechte = nil
	}

	return nil
}

func (d *kontaktdoppel) Erlauben(
	_ context.Context, kennung string, zweck domain.Zweck, _ time.Time,
) error {
	if d.scheitert != nil {
		return d.scheitert
	}

	kontakt, gibts := d.kontakte[kennung]
	if !gibts {
		return domain.MitGrund(domain.ErrNotFound, nil, "einen Zweck geben")
	}

	neu := *kontakt
	neu.Rechte = append(append([]domain.Zweck{}, kontakt.Rechte...), zweck)

	if fehler := neu.Validate(); fehler != nil {
		return fehler
	}

	kontakt.Rechte = neu.Rechte

	return nil
}

func (d *kontaktdoppel) Entziehen(
	_ context.Context, kennung string, zweck domain.Zweck,
) error {
	if d.scheitert != nil {
		return d.scheitert
	}

	kontakt, gibts := d.kontakte[kennung]
	if !gibts {
		return domain.MitGrund(domain.ErrNotFound, nil, "einen Zweck nehmen")
	}

	bleibt := make([]domain.Zweck, 0, len(kontakt.Rechte))

	for _, hat := range kontakt.Rechte {
		if hat != zweck {
			bleibt = append(bleibt, hat)
		}
	}

	kontakt.Rechte = bleibt

	return nil
}

func (d *kontaktdoppel) AdresseSetzen(
	_ context.Context, kennung, adresse string,
) error {
	if d.scheitert != nil {
		return d.scheitert
	}

	kontakt, gibts := d.kontakte[kennung]
	if !gibts {
		return domain.MitGrund(
			domain.ErrNotFound, nil, "eine Adresse setzen")
	}

	kontakt.Adresse = adresse

	return nil
}

func (d *kontaktdoppel) NotizSetzen(
	_ context.Context, kennung, notiz string,
) error {
	if d.scheitert != nil {
		return d.scheitert
	}

	kontakt, gibts := d.kontakte[kennung]
	if !gibts {
		return domain.MitGrund(domain.ErrNotFound, nil, "eine Notiz setzen")
	}

	kontakt.Notiz = notiz

	return nil
}

func (d *kontaktdoppel) AllesZurueck(context.Context) error {
	if d.scheitert != nil {
		return d.scheitert
	}

	for _, kontakt := range d.kontakte {
		if kontakt.Stand == domain.KontaktBestaetigt {
			kontakt.Stand = domain.KontaktUnbestaetigt
		}

		kontakt.Rechte = nil
	}

	return nil
}

// einePeeruhr steht still — dieselbe Zeit in jedem Lauf.
func einePeeruhr() time.Time {
	return time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
}

// einKontaktwerk gibt ein Werk über einem frischen Doppel.
func einKontaktwerk(t *testing.T) *service.Kontaktwerk {
	t.Helper()

	return service.NeuesKontaktwerk(neuesKontaktdoppel(), einePeeruhr)
}

// eineFremdeKennung gibt eine echte Peerkennung samt Abdruck.
func eineFremdeKennung(t *testing.T) (kennung peer.Kennung, abdruck string) {
	t.Helper()

	fremd, fehler := peer.Erzeugen(nil)
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return fremd.Kennung(), fremd.Fingerabdruck()
}

// **Eine Begegnung hinterlässt jemanden, der nichts darf außer sich
// vorzustellen.**
func TestEineBegegnungHinterlaesstJemandenOhneRechte(t *testing.T) {
	t.Parallel()

	werk := einKontaktwerk(t)
	kennung, _ := eineFremdeKennung(t)

	if fehler := werk.Begegnet(
		t.Context(), kennung, "Fremder Rechner"); fehler != nil {
		t.Fatalf("Begegnet: %v", fehler)
	}

	darf, fehler := werk.Darf(t.Context(), kennung, domain.ZweckHandel)
	if fehler != nil {
		t.Fatalf("Darf: %v", fehler)
	}

	if darf {
		t.Error("er darf handeln")
	}

	darf, fehler = werk.Darf(t.Context(), kennung, domain.ZweckPaarung)
	if fehler != nil {
		t.Fatalf("Darf: %v", fehler)
	}

	if !darf {
		t.Error("er darf sich nicht vorstellen")
	}
}

// **Eine Bestätigung ist ein Vergleich und kein Klick.**
//
// # Der Satz, für den es diesen Lauf gibt
//
// **Wer bestätigt, sagt, welchen Abdruck er verglichen hat.** Stimmt er nicht
// mit dem der Kennung überein, hat der Mensch etwas anderes vor sich gehabt
// als das, was hier steht — und dann ist es keine Bestätigung, sondern ein
// Irrtum, der still zur Freigabe würde.
func TestEineBestaetigungIstEinVergleich(t *testing.T) {
	t.Parallel()

	werk := einKontaktwerk(t)
	kennung, abdruck := eineFremdeKennung(t)

	if fehler := werk.Begegnet(t.Context(), kennung, "Wer"); fehler != nil {
		t.Fatalf("Begegnet: %v", fehler)
	}

	if fehler := werk.Bestaetigen(
		t.Context(), kennung, "0000-1111-2222-3333-4444-5555-6666-7777",
	); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("der falsche Abdruck kam durch: %v", fehler)
	}

	darf, fehler := werk.Darf(t.Context(), kennung, domain.ZweckTeilen)
	if fehler != nil {
		t.Fatalf("Darf: %v", fehler)
	}

	if darf {
		t.Fatal("der falsche Abdruck hat freigegeben")
	}

	if befund := werk.Bestaetigen(
		t.Context(), kennung, abdruck); befund != nil {
		t.Fatalf("Bestaetigen: %v", befund)
	}

	// **Und auch der richtige Abdruck gibt noch nichts frei.** Er macht das
	// Gegenüber prüfbar, nicht berechtigt — „Pairing allein gestattet keinen
	// Import oder Handel".
	darf, fehler = werk.Darf(t.Context(), kennung, domain.ZweckTeilen)
	if fehler != nil {
		t.Fatalf("Darf: %v", fehler)
	}

	if darf {
		t.Fatal("die Bestätigung allein hat freigegeben")
	}
}

// **Ein freigegebener Zweck gilt, ein entzogener nicht mehr.**
func TestEinFreigegebenerZweckGiltUndEinEntzogenerNicht(t *testing.T) {
	t.Parallel()

	werk := einKontaktwerk(t)
	kennung, abdruck := eineFremdeKennung(t)

	if fehler := werk.Begegnet(t.Context(), kennung, "Wer"); fehler != nil {
		t.Fatalf("Begegnet: %v", fehler)
	}

	if fehler := werk.Bestaetigen(
		t.Context(), kennung, abdruck); fehler != nil {
		t.Fatalf("Bestaetigen: %v", fehler)
	}

	if befund := werk.Erlauben(
		t.Context(), kennung, domain.ZweckTeilen); befund != nil {
		t.Fatalf("Erlauben: %v", befund)
	}

	darf, fehler := werk.Darf(t.Context(), kennung, domain.ZweckTeilen)
	if fehler != nil {
		t.Fatalf("Darf: %v", fehler)
	}

	if !darf {
		t.Fatal("der freigegebene Zweck gilt nicht")
	}

	if befund := werk.Entziehen(
		t.Context(), kennung, domain.ZweckTeilen); befund != nil {
		t.Fatalf("Entziehen: %v", befund)
	}

	darf, fehler = werk.Darf(t.Context(), kennung, domain.ZweckTeilen)
	if fehler != nil {
		t.Fatalf("Darf: %v", fehler)
	}

	if darf {
		t.Fatal("der entzogene Zweck gilt noch")
	}
}

// **Wen niemand kennt, der darf nichts — und das ist kein Fehler.**
//
// # Warum nicht `ErrNotFound`
//
// **Weil die Frage beantwortet ist.** „Darf dieses Gegenüber handeln?" hat für
// einen Unbekannten eine klare Antwort, und ein Fehler an dieser Stelle
// zwänge jeden Aufrufer, ihn wegzufangen — bis einer es vergisst und die
// Verbindung mit einer Meldung statt mit einem Nein endet.
func TestWenNiemandKenntDerDarfNichts(t *testing.T) {
	t.Parallel()

	werk := einKontaktwerk(t)
	kennung, _ := eineFremdeKennung(t)

	for _, zweck := range domain.DieZwecke {
		darf, fehler := werk.Darf(t.Context(), kennung, zweck)
		if fehler != nil {
			t.Fatalf("Darf(%q): %v", zweck, fehler)
		}

		// **Außer sich vorzustellen** — sonst käme niemand je in die Liste.
		if darf != (zweck == domain.ZweckPaarung) {
			t.Errorf("%q ist %v", zweck, darf)
		}
	}
}
