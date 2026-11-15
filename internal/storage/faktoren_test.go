package storage_test

// **Der zweite Faktor im Bestand** — Werkstück 7 von Phase 4.12
// (#1351).
//
// **Ohne `t.Parallel()`**, wie jeder Lauf in diesem Paket: `ImArbeitsspeicher`
// ist **eine** Datenquelle, und zwei Migrationen darauf sperren einander —
// „die Migration schlug fehl: database table is locked", gesehen und nicht
// vermutet (siehe `frischeDatenbank`).

import (
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Faktor kommt so zurück, wie er hineinging.**
func TestEinFaktorKommtZurueck(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	angelegt, fehler := ablage.FaktorAnlegen(ctx, domain.Faktor{
		NutzerID: 1, Art: domain.FaktorTOTP, Bezeichnung: "iPhone",
	})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if angelegt.ID == 0 {
		t.Fatal("ohne Kennung angelegt")
	}

	gelesen, lesefehler := ablage.Faktoren(ctx, 1)
	if lesefehler != nil {
		t.Fatalf("lesen: %v", lesefehler)
	}

	if len(gelesen) != 1 || gelesen[0].Art != domain.FaktorTOTP ||
		gelesen[0].Bezeichnung != "iPhone" {
		t.Errorf("gelesen wurde %+v", gelesen)
	}
}

// **Ein Nutzer ohne Faktoren bekommt eine leere Liste und kein `nil`** (§3).
func TestEinNutzerOhneFaktorenBekommtEineListe(t *testing.T) {
	gelesen, fehler := storage.NeueNutzerablage(frischeDatenbank(t)).
		Faktoren(probekontext(t), 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if gelesen == nil {
		t.Error("die Liste ist nil")
	}
}

// **Dieselbe Art zweimal ist ein Konflikt.**
//
// **Zwei TOTP-Faktoren wären zwei Geheimnisse für dieselbe Person**, und die
// Anmeldung müsste raten, welches gemeint ist.
func TestDieselbeArtZweimalIstEinKonflikt(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if _, fehler := ablage.FaktorAnlegen(ctx, domain.Faktor{
		NutzerID: 1, Art: domain.FaktorTOTP,
	}); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	_, fehler := ablage.FaktorAnlegen(ctx, domain.Faktor{
		NutzerID: 1, Art: domain.FaktorTOTP,
	})
	if !errors.Is(fehler, domain.ErrConflict) {
		t.Fatalf("angenommen wurde: %v", fehler)
	}
}

// **Ein Faktor lässt sich wieder herausnehmen.**
func TestEinFaktorLaesstSichHerausnehmen(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if _, fehler := ablage.FaktorAnlegen(ctx, domain.Faktor{
		NutzerID: 1, Art: domain.FaktorBiometrie,
	}); fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if fehler := ablage.FaktorLoeschen(ctx, 1, domain.FaktorBiometrie); fehler != nil {
		t.Fatalf("löschen: %v", fehler)
	}

	gelesen := mussFaktoren(t, ablage, 1)
	if len(gelesen) != 0 {
		t.Errorf("noch da: %+v", gelesen)
	}

	// **Und ein zweites Löschen sagt, dass es ihn nicht gibt.** „Schon weg"
	// und „hat er nie gehabt" sind dasselbe — aber „nichts getan" darf nicht
	// wie „getan" aussehen (§8).
	if fehler := ablage.FaktorLoeschen(ctx, 1, domain.FaktorBiometrie); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Wiederherstellungscodes werden gesetzt, gezählt und verbraucht.**
func TestWiederherstellungscodesWerdenVerbraucht(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if fehler := ablage.CodesSetzen(ctx, 1, []string{"eins", "zwei"}); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	frei, fehler := ablage.FreieCodes(ctx, 1)
	if fehler != nil {
		t.Fatalf("lesen: %v", fehler)
	}

	if len(frei) != 2 {
		t.Fatalf("frei sind %+v", frei)
	}

	if fehler := ablage.CodeVerbrauchen(ctx, frei[0].ID); fehler != nil {
		t.Fatalf("verbrauchen: %v", fehler)
	}

	danach := mussCodes(t, ablage, 1)
	if len(danach) != 1 {
		t.Errorf("frei sind %+v", danach)
	}

	// **Zweimal derselbe Code geht nicht.** Ein Wiederherstellungscode ist
	// einmalig, und ohne diese Zusage wäre er ein zweites Kennwort.
	if fehler := ablage.CodeVerbrauchen(ctx, frei[0].ID); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("gemeldet wurde: %v", fehler)
	}
}

// **Neue Codes ersetzen die alten und lassen keine liegen.**
func TestNeueCodesErsetzenDieAlten(t *testing.T) {
	ablage := storage.NeueNutzerablage(frischeDatenbank(t))
	ctx := probekontext(t)

	if fehler := ablage.CodesSetzen(ctx, 1, []string{"eins", "zwei"}); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	if fehler := ablage.CodesSetzen(ctx, 1, []string{"drei"}); fehler != nil {
		t.Fatalf("setzen: %v", fehler)
	}

	frei := mussCodes(t, ablage, 1)
	if len(frei) != 1 || frei[0].Gehuetet != "drei" {
		t.Errorf("frei sind %+v", frei)
	}
}

// **Wer geht, nimmt seine Faktoren und Codes mit.**
//
// **`ON DELETE CASCADE`, und das ist hier richtig** — anders als bei den Sets
// eines Profils: ein Faktor ohne seinen Nutzer ist kein Wert, den jemand
// später wiederfinden will, sondern ein Rest.
func TestWerGehtNimmtSeineFaktorenMit(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueNutzerablage(db)
	ctx := probekontext(t)

	zweiter, fehler := ablage.NutzerAnlegen(ctx, domain.Nutzer{Name: "Zweite"})
	if fehler != nil {
		t.Fatalf("anlegen: %v", fehler)
	}

	if _, err := ablage.FaktorAnlegen(ctx, domain.Faktor{
		NutzerID: zweiter.ID, Art: domain.FaktorTOTP,
	}); err != nil {
		t.Fatalf("faktor: %v", err)
	}

	if err := ablage.CodesSetzen(ctx, zweiter.ID, []string{"eins"}); err != nil {
		t.Fatalf("codes: %v", err)
	}

	if err := ablage.NutzerLoeschen(ctx, zweiter.ID, "Zweite"); err != nil {
		t.Fatalf("löschen: %v", err)
	}

	if gelesen := mussFaktoren(t, ablage, zweiter.ID); len(gelesen) != 0 {
		t.Errorf("Faktoren blieben: %+v", gelesen)
	}

	if frei := mussCodes(t, ablage, zweiter.ID); len(frei) != 0 {
		t.Errorf("Codes blieben: %+v", frei)
	}
}

// mussFaktoren liest die Faktoren und bricht ab, wenn es nicht glückt.
func mussFaktoren(
	t *testing.T, ablage *storage.Nutzerablage, nutzerID int64,
) []domain.Faktor {
	t.Helper()

	gelesen, fehler := ablage.Faktoren(probekontext(t), nutzerID)
	if fehler != nil {
		t.Fatalf("die Faktoren lesen: %v", fehler)
	}

	return gelesen
}

// mussCodes liest die freien Codes und bricht ab, wenn es nicht glückt.
func mussCodes(
	t *testing.T, ablage *storage.Nutzerablage, nutzerID int64,
) []domain.Wiederherstellungscode {
	t.Helper()

	frei, fehler := ablage.FreieCodes(probekontext(t), nutzerID)
	if fehler != nil {
		t.Fatalf("die Codes lesen: %v", fehler)
	}

	return frei
}
