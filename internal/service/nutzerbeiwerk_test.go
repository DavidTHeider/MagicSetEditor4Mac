// **Was neben dem Bestand an einem Nutzer hängt** — Befund 2 des Zweitblicks
// über die Phasen 4.0 bis 4.13 (2026-09-07).
//
// **Aus `nutzerpflege_test.go` herausgelöst**, als die Datei über 400 Zeilen
// wuchs (`AGENTS.md` §2). Der Schnitt folgt der Frage: dort steht, was mit
// **der Zeile** geschieht — Namensprüfung, „der letzte bleibt", die
// Fehlerwege der Ablage; hier, was mit dem geschieht, das **daneben** liegt:
// die Bilddatei im Anwendungsordner und das TOTP-Geheimnis im Schlüsselbund.
package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein gelöschter Nutzer nimmt sein Bild und sein Geheimnis mit.**
//
// # Der Befund
//
// **Zweitblick über die Phasen 4.0 bis 4.13** (2026-09-07): `NutzerLoeschen`
// rief die Ablage und sonst nichts — die löscht eine Datenbankzeile. Das
// Nutzerbild lag als Datei daneben, das TOTP-Geheimnis im Schlüsselbund, und
// beides blieb liegen. Die Datenschutzhilfe versprach derweil „samt allem, was
// an ihm hängt".
//
// **Ein persönliches Foto und ein Geheimnis, die eine Löschung überleben,
// sind genau das, wogegen es diese Zusage gibt.**
func TestEinGeloeschterNutzerNimmtSeinBeiwerkMit(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{nutzer: []domain.Nutzer{{ID: 3, Name: "Wer"}}}
	bilder := &bilddoppel{}
	geheimnisse := &geheimnisdoppel{}

	verwaltung := service.NeueNutzerverwaltung(doppel)
	verwaltung.MitBeiwerk(bilder, geheimnisse)

	if fehler := verwaltung.NutzerLoeschen(t.Context(), 3, "Wer"); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if bilder.weggenommen != 3 {
		t.Errorf("das Bild von %d wurde weggenommen, erwartet war 3", bilder.weggenommen)
	}

	if geheimnisse.genommen != 3 {
		t.Errorf("das Geheimnis von %d wurde genommen, erwartet war 3",
			geheimnisse.genommen)
	}
}

// **Und ein Fehlschlag dabei macht das Löschen nicht rückgängig** (§7,
// Regel 3) — er sagt nur, dass etwas nachzusehen ist.
//
// # Drei Fälle, weil die Ursache eine Liste ist
//
// **Geprüft war bis zum 2026-09-07 nur das Bild.** Der Zweig des Geheimnisses
// blieb unbetreten, und der Fall, für den es die Liste überhaupt gibt — beide
// scheitern —, ebenso. Eine Ursache, die zwei Dinge nennen können soll, hat
// nie zwei genannt (`AGENTS.md` §10).
//
// # Wo nachgesehen wird, und warum nicht in der Meldung
//
// **Der Satz für Menschen nennt die Ursachen nicht**, und das ist Werkstück 1
// von Phase 4.13: `domain.MitGrund` hängt den Grund in die **Kette** und
// lässt ihn aus `Error()` heraus, weil Kommandozeile, MCP und REST keinen
// Katalog haben und jeden Satz unverändert zeigen (§8).
//
// **Geprüft wird deshalb an [domain.GrundVon]** — dort steht, was
// `internal/app` ins Protokoll schreibt. Der erste Anlauf sah in `Error()`
// nach und wurde rot; er prüfte einen Vertrag, den es aus gutem Grund nicht
// gibt.
func TestEinLiegengebliebenesBeiwerkWirdGemeldet(t *testing.T) {
	t.Parallel()

	klemmt := errors.New("die Platte klemmt")
	zu := errors.New("der Schlüsselbund ist zu")

	for _, fall := range []struct {
		name            string
		bild, geheimnis error
		nennt           []string
	}{
		{"nur das Bild", klemmt, nil, []string{"die Platte klemmt"}},
		{"nur das Geheimnis", nil, zu, []string{"der Schlüsselbund ist zu"}},
		{
			"beide, und beide werden genannt", klemmt, zu,
			[]string{"die Platte klemmt", "der Schlüsselbund ist zu"},
		},
	} {
		t.Run(fall.name, func(t *testing.T) {
			t.Parallel()

			doppel := &nutzerdoppel{nutzer: []domain.Nutzer{{ID: 3, Name: "Wer"}}}

			verwaltung := service.NeueNutzerverwaltung(doppel)
			verwaltung.MitBeiwerk(
				&bilddoppel{fehler: fall.bild},
				&geheimnisdoppel{fehler: fall.geheimnis})

			fehler := verwaltung.NutzerLoeschen(t.Context(), 3, "Wer")
			if fehler == nil {
				t.Fatal("das liegengebliebene Beiwerk wurde verschwiegen")
			}

			grund := domain.GrundVon(fehler)
			if grund == nil {
				t.Fatalf("%v trägt keinen Grund — das Protokoll bliebe stumm",
					fehler)
			}

			for _, teil := range fall.nennt {
				if !strings.Contains(grund.Error(), teil) {
					t.Errorf("der Grund lautet %q und nennt %q nicht",
						grund.Error(), teil)
				}
			}

			// **Und der Satz für Menschen bleibt sauber** — die Zusage aus
			// Werkstück 1 von Phase 4.13, hier an ihrem jüngsten Aufrufer.
			if strings.Contains(fehler.Error(), fall.nennt[0]) {
				t.Errorf("die Meldung %q trägt die technische Ursache",
					fehler.Error())
			}

			// **Der Nutzer ist trotzdem fort.**
			if doppel.geloescht != 3 {
				t.Error("das Löschen wurde zurückgenommen")
			}
		})
	}
}

// **Und die Abzeichen seiner Profile gehen mit** — Nachlauf zum Zweitblick
// über die Phasen 4.0 bis 4.13 (2026-09-07), Befund 2.
//
// # Warum das derselbe Befund ist und nicht ein neuer
//
// **Ein Abzeichen ist ein Bild neben dem Bestand, genau wie das Nutzerfoto.**
// Es liegt im selben Verzeichnis, unter `profil-<id>.png`. Die erste Fassung
// der Beiwerknaht kannte nur das Foto — die Klasse war benannt und zur Hälfte
// geschlossen, und der Nachlauf hat genau das angestrichen.
//
// **Gelesen wird die Profilliste vor dem Löschen**, weil sie danach leer ist:
// die Kaskade räumt die Zeilen mit. Dieser Test hält beides — dass jedes
// Abzeichen drankommt und dass keines vergessen wird.
func TestEinGeloeschterNutzerNimmtDieAbzeichenSeinerProfileMit(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{
		nutzer: []domain.Nutzer{{ID: 3, Name: "Wer"}},
		profile: map[int64][]domain.Profil{
			3: {
				{ID: 7, NutzerID: 3, Name: "Erstes"},
				{ID: 9, NutzerID: 3, Name: "Turnier"},
			},
		},
	}

	bilder := &bilddoppel{}

	verwaltung := service.NeueNutzerverwaltung(doppel)
	verwaltung.MitBeiwerk(bilder, &geheimnisdoppel{})

	if fehler := verwaltung.NutzerLoeschen(t.Context(), 3, "Wer"); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if len(bilder.abzeichen) != 2 ||
		bilder.abzeichen[0] != 7 || bilder.abzeichen[1] != 9 {
		t.Errorf("weggenommen wurden die Abzeichen %v, erwartet waren 7 und 9",
			bilder.abzeichen)
	}
}

// **Und ein liegengebliebenes Abzeichen wird genannt, mit seiner Kennung.**
//
// **Die Kennung gehört dazu**: „ein Abzeichen ließ sich nicht wegräumen"
// schickt jemanden in ein Verzeichnis mit vielen Dateien. Der Grund fürs
// Protokoll nennt die Zahl (§8).
func TestEinLiegengebliebenesAbzeichenNenntSeinProfil(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{
		nutzer:  []domain.Nutzer{{ID: 3, Name: "Wer"}},
		profile: map[int64][]domain.Profil{3: {{ID: 9, NutzerID: 3, Name: "X"}}},
	}

	verwaltung := service.NeueNutzerverwaltung(doppel)
	verwaltung.MitBeiwerk(
		&bilddoppel{abzeichenfehler: errors.New("die Platte klemmt")},
		&geheimnisdoppel{})

	fehler := verwaltung.NutzerLoeschen(t.Context(), 3, "Wer")
	if fehler == nil {
		t.Fatal("das liegengebliebene Abzeichen wurde verschwiegen")
	}

	grund := domain.GrundVon(fehler)
	if grund == nil || !strings.Contains(grund.Error(), "Profil 9") {
		t.Errorf("der Grund lautet %v und nennt Profil 9 nicht", grund)
	}
}

// **Und ein gelöschtes Profil nimmt sein eigenes Abzeichen mit.**
func TestEinGeloeschtesProfilNimmtSeinAbzeichenMit(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{
		nutzer:  []domain.Nutzer{{ID: 3, Name: "Wer"}},
		profile: map[int64][]domain.Profil{3: {{ID: 9, NutzerID: 3, Name: "X"}}},
		zeiger:  1,
	}

	bilder := &bilddoppel{}

	verwaltung := service.NeueNutzerverwaltung(doppel)
	verwaltung.MitBeiwerk(bilder, &geheimnisdoppel{})

	if fehler := verwaltung.ProfilLoeschen(t.Context(), 9, "X"); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	if len(bilder.abzeichen) != 1 || bilder.abzeichen[0] != 9 {
		t.Errorf("weggenommen wurden %v, erwartet war allein 9", bilder.abzeichen)
	}
}

// **Und ein Profil, dessen Abzeichen liegen bleibt, sagt es.**
func TestEinLiegengebliebenesAbzeichenEinesProfilsWirdGemeldet(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{
		nutzer:  []domain.Nutzer{{ID: 3, Name: "Wer"}},
		profile: map[int64][]domain.Profil{3: {{ID: 9, NutzerID: 3, Name: "X"}}},
		zeiger:  1,
	}

	verwaltung := service.NeueNutzerverwaltung(doppel)
	verwaltung.MitBeiwerk(
		&bilddoppel{abzeichenfehler: errors.New("die Platte klemmt")},
		&geheimnisdoppel{})

	fehler := verwaltung.ProfilLoeschen(t.Context(), 9, "X")
	if fehler == nil {
		t.Fatal("das liegengebliebene Abzeichen wurde verschwiegen")
	}

	// **Das Profil ist trotzdem fort** (§7, Regel 3).
	if doppel.profilgeloescht != 9 {
		t.Error("das Löschen wurde zurückgenommen")
	}
}

// **Und ohne Beiwerknaht löscht der Dienst die Zeile und sonst nichts.**
//
// **Der Stand von vor dem Befund**, und keine Verschlechterung: eine Fläche,
// die nur liest, oder ein Lauf ohne Bilderverzeichnis baut die Verwaltung
// ohne sie. Der Zweig steht in beiden Löschwegen und wäre sonst unbetreten.
func TestOhneBeiwerknahtBleibtEsBeiDerZeile(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{
		nutzer:  []domain.Nutzer{{ID: 3, Name: "Wer"}},
		profile: map[int64][]domain.Profil{3: {{ID: 9, NutzerID: 3, Name: "X"}}},
		zeiger:  1,
	}

	verwaltung := service.NeueNutzerverwaltung(doppel)

	if fehler := verwaltung.ProfilLoeschen(t.Context(), 9, "X"); fehler != nil {
		t.Fatalf("Profil löschen: %v", fehler)
	}

	if fehler := verwaltung.NutzerLoeschen(t.Context(), 3, "Wer"); fehler != nil {
		t.Fatalf("Nutzer löschen: %v", fehler)
	}
}

// **Und ein Lesefehler bei den Profilen hält das Löschen nicht auf.**
//
// **Er wird gemeldet wie ein liegengebliebenes Bild**: der Mensch will
// gelöscht werden, und ein Verzeichnis, das gerade klemmt, ist kein Grund, ihn
// zu behalten.
func TestEinLesefehlerBeiDenProfilenHaeltDasLoeschenNichtAuf(t *testing.T) {
	t.Parallel()

	doppel := &nutzerdoppel{
		nutzer:       []domain.Nutzer{{ID: 3, Name: "Wer"}},
		profilfehler: errors.New("die Profile sind nicht zu lesen"),
	}

	verwaltung := service.NeueNutzerverwaltung(doppel)
	verwaltung.MitBeiwerk(&bilddoppel{}, &geheimnisdoppel{})

	fehler := verwaltung.NutzerLoeschen(t.Context(), 3, "Wer")
	if fehler == nil {
		t.Fatal("der Lesefehler wurde verschwiegen")
	}

	if doppel.geloescht != 3 {
		t.Error("das Löschen wurde zurückgenommen")
	}

	grund := domain.GrundVon(fehler)
	if grund == nil || !strings.Contains(grund.Error(), "Abzeichen bleiben liegen") {
		t.Errorf("der Grund lautet %v und sagt nicht, was liegen bleibt", grund)
	}
}

// bilddoppel merkt sich, wessen Bild und wessen Abzeichen weggenommen wurde.
type bilddoppel struct {
	weggenommen int64

	// abzeichen sammelt die Profilkennungen in der Reihenfolge der Rufe.
	abzeichen []int64

	fehler          error
	abzeichenfehler error
}

func (b *bilddoppel) NutzerbildWegnehmen(nutzerID int64) error {
	b.weggenommen = nutzerID

	return b.fehler
}

// **Tisch- und Fensterbild prüft der Lauf über die echte Ablage**
// (`nutzerhintergruende_test.go`); das Doppel nimmt sie nur entgegen.
func (b *bilddoppel) TischbildWegnehmen(int64) error { return nil }

func (b *bilddoppel) FensterbildWegnehmen(int64) error { return nil }

func (b *bilddoppel) ProfilbadgeWegnehmen(profilID int64) error {
	b.abzeichen = append(b.abzeichen, profilID)

	return b.abzeichenfehler
}

// geheimnisdoppel merkt sich, wessen Geheimnis genommen wurde.
type geheimnisdoppel struct {
	genommen int64
	fehler   error
}

func (g *geheimnisdoppel) Nimm(_ context.Context, nutzerID int64) error {
	g.genommen = nutzerID

	return g.fehler
}
