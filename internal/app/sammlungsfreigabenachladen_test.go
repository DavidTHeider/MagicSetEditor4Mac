package app_test

// **Ein Nachladefehler ist keine gescheiterte Sichtfreigabe** — §7, Regel 3
// (#1970).
//
// # Wie er hier entsteht
//
// **Ein Auslöser im echten Bestand** legt neben die erteilte oder
// widerrufene Freigabe eine zweite mit einem Ablauf, der kein Zeitpunkt ist.
// Geschrieben ist dann; die Liste danach scheitert an ihr. Bis #1970 kam das
// als `fehler` zurück — wer daraufhin widerrief, „widerrief ins Leere".

import (
	"database/sql"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
)

// freigabeVerbiegen setzt einen Auslöser, der nach `ereignis` eine krumme
// Freigabe an dasselbe Gegenüber legt — und gibt zurück, wie man aufräumt.
func freigabeVerbiegen(t *testing.T, db *sql.DB, ereignis, zeile string) func() {
	t.Helper()

	ctx := probekontext(t)

	if _, fehler := db.ExecContext(ctx, `CREATE TRIGGER krumme_freigabe `+ereignis+`
WHEN `+zeile+`.ablauf <> 'krumm'
BEGIN
  INSERT INTO sammlungsfreigaben (profil_id, empfaenger, umfang, zweck, ablauf, erteilt)
  VALUES (`+zeile+`.profil_id, `+zeile+`.empfaenger, `+zeile+`.umfang, 'teilen',
          'krumm', `+zeile+`.erteilt);
END`); fehler != nil {
		t.Fatalf("der Auslöser: %v", fehler)
	}

	return func() {
		for _, anweisung := range []string{
			`DROP TRIGGER krumme_freigabe`,
			`DELETE FROM sammlungsfreigaben WHERE ablauf = 'krumm'`,
		} {
			if _, fehler := db.ExecContext(ctx, anweisung); fehler != nil {
				t.Fatalf("%s: %v", anweisung, fehler)
			}
		}
	}
}

// sichtfreigabeNachgeladen prüft eine Antwort, deren Schritt galt und deren
// Liste sich danach nicht lesen ließ.
func sichtfreigabeNachgeladen(t *testing.T, schritt string, antwort app.SammlungsfreigabenAntwort) {
	t.Helper()

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("%s: als gescheitert gemeldet: %+v", schritt, antwort.Fehler)
	case antwort.Nachladefehler == nil:
		t.Fatalf("%s: der Nachladefehler fehlt: %+v", schritt, antwort)
	case antwort.Freigaben == nil || len(antwort.Umfaenge) == 0:
		t.Fatalf("%s: die Antwort trägt %+v", schritt, antwort)
	}
}

func TestEinNachladefehlerIstKeineGescheiterteSichtfreigabe(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	db := frischerBestand(t)
	flaeche := app.PeerflaecheUeber(db)
	ctx := probekontext(t)
	kennung := einGegenueber(t, flaeche)

	zaehle := func() (anzahl int) {
		if fehler := db.QueryRowContext(ctx,
			`SELECT count(*) FROM sammlungsfreigaben WHERE zweck = 'handel'`,
		).Scan(&anzahl); fehler != nil {
			t.Fatalf("zählen: %v", fehler)
		}

		return anzahl
	}

	aufraeumen := freigabeVerbiegen(t, db, `AFTER INSERT ON sammlungsfreigaben`, "NEW")
	sichtfreigabeNachgeladen(t, "freigeben", flaeche.SammlungFreigeben(ctx,
		app.SammlungsfreigabeDTO{Empfaenger: kennung, Umfang: "sammlung", Zweck: "handel"}))
	aufraeumen()

	if anzahl := zaehle(); anzahl != 1 {
		t.Fatalf("nach dem Freigeben stehen %d da", anzahl)
	}

	aufraeumen = freigabeVerbiegen(t, db, `AFTER DELETE ON sammlungsfreigaben`, "OLD")
	sichtfreigabeNachgeladen(t, "widerrufen",
		flaeche.SammlungsfreigabeWiderrufen(ctx, kennung, "handel", "sammlung"))
	aufraeumen()

	if anzahl := zaehle(); anzahl != 0 {
		t.Fatalf("nach dem Widerruf stehen %d da", anzahl)
	}

	// **Ein gescheiterter Schritt bleibt einer** — ein Widerruf ins Leere
	// meldet sich in `fehler`.
	if antwort := flaeche.SammlungsfreigabeWiderrufen(
		ctx, kennung, "handel", "sammlung"); antwort.Fehler == nil || antwort.Nachladefehler != nil {
		t.Fatalf("der Widerruf ins Leere: %+v", antwort)
	}
}
