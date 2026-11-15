package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eigene Merkmale sind Farben und sonst nichts** — UX-09.AK-04
// (#1681): deklarativ, eingeschränkt, keine Skriptausführung, keine
// unbemerkten Netzbezüge, keine Schriftdateien.
func TestEigeneMerkmaleNehmenNurBekannteFarben(t *testing.T) {
	gut := []struct {
		name string
		wert string
	}{
		{"leer", ""},
		{"ein Merkmal, lang", "--mse-grund: #112233;"},
		{"ein Merkmal, kurz", "--mse-text: #fff;"},
		{"große Buchstaben", "--mse-marke: #A0B1C2;"},
		{"mehrere Zeilen mit Leerzeilen", "--mse-grund: #112233;\n\n  --mse-text:#eeeeee;\n"},
		{"ohne Semikolon am Ende", "--mse-kante: #777777"},
	}
	for _, fall := range gut {
		t.Run(fall.name, func(t *testing.T) {
			if fehler := domain.PruefeGestaltung(fall.wert); fehler != nil {
				t.Fatalf("abgewiesen: %v", fehler)
			}
		})
	}

	schlecht := []struct {
		name  string
		wert  string
		zeile string
	}{
		{"ein fremdes Merkmal", "--mse-tisch-grund: #112233;", "Zeile 1"},
		{"eine Eigenschaft statt eines Merkmals", "color: #112233;", "Zeile 1"},
		{"eine Adresse", "--mse-grund: url(https://example.org/a.png);", "Zeile 1"},
		{"eine lokale Adresse", "--mse-grund: url(file:///etc/passwd);", "Zeile 1"},
		{"ein Import", "@import 'https://example.org/a.css';", "Zeile 1"},
		{"eine Schriftdatei", "@font-face { src: url(a.woff2) }", "Zeile 1"},
		{"ein Ausbruch aus der Angabe", "--mse-grund: #111; } body { display: none", "Zeile 1"},
		{"ein Farbname", "--mse-text: red;", "Zeile 1"},
		{"eine Farbfunktion", "--mse-text: rgb(1, 2, 3);", "Zeile 1"},
		{"ein Verweis auf ein anderes Merkmal", "--mse-text: var(--mse-grund);", "Zeile 1"},
		{"eine Maskierung", "--mse-grund: \\75rl(a);", "Zeile 1"},
		{"eine Hexzahl mit falscher Länge", "--mse-grund: #12345;", "Zeile 1"},
		{"ein Merkmal zweimal", "--mse-grund: #111111;\n--mse-grund: #222222;", "Zeile 2"},
		{"der Fehler in der dritten Zeile", "--mse-grund: #111111;\n\n--mse-text: blau;", "Zeile 3"},
	}
	for _, fall := range schlecht {
		t.Run(fall.name, func(t *testing.T) {
			fehler := domain.PruefeGestaltung(fall.wert)
			if !errors.Is(fehler, domain.ErrValidation) {
				t.Fatalf("durchgelassen oder falsch abgewiesen: %v", fehler)
			}
			if !strings.Contains(fehler.Error(), fall.zeile) {
				t.Fatalf("nennt die Stelle nicht (%s): %v", fall.zeile, fehler)
			}
		})
	}
}

// **Eine Grenze, bevor gelesen wird** — mehr als zehn Merkmale gibt es nicht,
// und ein Feld von Megabytes wäre keine Gestaltung.
func TestEigeneMerkmaleHabenEineGrenze(t *testing.T) {
	lang := strings.Repeat("--mse-grund: #111111;\n", 200)
	fehler := domain.PruefeGestaltung(lang)
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("durchgelassen: %v", fehler)
	}
	if !strings.Contains(fehler.Error(), "Zeichen") {
		t.Fatalf("nennt die Grenze nicht: %v", fehler)
	}
}

// **Die Liste der Merkmale ist die Zusage** — zehn, und die Tischfarben
// gehören nicht dazu.
func TestDieGestaltungsmerkmaleSindDieZehnFarben(t *testing.T) {
	merkmale := domain.Gestaltungsmerkmale()
	if len(merkmale) != 10 {
		t.Fatalf("%d Merkmale: %v", len(merkmale), merkmale)
	}
	for _, eines := range merkmale {
		if !strings.HasPrefix(eines, "--mse-") || strings.Contains(eines, "tisch") {
			t.Fatalf("unerwartetes Merkmal %q", eines)
		}
	}
}

// **Die Einstellung prüft mit** — über denselben Weg wie jede andere, damit
// MCP und Kommandozeile dieselbe Grenze haben wie das Fenster.
func TestDieEinstellungGestaltungPrueftIhrenWert(t *testing.T) {
	if fehler := domain.PruefeEinstellung(domain.EinstellungGestaltung,
		"--mse-grund: url(x);"); !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("durchgelassen: %v", fehler)
	}
	if fehler := domain.PruefeEinstellung(domain.EinstellungGestaltung,
		"--mse-grund: #101010;"); fehler != nil {
		t.Fatalf("abgewiesen: %v", fehler)
	}
}

// **Ein wohlgeformtes, aber fremdes Merkmal wird beim Namen genannt** — die
// Form stimmt, nur ändern lässt es sich nicht. `--mse-schleier` gibt es,
// aber es ist keine Farbe (`themen.css`).
func TestEinFremdesMerkmalWirdBeimNamenGenannt(t *testing.T) {
	fehler := domain.PruefeGestaltung("--mse-grund: #111111;\n--mse-schleier: #112233;")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("durchgelassen: %v", fehler)
	}

	if !strings.Contains(fehler.Error(), "Zeile 2: »--mse-schleier« ist kein Merkmal") {
		t.Fatalf("nennt das Merkmal nicht: %v", fehler)
	}
}
