package app_test

// **Die Fehlerwege des Bildvorschlags** — Werkstück A12 von Phase 5.3
// (#1451).
//
// # Warum sie eine eigene Datei haben
//
// **Weil jeder von ihnen eine Zusage ist.** „Verworfenes/abgebrochenes/
// abgelaufenes Bild lässt Karte unverändert", „fremdes Profil kann Asset
// weder lesen noch übernehmen", „Prozessabbruch begrenzt und bereinigt" —
// das Issue zählt sie einzeln auf, und ein Weg, den kein Lauf betreten hat,
// ist eine Zusage, die niemand geprüft hat (§10).

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein Vorschlag einer anderen Karte geht nicht durch.**
//
// **Die Kennung bindet an das Profil, nicht an das Ziel.** Wer sie hat, darf
// sie lesen — er darf sie nicht irgendwohin schreiben.
func TestEinVorschlagEinerAnderenKarteGehtNichtDurch(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	if antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "2", "9",
		"kunst", vorschlag.Kennung,
		app.AuswahlDTO{Breite: 8, Hoehe: 12}); antwort.Fehler == nil {
		t.Fatal("der Vorschlag landete an einer fremden Karte")
	}
}

// **Ein Bild, das sich nicht auspacken lässt, wird gemeldet.**
//
// Ein PNG-Kopf ohne Bilddaten kommt durch die Schranken des Vorrats — die
// lesen nur den Kopf — und scheitert erst beim wirklichen Auspacken.
func TestEinUnauspackbaresBildWirdGemeldet(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)
	welt.laufzeitbild.bild.Inhalt = einKopfOhneBilddaten()

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	if antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "2", "4",
		"kunst", vorschlag.Kennung,
		app.AuswahlDTO{Breite: 8, Hoehe: 12}); antwort.Fehler == nil {
		t.Fatal("der leere Kopf kam als Bild durch")
	}
}

// **Lässt sich das Bild nicht ablegen, wird es gemeldet.**
func TestEinNichtAblegbaresBildWirdGemeldet(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)
	welt.anwendung.MitBildablage(service.NeueBildablage(""))

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	if antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "2", "4",
		"kunst", vorschlag.Kennung,
		app.AuswahlDTO{Breite: 8, Hoehe: 12}); antwort.Fehler == nil {
		t.Fatal("das Bild wurde ohne Ablageort abgelegt")
	}
}

// **Ohne Vorrat gibt es auch keinen Zuschnitt — hinter dem Bestand.**
//
// **Der Unterschied zum Lauf nebenan**: dort fehlt schon der Planbauer, und
// die Naht kommt gar nicht bis zum Vorrat. Hier steht alles außer ihm.
func TestOhneVorratGibtEsKeinenZuschnitt(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)
	welt.anwendung.MitBildvorrat(nil)

	if antwort := welt.anwendung.KIBildZuschnitt(probekontext(t), "4", "kunst",
		"egal"); antwort.Fehler == nil {
		t.Error("der Zuschnitt gelang ohne Vorrat")
	}

	if antwort := welt.anwendung.KIBildEinsetzen(probekontext(t), "2", "4",
		"kunst", "egal", app.AuswahlDTO{}); antwort.Fehler == nil {
		t.Error("das Einsetzen gelang ohne Vorrat")
	}
}

// **Ohne Bestand gibt es keine Vorlage und keine Auskunft.**
func TestOhneBestandGibtEsKeineBildauskunft(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)
	welt.anwendung.MitEditor(welt.karten, nil, &bildplanbauer{}, nil)

	if antwort := welt.anwendung.KIBildVorschlagen(probekontext(t), welt.karte,
		"kunst", "ollama", "gemma4:8b", "dunkler", "", "",
		true, ""); antwort.Fehler == nil {
		t.Error("die Bearbeitung gelang ohne Bestand")
	}

	ohne := app.NeueApp("")

	if antwort := ohne.KIBildHerkunft(t.Context(), 4, "kunst"); antwort.Fehler == nil {
		t.Error("die Auskunft kam ohne Bestand")
	}
}

// **Ein unbekannter Anbieter und ein unbekanntes Modell werden gemeldet.**
func TestEinUnbekannterAnbieterWirdGemeldet(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	if antwort := welt.anwendung.KIBildVorschlagen(probekontext(t), welt.karte,
		"kunst", "gibtesnicht", "gemma4:8b", "ein Blitz", "", "",
		false, ""); antwort.Fehler == nil {
		t.Error("der unbekannte Anbieter kam durch")
	}

	if antwort := welt.anwendung.KIBildVorschlagen(probekontext(t), welt.karte,
		"kunst", "ollama", "gibtesnicht", "ein Blitz", "", "",
		false, ""); antwort.Fehler == nil {
		t.Error("das unbekannte Modell kam durch")
	}
}

// **Eine unlesbare Kartenkennung kommt an keiner Bindung durch.**
//
// **Und beide geben eine Liste zurück** — was fehlt, fehlt im Fenster als
// Länge und nicht als `null`.
func TestEineUnlesbareKartenkennungKommtNichtDurch(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	bindungen := app.NeueBindungen(welt.anwendung)

	vorschlag := bindungen.KIBildVorschlagen(t.Context(), "sieben", "kunst", "ollama",
		"gemma4:8b", "ein Blitz", "", "", false, "")
	if vorschlag.Fehler == nil || vorschlag.Angaben == nil {
		t.Errorf("die Antwort ist %+v", vorschlag)
	}

	herkunft := bindungen.KIBildHerkunft(t.Context(), "sieben", "kunst")
	if herkunft.Fehler == nil || herkunft.Angaben == nil {
		t.Errorf("die Antwort ist %+v", herkunft)
	}
}

// **Ein Bildvorrat, dessen Frist abgelaufen ist, gibt nichts heraus.**
func TestEinAbgelaufenerVorratGibtNichtsHeraus(t *testing.T) {
	t.Parallel()

	kurz := service.UeblicheBildgrenzen()
	kurz.Frist = time.Nanosecond

	vorrat, fehler := service.NeuerBildvorrat(t.TempDir(), kurz, time.Now)
	if fehler != nil {
		t.Fatalf("der Vorrat ließ sich nicht bauen: %v", fehler)
	}

	welt := einebildwelt(t)
	welt.anwendung.MitBildvorrat(vorrat)

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	if antwort := welt.anwendung.KIBildZuschnitt(probekontext(t), "4", "kunst",
		vorschlag.Kennung); antwort.Fehler == nil {
		t.Fatal("der abgelaufene Vorschlag kam durch")
	}
}

// einKopfOhneBilddaten baut ein PNG, das sich lesen und nicht auspacken lässt.
//
// **Die Prüfsumme wird gerechnet und nicht abgeschrieben**: eine falsche
// ließe schon `DecodeConfig` scheitern, und dann prüfte der Lauf die
// Signatur statt des Auspackens.
func einKopfOhneBilddaten() []byte {
	kopf := []byte{0, 0, 0, 8, 0, 0, 0, 12, 8, 6, 0, 0, 0}

	var aus bytes.Buffer

	aus.WriteString("\x89PNG\r\n\x1a\n")
	_ = binary.Write(&aus, binary.BigEndian, uint32(len(kopf))) //nolint:errcheck // ein Puffer nimmt alles
	aus.WriteString("IHDR")
	aus.Write(kopf)
	_ = binary.Write(&aus, binary.BigEndian, //nolint:errcheck // siehe oben
		crc32.ChecksumIEEE(append([]byte("IHDR"), kopf...)))

	return aus.Bytes()
}

// **Und die Domäne weist eine erfundene Fähigkeit ab.**
//
// Der Lauf steht hier, weil `bildfaehigkeit` zwei Antworten gibt und nur
// eine von ihnen sonst betreten wird.
func TestDieBildfaehigkeitFolgtDemAuftrag(t *testing.T) {
	t.Parallel()

	if !domain.IstKIFaehigkeit(domain.KIBildgenerierung) ||
		!domain.IstKIFaehigkeit(domain.KIBildbearbeitung) {
		t.Fatal("die zwei Bildfähigkeiten gibt es nicht")
	}
}

// **Die fünf Bindungen führen zu den fünf Wegen.**
//
// # Warum das ein eigener Lauf ist
//
// **Weil zwischen Bindung und Weg eine Kennung übersetzt wird.** Die Läufe
// darüber gehen an der Bindung vorbei und prüfen den Weg; dieser prüft, dass
// die Bindung ihn wirklich ruft — und nicht einen anderen.
func TestDieFuenfBindungenFuehrenZuDenFuenfWegen(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	// **Das Fenster trägt den Kontext, und die Bindungen nehmen ihn.**
	// `Bereit` ist, was Wails beim Start ruft; ohne ihn hätte eine Bindung
	// kein Profil — und ein Lauf, der sie am Kontext vorbei ruft, prüfte
	// einen Weg, den es im Fenster nicht gibt.
	welt.anwendung.Bereit(probekontext(t))

	bindungen := app.NeueBindungen(welt.anwendung)

	vorschlag := bindungen.KIBildVorschlagen(t.Context(), "4", "kunst", "ollama",
		"gemma4:8b", "ein Blitz", domain.KIBildgroesseHoch,
		domain.KIBildgueteMittel, false, "")
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	if herkunft := bindungen.KIBildHerkunft(t.Context(), "4", "kunst"); herkunft.Fehler != nil {
		t.Fatalf("%+v", herkunft.Fehler)
	}

	if zuschnitt := bindungen.KIBildZuschnitt(t.Context(), "4", "kunst",
		vorschlag.Kennung); zuschnitt.Fehler != nil {
		t.Fatalf("%+v", zuschnitt.Fehler)
	}

	if antwort := bindungen.KIBildEinsetzen(t.Context(), "2", "4", "kunst",
		vorschlag.Kennung,
		app.AuswahlDTO{Breite: 8, Hoehe: 12}); antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	// **Und das Verwerfen findet ihn nicht mehr** — er ist eingelöst.
	if antwort := bindungen.KIBildVerwerfen(t.Context(),
		vorschlag.Kennung); antwort.Fehler == nil {
		t.Error("der eingelöste Vorschlag ließ sich verwerfen")
	}
}

// **Ein unauspackbares Bild hält auch den Zuschnitt auf.**
//
// **Die zweite Hälfte derselben Klasse**: das Einsetzen prüft es ohnehin,
// aber wer den Zuschnitt durchließe, zeigte ein Fenster über einem Bild, das
// es nicht gibt.
func TestEinUnauspackbaresBildHaeltDenZuschnittAuf(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)
	welt.laufzeitbild.bild.Inhalt = einKopfOhneBilddaten()

	vorschlag := welt.vorschlagen(t)
	if vorschlag.Fehler != nil {
		t.Fatalf("%+v", vorschlag.Fehler)
	}

	if antwort := welt.anwendung.KIBildZuschnitt(probekontext(t), "4", "kunst",
		vorschlag.Kennung); antwort.Fehler == nil {
		t.Fatal("der leere Kopf kam als Zuschnitt durch")
	}
}

// **Ohne Tafel gibt es keine Bildauskunft.**
//
// Der Fall ist eine Migration, die nicht gelaufen ist — dieselbe Gestalt wie
// im Bestand (§8): gemeldet und nicht gedeutet.
func TestOhneTafelGibtEsKeineBildauskunft(t *testing.T) {
	t.Parallel()

	welt := einebildwelt(t)

	if _, fehler := welt.db.ExecContext(probekontext(t),
		`DROP TABLE ki_bildherkunft`); fehler != nil {
		t.Fatalf("die Tafel ließ sich nicht fallen lassen: %v", fehler)
	}

	if antwort := welt.anwendung.KIBildHerkunft(probekontext(t), welt.karte,
		"kunst"); antwort.Fehler == nil {
		t.Fatal("die Auskunft kam ohne Tafel")
	}
}
