//go:build fremdpakete

package app_test

import (
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Der Abstand zu MSE2, beziffert statt behauptet** — Phase A,
// #290,
// entschieden in
// [0021](../../docs/entscheidungen/0021-das-referenzbild-kommt-vom-nutzenden.md).
//
// # Was dieser Lauf ist, und was er nicht ist
//
// Er ist der **einzige** Vergleich dieses Projekts gegen einen fremden Maßstab.
// `render/goldbild_test.go` hält den Rasterer fest, `app/saatgoldbild_test.go`
// die Naht — beide gegen uns selbst. Ein Rundlauf prüft Selbstkonsistenz und
// nicht Wahrheit; das ist die Lehre aus Phase 9, und dieser Lauf ist die
// Antwort darauf.
//
// **Er ist rot, und er soll es sein.** Er misst den heutigen Abstand. Jede
// Phase von B bis F wird daran gemessen, ob sie die Zahl bewegt.
//
// # Was er nicht prüft
//
//   - **Das Bildfenster.** Die Illustration liegt uns nicht vor; das
//     Referenzbild ist ein fertig gerechnetes JPG. Ausgespart wird das Rechteck,
//     das die **Referenz selbst** unter `bildfenster:` nennt.
//
//     **Nicht die größte Bildebene des Plans.** Das war der erste Anlauf, und
//     der Zweitblick über Phase A hat ihn erledigt: in
//     `magic-m15-future-planeswalker` ist die Ebene `typeline` 339×464 groß und
//     die Illustration nur 328×309. Sobald eine spätere Phase solche Ebenen
//     baut, wechselte die ausgesparte Fläche — und eine „Verbesserung" entstünde
//     allein daraus, dass anderswo gemessen wird. Eine Geometrie sagt nicht, was
//     ein Bildfenster ist.
//
//   - **Verlustfreiheit.** JPG komprimiert verlustbehaftet. Schon die
//     Kompression allein erzeugt einen Abstand über der Toleranz. Die Zahl ist
//     als **Verlauf** zu lesen, nicht als absoluter Betrag.
//
//   - **Die Styling-Optionen der Karte.** Was der Urheber im Stil eingestellt
//     hat, steht nicht im Bild. Wo eine Referenz danebenliegt, weil eine Option
//     anders stand, ist das ein Fehler der Referenz und nicht der Anwendung —
//     deshalb steht neben jeder Zahl ihr Name.
func TestDerAbstandZuMSE2(t *testing.T) {
	referenzen := os.Getenv("MSE_REFERENZBILDER")
	if referenzen == "" {
		t.Skip("MSE_REFERENZBILDER ist nicht gesetzt")
	}

	wurzel := os.Getenv("MSE_FREMDPAKETE")
	if wurzel == "" {
		t.Skip("MSE_FREMDPAKETE ist nicht gesetzt")
	}

	eintraege, err := os.ReadDir(referenzen)
	if err != nil {
		t.Fatalf("Referenzen lesen: %v", err)
	}

	bestandsausweis(t, referenzen, eintraege)

	gemessen := 0

	for _, eintrag := range eintraege {
		name, istKarte := strings.CutSuffix(eintrag.Name(), ".karte")
		if !istKarte {
			continue
		}

		gemessen++

		t.Run(name, func(t *testing.T) {
			messeGegenMSE2(t, referenzen, wurzel, name)
		})
	}

	// **Ohne diese Zeile wäre ein leeres Verzeichnis ein Erfolg.** Ein Lauf, der
	// nichts vergleicht, ist grün und sagt nichts — die stillste Art, eine
	// Messung wegzudefinieren (§10).
	if gemessen == 0 {
		t.Fatalf("in %s liegt keine einzige .karte", referenzen)
	}
}

// messeGegenMSE2 rendert eine Referenzkarte und beziffert ihren Abstand.
func messeGegenMSE2(t *testing.T, referenzen, wurzel, name string) {
	t.Helper()

	karte, stilname, angaben := referenzkarte(t, filepath.Join(referenzen, name+".karte"))

	stil := fremdstil(t, wurzel, stilname)

	suche := schriften.Neue(schriften.Orte()).MitPaketen(filepath.Dir(wurzel))

	// **Dieselbe Bildquelle für beide Stufen**, so wie im Betrieb: der
	// Planbauer misst damit `content_width`, der Rasterer zeichnet damit
	// (§3 — eine Vorrichtung, die ärmer ist als der Betrieb, prüft den
	// falschen Weg).
	bilder := msepaket.NeueBildquelle(wurzel)

	plan, befunde, err := service.NeuerPlanbauer().
		MitSchriften(suche).MitBildern(bilder).Stilplan(
		probekontext(t), karte,
		fremdvorlage(t, stil, wurzel))
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}

	bild, rasterbefunde, err := render.
		NeuerRasterer(bilder).
		MitSchriften(suche).
		Zeichne(probekontext(t), plan, 1, render.GueteMittel)
	if err != nil {
		t.Fatalf("zeichnen: %v", err)
	}

	gold := liesReferenzbild(t, filepath.Join(referenzen, name+".jpg"))

	ganz, err := render.Vergleiche(gold, bild)
	if err != nil {
		t.Fatalf("vergleichen: %v", err)
	}

	fenster := bildfensterAus(t, angaben.Ersten("bildfenster"))

	ohne, err := render.VergleicheAusserhalb(gold, bild, fenster)
	if err != nil {
		t.Fatalf("vergleichen ohne Bildfenster: %v", err)
	}

	t.Logf("ganze Karte:      %s", ganz)
	t.Logf("ohne Bildfenster: %s  (ausgespart %v)", ohne, fenster)
	t.Logf("%d Befunde beim Planbauen, %d beim Rastern", len(befunde), len(rasterbefunde))

	// **Jeder Befund einzeln.** Eine Zahl sagt, *dass* etwas fehlt; erst der
	// Ort sagt, *was* — und ein Abstand von 33 ist ohne diese Liste nicht zu
	// deuten.
	for _, befund := range befunde {
		t.Logf("  Befund %s: %s", befund.Ort, befund.Meldung)
	}

	for _, befund := range rasterbefunde {
		t.Logf("  Rasterbefund: %s", befund)
	}

	legeBildAb(t, render.Differenzbild(gold, bild), name+"-differenz.png")
	legeBildAb(t, bild, name+"-unsere.png")

	if !ohne.Innerhalb() {
		t.Errorf("der Abstand zu MSE2 hält die Toleranz nicht: %s", ohne)
	}
}

// fremdstil öffnet ein bezogenes Stilpaket samt seinen Modulen.
func fremdstil(t *testing.T, wurzel, name string) msepaket.Stil {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join(wurzel, name+".mse-style"))
	if err != nil {
		t.Fatalf("Stilpaket %s: %v", name, err)
	}

	_ = paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)) //nolint:errcheck // Befunde sind hier kein Abbruch
	stil, _ := paket.Stilmodell()
	_ = paket.Schliessen() //nolint:errcheck // das Modell ist kopiert

	return stil
}

// fremdspielskript liest das Initskript des Spielpakets.
func fremdspielskript(t *testing.T, wurzel, spiel string) string {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join(wurzel, spiel+".mse-game"))
	if err != nil {
		t.Fatalf("Spielpaket %s: %v", spiel, err)
	}

	_ = paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)) //nolint:errcheck // dito
	skript := paket.Initskript()
	_ = paket.Schliessen() //nolint:errcheck // das Skript ist kopiert

	return skript
}

// liesReferenzbild lädt die MSE2-Ausgabe.
func liesReferenzbild(t *testing.T, pfad string) image.Image {
	t.Helper()

	datei, err := os.Open(pfad) //nolint:gosec // ein Pfad aus einer Umgebungsvariable des Laufs
	if err != nil {
		t.Fatalf("Referenzbild: %v", err)
	}

	defer func() { _ = datei.Close() }() //nolint:errcheck // gelesen ist gelesen

	bild, _, err := image.Decode(datei)
	if err != nil {
		t.Fatalf("Referenzbild lesen: %v", err)
	}

	return bild
}

// legeBildAb schreibt ein Bild neben den Lauf — das Differenzbild (§3) und
// unsere eigene Ausgabe.
//
// # Warum auch unsere eigene
//
// Eine Zahl allein sagt nicht, ob sie eine Aussage ist oder ein Fehler. Bei
// einem Abstand von 33 ist die erste Frage, ob überhaupt etwas gezeichnet
// wurde — und die beantwortet nur ein Blick.
//
// **In `MSE_AUSGABE` oder ins Temporärverzeichnis, nie ins Repository.** Ein
// Differenzbild gegen eine MSE2-Ausgabe zeigt deren Rahmen und Illustration;
// §11 lässt das nicht zu. Der Pfad steht im Protokoll.
func legeBildAb(t *testing.T, bild image.Image, name string) {
	t.Helper()

	ziel := os.Getenv("MSE_AUSGABE")
	if ziel == "" {
		ziel = t.TempDir()
	}

	pfad := filepath.Join(ziel, name)

	datei, err := os.Create(pfad) //nolint:gosec // ein Pfad unter t.TempDir() oder MSE_AUSGABE
	if err != nil {
		t.Fatalf("Bild anlegen: %v", err)
	}

	if err := png.Encode(datei, bild); err != nil {
		t.Fatalf("Bild schreiben: %v", err)
	}

	if err := datei.Close(); err != nil {
		t.Fatalf("Bild schließen: %v", err)
	}

	t.Logf("abgelegt: %s", pfad)
}

// fremdvorlage baut die Vorlage **so, wie der Betrieb sie baut**.
//
// # Warum das hier steht
//
// Bis zum 2026-08-16 reichte diese Messung nur Stil und Spielskript weiter —
// **keine Symbolschriften**. Sie maß damit eine Umgebung, die es im Betrieb
// nicht gibt, und ein `<sym>` im Regeltext konnte gar nicht ankommen.
//
// Es ist dieselbe Klasse, die schon `fremdpakete_test.go` getroffen hat: eine
// Vorrichtung, die **strenger** ist als der Betrieb, meldet Befunde, die
// niemand hat — und ihre Zahl taugt dann nicht als Maß.
func fremdvorlage(t *testing.T, stil msepaket.Stil, wurzel string) service.Vorlage {
	t.Helper()

	// **`symbolschriften` und nicht `schriften`**: der Name verdeckte sonst das
	// gleichnamige Paket, und ein späterer Zugriff darauf schlüge in dieser
	// Funktion fehl, ohne dass jemand den Grund sähe.
	symbolschriften, befunde := service.NeuerStilspeicher(nil, wurzel).Symbolschriften(stil)
	for _, befund := range befunde {
		t.Logf("  Symbolschrift: %s", befund.Meldung)
	}

	// **Und die Feldbeschreibungen des Spiels** — so wie `Stilspeicher.Vorlage`
	// sie im Betrieb mitgibt.
	//
	// Ohne sie hat kein Kartenfeld einen Vorgabewert, und `card.card_symbol`
	// ist nichts statt `none`. `magic-m15` schreibt
	// `left: { if card.card_symbol == "none" then 32 else 50 }` — der
	// Kartenname stand damit auf **jeder** gemessenen Karte 18 Punkte zu weit
	// rechts, und die Zahl daneben sprach über einen Weg, den niemand geht
	// (§3, #341).
	spielmodell, fassungen := fremdspielmodell(t, wurzel, stil.Spiel)

	return service.Vorlage{
		Stil:        stil,
		Spielskript: fremdspielskript(t, wurzel, stil.Spiel),

		Kartenfelder: spielmodell.Fields,
		Setfelder:    spielmodell.SetFields,

		Symbolschriften: symbolschriften,

		// **Und die Fassungen des Setsymbols aus dem Spiel** — `common`,
		// `uncommon`, `rare`, `mythic rare`. Ohne sie meldete jede
		// Referenzkarte „symbol_variation kennt die Fassung rare nicht" und
		// zeichnete kein Seltenheitssymbol: derselbe Mangel wie bei den
		// Symbolschriften, eine Vorrichtung **strenger** als der Betrieb
		// (#270).
		Symbolfassungen: fassungen,
	}
}

// fremdspielmodell liest die Feldbeschreibungen und Symbolfassungen eines Spiels.
func fremdspielmodell(
	t *testing.T, wurzel, spiel string,
) (domain.Game, []msepaket.Symbolfassung) {
	t.Helper()

	paket, err := msepaket.Oeffne(filepath.Join(wurzel, spiel+".mse-game"))
	if err != nil {
		t.Fatalf("Spielpaket %s: %v", spiel, err)
	}

	_ = paket.ModuleEinbinden(msepaket.NeuerPaketleser(wurzel)) //nolint:errcheck // Befunde sind hier kein Abbruch
	modell, _ := paket.Spielmodell()
	fassungen := paket.Spielsymbolfassungen()
	_ = paket.Schliessen() //nolint:errcheck // Modell und Fassungen sind kopiert

	return modell, fassungen
}
