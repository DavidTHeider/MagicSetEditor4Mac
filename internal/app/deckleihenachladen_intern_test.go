package app

// **Ein Nachladefehler ist kein gescheiterter Leihschritt** — §7, Regel 3,
// dieselbe Klasse wie bei den Kontakten
// (#1960).
//
// # Was hier gemessen wird
//
// **Jeder schreibende Leihweg, an einem echten Bestand.** Eine Zeile, deren
// Dokument kein JSON ist, lässt sich schreiben und nicht lesen: jede Änderung
// an einer **anderen** Leihe glückt, und das Lesen der Liste danach scheitert
// an ihr. Bis dahin kam das als `fehler` zurück — angeboten, angenommen,
// widerrufen oder zurückgegeben war trotzdem.

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// krummeLeihzeile legt eine Leihe an, deren Dokument kein JSON ist —
// **schreiben lässt sie sich, lesen nicht**.
func krummeLeihzeile(t *testing.T, bestand *sql.DB) {
	t.Helper()

	if _, fehler := bestand.ExecContext(t.Context(), `
INSERT INTO deckleihen (kennung, profil_id, dokument, abzug, verliehen)
VALUES ('krumm', 1, 'kein JSON', '{}', 0)`); fehler != nil {
		t.Fatalf("die krumme Zeile: %v", fehler)
	}
}

// geliehen prüft eine Antwort, die nach geglücktem Schreiben nicht nachladen
// konnte.
func geliehen(t *testing.T, schritt string, antwort DeckleiheAntwort) {
	t.Helper()

	switch {
	case antwort.Fehler != nil:
		t.Fatalf("%s: als gescheitert gemeldet: %+v", schritt, antwort.Fehler)
	case antwort.Nachladefehler == nil:
		t.Fatalf("%s: der Nachladefehler fehlt: %+v", schritt, antwort)
	case antwort.Leihen == nil || len(antwort.Leihen) != 0:
		t.Fatalf("%s: die Liste ist %+v", schritt, antwort.Leihen)
	}
}

// eineGelieheneKennung gibt die Kennung der einen heilen Leihe — gelesen an
// der Zeile, weil die Liste es nicht mehr hergibt.
func eineGelieheneKennung(t *testing.T, bestand *sql.DB) string {
	t.Helper()

	var kennung string
	if fehler := bestand.QueryRowContext(t.Context(),
		`SELECT kennung FROM deckleihen WHERE kennung <> 'krumm'`,
	).Scan(&kennung); fehler != nil {
		t.Fatalf("die Kennung der Leihe: %v", fehler)
	}

	return kennung
}

func TestEinNachladefehlerIstKeinGescheiterterLeihschritt(t *testing.T) {
	t.Parallel()

	verleiher, entleiher := eineLeihanwendung(t), eineLeihanwendung(t)
	ctx := riegelkontext(t)

	deckID := einDeckImBestand(t, verleiher.db)

	bert := entleiher.Peeridentitaet(ctx)
	if bert.Fehler != nil {
		t.Fatalf("die Kennung des Entleihers: %+v", bert.Fehler)
	}

	krummeLeihzeile(t, verleiher.db)
	krummeLeihzeile(t, entleiher.db)

	ziel := filepath.Join(t.TempDir(), "leihe.msepackage")

	angeboten := verleiher.DeckleiheAnbieten(ctx, LeihwunschDTO{
		Entleiher: bert.Kennung, DeckID: deckID,
		Bindungsart: "partie", Woran: "p1",
		Rechte: []string{"spielen"}, Stunden: 24, Modus: "zuende",
	}, ziel)
	geliehen(t, "anbieten", angeboten)

	// **Der Pfad geht trotzdem mit** — der Umschlag liegt, und wer ihn
	// weitergeben will, muss wissen, wo.
	if _, fehler := os.Stat(ziel); fehler != nil || angeboten.Pfad != ziel {
		t.Fatalf("anbieten: der Umschlag %q (%v)", angeboten.Pfad, fehler)
	}

	geliehen(t, "annehmen", entleiher.DeckleiheAnnehmen(ctx, ziel))

	kennung := eineGelieheneKennung(t, entleiher.db)

	geliehen(t, "zurückgeben", entleiher.DeckleiheZurueckgeben(ctx, kennung))
	geliehen(t, "widerrufen", verleiher.DeckleiheWiderrufen(ctx, kennung))

	// **Und geschrieben ist es wirklich** — nachgelesen an der einen Zeile,
	// die sich lesen lässt, auf beiden Seiten.
	for name, bestand := range map[string]*sql.DB{
		"beim Verleiher": verleiher.db, "beim Entleiher": entleiher.db,
	} {
		stand, fehler := storage.NeueLeihablage(bestand).Leihe(ctx, kennung)
		if fehler != nil || !stand.Widerrufen {
			t.Errorf("%s: nicht geschrieben: %+v (%v)", name, stand, fehler)
		}
	}

	// **Ein gescheiterter Schritt bleibt einer** — sein Fehler steht in
	// `fehler` und nicht daneben.
	antwort := verleiher.DeckleiheWiderrufen(ctx, "gibt es nicht")
	if antwort.Fehler == nil || antwort.Nachladefehler != nil {
		t.Fatalf("die unbekannte Leihe: %+v", antwort)
	}
}
