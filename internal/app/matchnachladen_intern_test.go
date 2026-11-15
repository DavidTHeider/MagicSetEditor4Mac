package app

// **Ein Nachladefehler ist kein gescheiterter Matchschritt** — §7, Regel 3
// (#1970).
//
// # Wie er hier entsteht
//
// **Ein Auslöser im echten Bestand** verbiegt nach dem Schreiben die Sitze
// der Bindung, die das Nachlesen als JSON liest. Das Schreiben selbst liest
// sie **vorher** und glückt; nur das Nachlesen danach scheitert. Bis #1970
// kam das als `fehler` zurück — gebunden, gemeldet oder übernommen war
// trotzdem, und eine Wiederholung lief gegen den geschriebenen Stand.

import (
	"strconv"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// sitzeVerbiegen setzt einen Auslöser, der nach jedem Schreiben in `tabelle`
// die Sitze der Bindung zu etwas macht, das kein JSON ist — und gibt zurück,
// wie man sie wiederherstellt.
func sitzeVerbiegen(t *testing.T, tisch turniertisch, ereignis string) func() {
	t.Helper()

	ctx := riegelkontext(t)
	db := tisch.anwendung.db

	var heil string
	if fehler := db.QueryRowContext(ctx,
		`SELECT sitze FROM matchbindungen WHERE paarung_id = ?`, tisch.paarungID,
	).Scan(&heil); fehler != nil {
		t.Fatalf("die heilen Sitze: %v", fehler)
	}

	if _, fehler := db.ExecContext(ctx, `CREATE TRIGGER krumme_sitze `+ereignis+`
BEGIN UPDATE matchbindungen SET sitze = 'krumm'; END`); fehler != nil {
		t.Fatalf("der Auslöser: %v", fehler)
	}

	return func() {
		for _, anweisung := range []string{
			`DROP TRIGGER krumme_sitze`,
			`UPDATE matchbindungen SET sitze = '` + heil + `'`,
		} {
			if _, fehler := db.ExecContext(ctx, anweisung); fehler != nil {
				t.Fatalf("%s: %v", anweisung, fehler)
			}
		}
	}
}

// matchNachgeladen prüft eine Antwort, deren Schritt galt und deren Stand
// sich danach nicht lesen ließ.
func matchNachgeladen(t *testing.T, schritt string, antwort MatchAntwort) {
	t.Helper()

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("%s: als gescheitert gemeldet: %+v", schritt, antwort.Fehler)
	case antwort.Nachladefehler == nil:
		t.Fatalf("%s: der Nachladefehler fehlt: %+v", schritt, antwort)
	case antwort.Sitze == nil || antwort.Meldungen == nil:
		t.Fatalf("%s: die Listen sind nil: %+v", schritt, antwort)
	}
}

// zaehleMeldungen gibt, wie viele Meldungen an der Paarung stehen.
func zaehleMeldungen(t *testing.T, tisch turniertisch) int {
	t.Helper()

	var anzahl int
	if fehler := tisch.anwendung.db.QueryRowContext(riegelkontext(t),
		`SELECT count(*) FROM matchmeldungen WHERE paarung_id = ?`, tisch.paarungID,
	).Scan(&anzahl); fehler != nil {
		t.Fatalf("zählen: %v", fehler)
	}

	return anzahl
}

func TestEinNachladefehlerIstKeinGescheiterterMatchschritt(t *testing.T) {
	tisch := einTurniertisch(t)
	ctx := riegelkontext(t)
	binden := func() MatchAntwort {
		return tisch.anwendung.TurnierMatchBinden(ctx, tisch.turnierID,
			tisch.paarungID, tisch.partieID, tisch.zuordnung)
	}

	if gebunden := binden(); gebunden.Fehler != nil {
		t.Fatalf("binden: %+v", gebunden.Fehler)
	}

	// **Binden** — die neue Bindung steht da, nur nicht lesbar.
	heilen := sitzeVerbiegen(t, tisch, `AFTER INSERT ON matchbindungen`)
	matchNachgeladen(t, "binden", binden())
	heilen()

	// **Melden** — beide Seiten, und beide Meldungen stehen da.
	stand := ErgebnisDTO{SpieleA: 2, SpieleB: 1}

	for anzahl, von := range []string{"anna", "bert"} {
		heilen = sitzeVerbiegen(t, tisch, `AFTER INSERT ON matchmeldungen`)
		matchNachgeladen(t, "melden "+von, tisch.anwendung.TurnierMatchMelden(
			ctx, tisch.paarungID, von, "letztesglied", stand))
		heilen()

		if gezaehlt := zaehleMeldungen(t, tisch); gezaehlt != anzahl+1 {
			t.Fatalf("nach dem Melden von %s stehen %d da", von, gezaehlt)
		}
	}

	// **Übernehmen** — das Turnier trägt das Ergebnis, und die Antwort sagt
	// trotzdem, welches.
	heilen = sitzeVerbiegen(t, tisch, `AFTER UPDATE ON turnierpaarungen`)
	uebernommen := tisch.anwendung.TurnierMatchUebernehmen(
		ctx, tisch.turnierID, tisch.paarungID)
	matchNachgeladen(t, "übernehmen", uebernommen)
	heilen()

	if uebernommen.Uebernommen == nil || uebernommen.Uebernommen.SpieleA != 2 {
		t.Fatalf("übernommen: %+v", uebernommen.Uebernommen)
	}

	turnier := tisch.anwendung.Turnier(ctx, tisch.turnierID)
	if ergebnis := turnier.Turnier.Runden[0].Paarungen[0].Ergebnis; turnier.Fehler != nil ||
		ergebnis == nil || ergebnis.SpieleA != 2 {
		t.Fatalf("das Turnier trägt %+v (%+v)", ergebnis, turnier.Fehler)
	}

	// **Ein gescheiterter Schritt bleibt einer** — eine Paarung, die es
	// nicht gibt, hat keinen Nachladefehler.
	if antwort := tisch.anwendung.TurnierMatchMelden(
		ctx, "999", "anna", "glied", stand); antwort.Fehler == nil ||
		antwort.Nachladefehler != nil {
		t.Fatalf("die fremde Paarung: %+v", antwort)
	}
}

// **Auch die Meldung aus der Partie** — der vierte Weg.
func TestEinNachladefehlerNachDerMeldungAusDerPartie(t *testing.T) {
	tisch := einTurniertisch(t)
	ctx := riegelkontext(t)

	if gebunden := tisch.anwendung.TurnierMatchBinden(ctx, tisch.turnierID,
		tisch.paarungID, tisch.partieID, tisch.zuordnung); gebunden.Fehler != nil {
		t.Fatalf("binden: %+v", gebunden.Fehler)
	}

	stand := tisch.anwendung.Partie(ctx, tisch.partieID, "anna")
	fassung, fehler := strconv.ParseInt(stand.Partie.Fassung, 10, 64)

	if stand.Fehler != nil || fehler != nil {
		t.Fatalf("Partie: %+v (%v)", stand.Fehler, fehler)
	}

	if aufgegeben := tisch.anwendung.PartieBefehl(ctx, domain.Spielbefehl{
		PartieID: tisch.partieID, Kennung: "b1", Absender: "bert",
		Art: "zug.aufgeben", Erwartet: fassung,
	}); aufgegeben.Fehler != nil {
		t.Fatalf("aufgeben: %+v", aufgegeben.Fehler)
	}

	heilen := sitzeVerbiegen(t, tisch, `AFTER INSERT ON matchmeldungen`)
	matchNachgeladen(t, "aus der Partie", tisch.anwendung.TurnierMatchAusDerPartie(
		ctx, tisch.paarungID, "anna"))
	heilen()

	if gezaehlt := zaehleMeldungen(t, tisch); gezaehlt != 1 {
		t.Fatalf("es stehen %d Meldungen da", gezaehlt)
	}
}
