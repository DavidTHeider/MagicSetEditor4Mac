package app

// **Die vier Verwaltungswege einer Partie am Fenster** — PLAY-03.AK-05
// (#1711).

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// verwaltungsanwendung baut eine Anwendung mit einer schweigsamen Leitung
// und einer Datei, die im Arbeitsspeicher liegt.
func verwaltungsanwendung(t *testing.T) (*App, *papierkorb) {
	t.Helper()

	korb := &papierkorb{}

	anwendung := NeueApp("")
	anwendung.db = bestandFuerDenRiegel(t)
	anwendung.profil.Store(1)
	anwendung.MitPartiedatei(korb)

	return anwendung.MitPartienaht(&schweigsameLeitung{}), korb
}

// papierkorb ist eine Datei, die es nur im Arbeitsspeicher gibt.
type papierkorb struct {
	inhalt map[string][]byte
	klagt  error
}

func (p *papierkorb) Schreibe(pfad string, inhalt []byte) error {
	if p.klagt != nil {
		return p.klagt
	}

	if p.inhalt == nil {
		p.inhalt = map[string][]byte{}
	}

	p.inhalt[pfad] = inhalt

	return nil
}

func (p *papierkorb) Lies(pfad string) ([]byte, error) {
	if p.klagt != nil {
		return nil, p.klagt
	}

	return p.inhalt[pfad], nil
}

// **Eine Partie verschwindet, und es bleibt keine halbe Antwort zurück.**
func TestEinePartieLaesstSichAmFensterLoeschen(t *testing.T) {
	t.Parallel()

	anwendung, _ := verwaltungsanwendung(t)

	antwort := anwendung.PartieLoeschen(t.Context(), "p1")

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Kennung != "":
		t.Errorf("gelöscht, und es kam eine Kennung zurück: %q", antwort.Kennung)
	case antwort.Befunde == nil:
		t.Error("die Befunde sind nil statt leer")
	}
}

// **Eine Kopie bekommt eine neue Kennung** — und die kommt von hier oben,
// wie beim Anlegen.
func TestEineKopieBekommtEineNeueKennung(t *testing.T) {
	t.Parallel()

	anwendung, _ := verwaltungsanwendung(t)

	antwort := anwendung.PartieKopieren(t.Context(), "p1")

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Kennung == "" || antwort.Kennung == "p1":
		t.Errorf("die Kopie heißt %q", antwort.Kennung)
	}
}

// **Was ausgeführt wird, liegt danach da** — und der Ort steht in der
// Antwort, damit ein Mensch ihn wiederfindet.
func TestDieAusfuhrLegtDieDateiAb(t *testing.T) {
	t.Parallel()

	anwendung, korb := verwaltungsanwendung(t)
	pfad := filepath.Join(t.TempDir(), "partie.msepartie")

	antwort := anwendung.PartieAusfuhrSchreiben(t.Context(), "p1", pfad)

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Pfad != pfad:
		t.Errorf("abgelegt wurde nach %q", antwort.Pfad)
	case string(korb.inhalt[pfad]) != "{}":
		t.Errorf("in der Datei steht %q", korb.inhalt[pfad])
	}
}

// **Und was eingelesen wird, wird zu einer neuen Partie — samt Befunden.**
func TestDieEinfuhrLegtEineNeuePartieAn(t *testing.T) {
	t.Parallel()

	anwendung, korb := verwaltungsanwendung(t)
	pfad := filepath.Join(t.TempDir(), "partie.msepartie")
	korb.inhalt = map[string][]byte{pfad: []byte("{}")}

	antwort := anwendung.PartieEinfuhrLesen(t.Context(), pfad)

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("Fehler: %v", antwort.Fehler)
	case antwort.Kennung == "":
		t.Error("die eingelesene Partie hat keine Kennung")
	case len(antwort.Befunde) != 1:
		t.Errorf("die Befunde kamen nicht mit: %v", antwort.Befunde)
	}
}

// **Jeder der vier Wege reicht den Fehler seiner Leitung durch** — und
// hinterlässt dabei keine halbe Antwort (§8).
func TestJederVerwaltungswegReichtSeinenFehlerDurch(t *testing.T) {
	t.Parallel()

	faelle := map[string]func(*App) PartieverwaltungAntwort{
		"Loeschen": func(a *App) PartieverwaltungAntwort {
			return a.PartieLoeschen(t.Context(), "p1")
		},
		"Kopieren": func(a *App) PartieverwaltungAntwort {
			return a.PartieKopieren(t.Context(), "p1")
		},
		"Ausfuhr": func(a *App) PartieverwaltungAntwort {
			return a.PartieAusfuhrSchreiben(t.Context(), "p1", "irgendwo")
		},
		"Einfuhr": func(a *App) PartieverwaltungAntwort {
			return a.PartieEinfuhrLesen(t.Context(), "irgendwo")
		},
	}

	for weg, fahre := range faelle {
		t.Run(weg, func(t *testing.T) {
			t.Parallel()

			anwendung, _ := verwaltungsanwendung(t)
			anwendung.MitPartienaht(&klagendePartieleitung{
				weg: weg, klagt: errors.New("die Leitung schweigt"),
			})

			antwort := fahre(anwendung)

			switch {
			case antwort.Fehler == nil:
				t.Fatalf("der Weg ging durch: %+v", antwort)
			case antwort.Kennung != "" || antwort.Pfad != "":
				t.Errorf("es blieb eine halbe Antwort: %+v", antwort)
			}
		})
	}
}

// **Und die Datei ebenso** — sie ist der zweite Weg, auf dem es schiefgehen
// kann, und er ist ein anderer als der der Leitung (§10).
func TestDieDateifehlerKommenAmFensterAn(t *testing.T) {
	t.Parallel()

	for _, weg := range []string{"Ausfuhr", "Einfuhr"} {
		t.Run(weg, func(t *testing.T) {
			t.Parallel()

			anwendung, korb := verwaltungsanwendung(t)
			korb.klagt = errors.New("die Platte ist voll")

			antwort := anwendung.PartieAusfuhrSchreiben(
				t.Context(), "p1", "irgendwo")

			if weg == "Einfuhr" {
				antwort = anwendung.PartieEinfuhrLesen(t.Context(), "irgendwo")
			}

			if antwort.Fehler == nil || antwort.Pfad != "" {
				t.Errorf("es kam %+v", antwort)
			}
		})
	}
}

// **Die beiden Textwege tragen ihre Fehler genauso** — sie stehen neben den
// Dateiwegen und nicht in ihnen, und ein Weg, den niemand prüft, ist ein Weg
// ohne Zusage (§10).
func TestDieTextwegeReichenIhreFehlerDurch(t *testing.T) {
	t.Parallel()

	anwendung, _ := verwaltungsanwendung(t)
	anwendung.MitPartienaht(&klagendePartieleitung{
		weg: "Ausfuhr", klagt: errors.New("die Leitung schweigt"),
	})

	if antwort := anwendung.PartieAusfuhrText(
		t.Context(), "p1"); antwort.Fehler == nil || antwort.Inhalt != "" {
		t.Errorf("die Ausfuhr gab %+v", antwort)
	}

	// **Und ohne die Rolle geschieht auch hier nichts.**
	ohneRolle, _ := verwaltungsanwendung(t)
	ohneRolle.MitWache(&abweisendeWache{})

	if antwort := ohneRolle.PartieEinfuhrText(
		t.Context(), "{}"); antwort.Fehler == nil || antwort.Kennung != "" {
		t.Errorf("die Einfuhr gab %+v", antwort)
	}
}

// **Ohne die Rolle geschieht nichts** — und gefragt wird nach der Wirkung,
// die der Weg wirklich hat: drei schreiben, die Ausfuhr liest (§10).
func TestOhneRolleGehtKeinVerwaltungsweg(t *testing.T) {
	t.Parallel()

	faelle := map[string]struct {
		fahre   func(*App) PartieverwaltungAntwort
		wirkung domain.Wirkung
	}{
		"Loeschen": {
			func(a *App) PartieverwaltungAntwort {
				return a.PartieLoeschen(t.Context(), "p1")
			}, domain.WirkungSchreiben,
		},
		"Kopieren": {
			func(a *App) PartieverwaltungAntwort {
				return a.PartieKopieren(t.Context(), "p1")
			}, domain.WirkungSchreiben,
		},
		"Ausfuhr": {
			func(a *App) PartieverwaltungAntwort {
				return a.PartieAusfuhrSchreiben(t.Context(), "p1", "irgendwo")
			}, domain.WirkungLesen,
		},
		"Einfuhr": {
			func(a *App) PartieverwaltungAntwort {
				return a.PartieEinfuhrLesen(t.Context(), "irgendwo")
			}, domain.WirkungSchreiben,
		},
	}

	for name, fall := range faelle {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			anwendung, korb := verwaltungsanwendung(t)
			wache := &abweisendeWache{}
			anwendung.MitWache(wache)

			if antwort := fall.fahre(anwendung); antwort.Fehler == nil {
				t.Fatalf("die Abweisung kam nicht an: %+v", antwort)
			}

			if wache.gefragt != fall.wirkung {
				t.Errorf("gefragt wurde nach %q statt nach %q",
					wache.gefragt, fall.wirkung)
			}

			// **Und die Datei ist nicht angefasst worden** — eine
			// Abweisung, die trotzdem schreibt, ist keine.
			if len(korb.inhalt) != 0 {
				t.Errorf("geschrieben wurde trotzdem: %v", korb.inhalt)
			}
		})
	}
}
