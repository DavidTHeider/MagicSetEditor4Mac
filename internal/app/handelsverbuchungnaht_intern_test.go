package app

// **Der geglückte Abschluss an der Naht** — Werkstück `T6` von Phase 5.6
// (#1471).
//
// # Warum dieser Lauf paketintern steht
//
// **Weil er einen Schlüsselbund braucht, den es im Test nicht gibt.** Eine
// Unterschrift entsteht aus der Peeridentität, und die liegt im Anmeldebund;
// Testläufe bekommen ihn nicht (§3). `MitPeerbund` setzt einen eigenen ein,
// und der ist paketintern — dieselbe Vorrichtung wie in
// `peerausweis_intern_test.go`.
//
// # Und warum der Handel von Hand hingelegt wird
//
// **Weil hier die Naht gemessen wird und nicht der Weg dorthin.** Dass ein
// Handel über die Flächen bis zur bestätigten Übergabe kommt, fährt
// `internal/flaechen/handelsleitung_test.go`
// (#1637);
// hier steht der Abschluss schon da, damit dieser Lauf zeigen kann, dass die
// Naht ihn fährt, den Beleg **liest** statt ihn zu folgern und die
// Unterschrift heil weiterreicht.
//
// **Bis zum 2026-09-19 stand hier ein anderer Grund** — die Schritte davor
// hätten keine Fläche. Den gibt es nicht mehr, und ein Grund, der nicht
// stimmt, ist schlimmer als keiner (§10).

import (
	"database/sql"
	"encoding/hex"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// derNahtvorgang ist der Handel dieses Laufs.
const derNahtvorgang = "v-naht"

// legeHandelHin schreibt Angebot, Besitz und eine bestätigte Zusage.
func legeHandelHin(t *testing.T, db *sql.DB, ich peer.Kennung) {
	t.Helper()

	ctx := riegelkontext(t)

	bestand := storage.NeuerSammlungsspeicher(db)

	exemplar := &domain.Exemplar{
		Printing: "bolt-m10", Menge: 4, Zustand: domain.ZustandNearMint,
		Sprache: "de", Lagerort: "Kiste 4",
	}

	kennung, fehler := bestand.Anlegen(ctx, exemplar)
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := storage.NeueReservierungsablage(db).Festhalten(ctx,
		domain.Reservierung{
			Vorgang: derNahtvorgang, Revision: 1, ExemplarID: kennung,
			Menge: 2, Fassung: 1, Frist: time.Now().Add(time.Hour),
			Uebergeben: true,
		}, domain.Zusagevertrag{{
			Printing: "bolt-m10", Menge: 99, Guete: "nearMint", Sprache: "de",
		}}); fehler != nil {
		t.Fatalf("festhalten: %v", fehler)
	}

	if fehler := storage.NeueHandelsablage(db).Schreiben(ctx,
		domain.Handelsangebot{
			Vertragsfassung: 1, Vorgang: derNahtvorgang, Revision: 1,
			Art: domain.HandelTausch, Initiator: string(ich),
			Parteien: []string{string(ich), "bgegenueber"},
			Zustand:  domain.HandelUebergabeOffen,
			Positionen: []domain.Handelsposition{
				{
					Von: string(ich), Printing: "bolt-m10", Menge: 2,
					Guete: string(domain.ZustandNearMint), Sprache: "de", Finish: "nonfoil",
				},
				{
					Von: "bgegenueber", Printing: "shivan-lea", Menge: 1,
					Guete: string(domain.ZustandNearMint), Sprache: "en", Finish: "nonfoil",
				},
			},
		}); fehler != nil {
		t.Fatalf("den Handel schreiben: %v", fehler)
	}
}

func TestDieNahtVerbuchtUndReichtDenBelegHeilDurch(t *testing.T) {
	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeeridentitaetSicherstellen(ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	ich, fehler := anwendung.Ausweis(ctx)
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	legeHandelHin(t, anwendung.db, ich.Kennung())

	// **Die Freigabe kommt über die Fläche und nicht aus der Luft** (T10,
	// #1475): ein Nachweis für diesen Vorgang in diesem Stand, ausgestellt
	// dort, wo ein Mensch zustimmt.
	freigabe := anwendung.HandelFreigabeErteilen(ctx, derNahtvorgang, "verbuchen", "")
	if freigabe.Fehler != nil {
		t.Fatalf("Freigabe: %+v", freigabe.Fehler)
	}

	antwort := anwendung.Handelsverbuchung(ctx, derNahtvorgang, freigabe.Kennung)
	if antwort.Fehler != nil {
		t.Fatalf("Handelsverbuchung: %+v", antwort.Fehler)
	}

	if antwort.Zustand != string(domain.HandelLokalVerbucht) {
		t.Errorf("der Vorgang steht auf %q", antwort.Zustand)
	}

	// **Der Beleg kommt mit** — gelesen und nicht gefolgert.
	if !antwort.HatBeleg {
		t.Fatal("es kam kein Beleg")
	}

	if antwort.Beleg.Vorgang != derNahtvorgang || antwort.Beleg.Abdruck == "" {
		t.Errorf("der Beleg ist %+v", antwort.Beleg)
	}

	if antwort.Beleg.Beidseitig {
		t.Error("der Beleg gibt sich als beidseitig aus")
	}

	// **Und die Unterschrift hält** — sie ist Hex und gehört zu dieser
	// Identität.
	roh, hexfehler := hex.DecodeString(antwort.Beleg.Unterschrift)
	if hexfehler != nil {
		t.Fatalf("die Unterschrift ist kein Hex: %q", antwort.Beleg.Unterschrift)
	}

	if len(roh) == 0 {
		t.Fatal("die Unterschrift ist leer")
	}

	if _, fehler := time.Parse(time.RFC3339, antwort.Beleg.Gebucht); fehler != nil {
		t.Errorf("die Uhr lautet %q: %v", antwort.Beleg.Gebucht, fehler)
	}

	// **Ein zweiter Ruf bewegt nichts** — der Beleg steht schon da, und er
	// verbraucht deshalb auch keine zweite Freigabe: die erste ist
	// eingelöst, und ein Mensch, der hier noch einmal zustimmen müsste,
	// stimmte einer Sache zu, die schon geschehen ist.
	nochmal := anwendung.Handelsverbuchung(ctx, derNahtvorgang, "")
	if nochmal.Fehler != nil {
		t.Fatalf("zweiter Ruf: %+v", nochmal.Fehler)
	}

	if nochmal.Beleg.Unterschrift != antwort.Beleg.Unterschrift {
		t.Error("der zweite Ruf gab eine andere Unterschrift")
	}

	// **Und das Journal führt beide Bewegungen.**
	journal := anwendung.Handelsjournal(ctx, derNahtvorgang)
	if journal.Fehler != nil {
		t.Fatalf("Handelsjournal: %+v", journal.Fehler)
	}

	if len(journal.Zeilen) != 2 {
		t.Fatalf("das Journal führt %d Zeilen", len(journal.Zeilen))
	}
}

// **Ein Lesefehler beim Beleg nimmt den Schritt nicht zurück.**
//
// # Warum dieser Zweig eigens gefahren wird
//
// **Weil er die Regel 3 aus §7 einlöst**: geschrieben ist geschrieben. Der
// Schritt ist getan und abgelegt; dass der Beleg danach nicht zu lesen ist,
// wird **gemeldet** — und der Zustand kommt trotzdem mit, sonst hielte der
// Aufrufer einen getanen Schritt für misslungen und ginge ihn noch einmal.
//
// **Mit einer geschlossenen Datenbank**, weil ein echter Treiber diesen Fall
// nicht hergibt: dieselbe Verbindung, die eben geschrieben hat, liest auch.
func TestEinLesefehlerAmBelegNimmtDenSchrittNichtZurueck(t *testing.T) {
	db := bestandFuerDenRiegel(t)
	anwendung := PeerflaecheUeber(db).MitPeerbund(&peerprobebund{})

	if fehler := db.Close(); fehler != nil {
		t.Fatalf("schließen: %v", fehler)
	}

	antwort := anwendung.nachVerbuchung(
		anwendung.imProfil(riegelkontext(t)), derNahtvorgang,
		domain.Handelsangebot{
			Vorgang: derNahtvorgang, Zustand: domain.HandelLokalVerbucht,
		}, nil)

	if antwort.Fehler == nil {
		t.Fatal("der Lesefehler ging verloren")
	}

	// **Und der Zustand steht trotzdem da.**
	if antwort.Zustand != string(domain.HandelLokalVerbucht) {
		t.Errorf("der Zustand ging verloren: %q", antwort.Zustand)
	}

	if antwort.HatBeleg {
		t.Error("es kam ein Beleg")
	}
}

// **Ein geprüfter Gegenbeleg kommt mit Kennung und Uhrzeit ans Fenster** —
// #1650.
//
// **Und ohne einen bleiben beide Felder leer** — nicht `0001-01-01`: eine
// Nullzeit sieht aus wie eine Angabe.
func TestEinGegenbelegKommtAnsFenster(t *testing.T) {
	t.Parallel()

	anwendung := PeerflaecheUeber(
		bestandFuerDenRiegel(t)).MitPeerbund(&peerprobebund{})

	ctx := anwendung.imProfil(riegelkontext(t))

	if antwort := anwendung.PeeridentitaetSicherstellen(
		ctx); antwort.Fehler != nil {
		t.Fatalf("Identität anlegen: %+v", antwort.Fehler)
	}

	ich, fehler := anwendung.Ausweis(ctx)
	if fehler != nil {
		t.Fatalf("Ausweis: %v", fehler)
	}

	legeHandelHin(t, anwendung.db, ich.Kennung())

	freigabe := anwendung.HandelFreigabeErteilen(
		ctx, derNahtvorgang, "verbuchen", "")
	if freigabe.Fehler != nil {
		t.Fatalf("Freigabe: %+v", freigabe.Fehler)
	}

	if antwort := anwendung.Handelsverbuchung(
		ctx, derNahtvorgang, freigabe.Kennung); antwort.Fehler != nil {
		t.Fatalf("Handelsverbuchung: %+v", antwort.Fehler)
	}

	ablage := storage.NeueAbschlussablage(anwendung.db)

	// **Ohne Gegenbeleg steht nichts da** — und zwar **leer** und nicht
	// `0001-01-01`: eine Nullzeit sieht aus wie eine Angabe.
	ohne, fehler := ablage.Beleg(ctx, derNahtvorgang)
	if fehler != nil || ohne == nil {
		t.Fatalf("Beleg: %v / %v", ohne, fehler)
	}

	if dto := alsHandelsbelegDTO(*ohne); dto.Gegenzeichner != "" ||
		dto.Gegengezeichnet != "" {
		t.Errorf("ohne Gegenbeleg steht %+v", dto)
	}

	// **Und mit einem steht, von wem er kam und wann.**
	if abgewiesen := ablage.Gegenzeichnen(ctx, domain.Fremdbeleg{
		Vorgang: derNahtvorgang, Von: "peer-bert", Abdruck: "a1",
		Unterschrift: "deadbeef",
	}); abgewiesen != nil {
		t.Fatalf("Gegenzeichnen: %v", abgewiesen)
	}

	mit, fehler := ablage.Beleg(ctx, derNahtvorgang)
	if fehler != nil || mit == nil {
		t.Fatalf("Beleg: %v / %v", mit, fehler)
	}

	dto := alsHandelsbelegDTO(*mit)

	if !dto.Beidseitig {
		t.Error("der Beleg ist nicht beidseitig")
	}

	if dto.Gegenzeichner != "peer-bert" {
		t.Errorf("gegengezeichnet hat %q", dto.Gegenzeichner)
	}

	if dto.Gegengezeichnet == "" {
		t.Error("die Uhrzeit des Gegenbelegs fehlt")
	}
}
