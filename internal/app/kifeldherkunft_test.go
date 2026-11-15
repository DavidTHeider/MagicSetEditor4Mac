package app_test

// **Die Herkunft eines übernommenen Werts** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Was hier gemessen wird
//
// **Der Weg zurück.** Geschrieben wird sie in der Transaktion der Übernahme —
// das prüft `storage`. Hier steht, dass sie **ankommt**: eine Herkunft, die
// niemand lesen kann, ist keine.

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Was geschrieben wurde, kommt zurück.**
//
// # Warum hier über den Bestand und nicht über die Übernahme
//
// **Weil die Vorrichtung dieser Läufe eine Kartenablage einsetzt** — sie
// zählt, was ankommt, und schreibt nichts in die Datenbank. Der Weg der
// Übernahme ist in `internal/service` und `internal/storage` geprüft; was
// **hier** offen ist, ist der Weg **zurück**: aus dem Bestand ins Fenster.
func TestWasGeschriebenWurdeKommtZurueck(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	karteID := schreibeHerkunft(t, welt, "Alt", "Schreibe den Regeltext.")

	antwort := welt.anwendung.KIFeldherkunft(probekontext(t), karteID)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Herkunft) != 1 {
		t.Fatalf("die Herkunft ist %+v", antwort.Herkunft)
	}

	einer := antwort.Herkunft[0]

	if einer.Feld != "name" || einer.Anbieter != "ollama" ||
		einer.Modell != "gemma4:8b-q4" {
		t.Errorf("die Herkunft ist %+v", einer)
	}

	// **Die Aufgabe ist, was getippt wurde** — und nicht der Rohprompt.
	if einer.Aufgabe != "Schreibe den Regeltext." {
		t.Errorf("die Aufgabe ist %q", einer.Aufgabe)
	}

	if !strings.HasPrefix(einer.Zeitpunkt, "20") || einer.Auftrag == "" {
		t.Errorf("die Herkunft ist %+v", einer)
	}

	// **Und der Wert ist noch der übernommene** — die Karte trägt „Alt", und
	// der Abdruck spricht von „Alt".
	if einer.Veraendert {
		t.Error("die Herkunft sagt »von Hand geändert«, ohne dass jemand war")
	}
}

// **Wer den Wert danach ändert, bekommt es gesagt.**
//
// # Warum das keine Klage ist
//
// **Manuelle Nachbearbeitung ist erlaubt** — das Issue sagt es ausdrücklich.
// Was nicht erlaubt ist: dass die Herkunft danach behauptet, der Inhalt
// stamme von der KI.
func TestEineHandarbeitDanachStehtInDerHerkunft(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	// **Der Abdruck spricht von einem Wert, den die Karte nicht trägt** —
	// genau das, was nach einem Handgriff dasteht.
	karteID := schreibeHerkunft(t, welt, "Blitzschlag", "x")

	antwort := welt.anwendung.KIFeldherkunft(probekontext(t), karteID)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Herkunft) != 1 || !antwort.Herkunft[0].Veraendert {
		t.Fatalf("die Herkunft ist %+v", antwort.Herkunft)
	}
}

// schreibeHerkunft legt eine Herkunft in den Bestand — über den echten Weg.
func schreibeHerkunft(t *testing.T, welt kiwelt, wert, aufgabe string) int64 {
	t.Helper()

	karteID := eineEchteKarte(t, welt)

	if _, _, fehler := storage.NeuerKartenschreiber(welt.db).
		FelderSetzenMitHerkunft(probekontext(t), karteID,
			[]domain.Value{{Field: "name", Text: wert}},
			"ki", []domain.KIFeldherkunft{{
				Feld: "name", Anbieter: "ollama", Modell: "gemma4:8b-q4",
				Aufgabe: aufgabe, Auftrag: "ki-4-1",
				Zeitpunkt: time.Now().UTC(),
				Bezug:     domain.KIInhaltsbezug([]string{wert}),
			}}, nil, ""); fehler != nil {
		t.Fatalf("die Herkunft schreiben: %v", fehler)
	}

	welt.karten.zeile.ID = karteID

	return karteID
}

// eineEchteKarte legt eine Karte im Bestand an — die Ablage der Vorrichtung
// hat keine.
func eineEchteKarte(t *testing.T, welt kiwelt) int64 {
	t.Helper()

	setID, fehler := storage.NeuerSetSpeicher(welt.db).SetAnlegen(
		probekontext(t), &domain.Set{
			Name: "Probeset", Game: "magic",
			Cards: []*domain.Card{{
				Values: []domain.Value{{Field: "name", Text: "Alt"}},
			}},
		}, nil)
	if fehler != nil {
		t.Fatalf("das Set anlegen: %v", fehler)
	}

	karten, fehler := storage.NeuerSetLeser(welt.db).Karten(probekontext(t), setID)
	if fehler != nil {
		t.Fatalf("die Karten lesen: %v", fehler)
	}

	return karten[0].ID
}

// **Ohne Übernahme steht keine Herkunft da — und das ist kein Fehler.**
func TestOhneUebernahmeStehtKeineHerkunftDa(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	antwort := welt.anwendung.KIFeldherkunft(probekontext(t), welt.karte)
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Herkunft) != 0 {
		t.Errorf("die Herkunft ist %+v", antwort.Herkunft)
	}
}

// **Ohne Bestand gibt sie eine Liste und einen Fehler.**
//
// **Beides, und das ist der Punkt**: `null.length` hält das Fenster an — das
// ist in A9 einmal passiert, und dort stand derselbe Satz.
func TestOhneBestandGibtDieHerkunftEineLeereListe(t *testing.T) {
	t.Parallel()

	antwort := app.NeueApp("").KIFeldherkunft(probekontext(t), 4)
	if antwort.Fehler == nil {
		t.Fatal("ohne Bestand kam kein Fehler")
	}

	if antwort.Herkunft == nil {
		t.Error("die Herkunft ist null statt einer leeren Liste")
	}
}

// **Und eine Karte, die sich nicht lesen lässt, hält sie auf.**
//
// **Weil der Vergleich sie braucht**: ohne den heutigen Wert ließe sich nicht
// sagen, ob von Hand daran gearbeitet wurde — und „nein" zu antworten, ohne
// nachzusehen, wäre eine Behauptung.
func TestEineUnlesbareKarteHaeltDieHerkunftAuf(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.karten.fehler = errUnlesbar

	if antwort := welt.anwendung.KIFeldherkunft(probekontext(t),
		welt.karte); antwort.Fehler == nil {
		t.Fatal("die unlesbare Karte gab eine Herkunft")
	}
}

// errUnlesbar ist der Fehler, mit dem die Kartenablage sich weigert.
var errUnlesbar = errors.New("die Karte ließ sich nicht lesen")

// **Und ein Bestand, der nicht mehr antwortet, wird gemeldet.**
//
// **Nicht als leere Herkunft**: „kein Feld stammt aus einem Vorschlag" und
// „ich konnte nicht nachsehen" sind zwei verschiedene Auskünfte, und die
// zweite ist die, nach der jemand handelt.
func TestEinGeschlossenerBestandHaeltDieHerkunftAuf(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)

	if fehler := welt.db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	if antwort := welt.anwendung.KIFeldherkunft(probekontext(t),
		welt.karte); antwort.Fehler == nil {
		t.Fatal("der geschlossene Bestand gab eine Herkunft")
	}
}

// **Ohne Kartenleser steht der Vorschlag ohne Vergleich da — und trotzdem da.**
//
// # Bestenfalls-Arbeit (§7, Regel 3)
//
// **Ein Vorschlag ohne den alten Wert ist weniger wert als einer mit; gar
// keiner ist schlechter als beides.** Deshalb hält ein Lesefehler hier nichts
// auf: was fehlt, ist der Vergleich, nicht der Vorschlag.
func TestOhneKartenleserStehtDerVorschlagOhneVergleichDa(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	welt.karten.fehler = errUnlesbar

	// **Der Vergleich allein, ohne den Weg davor.** Über [App.KIVorschlagen]
	// käme man hier nicht an: der Auftrag friert die Revision ein und
	// scheitert an derselben Ablage. Was bliebe, wäre ein Zweig, den kein
	// Lauf betritt — und der sähe aus wie eine geprüfte Zusage (§10).
	felder := app.MitBisherigemZumPruefen(welt.anwendung, probekontext(t),
		welt.karte, []app.KIFeldwertDTO{
			{Feld: "name", Werte: []string{"Blitzschlag"}},
		})

	if len(felder) != 1 || felder[0].Bisher != "" {
		t.Errorf("die Felder sind %+v", felder)
	}
}

// **Und ohne jeden Bestand gibt es keinen Vergleich und keine Klage.**
func TestOhneBestandGibtEsKeinenVergleich(t *testing.T) {
	t.Parallel()

	felder := app.MitBisherigemZumPruefen(app.NeueApp(""), probekontext(t), 4,
		[]app.KIFeldwertDTO{{Feld: "name", Werte: []string{"Blitzschlag"}}})

	if len(felder) != 1 || felder[0].Bisher != "" {
		t.Errorf("die Felder sind %+v", felder)
	}
}

// **Die Bindung gibt die Herkunft — und weist eine unlesbare Kennung ab.**
//
// **Beides am selben Ort**: die Bindung ist der Weg, den das Fenster geht,
// und sie bekommt ihre Kennung als **Text**. Wer sie nicht prüfte, reichte
// eine 0 weiter und bekäme die Herkunft von gar nichts.
func TestDieBindungGibtDieHerkunft(t *testing.T) {
	t.Parallel()

	welt := einekiwelt(t)
	karteID := schreibeHerkunft(t, welt, "Alt", "x")

	// **Die Bindung baut ihren Kontext selbst** — und das Profil kommt aus
	// dem Fenster, das `Bereit` setzt.
	welt.anwendung.Bereit(probekontext(t))

	bindungen := app.NeueBindungen(welt.anwendung)

	antwort := bindungen.KIFeldherkunft(t.Context(), strconv.FormatInt(karteID, 10))
	if antwort.Fehler != nil {
		t.Fatalf("%+v", antwort.Fehler)
	}

	if len(antwort.Herkunft) != 1 {
		t.Fatalf("die Herkunft ist %+v", antwort.Herkunft)
	}

	// **Und eine Kennung, die keine Zahl ist, gibt eine Liste und einen
	// Fehler** — `null.length` hält das Fenster an.
	krumm := bindungen.KIFeldherkunft(t.Context(), "keine Zahl")
	if krumm.Fehler == nil || krumm.Herkunft == nil {
		t.Errorf("die krumme Kennung gab %+v", krumm)
	}
}
