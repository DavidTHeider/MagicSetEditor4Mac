package service

// **Eine Geste wird im Kern gedeutet** — P3
// (#2371),
// Befund C-03. Eine Karte von der Hand aufs Schlachtfeld zu ziehen stellte
// `objekt.schieben`: ohne Landzählung, ohne Stapel, und die Sitzangabe des
// Ziels fiel weg. Der Knopf *Als Land spielen* prüfte dagegen — zwei Wege,
// zwei Regelausgänge, und der bequeme war der ungeprüfte.
//
// # Ein Befehl für die Geste
//
// **`objekt.ziehen` nennt Objekt, Zone und Ziel-Sitz**, und der Kern macht
// daraus den Befehl, den das Profil verlangt — an einer Stelle für alle vier
// Flächen. Das Fenster stellt keinen eigenen Regelausleger (Vermerk 0126).
//
// # Was das Profil entscheidet
//
// **Im Sandkasten ist jede Geste ein Schieben** — der Tisch gehorcht dem
// Menschen ([domain.StrengeSandkasten]), und eine Partie ohne laufenden Zug
// ist einer. **Sonst wird gedeutet**: aus der Hand oder der Kommandozone auf
// Schlachtfeld oder Stapel heißt ein Land spielen (CR 116.2a, 305.1, 305.2)
// oder wirken (CR 601.2a, 903.8). Was die Deutung nicht kennt, schiebt
// „begleitet“; „streng“ weist es mit Grund ab. **Ausdrücklich frei**
// (`frei`, am Tisch ⌥) schiebt, außer unter strenger Prüfung.
//
// # Was immer gilt
//
// **Der Ziel-Sitz wird geprüft statt verworfen.** Bibliothek, Hand und
// Friedhof gehören dem Eigner (CR 400.3), und auf den Tisch eines anderen
// Sitzes kommt ein Objekt nur durch einen Kontrollwechsel (CR 110.2).
//
// Die Regeln sind nachgelesen in der Fassung vom 2026-09-25.

import (
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// BefehlZiehen deutet eine Geste: ein Objekt in eine Zone eines Sitzes.
const BefehlZiehen = "objekt.ziehen"

// Deutung ist der Befehl, den eine Geste meint.
type Deutung struct {
	Art      string
	Nutzlast map[string]any
}

// ziehen deutet die Geste und legt aus, was sie meint.
//
// **Über `auslegen` und nicht über `Auslegen`**: die Prüfungen davor —
// Absender, Beginn — sind gelaufen, und die zustandsbasierten Aktionen
// danach laufen ohnehin, wenn dieser Ruf zurückkommt.
func (s *Sandkasten) ziehen(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	deutung, fehler := s.deute(partie, befehl)
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	befehl.Art, befehl.Nutzlast = deutung.Art, deutung.Nutzlast

	return s.auslegen(partie, befehl)
}

// deute sagt, welchen Befehl eine Geste meint — oder warum keinen.
func (s *Sandkasten) deute(
	partie domain.Partie, befehl domain.Spielbefehl,
) (Deutung, error) {
	nutzlast := befehl.Nutzlast

	stelle, fehler := s.objektstelle(partie, befehl)
	if fehler != nil {
		return Deutung{}, fehler
	}

	objekt := partie.Objekte[stelle]

	ziel, fehler := pruefeGestenziel(partie, objekt, befehl)
	if fehler != nil {
		return Deutung{}, fehler
	}

	zone := ziel.zone
	schieben := Deutung{Art: BefehlSchieben, Nutzlast: map[string]any{
		"objekt": objekt.Kennung, "zone": string(zone),
	}}
	if stelle, steht := nutzlast["stelle"]; steht {
		schieben.Nutzlast["stelle"] = stelle
	}

	strenge := partie.Regelwerk.Strenge
	frei, _ := nutzlast["frei"].(bool) //nolint:errcheck // siehe [Sandkasten.objektstelle]

	switch {
	case frei && strenge == domain.StrengeStreng:
		return Deutung{}, fmt.Errorf(
			"%w: unter strenger Prüfung gibt es kein freies Schieben — die "+
				"Geste wird nach den Regeln gedeutet", domain.ErrValidation)
	case frei, strenge == domain.StrengeSandkasten, !prioritaetGilt(partie):
		return schieben, nil
	}

	gemeint, gedeutet, fehler := regelhandlung(partie, objekt, zone, nutzlast)
	switch {
	case fehler != nil:
		return Deutung{}, fehler
	case gedeutet:
		return gemeint, nil
	}

	if strenge == domain.StrengeStreng {
		return Deutung{}, fmt.Errorf(
			"%w: unter strenger Prüfung bewegt %q nur ein Befehl, der sagt, "+
				"warum — aus %q nach %q deutet die Geste nichts",
			domain.ErrValidation, objekt.Kennung, objekt.Zone, zone)
	}

	return schieben, nil
}

// gestenziel ist, wohin die Geste zeigt.
type gestenziel struct {
	zone domain.Zone
	sitz string
}

// pruefeGestenziel prüft Zone und Ziel-Sitz einer Geste.
func pruefeGestenziel(
	partie domain.Partie, objekt domain.Spielobjekt, befehl domain.Spielbefehl,
) (gestenziel, error) {
	kennung := objekt.Kennung
	nutzlast := befehl.Nutzlast

	zone, fehler := gemeinteZone(befehl)
	if fehler != nil {
		return gestenziel{}, fehler
	}

	sitz, _ := nutzlast["sitz"].(string) //nolint:errcheck // siehe [Sandkasten.objektstelle]

	switch {
	case sitz == "":
		return gestenziel{}, fmt.Errorf(
			"%w: die Geste nennt keinen Ziel-Sitz — wessen Zone gemeint ist, "+
				"sagt erst er", domain.ErrValidation)
	case partie.SitzVon(sitz) == nil || partie.Gegangen(sitz):
		return gestenziel{}, fmt.Errorf(
			"%w: den Ziel-Sitz %q gibt es an diesem Tisch nicht",
			domain.ErrValidation, sitz)
	}

	switch zone {
	case domain.ZoneBibliothek, domain.ZoneHand, domain.ZoneFriedhof:
		if sitz != objekt.Eigner {
			return gestenziel{}, fmt.Errorf(
				"%w: %q geht nur in die Zone %q seines Eigners %q, nicht in "+
					"die von %q (CR 400.3)",
				domain.ErrValidation, kennung, zone, objekt.Eigner, sitz)
		}

	case domain.ZoneTisch:
		if sitz != objekt.Beherrscher {
			return gestenziel{}, fmt.Errorf(
				"%w: auf den Tisch von %q kommt %q nur durch einen "+
					"Kontrollwechsel — „Beherrschen“ (CR 110.2)",
				domain.ErrValidation, sitz, kennung)
		}
	}

	return gestenziel{zone: zone, sitz: sitz}, nil
}

// regelhandlung deutet, was die Regeln für eine Geste kennen.
//
// **Drei Ausgänge**: gedeutet, ein Grund, warum die Geste nicht geht, oder
// nichts davon — dann entscheidet die Strenge.
func regelhandlung(
	partie domain.Partie, objekt domain.Spielobjekt, zone domain.Zone,
	nutzlast map[string]any,
) (Deutung, bool, error) {
	ins := zone == domain.ZoneTisch || zone == domain.ZoneStapel

	switch {
	case objekt.Zone == domain.ZoneHand && ins && istLand(partie, objekt):
		if zone == domain.ZoneStapel {
			return Deutung{}, false, fmt.Errorf(
				"%w: %q ist ein Land — es wird gespielt, nicht gewirkt "+
					"(CR 305.1)", domain.ErrValidation, objekt.Kennung)
		}

		return Deutung{Art: BefehlLandspielen, Nutzlast: map[string]any{
			"objekt": objekt.Kennung,
		}}, true, nil

	case (objekt.Zone == domain.ZoneHand || objekt.Zone == domain.ZoneKommando) && ins:
		gemeint := map[string]any{"objekt": objekt.Kennung}
		if ziele, stehen := nutzlast["ziele"]; stehen {
			gemeint["ziele"] = ziele
		}

		return Deutung{Art: BefehlWirken, Nutzlast: gemeint}, true, nil
	}

	return Deutung{}, false, nil
}
