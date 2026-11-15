package service_test

// **Der Profilwechsel** — Werkstück 3 von Phase 4.12
// (#1351).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Wechsel setzt den Zeiger und gibt das Profil zurück.**
//
// **Das Profil und nicht nur ein „gut"**: der Rufer schreibt es sich an die
// Fläche — das Fenster in seinen Zustand, die Kommandozeile in ihre Ausgabe.
// Wer es danach erst holte, hätte zwei Rufe und zwei Stände (§7, Regel 2).
func TestEinWechselSetztDenZeigerUndNenntDasProfil(t *testing.T) {
	ablage := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{
			1: {{ID: 1, NutzerID: 1, Name: "Erstes"}, {ID: 2, NutzerID: 1, Name: "Zweites"}},
		},
		zeiger: 1,
	}

	profil, fehler := service.NeueNutzerverwaltung(ablage).Wechseln(t.Context(), 2, "")
	if fehler != nil {
		t.Fatalf("wechseln: %v", fehler)
	}

	if profil.Name != "Zweites" {
		t.Errorf("Profil %q, erwartet war „Zweites", profil.Name)
	}

	if ablage.zeiger != 2 {
		t.Errorf("der Zeiger steht auf %d", ablage.zeiger)
	}
}

// **Ein Wechsel auf das Profil, das schon gezeigt wird, ist kein Fehler.**
//
// Er ist auch kein Schreiben: ein Zeiger, der schon steht, noch einmal gesetzt
// zu werden, wäre ein Bestandsschreiben ohne Wirkung — und im Verlauf einer
// Beobachtbarkeit ein Ereignis, das nichts bedeutet.
func TestEinWechselAufDasGezeigteProfilSchreibtNicht(t *testing.T) {
	ablage := &nutzerdoppel{
		nutzer:  []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{1: {{ID: 1, NutzerID: 1, Name: "Erstes"}}},
		zeiger:  1,
	}

	if _, fehler := service.NeueNutzerverwaltung(ablage).Wechseln(t.Context(), 1, ""); fehler != nil {
		t.Fatalf("wechseln: %v", fehler)
	}

	if ablage.gesetzt != 0 {
		t.Errorf("%d Schreibvorgänge, erwartet war keiner", ablage.gesetzt)
	}
}

// **Ein Profil, das es nicht gibt, ist ein Nein.**
func TestEinWechselAufEinFehlendesProfilWirdAbgewiesen(t *testing.T) {
	ablage := &nutzerdoppel{
		nutzer:  []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{1: {{ID: 1, NutzerID: 1, Name: "Erstes"}}},
		zeiger:  1,
	}

	_, fehler := service.NeueNutzerverwaltung(ablage).Wechseln(t.Context(), 4711, "")
	if fehler == nil {
		t.Fatal("das fehlende Profil kam durch")
	}

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Errorf("Fehler %v, erwartet war ErrNotFound", fehler)
	}
}

// **Das gezeigte Profil lässt sich erfragen.**
func TestDasGezeigteProfilLaesstSichErfragen(t *testing.T) {
	ablage := &nutzerdoppel{
		nutzer:  []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{1: {{ID: 7, NutzerID: 1, Name: "Eines"}}},
		zeiger:  7,
	}

	profil, fehler := service.NeueNutzerverwaltung(ablage).Gezeigt(t.Context())
	if fehler != nil {
		t.Fatalf("erfragen: %v", fehler)
	}

	if profil.ID != 7 {
		t.Errorf("Profil %d, erwartet war 7", profil.ID)
	}
}

// **Das Arbeitsprofil ist das des Rufs — nicht das des Zeigers.**
//
// # Warum es diese Frage zusätzlich zu [Nutzerverwaltung.Gezeigt] gibt
//
// **Weil ein Ruf mit Token nicht im gezeigten Profil arbeitet** — Werkstück
// A14 von Phase 5.3
// (#1453).
// `app.Tokenwache` hängt das Profil des Tokens an den Kontext; jede Abfrage
// des Bestands liest von dort. Wer daneben den **Zeiger** nennte, sagte einem
// Agenten ein Profil, in dem seine eigenen Antworten nicht stehen.
func TestDasArbeitsprofilIstDasDesRufs(t *testing.T) {
	t.Parallel()

	ablage := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}},
		profile: map[int64][]domain.Profil{1: {
			{ID: 7, NutzerID: 1, Name: "Gezeigtes"},
			{ID: 9, NutzerID: 1, Name: "Des Tokens"},
		}},
		zeiger:        7,
		arbeitsprofil: 9,
	}

	profil, fehler := service.NeueNutzerverwaltung(ablage).Arbeitsprofil(t.Context())
	if fehler != nil {
		t.Fatalf("erfragen: %v", fehler)
	}

	if profil.ID != 9 {
		t.Errorf("Profil %d, erwartet war 9 — der Zeiger steht auf 7", profil.ID)
	}
}

// **Und ein unlesbares Arbeitsprofil reicht durch.**
func TestEinUnlesbaresArbeitsprofilReichtDurch(t *testing.T) {
	t.Parallel()

	ablage := &nutzerdoppel{arbeitsprofilfehler: errors.New("kaputt")}

	if _, fehler := service.NeueNutzerverwaltung(ablage).Arbeitsprofil(
		t.Context()); fehler == nil {
		t.Fatal("der Fehler kam nicht durch")
	}
}

// **Jeder Fehlerweg des Wechsels reicht durch.**
//
// **Vier und nicht drei**: der Zeiger wird **gelesen** und **geschrieben**,
// und beides kann scheitern. Der erste Anlauf hatte nur das Schreiben — und
// ein Lesefehler wäre still als „steht schon" durchgegangen.
func TestJederFehlerwegDesWechselsReichtDurch(t *testing.T) {
	welt := func() *nutzerdoppel {
		return &nutzerdoppel{
			nutzer:  []domain.Nutzer{{ID: 1, Name: "Wer"}},
			profile: map[int64][]domain.Profil{1: {{ID: 1, NutzerID: 1, Name: "Erstes"}, {ID: 2, NutzerID: 1, Name: "Zweites"}}},
			zeiger:  1,
		}
	}

	faelle := map[string]func(*nutzerdoppel){
		"die Übersicht scheitert": func(d *nutzerdoppel) { d.fehler = errors.New("die Platte schweigt") },
		"die Profile scheitern":   func(d *nutzerdoppel) { d.profilfehler = errors.New("nichts zu lesen") },
		"der Zeiger scheitert":    func(d *nutzerdoppel) { d.zeigerfehler = errors.New("kein Zeiger") },
		"das Setzen scheitert":    func(d *nutzerdoppel) { d.setzfehler = errors.New("kein Schreiben") },
	}

	for name, stoere := range faelle {
		t.Run(name, func(t *testing.T) {
			ablage := welt()
			stoere(ablage)

			if _, fehler := service.NeueNutzerverwaltung(ablage).Wechseln(
				t.Context(), 2, ""); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Und die Frage nach dem gezeigten Profil ebenso.**
func TestJederFehlerwegDerFrageReichtDurch(t *testing.T) {
	faelle := map[string]*nutzerdoppel{
		"der Zeiger scheitert": {zeigerfehler: errors.New("kein Zeiger")},
		"die Übersicht scheitert": {
			zeiger: 1,
			fehler: errors.New("die Platte schweigt"),
		},
	}

	for name, ablage := range faelle {
		t.Run(name, func(t *testing.T) {
			if _, fehler := service.NeueNutzerverwaltung(ablage).Gezeigt(
				t.Context()); fehler == nil {
				t.Error("der Fehler ging verloren")
			}
		})
	}
}

// **Ein angelegter Nutzer nennt sein erstes Profil.**
//
// # Warum das hier steht und nicht bei den Anlagewegen
//
// **Weil es der Wechsel gebraucht hat.** Bis zum 2026-09-04 gab
// [service.Nutzerverwaltung.NutzerAnlegen] den Nutzer **ohne** Profile
// zurück — der Speicher legt eines an, in derselben Transaktion, und niemand
// erfuhr seine Kennung. Wer danach hinwechseln wollte, musste die ganze
// Übersicht neu holen: zwei Rufe, zwei Stände (§7, Regel 2).
//
// Die leere Liste war dabei nicht bloß unbequem, sondern **falsch**: sie
// behauptet, der Nutzer habe kein Profil. Genau das kann er nicht.
func TestEinAngelegterNutzerNenntSeinErstesProfil(t *testing.T) {
	ablage := &nutzerdoppel{nutzer: []domain.Nutzer{{ID: 1, Name: "Wer"}}}

	zeile, fehler := service.NeueNutzerverwaltung(ablage).NutzerAnlegen(
		t.Context(), domain.Nutzer{Name: "Zweiter"})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if len(zeile.Profile) != 1 {
		t.Fatalf("%d Profile, erwartet war eines", len(zeile.Profile))
	}

	if zeile.Profile[0].ID == 0 {
		t.Error("das erste Profil hat keine Kennung")
	}
}

// **Ein Lesefehler nach dem Anlegen reißt ab.**
//
// Eine halbe Antwort — der Nutzer da, seine Profile nicht — sähe aus wie ein
// Nutzer ohne Arbeitsumgebung, und den kann es nicht geben (§8).
func TestEinLesefehlerNachDemAnlegenReisstAb(t *testing.T) {
	ablage := &nutzerdoppel{profilfehler: errors.New("nichts zu lesen")}

	if _, fehler := service.NeueNutzerverwaltung(ablage).NutzerAnlegen(
		t.Context(), domain.Nutzer{Name: "Zweiter"}); fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}

// **Und ein Schreibfehler beim Anlegen ebenso.**
func TestEinSchreibfehlerBeimAnlegenReichtDurch(t *testing.T) {
	ablage := &nutzerdoppel{fehler: errors.New("die Platte schweigt")}

	if _, fehler := service.NeueNutzerverwaltung(ablage).NutzerAnlegen(
		t.Context(), domain.Nutzer{Name: "Zweiter"}); fehler == nil {
		t.Error("der Fehler ging verloren")
	}
}
