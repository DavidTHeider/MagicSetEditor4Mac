// **Die Unsicherheit überlebt den Prozess** — Befund 10 des Zweitblicks über
// Phase 5.2 (#1435).
//
// **Ein weißer Lauf**, weil die Marke eine Datei neben dem Bestand ist und
// ihre Gestalt hier entsteht — der Weg dorthin geht durch `Zurueckspielen`,
// und der braucht ein Archiv, zwei Verzeichnisse und eine offene Datenbank.
package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// wegMitMarke gibt einen Bestandsweg, dessen Marke in einem eigenen Ordner
// liegt — samt seinem Buch.
func wegMitMarke(t *testing.T) (*Bestandsweg, *merkbuch) {
	t.Helper()

	buch := &merkbuch{}

	weg := &Bestandsweg{
		orte: Sicherungsorte{Datenbank: filepath.Join(t.TempDir(), "bestand.db")},
		uhr:  time.Now,
	}
	weg.MitRuecksprungschau(nil, nil, buch)

	return weg, buch
}

// **Ohne unklare Mengen entsteht keine Marke.**
func TestOhneUnklareMengenKeineEpochenmarke(t *testing.T) {
	weg, _ := wegMitMarke(t)

	if fehler := weg.epochenmarkeSchreiben(domain.Ruecksprunglage{}, "probe.zip",
		time.Now()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if _, err := os.Stat(weg.epochenmarkenpfad()); !os.IsNotExist(err) {
		t.Errorf("die Marke steht da: %v", err)
	}
}

// **Eine liegengebliebene Rückkehr wird nachgetragen — und die Marke geht.**
func TestEineLiegengebliebeneRueckkehrWirdNachgetragen(t *testing.T) {
	weg, buch := wegMitMarke(t)

	jetzt := time.Date(2026, time.September, 11, 8, 0, 0, 0, time.UTC)

	lage := domain.Ruecksprunglage{
		Datenstand: jetzt.Add(-time.Hour),
		Unklar: []domain.Sammlungsmenge{
			{Profil: 1, Schluessel: "probe-druck", Menge: 2},
		},
	}

	if fehler := weg.epochenmarkeSchreiben(lage, "probe.zip", jetzt); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if rest := weg.EpocheNachholen(t.Context()); rest != "" {
		t.Fatalf("der Rest ist %q", rest)
	}

	if len(buch.epochen) != 1 {
		t.Fatalf("im Buch stehen %d Epochen", len(buch.epochen))
	}

	// **Der Zeitpunkt kommt aus der Marke** und nicht von der Uhr beim
	// Nachholen: die Epoche gehört zum Tausch, und der war vorher.
	if !buch.epochen[0].Zurueckgespielt.Equal(jetzt) ||
		buch.epochen[0].Quelle != "probe.zip" ||
		len(buch.epochen[0].Mengen) != 1 {
		t.Errorf("die Epoche ist %+v", buch.epochen[0])
	}

	if _, err := os.Stat(weg.epochenmarkenpfad()); !os.IsNotExist(err) {
		t.Errorf("die Marke blieb liegen: %v", err)
	}
}

// **Und wenn das Buch klemmt, bleibt sie liegen.**
//
// Das ist die Zusage: sie wegzuräumen hieße, die Unsicherheit für geklärt zu
// erklären, ohne dass jemand nachgesehen hat.
func TestEineNichtVermerkteRueckkehrBleibtLiegen(t *testing.T) {
	weg, buch := wegMitMarke(t)

	if fehler := weg.epochenmarkeSchreiben(domain.Ruecksprunglage{
		HandpruefungNoetig: true,
	}, "probe.zip", time.Now()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	buch.fehler = errKeinBuch

	if rest := weg.EpocheNachholen(t.Context()); rest == "" {
		t.Error("es kam kein Satz fürs Fenster zurück")
	}

	if _, err := os.Stat(weg.epochenmarkenpfad()); err != nil {
		t.Errorf("die Marke ist weg: %v", err)
	}
}

// **Eine unlesbare Marke sagt es — und bleibt liegen.**
func TestEineUnlesbareEpochenmarkeMeldetSich(t *testing.T) {
	weg, _ := wegMitMarke(t)

	if fehler := os.WriteFile(weg.epochenmarkenpfad(), []byte("{"), 0o600); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if rest := weg.EpocheNachholen(t.Context()); rest == "" {
		t.Error("es kam kein Satz fürs Fenster zurück")
	}
}

// **Und ohne Marke geschieht nichts.**
func TestOhneEpochenmarkeGeschiehtNichts(t *testing.T) {
	weg, buch := wegMitMarke(t)

	if rest := weg.EpocheNachholen(t.Context()); rest != "" {
		t.Errorf("der Rest ist %q", rest)
	}

	if len(buch.epochen) != 0 {
		t.Errorf("im Buch stehen %d Epochen", len(buch.epochen))
	}
}

// errKeinBuch ist der Fehler, mit dem das Doppel klemmt.
var errKeinBuch = &markenfehler{}

type markenfehler struct{}

func (*markenfehler) Error() string { return "das Buch klemmt" }

// **Was die Marke trägt, kommt auch wieder heraus.**
func TestDieEpochenmarkeTraegtIhreAngaben(t *testing.T) {
	weg, _ := wegMitMarke(t)

	jetzt := time.Date(2026, time.September, 11, 8, 0, 0, 0, time.UTC)

	if fehler := weg.epochenmarkeSchreiben(domain.Ruecksprunglage{
		Datenstand: jetzt, HandpruefungNoetig: true,
	}, "probe.zip", jetzt); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	roh, fehler := os.ReadFile(weg.epochenmarkenpfad())
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	var stand epochenstand
	if fehler := json.Unmarshal(roh, &stand); fehler != nil {
		t.Fatalf("deuten: %v", fehler)
	}

	if stand.Quelle != "probe.zip" || !stand.Zurueckgespielt.Equal(jetzt) {
		t.Errorf("die Marke ist %+v", stand)
	}
}

// **Eine Marke, die sich nicht schreiben lässt, hält den Tausch auf.**
//
// **Und das ist Absicht**: sie steht **vor** dem Tausch, und ohne sie gäbe es
// keinen Nachweis für die Unsicherheit — also findet der Tausch nicht statt.
func TestEineUnschreibbareEpochenmarkeMeldetSich(t *testing.T) {
	weg := &Bestandsweg{
		orte: Sicherungsorte{
			Datenbank: filepath.Join(t.TempDir(), "gibtesnicht", "bestand.db"),
		},
		uhr: time.Now,
	}

	if fehler := weg.epochenmarkeSchreiben(domain.Ruecksprunglage{
		HandpruefungNoetig: true,
	}, "probe.zip", time.Now()); fehler == nil {
		t.Error("es kam durch")
	}
}

// **Eine Marke, die kein Aufrufer lesen kann, sagt es.**
func TestEineUnlesbareDateiMeldetSichAlsMarke(t *testing.T) {
	weg, _ := wegMitMarke(t)

	// **Ein Verzeichnis an ihrer Stelle** — `os.ReadFile` scheitert daran
	// mit etwas anderem als „gibt es nicht".
	if fehler := os.Mkdir(weg.epochenmarkenpfad(), 0o750); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if rest := weg.EpocheNachholen(t.Context()); rest == "" {
		t.Error("es kam kein Satz fürs Fenster zurück")
	}
}

// **Ohne Buch wird nichts nachgetragen** — und die Marke bleibt liegen.
func TestOhneBuchBleibtDieEpochenmarkeLiegen(t *testing.T) {
	weg := &Bestandsweg{
		orte: Sicherungsorte{Datenbank: filepath.Join(t.TempDir(), "bestand.db")},
		uhr:  time.Now,
	}

	if fehler := weg.epochenmarkeSchreiben(domain.Ruecksprunglage{
		HandpruefungNoetig: true,
	}, "probe.zip", time.Now()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	if rest := weg.EpocheNachholen(t.Context()); rest == "" {
		t.Error("es kam kein Satz fürs Fenster zurück")
	}

	if _, err := os.Stat(weg.epochenmarkenpfad()); err != nil {
		t.Errorf("die Marke ist weg: %v", err)
	}
}

// **Und eine Marke, die sich nicht wegräumen lässt, hält den Start nicht auf.**
func TestEineUnraeumbareEpochenmarkeHaeltNichtAuf(t *testing.T) {
	weg, buch := wegMitMarke(t)

	if fehler := weg.epochenmarkeSchreiben(domain.Ruecksprunglage{
		HandpruefungNoetig: true,
	}, "probe.zip", time.Now()); fehler != nil {
		t.Fatalf("schreiben: %v", fehler)
	}

	// **Der Ordner wird schreibgeschützt** — dann geht das Wegräumen nicht,
	// und das Lesen davor sehr wohl.
	ordner := filepath.Dir(weg.epochenmarkenpfad())
	schuetzen(t, ordner, 0o500)

	t.Cleanup(func() { schuetzen(t, ordner, 0o700) })

	if rest := weg.EpocheNachholen(t.Context()); rest != "" {
		t.Errorf("der Rest ist %q", rest)
	}

	if len(buch.epochen) != 1 {
		t.Errorf("im Buch stehen %d Epochen", len(buch.epochen))
	}
}

// **Lässt sich die Marke nicht schreiben, findet der Tausch nicht statt.**
//
// Sie steht **vor** dem Tausch, und ohne sie gäbe es keinen Nachweis für die
// Unsicherheit — der Weg endet also, statt einen Bestand zu tauschen, über
// dessen Mengen danach niemand etwas sagen könnte (Befund 10).
func TestOhneSchreibbareMarkeKeinTausch(t *testing.T) {
	bau := neuerWegbau(t)

	mitStaenden(bau.weg, bau.buch,
		besitz(4, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 4}),
		besitz(1, domain.Sammlungsmenge{Profil: 1, Schluessel: "scryfall:abc", Menge: 1}),
	)

	sicherung := probesicherung(t, bau)

	// **An der Stelle der Marke steht ein Verzeichnis.** Damit scheitert
	// genau ihr Schreiben — und sonst nichts: ein schreibgeschützter Ordner
	// ließe schon das Auspacken oder die Sicherung scheitern, und der Lauf
	// bewiese dann etwas anderes (§10).
	if fehler := os.Mkdir(bau.weg.epochenmarkenpfad(), 0o750); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if _, fehler := bau.weg.Zurueckspielen(t.Context(), sicherung); fehler == nil {
		t.Fatal("es kam durch")
	}

	if len(bau.buch.epochen) != 0 {
		t.Errorf("im Buch stehen %d Epochen", len(bau.buch.epochen))
	}
}

// schuetzen stellt die Rechte eines Ordners — und sagt es, wenn das nicht geht.
//
// **Eine eigene Funktion, weil `gosec` sonst jede Zahl anstreicht**: G302 will
// 0600 oder weniger, und hier geht es um einen **Ordner**, der für die Dauer
// eines Laufs nicht beschreibbar sein soll.
func schuetzen(t *testing.T, ordner string, rechte os.FileMode) {
	t.Helper()

	if fehler := os.Chmod(ordner, rechte); fehler != nil { //nolint:gosec // ein Ordner im TempDir des Laufs
		t.Fatalf("die Rechte von %s stellen: %v", ordner, fehler)
	}
}
