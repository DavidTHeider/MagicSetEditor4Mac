//go:build fremdpakete || netzlauf

// **Die Wege zu den kanonischen Sätzen und zum Zeichenplan.**
//
// Eigene Datei mit **beiden** Marken, seit Werkstück Ω.6
// (#1029):
// der Lauf gegen die gedruckten Karten braucht dieselben Wege und steht hinter
// `netzlauf`.
//
// **Nicht `netzlauf && fremdpakete` an jenem Lauf.** Die Stufe `lint` fährt
// golangci-lint je Marke **einzeln**; eine Datei, die beide verlangt, würde
// unter keiner von ihnen gelesen — ein Wächter, der wegsieht, weil zwei
// Bedingungen nicht zugleich gelten.
package app

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// fremdpaketwurzelOderSkip gibt das Paketverzeichnis oder überspringt.
func fremdpaketwurzelOderSkip(t *testing.T) string {
	t.Helper()

	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	return wurzel
}

// kanonsatzpfad gibt den Pfad eines kanonischen Sets.
//
// **`MSE_KANON` geht vor `HOME`.** Der Lauf gegen die gedruckten Karten lenkt
// `HOME` in ein leeres Verzeichnis um (§3), damit er den Zwischenspeicher
// dieses Rechners nicht sieht — und fände die Sätze dann nicht mehr, denn
// [kanonort] baut seinen Weg aus `HOME`. Wer beides will, nennt den Ort
// ausdrücklich; so tut es auch `kanonsatz` im Messlauf.
func kanonsatzpfad(t *testing.T, satz string) string {
	t.Helper()

	ort := os.Getenv("MSE_KANON")
	if ort == "" {
		ort = kanonort()
	}

	pfad := filepath.Join(ort, satz+".mse-set")
	if _, err := os.Stat(pfad); err != nil { // #nosec G703 -- ein Pfad aus dem Kanon
		t.Skipf("%s: %v", pfad, err)
	}

	return pfad
}

// planZurKarte holt den Zeichenplan einer Karte — denselben, aus dem die
// Vorschau entstanden ist.
//
// **Über dieselben Nähte wie [App.Vorschau]**, nicht über einen zweiten Weg:
// ein Plan aus einer eigenen Kette wäre nicht der, der gezeichnet wurde, und
// die Flächen lägen woanders. **Auch über `imProfil`** — hier fehlte es, und
// seit es Profile gibt, las der Planbauer den Setkopf nicht und fiel still auf
// den Behelfsplan zurück: jeder Lauf darüber maß eine Feldliste statt der
// Karte, zwei Kanonläufe wurden rot, und die Profil- und Messläufe maßen die
// Zeit für die Feldliste (QA-03, #1656).
//
// **Der Behelfsplan ist ein Fehlschlag — ein Befund im Stilplan nicht.** Der
// Behelfsplan kommt samt Befund zurück, und wer ihn wegwirft, misst ihn; also
// bricht der Lauf ab, wenn er ihn bekommt. Hier stand »ein Befund ist ein
// Fehlschlag«, und damit brach die Zufallsstichprobe an Karte 14944 ab, deren
// Manakosten ein `\n` tragen: ein echter Stilplan mit drei Hinweisen der
// Symbolschrift, kein Rückfall. **Erkannt wird der Rückfall am Plan selbst** —
// er ist für eine Karte immer derselbe, und derselbe Planbauer baut ihn zum
// Vergleich.
func planZurKarte(tb testing.TB, anwendung *App, karteID string) *render.Plan {
	tb.Helper()

	ctx := anwendung.imProfil(context.Background())

	kennung, befund := kennungAus(karteID)
	if befund != nil {
		tb.Fatalf("Karte %s: %+v", karteID, befund)
	}

	karte, err := anwendung.kartenleser.KarteLesen(ctx, kennung)
	if err != nil {
		tb.Fatalf("Karte %s lesen: %v", karteID, err)
	}

	plan, befunde, err := anwendung.planbauer.Kartenplan(ctx, karte)
	if err != nil {
		tb.Fatalf("Karte %s hat keinen Plan: %v", karteID, err)
	}

	behelfer, kann := anwendung.planbauer.(interface {
		Behelfsplan(ctx context.Context, karte *domain.CardRow) (*render.Plan, error)
	})
	if !kann {
		tb.Fatalf("der Planbauer baut keinen Behelfsplan — dann ist der Rückfall nicht zu erkennen")
	}

	behelf, err := behelfer.Behelfsplan(ctx, karte)
	if err != nil || reflect.DeepEqual(plan, behelf) {
		tb.Fatalf("Karte %s bekam den Behelfsplan statt ihres Stils: %v %v", karteID, err, befunde)
	}

	return plan
}

// kanonort gibt das Verzeichnis der kanonischen Sätze unter `HOME`.
//
// **Neben dem Repositorium und nicht darin** (§11): sechs der zehn Sätze
// liegen je über 227 MB.
func kanonort() string {
	// **`MSE_KANON` geht vor `HOME`** — derselbe Grund wie in
	// [kanonverzeichnis], und dieselbe Klasse: unter dem Testzuhause aus
	// `TestMain` findet dieser Weg sonst nichts, und der Lauf springt ab
	// statt rot zu werden (#1170).
	if ort := os.Getenv("MSE_KANON"); ort != "" {
		return ort
	}

	return filepath.Join(os.Getenv("HOME"), "Library", "Application Support",
		"MagicSetEditor4Mac", "kanonische-sets")
}
