package service_test

// **Die offenen Aufträge über einen Neubau hinweg** — Werkstück A7 von
// Phase 5.3 (#1446).

import (
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// einWerk baut ein Auftragswerk mit Karten, die sich nicht ändern.
func einKIWerk(t *testing.T) *service.KIAuftragswerk {
	t.Helper()

	return service.NeuesKIAuftragswerk(
		derKatalog(t), &kiKartenprobe{revision: "stand-abc"}, &kostenprobe{},
		&schemaprobe{}, service.KIOrtsregel{},
		func() time.Time { return derLauf })
}

// **Ein Vorschlag überlebt den Neubau des Werks.**
//
// # Warum es diesen Weg gibt
//
// **Weil der Katalog bei jedem Ruf neu entsteht und die Aufträge nicht.** Die
// Naht baut das Werk neu, sobald sich die Laufzeit geändert haben könnte;
// ohne die Übergabe wäre jeder Vorschlag beim Übernehmen „gibt es nicht".
func TestEinVorschlagUeberlebtDenNeubauDesWerks(t *testing.T) {
	t.Parallel()

	erstes := einKIWerk(t)

	auftrag, fehler := erstes.Anlegen(t.Context(), einWunsch())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, fehler := erstes.Vorschlagen(t.Context(), domain.KIVorschlag{
		Auftrag: auftrag.Kennung, Felder: map[string][]string{"name": {"Probe"}},
	}); fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	zweites := einKIWerk(t).MitOffenen(erstes)

	uebernahme, uebernahmefehler := zweites.Uebernehmen(
		t.Context(), auftrag.Kennung, []string{"name"}, 7)
	if uebernahmefehler != nil {
		t.Fatalf("Uebernehmen am neuen Werk: %v", uebernahmefehler)
	}

	if uebernahme.Felder != 1 {
		t.Errorf("%d Felder geschrieben, erwartet 1", uebernahme.Felder)
	}
}

// **Ohne Vorgänger bleibt das Werk, wie es ist.**
//
// **Die Gegenprobe**: ein `nil` heißt „es gab keinen", und ein Werk, das
// daran scheiterte, ließe sich beim ersten Ruf gar nicht bauen.
func TestOhneVorgaengerBleibtDasWerkWieEsIst(t *testing.T) {
	t.Parallel()

	werk := einKIWerk(t).MitOffenen(nil)

	auftrag, fehler := werk.Anlegen(t.Context(), einWunsch())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if auftrag.Kennung == "" {
		t.Error("der Auftrag hat keine Kennung")
	}
}

// **Und ein Vorgänger ohne offene Aufträge gibt nichts weiter.**
func TestEinLeererVorgaengerGibtNichtsWeiter(t *testing.T) {
	t.Parallel()

	zweites := einKIWerk(t).MitOffenen(einKIWerk(t))

	if _, fehler := zweites.Uebernehmen(t.Context(), "ki-4-1", []string{"name"}, 7); fehler == nil {
		t.Fatal("ein Auftrag, den es nie gab, ließ sich übernehmen")
	}
}

// **Ein Auftrag, der erst nach dem Neubau entsteht, ist trotzdem da.**
//
// # Der Befund dahinter
//
// **Zwischen „Werk bauen" und „Auftrag anlegen" liegt eine Lücke.** Die Naht
// gibt das Werk zurück, und **danach** legt der Aufrufer den Auftrag an. Wer
// in dieser Lücke ein zweites Werk baute, kopierte eine Karte ohne diesen
// Auftrag — er landete im ersten, und die Übernahme suchte ihn im zweiten.
//
// Angestrichen vom Zweitblick über Phase 5.3, Befund 1. Der Fall steht hier
// **ohne Goroutinen**: die Reihenfolge »zweites Werk bauen, dann im ersten
// anlegen« ist genau der Zustand, den zwei gleichzeitige Bindungsrufe
// herstellen — und sie ist bestimmt statt zufällig.
func TestEinAuftragNachDemNeubauIstTrotzdemDa(t *testing.T) {
	t.Parallel()

	erstes := einKIWerk(t)

	// **Das zweite Werk entsteht, bevor der Auftrag steht** — die Lücke.
	zweites := einKIWerk(t).MitOffenen(erstes)

	auftrag, fehler := erstes.Anlegen(t.Context(), einWunsch())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, fehler := erstes.Vorschlagen(t.Context(), domain.KIVorschlag{
		Auftrag: auftrag.Kennung, Felder: map[string][]string{"name": {"Probe"}},
	}); fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	uebernahme, uebernahmefehler := zweites.Uebernehmen(
		t.Context(), auftrag.Kennung, []string{"name"}, 7)
	if uebernahmefehler != nil {
		t.Fatalf("Uebernehmen am neuen Werk: %v", uebernahmefehler)
	}

	if uebernahme.Felder != 1 {
		t.Errorf("%d Felder geschrieben, erwartet 1", uebernahme.Felder)
	}
}

// **Und eine Übernahme am alten Werk verbraucht die Kennung auch für das neue.**
//
// **Die Gegenprobe zum Teilen**: hätte jedes Werk seine eigene Karte, stünde
// der eingelöste Auftrag im anderen weiterhin offen — und ließe sich ein
// zweites Mal auf denselben Stand legen.
func TestEineUebernahmeAmAltenWerkVerbrauchtDieKennungAuchImNeuen(t *testing.T) {
	t.Parallel()

	erstes := einKIWerk(t)

	auftrag, fehler := erstes.Anlegen(t.Context(), einWunsch())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, fehler := erstes.Vorschlagen(t.Context(), domain.KIVorschlag{
		Auftrag: auftrag.Kennung, Felder: map[string][]string{"name": {"Probe"}},
	}); fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	zweites := einKIWerk(t).MitOffenen(erstes)

	if _, fehler := erstes.Uebernehmen(
		t.Context(), auftrag.Kennung, []string{"name"}, 7); fehler != nil {
		t.Fatalf("Uebernehmen am alten Werk: %v", fehler)
	}

	if _, fehler := zweites.Uebernehmen(
		t.Context(), auftrag.Kennung, []string{"name"}, 7); fehler == nil {
		t.Fatal("derselbe Vorschlag ließ sich am neuen Werk ein zweites Mal übernehmen")
	}
}

// **Der Stand des Auftrags reist bis zur Ablage mit** — Befund 2 des
// Zweitblicks über Phase 5.3.
//
// # Warum das hier steht und nicht nur im Bestand
//
// **Weil `pruefeStand` ihn nur liest.** Dass die Karte sich seither nicht
// geändert hat, ist eine Auskunft über die Vergangenheit; eine Bedingung wird
// daraus erst, wenn sie **mit** dem Schreiben stattfindet. Die Ablage kann sie
// nur halten, wenn sie sie bekommt — und genau das prüft dieser Fall.
func TestDerStandDesAuftragsReistBisZurAblage(t *testing.T) {
	t.Parallel()

	karten := &kiKartenprobe{revision: "stand-abc"}

	werk := service.NeuesKIAuftragswerk(
		derKatalog(t), karten, &kostenprobe{}, &schemaprobe{},
		service.KIOrtsregel{}, func() time.Time { return derLauf })

	auftrag, fehler := werk.Anlegen(t.Context(), einWunsch())
	if fehler != nil {
		t.Fatalf("Anlegen: %v", fehler)
	}

	if _, fehler := werk.Vorschlagen(t.Context(), domain.KIVorschlag{
		Auftrag: auftrag.Kennung, Felder: map[string][]string{"name": {"Probe"}},
	}); fehler != nil {
		t.Fatalf("Vorschlagen: %v", fehler)
	}

	if _, fehler := werk.Uebernehmen(
		t.Context(), auftrag.Kennung, []string{"name"}, 7); fehler != nil {
		t.Fatalf("Uebernehmen: %v", fehler)
	}

	// **Die Revision des Auftrags und nicht die von heute** — sie sind hier
	// dieselbe Zeichenkette, und das ist der Punkt: der Auftrag hat sie
	// eingefroren, und genau die muss die Ablage halten.
	if len(karten.erwartet) != 1 || karten.erwartet[0] != auftrag.Zielrevision {
		t.Fatalf("an der Ablage kam %q an, erwartet %q",
			karten.erwartet, auftrag.Zielrevision)
	}
}
