package service_test

// **Eine Partie aus- und einführen** — PLAY-03.AK-05
// (#1711).

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Was hinausgeht, kommt als neue Partie zurück.**
func TestEinePartieGehtDurchDieDateiUndZurueck(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineAbgelegtePartie(t, domain.StandBeendet, 1)

	roh, fehler := leitung.Ausfuhr(context.Background(), "p1")
	if fehler != nil {
		t.Fatalf("Ausfuhr: %v", fehler)
	}

	eingelesen, befunde, fehler := leitung.Einfuhr(
		context.Background(), roh, "p2")
	if fehler != nil {
		t.Fatalf("Einfuhr: %v", fehler)
	}

	if len(befunde) != 0 {
		t.Errorf("unerwartete Befunde: %v", befunde)
	}

	if eingelesen.ID != "p2" || eingelesen.Stand != domain.StandLaeuft {
		t.Errorf("eingelesen wurde %q im Stand %q",
			eingelesen.ID, eingelesen.Stand)
	}

	if eingelesen.Fassung != 1 {
		t.Errorf("die eingelesene Partie steht bei Fassung %d",
			eingelesen.Fassung)
	}

	// **Ohne alte Peer-Zustimmungen** — die Einwilligung der Vorlage bleibt
	// bei der Vorlage.
	zahl, fehler := speicher.Offenlegungszahl(context.Background(), "p2")
	if fehler != nil || zahl != 0 {
		t.Errorf("die eingelesene Partie trägt %d Einwilligungen (%v)",
			zahl, fehler)
	}
}

// **Eine Partie, die es nicht gibt, geht nicht hinaus.**
func TestDieAusfuhrMeldetDieUnbekanntePartie(t *testing.T) {
	t.Parallel()

	leitung, _ := eineLeitung(t)

	if _, fehler := leitung.Ausfuhr(
		context.Background(), "gibtsnicht"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("erwartet ErrNotFound, bekam %v", fehler)
	}
}

// **Eine Datei, die nicht taugt, legt nichts an.**
func TestDieEinfuhrWeistEineFremdeDateiAb(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineLeitung(t)

	_, _, fehler := leitung.Einfuhr(
		context.Background(), []byte(`{"fassung":99}`), "p2")
	if !errors.Is(fehler, domain.ErrValidation) {
		t.Fatalf("erwartet ErrValidation, bekam %v", fehler)
	}

	if len(speicher.partien) != 0 {
		t.Errorf("angelegt wurde trotzdem: %v", speicher.partien)
	}
}

// **Ein fremdes Feld ist ein Befund und kein Fehler** — die Partie entsteht,
// und dass etwas darin stand, das niemand kennt, wird gesagt (§8, §16).
func TestDieEinfuhrReichtDenBefundNachOben(t *testing.T) {
	t.Parallel()

	leitung, _ := eineAbgelegtePartie(t, domain.StandLaeuft, 0)

	roh, fehler := leitung.Ausfuhr(context.Background(), "p1")
	if fehler != nil {
		t.Fatalf("Ausfuhr: %v", fehler)
	}

	var felder map[string]json.RawMessage

	if lesefehler := json.Unmarshal(roh, &felder); lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	felder["übermorgen"] = json.RawMessage(`true`)

	mitFremdem, fehler := json.Marshal(felder)
	if fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	_, befunde, fehler := leitung.Einfuhr(
		context.Background(), mitFremdem, "p2")
	if fehler != nil {
		t.Fatalf("Einfuhr: %v", fehler)
	}

	if len(befunde) != 1 || !strings.Contains(befunde[0].Grund, "übermorgen") {
		t.Fatalf("erwartet ein Befund, bekam %v", befunde)
	}
}

// **Und ein Bestand, der beim Anlegen klagt, legt nichts an.**
func TestDieEinfuhrReichtDenSchreibfehlerDurch(t *testing.T) {
	t.Parallel()

	leitung, speicher := eineAbgelegtePartie(t, domain.StandLaeuft, 0)

	roh, fehler := leitung.Ausfuhr(context.Background(), "p1")
	if fehler != nil {
		t.Fatalf("Ausfuhr: %v", fehler)
	}

	speicher.stolpert = map[string]error{
		"anlegen": errors.New("der Bestand ist voll"),
	}

	if _, _, fehler := leitung.Einfuhr(
		context.Background(), roh, "p2"); fehler == nil ||
		!strings.Contains(fehler.Error(), "voll") {
		t.Fatalf("erwartet den Schreibfehler, bekam %v", fehler)
	}
}
