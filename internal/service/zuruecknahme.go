package service

// **Begrenztes Rückgängig** — P8
// (#2376),
// Befund C-07. Die Regeln stehen in Vermerk 0197; hier steht, wie der Kern
// sie einlöst.
//
// # Was zurückgenommen wird
//
// **Der letzte Befehl der Partie, und nur, wenn er dem Absender gehört** und
// auf der Liste steht — Tappen, Enttappen, Leben, Vergiftung: Befehle, deren
// Wirkung ganz im öffentlichen Ereignis steht und sich ohne Rest umkehren
// lässt. **Keine Regel des Spiels**: CR 733.1 (Fassung 2026-09-25) gilt für
// regelwidrige Handlungen, und eine Rücknahme regelgerechter Züge stützt
// sich nicht darauf. Sie ist eine Absprache dieses Tisches.
//
// # Wie
//
// **Angehängt, nicht gelöscht**: die Rücknahme schreibt ein eigenes Ereignis
// mit der Folge, die sie meint, und danach das Ereignis der Umkehr — die
// durch **dieselbe Auslegung** geht wie ein Befehl von Hand. Eine Abweisung
// schreibt nichts.

import (
	"context"
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// BefehlZuruecknehmen nimmt den letzten eigenen Zug zurück — P8 (#2376).
const BefehlZuruecknehmen = "spieler.zuruecknehmen"

// Leitungsbefehle sind die Spielbefehle, die nicht der Sandkasten auslegt,
// sondern die Leitung — sie meinen den Strom und nicht nur den Stand.
func Leitungsbefehle() []string {
	return []string{BefehlZuruecknehmen}
}

// zufaellig sind Befehle, deren zweiter Versuch ein neuer Wurf wäre.
var zufaellig = []string{
	BefehlMuenzwurf, BefehlWuerfeln, BefehlMischen, BefehlStarthand,
	BefehlMulligan,
}

// aufdeckend sind Befehle, nach denen etwas gesehen oder bewegt ist.
var aufdeckend = []string{
	BefehlUmdrehen, BefehlSchieben, BefehlZiehen, BefehlWirken,
	BefehlLandspielen, BefehlExilieren, BefehlOpfern, BefehlZerstoeren,
	BefehlNeutralisieren, BefehlMahlen, BefehlBlick, BefehlHellsehen,
	BefehlUeberwachen, BefehlUnterlegen, BefehlHeimholen,
}

// abschliessend sind Befehle, die das Wort weitergeben oder etwas beenden.
var abschliessend = []string{
	BefehlPassen, BefehlAbkuerzen, BefehlZugende, BefehlBehalten,
	BefehlAufgeben, BefehlAngreifer, BefehlBlocker, BefehlReihung,
	BefehlLegende,
}

// umkehrVon sagt, welcher Befehl den letzten zurücknimmt — und welche Folge
// er meint. `letzte` sind die Ereignisse des letzten Befehls der Partie.
func umkehrVon(
	letzte []domain.Ereignis, absender string,
) (domain.Spielbefehl, int64, error) {
	if len(letzte) == 0 {
		return domain.Spielbefehl{}, 0, fmt.Errorf(
			"%w: an diesem Tisch ist noch nichts geschehen, das sich "+
				"zurücknehmen ließe", domain.ErrValidation)
	}

	erstes := letzte[0]
	sitz, _ := erstes.Oeffentlich["sitz"].(string) //nolint:errcheck // leer heißt: niemandes

	switch {
	case sitz != absender:
		return domain.Spielbefehl{}, 0, fmt.Errorf(
			"%w: seit deinem letzten Zug hat ein anderer Sitz gehandelt — "+
				"zurücknehmen lässt sich nur der eigene letzte Zug",
			domain.ErrConflict)
	case erstes.Art == BefehlZuruecknehmen:
		return domain.Spielbefehl{}, 0, fmt.Errorf(
			"%w: eine Rücknahme nimmt man nicht zurück", domain.ErrValidation)
	case len(letzte) > 1:
		return domain.Spielbefehl{}, 0, fmt.Errorf(
			"%w: dieser Zug hatte mehr als eine Folge — er lässt sich "+
				"nicht zurücknehmen", domain.ErrValidation)

	// **Wer etwas nur für sich erfuhr, hat es erfahren** — gleich, welcher
	// Befehl es war.
	case erstes.Erfahren():
		return domain.Spielbefehl{}, 0, fmt.Errorf("%w: %s",
			domain.ErrValidation, gesehen)
	}

	umkehr, fehler := umkehrbefehl(erstes)

	return umkehr, erstes.Folge, fehler
}

// umkehrbefehl gibt den Befehl, der ein Ereignis der Liste umkehrt.
func umkehrbefehl(ereignis domain.Ereignis) (domain.Spielbefehl, error) {
	switch ereignis.Art {
	case BefehlTappen, BefehlEnttappen:
		gegen := map[string]string{
			BefehlTappen: BefehlEnttappen, BefehlEnttappen: BefehlTappen,
		}[ereignis.Art]

		return domain.Spielbefehl{
			Art:      gegen,
			Nutzlast: map[string]any{"objekt": ereignis.Oeffentlich["objekt"]},
		}, nil
	case BefehlLeben, BefehlVergiftung:
		um, _ := ereignis.Oeffentlich["um"].(float64) //nolint:errcheck // null kehrt null um

		return domain.Spielbefehl{
			Art: ereignis.Art, Nutzlast: map[string]any{"um": -um},
		}, nil
	}

	return domain.Spielbefehl{}, fmt.Errorf("%w: %s",
		domain.ErrValidation, warumNicht(ereignis))
}

// gesehen ist der Grund für alles, was aufgedeckt oder bewegt wurde.
const gesehen = "was aufgedeckt oder bewegt wurde, bleibt gesehen — dieser " +
	"Zug lässt sich nicht zurücknehmen"

// warumNicht nennt den Grund, aus dem ein Ereignis nicht auf der Liste steht.
func warumNicht(ereignis domain.Ereignis) string {
	switch {
	case slices.Contains(zufaellig, ereignis.Art):
		return "ein Zufall lässt sich nicht zurücknehmen — ein zweiter " +
			"Versuch wäre ein neuer Wurf"
	case slices.Contains(aufdeckend, ereignis.Art):
		return gesehen
	case slices.Contains(abschliessend, ereignis.Art):
		return "dieser Zug hat das Wort weitergegeben oder etwas " +
			"abgeschlossen — er lässt sich nicht zurücknehmen"
	}

	return fmt.Sprintf("%q steht nicht auf der Liste der Züge, die sich "+
		"zurücknehmen lassen", ereignis.Art)
}

// zuruecknehmen legt eine Rücknahme aus und schreibt sie.
func (l *Partieleitung) zuruecknehmen(
	ctx context.Context, partie domain.Partie, befehl domain.Spielbefehl,
) (Befehlsausgang, error) {
	letzte, fehler := l.partien.LetzterBefehl(ctx, partie.ID)
	if fehler != nil {
		return Befehlsausgang{}, fehler
	}

	umkehr, folge, fehler := umkehrVon(letzte, befehl.Absender)
	if fehler != nil {
		return Befehlsausgang{}, fehler
	}

	umkehr.Kennung = befehl.Kennung
	umkehr.PartieID = befehl.PartieID
	umkehr.Absender = befehl.Absender
	umkehr.Gestellt = befehl.Gestellt

	gefolgt, danach, fehler := l.ausleger.Auslegen(partie, umkehr)
	if fehler != nil {
		return Befehlsausgang{}, fehler
	}

	ereignisse := append([]domain.Ereignis{{
		Art: BefehlZuruecknehmen,
		Oeffentlich: map[string]any{
			"sitz": befehl.Absender, "folge": float64(folge),
			"befehl": letzte[0].Art,
		},
	}}, gefolgt...)

	return l.schreiben(ctx, partie, befehl.Kennung, danach, ereignisse)
}
