package service

// **Die atomare Gruppe** — Werkstück A13.1b von Phase 5.3
// (#1452).
//
// # Der Satz, für den es diese Datei gibt
//
// > Gewählte Gruppe entweder vollständig mit Herkunft übernommen oder klarer
// > Konflikt **ohne Teilbestand**; kein Rückrollen bereits separat
// > bestätigter Gruppen.
//
// # Warum das nicht die Schleife über [KIAuftragswerk.Uebernehmen] ist
//
// **Weil die zweite Zusage die erste bricht.** Eine Schleife, die beim
// Scheitern zurücknimmt, nimmt auch zurück, was eine **frühere**, bestätigte
// Gruppe geschrieben hat — sie kennt den Unterschied nicht. Und eine Schleife
// **ohne** Rücknahme lässt genau den Teilbestand stehen, den der Satz
// ausschließt.
//
// Deshalb zerfällt der Weg in zwei Hälften: **erst prüfen, dann schreiben**.
// Geprüft wird jeder Schritt vollständig — Auftrag, Vorschlag, Auswahl,
// Frist, Stand —, und erst wenn **alle** durch sind, geht **ein** Ruf an die
// Ablage. Zurückzunehmen gibt es dann nichts, und die frühere Gruppe wird nie
// angefasst.

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// KIGruppenwahl ist ein Vorschlag mit den Feldern, die daraus gelten sollen.
type KIGruppenwahl struct {
	Kennung string
	Felder  []string
}

// KIGruppenschritt ist, was die Ablage je Karte zu schreiben bekommt.
//
// **Er trägt die Herkunft mit** und nicht daneben: eine Übernahme ohne sie
// ist von einem getippten Wert später nicht zu unterscheiden (A11, #1450).
type KIGruppenschritt struct {
	Karte    int64
	Felder   map[string][]string
	Herkunft []domain.KIFeldherkunft

	// Erwartet ist der Stand, auf dem die Karte stehen muss.
	//
	// **Er reist bis in die Transaktion mit** (Befund 2 des Zweitblicks über
	// Phase 5.3): `gruppePruefen` liest ihn, bevor irgendetwas geschrieben
	// ist, und zwischen dieser Prüfung und dem Commit kann eine zweite
	// Übernahme dieselbe Karte ändern.
	Erwartet string
}

// KIGruppenuebernahme sagt, was aus der Gruppe wurde.
type KIGruppenuebernahme struct {
	// Karten ist die Zahl der geschriebenen Karten.
	Karten int

	// Felder ist die Zahl der geschriebenen Felder über alle Karten.
	Felder int

	// Namen sind die geschriebenen Feldnamen je Karte, in fester Ordnung.
	Namen map[int64][]string
}

// GruppeUebernehmen schreibt mehrere Vorschläge als **eine** Gruppe.
//
// **Ganz oder gar nicht.** Scheitert ein Schritt, ist nichts geschrieben und
// **kein** Schlüssel verbraucht — ein zweiter Versuch ist möglich.
// **`profil` ist der Mensch, der jetzt fragt** — dieselbe Grenze wie bei
// [KIAuftragswerk.Uebernehmen], und je Schritt: eine Gruppe aus zwei Profilen
// gibt es nicht.
func (w *KIAuftragswerk) GruppeUebernehmen(
	ctx context.Context, wahl []KIGruppenwahl, profil int64,
) (KIGruppenuebernahme, error) {
	if len(wahl) == 0 {
		return KIGruppenuebernahme{}, fmt.Errorf(
			"%w: die Gruppe ist leer — eine Übernahme sagt, was sie "+
				"übernimmt", domain.ErrValidation)
	}

	schritte, laeufe, fehler := w.gruppePruefen(ctx, wahl, profil)
	if fehler != nil {
		return KIGruppenuebernahme{}, fehler
	}

	// **Ein Ruf und nicht einer je Karte.** Zwei Rufe wären zwei
	// Transaktionen, und dazwischen läge der halbe Stand.
	if fehler := w.karten.FelderGruppeSetzen(ctx, schritte); fehler != nil {
		return KIGruppenuebernahme{}, fehler
	}

	// **Die Schlüssel erst danach** — scheitert das Schreiben, bleiben die
	// Aufträge offen, so wie bei der einzelnen Übernahme.
	w.buch.verbrauche(laeufe...)

	return gruppenbericht(schritte), nil
}

// gruppePruefen prüft **jeden** Schritt und baut, was zu schreiben wäre.
//
// **Es schreibt nichts.** Das ist die ganze Arbeitsteilung dieser Datei: was
// hier durchkommt, geht danach in einem Zug hinaus.
func (w *KIAuftragswerk) gruppePruefen(
	ctx context.Context, wahl []KIGruppenwahl, profil int64,
) (schritte []KIGruppenschritt, kennungen []string, fehler error) {
	schritte = make([]KIGruppenschritt, 0, len(wahl))
	kennungen = make([]string, 0, len(wahl))

	// **Eine Karte darf nur einmal vorkommen.** Zweimal hieße, dass die
	// zweite Zeile die erste in derselben Transaktion überschreibt — welcher
	// Wert am Ende stünde, hinge an der Reihenfolge der Auswahl.
	gesehen := make(map[int64]bool, len(wahl))

	for _, eine := range wahl {
		lauf, gibts := w.lauf(eine.Kennung)
		if !gibts || lauf.auftrag.ProfilID != profil {
			// **Die Profilgrenze je Schritt** — dieselbe Meldung wie bei der
			// einzelnen Übernahme, und aus demselben Grund (Befund 1 des
			// Nachlaufs über Phase 5.3).
			return nil, nil, fmt.Errorf("%w: den Auftrag %q gibt es nicht",
				domain.ErrNotFound, eine.Kennung)
		}

		if lauf.vorschlag == nil {
			return nil, nil, fmt.Errorf("%w: zu %q gibt es keinen Vorschlag",
				domain.ErrNotFound, eine.Kennung)
		}

		gewaehlt, fehler := gewaehlteFelder(lauf.vorschlag.Felder, eine.Felder)
		if fehler != nil {
			return nil, nil, fehler
		}

		if lauf.auftrag.Abgelaufen(w.jetzt()) {
			return nil, nil, fmt.Errorf(
				"%w: die Frist für %q ist abgelaufen — ein Vorschlag von "+
					"gestern trifft den Stand von heute vielleicht nicht mehr",
				domain.ErrConflict, eine.Kennung)
		}

		if fehler := w.pruefeStand(ctx, lauf.auftrag); fehler != nil {
			return nil, nil, fehler
		}

		if gesehen[lauf.auftrag.Zielkarte] {
			return nil, nil, fmt.Errorf(
				"%w: die Karte %d steht zweimal in der Gruppe — die zweite "+
					"Zeile überschriebe die erste, ohne dass es jemand sieht",
				domain.ErrValidation, lauf.auftrag.Zielkarte)
		}

		gesehen[lauf.auftrag.Zielkarte] = true

		schritte = append(schritte, KIGruppenschritt{
			Karte: lauf.auftrag.Zielkarte, Felder: gewaehlt,
			Herkunft: w.herkunftFuer(lauf, gewaehlt),

			// **Der Stand reist mit bis in die Transaktion** (Befund 2 des
			// Zweitblicks über Phase 5.3): `pruefeStand` hat ihn gerade
			// gelesen, und bis zum Commit kann ihn ein zweiter Schreibweg
			// verschieben.
			Erwartet: lauf.auftrag.Zielrevision,
		})
		kennungen = append(kennungen, eine.Kennung)
	}

	return schritte, kennungen, nil
}

// gruppenbericht zählt, was geschrieben wurde.
func gruppenbericht(schritte []KIGruppenschritt) KIGruppenuebernahme {
	bericht := KIGruppenuebernahme{
		Karten: len(schritte),
		Namen:  make(map[int64][]string, len(schritte)),
	}

	for _, schritt := range schritte {
		bericht.Felder += len(schritt.Felder)
		bericht.Namen[schritt.Karte] = feldnamenAus(schritt.Felder)
	}

	return bericht
}
