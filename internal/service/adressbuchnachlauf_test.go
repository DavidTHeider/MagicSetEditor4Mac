package service_test

// **Der Nachlauf über ID-04** — zwei Befunde am Abgleich
// (#1703).

import (
	"errors"
	"strings"
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/service"
)

// **Gleiche Karten ohne UID und ohne Weg sind dieselbe** — Befund 1: ohne
// Ursprung verlangte der Abgleich einen gemeinsamen Weg, und zwei Karten,
// die nur einen Namen tragen, kamen doppelt durch — in einer Datei wie bei
// einem zweiten Import.
func TestGleicheKartenNurMitNamenSindDieselbe(t *testing.T) {
	t.Parallel()

	karte := "BEGIN:VCARD\r\nFN:Bo\r\nEND:VCARD\r\n"

	bericht, fehler := service.NeuesAdressbuch(&adressbuchdoppel{}).Vorschau(t.Context(), []byte(karte+karte))
	if fehler != nil || len(bericht.Karten) != 1 || len(bericht.Befunde) != 1 {
		t.Errorf("in einer Datei: %+v %v", bericht, fehler)
	}

	doppel := &adressbuchdoppel{kontakte: []domain.Kontakt{{ID: 3, Fassung: 1, Name: "Bo", Wege: []domain.Kontaktweg{}}}}

	zweiter, fehler := service.NeuesAdressbuch(doppel).Vorschau(t.Context(), []byte(karte))
	if fehler != nil || len(zweiter.Karten) != 1 || zweiter.Karten[0].Abgleich != domain.KontaktGleich {
		t.Errorf("ein zweiter Import: %+v %v", zweiter, fehler)
	}

	if _, _, fehler := service.NeuesAdressbuch(doppel).Uebernehmen(t.Context(), []domain.Kontakt{{Name: "Bo"}}); !errors.Is(fehler, domain.ErrDoppelt) {
		t.Errorf("übernommen trotz gleicher Angaben: %v", fehler)
	}
}

// **Eine Vermutung verschwindet nicht aus der Vorschau** — Befund 2: zwei
// Karten mit verschiedenen UIDs, gleichem Namen und geteilter Adresse sind
// vielleicht zwei Menschen, und die zweite fiel still heraus.
func TestZweiVermutlichGleicheKartenStehenBeideDa(t *testing.T) {
	t.Parallel()

	karte := func(uid string) string {
		return "BEGIN:VCARD\r\nFN:Anna Muster\r\nUID:" + uid + "\r\nEMAIL:familie@example.org\r\nEND:VCARD\r\n"
	}

	bericht, fehler := service.NeuesAdressbuch(&adressbuchdoppel{}).Vorschau(t.Context(), []byte(karte("a")+karte("b")))
	if fehler != nil || len(bericht.Karten) != 2 || len(bericht.Befunde) != 0 {
		t.Fatalf("%+v %v", bericht, fehler)
	}

	// **Und beide lassen sich übernehmen**: der Mensch hat beide gesehen.
	doppel := &adressbuchdoppel{}
	angelegt, _, fehler := service.NeuesAdressbuch(doppel).Uebernehmen(t.Context(),
		[]domain.Kontakt{bericht.Karten[0].Kontakt, bericht.Karten[1].Kontakt})
	if fehler != nil || angelegt != 2 {
		t.Errorf("%d angelegt: %v", angelegt, fehler)
	}

	// **Dieselbe UID zweimal in einer Datei bleibt eine Karte.**
	nochmal, fehler := service.NeuesAdressbuch(&adressbuchdoppel{}).Vorschau(t.Context(), []byte(karte("a")+karte("a")))
	if fehler != nil || len(nochmal.Karten) != 1 || !strings.Contains(strings.Join(nochmal.Befunde, " "), "wiederholt") {
		t.Errorf("dieselbe UID zweimal: %+v", nochmal)
	}
}
