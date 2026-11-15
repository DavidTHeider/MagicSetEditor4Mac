package storage_test

// **Die Kontaktablage** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Was hier gemessen wird
//
// **Dass ein Kontakt dem Profil gehört, das ihn kennt.** Eine Freigabe, die
// nach einem Profilwechsel weitergälte, wäre eine, die niemand gegeben hat.
//
// **Und dass eine Rotation jede Bestätigung aufhebt.** Die Abnahme des Issues
// verlangt es: „Rotation erzwingt Neuprüfung bekannter Kontakte."

import (
	"context"
	"errors"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// begegne trägt ein Gegenüber ein.
func begegne(t *testing.T, ablage *storage.Kontaktablage,
	ctx context.Context, kennung, name string,
) {
	t.Helper()

	if fehler := ablage.Begegnet(
		ctx, kennung, "abcd-ef01", name, derPeerlauf); fehler != nil {
		t.Fatalf("Begegnet: %v", fehler)
	}
}

// **Wer noch niemandem begegnet ist, hat eine leere Liste.**
func TestOhneBegegnungIstDieListeLeer(t *testing.T) {
	db := frischeDatenbank(t)

	liste, fehler := storage.NeueKontaktablage(db).Liste(probekontext(t))
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	if len(liste) != 0 {
		t.Fatalf("es stehen %d darin", len(liste))
	}
}

// **Eine Begegnung hinterlässt einen unbestätigten Kontakt.**
func TestEineBegegnungHinterlaesstEinenUnbestaetigtenKontakt(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)

	begegne(t, ablage, probekontext(t), "bgegenueber", "Fremder Rechner")

	gelesen, fehler := ablage.Lesen(probekontext(t), "bgegenueber")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if gelesen.Stand != domain.KontaktUnbestaetigt {
		t.Errorf("der Stand ist %q", gelesen.Stand)
	}

	if gelesen.SelbstgewaehlterName != "Fremder Rechner" {
		t.Errorf("der Name ist %q", gelesen.SelbstgewaehlterName)
	}

	if len(gelesen.Rechte) != 0 {
		t.Errorf("er hat schon Rechte: %v", gelesen.Rechte)
	}
}

// **Eine zweite Begegnung hebt die Bestätigung nicht auf.**
//
// # Der Fund
//
// **Ein Einfügen mit Ersetzen täte genau das.** Wer den Kontakt bei jeder
// Begegnung neu schreibt, setzt den Stand zurück auf `unbestaetigt` — und
// jede Freigabe wäre nach dem nächsten Hallo fort, ohne dass jemand etwas
// entzogen hätte.
func TestEineZweiteBegegnungHebtDieBestaetigungNichtAuf(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bgegenueber", "Fremder Rechner")

	if fehler := ablage.Stand(
		ctx, "bgegenueber", domain.KontaktBestaetigt); fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	if fehler := ablage.Erlauben(
		ctx, "bgegenueber", domain.ZweckTeilen, derPeerlauf); fehler != nil {
		t.Fatalf("Erlauben: %v", fehler)
	}

	begegne(t, ablage, ctx, "bgegenueber", "Neuer Name")

	gelesen, fehler := ablage.Lesen(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if gelesen.Stand != domain.KontaktBestaetigt {
		t.Errorf("der Stand ist %q", gelesen.Stand)
	}

	if len(gelesen.Rechte) != 1 || gelesen.Rechte[0] != domain.ZweckTeilen {
		t.Errorf("die Rechte sind %v", gelesen.Rechte)
	}

	// **Der selbstgewählte Name folgt dagegen** — er ist eine Auskunft des
	// Gegenübers und kein Stand dieses Profils.
	if gelesen.SelbstgewaehlterName != "Neuer Name" {
		t.Errorf("der Name ist %q", gelesen.SelbstgewaehlterName)
	}
}

// **Ein Kontakt eines anderen Profils ist keiner.**
func TestEinKontaktEinesAnderenProfilsIstKeiner(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)

	begegne(t, ablage, probekontext(t), "bgegenueber", "Fremder Rechner")

	zweites := storage.MitProfil(t.Context(), zweitesProfil(t, storage.NeueNutzerablage(db), "Turnier"))

	if _, fehler := ablage.Lesen(zweites, "bgegenueber"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Fatalf("das fremde Profil las ihn: %v", fehler)
	}

	liste, fehler := ablage.Liste(zweites)
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	if len(liste) != 0 {
		t.Fatalf("das fremde Profil sieht %d Kontakte", len(liste))
	}
}

// **Wen es nicht gibt, den gibt es nicht.**
func TestWenEsNichtGibtDenGibtEsNicht(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	if _, fehler := ablage.Lesen(ctx, "bniemand"); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Lesen: %v", fehler)
	}

	if fehler := ablage.Stand(
		ctx, "bniemand", domain.KontaktBestaetigt); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Stand: %v", fehler)
	}

	if fehler := ablage.Erlauben(
		ctx, "bniemand", domain.ZweckTeilen, derPeerlauf); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Erlauben: %v", fehler)
	}

	if fehler := ablage.Entziehen(
		ctx, "bniemand", domain.ZweckTeilen); !errors.Is(
		fehler, domain.ErrNotFound) {
		t.Errorf("Entziehen: %v", fehler)
	}
}

// **Auch nicht, wenn jemand es gern hätte.** Der Weg dorthin ist
// [storage.Kontaktablage.Stand] und ein Mensch, der einen Fingerabdruck
// vergleicht — die Begegnung selbst weiß darüber nichts.
func TestEineBegegnungMachtNiemandenBestaetigt(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bgegenueber", "Fremder Rechner")

	gelesen, fehler := ablage.Lesen(ctx, "bgegenueber")
	if fehler != nil {
		t.Fatalf("Lesen: %v", fehler)
	}

	if gelesen.Darf(domain.ZweckTeilen) || gelesen.Darf(domain.ZweckHandel) {
		t.Fatalf("er darf schon etwas: %+v", gelesen)
	}

	if !gelesen.Darf(domain.ZweckPaarung) {
		t.Error("er darf sich nicht einmal vorstellen")
	}
}

// **Eine leere Kennung ist kein Kontakt.**
func TestEineLeereKennungIstKeinKontakt(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)

	if fehler := ablage.Begegnet(
		probekontext(t), "", "", "", derPeerlauf); !errors.Is(
		fehler, domain.ErrValidation) {
		t.Fatalf("sie kam durch: %v", fehler)
	}
}

// **Die Liste zeigt die Kontakte samt ihren Rechten.**
func TestDieListeZeigtDieKontakteSamtRechten(t *testing.T) {
	db := frischeDatenbank(t)
	ablage := storage.NeueKontaktablage(db)
	ctx := probekontext(t)

	begegne(t, ablage, ctx, "bzweiter", "Zweiter")
	begegne(t, ablage, ctx, "berster", "Erster")

	if fehler := ablage.Stand(
		ctx, "berster", domain.KontaktBestaetigt); fehler != nil {
		t.Fatalf("Stand: %v", fehler)
	}

	if fehler := ablage.Erlauben(
		ctx, "berster", domain.ZweckPartie, derPeerlauf); fehler != nil {
		t.Fatalf("Erlauben: %v", fehler)
	}

	liste, fehler := ablage.Liste(ctx)
	if fehler != nil {
		t.Fatalf("Liste: %v", fehler)
	}

	if len(liste) != 2 {
		t.Fatalf("es stehen %d darin", len(liste))
	}

	// **Nach Kennung geordnet** — eine Liste ohne Ordnung sieht bei jedem
	// Lauf anders aus, und niemand könnte sie vergleichen.
	if liste[0].Kennung != "berster" || liste[1].Kennung != "bzweiter" {
		t.Fatalf("die Ordnung ist %q, %q", liste[0].Kennung, liste[1].Kennung)
	}

	if len(liste[0].Rechte) != 1 || liste[0].Rechte[0] != domain.ZweckPartie {
		t.Errorf("die Rechte des ersten sind %v", liste[0].Rechte)
	}

	if len(liste[1].Rechte) != 0 {
		t.Errorf("der zweite hat schon Rechte: %v", liste[1].Rechte)
	}
}
