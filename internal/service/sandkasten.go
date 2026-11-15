// **Der Sandkasten** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was er ist und was er ausdrücklich nicht ist
//
// **Er ist der Regelausleger, der nichts über Magic weiß.** Er bewegt
// Objekte, tappt sie, dreht sie um und legt Zähler darauf — genau das, was
// zwei Menschen auch mit echten Karten auf einem Tisch tun. Was in welcher
// Reihenfolge erlaubt ist, entscheidet in ihm **niemand**.
//
// **Und das ist keine Lücke, sondern die Zusage aus #1458**: „Sandbox lockert
// keine Objekt-, Sichtbarkeits-, Integritäts- oder Rechteinvarianten." Was er
// lockert, sind **Regelprüfungen** — wer hier eine Karte auf den Tisch legt,
// die dort nicht hingehört, darf das; **wessen Hand er dabei sieht, ändert
// sich nicht**.
//
// # Warum er trotzdem prüft
//
// **Weil ein Befehl auf ein Objekt zeigt, das es geben muss.** Ein Tappen
// ohne Objekt ist kein gelockertes Spiel, sondern ein Befehl ins Leere — und
// ein Ereignis darüber wäre eine Zeile im Strom, die niemand erklären kann.

package service

import (
	"fmt"
	"slices"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Die Befehlsarten, die der Sandkasten kennt.
//
// **Als Konstanten und nicht als Zeichenketten im Aufrufer** — wer sie
// tippte, bekäme beim Vertippen ein „unbekannter Befehl" und suchte im
// falschen Paket.
const (
	BefehlTappen    = "objekt.tappen"
	BefehlEnttappen = "objekt.enttappen"
	BefehlUmdrehen  = "objekt.umdrehen"
	BefehlSchieben  = "objekt.schieben"
	BefehlZaehler   = "objekt.zaehler"
	BefehlAnhaengen = "objekt.anhaengen"
	BefehlLoesen    = "objekt.loesen"

	// **Der Haufen** — CR 400.5, für PLAY-03.AK-03 und AK-04
	// (#1756).
	// Eine Anordnung und keine Regelbewegung: wer sie fürs Regelwerk
	// hielte, suchte sie in CR 400.1 unter den Zonen.
	BefehlHaeufen    = "objekt.haeufen"
	BefehlEnthaeufen = "objekt.enthaeufen"
)

// erlaubtOhnePrioritaet gibt die Befehle, die auch ohne sie gehen.
func erlaubtOhnePrioritaet() []string {
	frei := []string{}

	for _, art := range Sandkastenbefehle() {
		if !brauchtPrioritaet(art) {
			frei = append(frei, art)
		}
	}

	return frei
}

// Sandkastenbefehle gibt sie alle.
//
// **Ohne Zahl im Satz**: eine Zahl in einem Kommentar steht unter keinem
// Wächter (§16). Gezählt wird sie im Lauf daneben.
//
// **Fünf kamen mit G9 dazu** (#1592): passen, abwerfen, Leben,
// Vergiftungsmarken und markierter Schaden; **aufgeben** mit M17 (#1618),
// **abkürzen und Zugende** mit M1 (#1603), **Mana und das Grundland** mit M3
// (#1605),
// **Wirken und Landspielen** mit M4
// (#1606),
// **Angreifen und Blocken** mit M6
// (#1608),
// **acht Handlungen mit Schlüsselwort** mit M7
// (#1609),
// **Münzwurf und Würfel** mit M15
// (#1617),
// **Häufen und Enthäufen** mit
// #1756,
// **Starthand, Mulligan, Unterlegen und Behalten** mit
// #1839.
// Sie stehen hier und nicht daneben, weil sie durch denselben Weg gehen — und
// damit an allen vier Flächen zugleich da sind (§16).
func Sandkastenbefehle() []string {
	befehle := append([]string{
		BefehlTappen, BefehlEnttappen, BefehlUmdrehen, BefehlSchieben,
		BefehlZiehen,
		BefehlZaehler, BefehlAnhaengen, BefehlLoesen,
		BefehlHaeufen, BefehlEnthaeufen,
		BefehlBlick, BefehlMischen,
	}, Zugbefehle()...)

	return append(befehle, Eroeffnungsbefehle()...)
}

// Sandkasten legt Befehle aus, ohne Regeln zu kennen.
type Sandkasten struct {
	// zufall ist, woher die Mischung kommt — eingesetzt, damit ein Lauf
	// zweimal dasselbe Brett sieht (§3, und #1458 wörtlich).
	zufall Zufallsquelle

	// skript liest die Kartenskripte — **beim Verrechnen** (CR 608.2c,
	// #1625).
	//
	// **Einer je Sandkasten und nicht je Ruf**: er hält nur seine Frist, und
	// die ist für jede Partie dieselbe.
	skript *Kartenskript
}

// NeuerSandkasten baut ihn über einer Zufallsquelle.
func NeuerSandkasten(zufall Zufallsquelle) *Sandkasten {
	return &Sandkasten{zufall: zufall, skript: NeuesKartenskript()}
}

// Auslegen gibt die Ereignisse und den Zustand danach.
func (s *Sandkasten) Auslegen(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	// **Der Deuter wird hier eingesetzt und nirgends sonst**
	// (#1625).
	//
	// **Hier, weil jeder Befehl hier vorbeikommt** — auch `passen`, und
	// genau in ihm verrechnet der Stapel (CR 117.3b). Eine Partie kommt
	// ohne ihn aus dem Bestand; wer ihn an einer zweiten Stelle setzte,
	// hätte eine, die ihn manchmal hat.
	partie = partie.MitDeuter(skriptdeuter{skript: s.skript})

	// **Wer gegangen ist, stellt nichts mehr** — CR 800.4a (Fassung
	// 2026-09-25), #1979. Hier aus demselben Grund wie der Beginn darunter:
	// jede Fläche kommt an diesem Ausleger vorbei.
	if fehler := partie.PruefeAbsender(befehl.Absender); fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	// **Vor dem ersten Zug nur der Beginn** — CR 103.8 (Fassung 2026-09-25).
	// Hier und nicht erst in `Erlaubt`: das sagt nur, was angeboten wird, und
	// jede Fläche kommt an diesem Ausleger vorbei (§16).
	if beginnHaelt(partie) && !slices.Contains(waehrendDesBeginns(), befehl.Art) {
		return nil, domain.Partie{}, fmt.Errorf(
			"%w: vor dem ersten Zug steht der Beginn — erst die Starthand "+
				"und ihr Mulligan, dann %q (CR 103.8)",
			domain.ErrValidation, befehl.Art)
	}

	ereignisse, danach, fehler := s.auslegen(partie, befehl)
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	// **Nach einem Schritt des Beginns wird nicht geprüft** (#1839) —
	// CR 704.3 prüft, wann immer jemand Priorität bekäme, und während des
	// Beginns bekommt sie niemand. Wer dabei ins Leere gezogen hat, verliert
	// deshalb beim nächsten Mal und nicht mitten im Mulligan: so sagt es
	// CR 104.3c (Fassung 2026-09-25).
	if istEroeffnungsbefehl(befehl.Art) {
		return ereignisse, danach, nil
	}

	// **Und danach laufen die zustandsbasierten Aktionen** — CR 704.3 (M5,
	// #1607).
	//
	// # Warum hier und nicht je Befehl
	//
	// **Weil CR 704.3 keinen Befehl kennt.** Sie sagt: wann immer jemand
	// Priorität bekäme. Nach einem Befehl behält der Handelnde sie — also
	// ist eine Prüfung fällig. Sie hier zu ziehen heißt, dass ein **neuer**
	// Befehl sie mitbringt, statt sie zu vergessen; bis M5 stand sie je
	// Befehl, und zwar an genau einem.
	//
	// **Sie gilt nur, wo das Profil sie trägt.** Eine Partie aus G1 bis G8
	// ist ein Sandkasten: dort setzt nichts etwas durch, und eine Kreatur
	// mit null Widerstandskraft bleibt liegen, wenn jemand sie so hinlegt
	// (§15, Punkt 4 — alte Partien laufen unverändert weiter).
	return ereignisse, s.zustandsgepruefft(danach), nil
}

// zustandsgepruefft wendet die zustandsbasierten Aktionen an, wo das Profil
// sie trägt.
func (s *Sandkasten) zustandsgepruefft(partie domain.Partie) domain.Partie {
	if partie.Profil.Matrix.Grad(domain.BereichZustandsbasiert) !=
		domain.UnterstuetzungAutomatisch {
		return partie
	}

	return partie.Zustandspruefung()
}

// auslegen legt den Befehl aus — ohne die Prüfung danach.
func (s *Sandkasten) auslegen(
	partie domain.Partie, befehl domain.Spielbefehl,
) ([]domain.Ereignis, domain.Partie, error) {
	// **Eine Geste wird zuerst gedeutet** (P3, #2371) — was sie meint, geht
	// danach denselben Weg wie jeder andere Befehl.
	if befehl.Art == BefehlZiehen {
		return s.ziehen(partie, befehl)
	}

	// **Zuerst die Zonenbefehle.** Sie zeigen auf keine Objektkennung, und
	// ein `objektstelle` davor suchte ein Objekt, das gar nicht gemeint ist.
	if istZonenbefehl(befehl.Art) {
		return s.zonenbefehl(partie, befehl)
	}

	// **Und dann die, die einen Sitz meinen statt eines Objekts** (G9,
	// #1592): passen, abwerfen, Leben und Vergiftungsmarken.
	if istZugbefehl(befehl.Art) {
		return s.zugbefehl(partie, befehl)
	}

	// **Und das Grundland vor der Objektsuche** (M3,
	// #1605).
	//
	// **Es meint ein Objekt und wirkt trotzdem auf die Partie**: es tappt das
	// Land **und** legt Mana in den Vorrat eines Sitzes. `wirken` bekommt nur
	// das Objekt; wer diesen Befehl dort hineinlegte, hätte die Hälfte der
	// Wirkung an einer Stelle, an der sie niemand vermutet.
	if befehl.Art == BefehlLandmana {
		return s.landmanabefehl(partie, befehl)
	}

	// **Und die zwei aus M4 ebenso** (#1606): beide meinen ein Objekt und
	// wirken auf die Partie — der Zauber nimmt Mana aus einem Vorrat, das
	// Land zählt am Zug mit.
	if befehl.Art == BefehlWirken || befehl.Art == BefehlLandspielen {
		return s.wirkbefehl(partie, befehl)
	}

	// **Und die zwei aus M6** (#1608): sie meinen gar kein einzelnes
	// Objekt, sondern eine **Liste von Paaren**.
	if befehl.Art == BefehlAngreifer || befehl.Art == BefehlBlocker {
		return s.kampfbefehl(partie, befehl)
	}

	// **Und die acht aus M7** (#1609): sie bewegen Karten zwischen Zonen
	// und wirken damit auf die ganze Partie.
	if istHandlung(befehl.Art) {
		return s.handlungsbefehl(partie, befehl)
	}

	// **Und die vier aus M12** (#1614):
	// zwei erzeugen ein Objekt, das es noch nicht gibt, und zwei ändern
	// eines, das nicht über `wirken` geht.
	if istGestaltbefehl(befehl.Art) {
		return s.gestaltbefehl(partie, befehl)
	}

	// **Und die zwei aus M15** (#1617): sie ändern nichts am Tisch und
	// hinterlassen nur ein Ergebnis im Strom.
	if befehl.Art == BefehlMuenzwurf || befehl.Art == BefehlWuerfeln {
		return s.zufallsbefehl(partie, befehl)
	}

	// **Und die vier des Beginns** (#1839): sie meinen einen Sitz, seine
	// Bibliothek und seine Hand.
	if istEroeffnungsbefehl(befehl.Art) {
		return s.eroeffnungsbefehl(partie, befehl)
	}

	stelle, fehler := s.objektstelle(partie, befehl)
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	danach := s.kopie(partie)

	geschehen, fehler := s.wirken(&danach, stelle, befehl)
	if fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	objekt := &danach.Objekte[stelle]

	if fehler := objekt.Pruefe(); fehler != nil {
		return nil, domain.Partie{}, fehler
	}

	geschehen["objekt"] = objekt.Kennung
	geschehen["sitz"] = befehl.Absender

	return []domain.Ereignis{{
		Art:         befehl.Art,
		Oeffentlich: geschehen,
	}}, danach, nil
}

// objektstelle findet das gemeinte Objekt.
func (s *Sandkasten) objektstelle(
	partie domain.Partie, befehl domain.Spielbefehl,
) (int, error) {
	kennung, _ := befehl.Nutzlast["objekt"].(string) //nolint:errcheck // siehe darunter

	// **Die misslungene Zusicherung ist hier kein eigener Fall.** Was nicht
	// als Zeichenkette hereinkam, ist die leere Zeichenkette — und „kein
	// Objekt genannt" ist derselbe Satz für „gar nichts" und für „eine
	// Zahl". Zwei Meldungen für denselben Irrtum wären zwei Wege zu
	// derselben Stelle.
	if kennung == "" {
		return 0, fmt.Errorf(
			"%w: der Befehl %q nennt kein Objekt", domain.ErrValidation,
			befehl.Art)
	}

	for i := range partie.Objekte {
		if partie.Objekte[i].Kennung == kennung {
			return i, nil
		}
	}

	return 0, fmt.Errorf(
		"%w: in dieser Partie liegt kein Objekt %q", domain.ErrNotFound,
		kennung)
}

// kopie gibt eine Partie, deren Objekte sich ändern lassen.
//
// **Ohne sie schriebe der Ausleger in den Stand, den der Aufrufer noch
// hält** — und ein Fehlschlag danach ließe eine halb geänderte Partie zurück.
func (s *Sandkasten) kopie(partie domain.Partie) domain.Partie {
	objekte := make([]domain.Spielobjekt, len(partie.Objekte))

	for i, objekt := range partie.Objekte {
		objekte[i] = objekt

		if objekt.Zaehler != nil {
			objekte[i].Zaehler = make(map[string]int, len(objekt.Zaehler))

			for art, menge := range objekt.Zaehler {
				objekte[i].Zaehler[art] = menge
			}
		}
	}

	partie.Objekte = objekte

	// **Und die Sitze ebenso** — Befund beim Bauen von M3
	// (#1605),
	// gemessen am 2026-09-17.
	//
	// **Hier stand nur die Objektreihe**, und `markenbefehl` schrieb mit
	// `danach.Sitze[stelle].Leben += um` in den Speicher der Partie, die der
	// Aufrufer hält: ein `spieler.leben` mit −5 ließ die **übergebene**
	// Partie bei 15 statt bei 20 stehen. Auf dem Erfolgsweg fiel es nicht
	// auf, weil der alte Stand ohnehin weggeworfen wird — auf dem Fehlerweg
	// kommt ein Fehler zurück und der Stand ist trotzdem verändert (§10).
	//
	// **Seit `M3` trägt ein Sitz eine Tabelle** (`Manavorrat`), und eine
	// geteilte Tabelle reicht weiter als eine geteilte Reihe.
	//
	// **Was hier ausdrücklich nicht kopiert wird, ist [domain.Spielobjekt.Felder]**
	// (M12, #1614).
	// Die gedruckten Eigenschaften eines Spielsteins werden **einmal**
	// geschrieben — beim Erzeugen, und dort aus einer eigenen Kopie
	// ([domain.Partie.Spielstein]). Ein Zweig, der sie hier noch einmal
	// kopierte, wäre ein Versprechen, das niemand einlöst (§10): kein
	// Befehl schreibt hinein.
	//
	// **Das ändert sich mit `M10`** (#1612):
	// wo Schichten Eigenschaften verändern, wird geschrieben — und dann
	// gehört die Kopie hierher.
	sitze := make([]domain.Sitz, len(partie.Sitze))

	for i, sitz := range partie.Sitze {
		sitze[i] = sitz
		sitze[i].Manavorrat = sitz.Manavorrat.Kopie()
	}

	partie.Sitze = sitze

	return partie
}
