// **Die Matrix hält, was sie zusagt — soweit ein Wächter das kann** —
// Werkstück O1 von Phase 5.2
// (#1425),
// nach [0078](../../docs/entscheidungen/0078-der-wirksame-akteur.md).
//
// # Was er hält
//
// Jede Fähigkeit aus `FAEHIGKEITEN.md` steht in `AUTORISIERUNG.md` genau
// einmal, jede Wirkung ist eine der fünf aus `domain`, jede Mindestrolle ist
// die, die `domain` für die Klasse nennt, und die Zahlen im Kopf sind
// gezählt.
//
// # Was er nicht hält
//
// Ob der Träger stimmt, und ob die Durchsetzung dort sitzt, wo die Zeile es
// sagt. Das erste ist eine Frage ans Review; das zweite belegen die Läufe
// am gestarteten Fenster und an der gebauten Kommandozeile.
package werkzeuge

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

const matrixpfad = "../../docs/AUTORISIERUNG.md"

// matrixzeile ist eine Zeile der Matrix.
type matrixzeile struct {
	faehigkeit, traeger, wirkung, rolle, durchsetzung string
}

// matrixzeilen liest die Tabelle.
func matrixzeilen(t *testing.T) []matrixzeile {
	t.Helper()

	roh, err := os.ReadFile(matrixpfad) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", matrixpfad, err)
	}

	var zeilen []matrixzeile

	for _, zeile := range strings.Split(string(roh), "\n") {
		if !strings.HasPrefix(zeile, "| ") || strings.Count(zeile, "|") != 6 ||
			strings.HasPrefix(zeile, "| Fähigkeit |") || strings.Contains(zeile, "---") {
			continue
		}

		teile := strings.Split(strings.TrimSuffix(strings.TrimPrefix(zeile, "| "), " |"), " | ")
		if len(teile) != 5 {
			t.Fatalf("die Zeile hat %d Spalten statt 5: %s", len(teile), zeile)
		}

		zeilen = append(zeilen, matrixzeile{
			faehigkeit: teile[0], traeger: teile[1],
			wirkung: strings.Trim(teile[2], "`"), rolle: teile[3], durchsetzung: teile[4],
		})
	}

	if len(zeilen) == 0 {
		t.Fatal("keine einzige Matrixzeile gefunden — der Leser greift daneben")
	}

	return zeilen
}

// **Jede Fähigkeit genau einmal — und keine, die es nicht gibt.**
func TestDieMatrixNenntJedeFaehigkeitGenauEinmal(t *testing.T) {
	t.Parallel()

	erwartet := map[string]int{}

	for _, zeile := range faehigkeitszeilen(t) {
		name := strings.TrimSpace(strings.Split(strings.TrimPrefix(zeile, "| "), " | ")[0])
		erwartet[name]++
	}

	gesehen := map[string]int{}

	for _, zeile := range matrixzeilen(t) {
		gesehen[zeile.faehigkeit]++

		if erwartet[zeile.faehigkeit] == 0 {
			t.Errorf("die Matrix nennt %q, FAEHIGKEITEN.md nicht", zeile.faehigkeit)
		}
	}

	for name := range erwartet {
		if gesehen[name] != 1 {
			t.Errorf("%q steht %d-mal in der Matrix, erwartet einmal", name, gesehen[name])
		}
	}
}

// mindestrolle ist, was `domain` für eine Klasse an einem Profil verlangt —
// **aus dem Code gelesen**, nicht hier noch einmal aufgeschrieben.
func mindestrolle(wirkung domain.Wirkung) string {
	for _, rolle := range []domain.Rolle{domain.RolleLeser, domain.RolleMitarbeiter, domain.RolleVerwalter} {
		akteur := domain.Akteur{Rolle: rolle, NutzerID: 1, ProfilID: 1, EignerID: 1}
		if akteur.Darf(wirkung) == nil {
			return string(rolle)
		}
	}

	return "verwalter + Nachweis"
}

// **Jede Wirkung ist eine der fünf, jede Mindestrolle die aus `domain` — und
// am Profil steht nie ein Strich.**
func TestJedeZeileTraegtEineWirkungUndDieRolleDazu(t *testing.T) {
	t.Parallel()

	bekannt := map[string]bool{}
	for _, wirkung := range domain.Wirkungen() {
		bekannt[string(wirkung)] = true
	}

	traeger := map[string]bool{"Profil": true, "Nutzer": true, "Bestand": true, "Rechner": true}

	for _, zeile := range matrixzeilen(t) {
		if !bekannt[zeile.wirkung] {
			t.Errorf("%q: die Wirkung %q gibt es nicht", zeile.faehigkeit, zeile.wirkung)

			continue
		}

		if !traeger[zeile.traeger] {
			t.Errorf("%q: der Träger %q ist keiner der vier", zeile.faehigkeit, zeile.traeger)
		}

		if strings.TrimSpace(zeile.durchsetzung) == "" {
			t.Errorf("%q: die Durchsetzung fehlt", zeile.faehigkeit)
		}

		erwartet := "—"
		if zeile.traeger == "Profil" {
			erwartet = mindestrolle(domain.Wirkung(zeile.wirkung))
		}

		if zeile.rolle != erwartet {
			t.Errorf("%q: Mindestrolle %q, aus domain folgt %q", zeile.faehigkeit, zeile.rolle, erwartet)
		}
	}
}

// **Und die Zahlen im Kopf sind gezählt.**
func TestDieZahlenDerMatrixSindGezaehlt(t *testing.T) {
	t.Parallel()

	roh, err := os.ReadFile(matrixpfad) //nolint:gosec // ein fester Pfad im Repositorium
	if err != nil {
		t.Fatalf("%s lesen: %v", matrixpfad, err)
	}

	zeilen := matrixzeilen(t)
	wirkungen := map[string]int{}
	traeger := map[string]int{}

	for _, zeile := range zeilen {
		wirkungen[zeile.wirkung]++
		traeger[zeile.traeger]++
	}

	erwartet := map[string]int{"| Fähigkeiten | **%d**": len(zeilen)}
	for _, wirkung := range domain.Wirkungen() {
		erwartet["| `"+string(wirkung)+"` | **%d**"] = wirkungen[string(wirkung)]
	}

	for name, zahl := range traeger {
		erwartet["| am Träger **"+name+"** | **%d**"] = zahl
	}

	for muster, zahl := range erwartet {
		gesucht := fmt.Sprintf(muster, zahl)
		if !strings.Contains(string(roh), gesucht) {
			t.Errorf("%s nennt nicht %q", matrixpfad, gesucht)
		}
	}

	// **Und keine andere Zahl an dieser Stelle** — sonst stünde die alte
	// neben der neuen.
	for muster := range erwartet {
		stamm := strings.TrimSuffix(muster, " | **%d**")
		treffer := regexp.MustCompile(regexp.QuoteMeta(stamm)+` \| \*\*(\d+)\*\*`).FindAllString(string(roh), -1)

		if len(treffer) != 1 {
			t.Errorf("%q steht %d-mal im Kopf, erwartet einmal", stamm, len(treffer))
		}
	}
}
