package service_test

// **Wer darf was zurücksetzen** — Werkstück O4 von Phase 5.2
// (#1428).
//
// Die drei Zusagen der Abnahme, jede mit ihrem Test:
//
//  1. „Gesamtreset verlangt passende lokale Bedien-/Administratorautorisierung."
//  2. „Kein Profilverwalter erhält durch umfassenden Reset Rechte über andere
//     Eigentümer."
//  3. „Offene … Jobs vor Löschung geordnet stoppen."

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// wachdoppel spielt die Autorisierung.
type wachdoppel struct {
	akteur  domain.Akteur
	fehler  error
	absage  error
	kennung string
	bindung service.Nachweisbindung
	wirkung domain.Wirkung
}

func (w *wachdoppel) Akteur(context.Context) (domain.Akteur, error) {
	return w.akteur, w.fehler
}

func (w *wachdoppel) PruefeMitNachweis(
	_ context.Context, wirkung domain.Wirkung, kennung string,
	bindung service.Nachweisbindung,
) error {
	w.wirkung = wirkung
	w.kennung = kennung
	w.bindung = bindung

	return w.absage
}

// wache gibt eine Wache für den Menschen 7 im Profil 3.
func wache() *wachdoppel {
	return &wachdoppel{akteur: domain.Akteur{NutzerID: 7, ProfilID: 3, EignerID: 7}}
}

// **Ohne Nachweis geht die ganze Anwendung nicht** — und die Vorschau sagt es
// vorher, ohne sich als gesperrt auszugeben.
func TestDieGanzeAnwendungVerlangtEinenNachweis(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{}
	wacht := wache()
	setzung := service.NeueRuecksetzung(doppel, doppel, wacht, doppel.neuanfang)

	vorschau, fehler := setzung.Vorschau(t.Context(),
		domain.RuecksetzungAnwendung, service.Ruecksetzungsziel{})
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if !vorschau.NachweisNoetig {
		t.Error("die Vorschau sagt nicht, dass ein Nachweis nötig ist")
	}

	// **Und sie ist trotzdem bereit** — der Nachweis fehlt beim Ausführen,
	// nicht beim Nachsehen.
	if !vorschau.Bereit() {
		t.Errorf("die Vorschau nennt Hindernisse: %v", vorschau.Hindernisse)
	}

	_, fehler = setzung.Zuruecksetzen(t.Context(),
		domain.RuecksetzungAnwendung, service.Ruecksetzungsziel{}, "",
		gesehen(t, setzung, domain.RuecksetzungAnwendung, service.Ruecksetzungsziel{}))
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("ohne Nachweis: %v", fehler)
	}

	if doppel.neuanfaenge != 0 {
		t.Error("der Bestand wurde trotzdem beiseitegelegt")
	}
}

// **Mit Nachweis wird er eingelöst** — gebunden an Zweck und Art.
func TestDerNachweisWirdAnZweckUndArtGebunden(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{}
	wacht := wache()
	setzung := service.NeueRuecksetzung(doppel, doppel, wacht, doppel.neuanfang)

	bilanz, fehler := setzung.Zuruecksetzen(t.Context(),
		domain.RuecksetzungAnwendung, service.Ruecksetzungsziel{}, "abc123",
		gesehen(t, setzung, domain.RuecksetzungAnwendung, service.Ruecksetzungsziel{}))
	if fehler != nil {
		t.Fatalf("Zuruecksetzen: %v", fehler)
	}

	if bilanz.Beiseite == "" {
		t.Error("die Bilanz nennt nicht, wo der alte Bestand liegt")
	}

	if wacht.kennung != "abc123" {
		t.Errorf("eingelöst wurde %q", wacht.kennung)
	}

	if wacht.bindung.Zweck != domain.RuecksetzungsZweck ||
		wacht.bindung.Ziel != string(domain.RuecksetzungAnwendung) {
		t.Errorf("die Bindung lautet %+v", wacht.bindung)
	}

	if wacht.wirkung != domain.WirkungSchreiben {
		t.Errorf("geprüft wurde die Wirkung %q", wacht.wirkung)
	}
}

// **Eine Absage der Wache hält an** — und zwar vor der Wirkung.
func TestEineAbsageDerWacheHaeltDieRuecksetzungAn(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{}
	wacht := wache()
	wacht.absage = domain.ErrErlaubnis
	setzung := service.NeueRuecksetzung(doppel, doppel, wacht, doppel.neuanfang)

	_, fehler := setzung.Zuruecksetzen(t.Context(),
		domain.RuecksetzungLayout, service.Ruecksetzungsziel{}, "", "")
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("mit Absage: %v", fehler)
	}

	if doppel.zurueck != 0 {
		t.Error("es wurde trotzdem zurückgesetzt")
	}
}

// **Ein fremdes Profil ist ein Hindernis** — und keine Löschung.
func TestEinFremdesProfilLaesstSichNichtZuruecksetzen(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{profile: []domain.Profil{{ID: 3, Name: "meins"}}}
	wacht := wache()
	setzung := service.NeueRuecksetzung(doppel, doppel, wacht, doppel.neuanfang)

	vorschau, fehler := setzung.Vorschau(t.Context(),
		domain.RuecksetzungProfil, service.Ruecksetzungsziel{ProfilID: 99})
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if vorschau.Bereit() {
		t.Fatal("ein fremdes Profil gilt als zurücksetzbar")
	}

	if vorschau.Hindernisse[0].Art != domain.RuecksetzungFremdesZiel {
		t.Errorf("das Hindernis heißt %q", vorschau.Hindernisse[0].Art)
	}

	if _, fehler := setzung.Zuruecksetzen(t.Context(), domain.RuecksetzungProfil,
		service.Ruecksetzungsziel{ProfilID: 99, Name: "meins"}, "", ""); !errors.Is(
		fehler, domain.ErrConflict) {
		t.Fatalf("beim Ausführen: %v", fehler)
	}

	if doppel.geloescht != "" {
		t.Errorf("gelöscht wurde %q", doppel.geloescht)
	}
}

// **Ein fremder Mensch auch nicht.**
func TestEinFremderMenschLaesstSichNichtLoeschen(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{nutzer: 4}
	wacht := wache()
	setzung := service.NeueRuecksetzung(doppel, doppel, wacht, doppel.neuanfang)

	vorschau, fehler := setzung.Vorschau(t.Context(),
		domain.RuecksetzungNutzer, service.Ruecksetzungsziel{NutzerID: 8})
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if vorschau.Bereit() ||
		vorschau.Hindernisse[0].Art != domain.RuecksetzungFremdesZiel {
		t.Errorf("die Hindernisse lauten %v", vorschau.Hindernisse)
	}
}

// **Das eigene Profil geht** — und ohne Nennung ist es das gezeigte.
func TestOhneNennungIstDasZielDasGezeigteProfil(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{profile: []domain.Profil{{ID: 3, Name: "meins"}}}
	wacht := wache()
	setzung := service.NeueRuecksetzung(doppel, doppel, wacht, doppel.neuanfang)

	ziel := service.Ruecksetzungsziel{Name: "meins"}

	if _, fehler := setzung.Zuruecksetzen(t.Context(), domain.RuecksetzungProfil,
		ziel, "", gesehen(t, setzung, domain.RuecksetzungProfil, ziel)); fehler != nil {
		t.Fatalf("Zuruecksetzen: %v", fehler)
	}

	if doppel.geloescht != "profil:meins" {
		t.Errorf("gelöscht wurde %q", doppel.geloescht)
	}
}

// **Ein laufender Auftrag sperrt, was löscht** — und nur das.
func TestEinLaufenderAuftragSperrtDieLoeschung(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{
		nutzer: 4, profile: []domain.Profil{{ID: 3, Name: "meins"}},
	}
	wacht := wache()
	setzung := service.NeueRuecksetzung(doppel, doppel, wacht, doppel.neuanfang).
		MitAuftraegen(func() []domain.Auftrag {
			return []domain.Auftrag{
				{ID: 1, Was: "fertig", Stand: domain.AuftragFertig},
				{ID: 2, Was: "die Ausfuhr", Stand: domain.AuftragLaeuft},
			}
		})

	vorschau, fehler := setzung.Vorschau(t.Context(),
		domain.RuecksetzungNutzer, service.Ruecksetzungsziel{})
	if fehler != nil {
		t.Fatalf("Vorschau: %v", fehler)
	}

	if vorschau.Bereit() {
		t.Fatal("mit laufendem Auftrag gilt die Löschung als bereit")
	}

	if vorschau.Hindernisse[0].Art != domain.RuecksetzungLaufenderAuftrag ||
		vorschau.Hindernisse[0].Stelle != "die Ausfuhr" {
		t.Errorf("das Hindernis lautet %+v", vorschau.Hindernisse[0])
	}

	// **Und die Ausführung weist ab, mit dem Satz zum Auftrag.**
	_, ausfuehrfehler := setzung.Zuruecksetzen(t.Context(),
		domain.RuecksetzungNutzer, service.Ruecksetzungsziel{Name: "David"}, "", "")
	if ausfuehrfehler == nil {
		t.Fatal("mit laufendem Auftrag wurde gelöscht")
	}

	if !strings.Contains(ausfuehrfehler.Error(), "die Ausfuhr") {
		t.Errorf("der Satz nennt den Auftrag nicht: %v", ausfuehrfehler)
	}

	// **Eine Wahlrücknahme läuft weiter** — sie nimmt niemandem etwas weg.
	layout, fehler := setzung.Vorschau(t.Context(),
		domain.RuecksetzungLayout, service.Ruecksetzungsziel{})
	if fehler != nil {
		t.Fatalf("Vorschau Layout: %v", fehler)
	}

	if !layout.Bereit() {
		t.Errorf("die Layoutrücksetzung ist gesperrt: %v", layout.Hindernisse)
	}
}

// **Ein Fehler der Wache kommt durch** — und nicht als leerer Akteur.
func TestEinFehlerDerWacheKommtDurch(t *testing.T) {
	t.Parallel()

	doppel := &ruecksetzungsdoppel{}
	wacht := wache()
	wacht.fehler = errors.New("kein Zeiger")
	setzung := service.NeueRuecksetzung(doppel, doppel, wacht, doppel.neuanfang)

	if _, fehler := setzung.Vorschau(t.Context(), domain.RuecksetzungLayout,
		service.Ruecksetzungsziel{}); fehler == nil {
		t.Fatal("der Fehler der Wache ist verschwunden")
	}
}
