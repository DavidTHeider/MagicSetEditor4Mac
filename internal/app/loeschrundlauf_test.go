package app_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// **Die schärfste Probe auf die Vollständigkeit des Rückwegs**
// (#621):
// löschen, zurücknehmen, schreiben — und die Datei ist byteweise dieselbe.
//
// # Warum nicht der Test in `internal/storage` genügt
//
// Dort steht, dass Kennung, Reihenfolge und Feldwerte zurückkommen. Was dort
// **nicht** steht: ob die **Rohzeilen** der Karte mit zurückkommen. `.mse-set`
// ist keine Tabelle, sondern eine eingerückte Textdatei; Kommentare, Abstände
// und die Reihenfolge der Zeilen stehen in `card_raw` und hängen an derselben
// `ON DELETE CASCADE` wie die Werte. Ein Rückweg, der sie vergisst, ist in
// jedem Datenbanktest grün und schreibt trotzdem eine andere Datei
// (#48).
//
// **Zwei Stellen, weil die Ränder anders liegen**: eine Karte in der Mitte
// lässt eine Lücke zwischen zwei Nachbarn, die letzte lässt sie am Ende. Wer
// nur die Mitte prüft, prüft die Lücke am Ende nicht.
func TestEineGeloeschteUndZurueckgeholteKarteSchreibtDieselbeDatei(t *testing.T) {
	for _, fall := range []struct {
		name   string
		stelle func(anzahl int) int
	}{
		{name: "die erste", stelle: func(int) int { return 0 }},
		{name: "die letzte", stelle: func(anzahl int) int { return anzahl - 1 }},
	} {
		t.Run(fall.name, func(t *testing.T) {
			anwendung, setID := geoeffnetesProbeset(t)
			quelle := probesetpfad(t)

			karten := anwendung.Karten(t.Context(), setID)
			if karten.Fehler != nil {
				t.Fatalf("Karten: %+v", karten.Fehler)
			}

			if len(karten.Karten) < 2 {
				t.Fatalf("das Prüfstück trägt %d Karten, gebraucht werden zwei", len(karten.Karten))
			}

			betroffen := karten.Karten[fall.stelle(len(karten.Karten))].ID

			if antwort := anwendung.KarteLoeschen(t.Context(), setID, betroffen, false); antwort.Fehler != nil {
				t.Fatalf("löschen: %+v", antwort.Fehler)
			}

			// **Die Gegenprobe zuerst**: ohne sie prüfte der Test eine
			// Rücknahme, die nichts zurückzunehmen hatte.
			danach := anwendung.Karten(t.Context(), setID)
			if danach.Fehler != nil {
				t.Fatalf("Karten: %+v", danach.Fehler)
			}

			if len(danach.Karten) != len(karten.Karten)-1 {
				t.Fatalf("nach dem Löschen %d Karten, erwartet %d",
					len(danach.Karten), len(karten.Karten)-1)
			}

			zurueck := anwendung.Zuruecknehmen(t.Context(), setID, "")
			if zurueck.Fehler != nil {
				t.Fatalf("zurücknehmen: %+v", zurueck.Fehler)
			}

			ziel := filepath.Join(t.TempDir(), "wieder.mse-set")

			gespeichert := anwendung.SetSpeichern(t.Context(), setID, ziel)
			if gespeichert.Fehler != nil {
				t.Fatalf("speichern: %+v", gespeichert.Fehler)
			}

			vorher := lies(t, filepath.Join(quelle, "set"))
			nachher := lies(t, filepath.Join(ziel, "set"))

			if !bytes.Equal(nachher, vorher) {
				t.Errorf("die Datei ist eine andere:\n%s", unterschied(vorher, nachher))
			}
		})
	}
}

// **Und ohne Rücknahme steht die Karte nicht mehr in der Datei** — genau
// einmal weniger, und die übrigen unverändert.
//
// Der Weg dorthin ging über eine Lücke, die jeder Datenbanktest übersah: der
// Schreiber paarte die `card:`-Blöcke der Quelldatei der Reihe nach mit den
// Karten des Bestands. Nach dem Löschen der mittleren von drei stand die
// dritte **zweimal** in der Datei und die zweite gar nicht
// (#621).
func TestEineGeloeschteKarteStehtNichtMehrInDerDatei(t *testing.T) {
	anwendung, setID := geoeffnetesProbeset(t)

	karten := anwendung.Karten(t.Context(), setID)
	if karten.Fehler != nil {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	if antwort := anwendung.KarteLoeschen(t.Context(), setID, karten.Karten[1].ID, false); antwort.Fehler != nil {
		t.Fatalf("löschen: %+v", antwort.Fehler)
	}

	geschrieben := string(gespeichertesSet(t, anwendung, setID))

	if zahl := strings.Count(geschrieben, "card:"); zahl != len(karten.Karten)-1 {
		t.Errorf("%d Kartenblöcke, erwartet %d:\n%s", zahl, len(karten.Karten)-1, geschrieben)
	}

	for _, name := range []string{"Feuerball", "Doppelt"} {
		if zahl := strings.Count(geschrieben, name); zahl != 1 {
			t.Errorf("%q steht %dmal in der Datei, erwartet einmal:\n%s", name, zahl, geschrieben)
		}
	}

	if strings.Contains(geschrieben, "Eisball") {
		t.Errorf("die gelöschte Karte steht noch in der Datei:\n%s", geschrieben)
	}

	// **Was nicht die Karten betrifft, bleibt unangetastet** — der Rest der
	// Datei ist keine Sammeländerung.
	if !strings.Contains(geschrieben, "unverstandener block:") {
		t.Errorf("der unverstandene Block ist weg:\n%s", geschrieben)
	}
}

// **Eine neu angelegte Karte steht in der geschriebenen Datei**
// (#623).
//
// Bis hierher nicht: der Schreiber lief nur die Blöcke durch, die die Datei
// schon hatte. Wer eine Karte anlegte und speicherte, bekam eine Datei ohne
// sie — stiller Datenverlust an der Stelle, die #562 gebaut hat.
func TestEineNeueKarteStehtInDerGeschriebenenDatei(t *testing.T) {
	anwendung, setID := geoeffnetesProbeset(t)

	angelegt := anwendung.KarteAnlegen(t.Context(), setID)
	if angelegt.Fehler != nil {
		t.Fatalf("anlegen: %+v", angelegt.Fehler)
	}

	if antwort := anwendung.FeldSetzen(t.Context(), setID, angelegt.KarteID, "name", "Neuling"); antwort.Fehler != nil {
		t.Fatalf("setzen: %+v", antwort.Fehler)
	}

	geschrieben := string(gespeichertesSet(t, anwendung, setID))

	if !strings.Contains(geschrieben, "name: Neuling") {
		t.Errorf("die neue Karte fehlt:\n%s", geschrieben)
	}

	// **Sie steht bei den Karten**, nicht hinter dem unverstandenen Block.
	if strings.Index(geschrieben, "Neuling") > strings.Index(geschrieben, "unverstandener block:") {
		t.Errorf("die neue Karte steht hinter dem unverstandenen Block:\n%s", geschrieben)
	}
}

// geoeffnetesProbeset gibt eine laufende Anwendung mit geöffnetem Prüfstück.
func geoeffnetesProbeset(t *testing.T) (anwendung *app.App, setID string) {
	t.Helper()

	anwendung = app.NeueApp(filepath.Join(t.TempDir(), "bestand.db"))

	if err := anwendung.Starten(t.Context()); err != nil {
		t.Fatalf("starten: %v", err)
	}

	t.Cleanup(func() {
		if fehler := anwendung.Beenden(); fehler != nil {
			t.Errorf("beenden: %v", fehler)
		}
	})

	geoeffnet := anwendung.SetOeffnen(t.Context(), probesetpfad(t))
	if geoeffnet.Fehler != nil {
		t.Fatalf("öffnen: %+v", geoeffnet.Fehler)
	}

	return anwendung, geoeffnet.SetID
}

// probesetpfad gibt den Ort des Prüfstücks.
func probesetpfad(t *testing.T) string {
	t.Helper()

	pfad, err := filepath.Abs(filepath.Join("..", "..", "testdaten", "probe.mse-set"))
	if err != nil {
		t.Fatalf("Pfad: %v", err)
	}

	return pfad
}

// gespeichertesSet schreibt das Set an einen frischen Ort und liest es zurück.
func gespeichertesSet(t *testing.T, anwendung *app.App, setID string) []byte {
	t.Helper()

	ziel := filepath.Join(t.TempDir(), "wieder.mse-set")

	if gespeichert := anwendung.SetSpeichern(t.Context(), setID, ziel); gespeichert.Fehler != nil {
		t.Fatalf("speichern: %+v", gespeichert.Fehler)
	}

	return lies(t, filepath.Join(ziel, "set"))
}
