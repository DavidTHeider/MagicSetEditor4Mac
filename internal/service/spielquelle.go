package service

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Das Spielpaket als Quelle des Planbauers.
//
// **Von `stilquelle.go` getrennt**, als sie mit den Schlüsselwörtern über 400
// Zeilen wuchs (§2). Der Schnitt folgt der Frage: dort, wo ein **Stil**
// herkommt, hier, was das **Spiel** dazu beisteuert.

// Spielangaben ist, was ein Spielpaket dem Planbauer mitgibt.
//
// **Ein Verbund und keine fünf Rückgabewerte.** Er ist von vier auf fünf
// gewachsen, als die Schlüsselwörter dazukamen (#1061); ab da liest niemand
// mehr, welcher Wert welcher ist.
type Spielangaben struct {
	Skript          string
	Modell          domain.Game
	Symbolfassungen []msepaket.Symbolfassung

	// Schluesselwoerter ist die **vorbereitete** Wortliste — einmal je
	// Spielpaket, nicht je Karte.
	Schluesselwoerter Wortliste

	// Wortlisten sind die `word list:`-Blöcke — die Auswahllisten der
	// Textfelder (#1074).
	Wortlisten []msepaket.Wortliste

	// Rohwoerter sind die Schlüsselwörter des Spiels, **unvorbereitet**
	// (#1062).
	//
	// **Neben der vorbereiteten Liste und nicht statt ihrer.** Die
	// vorbereitete rechnet auf der Karte; diese hier wird **angezeigt**, und
	// dafür braucht es die Felder, wie sie in der Datei stehen — samt derer,
	// die kein gültiges Muster haben und deshalb gar nicht vorbereitet werden.
	Rohwoerter []msepaket.Schluesselwort

	// Modi sind die `keyword mode:`-Blöcke — `core`, `expert`, `custom`.
	Modi []msepaket.Schluesselwortmodus

	// Parameterarten braucht, wer ein Muster auf Gültigkeit prüft.
	Parameterarten []msepaket.Schluesselwortparameterart

	// Kartenskripte sind die `add cards script:`-Blöcke
	// (#1337).
	//
	// **Hier und nicht in einem eigenen Ruf**, weil das Spielpaket ohnehin
	// offen ist: ein zweiter Weg dorthin öffnete es ein zweites Mal, samt
	// seinen 19.000 Zeilen Skript.
	Kartenskripte []Kartenskriptangabe

	// Auszaehldimensionen sind die `statistics dimension:`-Blöcke — wonach
	// dieses Spiel auszählen lässt
	// (#1338).
	Auszaehldimensionen []Auszaehlangabe

	// Packs sind die `pack type:`-Blöcke — die Booster des Spiels
	// (#1072).
	Packs []msepaket.Pack
}

// spielskript liest das `init script:` des Spielpakets.
//
// **Kein Spielpaket ist ein Befund, kein Abbruch.** Ein Stil steht auch ohne
// es; was ohne seine Funktionen nicht ausrechenbar ist, benennt der Planbauer
// Feld für Feld.
//
// **Das ganze Spielmodell und nicht seine Kartenfelder**: es trägt seit #321
// auch die Setfelder, und eine vierte Rückgabe daneben wäre eine Liste, an die
// jede weitere Erkenntnis eine fünfte hängte.
func (s *Stilspeicher) spielskript(spiel string, leser msepaket.Paketquelle) (Spielangaben, []domain.Befund) {
	if spiel == "" {
		return Spielangaben{}, nil
	}

	paketname := spiel + "." + string(msepaket.ArtSpiel)
	var oeffnungsbefunde []domain.Befund

	for _, wurzel := range s.wurzeln {
		pfad := filepath.Join(wurzel, paketname)
		paket, err := msepaket.Oeffne(pfad)
		if err != nil {
			if befund := paketOeffnungsbefund(pfad, "Spielpaket", err); befund != nil {
				oeffnungsbefunde = append(oeffnungsbefunde, *befund)
			}

			continue
		}
		befunde := append([]domain.Befund(nil), oeffnungsbefunde...)
		befunde = append(befunde,
			ausMsepaket(paket.ModuleEinbinden(leser))...)
		skript := paket.Initskript()

		// **Und die Kartenfelder** — für die Vorgabewerte
		// (`kartenvorgaben.go`). Ein Feld, das die Karte nicht führt, trägt den
		// Wert des Spiels und nicht `nil`; ohne ihn entscheidet jede
		// Positionsangabe, die `card.<feld>` vergleicht, falsch herum
		// (#276).
		spielmodell, modellbefunde := paket.Spielmodell()
		befunde = append(befunde, ausMsepaket(modellbefunde)...)

		// **Und die Fassungen des Setsymbols** — das Spiel erklärt die
		// Grundfassungen `common`, `uncommon`, `rare` und `mythic rare`
		// (#323).
		fassungen := paket.Spielsymbolfassungen()

		// **Und die Schlüsselwörter** — vorbereitet, nicht roh
		// (#1061).
		//
		// **Einmal je Spielpaket und nicht je Karte.** `magic.mse-game` führt
		// 334 Schlüsselwörter; ihre Ausdrücke bei jeder Vorschau neu zu
		// übersetzen hieße, bei jedem Tastendruck 334 Regexe zu bauen.
		woerter, wortbefunde := NeueWortliste(
			paket.Schluesselwoerter(), paket.Schluesselwortparameterarten())
		befunde = append(befunde, wortbefunde...)

		_ = paket.Schliessen() //nolint:errcheck // Skript und Modell sind kopiert

		return Spielangaben{
			Skript: skript, Modell: spielmodell,
			Kartenskripte:       alsKartenskriptangaben(paket.Kartenskripte()),
			Auszaehldimensionen: mitAutomatischen(spielmodell.Fields, Auszaehlangaben(paket.Auszaehldimensionen())),
			Symbolfassungen:     fassungen, Schluesselwoerter: woerter,
			Wortlisten: paket.Wortlisten(),
			Rohwoerter: paket.Schluesselwoerter(), Modi: paket.Schluesselwortmodi(),
			Parameterarten: paket.Schluesselwortparameterarten(),
			Packs:          paket.Packs(),
		}, befunde
	}
	if len(oeffnungsbefunde) > 0 {
		return Spielangaben{}, oeffnungsbefunde
	}

	return Spielangaben{}, befundliste(
		"das Spielpaket %q liegt in keinem Paketverzeichnis — "+
			"Feldstile, die seine Funktionen rufen, bleiben ungedeutet", paketname)
}

// paketOeffnungsbefund trennt „nicht installiert" von „vorhanden, aber kaputt".
func paketOeffnungsbefund(pfad, art string, err error) *domain.Befund {
	if errors.Is(err, fs.ErrNotExist) {
		if _, statfehler := os.Stat(pfad); errors.Is(statfehler, fs.ErrNotExist) {
			return nil
		}
	}

	return &domain.Befund{
		Datei:   filepath.Base(pfad),
		Meldung: "das " + art + " ließ sich nicht öffnen: " + err.Error(),
	}
}

// kopfMit sucht einen Setkopf.
func kopfMit(koepfe []domain.SetHead, setID int64) (domain.SetHead, bool) {
	for _, kopf := range koepfe {
		if kopf.ID == setID {
			return kopf, true
		}
	}

	return domain.SetHead{}, false
}

// befundliste macht aus einem Satz eine Befundliste.
func befundliste(muster string, teile ...any) []domain.Befund {
	return []domain.Befund{{Meldung: fmt.Sprintf(muster, teile...)}}
}

// MitBildablage sagt dem Speicher, wo eingesetzte Bilder liegen.
//
// **Eine Naht und kein Pfad**: wo das Benutzerverzeichnis liegt, weiß `app`
// (§2) — `service` weiß, dass es eines gibt.
func (s *Stilspeicher) MitBildablage(ablage *Bildablage) *Stilspeicher {
	s.bilder = ablage

	return s
}

// setort trägt Herkunft und Bildablage eines Sets in eine Vorlage ein.
//
// **An einer Stelle für drei Aufrufer.** Sie standen als `vorlage.Setpaket =
// kopf.Origin` dreimal nebeneinander; die zweite Angabe daneben wäre beim
// dritten Mal vergessen worden.
func (s *Stilspeicher) setort(vorlage *Vorlage, kopf domain.SetHead) {
	vorlage.Setpaket = kopf.Origin

	if s.bilder != nil {
		vorlage.Bilderverzeichnis = s.bilder.Verzeichnis(kopf.ID)
	}
}
