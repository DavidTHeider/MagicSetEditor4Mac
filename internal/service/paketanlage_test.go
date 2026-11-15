// **Ein Spiel- und ein Stilpaket lassen sich anlegen** —
// #1350.
//
// # Was fehlte
//
// Ein `.mse-set` ließ sich schreiben, eine Datei **in** einem installierten
// Paket ändern — ein `.mse-game` oder `.mse-style` **anlegen** nicht. Wer sich
// eine eigene Vorlage bauen wollte, legte die Dateien von Hand an und band das
// Paket danach ein.
//
// # Warum beide und nicht nur das Spiel
//
// Die Abnahme des Issues verlangt, dass sich auf dem neuen Spiel **ein Set
// anlegen** lässt — und ein Set trägt Spiel **und** Stil. Ein Spielpaket
// allein wäre eine Fähigkeit, die man nicht benutzen kann; das ist die erste
// Dimension aus §10.
package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein angelegtes Spiel steht im Katalog und auf der Platte.**
func TestEinAngelegtesSpielStehtImKatalog(t *testing.T) {
	ort, verwalter, katalog := anlageplatz(t)

	angelegt, err := verwalter.Anlegen(t.Context(),
		service.Neuanlage{Art: "mse-game", Name: "eigenes", Vollname: "Eigenes Spiel"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if angelegt.Art != "mse-game" || angelegt.Name != "eigenes" {
		t.Fatalf("angelegt: %+v", angelegt)
	}

	if len(katalog.pakete) != 1 {
		t.Errorf("%d Pakete im Katalog", len(katalog.pakete))
	}

	// **Ein angelegtes Paket ist ein eigenes** (DOM-01.AK-01, #1696): ohne
	// Herkunft — der Bauort ist keine — und ohne Abdruck, denn bei einem
	// eigenen Paket sind Änderungen der Normalfall.
	if angelegt.Herkunft != "" || katalog.pakete[0].Herkunft != "" || katalog.pakete[0].Abdruck != "" ||
		katalog.pakete[0].HerkunftsartIn(ort) != domain.PaketEigen {
		t.Errorf("das angelegte Paket: %+v, im Katalog %+v", angelegt, katalog.pakete[0])
	}

	// **Die Datei heißt `game`**, wie MSE2 sie sucht.
	inhalt := datei(t, filepath.Join(ort, "eigenes.mse-game", "game"))

	for _, zeile := range []string{
		"mse version: 2.0.0",
		"short name: eigenes",
		"full name: Eigenes Spiel",
		"card field:",
	} {
		if !strings.Contains(inhalt, zeile) {
			t.Errorf("die Datei führt %q nicht:\n%s", zeile, inhalt)
		}
	}
}

// **Ein angelegter Stil nennt sein Spiel und hängt daran.**
//
// Ohne beides öffnete MSE2 ihn nicht: `game:` sagt, wozu er gehört,
// `depends on:` hält die Fassung fest, gegen die er gebaut wurde.
func TestEinAngelegterStilNenntSeinSpiel(t *testing.T) {
	ort, verwalter, _ := anlageplatz(t)
	anlage := verwalter

	if _, err := anlage.Anlegen(t.Context(), service.Neuanlage{
		Art: "mse-game", Name: "eigenes", Vollname: "Eigenes Spiel",
	}); err != nil {
		t.Fatalf("das Spiel anlegen: %v", err)
	}

	angelegt, err := anlage.Anlegen(t.Context(), service.Neuanlage{
		Art: "mse-style", Name: "schlicht", Vollname: "Schlichter Stil", Spiel: "eigenes",
	})
	if err != nil {
		t.Fatalf("den Stil anlegen: %v", err)
	}

	if angelegt.Art != "mse-style" {
		t.Fatalf("angelegt: %+v", angelegt)
	}

	inhalt := datei(t, filepath.Join(ort, "eigenes-schlicht.mse-style", "style"))

	for _, zeile := range []string{
		"game: eigenes",
		"depends on:",
		"package: eigenes.mse-game",
		"card style:",
	} {
		if !strings.Contains(inhalt, zeile) {
			t.Errorf("die Datei führt %q nicht:\n%s", zeile, inhalt)
		}
	}
}

// **Der Stil legt jedes Feld des Spiels aus.**
//
// Sonst zeichnete die Karte nichts: ein `.mse-style` ohne `card style` für ein
// Feld lässt es weg. Ein Paar aus Spiel und Stil, das eine leere Karte malt,
// ist keine Vorlage — dieselbe Prüfung wie an
// [0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md), nur
// eine Ebene tiefer.
func TestDerStilLegtJedesFeldDesSpielsAus(t *testing.T) {
	ort, verwalter, _ := anlageplatz(t)
	anlage := verwalter

	if _, err := anlage.Anlegen(t.Context(), service.Neuanlage{
		Art: "mse-game", Name: "eigenes", Vollname: "Eigenes Spiel",
	}); err != nil {
		t.Fatalf("das Spiel anlegen: %v", err)
	}

	if _, err := anlage.Anlegen(t.Context(), service.Neuanlage{
		Art: "mse-style", Name: "schlicht", Vollname: "Schlicht", Spiel: "eigenes",
	}); err != nil {
		t.Fatalf("den Stil anlegen: %v", err)
	}

	spiel := datei(t, filepath.Join(ort, "eigenes.mse-game", "game"))
	stil := datei(t, filepath.Join(ort, "eigenes-schlicht.mse-style", "style"))

	for _, feld := range feldnamenAus(spiel) {
		if !strings.Contains(stil, "\t"+feld+":") {
			t.Errorf("der Stil legt das Feld %q nicht aus:\n%s", feld, stil)
		}
	}
}

// **Ein Stil ohne Spiel wird abgewiesen, und zwar am Feld.**
func TestEinStilOhneSpielWirdAbgewiesen(t *testing.T) {
	_, verwalter, _ := anlageplatz(t)

	_, err := verwalter.Anlegen(t.Context(),
		service.Neuanlage{Art: "mse-style", Name: "schlicht", Vollname: "Schlicht"})

	var feldfehler *domain.FieldError
	if !errors.As(err, &feldfehler) || feldfehler.Field != "spiel" {
		t.Fatalf("Fehler ist %v", err)
	}
}

// **Und ein Name, der kein Verzeichnisname ist, ebenso.**
//
// Ein Schrägstrich legte das Paket woanders an, ein `..` eine Ebene höher —
// beides außerhalb des verwalteten Verzeichnisses, das §11 verlangt.
func TestEinKrummerNameWirdAbgewiesen(t *testing.T) {
	_, verwalter, _ := anlageplatz(t)
	anlage := verwalter

	// **Die Liste ist geschlossen, seit die Meldung sie versprach** — Befund 2
	// des Zweitblicks über Phase 4.10. `stil!` und `späß` gingen durch,
	// während der Satz daneben „Buchstaben, Ziffern und Bindestriche" nannte.
	for _, name := range []string{
		"", "  ", "mit/schrägstrich", "..", "mit.punkt",
		"stil!", "späß", "mit leerzeichen", "back\\slash",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := anlage.Anlegen(t.Context(),
				service.Neuanlage{Art: "mse-game", Name: name, Vollname: "X"})

			var feldfehler *domain.FieldError
			if !errors.As(err, &feldfehler) || feldfehler.Field != "name" {
				t.Errorf("Fehler ist %v", err)
			}
		})
	}
}

// anlageplatz gibt Verwaltungsort, Verwalter und Katalog.
func anlageplatz(t *testing.T) (ort string, verwalter *service.Paketverwalter, katalog *ablage) {
	t.Helper()

	ort = t.TempDir()
	katalog = &ablage{}

	return ort, service.NeuerPaketverwalter(katalog, ort, ort), katalog
}

// datei liest eine geschriebene Datei.
func datei(t *testing.T, pfad string) string {
	t.Helper()

	inhalt, err := os.ReadFile(pfad) //nolint:gosec // Pfad aus t.TempDir()
	if err != nil {
		t.Fatalf("%s lesen: %v", pfad, err)
	}

	return string(inhalt)
}

// feldnamenAus liest die `name:`-Zeilen unter `card field:`.
func feldnamenAus(spiel string) []string {
	var (
		namen []string
		drin  bool
	)

	for _, zeile := range strings.Split(spiel, "\n") {
		if strings.HasPrefix(zeile, "card field:") {
			drin = true

			continue
		}

		if drin && strings.HasPrefix(zeile, "\tname: ") {
			namen = append(namen, strings.TrimPrefix(zeile, "\tname: "))
			drin = false
		}
	}

	return namen
}

// **Ein Stilname mit Spiel davor wird nicht doppelt vorangestellt** (#1350).
//
// Der Name, den man eingibt, ist nicht der, den man danach braucht: wer
// `schlicht` anlegt, bekommt `eigenes-schlicht` — und **genau das** verlangt
// `set-neu`. Der Flächenlauf ist darüber gestolpert. Beide Schreibweisen
// anzunehmen ist eindeutig; zweimal voranzustellen wäre
// `eigenes-eigenes-schlicht`.
func TestEinStilnameMitSpielDavorWirdNichtVerdoppelt(t *testing.T) {
	ort, verwalter, _ := anlageplatz(t)

	if _, err := verwalter.Anlegen(t.Context(), service.Neuanlage{
		Art: "mse-game", Name: "eigenes", Vollname: "Eigenes",
	}); err != nil {
		t.Fatalf("das Spiel anlegen: %v", err)
	}

	for _, name := range []string{"schlicht", "eigenes-bunt"} {
		t.Run(name, func(t *testing.T) {
			angelegt, err := verwalter.Anlegen(t.Context(), service.Neuanlage{
				Art: "mse-style", Name: name, Vollname: "Stil", Spiel: "eigenes",
			})
			if err != nil {
				t.Fatalf("anlegen: %v", err)
			}

			erwartet := "eigenes-" + strings.TrimPrefix(name, "eigenes-")
			if angelegt.Name != erwartet {
				t.Errorf("das Paket heißt %q, erwartet war %q", angelegt.Name, erwartet)
			}

			if _, fehler := os.Stat(
				filepath.Join(ort, erwartet+".mse-style", "style")); fehler != nil {
				t.Errorf("nicht geschrieben: %v", fehler)
			}
		})
	}
}

// **Und ein Name aus dem, was echte Pakete benutzen, geht durch.**
//
// Die Gegenprobe zur Liste darüber: ohne sie wäre eine Prüfung, die **alles**
// abweist, ebenso grün — und `magic-counter-sheet_744` ließe sich nicht
// anlegen, obwohl es im Pack liegt. **Gezählt über alle 526 Pakete**: außer
// Buchstaben und Ziffern kommen genau `-` und `_` vor.
func TestEinNameWieImPackGehtDurch(t *testing.T) {
	_, verwalter, _ := anlageplatz(t)

	for _, name := range []string{"magic", "magic-m15", "magic_counter_sheet_744", "vs2"} {
		t.Run(name, func(t *testing.T) {
			if _, err := verwalter.Anlegen(t.Context(), service.Neuanlage{
				Art: "mse-game", Name: name, Vollname: "Probe",
			}); err != nil {
				t.Errorf("%q wurde abgewiesen: %v", name, err)
			}
		})
	}
}

// **Ein Spiel, das schon da ist, darf heißen wie es will** — Befund 2 des
// Nachlaufs über Phase 4.10.
//
// # Der Befund
//
// `Installieren` und `Uebernehmen` nehmen einen Paketnamen so, wie er auf der
// Platte steht — geprüft wird nur die Endung. Ein `späß.mse-game` kann also im
// Katalog stehen und in der Spielwahl erscheinen. Die geschlossene Namensliste
// wies einen Stil darauf trotzdem ab: **eingebunden ja, benutzbar nein.**
//
// # Warum die Trennung stimmt
//
// `Name` wird zu einem Verzeichnis, das **wir** anlegen — dort gilt die
// geschlossene Liste. `Spiel` benennt eines, das **schon da ist** — dort gilt
// nur die Wegsicherheit.
func TestEinVorhandenesSpielDarfHeissenWieEsWill(t *testing.T) {
	ort := t.TempDir()
	spielAnlegenVonHand(t, ort, "späß")

	verwalter := service.NeuerPaketverwalter(&ablage{}, ort, ort)

	angelegt, err := verwalter.Anlegen(t.Context(), service.Neuanlage{
		Art: "mse-style", Name: "schlicht", Vollname: "Schlicht", Spiel: "späß",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if angelegt.Name != "späß-schlicht" {
		t.Errorf("das Paket heißt %q", angelegt.Name)
	}
}

// spielAnlegenVonHand legt ein Spielpaket unter diesem Namen ab — so, wie es
// `Installieren` täte.
func spielAnlegenVonHand(t *testing.T, ort, name string) {
	t.Helper()

	verzeichnis := filepath.Join(ort, name+".mse-game")
	if err := os.MkdirAll(verzeichnis, 0o750); err != nil {
		t.Fatalf("das Spiel anlegen: %v", err)
	}

	inhalt := "mse version: 2.0.0\nshort name: " + name +
		"\ncard field:\n\ttype: text\n\tname: name\n"

	if err := os.WriteFile(
		filepath.Join(verzeichnis, "game"), []byte(inhalt), 0o600); err != nil {
		t.Fatalf("das Spiel schreiben: %v", err)
	}
}

// **Ein wegunsicherer Spielname wird abgewiesen** (#1350).
//
// Angestrichen von `gosec` (G304) — **zu Recht**: die erste Fassung prüfte nur
// `Name`, und der Spielname wird genauso an eine Paketwurzel gehängt. Ein `..`
// darin läse eine Datei außerhalb jeder Wurzel.
func TestEinWegunsichererSpielnameWirdAbgewiesen(t *testing.T) {
	_, verwalter, _ := anlageplatz(t)

	for _, spiel := range []string{"..", "../../etc", "mit/schrägstrich", "mit.punkt"} {
		t.Run(spiel, func(t *testing.T) {
			_, err := verwalter.Anlegen(t.Context(), service.Neuanlage{
				Art: "mse-style", Name: "schlicht", Vollname: "S", Spiel: spiel,
			})

			var feldfehler *domain.FieldError
			if !errors.As(err, &feldfehler) || feldfehler.Field != "spiel" {
				t.Errorf("Fehler ist %v", err)
			}
		})
	}
}
