package app

// **Die Fehlerwege des Umschlags** — Werkstück N4 von Phase 5.4
// (#1457).
//
// # Warum sie einen eigenen Lauf bekommen
//
// **Weil sie sonst niemand betritt.** Der Abnahmelauf daneben fährt den Weg,
// der glückt; §10 sagt dazu: „wer nur den Erfolgsweg prüft, hat die Hälfte
// ausgelassen." Und jeder dieser Wege endet in einem Satz, den ein Mensch
// liest — ein `null` darin wäre eine leere Meldung.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Eine Kennung, die keine Zahl ist, kommt nicht durch.**
func TestEineUngueltigeSetkennungWirdAbgewiesen(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	antwort := anwendung.UmschlagSchnueren(t.Context(), "", "keine Zahl",
		filepath.Join(t.TempDir(), "x.msepackage"))

	if antwort.Fehler == nil {
		t.Fatal("die Kennung kam durch")
	}

	if antwort.Objekte == nil {
		t.Error("die Liste ist nil")
	}
}

// **Ohne Identität wird nichts unterschrieben.**
//
// **Und das ist der Grund, warum der Ausweis vor dem Schreiben steht**: eine
// halbe Datei am Ziel wäre schlimmer als gar keine.
func TestOhnePeeridentitaetWirdNichtsGeschnuert(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	ctx := t.Context()

	geoeffnet := anwendung.SetOeffnen(ctx, saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "ohne.msepackage")

	antwort := anwendung.UmschlagSchnueren(ctx, "", geoeffnet.SetID, ziel)
	if antwort.Fehler == nil {
		t.Fatal("ohne Identität wurde geschnürt")
	}

	if _, fehler := os.Stat(ziel); !errors.Is(fehler, os.ErrNotExist) {
		t.Error("am Ziel liegt etwas, obwohl nichts geschrieben werden konnte")
	}
}

// **Ein Set, das es nicht gibt, lässt sich nicht schnüren.**
func TestEinUnbekanntesSetLaesstSichNichtSchnueren(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	ctx := t.Context()

	if antwort := anwendung.PeeridentitaetSicherstellen(
		anwendung.imProfil(ctx)); antwort.Fehler != nil {
		t.Fatalf("Identität: %+v", antwort.Fehler)
	}

	antwort := anwendung.UmschlagSchnueren(ctx, "", "999999",
		filepath.Join(t.TempDir(), "x.msepackage"))
	if antwort.Fehler == nil {
		t.Fatal("ein Set ohne Bestand wurde geschnürt")
	}
}

// **Was kein Umschlag ist, kommt nicht herein.**
func TestEineFremdeDateiIstKeinUmschlag(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	fremd := filepath.Join(t.TempDir(), "fremd.msepackage")
	if fehler := os.WriteFile(fremd, []byte("kein ZIP"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	ctx := t.Context()

	angesehen := anwendung.UmschlagAnsehen(ctx, fremd)
	if angesehen.Fehler == nil {
		t.Error("eine fremde Datei wurde angesehen")
	}

	if angesehen.Objekte == nil {
		t.Error("die Liste ist nil")
	}

	// **Und auf demselben Weg auch beim Übernehmen** — sonst stünde hier ein
	// Wegwerfort mit fremdem Inhalt.
	uebernommen := anwendung.UmschlagUebernehmen(ctx, fremd)
	if uebernommen.Fehler == nil {
		t.Error("eine fremde Datei wurde übernommen")
	}

	if uebernommen.Befunde == nil || uebernommen.Angelegt == nil {
		t.Error("eine Liste der Antwort ist nil")
	}
}

// **Die drei Ausgänge tragen ihren Schlüssel.**
//
// **Der Schlüssel und nicht der Satz** ist das, woran das Fenster übersetzt;
// ohne ihn stünde dort ein deutscher Satz in einer englischen Oberfläche (§9).
func TestDieDreiUmschlagausgaengeTragenIhrenSchluessel(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		fehler     error
		schluessel domain.Schluessel
	}{
		"Fassung": {
			domain.ErrUmschlagFassung, domain.SchluesselUmschlagFassung,
		},
		"Unterschrift": {
			domain.ErrUmschlagUnterschrift, domain.SchluesselUmschlagUnterschrift,
		},
		"Inhalt": {
			domain.ErrUmschlagInhalt, domain.SchluesselUmschlagInhalt,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			uebersetzt := umschlagausgang(fall.fehler)
			if uebersetzt == nil {
				t.Fatal("kein Umschlagausgang erkannt")
			}

			if uebersetzt.Schluessel != fall.schluessel {
				t.Errorf("der Schlüssel ist %q", uebersetzt.Schluessel)
			}
		})
	}

	// **Die Gegenprobe**: was keiner der drei ist, bekommt hier auch keinen
	// Schlüssel — sonst trüge jeder Fehler einen falschen.
	if umschlagausgang(errors.New("irgendetwas")) != nil {
		t.Error("ein fremder Fehler wurde als Umschlagausgang gedeutet")
	}
}

// **Ein Umschlag mit einem Symlink darin nimmt ihn nicht mit.**
//
// **Gemessen und nicht behauptet**: `setEntfalten` überspringt alles, was
// keine gewöhnliche Datei ist — und #1457 verbietet Symlinks ausdrücklich.
func TestEinSymlinkGehtNichtMit(t *testing.T) {
	t.Parallel()

	ordner := t.TempDir()

	if fehler := os.WriteFile(
		filepath.Join(ordner, "echt.txt"), []byte("da"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if fehler := os.Symlink("/etc/passwd",
		filepath.Join(ordner, "zeiger.txt")); fehler != nil {
		t.Fatalf("Symlink: %v", fehler)
	}

	inhalte, fehler := setEntfalten(ordner)
	if fehler != nil {
		t.Fatalf("entfalten: %v", fehler)
	}

	if len(inhalte) != 1 || inhalte[0].Bezeichnung != "echt.txt" {
		t.Errorf("mitgekommen ist %+v", inhalte)
	}
}

// **Ein Ordner, den es nicht gibt, meldet sich statt still leer zu bleiben.**
func TestEinFehlenderSetordnerMeldetSich(t *testing.T) {
	t.Parallel()

	if _, fehler := setEntfalten(
		filepath.Join(t.TempDir(), "gibtesnicht")); fehler == nil {
		t.Error("ein fehlender Ordner kam als leerer Umschlag durch")
	}
}

// **Ohne Ort für die Zwischendateien wird nichts geschnürt und nichts
// gelesen.**
//
// # Warum das hier steht
//
// **Weil beide Wege einen Wegwerfort anlegen** — und ein Fehler dabei ist der
// einzige Zweig, den kein fachlicher Fall betritt. Ohne diesen Lauf stünde er
// da und niemand wüsste, ob er eine Meldung gibt oder einen leeren Umschlag.
//
// **Gemessen über `TMPDIR`**, denn [os.MkdirTemp] mit leerem ersten Argument
// fragt [os.TempDir], und das liest die Umgebung.
func TestOhneWegwerfortMeldenSichBeideWege(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	ctx := t.Context()

	geoeffnet := anwendung.SetOeffnen(ctx, saatset(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	if antwort := anwendung.PeeridentitaetSicherstellen(
		anwendung.imProfil(ctx)); antwort.Fehler != nil {
		t.Fatalf("Identität: %+v", antwort.Fehler)
	}

	ziel := filepath.Join(t.TempDir(), "ohneort.msepackage")

	// **Erst danach umgelenkt** — das Vorgeplänkel braucht einen echten Ort.
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "gibtesnicht"))

	if antwort := anwendung.UmschlagSchnueren(
		ctx, "", geoeffnet.SetID, ziel); antwort.Fehler == nil {
		t.Error("ohne Wegwerfort wurde geschnürt")
	}

	if antwort := anwendung.UmschlagAnsehen(ctx, ziel); antwort.Fehler == nil {
		t.Error("ohne Wegwerfort wurde gelesen")
	}
}

// **Ein Ordner, in den niemand hineinsehen darf, meldet sich.**
//
// **Und eine Datei, die niemand lesen darf, auch** — zwei Zweige, die nur ein
// echter Lauf betritt: `WalkDir` reicht den einen herein, `ReadFile` gibt den
// anderen zurück.
func TestEinUnlesbaresStueckMeldetSich(t *testing.T) {
	t.Parallel()

	for name, versperre := range map[string]func(*testing.T, string){
		"ein versperrter Ordner": func(t *testing.T, ordner string) {
			t.Helper()

			if fehler := os.Mkdir(
				filepath.Join(ordner, "zu"), 0o000); fehler != nil {
				t.Fatalf("anlegen: %v", fehler)
			}
		},
		"eine versperrte Datei": func(t *testing.T, ordner string) {
			t.Helper()

			if fehler := os.WriteFile(
				filepath.Join(ordner, "zu"), []byte("x"), 0o000); fehler != nil {
				t.Fatalf("schreiben: %v", fehler)
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ordner := t.TempDir()
			versperre(t, ordner)

			// **Aufgeräumt wird ausdrücklich** — `t.TempDir` räumt selbst
			// ab und käme an einem Ordner ohne Rechte nicht vorbei.
			t.Cleanup(func() {
				// **0700 und nicht 0600**: der eine Fall ist ein Ordner,
				// und ohne das Ausführungsrecht käme `t.TempDir` nicht
				// hinein, um ihn abzuräumen.
				//nolint:errcheck,gosec // Aufräumen eines Wegwerforts
				_ = os.Chmod(filepath.Join(ordner, "zu"), 0o700)
			})

			if _, fehler := setEntfalten(ordner); fehler == nil {
				t.Error("das unlesbare Stück kam durch")
			}
		})
	}
}

// **Ein Set, das dasteht, aber nicht zu lesen ist, wird nicht geschnürt.**
//
// # Warum über eine Vorrichtung und nicht über einen echten Weg
//
// **Weil dieser Zweig zwischen zwei Schritten liegt.** Der Setschreiber hat
// geschrieben, und erst das Entfalten scheitert — dazwischen kommt kein
// fachlicher Fall. Die Vorrichtung sagt genau das: „geschrieben — dorthin",
// und dorthin führt kein Weg.
//
// **Die Zusage, die dabei zählt**, ist die Reihenfolge: es wird **nichts** am
// Ziel abgelegt. Ein halber Umschlag neben einem gemeldeten Fehler wäre
// schlimmer als kein Umschlag.
func TestEinUnlesbaresSetWirdNichtGeschnuert(t *testing.T) {
	anwendung := mitSaat(t).MitPeerbund(&peerprobebund{})

	ctx := t.Context()

	if antwort := anwendung.PeeridentitaetSicherstellen(
		anwendung.imProfil(ctx)); antwort.Fehler != nil {
		t.Fatalf("Identität: %+v", antwort.Fehler)
	}

	anwendung.MitSpeicherer(schreiberInsLeere{})

	ziel := filepath.Join(t.TempDir(), "unlesbar.msepackage")

	if antwort := anwendung.UmschlagSchnueren(ctx, "", "1", ziel); antwort.Fehler == nil {
		t.Error("ein unlesbares Set wurde geschnürt")
	}

	if _, fehler := os.Stat(ziel); !errors.Is(fehler, os.ErrNotExist) {
		t.Error("am Ziel liegt etwas, obwohl nichts zu lesen war")
	}
}

// schreiberInsLeere meldet ein geschriebenes Set an einem Ort, den es nicht
// gibt.
type schreiberInsLeere struct{}

func (schreiberInsLeere) Schreibe(_ context.Context, _ int64, ziel string) (
	*service.Speicherung, error,
) {
	return &service.Speicherung{Pfad: filepath.Join(ziel, "gibtesnicht")}, nil
}
