package app

// **Die Herkunft eines Feldes lesen** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Warum sie gelesen werden muss
//
// **Weil eine Herkunft, die niemand sieht, keine ist.** Das Issue verlangt
// sie „bei erster Übernahme verpflichtend" — und im selben Atemzug, dass im
// Inspector steht, woher ein Wert kommt. Geschrieben wird sie in der
// Transaktion der Übernahme ([storage.Kartenschreiber.FelderSetzenMitHerkunft]);
// hier steht der Weg zurück.
//
// # Und sie sagt, ob der Wert noch derselbe ist
//
// **Weil manuelle Nachbearbeitung erlaubt ist** (das Issue sagt es
// ausdrücklich). Wer den Wert danach ändert, darf das — und dann soll die
// Herkunft nicht behaupten, der Inhalt stamme von der KI. Der Abdruck sagt
// es: stimmt er nicht mehr, steht `Verändert` auf `true`.

import (
	"context"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// KIFeldherkunftDTO ist eine Herkunft, wie das Fenster sie sieht.
type KIFeldherkunftDTO struct {
	Feld     string `json:"feld"`
	Anbieter string `json:"anbieter"`
	Modell   string `json:"modell"`

	// Aufgabe ist, was ein Mensch getippt hat — **nicht** der Rohprompt.
	Aufgabe string `json:"aufgabe"`

	// Auftrag bindet die Felder **einer** Übernahme zusammen.
	Auftrag string `json:"auftrag"`

	// Zeitpunkt ist RFC 3339.
	Zeitpunkt string `json:"zeitpunkt"`

	// Veraendert sagt, ob am Feld seither von Hand gearbeitet wurde.
	Veraendert bool `json:"veraendert"`
}

// KIFeldherkunftAntwort trägt sie samt Fehler.
type KIFeldherkunftAntwort struct {
	Herkunft []KIFeldherkunftDTO `json:"herkunft"`

	Fehler *Fehler `json:"fehler"`
}

// KIFeldherkunftLeser ist, was diese Naht vom Bestand braucht.
type KIFeldherkunftLeser interface {
	Lesen(ctx context.Context, karteID int64) ([]domain.KIFeldherkunft, error)

	// Bildangaben gibt, was über ein übernommenes Bild zusätzlich feststeht
	// (A12, #1451) — leer bei jedem anderen Feld.
	Bildangaben(ctx context.Context, karteID int64, feld string) (
		[]domain.KIBildparameter, error)
}

// herkunftsleser gibt den Leser über dem offenen Bestand.
//
// # Warum kein Feld an der Anwendung
//
// **Weil er nichts hält.** Er liest über die **laufende** Verbindung, und die
// wechselt, wenn jemand einen anderen Bestand öffnet. Ein einmal gesetztes
// Feld zeigte danach auf die alte Datei — dieselbe Überlegung wie bei
// [App.probenablage].
func (a *App) herkunftsleser() KIFeldherkunftLeser {
	db := a.laufendeVerbindung()
	if db == nil {
		return nil
	}

	return storage.NeueFeldherkunft(db)
}

// KIFeldherkunft gibt, woher die Werte einer Karte stammen.
//
// **Eine Auskunft und keine Frage nach draußen**: hier geht kein Aufruf an
// einen Anbieter, und es entsteht kein Auftrag. Wer sie liest, liest den
// Bestand.
func (a *App) KIFeldherkunft(
	ctx context.Context, karteID int64,
) KIFeldherkunftAntwort {
	ctx = a.imProfil(ctx)

	// **Auch der Fehlerweg gibt eine Liste und kein `null`.** Was hier fehlt,
	// fehlt im Fenster als `herkunft.length` — und `null.length` hält die
	// Seite an (A9, dort einmal passiert).
	leer := KIFeldherkunftAntwort{Herkunft: []KIFeldherkunftDTO{}}

	leser := a.herkunftsleser()
	if leser == nil || a.kartenleser == nil {
		leer.Fehler = Uebersetze(ctx, fehlenderBestand())

		return leer
	}

	gelesen, fehler := leser.Lesen(ctx, karteID)
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	zeile, fehler := a.kartenleser.KarteLesen(ctx, karteID)
	if fehler != nil {
		leer.Fehler = Uebersetze(ctx, fehler)

		return leer
	}

	return KIFeldherkunftAntwort{Herkunft: alsHerkunftsDTO(gelesen, zeile)}
}

// alsHerkunftsDTO macht aus den Zeilen die Liste für das Fenster.
func alsHerkunftsDTO(
	gelesen []domain.KIFeldherkunft, zeile *domain.CardRow,
) []KIFeldherkunftDTO {
	heute := make(map[string]string, len(zeile.Values))
	for _, wert := range zeile.Values {
		heute[wert.Field] = wert.Text
	}

	liste := make([]KIFeldherkunftDTO, 0, len(gelesen))

	for _, herkunft := range gelesen {
		liste = append(liste, KIFeldherkunftDTO{
			Feld: herkunft.Feld, Anbieter: herkunft.Anbieter,
			Modell: herkunft.Modell, Aufgabe: herkunft.Aufgabe,
			Auftrag:   herkunft.Auftrag,
			Zeitpunkt: herkunft.Zeitpunkt.Format(time.RFC3339),

			// **Verglichen wird gegen den Text, der im Feld steht** — genau
			// so, wie der Abdruck bei der Übernahme entstanden ist
			// ([kikarten.FelderSetzen]).
			Veraendert: domain.KIInhaltsbezug(
				[]string{heute[herkunft.Feld]}) != herkunft.Bezug,
		})
	}

	return liste
}
