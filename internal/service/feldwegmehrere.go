package service

import (
	"context"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// FelderVonAussen schreibt mehrere übernommene Werte — **alle oder keinen**.
//
// **Eine eigene Datei**, weil `feldweg.go` mit ihr über 400 Zeilen ginge (§2)
// — und weil der Schnitt einer Frage folgt: dort steht, **wohin** ein Wert
// gehört, hier, **wie viele** auf einmal geschrieben werden.
//
// # Wofür
//
// Eine Scryfall-Übernahme ändert selten ein Feld. Bis zum 2026-08-24 schrieb
// sie eines nach dem anderen, jedes in seiner eigenen Transaktion: scheiterte
// das dritte von fünf, standen zwei neu da und drei alt — und der Aufrufer
// bekam einen Fehler ohne die Angabe, welche Hälfte durch war. Angestrichen
// vom Zweitblick durch Codex am 2026-08-24.
//
// # Was sie zusagt, und was nicht
//
// **Alles oder nichts**, denn `storage.FelderSetzenMitVorwert` sagt es zu
// (#714).
//
// **Ein Verlaufsschritt für alle** (#2396): ein ⌘Z nimmt die Übernahme ganz
// zurück, so wie sie ganz geschrieben wurde. Bis dahin war es je Feld einer —
// und seit dieser Weg nachzieht, wer ein Feld liest, hätte schon eine einzige
// Bearbeitung mehrere ⌘Z gebraucht.
//
// **Und jeder Wert geht durch dieselbe Auflösung** wie ein einzeln
// geschriebener ([aufloesen]): ein Verbundfeld zerfällt in seine Teile, ein
// nicht speicherbares wird abgewiesen. Ein zweiter Regelsatz hier liefe
// auseinander, und es fiele niemandem auf.
//
// **Was sie nicht kann**: zwei Werte auf dasselbe Feld. Das wäre eine Eingabe,
// die sich selbst widerspricht — sie käme hier als zwei Zeilen an, und die
// zweite gewänne. Der Abgleich baut die Liste aus einem Vorschlag, und darin
// steht jedes Feld einmal.
func (f *Feldweg) FelderVonAussen(
	ctx context.Context, karteID int64, werte []domain.Value, herkunft string,
) (vorher string, hatte bool, fehler error) {
	ziele, stand, fehler := f.aufgeloest(ctx, karteID, werte)
	if fehler != nil {
		return "", false, fehler
	}

	return f.ablage.FelderSetzenVonAussen(ctx, karteID, ziele, herkunft, stand)
}

// aufgeloest zerlegt jeden Wert in das, was der Bestand trägt.
//
// **Einmal gelesen für alle Werte.** Ein Aufruf je Wert ginge bei einer
// Übernahme von zwölf Feldern zwölfmal ans Paket.
//
// **Ohne lesbare Vorlage gehen die Werte roh hinaus** — dieselbe Duldung wie
// beim einzelnen Schreiben, und aus demselben Grund: ein Set aus MTG JSON
// bliebe sonst unbearbeitbar, solange sein Spielpaket fehlt.
func (f *Feldweg) aufgeloest(
	ctx context.Context, karteID int64, werte []domain.Value,
) (ziele []domain.Value, stand string, fehler error) {
	karte, felder, lesefehler := f.karteUndFelder(ctx, karteID)
	if lesefehler != nil {
		// **Der Lesefehler wird hier nicht weitergereicht, und das ist die
		// Duldung**: ein Set aus MTG JSON bliebe sonst unbearbeitbar,
		// solange sein Spielpaket fehlt. Was hinausgeht, sind die rohen
		// Werte.
		return werte, "", nil //nolint:nilerr // siehe oben
	}

	ziele = make([]domain.Value, 0, len(werte))

	for _, wert := range werte {
		teile, fehler := aufloesen(felder, wert.Field, wert.Text)
		if fehler != nil {
			return nil, "", fehler
		}

		if teile == nil {
			teile = []domain.Value{wert}
		}

		ziele = append(ziele, teile...)
	}

	// **Gespeichert wird, was gerechnet ist** — wie beim einzelnen Feld
	// (#2396, `feldwegnachzug.go`), samt dem Stand, aus dem gerechnet wurde.
	ziele, stand = f.nachzug.nachgezogen(ctx, karte, ziele)

	return ziele, stand, nil
}

// FelderVonAussenMitHerkunft schreibt Werte **und ihre Herkunft in einem Zug**.
//
// # Warum derselbe Weg und nicht ein zweiter
//
// **Weil jeder Wert durch dieselbe Auflösung muss** wie ein gewöhnlich
// geschriebener ([aufloesen]): ein Verbundfeld zerfällt in seine Teile. Ein
// zweiter Regelsatz hier liefe auseinander, und es fiele niemandem auf — das
// steht schon über [Feldweg.FelderVonAussen] und gilt hier genauso.
//
// **Was hinzukommt, ist die Herkunft** (A11, #1450), und sie geht in
// derselben Transaktion hinaus wie die Werte.
// **`erwartet` ist der Stand, auf dem die Karte stehen muss** — leer heißt:
// ohne Bedingung (Befund 2 des Zweitblicks über Phase 5.3).
//
// **Die Auflösung geschieht davor und nicht darunter.** Ein Verbundfeld
// zerfällt in seine Teile, und der Stand der Karte ist derselbe, gleich
// welche Teile geschrieben werden.
func (f *Feldweg) FelderVonAussenMitHerkunft(
	ctx context.Context, karteID int64, werte []domain.Value,
	herkunft string, herkuenfte []domain.KIFeldherkunft,
	angaben map[string][]domain.KIBildparameter, erwartet string,
) (vorher string, hatte bool, fehler error) {
	ziele, stand, fehler := f.aufgeloest(ctx, karteID, werte)
	if fehler != nil {
		return "", false, fehler
	}

	erwartet, fehler = aufDemRechenstand(erwartet, stand)
	if fehler != nil {
		return "", false, fehler
	}

	return f.ablage.FelderSetzenMitHerkunft(
		ctx, karteID, ziele, herkunft, herkuenfte, angaben, erwartet)
}

// GruppeSetzenMitHerkunft schreibt eine atomare Gruppe (A13.1b, #1452) —
// **jede Karte durch dieselbe Auflösung** wie eine einzelne.
//
// Bis #2396 ging die Gruppe an diesem Weg vorbei: ein Verbundfeld landete als
// Zeile, die die Vorlage nie liest (#696), und kein Wertskript rechnete.
//
// **Aufgelöst wird alles vor dem ersten Schreiben**: weist ein Schritt ab,
// bleibt die ganze Gruppe ungeschrieben — die Zusage der Ablage, eine
// Transaktion für alle, gälte sonst nur für die Hälfte davor.
func (f *Feldweg) GruppeSetzenMitHerkunft(ctx context.Context, schritte []domain.Kartenschritt) error {
	aufgeloest := make([]domain.Kartenschritt, 0, len(schritte))

	for _, schritt := range schritte {
		ziele, stand, fehler := f.aufgeloest(ctx, schritt.KarteID, schritt.Werte)
		if fehler != nil {
			return fehler
		}

		schritt.Werte = ziele

		schritt.Erwartet, fehler = aufDemRechenstand(schritt.Erwartet, stand)
		if fehler != nil {
			return fehler
		}
		aufgeloest = append(aufgeloest, schritt)
	}

	return f.ablage.GruppeSetzenMitHerkunft(ctx, aufgeloest)
}

// aufDemRechenstand gibt den Stand, auf dem geschrieben wird: den, aus dem
// gerechnet wurde. **Nennt der Auftrag einen anderen, ist das ein Konflikt**
// (Nachlauf über das Gesamtreview über 0.9.0, Befund 3) — sonst prüfte die
// Transaktion den Stand des Auftrags, während aus einem anderen gerechnet
// wurde, und eine Rücknahme auf den alten ließe die neuen Werte durch. Ohne
// Rechnung gilt der Stand des Auftrags allein.
func aufDemRechenstand(auftrag, gerechnet string) (string, error) {
	switch {
	case gerechnet == "":
		return auftrag, nil
	case auftrag == "" || auftrag == gerechnet || nurKarte(auftrag, gerechnet):
		return gerechnet, nil
	default:
		return "", fmt.Errorf("%w: die Karte hat sich seit dem Auftrag geändert — "+
			"der Vorschlag gehört zu einem Stand, den es so nicht mehr gibt", domain.ErrConflict)
	}
}

// nurKarte sagt, ob ein Auftrag mit reinem Kartenstand zur Rechnung passt —
// **ein KI-Auftrag nennt keinen Setteil** (#2533), und verglichen wird dann
// der Kartenteil; den Setteil hält die Transaktion.
func nurKarte(auftrag, gerechnet string) bool {
	karte, set := domain.StandTeile(auftrag)
	gerechneteKarte, _ := domain.StandTeile(gerechnet)

	return set == "" && karte == gerechneteKarte
}
