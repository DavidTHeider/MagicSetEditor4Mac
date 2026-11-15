package flaechen_test

// **Die Sicht eines Gegenübers über einen echten Draht** — P11
// (#2379).
//
// Zwei `mse`-Prozesse mit getrennten Beständen, TCP mit TLS 1.3 wie in
// `handelsdraht_test.go`: Bert gibt Anna seine Sammlung zum Handeln frei und
// horcht, Anna fragt mit `mse sicht-anfragen` — und bekommt genau Berts
// freigegebene Zeilen. Nach dem Widerruf eine Absage, ohne Zeile.

import (
	"strings"
	"testing"
)

func TestEineSichtKommtUeberEinenEchtenDraht(t *testing.T) {
	anna := stelleAuf(t)
	bert := stelleAuf(t)

	kenneUndErlaube(t, anna, bert)
	kenneUndErlaube(t, bert, anna)

	mussKommandozeile(t, bert.umgebung, "exemplar",
		"00000000-0000-4000-8000-0000000b1002", "2", "nearMint", "de", "nonfoil")
	mussKommandozeile(t, bert.umgebung, "kontakt-freigeben", anna.kennung, "handel", "sammlung")

	adresse, protokoll := horcheImHintergrund(t, bert)
	mussKommandozeile(t, anna.umgebung, "kontakt-adresse", bert.kennung, adresse)

	sicht := mussKommandozeile(t, anna.umgebung, "sicht-anfragen", bert.kennung)
	if zeilen := strings.Split(strings.TrimSpace(sicht), "\n"); len(zeilen) != 1 ||
		!strings.Contains(zeilen[0], "\t2\tnearMint\tde\tnonfoil\t") ||
		!strings.HasSuffix(zeilen[0], "\tangebot") {
		t.Fatalf("die Sicht: %q\nBerts Protokoll:\n%s", sicht, protokoll.text())
	}

	// **Ein Wunsch heißt in der Ausgabe so** (Befund 7 des Zweitblicks über
	// Phase 8): ohne die Spalte sah er aus wie ein Angebot desselben Drucks.
	mussKommandozeile(t, bert.umgebung, "exemplar",
		"00000000-0000-4000-8000-0000000b1003", "1", "nearMint", "de", "nonfoil",
		"", "", "", "", "", "", "", "", "wunsch")
	mussKommandozeile(t, bert.umgebung, "kontakt-freigeben", anna.kennung, "handel", "wuensche")

	wuensche := mussKommandozeile(t, anna.umgebung, "sicht-anfragen", bert.kennung)
	if !strings.Contains(wuensche, "\twunsch\n") {
		t.Errorf("der Wunsch: %q", wuensche)
	}

	// **Nach dem Widerruf eine Absage** — und keine Zeile. Die Freigaben
	// stehen je Umfang, also beide.
	mussKommandozeile(t, bert.umgebung, "kontakt-widerrufen", anna.kennung, "handel", "sammlung")
	mussKommandozeile(t, bert.umgebung, "kontakt-widerrufen", anna.kennung, "handel", "wuensche")

	aus, fehler, code := kommandozeile(t, anna.umgebung, "sicht-anfragen", bert.kennung)
	if code == 0 || strings.TrimSpace(aus) != "" || !strings.Contains(fehler, "nichts frei") {
		t.Errorf("nach dem Widerruf: Code %d, %q, %q", code, aus, fehler)
	}
}
