package service

// **Was an einem Objekt zu tun ist** — P3
// (#2371),
// Befund C-10. `Erlaubt` sagt je Sitz, welche Befehlsarten er stellen darf;
// hier steht, **welche davon an diesem Objekt in dieser Zone gelten** und
// welche ein Klick und ein Doppelklick auslösen (P4). Im Kern und nicht im
// Fenster: sonst entstünde der zweite Regelausleger, den Vermerk 0126
// verwirft.
//
// # Woher die Zonen kommen
//
// **Wo der Kern eine Zone prüft, gilt seine Prüfung**: Zerstören, Opfern,
// Kämpfen, Beherrschen und Umwandeln auf dem Schlachtfeld, Neutralisieren
// auf dem Stapel, Exilieren von überall außer dem Exil
// (`schluesselhandlungen.go`, `kartengestalten.go`). **Wo er keine prüft**
// — die Befehle des Sandkastens —, gilt die Regel, an welchem Objekt die
// Handlung überhaupt etwas bedeutet, nachgelesen in der CR (Fassung
// 2026-09-25): getappt und enttappt werden bleibende Karten (701.26a/b),
// Schaden wird bleibenden Karten zugefügt (120.1), Angelegtes hängt auf dem
// Schlachtfeld (701.3a), Zähler liegen auf Objekten und vergehen mit einem
// Zonenwechsel (122.1, 122.2) — also in den offenen Zonen. Gewirkt wird aus
// der Hand und, für den Kommandeur, aus der Kommandozone (601.2, 903.8), ein
// Land gespielt aus der Hand (305.1).
//
// # Was ein Sitz nicht erfährt
//
// **An einem Objekt, dessen Karte der Sitz nicht kennt** — die eigene
// Bibliothek —, hängen nur Handlungen, die nicht von der Karte abhängen; ob
// dort ein Land liegt, verrät keine Auskunft. **Eine fremde Hand und eine
// fremde Bibliothek bekommen gar keine.**
//
// # Eigen und fremd
//
// **Primär und Zweit nur am eigenen Objekt.** Spielen, Wirken, Mana und
// Opfern setzen voraus, dass der Sitz das Objekt beherrscht (305.1, 601.2,
// 605.3a, 701.21a). Was eine Wirkung an einem fremden Objekt ausführt —
// Zerstören, Schaden, Zähler —, bleibt angeboten: der Tisch führt aus, was
// ein Zauber sagt, und das trifft auch fremde Objekte.

import (
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// handlungenJeZone ordnet jeder Zone die Befehlsarten zu, die dort an einem
// Objekt gelten — in der Reihenfolge, in der eine Fläche sie zeigt.
var handlungenJeZone = map[domain.Zone][]string{
	domain.ZoneHand: {
		BefehlLandspielen, BefehlWirken, BefehlExilieren,
		BefehlSchieben, BefehlZiehen,
	},
	domain.ZoneTisch: {
		BefehlTappen, BefehlEnttappen, BefehlLandmana, BefehlZaehler,
		BefehlSchaden, BefehlAnhaengen, BefehlLoesen, BefehlHaeufen,
		BefehlEnthaeufen, BefehlUmdrehen, BefehlUmwandeln, BefehlKaempfen,
		BefehlBeherrschen, BefehlKopieren, BefehlZerstoeren, BefehlOpfern,
		BefehlExilieren, BefehlSchieben, BefehlZiehen,
	},
	domain.ZoneStapel: {
		BefehlNeutralisieren, BefehlZaehler, BefehlUmdrehen,
		BefehlKopieren, BefehlExilieren, BefehlSchieben, BefehlZiehen,
	},
	domain.ZoneKommando: {
		BefehlWirken, BefehlZaehler, BefehlExilieren,
		BefehlSchieben, BefehlZiehen,
	},
	domain.ZoneFriedhof:   {BefehlZaehler, BefehlExilieren, BefehlSchieben, BefehlZiehen},
	domain.ZoneExil:       {BefehlZaehler, BefehlUmdrehen, BefehlSchieben, BefehlZiehen},
	domain.ZoneBibliothek: {BefehlExilieren, BefehlSchieben, BefehlZiehen},
}

// nurEigen sind die Befehle, die voraussetzen, dass der Sitz das Objekt
// beherrscht.
var nurEigen = []string{
	BefehlLandspielen, BefehlWirken, BefehlLandmana, BefehlOpfern,
}

// nachKarte sind die Befehle, deren Gelten von der Karte abhängt — an einem
// Objekt, dessen Karte der Sitz nicht kennt, stehen sie nie.
var nachKarte = []string{BefehlLandspielen, BefehlWirken, BefehlLandmana}

// Handlungsauskunft gibt je Objekt, was `fuer` daran gerade tun kann.
//
// **`erlaubt` ist die Antwort von [Sandkasten.Erlaubt]** für denselben
// Stand — Priorität, Beginn und Gehen stehen schon darin. Ein Objekt ohne
// eine einzige Handlung steht nicht in der Liste.
func Handlungsauskunft(
	partie domain.Partie, fuer string, erlaubt []string,
) []domain.Objekthandlungen {
	auskunft := []domain.Objekthandlungen{}

	for _, objekt := range partie.Objekte {
		eine := objekthandlungen(partie, objekt, fuer, erlaubt)
		if len(eine.Handlungen) > 0 {
			auskunft = append(auskunft, eine)
		}
	}

	return auskunft
}

func objekthandlungen(
	partie domain.Partie, objekt domain.Spielobjekt, fuer string,
	erlaubt []string,
) domain.Objekthandlungen {
	eigen := objekt.Beherrscher == fuer
	kennt := domain.KennenDieKarte(objekt, fuer)
	// **Unter der Sichtkennung** (#2480): die Auskunft geht im Tischbild an
	// den Gast.
	eine := domain.Objekthandlungen{Objekt: partie.Sichtkennung(objekt), Handlungen: []string{}}

	// **Eine fremde Hand und eine fremde Bibliothek bekommen nichts** — was
	// dort liegt, ist geheim, und auch die Frage danach wäre eine Auskunft.
	if !eigen && !kennt {
		return eine
	}

	land := kennt && istLand(partie, objekt)

	for _, art := range handlungenJeZone[objekt.Zone] {
		switch {
		case !slices.Contains(erlaubt, art),
			!eigen && slices.Contains(nurEigen, art),
			!kennt && slices.Contains(nachKarte, art),
			!gilt(partie, objekt, art, land):
			continue
		}

		eine.Handlungen = append(eine.Handlungen, art)
	}

	if eigen {
		eine.Primaer, eine.Zweit = primaerUndZweit(objekt, land, eine.Handlungen)
	}

	return eine
}

// gilt prüft, was über die Zone hinaus am Zustand des Objekts hängt.
func gilt(
	partie domain.Partie, objekt domain.Spielobjekt, art string, land bool,
) bool {
	switch art {
	case BefehlLandspielen:
		return land
	case BefehlWirken:
		return !land
	case BefehlTappen:
		return !objekt.Getappt
	case BefehlEnttappen:
		return objekt.Getappt
	case BefehlLandmana:
		return land && !objekt.Getappt && partie.HatGrundlandart(objekt)
	case BefehlLoesen:
		return objekt.HaengtAn != ""
	case BefehlEnthaeufen:
		return objekt.Haufen != ""
	}

	return true
}

// primaerUndZweit wählt aus den Handlungen die für Klick und Doppelklick.
//
// **Aus den Handlungen und nicht daneben** — was nicht angeboten ist, kann
// auch kein Klick auslösen.
func primaerUndZweit(
	objekt domain.Spielobjekt, land bool, handlungen []string,
) (primaer, zweit string) {
	erste := func(arten ...string) string {
		for _, art := range arten {
			if slices.Contains(handlungen, art) {
				return art
			}
		}

		return ""
	}

	switch objekt.Zone {
	case domain.ZoneHand:
		if land {
			return erste(BefehlLandspielen), ""
		}

		return erste(BefehlWirken), ""

	case domain.ZoneTisch:
		return erste(BefehlTappen, BefehlEnttappen), erste(BefehlLandmana)

	case domain.ZoneKommando:
		return erste(BefehlWirken), ""
	}

	return "", ""
}

// istLand sagt, ob die Karte eines Objekts ein Land ist — CR 305.
func istLand(partie domain.Partie, objekt domain.Spielobjekt) bool {
	eigenschaften, da := partie.Eigenschaften(objekt)

	return da && eigenschaften.Typzeile.Ist(domain.TypLand)
}
