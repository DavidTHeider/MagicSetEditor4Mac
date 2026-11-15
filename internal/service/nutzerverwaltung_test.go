package service_test

// **Nutzer und Profile verwalten** — Werkstück 1 von Phase 4.12
// (#1351),
// nach Vermerk 0066.

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein krummer Name wird am Feld abgewiesen** — und zwar hier, nicht erst im
// Bestand.
//
// Der `CHECK` in der Wanderung ist die zweite Reihe; er gäbe einen
// SQLite-Fehler, und der nennt kein Feld (§8).
func TestEinKrummerNameWirdImDienstAbgewiesen(t *testing.T) {
	t.Parallel()

	verwaltung := service.NeueNutzerverwaltung(&nutzerdoppel{})

	faelle := map[string]domain.Nutzer{
		"leer":         {Name: "  "},
		"mit Rändern":  {Name: " David "},
		"krummes Bild": {Name: "David", Bild: "../heimlich.png"},
	}

	for name, nutzer := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var feld *domain.FieldError
			if _, err := verwaltung.NutzerAnlegen(t.Context(), nutzer); !errors.As(
				err, &feld) {
				t.Errorf("Fehler ist %v", err)
			}
		})
	}
}

// **Ein gültiger Nutzer erreicht den Bestand.**
//
// Die Gegenprobe zum Lauf darüber: ohne sie wäre eine Prüfung, die **alles**
// abweist, ebenso grün.
func TestEinGueltigerNutzerErreichtDenBestand(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{}

	if _, err := service.NeueNutzerverwaltung(doppel).NutzerAnlegen(
		t.Context(), domain.Nutzer{Name: "Zweite"}); err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if len(doppel.angelegt) != 1 || doppel.angelegt[0].Name != "Zweite" {
		t.Errorf("angelegt: %+v", doppel.angelegt)
	}
}

// **Ein krummes Profil ebenso.**
func TestEinKrummesProfilWirdImDienstAbgewiesen(t *testing.T) {
	t.Parallel()

	var feld *domain.FieldError
	if _, err := service.NeueNutzerverwaltung(&nutzerdoppel{}).ProfilAnlegen(
		t.Context(), domain.Profil{NutzerID: 1, Name: " "}); !errors.As(err, &feld) {
		t.Errorf("Fehler ist %v", err)
	}
}

// **Die Übersicht nennt zu jedem Nutzer seine Profile.**
//
// **Ein Ruf und nicht zwei je Nutzer**: die Profilwahl zeigt beides
// nebeneinander, und wer sie einzeln holte, hätte für jeden Nutzer eine
// eigene Antwort — und damit Stände, die auseinanderlaufen können (§7,
// Regel 2).
func TestDieUebersichtNenntZuJedemNutzerSeineProfile(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Eine"}, {ID: 2, Name: "Zweite"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 10, NutzerID: 1, Name: "Mein Profil"}},
			2: {{ID: 20, NutzerID: 2, Name: "Arbeit"}, {ID: 21, NutzerID: 2, Name: "Turnier"}},
		},
	}

	uebersicht, err := service.NeueNutzerverwaltung(doppel).Uebersicht(t.Context())
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(uebersicht) != 2 {
		t.Fatalf("%d Zeilen: %+v", len(uebersicht), uebersicht)
	}

	if len(uebersicht[1].Profile) != 2 {
		t.Errorf("die zweite hat %d Profile", len(uebersicht[1].Profile))
	}
}

// **Ein Lesefehler bei den Profilen reißt die Übersicht nicht ab.**
//
// Er wird gemeldet. Eine halbe Übersicht, die aussieht wie eine ganze, wäre
// schlimmer: wer einen Nutzer ohne Profile sieht, hält ihn für leer.
func TestEinLesefehlerBeiDenProfilenWirdGemeldet(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{
		nutzer:       []domain.Nutzer{{ID: 1, Name: "Eine"}},
		profilfehler: errors.New("die Platte schweigt"),
	}

	if _, err := service.NeueNutzerverwaltung(doppel).Uebersicht(t.Context()); err == nil {
		t.Error("der Fehler ging verloren")
	}
}

// nutzerdoppel ist eine Nutzerablage im Arbeitsspeicher.
type nutzerdoppel struct {
	nutzer   []domain.Nutzer
	profile  map[int64][]domain.Profil
	angelegt []domain.Nutzer

	fehler       error
	profilfehler error

	// **Die Revision der Rechte und das Arbeitsprofil** — Werkstück O1 von
	// Phase 5.2 (#1425).
	revision            int64
	revisionsfehler     error
	arbeitsprofil       int64
	arbeitsprofilfehler error

	// **Ändern und Löschen** — Werkstück 5 (#1351).
	geaendert       domain.Nutzer
	profilgeandert  domain.Profil
	geloescht       int64
	profilgeloescht int64
	pflegefehler    error

	// **Der Zeiger und was an ihm geschieht** — Werkstück 3 (#1351).
	zeiger int64

	// nutzerzeiger ist, wer arbeitet — 0 heißt: der Eigner des gezeigten
	// Profils, wie im Bestand ohne Eintrag.
	nutzerzeiger int64
	gesetzt      int
	zeigerfehler error
	setzfehler   error

	// **Ein eigener Fehler für den Nutzerzeiger** — Werkstück 17 von Phase
	// 4.12.
	//
	// **Getrennt von [nutzerdoppel.zeigerfehler]**, weil die beiden Zeiger
	// getrennt gelesen werden und `Wechseln` den einen **vor** dem anderen
	// liest. Ein gemeinsamer Fehler ließe den zweiten Zweig unbetreten — und
	// genau so ist er es beim ersten Anlauf gewesen.
	nutzerzeigerfehler error

	// **Ein Fehler ab dem zweiten Ruf von [nutzerdoppel.Nutzer]** — siehe
	// dort.
	nachErlaubnis error
	gefragt       int
}

func (d *nutzerdoppel) NutzerAendern(_ context.Context, geaendert domain.Nutzer) (
	domain.Nutzer, error,
) {
	if d.pflegefehler != nil {
		return domain.Nutzer{}, d.pflegefehler
	}

	d.geaendert = geaendert

	return geaendert, nil
}

func (d *nutzerdoppel) ProfilAendern(_ context.Context, geaendert domain.Profil) (
	domain.Profil, error,
) {
	if d.pflegefehler != nil {
		return domain.Profil{}, d.pflegefehler
	}

	d.profilgeandert = geaendert

	return geaendert, nil
}

func (d *nutzerdoppel) NutzerLoeschen(_ context.Context, nutzerID int64, _ string,
	_ ...func([]domain.Profil, [][]domain.Betroffenes) error,
) error {
	if d.pflegefehler != nil {
		return d.pflegefehler
	}

	d.geloescht = nutzerID

	return nil
}

// ProfilLoeschen löscht **und** stellt den Zeiger — wie die echte Ablage.
//
// **Beides in einem Zug** (Befund 1 des Zweitblicks über die Phasen 4.0 bis
// 4.13): dort hält es eine Transaktion, hier diese Reihenfolge. Ein Doppel,
// das nur löschte, ließe den Lauf über den Zeiger ins Leere greifen.
func (d *nutzerdoppel) ProfilLoeschen(
	_ context.Context, profilID int64, _ string, ausweich int64,
	_ ...func([]domain.Profil, [][]domain.Betroffenes) error,
) error {
	if d.pflegefehler != nil {
		return d.pflegefehler
	}

	d.profilgeloescht = profilID

	if ausweich != 0 {
		d.zeiger = ausweich
	}

	return nil
}

func (d *nutzerdoppel) Profilzeiger(context.Context) (int64, error) {
	return d.zeiger, d.zeigerfehler
}

func (d *nutzerdoppel) ProfilzeigerSetzen(_ context.Context, profilID int64) error {
	if d.setzfehler != nil {
		return d.setzfehler
	}

	d.gesetzt++
	d.zeiger = profilID

	return nil
}

// Nutzerzeiger sagt, wer arbeitet — Werkstück 17 von Phase 4.12.
//
// **Der Rückfall ist derselbe wie im Bestand**: ohne gesetzten Zeiger gilt der
// Eigner des gezeigten Profils. Ohne ihn prüfte jeder Lauf, der ihn nicht
// eigens setzt, eine Lage, die es im Betrieb nicht gibt.
func (d *nutzerdoppel) Nutzerzeiger(context.Context) (int64, error) {
	// **Nur der eigene Fehler, nicht der des Profilzeigers.**
	//
	// **Sonst bliebe ein Zweig unbetreten**: `Wechseln` liest beide Zeiger,
	// den Nutzerzeiger zuerst. Wer hier auch `zeigerfehler` beantwortete,
	// ließe den Ruf nie bis zum Profilzeiger kommen — und dessen Fehlerweg
	// wäre von außen unerreichbar (§3). Gefunden an der Abdeckung.
	if d.nutzerzeigerfehler != nil {
		return 0, d.nutzerzeigerfehler
	}

	if d.nutzerzeiger != 0 {
		return d.nutzerzeiger, nil
	}

	for nutzerID, profile := range d.profile {
		for _, profil := range profile {
			if profil.ID == d.zeiger {
				return nutzerID, nil
			}
		}
	}

	return 0, nil
}

func (d *nutzerdoppel) NutzerzeigerSetzen(_ context.Context, nutzerID int64) error {
	if d.setzfehler != nil {
		return d.setzfehler
	}

	d.nutzerzeiger = nutzerID

	return nil
}

// ZeigerSetzen stellt beide Zeiger — **oder keinen**.
//
// **Die Ganzheit ist die Zusage** (Befund 1 des Zweitblicks über die Phasen
// 4.0 bis 4.13): im Bestand hält sie eine Transaktion, hier hält sie diese
// Reihenfolge — erst prüfen, dann beide setzen.
func (d *nutzerdoppel) ZeigerSetzen(_ context.Context, nutzerID, profilID int64) error {
	if d.setzfehler != nil {
		return d.setzfehler
	}

	d.nutzerzeiger = nutzerID
	d.zeiger = profilID

	return nil
}

func (d *nutzerdoppel) Nutzer(context.Context) ([]domain.Nutzer, error) {
	// **Ein Fehler, der erst beim zweiten Ruf kommt** — Werkstück 12 (#1351).
	//
	// **Wofür**: die Deckübernahme fragt die Übersicht **zweimal** — einmal
	// für die Erlaubnis, einmal für den Namen des Quellprofils. Ein Fehler,
	// der schon beim ersten kommt, prüft den ersten; der zweite Weg bliebe
	// unbetreten (§10).

	d.gefragt++

	if d.gefragt > 1 && d.nachErlaubnis != nil {
		return nil, d.nachErlaubnis
	}

	return d.nutzer, d.fehler
}

func (d *nutzerdoppel) Profile(_ context.Context, nutzerID int64) (
	[]domain.Profil, error,
) {
	if d.profilfehler != nil {
		return nil, d.profilfehler
	}

	return d.profile[nutzerID], nil
}

func (d *nutzerdoppel) NutzerAnlegen(_ context.Context, neu domain.Nutzer) (
	domain.Nutzer, error,
) {
	if d.fehler != nil {
		return domain.Nutzer{}, d.fehler
	}

	neu.ID = int64(len(d.angelegt) + 1)
	d.angelegt = append(d.angelegt, neu)

	// **Und ein erstes Profil**, wie der Speicher es in derselben Transaktion
	// anlegt. Ein Doppel, das den Nutzer ohne eines gibt, wäre großzügiger als
	// der Betrieb (§3).
	if d.profile == nil {
		d.profile = map[int64][]domain.Profil{}
	}

	d.profile[neu.ID] = append(d.profile[neu.ID],
		domain.Profil{ID: neu.ID * 100, NutzerID: neu.ID, Name: "Standard"})

	return neu, nil
}

func (d *nutzerdoppel) ProfilAnlegen(_ context.Context, neu domain.Profil) (
	domain.Profil, error,
) {
	if d.fehler != nil {
		return domain.Profil{}, d.fehler
	}

	neu.ID = 99

	return neu, nil
}

// **Ein Lesefehler bei den Nutzern reicht durch.**
func TestEinLesefehlerBeiDenNutzernWirdGemeldet(t *testing.T) {
	t.Parallel()

	if _, err := service.NeueNutzerverwaltung(
		&nutzerdoppel{fehler: errors.New("die Platte schweigt")},
	).Uebersicht(t.Context()); err == nil {
		t.Error("der Fehler ging verloren")
	}
}

// **Ein gültiges Profil erreicht den Bestand.**
func TestEinGueltigesProfilErreichtDenBestand(t *testing.T) {
	t.Parallel()

	angelegt, err := service.NeueNutzerverwaltung(&nutzerdoppel{}).ProfilAnlegen(
		t.Context(), domain.Profil{NutzerID: 1, Name: "Turnier"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if angelegt.ID == 0 {
		t.Error("das Profil hat keine Kennung")
	}
}
