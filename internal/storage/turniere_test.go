package storage_test

// **Der Turnierspeicher** — Werkstück 3 von Phase 4.7
// (#1294),
// nach Vermerk [0062](../../docs/entscheidungen/0062-was-ein-turnier-hier-ist.md).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// turnierablage gibt eine frische Ablage samt Datenbank.
func turnierablage(t *testing.T) *storage.Turnierablage {
	t.Helper()

	return storage.NeueTurnierablage(frischeDatenbank(t))
}

// **Ein Turnier lässt sich anlegen und wiederfinden.**
func TestEinTurnierLaesstSichAnlegenUndWiederfinden(t *testing.T) {
	ablage := turnierablage(t)

	angelegt, err := ablage.Anlegen(probekontext(t), domain.Turnier{
		Name: "Küchentisch", Spiel: "magic", Notiz: "Freitagabend",
	})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if angelegt.ID == 0 || angelegt.Fassung != 1 {
		t.Errorf("%+v", angelegt)
	}

	gelesen, err := ablage.Turnier(probekontext(t), angelegt.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if gelesen.Name != "Küchentisch" || gelesen.Spiel != "magic" ||
		gelesen.Notiz != "Freitagabend" {
		t.Errorf("%+v", gelesen)
	}
}

// **Ein Turnier, das es nicht gibt, ist ein fachliches Nein.**
func TestEinFehlendesTurnierIstNichtGefunden(t *testing.T) {
	if _, err := turnierablage(t).Turnier(probekontext(t), 999); !errors.Is(
		err, domain.ErrNotFound,
	) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Zwei Turniere desselben Spiels dürfen nicht gleich heißen.**
func TestZweiTurniereDesselbenSpielsBrauchenVerschiedeneNamen(t *testing.T) {
	ablage := turnierablage(t)

	turnier := domain.Turnier{Name: "Küchentisch", Spiel: "magic"}

	if _, err := ablage.Anlegen(probekontext(t), turnier); err != nil {
		t.Fatalf("das erste: %v", err)
	}

	if _, err := ablage.Anlegen(probekontext(t), turnier); !errors.Is(
		err, domain.ErrConflict,
	) {
		t.Errorf("Fehler ist %v, erwartet ErrConflict", err)
	}
}

// **Ein Teilnehmer bekommt seine Stelle vom Speicher.**
//
// **Und nicht vom Aufrufer**: zwei Flächen, die gleichzeitig anmelden,
// vergäben sonst dieselbe — und der eindeutige Index wiese die zweite ab,
// obwohl beide gültig sind.
func TestEinTeilnehmerBekommtSeineStelleVomSpeicher(t *testing.T) {
	ablage := turnierablage(t)

	turnier, err := ablage.Anlegen(probekontext(t),
		domain.Turnier{Name: "Küchentisch", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	for nummer, name := range []string{"Ada", "Bert", "Cem"} {
		angemeldet, anmeldefehler := ablage.Anmelden(probekontext(t), turnier.ID,
			domain.Teilnehmer{Name: name})
		if anmeldefehler != nil {
			t.Fatalf("%s anmelden: %v", name, anmeldefehler)
		}

		if angemeldet.Stelle != nummer+1 {
			t.Errorf("%s bekam Stelle %d, erwartet %d",
				name, angemeldet.Stelle, nummer+1)
		}
	}

	gelesen, err := ablage.Turnier(probekontext(t), turnier.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen.Teilnehmer) != 3 || gelesen.Teilnehmer[0].Name != "Ada" {
		t.Errorf("%+v", gelesen.Teilnehmer)
	}
}

// **Eine Runde entsteht in einem Zug — mit allen ihren Paarungen.**
//
// Das ist die Zusage aus dem Issue: „Ein Absturz mitten in der
// Ergebniserfassung hinterlässt keine halbe Runde." Sie löst kein
// Zustandsfeld ein, sondern eine **Transaktionsgrenze** (Vermerk 0062,
// Entscheidung 10).
func TestEineRundeEntstehtInEinemZug(t *testing.T) {
	ablage := turnierablage(t)

	turnier := turnierMitTeilnehmern(t, ablage, "Ada", "Bert", "Cem")

	runde, err := ablage.RundeLegen(probekontext(t), turnier.ID, 1, []domain.Paarung{
		{A: turnier.Teilnehmer[0].ID, B: turnier.Teilnehmer[1].ID},
		{A: turnier.Teilnehmer[2].ID}, // Bye
	})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	if runde.Nummer != 1 || len(runde.Paarungen) != 2 {
		t.Errorf("%+v", runde)
	}

	if !runde.Paarungen[1].IstBye() {
		t.Errorf("die zweite Paarung ist kein Bye: %+v", runde.Paarungen[1])
	}
}

// **Eine Runde mit einer unmöglichen Paarung entsteht gar nicht.**
//
// **Die Gegenprobe zur Transaktion.** Ohne sie stünde die erste Paarung im
// Bestand und die zweite nicht — genau die halbe Runde.
func TestEineHalbeRundeEntstehtNicht(t *testing.T) {
	ablage := turnierablage(t)

	turnier := turnierMitTeilnehmern(t, ablage, "Ada", "Bert")

	_, err := ablage.RundeLegen(probekontext(t), turnier.ID, 1, []domain.Paarung{
		{A: turnier.Teilnehmer[0].ID, B: turnier.Teilnehmer[1].ID},
		// **Gegen sich selbst** — der Bestand weist es erst beim Schreiben
		// ab, nachdem die erste Paarung schon steht. Bis #1815 stand hier ein
		// Teilnehmer, den es nicht gibt; den weist seitdem die Frage nach dem
		// Besitz ab, bevor irgendetwas geschrieben ist, und die halbe Runde
		// entstünde auch ohne Transaktion nicht.
		{A: turnier.Teilnehmer[1].ID, B: turnier.Teilnehmer[1].ID},
	})
	if err == nil {
		t.Fatal("angenommen, erwartet abgewiesen")
	}

	gelesen, lesefehler := ablage.Turnier(probekontext(t), turnier.ID)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(gelesen.Runden) != 0 {
		t.Errorf("%d Runden im Bestand — erwartet keine", len(gelesen.Runden))
	}
}

// **Ein Ergebnis lässt sich eintragen und berichtigen.**
//
// Berichtigen ist der Normalfall am Turniertisch, nicht der Ausnahmefall —
// deshalb wird die Rangliste gerechnet und nicht abgelegt.
func TestEinErgebnisLaesstSichBerichtigen(t *testing.T) {
	ablage := turnierablage(t)

	turnier := turnierMitTeilnehmern(t, ablage, "Ada", "Bert")

	runde, err := ablage.RundeLegen(probekontext(t), turnier.ID, 1, []domain.Paarung{
		{A: turnier.Teilnehmer[0].ID, B: turnier.Teilnehmer[1].ID},
	})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	paarung := runde.Paarungen[0].ID

	if setzfehler := ablage.ErgebnisSetzen(probekontext(t),
		fassungVon(t, ablage, turnier.ID), paarung,
		&domain.Ergebnis{SpieleA: 2}); setzfehler != nil {
		t.Fatalf("eintragen: %v", setzfehler)
	}

	if setzfehler := ablage.ErgebnisSetzen(probekontext(t),
		fassungVon(t, ablage, turnier.ID), paarung,
		&domain.Ergebnis{SpieleA: 1, SpieleB: 2}); setzfehler != nil {
		t.Fatalf("berichtigen: %v", setzfehler)
	}

	gelesen, err := ablage.Turnier(probekontext(t), turnier.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	stand := gelesen.Runden[0].Paarungen[0].Ergebnis
	if stand == nil || stand.SpieleA != 1 || stand.SpieleB != 2 {
		t.Errorf("das Ergebnis ist %+v", stand)
	}
}

// **Und es lässt sich wieder herausnehmen.**
func TestEinErgebnisLaesstSichZuruecknehmen(t *testing.T) {
	ablage := turnierablage(t)

	turnier := turnierMitTeilnehmern(t, ablage, "Ada", "Bert")

	runde, err := ablage.RundeLegen(probekontext(t), turnier.ID, 1, []domain.Paarung{
		{A: turnier.Teilnehmer[0].ID, B: turnier.Teilnehmer[1].ID},
	})
	if err != nil {
		t.Fatalf("legen: %v", err)
	}

	if setzfehler := ablage.ErgebnisSetzen(probekontext(t),
		fassungVon(t, ablage, turnier.ID), runde.Paarungen[0].ID,
		&domain.Ergebnis{SpieleA: 2}); setzfehler != nil {
		t.Fatalf("eintragen: %v", setzfehler)
	}

	if ruecknahme := ablage.ErgebnisSetzen(probekontext(t),
		fassungVon(t, ablage, turnier.ID), runde.Paarungen[0].ID,
		nil); ruecknahme != nil {
		t.Fatalf("zurücknehmen: %v", ruecknahme)
	}

	gelesen, err := ablage.Turnier(probekontext(t), turnier.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if stand := gelesen.Runden[0].Paarungen[0].Ergebnis; stand != nil {
		t.Errorf("das Ergebnis ist %+v, erwartet keines", stand)
	}
}

// turnierMitTeilnehmern legt ein Turnier an und meldet Namen an.
func turnierMitTeilnehmern(
	t *testing.T, ablage *storage.Turnierablage, namen ...string,
) domain.Turnier {
	t.Helper()

	turnier, err := ablage.Anlegen(probekontext(t),
		domain.Turnier{Name: "Küchentisch", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	for _, name := range namen {
		angemeldet, anmeldefehler := ablage.Anmelden(probekontext(t), turnier.ID,
			domain.Teilnehmer{Name: name})
		if anmeldefehler != nil {
			t.Fatalf("%s anmelden: %v", name, anmeldefehler)
		}

		turnier.Teilnehmer = append(turnier.Teilnehmer, angemeldet)
	}

	return turnier
}

// **Ein Ergebnis für eine Paarung, die es nicht gibt, ist ein fachliches
// Nein** — kein Absturz und kein stilles Nichts.
func TestEinErgebnisOhnePaarungIstNichtGefunden(t *testing.T) {
	if err := turnierablage(t).ErgebnisSetzen(probekontext(t), 1, 999,
		&domain.Ergebnis{SpieleA: 2}); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Fehler ist %v, erwartet ErrNotFound", err)
	}
}

// **Eine Runde ohne Paarungen ist eine Runde ohne Paarungen** — kein Absturz.
//
// # Wie sie entsteht
//
// **Nicht über `RundeLegen`** — das weist eine leere Runde ab. Wohl aber in
// einem Bestand, den jemand von Hand berichtigt hat, und dann ist eine leere
// Runde die richtige Antwort. Der `LEFT JOIN` gibt sie als Zeile mit NULLen,
// und wer die nicht abfängt, liest eine Paarung mit der Kennung 0.
func TestEineRundeOhnePaarungenWirdGelesen(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueTurnierablage(db)

	turnier, err := ablage.Anlegen(probekontext(t),
		domain.Turnier{Name: "Küchentisch", Spiel: "magic"})
	if err != nil {
		t.Fatalf("anlegen: %v", err)
	}

	if _, einfuegen := db.ExecContext(probekontext(t),
		`INSERT INTO turnierrunden (turnier_id, nummer) VALUES (?, 1)`,
		turnier.ID); einfuegen != nil {
		t.Fatalf("die leere Runde einfügen: %v", einfuegen)
	}

	gelesen, err := ablage.Turnier(probekontext(t), turnier.ID)
	if err != nil {
		t.Fatalf("lesen: %v", err)
	}

	if len(gelesen.Runden) != 1 {
		t.Fatalf("%d Runden, erwartet eine", len(gelesen.Runden))
	}

	if len(gelesen.Runden[0].Paarungen) != 0 {
		t.Errorf("%d Paarungen, erwartet keine: %+v",
			len(gelesen.Runden[0].Paarungen), gelesen.Runden[0].Paarungen)
	}
}

// fassungVon liest die aktuelle Fassung eines Turniers.
//
// **Die Tests lesen sie frisch**, weil jeder Schreibweg sie erhöht — genau
// wie ein echter Aufrufer es tun muss (§7, Regel 2).
func fassungVon(t *testing.T, ablage *storage.Turnierablage, turnierID int64) int64 {
	t.Helper()

	gelesen, err := ablage.Turnier(probekontext(t), turnierID)
	if err != nil {
		t.Fatalf("Fassung lesen: %v", err)
	}

	return gelesen.Fassung
}
