package app

// **Auch das Fenster speichert, was gerechnet ist** — #2396
// (Issue (#2396)).
//
// Die Kommandozeile und der Dienst hält `internal/flaechen/wertnachzug_test.go`;
// das Fenster hat seinen eigenen Composition Root (`verdrahtung.go`), und der
// Nachzug hängt dort erst am Zeichenwerk. Ein Root ohne ihn speicherte den
// eingegebenen Wert — und zeigte nach #2396 einen anderen als die Flächen ohne
// Fenster.

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// nachzugsfenster startet die Anwendung über einer Saat, deren Spiel zwei
// Felder mit Wertskript trägt, und gibt die erste Karte samt einem Leser
// ihres Bestands.
func nachzugsfenster(t *testing.T) (anwendung *App, setID string, karteID int64, werte func() map[string]string) {
	t.Helper()

	wurzel, fehler := filepath.Abs(filepath.Join("..", ".."))
	if fehler != nil {
		t.Fatalf("Wurzel: %v", fehler)
	}

	// **Eine Kopie der Saat**, deren Spiel zwei Felder mit Wertskript trägt.
	saat := filepath.Join(t.TempDir(), "saat")
	if fehler = os.CopyFS(saat, os.DirFS(filepath.Join(wurzel, "saat"))); fehler != nil {
		t.Fatalf("Saat kopieren: %v", fehler)
	}

	spiel := filepath.Join(saat, "probe.mse-game", "game")

	roh, fehler := os.ReadFile(spiel) //nolint:gosec // die Kopie dieses Laufs
	if fehler != nil {
		t.Fatalf("Spiel lesen: %v", fehler)
	}

	const felder = "card field:\n\ttype: text\n\tname: zeile\n\tscript: \"[\" + value + \"]\"\n" +
		"card field:\n\ttype: text\n\tname: echo\n\tscript: card.zeile + \"!\"\n\n"

	//nolint:gosec // G703: die Kopie dieses Laufs
	if fehler = os.WriteFile(spiel, []byte(strings.Replace(string(roh), "set field:",
		felder+"set field:", 1)), 0o600); fehler != nil {
		t.Fatalf("Spiel schreiben: %v", fehler)
	}

	anwendung = gestartetMit(t, func() string { return saat })

	quelle := anwendung.SetOeffnen(t.Context(), saatset(t))
	if quelle.Fehler != nil {
		t.Fatalf("öffnen: %+v", quelle.Fehler)
	}

	karten := anwendung.Karten(t.Context(), quelle.SetID)
	if karten.Fehler != nil || len(karten.Karten) == 0 {
		t.Fatalf("Karten: %+v", karten.Fehler)
	}

	karteID, _ = strconv.ParseInt(karten.Karten[0].ID, 10, 64) //nolint:errcheck // aus der Liste

	werte = func() map[string]string {
		karte, fehler := storage.NeuerSetLeser(anwendung.db).KarteLesen(probekontext(t), karteID)
		if fehler != nil || karte == nil {
			t.Fatalf("lesen: %v", fehler)
		}

		gelesen := map[string]string{}
		for _, wert := range karte.Values {
			gelesen[wert.Field] = wert.Text
		}

		return gelesen
	}

	return anwendung, quelle.SetID, karteID, werte
}

func TestDasFensterSpeichertDenGerechnetenWert(t *testing.T) {
	anwendung, setID, karteID, werte := nachzugsfenster(t)

	if gesetzt := anwendung.FeldSetzen(probekontext(t), setID, strconv.FormatInt(karteID, 10),
		"zeile", "Funke"); gesetzt.Fehler != nil {
		t.Fatalf("FeldSetzen: %+v", gesetzt.Fehler)
	}

	if gelesen := werte(); gelesen["zeile"] != "[Funke]" || gelesen["echo"] != "[Funke]!" {
		t.Errorf("im Bestand steht zeile=%q, echo=%q — erwartet [Funke] und [Funke]!", gelesen["zeile"], gelesen["echo"])
	}
}

// **Auch die Übernahme einer Gruppe** (A13.1b, #1452): sie geht im Fenster an
// der Aufsicht vorbei in den Bestand, wenn die Naht sie nicht über den
// Feldweg reicht.
func TestDasFensterSpeichertAuchInEinerGruppeDenGerechnetenWert(t *testing.T) {
	anwendung, setID, karteID, werte := nachzugsfenster(t)

	set, _ := strconv.ParseInt(setID, 10, 64) //nolint:errcheck // vom Öffnen

	if fehler := anwendung.schreiber.GruppeVonAussenMitHerkunft(probekontext(t), []domain.Kartenschritt{{
		KarteID: karteID, SetID: set, Herkunft: "ki",
		Werte: []domain.Value{{Field: "zeile", Text: "Glut"}},
		Herkuenfte: []domain.KIFeldherkunft{{
			Feld: "zeile", Anbieter: "probe", Modell: "probe", Aufgabe: "probe",
			Auftrag: "probe", Bezug: "probe", Zeitpunkt: time.Unix(1, 0),
		}},
	}}); fehler != nil {
		t.Fatalf("Gruppe: %v", fehler)
	}

	if gelesen := werte(); gelesen["zeile"] != "[Glut]" || gelesen["echo"] != "[Glut]!" {
		t.Errorf("im Bestand steht zeile=%q, echo=%q — erwartet [Glut] und [Glut]!", gelesen["zeile"], gelesen["echo"])
	}
}
