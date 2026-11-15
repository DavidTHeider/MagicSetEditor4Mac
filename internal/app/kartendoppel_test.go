package app_test

// **Das Kartendoppel der KI-Welt** — abgetrennt am 2026-09-13, weil
// `kivorschlag_test.go` mit der gezählten Lesung über die Zeilengrenze aus §2
// ging (A13.1b, #1452).
//
// # Der Schnitt folgt der Frage
//
// Dort stehen die **Läufe** über den Vorschlag; hier die **Ablage**, gegen
// die sie laufen — was sie festhält, was sie zurückgibt und woran sie sich
// auf Wunsch verschluckt.

import (
	"context"
	"slices"
	"sync"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// kartendoppel hält eine Karte im Speicher.
type kartendoppel struct {
	zeile         domain.CardRow
	herkunft      string
	fehler        error
	schreibfehler error

	// geschrieben zählt die Felder, die wirklich gelandet sind — ein Batch
	// sagt zu, während der Generierung **nichts** zu schreiben, und eine
	// Zusicherung über den Rückgabewert prüfte das nicht (A13.1, #1452).
	geschrieben int

	// abLesung lässt das Lesen **erst ab dem n-ten Ruf** scheitern — A13.1b
	// (#1452), null heißt: nie.
	//
	// # Wofür
	//
	// **Weil `service` den Stand prüft, bevor die Naht liest.** Ein Doppel,
	// das immer scheitert, bringt die Gruppe schon dort zu Fall, und der
	// Lesefehler **in** der Naht bleibt unbetreten — obwohl er im Betrieb
	// vorkommt: zwischen der Standprüfung und dem Schreiben kann ein zweiter
	// Prozess die Karte löschen ([0010](../../docs/entscheidungen/0010-mcp-als-eigener-prozess.md)
	// sieht mehrere ausdrücklich vor).
	//
	// # Und warum ein Schloss dazugehört
	//
	// **Weil ein Batch seine Ziele nebenläufig fährt** (`app.batchFahren`).
	// Der erste Anlauf zählte ohne, und der Race-Detektor hat es beim ersten
	// vollen Lauf angestrichen — an **diesem** Zähler, nicht am
	// Produktivcode. Ein Doppel mit Zustand ist ein Doppel mit Nebenläufigkeit.
	lesezaehler sync.Mutex
	abLesung    int
	lesungen    int
	leseklemm   error

	// feldherkuenfte hält, was eine Übernahme mitschickte (A11, #1450).
	feldherkuenfte []domain.KIFeldherkunft

	// bildangaben hält, was eine Bildübernahme daneben mitschickte
	// (A12, #1451).
	bildangaben map[string][]domain.KIBildparameter

	// erwartet hält, welcher Stand als Bedingung ankam (Befund 2 des
	// Zweitblicks über Phase 5.3).
	erwartet []string

	// panikBei lässt das Lesen dieser Karte **zusammenbrechen** — BK-05b-KI
	// (#2112); null heißt: nie.
	panikBei int64
}

func (k *kartendoppel) KarteLesen(_ context.Context, karteID int64) (*domain.CardRow, error) {
	if k.panikBei != 0 && karteID == k.panikBei {
		panic("das Lesen dieser Karte bricht zusammen")
	}

	if k.fehler != nil {
		return nil, k.fehler
	}

	k.lesezaehler.Lock()
	k.lesungen++
	klemmt := k.abLesung > 0 && k.lesungen >= k.abLesung
	klemme := k.leseklemm
	k.lesezaehler.Unlock()

	if klemmt {
		return nil, klemme
	}

	kopie := k.zeile
	kopie.Values = slices.Clone(k.zeile.Values)

	return &kopie, nil
}

func (k *kartendoppel) Herkuenfte(context.Context, int64) (map[string]string, error) {
	return map[string]string{}, nil
}

func (k *kartendoppel) FeldSetzenMitVorwert(
	context.Context, int64, int64, string, string,
) (vorher string, hatte bool, fehler error) {
	return "", false, nil
}

func (k *kartendoppel) FeldSetzenVonAussen(
	context.Context, int64, int64, string, string, string,
) (vorher string, hatte bool, fehler error) {
	return "", false, nil
}

// FelderVonAussen schreibt — und merkt sich die Herkunft.
// FelderVonAussenMitHerkunft ist derselbe Weg samt Feldherkunft (A11, #1450).
func (k *kartendoppel) FelderVonAussenMitHerkunft(
	ctx context.Context, setID, karteID int64, werte []domain.Value,
	_ string, herkuenfte []domain.KIFeldherkunft,
	angaben map[string][]domain.KIBildparameter, erwartet string,
) (vorher string, hatte bool, fehler error) {
	k.feldherkuenfte = append(k.feldherkuenfte, herkuenfte...)
	k.bildangaben = angaben
	k.erwartet = append(k.erwartet, erwartet)

	return k.FelderVonAussen(ctx, setID, karteID, werte, "ki")
}

func (k *kartendoppel) FelderVonAussen(
	_ context.Context, _, _ int64, werte []domain.Value, herkunft string,
) (vorher string, hatte bool, fehler error) {
	if k.schreibfehler != nil {
		return "", false, k.schreibfehler
	}

	k.herkunft = herkunft
	k.geschrieben += len(werte)

	for _, wert := range werte {
		gesetzt := false

		for i := range k.zeile.Values {
			if k.zeile.Values[i].Field == wert.Field {
				k.zeile.Values[i].Text = wert.Text
				gesetzt = true
			}
		}

		if !gesetzt {
			k.zeile.Values = append(k.zeile.Values, wert)
		}
	}

	return "", false, nil
}

// gelesen gibt, wie oft dieses Doppel gelesen wurde — **unter dem Schloss**.
//
// **Ein eigener Leser und kein Griff ans Feld**: der Batch schreibt es aus
// mehreren Goroutinen, und ein Test, der nebenher hineinsieht, ist selbst das
// Rennen.
func (k *kartendoppel) gelesen() int {
	k.lesezaehler.Lock()
	defer k.lesezaehler.Unlock()

	return k.lesungen
}

// klemmeAb lässt das Lesen ab dem n-ten Ruf scheitern — unter dem Schloss.
func (k *kartendoppel) klemmeAb(ab int, grund error) {
	k.lesezaehler.Lock()
	defer k.lesezaehler.Unlock()

	k.abLesung, k.leseklemm = ab, grund
}
