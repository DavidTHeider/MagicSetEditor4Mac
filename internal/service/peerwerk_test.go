package service_test

// **Die Identität dieses Profils gegenüber anderen** — Werkstück N1 von
// Phase 5.4
// (#1454).
//
// # Die Frage, für die es dieses Werk gibt
//
// **Zwei Quellen, eine Identität.** Der private Schlüssel liegt im
// Schlüsselbund, die Kennung im Bestand — und was gilt, sagt erst der
// Vergleich. Das Issue ist an dieser Stelle scharf: „Schlüsselverlust,
// Rotation und Profilkopie dürfen **nie still** eine fremde Identität
// übernehmen."
//
// **Vier Stände, und drei davon sind kein Fehler**: es gab noch keine, es
// gibt eine, sie ist verloren, sie widerspricht sich. Nur der letzte ist ein
// Grund anzuhalten — und der vorletzte ist es auch, aber aus dem anderen
// Grund: dort **fehlt** etwas, statt dass zwei Dinge streiten.

import (
	"context"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// derPeerlauf ist der Zeitpunkt dieser Läufe.
var derPeerlauf = time.Date(2026, 9, 14, 3, 0, 0, 0, time.UTC)

// vermerkprobe ist der Bestand, der die Kennung hält.
type vermerkprobe struct {
	vermerk    domain.Peervermerk
	lesefehler error
	merkfehler error
	rotiert    int
}

func (v *vermerkprobe) Lesen(context.Context) (domain.Peervermerk, error) {
	return v.vermerk, v.lesefehler
}

func (v *vermerkprobe) Merken(_ context.Context, kennung string, jetzt time.Time) error {
	if v.merkfehler != nil {
		return v.merkfehler
	}

	v.vermerk = domain.Peervermerk{Kennung: kennung, Seit: jetzt}

	return nil
}

func (v *vermerkprobe) Rotieren(_ context.Context, kennung string, jetzt time.Time) error {
	if v.merkfehler != nil {
		return v.merkfehler
	}

	v.rotiert++
	v.vermerk = domain.Peervermerk{
		Kennung: kennung, Seit: jetzt, Vorher: v.vermerk.Kennung,
	}

	return nil
}

// bundprobe ist der Schlüsselbund.
type bundprobe struct {
	geheimnis     string
	lesefehler    error
	schreibfehler error
	geschrieben   int
	gelesen       int

	// lesefehlerAb lässt erst den n-ten Leseruf scheitern — für den Weg,
	// der **zweimal** liest (`Ausweis` fragt die Lage und danach den Bund).
	lesefehlerAb int
}

func (b *bundprobe) Lies(context.Context, string) (string, error) {
	b.gelesen++

	if b.lesefehlerAb != 0 && b.gelesen >= b.lesefehlerAb {
		return "", domain.MitGrund(domain.ErrConflict, nil,
			"der Bund antwortet nicht mehr")
	}

	if b.lesefehler != nil {
		return "", b.lesefehler
	}

	if b.geheimnis == "" {
		return "", domain.MitGrund(domain.ErrNotFound, nil, "kein Geheimnis")
	}

	return b.geheimnis, nil
}

func (b *bundprobe) Schreibe(_ context.Context, _, geheimnis string) error {
	if b.schreibfehler != nil {
		return b.schreibfehler
	}

	b.geschrieben++
	b.geheimnis = geheimnis

	return nil
}

// einPeerwerk baut das Werk über zwei Proben.
func einPeerwerk(
	t *testing.T, vermerk *vermerkprobe, bund *bundprobe,
) *service.Peerwerk {
	t.Helper()

	return service.NeuesPeerwerk(vermerk, bund, &kontaktprobe{},
		func() time.Time { return derPeerlauf })
}

// kontaktprobe zählt, wie oft eine Rotation die Kontakte zurückstellt.
type kontaktprobe struct {
	gerufen   int
	scheitert error
}

func (k *kontaktprobe) NachRotation(context.Context) error {
	k.gerufen++

	return k.scheitert
}

// eineKennung gibt eine gültige Identität samt ihrem Geheimnis.
func eineKennung(t *testing.T, stand byte) (kennung peer.Kennung, geheimnis string) {
	t.Helper()

	ich, fehler := peer.Erzeugen(&fester{stand: stand})
	if fehler != nil {
		t.Fatalf("Erzeugen: %v", fehler)
	}

	return ich.Kennung(), ich.Geheimnis()
}

// fester gibt immer dieselben Bytes.
type fester struct{ stand byte }

func (f *fester) Read(ziel []byte) (int, error) {
	for i := range ziel {
		ziel[i] = f.stand
		f.stand++
	}

	return len(ziel), nil
}

// **Ohne beides gibt es noch keine Identität — und das ist kein Fehler.**
func TestOhneSchluesselUndOhneVermerkGibtEsNochKeine(t *testing.T) {
	t.Parallel()

	lage, fehler := einPeerwerk(t, &vermerkprobe{}, &bundprobe{}).
		Lage(t.Context())
	if fehler != nil {
		t.Fatalf("Lage: %v", fehler)
	}

	if lage.Stand != domain.PeerstandOhne {
		t.Fatalf("der Stand ist %q", lage.Stand)
	}

	if lage.Kennung != "" {
		t.Errorf("ohne Identität steht %q da", lage.Kennung)
	}
}

// **Passen Schlüssel und Vermerk zusammen, ist sie bereit.**
func TestSchluesselUndVermerkZusammenSindBereit(t *testing.T) {
	t.Parallel()

	kennung, geheimnis := eineKennung(t, 1)

	lage, fehler := einPeerwerk(t,
		&vermerkprobe{vermerk: domain.Peervermerk{
			Kennung: string(kennung), Seit: derPeerlauf,
		}},
		&bundprobe{geheimnis: geheimnis}).Lage(t.Context())
	if fehler != nil {
		t.Fatalf("Lage: %v", fehler)
	}

	if lage.Stand != domain.PeerstandBereit {
		t.Fatalf("der Stand ist %q", lage.Stand)
	}

	if lage.Kennung != kennung {
		t.Errorf("die Kennung ist %q", lage.Kennung)
	}

	// **Der Fingerabdruck steht dabei** — er ist das, was zwei Menschen
	// vergleichen, und ihn erst auf Nachfrage zu rechnen hieße, dass die
	// Fläche ihn vergessen kann.
	if lage.Fingerabdruck != peer.Fingerabdruck(kennung) {
		t.Errorf("der Abdruck ist %q", lage.Fingerabdruck)
	}
}

// **Ein Vermerk ohne Schlüssel ist ein Verlust — und wird als solcher gesagt.**
//
// **Das ist die kopierte Bestandsdatei**: sie bringt die Kennung mit, der
// fremde Rechner hat den Schlüssel nicht. Wer hier still eine neue Identität
// anlegte, hätte den Peer getauscht, und jeder vorgelesene Fingerabdruck
// zeigte ins Leere.
func TestEinVermerkOhneSchluesselIstEinVerlust(t *testing.T) {
	t.Parallel()

	kennung, _ := eineKennung(t, 1)

	werk := einPeerwerk(t,
		&vermerkprobe{vermerk: domain.Peervermerk{
			Kennung: string(kennung), Seit: derPeerlauf,
		}},
		&bundprobe{})

	lage, fehler := werk.Lage(t.Context())
	if fehler != nil {
		t.Fatalf("Lage: %v", fehler)
	}

	if lage.Stand != domain.PeerstandVerloren {
		t.Fatalf("der Stand ist %q", lage.Stand)
	}

	// **Die verlorene Kennung steht dabei** — ohne sie könnte niemand einem
	// Kontakt sagen, welche Identität fort ist.
	if lage.Kennung != kennung {
		t.Errorf("die Kennung ist %q", lage.Kennung)
	}
}

// **Und ein Schlüssel, der nicht zum Vermerk passt, ist ein Widerspruch.**
func TestEinFremderSchluesselIstEinWiderspruch(t *testing.T) {
	t.Parallel()

	vermerkt, _ := eineKennung(t, 1)
	_, anderes := eineKennung(t, 99)

	lage, fehler := einPeerwerk(t,
		&vermerkprobe{vermerk: domain.Peervermerk{
			Kennung: string(vermerkt), Seit: derPeerlauf,
		}},
		&bundprobe{geheimnis: anderes}).Lage(t.Context())
	if fehler != nil {
		t.Fatalf("Lage: %v", fehler)
	}

	if lage.Stand != domain.PeerstandFremd {
		t.Fatalf("der Stand ist %q", lage.Stand)
	}
}

// **Ein Schlüssel ohne Vermerk wird vermerkt — und das steht dabei.**
//
// **Nicht still**: der Bestand kann zurückgesetzt worden sein, und dann ist
// die Reparatur richtig. Sie **stillschweigend** zu tun hieße, dass niemand
// merkt, dass der Bestand jünger ist als die Identität.
func TestEinSchluesselOhneVermerkWirdVermerkt(t *testing.T) {
	t.Parallel()

	kennung, geheimnis := eineKennung(t, 1)
	vermerke := &vermerkprobe{}

	lage, fehler := einPeerwerk(t, vermerke, &bundprobe{geheimnis: geheimnis}).
		Lage(t.Context())
	if fehler != nil {
		t.Fatalf("Lage: %v", fehler)
	}

	if lage.Stand != domain.PeerstandBereit {
		t.Fatalf("der Stand ist %q", lage.Stand)
	}

	if vermerke.vermerk.Kennung != string(kennung) {
		t.Errorf("vermerkt wurde %q", vermerke.vermerk.Kennung)
	}

	if len(lage.Befunde) != 1 {
		t.Fatalf("die Befunde sind %v", lage.Befunde)
	}
}
