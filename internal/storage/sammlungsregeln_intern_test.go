package storage

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// **Die Sammlung nimmt nur gültige Zeilen — auch an der Domäne vorbei** —
// ENG-05.AK-01 (#1730).
//
// # Warum das eine kritische Regel ist
//
// **An ihr hängen Besitz und Handel.** Eine Zeile mit Menge null gibt einen
// Besitz vor, den es nicht gibt; ein Zustand oder eine Ausführung außerhalb
// der Aufzählung geht an der Eindeutigkeit vorbei, die Dubletten erkennt
// (F08) — „Foil" und „foil" wären zwei Zeilen.
//
// **Zwei Orte, zwei Läufe**: die Domäne hält es in `Exemplar.Validate`
// (`internal/domain/exemplar_test.go`), der Bestand in den Auslösern aus
// Wanderung 0080 — geprüft hier, **ohne** die Domäne zu fragen.
//
// # Und die Aufzählung ist dieselbe
//
// Jeder Zustand und jede Ausführung, die die Domäne kennt, geht hinein. Ein
// neuer Zustand ohne neue Wanderung wird hier rot und nicht erst im Fenster.
func TestDieSammlungNimmtNurGueltigeZeilen(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)
	ctx := probekontext(t)

	nummer := 0
	einfuegen := func(menge int, zustand, finish, sprache string) error {
		nummer++

		_, fehler := db.ExecContext(ctx,
			`INSERT INTO sammlung (printing, menge, zustand, sprache, finish, profil_id)
			 VALUES (?, ?, ?, ?, ?, 1)`,
			"druck-"+string(rune('a'+nummer%26))+string(rune('a'+nummer/26)),
			menge, zustand, sprache, finish)

		return einordnen(fehler)
	}

	for _, zustand := range domain.Zustaende() {
		for _, finish := range []domain.Finish{domain.FinishNonfoil, domain.FinishFoil, domain.FinishEtched} {
			if !domain.IstFinish(finish) {
				t.Fatalf("%s ist keine Ausführung der Domäne mehr", finish)
			}

			if fehler := einfuegen(1, string(zustand), string(finish), "en"); fehler != nil {
				t.Errorf("%s/%s: %v", zustand, finish, fehler)
			}
		}
	}

	for name, fehler := range map[string]error{
		"Menge null":            einfuegen(0, "nearMint", "nonfoil", "en"),
		"Menge negativ":         einfuegen(-2, "nearMint", "nonfoil", "en"),
		"unbekannter Zustand":   einfuegen(1, "NM", "nonfoil", "en"),
		"unbekannte Ausführung": einfuegen(1, "nearMint", "Foil", "en"),
		"ohne Sprache":          einfuegen(1, "nearMint", "nonfoil", "  "),
	} {
		if !errors.Is(fehler, domain.ErrValidation) {
			t.Errorf("%s: %v, erwartet eine Validierung", name, fehler)
		}
	}
}

// **Ein Nachweis mit Betragsgrenze trägt eine gültige** — dieselbe Regel wie
// `Nachweis.Validate`: nicht negativ, und mit Betrag eine Währung.
func TestEinNachweisTraegtEineGueltigeGrenze(t *testing.T) {
	t.Parallel()

	db := migrierteDatenbank(t)
	ctx := probekontext(t)

	nachweis := func(kennung string, betrag int, waehrung string) error {
		_, fehler := db.ExecContext(ctx,
			`INSERT INTO freigabenachweise
			   (kennung, nutzer_id, profil_id, zweck, ziel, revision, grenze_betrag,
			    grenze_waehrung, ausgestellt, ablauf)
			 VALUES (?, 1, 1, 'handel', 'x', 1, ?, ?,
			         '2026-09-23T00:00:00.000Z', '2026-09-24T00:00:00.000Z')`,
			kennung, betrag, waehrung)

		return einordnen(fehler)
	}

	for name, fall := range map[string]struct {
		fehler   error
		erwartet error
	}{
		"ohne Grenze":            {nachweis("a", 0, ""), nil},
		"mit Grenze und Währung": {nachweis("b", 500, "EUR"), nil},
		"negative Grenze":        {nachweis("c", -1, "EUR"), domain.ErrValidation},
		"Grenze ohne Währung":    {nachweis("d", 500, " "), domain.ErrValidation},
	} {
		if (fall.erwartet == nil) != (fall.fehler == nil) ||
			(fall.erwartet != nil && !errors.Is(fall.fehler, fall.erwartet)) {
			t.Errorf("%s: %v, erwartet %v", name, fall.fehler, fall.erwartet)
		}
	}
}

// **Eine alte Zeile bleibt bearbeitbar** — der Auslöser prüft, was sich
// ändert, nicht was dasteht.
//
// Ein Bestand aus 0.7.0 kann eine Zeile tragen, die vor 0080 entstand und
// einen Zustand hat, den es heute nicht mehr gibt. Ein Auslöser, der jede
// Änderung der Zeile abwiese, ließe ihre Notiz und ihre Menge nicht mehr
// setzen.
//
// **Seit es nur die Grundlinie gibt** (Vermerk 0189) entsteht eine solche
// Zeile nicht mehr über die Wanderung; sie wird an dem Auslöser vorbei
// geschrieben, der sie heute abwiese, und der Auslöser danach wieder
// angelegt — **aus seinem eigenen Text** in `sqlite_master`, nicht aus einer
// Abschrift.
func TestEineAlteZeileBleibtBearbeitbar(t *testing.T) {
	t.Parallel()

	db := offeneDatenbank(t)
	ctx := probekontext(t)

	if fehler := Migriere(ctx, db, nil); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	var ausloeser string
	if fehler := db.QueryRowContext(ctx,
		`SELECT sql FROM sqlite_master WHERE type = 'trigger' AND name = 'sammlung_regeln_neu'`).Scan(&ausloeser); fehler != nil {
		t.Fatalf("den Auslöser lesen: %v", fehler)
	}

	for _, anweisung := range []string{
		`DROP TRIGGER sammlung_regeln_neu`,
		`INSERT INTO sammlung (id, printing, menge, zustand, sprache, finish, profil_id)
		 VALUES (1, 'alt', 2, 'NM', 'en', 'nonfoil', 1)`,
		ausloeser,
	} {
		if _, fehler := db.ExecContext(ctx, anweisung); fehler != nil {
			t.Fatalf("die alte Zeile: %v", fehler)
		}
	}

	for anweisung, erwartet := range map[string]error{
		`UPDATE sammlung SET notiz = 'neu' WHERE id = 1`:      nil,
		`UPDATE sammlung SET menge = 3 WHERE id = 1`:          nil,
		`UPDATE sammlung SET menge = 0 WHERE id = 1`:          domain.ErrValidation,
		`UPDATE sammlung SET zustand = 'kaputt' WHERE id = 1`: domain.ErrValidation,
		`UPDATE sammlung SET zustand = 'mint' WHERE id = 1`:   nil,
	} {
		_, fehler := db.ExecContext(ctx, anweisung)

		if eingeordnet := einordnen(fehler); (erwartet == nil) != (eingeordnet == nil) ||
			(erwartet != nil && !errors.Is(eingeordnet, erwartet)) {
			t.Errorf("%s: %v, erwartet %v", anweisung, eingeordnet, erwartet)
		}
	}
}
