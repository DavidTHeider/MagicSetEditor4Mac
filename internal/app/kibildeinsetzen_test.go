package app_test

// **Vom Vorschlag in die Karte** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Was hier gemessen wird
//
// **Dass die Karte sich erst beim Einsetzen ändert — und dann vollständig.**
// Das Feld trägt den Dateinamen, die Herkunft nennt Anbieter und Modell, die
// Angaben nennen Format und Güte, und der Vorschlag ist danach fort.
//
// **Und dass ein Vorschlag nur dorthin geht, wofür er gedacht war.** Die
// Kennung sagt, dass er diesem Profil gehört; sie sagt nicht, dass er in
// dieses Feld gehört.

import (
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Eingesetzt steht das Bild in der Karte — samt seiner Herkunft.**
func TestEinEingesetztesBildTraegtSeineHerkunft(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "2", "4",
		"kunst", vorschlag.Kennung, app.AuswahlDTO{Breite: 8, Hoehe: 12})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if werte := welt.kartenwerte(t); !strings.Contains(werte, "kunst") {
		t.Errorf("die Karte trägt %q", werte)
	}

	// **Die Herkunft ist mitgegangen** — in derselben Transaktion wie der
	// Wert (A11, #1450).
	herkuenfte := welt.karten.feldherkuenfte
	if len(herkuenfte) != 1 || herkuenfte[0].Feld != "kunst" ||
		herkuenfte[0].Anbieter != "ollama" {
		t.Fatalf("die Herkünfte sind %+v", herkuenfte)
	}

	if herkuenfte[0].Modell != "gpt-image-2.5-flare" {
		t.Errorf("das Modell ist %q — geantwortet hat ein anderes",
			herkuenfte[0].Modell)
	}

	// **Und die Bildangaben daneben.**
	if len(welt.karten.bildangaben["kunst"]) != 4 {
		t.Errorf("die Angaben sind %+v", welt.karten.bildangaben)
	}

	// **Der Vorschlag ist danach fort** — er ist eingelöst.
	if _, fehler := welt.vorrat.Schauen(vorschlag.Kennung, 1); fehler == nil {
		t.Error("der Vorschlag liegt noch im Vorrat")
	}
}

// **Ein Vorschlag für ein anderes Feld geht nicht durch.**
//
// # Warum das eine eigene Prüfung ist
//
// **Weil die Kennung nur das Profil bindet.** Ein Bild, das für die
// Illustration erzeugt wurde, gehört nicht in das Wasserzeichen — und wer
// sich vertut, hat es dort stehen, ohne dass etwas gemeldet wurde.
func TestEinVorschlagFuerEinAnderesFeldGehtNichtDurch(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "2", "4",
		"wasserzeichen", vorschlag.Kennung,
		app.AuswahlDTO{Breite: 8, Hoehe: 12})
	if antwort.Fehler == nil {
		t.Fatal("der Vorschlag landete im falschen Feld")
	}

	if !strings.Contains(antwort.Fehler.Meldung, "kunst") {
		t.Errorf("die Meldung nennt das richtige Feld nicht: %s",
			antwort.Fehler.Meldung)
	}
}

// **Der Zuschnitt geht durch denselben Weg wie eine Datei.**
func TestDerZuschnittEinesVorschlagsKommtAn(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	antwort := welt.anwendung.KIBildZuschnitt(probekontext(t), "4", "kunst",
		vorschlag.Kennung)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if antwort.QuelleBreite != 8 || antwort.QuelleHoehe != 12 {
		t.Errorf("die Quelle ist %d×%d", antwort.QuelleBreite, antwort.QuelleHoehe)
	}

	// **Und der Vorschlag liegt noch da** — ein Zuschnitt lässt sich
	// abbrechen (§7, Regel 3).
	if _, fehler := welt.vorrat.Schauen(vorschlag.Kennung, 1); fehler != nil {
		t.Errorf("der Vorschlag ist fort: %v", fehler)
	}
}

// **Ein verworfener Vorschlag ist weg — und die Karte unverändert.**
func TestEinVerworfenerVorschlagLaesstDieKarteUnveraendert(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	vorher := welt.kartenwerte(t)

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	if antwort := welt.anwendung.KIBildVerwerfen(probekontext(t),
		vorschlag.Kennung); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if nachher := welt.kartenwerte(t); nachher != vorher {
		t.Errorf("die Karte wurde geändert: %q statt %q", nachher, vorher)
	}

	// **Und ein zweites Mal geht nicht** — er ist weg.
	if antwort := welt.anwendung.KIBildVerwerfen(probekontext(t),
		vorschlag.Kennung); antwort.Fehler == nil {
		t.Error("der Vorschlag ließ sich zweimal verwerfen")
	}
}

// **Eine unbekannte Kennung setzt nichts ein.**
func TestEineUnbekannteKennungSetztNichtsEin(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	if antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "2", "4",
		"kunst", "gibtesnicht",
		app.AuswahlDTO{Breite: 8, Hoehe: 12}); antwort.Fehler == nil {
		t.Error("die unbekannte Kennung kam durch")
	}

	if antwort := welt.anwendung.KIBildZuschnitt(probekontext(t), "4", "kunst",
		"gibtesnicht"); antwort.Fehler == nil {
		t.Error("die unbekannte Kennung kam beim Zuschnitt durch")
	}

	// **Und ein Feld, das kein Bild trägt, auch nicht** — es gibt keine
	// Fläche, auf die zugeschnitten würde.
	if antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "2", "4",
		"name", "egal", app.AuswahlDTO{}); antwort.Fehler == nil {
		t.Error("das Textfeld kam als Bildziel durch")
	}
}

// **Scheitert das Schreiben, bleibt kein Bild ohne Verweis liegen.**
//
// **Dieselbe Zusage wie beim Weg von der Platte** (#1163, Z7-02): das Bild
// liegt, bevor ein Feld auf es zeigt.
func TestEinGescheitertesSchreibenLaesstKeinBildLiegen(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	welt.karten.schreibfehler = domain.MitGrund(domain.ErrConflict, nil,
		"der Bestand nimmt nichts an")

	if antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "2", "4",
		"kunst", vorschlag.Kennung,
		app.AuswahlDTO{Breite: 8, Hoehe: 12}); antwort.Fehler == nil {
		t.Fatal("das gescheiterte Schreiben kam als Erfolg zurück")
	}

	// **Und der Vorschlag liegt noch da** — wer es erneut versuchen will,
	// hat noch etwas dafür.
	if _, fehler := welt.vorrat.Schauen(vorschlag.Kennung, 1); fehler != nil {
		t.Errorf("der Vorschlag ist fort: %v", fehler)
	}
}

// **Die Angaben eines übernommenen Bildes lassen sich lesen.**
func TestDieAngabenEinesUebernommenenBildesStehenDa(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	antwort := welt.anwendung.KIBildHerkunft(probekontext(t), welt.karte, "kunst")
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Eine Liste und kein `null`** — auch ohne Angaben.
	if antwort.Angaben == nil {
		t.Fatal("die Angaben sind null")
	}
}

// **Der Bildvorschlag trägt den Stand der Karte bis zum Einsetzen** — Befund 2
// des Nachlaufs über Phase 5.3.
//
// # Der Befund
//
// **Hier stand, ein Bildvorschlag friere keinen Stand ein.** Er tut es:
// `KIBildVorschlagen` ruft `werk.Anlegen`, und das liest die Revision der
// Karte — auch für ein Bild. Nur kam sie nicht im Vorrat an, und das
// Einsetzen schrieb ohne Bedingung.
//
// **Ein Bildauftrag dauert Sekunden bis Minuten.** Genau in dieser Zeit ändert
// jemand die Karte, und der Einsatz überschrieb sie still — dieselbe Klasse,
// gegen die der Textweg seit dem Zweitblick geschützt ist.
//
// **Gemessen wird die Naht und nicht die Wirkung**: dass eine Bedingung in
// der Transaktion **hält**, prüft `internal/storage/kistandbedingung_test.go`.
// Hier steht, dass sie überhaupt dorthin gelangt.
func TestEinBildvorschlagTraegtDenStandBisZumEinsetzen(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	stand := welt.karten.zeile.Stand()

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "2", "4",
		"kunst", vorschlag.Kennung, app.AuswahlDTO{Breite: 8, Hoehe: 12})
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(welt.karten.erwartet) != 1 || welt.karten.erwartet[0] != stand {
		t.Fatalf("an der Ablage kam %q an, erwartet %q",
			welt.karten.erwartet, stand)
	}
}
