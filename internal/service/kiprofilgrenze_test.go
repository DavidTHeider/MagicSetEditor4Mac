package service_test

// **Ein Vorschlag gehört einem Profil** — Befund 1 des Nachlaufs über Phase
// 5.3.
//
// # Der Befund
//
// **Das Auftragsbuch ist seit dem Zweitblick geteilt**, und es überlebt einen
// Profilwechsel: ein in Profil A angelegter Vorschlag stand danach in Profil
// B weiter offen. Übernommen hätte er die Karte von **A** geschrieben — die
// Kartenkennung steht im Auftrag, und der Kartenleser bindet sie nicht an ein
// Profil —, während die Herkunft das Profil von **B** trüge.
//
// Das ist §7, Regel 1 (ein Wechsel des Bezugsobjekts räumt den offenen
// Kontext) und Regel 2 (eine Antwort gehört zu dem Stand, aus dem sie stammt)
// in einem.
//
// **Geprüft wird hier die Grenze am Auftrag.** Dass das Fenster das Buch beim
// Wechsel außerdem **räumt**, prüft `internal/app`; beides zusammen, weil das
// eine ohne das andere eine halbe Zusage wäre: der Wechsel geht auch über
// Wege, die den Zeiger nicht selbst umstellen.

import (
	"errors"
	"testing"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Ein fremdes Profil übernimmt den Vorschlag nicht.**
func TestEinFremdesProfilUebernimmtDenVorschlagNicht(t *testing.T) {
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

	// **Der Wunsch trug Profil 7** — hier fragt ein anderes.
	_, uebernahmefehler := werk.Uebernehmen(
		t.Context(), auftrag.Kennung, []string{"name"}, 8)

	if !errors.Is(uebernahmefehler, domain.ErrNotFound) {
		t.Fatalf("das fremde Profil kam durch: %v", uebernahmefehler)
	}

	if karten.schreibvorgaenge != 0 {
		t.Errorf("geschrieben wurde %d mal", karten.schreibvorgaenge)
	}

	// **Und im eigenen Profil geht er weiterhin** — die Gegenprobe: eine
	// Grenze, die jeden abwiese, machte die Übernahme unmöglich.
	if _, fehler := werk.Uebernehmen(
		t.Context(), auftrag.Kennung, []string{"name"}, 7); fehler != nil {
		t.Fatalf("das eigene Profil kam nicht durch: %v", fehler)
	}
}

// **Und eine Gruppe hält dieselbe Grenze — für jeden ihrer Schritte.**
func TestEineGruppeHaeltDieProfilgrenzeJeSchritt(t *testing.T) {
	t.Parallel()

	hand := einWerkMitZweiVorschlaegen(t)

	_, fehler := hand.werk.GruppeUebernehmen(t.Context(),
		[]service.KIGruppenwahl{
			{Kennung: hand.eins, Felder: []string{"name"}},
			{Kennung: hand.zwei, Felder: []string{"name"}},
		}, 8)

	if !errors.Is(fehler, domain.ErrNotFound) {
		t.Fatalf("die fremde Gruppe kam durch: %v", fehler)
	}

	if hand.karten.gruppenrufe != 0 {
		t.Errorf("die Ablage bekam %d Rufe", hand.karten.gruppenrufe)
	}
}
