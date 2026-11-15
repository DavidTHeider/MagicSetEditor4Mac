// Der Wächter über die Naht, die #1257 gefehlt hat.
//
// # Warum es ihn braucht
//
// Die Bild-Zwischenablage war **gebaut, geprüft und nicht erreichbar**:
// `internal/service` konnte sie, `ablagebilder_test.go` prüfte sie über eine
// selbst gebaute `Kartenablage` — und der Composition Root des Fensters rief
// `MitBildern` nicht. Über diesen Weg reiste ein Bildfeld weiter ohne seine
// Datei, und `docs/PARITAET.md` sagte das Gegenteil.
//
// **Das ist die erste Dimension aus §10**, und sie steht dort zuerst, weil sie
// schon einmal gefehlt hat: ein Paket ohne Aufrufer ist keine
// Anwendungsfunktion. Angestrichen vom Nachlauf über die Nacharbeit
// (#1282,
// Befund 1).
//
// # Warum durch die ganze Anwendung
//
// Weil genau dazwischen der Fehler lag. Ein Lauf, der eine `Kartenablage`
// selbst baut, hätte ihn nie gesehen — er baut ja gerade die Naht, die fehlte.
package app

import (
	"context"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// **Ein Bildfeld nimmt seine Datei mit — über den Weg des Fensters.**
func TestDieZwischenablageNimmtBilderMit(t *testing.T) {
	anwendung := mitSaat(t)

	quelle := anwendung.SetOeffnen(t.Context(), saatset(t))
	if quelle.Fehler != nil {
		t.Fatalf("öffnen: %+v", quelle.Fehler)
	}

	karten := anwendung.Karten(t.Context(), quelle.SetID)
	if karten.Fehler != nil || len(karten.Karten) == 0 {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	// **Ein echtes Bild an einer echten Karte**, über denselben Weg, den das
	// Fenster nimmt: die Ablage vergibt den Namen, die Karte trägt ihn.
	karteID := karten.Karten[0].ID

	quellkennung, befund := kennungAus(quelle.SetID)
	if befund != nil {
		t.Fatalf("Kennung: %+v", befund)
	}

	bild := image.NewRGBA(image.Rect(0, 0, 2, 2))
	bild.Set(0, 0, color.RGBA{R: 200, G: 30, B: 30, A: 255})

	bildname, ablegefehler := anwendung.ablagen().Ablegen(quellkennung, "image", bild)
	if ablegefehler != nil {
		t.Fatalf("Bild ablegen: %v", ablegefehler)
	}

	if gesetzt := anwendung.FeldSetzen(
		probekontext(t), quelle.SetID, karteID, "image", bildname,
	); gesetzt.Fehler != nil {
		t.Fatalf("FeldSetzen: %+v", gesetzt.Fehler)
	}

	var inhalt string

	anwendung.ablageSchreiben = func(_ context.Context, text string) error {
		inhalt = text

		return nil
	}

	if kopiert := anwendung.AblageKopieren(
		probekontext(t), quelle.SetID, []string{karteID},
	); kopiert.Fehler != nil {
		t.Fatalf("AblageKopieren: %+v", kopiert.Fehler)
	}

	if !strings.Contains(inhalt, "image_data") {
		t.Fatalf("der Ablageinhalt trägt keine Bilddaten — `MitBildern` fehlt "+
			"im Composition Root:\n%.400s", inhalt)
	}

	// **Ein zweites Set desselben Spiels**, damit der Name im Ziel neu
	// vergeben werden muss. `saatset` legt bei jedem Ruf eine eigene Datei an.
	ziel := anwendung.SetOeffnen(t.Context(), saatset(t))
	if ziel.Fehler != nil {
		t.Fatalf("zweites Set öffnen: %+v", ziel.Fehler)
	}

	anwendung.ablageLesen = func(context.Context) (string, error) { return inhalt, nil }

	eingefuegt := anwendung.AblageEinfuegen(t.Context(), ziel.SetID)
	if eingefuegt.Fehler != nil {
		t.Fatalf("AblageEinfuegen: %+v", eingefuegt.Fehler)
	}

	if len(eingefuegt.Karten) == 0 {
		t.Fatal("keine Karte eingefügt")
	}

	gelesen := anwendung.Karte(t.Context(), eingefuegt.Karten[0])
	if gelesen.Fehler != nil {
		t.Fatalf("Karte: %+v", gelesen.Fehler)
	}

	var neu string

	for _, wert := range gelesen.Karte.Werte {
		if wert.Feld == "image" {
			neu = wert.Wert
		}
	}

	if neu == "" {
		t.Fatal("die eingefügte Karte trägt keinen Bildwert")
	}

	// **Und im Zielset liegt wirklich eine Datei** — nicht der Verweis auf die
	// des Quellsets.
	zielkennung, befund := kennungAus(ziel.SetID)
	if befund != nil {
		t.Fatalf("Kennung: %+v", befund)
	}

	pfad := filepath.Join(anwendung.ablagen().Verzeichnis(zielkennung), neu)
	if _, err := os.Stat(pfad); err != nil {
		t.Errorf("im Zielset liegt keine Datei %q: %v — der Verweis zeigt ins "+
			"Leere, und die Quelle heißt %q", neu, err, bildname)
	}
}

// **Ohne Zwischenspeicher gibt es keinen Bildvorrat — und keinen Absturz.**
//
// # Warum das eine eigene Zusage ist
//
// **Weil der Vorrat eine Beigabe ist und kein Fundament** (A12, #1451). Wo
// sich kein Zwischenspeicherverzeichnis ermitteln lässt, soll die Anwendung
// weiterlaufen; was fehlt, sind die Bildvorschläge, und jede ihrer Bindungen
// sagt das.
//
// **Zwei Wege dorthin**, und beide kommen auf echten Rechnern vor: eine
// Umgebung ohne Zuhause und ein Ort, an dem sich nichts anlegen lässt.
func TestOhneZwischenspeicherGibtEsKeinenBildvorrat(t *testing.T) {
	for name, zuhause := range map[string]func(*testing.T) string{
		"ohne Zuhause": func(*testing.T) string { return "" },
		"mit unbeschreibbarem Zuhause": func(t *testing.T) string {
			t.Helper()

			heim := t.TempDir()
			if fehler := os.Chmod(heim, 0o500); fehler != nil { //nolint:gosec // ein Verzeichnis braucht das Ausführrecht
				t.Fatalf("die Rechte ließen sich nicht setzen: %v", fehler)
			}

			t.Cleanup(func() { _ = os.Chmod(heim, 0o700) }) //nolint:errcheck,gosec // damit t.TempDir aufräumen kann

			return heim
		},
	} {
		t.Run(name, func(t *testing.T) {
			// **Nicht parallel**: `t.Setenv` verträgt das nicht, und das
			// Zuhause ist eine Eigenschaft des Prozesses.
			t.Setenv("HOME", zuhause(t))

			anwendung := NeueApp("")
			anwendung.ablagen()

			if anwendung.bildvorrat != nil {
				t.Fatal("es entstand ein Vorrat, obwohl es keinen Ort gibt")
			}
		})
	}
}
