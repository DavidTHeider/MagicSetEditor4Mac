package service_test

// **Nachweise ausstellen und einlösen** — Werkstück O1 von Phase 5.2
// (#1425).
//
// # Was hier geprüft wird
//
// Ein Nachweis entsteht dort, wo der Mensch ist — am Fenster, ohne Token —,
// und er wird **einmal** eingelöst: gegen die Bindung, die der Aufruf
// mitbringt. Was davon abweicht, erlaubt nichts Zusätzliches — das ist die
// Abnahme des Issues, Zeile 4.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// nachweisdoppel ist eine Nachweisablage im Arbeitsspeicher.
type nachweisdoppel struct {
	abgelegt map[string]domain.Nachweis
	kaputt   error

	// einloesefehler ist ein Fehler, den nur das Einlösen gibt — für den
	// Weg, auf dem die Ablage nach dem Lesen etwas anderes sagt als vorher.
	einloesefehler error
}

func neuesNachweisdoppel() *nachweisdoppel {
	return &nachweisdoppel{abgelegt: map[string]domain.Nachweis{}}
}

func (d *nachweisdoppel) NachweisAnlegen(_ context.Context, n domain.Nachweis) error {
	if d.kaputt != nil {
		return d.kaputt
	}

	d.abgelegt[n.Kennung] = n

	return nil
}

func (d *nachweisdoppel) Nachweis(_ context.Context, kennung string) (domain.Nachweis, error) {
	if d.kaputt != nil {
		return domain.Nachweis{}, d.kaputt
	}

	n, da := d.abgelegt[kennung]
	if !da {
		return domain.Nachweis{}, domain.ErrNotFound
	}

	return n, nil
}

func (d *nachweisdoppel) NachweisEinloesen(_ context.Context, kennung string) error {
	if d.einloesefehler != nil {
		return d.einloesefehler
	}

	n, da := d.abgelegt[kennung]
	if !da {
		return domain.ErrNotFound
	}

	if n.Eingeloest {
		return domain.ErrConflict
	}

	n.Eingeloest = true
	d.abgelegt[kennung] = n

	return nil
}

var uhrzeit = time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)

func nachweise(ablage *nachweisdoppel) *service.Nachweise {
	return service.NeueNachweise(ablage).
		MitUhr(func() time.Time { return uhrzeit }).
		MitKennung(func() string { return "n-1" })
}

func verwalterAmFenster() domain.Akteur {
	return domain.Akteur{
		NutzerID: 1, ProfilID: 7, EignerID: 1, Rolle: domain.RolleVerwalter,
		Modus: domain.ModusFenster, Revision: 3,
	}
}

func antrag() service.Nachweisantrag {
	return service.Nachweisantrag{
		Zweck: "handel.angebot", Ziel: "angebot-42", Revision: 5,
		Grenze: domain.Kostengrenze{Betrag: 1500, Waehrung: "EUR"},
		Frist:  10 * time.Minute,
	}
}

func bindungFuer(n domain.Nachweis) domain.Nachweisbindung {
	return domain.Nachweisbindung{
		NutzerID: n.NutzerID, ProfilID: n.ProfilID, Zweck: n.Zweck, Ziel: n.Ziel,
		Revision: n.Revision, Kosten: 1200, Waehrung: "EUR",
	}
}

func TestEinNachweisWirdAusgestelltUndEinmalEingeloest(t *testing.T) {
	t.Parallel()

	ablage := neuesNachweisdoppel()
	n, fehler := nachweise(ablage).Ausstellen(context.Background(), verwalterAmFenster(), antrag())
	if fehler != nil {
		t.Fatalf("Ausstellen: %v", fehler)
	}

	if n.Kennung != "n-1" || n.NutzerID != 1 || n.ProfilID != 7 || n.Zweck != "handel.angebot" ||
		n.Ablauf != uhrzeit.Add(10*time.Minute) || n.Ausgestellt != uhrzeit || n.Eingeloest {
		t.Fatalf("Nachweis %+v", n)
	}

	if einloesefehler := nachweise(ablage).Einloesen(context.Background(), "n-1", bindungFuer(n)); einloesefehler != nil {
		t.Fatalf("Einlösen: %v", einloesefehler)
	}

	// **Ein zweites Mal geht nicht.**
	fehler = nachweise(ablage).Einloesen(context.Background(), "n-1", bindungFuer(n))
	if !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("zweites Einlösen: %v, erwartet ErrErlaubnis", fehler)
	}
}

// **Nur ein Mensch am Fenster stellt aus** — kein Token, keine Maschine,
// kein Leser.
func TestNurEinMenschAmFensterStelltAus(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*domain.Akteur){
		"über den Dienst":      func(a *domain.Akteur) { a.Modus = domain.ModusDienst },
		"über MCP":             func(a *domain.Akteur) { a.Modus = domain.ModusMCP },
		"mit Token":            func(a *domain.Akteur) { a.Token = domain.ScopeSchreiben },
		"als Leser":            func(a *domain.Akteur) { a.Rolle = domain.RolleLeser; a.NurLesend = true },
		"ohne Rolle":           func(a *domain.Akteur) { a.Rolle = "" },
		"an der Kommandozeile": func(a *domain.Akteur) { a.Modus = domain.ModusKommandozeile },
	}

	for name, verschiebe := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wer := verwalterAmFenster()
			verschiebe(&wer)

			_, fehler := nachweise(neuesNachweisdoppel()).Ausstellen(context.Background(), wer, antrag())
			if !errors.Is(fehler, domain.ErrErlaubnis) {
				t.Fatalf("%s: %v, erwartet ErrErlaubnis", name, fehler)
			}
		})
	}
}

// **Ein Mitarbeiter darf ausstellen** — die Rolle entscheidet über die
// Wirkung, der Nachweis über das Eine.
func TestEinMitarbeiterStelltAus(t *testing.T) {
	t.Parallel()

	wer := verwalterAmFenster()
	wer.NutzerID, wer.Rolle = 2, domain.RolleMitarbeiter

	if _, fehler := nachweise(neuesNachweisdoppel()).Ausstellen(context.Background(), wer, antrag()); fehler != nil {
		t.Fatalf("Ausstellen als Mitarbeiter: %v", fehler)
	}
}

// **Ein Antrag ohne Frist oder mit einer über einem Tag ist keiner.**
func TestEinAntragBrauchtEineFristInGrenzen(t *testing.T) {
	t.Parallel()

	for name, frist := range map[string]time.Duration{"ohne": 0, "negativ": -time.Minute, "über einem Tag": 25 * time.Hour} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			a := antrag()
			a.Frist = frist

			_, fehler := nachweise(neuesNachweisdoppel()).Ausstellen(context.Background(), verwalterAmFenster(), a)
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("Frist %v: %v, erwartet ErrValidation", frist, fehler)
			}
		})
	}

	// **Und ein Antrag ohne Ziel scheitert an der Prüfung des Nachweises.**
	a := antrag()
	a.Ziel = ""
	if _, fehler := nachweise(neuesNachweisdoppel()).Ausstellen(context.Background(), verwalterAmFenster(), a); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("ohne Ziel: %v, erwartet ErrValidation", fehler)
	}
}

// **Die Fehlerwege der Ablage und der Kennung** werden betreten (§3).
func TestDieFehlerwegeDerNachweise(t *testing.T) {
	t.Parallel()

	kaputt := errors.New("kaputt")

	ablage := neuesNachweisdoppel()
	ablage.kaputt = kaputt
	if _, fehler := nachweise(ablage).Ausstellen(context.Background(), verwalterAmFenster(), antrag()); !errors.Is(fehler, kaputt) {
		t.Errorf("kaputte Ablage beim Ausstellen: %v", fehler)
	}

	if fehler := nachweise(ablage).Einloesen(context.Background(), "n-1", domain.Nachweisbindung{}); !errors.Is(fehler, kaputt) {
		t.Errorf("kaputte Ablage beim Einlösen: %v", fehler)
	}

	// **Ein Fehler der Ablage beim Einlösen bleibt, was er ist** — nur der
	// Konflikt wird zur Absage.
	stoerrisch := neuesNachweisdoppel()
	stoerrischer := mussAusstellen(t, stoerrisch)
	stoerrisch.einloesefehler = kaputt
	if fehler := nachweise(stoerrisch).Einloesen(context.Background(), "n-1", bindungFuer(stoerrischer)); !errors.Is(fehler, kaputt) {
		t.Errorf("Ablagefehler beim Einlösen: %v", fehler)
	}

	// **Zwei zugleich, und die Ablage entscheidet.** Beide lesen den Nachweis
	// als offen, beide prüfen die Bindung, und nur einer trifft beim
	// Einlösen eine Zeile — der andere hört, dass jemand vor ihm war.
	wettlauf := neuesNachweisdoppel()
	umstritten := mussAusstellen(t, wettlauf)
	wettlauf.einloesefehler = fmt.Errorf("%w: schon weg", domain.ErrConflict)
	if fehler := nachweise(wettlauf).Einloesen(context.Background(), "n-1", bindungFuer(umstritten)); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("Wettlauf beim Einlösen: %v, erwartet ErrErlaubnis", fehler)
	}

	// **Unbekannt heißt unbekannt** — und ein abgelaufener Nachweis erlaubt
	// nichts.
	if fehler := nachweise(neuesNachweisdoppel()).Einloesen(context.Background(), "gibts-nicht", domain.Nachweisbindung{}); !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("unbekannter Nachweis: %v", fehler)
	}

	abgelaufen := neuesNachweisdoppel()
	n := mussAusstellen(t, abgelaufen)
	spaeter := service.NeueNachweise(abgelaufen).MitUhr(func() time.Time { return uhrzeit.Add(time.Hour) })
	if fehler := spaeter.Einloesen(context.Background(), "n-1", bindungFuer(n)); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("abgelaufener Nachweis: %v", fehler)
	}

	// **Die Vorgaben tragen sich selbst**: eine echte Uhr, eine zufällige,
	// eindeutige Kennung.
	echt := service.NeueNachweise(neuesNachweisdoppel())
	eins, fehler := echt.Ausstellen(context.Background(), verwalterAmFenster(), antrag())
	zwei, fehler2 := echt.Ausstellen(context.Background(), verwalterAmFenster(), antrag())
	if fehler != nil || fehler2 != nil || eins.Kennung == zwei.Kennung || len(eins.Kennung) < 16 {
		t.Errorf("Vorgaben: %v %v %q %q", fehler, fehler2, eins.Kennung, zwei.Kennung)
	}
}

// **Die Prüfung mit Nachweis**: die Rolle muss reichen **und** der Nachweis
// passen — und der Nachweis ist danach verbraucht.
func TestDiePruefungMitNachweis(t *testing.T) {
	t.Parallel()

	nutzer, freigaben := bestandFuerAkteure("")
	ablage := neuesNachweisdoppel()
	a := autorisierung(nutzer, freigaben).MitNachweisen(nachweise(ablage))

	wer, akteurfehler := a.Akteur(context.Background())
	if akteurfehler != nil {
		t.Fatalf("Akteur: %v", akteurfehler)
	}

	n, fehler := nachweise(ablage).Ausstellen(context.Background(), wer, antrag())
	if fehler != nil {
		t.Fatalf("Ausstellen: %v", fehler)
	}

	bindung := service.Nachweisbindung{Zweck: n.Zweck, Ziel: n.Ziel, Revision: n.Revision, Kosten: 1000, Waehrung: "EUR"}

	if fehler := a.PruefeMitNachweis(context.Background(), domain.WirkungExternSchreiben, "n-1", bindung); fehler != nil {
		t.Fatalf("mit passendem Nachweis: %v", fehler)
	}

	if fehler := a.PruefeMitNachweis(context.Background(), domain.WirkungExternSchreiben, "n-1", bindung); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("verbrauchter Nachweis: %v", fehler)
	}

	// **Eine Wirkung, die die Rolle nicht hergibt, scheitert vor dem
	// Nachweis** — der bleibt dann unverbraucht.
	nutzer.nutzerzeiger = 2
	freigaben.abgelegt = []domain.Profilfreigabe{{ProfilID: 7, NutzerID: 2, Rolle: domain.RolleMitarbeiter}}
	zweiter := mussAusstellen(t, ablage)
	ablage.abgelegt["n-2"] = func() domain.Nachweis { z := zweiter; z.Kennung = "n-2"; z.NutzerID = 2; return z }()

	if fehler := a.PruefeMitNachweis(context.Background(), domain.WirkungExternSchreiben, "n-2", bindung); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Fatalf("Mitarbeiter schreibt draußen: %v", fehler)
	}

	if ablage.abgelegt["n-2"].Eingeloest {
		t.Error("der Nachweis wurde verbraucht, obwohl die Rolle nicht reichte")
	}

	// **Ein Aufruf ohne Arbeitsprofil scheitert vor allem anderen.**
	nutzer.arbeitsprofilfehler = errors.New("kein Profil")
	if fehler := a.PruefeMitNachweis(context.Background(), domain.WirkungSchreiben, "", bindung); fehler == nil || !strings.Contains(fehler.Error(), "kein Profil") {
		t.Errorf("ohne Profil: %v", fehler)
	}

	nutzer.arbeitsprofilfehler = nil

	// **Ohne Nachweise-Dienst gibt es keinen Nachweis** — die Absage sagt es.
	ohne := autorisierung(nutzer, freigaben)
	if fehler := ohne.PruefeMitNachweis(context.Background(), domain.WirkungSchreiben, "", bindung); fehler != nil {
		t.Errorf("Schreiben ohne Nachweisdienst: %v", fehler)
	}

	nutzer.nutzerzeiger = 1
	if fehler := ohne.PruefeMitNachweis(context.Background(), domain.WirkungExternSchreiben, "n-1", bindung); !errors.Is(fehler, domain.ErrErlaubnis) {
		t.Errorf("draußen ohne Nachweisdienst: %v", fehler)
	}
}

// mussAusstellen stellt einen Nachweis des Verwalters am Fenster aus — und
// bricht ab, wenn das nicht geht.
func mussAusstellen(t *testing.T, ablage *nachweisdoppel) domain.Nachweis {
	t.Helper()

	n, fehler := nachweise(ablage).Ausstellen(context.Background(), verwalterAmFenster(), antrag())
	if fehler != nil {
		t.Fatalf("Ausstellen: %v", fehler)
	}

	return n
}
