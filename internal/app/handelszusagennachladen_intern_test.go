package app

// **Ein Nachladefehler ist keine gescheiterte Zusage** — §7, Regel 3
// (#1970).
//
// # Wie er hier entsteht
//
// **Ein Auslöser im echten Bestand** legt neben jede neue Zusage eine zweite
// mit einer Frist, die kein Zeitpunkt ist. Das Schreiben glückt — es liest
// die Zusagen **vor** dem Eintragen —, und das Nachlesen danach scheitert an
// ihr. Bis #1970 kam das als `fehler` zurück, und wer es wiederholte, hielt
// dieselbe Besitzzeile zweimal fest.

import (
	"testing"
)

func TestEinNachladefehlerIstKeineGescheiterteZusage(t *testing.T) {
	t.Parallel()

	anwendung, exemplarID := einHandelZumFesthalten(t)
	ctx := riegelkontext(t)

	if _, fehler := anwendung.db.ExecContext(ctx, `
CREATE TRIGGER krumme_zusage AFTER INSERT ON reservierungen
WHEN NEW.frist <> 'krumm'
BEGIN
  INSERT INTO reservierungen
    (profil_id, vorgang, revision, exemplar_id, menge, fassung, frist)
  VALUES (NEW.profil_id, NEW.vorgang, NEW.revision, NEW.exemplar_id, 1,
          NEW.fassung, 'krumm');
END`); fehler != nil {
		t.Fatalf("der Auslöser: %v", fehler)
	}

	antwort := anwendung.HandelFesthalten(ctx, "v-halt", exemplarID, 2, 1)

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("als gescheitert gemeldet: %+v", antwort.Fehler)
	case antwort.Nachladefehler == nil:
		t.Fatalf("der Nachladefehler fehlt: %+v", antwort)
	case antwort.Zusagen == nil || len(antwort.Zusagen) != 0:
		t.Fatalf("die Liste ist %+v", antwort.Zusagen)
	}

	// **Und festgehalten ist sie** — gezählt an der Zeile, nicht über die
	// Liste, die es nicht hergibt.
	var menge int
	if fehler := anwendung.db.QueryRowContext(ctx,
		`SELECT menge FROM reservierungen WHERE frist <> 'krumm'`,
	).Scan(&menge); fehler != nil || menge != 2 {
		t.Fatalf("die Zusage: %d (%v)", menge, fehler)
	}

	// **Ein gescheiterter Schritt bleibt einer**: mehr als dasteht, wird
	// abgewiesen, und das Nein steht in `fehler`.
	if abgewiesen := anwendung.HandelFesthalten(
		ctx, "v-halt", exemplarID, 9, 1); abgewiesen.Fehler == nil ||
		abgewiesen.Nachladefehler != nil {
		t.Fatalf("zu viel: %+v", abgewiesen)
	}
}
