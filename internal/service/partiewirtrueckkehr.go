package service

// **Wie der Wirt einen Gast wieder hereinholt** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Warum der Wirt rechnet und nicht der Gast
//
// **Weil nur er den Strom hat.** Der Gast sagt, wo er stehengeblieben ist;
// **ob** das derselbe Strom ist, kann allein der Wirt beantworten. Die
// Rechnung selbst steht in `domain` — damit beide Seiten dieselbe Antwort
// bekämen, wenn sie sie stellten (§2).
//
// # Und warum die Partie aus der Sitzung kommt
//
// **Weil sie im Wunsch eine Behauptung des Gastes ist** — dieselbe Regel wie
// bei `Spielbefehl.PartieID` und beim Sprecher eines Satzes, und derselbe
// Schaden, wenn sie fehlt.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// rueckkehr beantwortet den Wunsch eines zurückkommenden Gastes.
func (w *Partiewirt) rueckkehr(ctx context.Context, roh []byte) error {
	var wunsch domain.Wiederaufnahmewunsch

	if fehler := json.Unmarshal(roh, &wunsch); fehler != nil {
		return fmt.Errorf("%w: der Inhalt ist kein Wiederaufnahmewunsch: %s",
			domain.ErrValidation, fehler.Error())
	}

	// **Die Partie steht fest und wird nicht gelesen** (siehe den Kopf).
	wunsch.Partie = w.partie

	// **Was der Gast kann, erklärt er hier** (Nachlauf über Phase 7, R4) —
	// vor allem anderen, denn auch ein abgewiesener sitzt am Platz.
	if w.runde != nil {
		w.runde.Kann(w.gast, wunsch.Kann)
	}

	// **Wer eingeladen ist, setzt sich zuerst** (P9b, #2377) — danach ist
	// der Stand, zu dem er zurückkommt, einer mit seinem Deck.
	abgewiesen, fehler := w.hinsetzen(ctx, wunsch.Mitgebracht)
	if fehler != nil {
		return fehler
	}

	// **Stand, Antwort und Nachgereichtes in der Reihe** (#2481): ein Zug,
	// der dazwischen verkündet würde, käme vor dem Strom an, aus dem er folgt
	// — und ein Bild von vorhin danach.
	return w.nehmer.InReihe(func() error {
		stand, fehler := w.nehmer.Wirtsstand(ctx, w.partie, w.gast)
		if fehler != nil {
			return fehler
		}

		weg := domain.Wiederaufnahme(wunsch, stand)
		if abgewiesen != "" {
			weg = domain.Wiederaufnahmeweg{Art: domain.WegAbbruch, Grund: abgewiesen}
		}

		if fehler := w.antworte(ctx, weg, stand); fehler != nil {
			return fehler
		}

		return w.nachholen(ctx, weg)
	})
}

// antworte schickt, was aus dem Wunsch wurde.
//
// **Auch beim Abbruch.** Ein Gast, dem die Leitung ohne ein Wort zugeht, sucht
// den Fehler im Netz; einer, der den Grund gelesen hat, weiß, dass sein Stand
// nicht angenommen wurde (§8).
func (w *Partiewirt) antworte(
	ctx context.Context, weg domain.Wiederaufnahmeweg, stand domain.Wirtsstand,
) error {
	antwort := domain.Rueckkehr{
		Art:     weg.Art,
		Grund:   weg.Grund,
		AbFolge: folgeAus(weg.AbFolge),
		Stand:   folgeAus(stand.Folge),
		Abdruck: stand.Kette,
	}

	if weg.Art == domain.WegAbzug {
		bild, fehler := w.tischbild(ctx)
		if fehler != nil {
			return fehler
		}

		antwort.Bild = bild
	}

	// **Kein Fehlerzweig, und derselbe Grund wie beim Tischbild**: eine
	// [domain.Rueckkehr] trägt Zeichenketten, Zahlen und ein Tischbild — und
	// dass **das** nichts Offenes enthält, hält
	// `internal/domain/sichtgestalt_test.go` (§10).
	inhalt, _ := domain.VerpackeKontrolle( //nolint:errcheck // der Grund steht darüber
		domain.KontrolleAbzug, antwort)

	return w.sitzung.Sende(ctx, domain.ArtKontrolle, inhalt)
}

// nachholen reicht die fehlenden Ereignisse nach — und danach das Bild.
//
// **In dieser Reihenfolge**, und aus demselben Grund wie nach einem Zug: das
// Bild zeigt den Stand **nach** allem, was davor geschah.
func (w *Partiewirt) nachholen(
	ctx context.Context, weg domain.Wiederaufnahmeweg,
) error {
	switch weg.Art {
	case domain.WegAbbruch:
		// **Ein Abbruch legt die Leitung nieder.** Das Issue verlangt es
		// wörtlich: „blockieren statt weiterzuspielen".
		//
		// **Aber erst, wenn der Grund draußen ist** (Zweitblick über Phase
		// 7, Befund 7): `Partienetz.Bediene` gibt nach einem Fehler nichts
		// mehr hinaus, und der Gast läse eine abgerissene Leitung statt der
		// Absage.
		if fehler := w.sitzung.Leere(ctx); fehler != nil {
			return fehler
		}

		return fmt.Errorf("%w: %s", domain.ErrConflict, weg.Grund)

	case domain.WegAbzug:
		// **Das Bild steckt schon in der Antwort.** Ein zweites wäre eine
		// zweite Wahrheit über denselben Tisch.
		return nil

	case domain.WegEreignisse:
		strom, fehler := w.nehmer.Nachreichen(
			ctx, w.partie, w.gast, folgeAus(weg.AbFolge))
		if fehler != nil {
			return fehler
		}

		if fehler := w.verkuenden(
			ctx, Befehlsausgang{Ereignisse: strom}); fehler != nil {
			return fehler
		}
	case domain.WegWeiter:
	}

	return w.tischbildSchicken(ctx)
}

// folgeAus macht aus einer Folgenummer eine Bestandsfolge.
//
// # Warum hier keine Grenze nach oben steht
//
// **Weil jede Zahl, die hier ankommt, aus einem `int64` stammt.** Der Stand
// des Wirts kommt über [folgeZahl] aus dem Bestand, und die nachzureichende
// Folge ist die des Gastes **plus eins** — und ein Gast, der weiter steht als
// der Wirt, kommt gar nicht bis hierher: [domain.Wiederaufnahme] bricht ab
// („das ist nicht dieselbe Partie").
//
// **Eine Grenze wäre ein Zweig, den kein Lauf betreten kann** (§10).
func folgeAus(folge uint64) int64 {
	// #nosec G115 -- der Grund steht darüber; gosec liest kein //nolint (§4)
	return int64(folge)
}
