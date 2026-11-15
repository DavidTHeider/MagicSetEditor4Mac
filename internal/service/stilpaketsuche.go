package service

// Die Suche nach dem Stilpaket eines Sets — **welchen Namen** es trägt und in
// welcher Wurzel es liegt.
//
// **Eigene Datei, seit der Ort des Sets dazukam**
// (#380):
// `stilquelle.go` beantwortet, was ein Set an Vorlage bekommt; hier steht, wie
// das Paket dazu gefunden wird. Der Anlass war die 400-Zeilen-Grenze (§2), der
// Schnitt folgt der Frage, die eine Datei beantwortet.

import (
	"fmt"
	"path/filepath"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// stilpaketnamen sind die Namen, unter denen ein `stylesheet:` zu suchen ist.
//
// # `stylesheet:` trägt die Kurzform
//
// In einem echten `.mse-set` steht **nicht** der Paketname, sondern der Name
// **ohne das Spiel davor**: `game: magic` mit `stylesheet: m15-altered` meint
// das Paket `magic-m15-altered.mse-style`.
//
// **Quelle:** `magic-set-generator` (MIT), `src/mse.rs`, Fassung 1b8fb7f vom 2021-08-09. Es legt ein neues Set
// mit `("stylesheet", … "m15-altered")` an und benennt den zugehörigen
// Stylingblock über `format!("{}-{}", mse_game, stylesheet.render())` als
// `magic-m15-altered`. Ein echtes `.mse-set` liegt hier nicht vor; dies ist
// der einzige Beleg dafür, was in einem steht.
//
// # Warum beide Formen gesucht werden
//
// Die zusammengesetzte **zuerst**, denn sie ist die belegte. Der rohe Wert
// danach, weil §8 fremde Daten zum Normalfall macht: ein Set, das den vollen
// Paketnamen trägt, abzuweisen wäre eine Strenge ohne Gegenwert — und unsere
// eigene Saat schreibt ihn so.
//
// Ohne Spielangabe gibt es keine Kurzform aufzulösen; ein vorangestellter
// Bindestrich wäre ein Name, den kein Paket trägt.
//
// **Auch wenn der Wert schon mit dem Spielnamen beginnt**, wird beides gesucht.
// Der erste Anlauf sprang bei `magic-foo` gleich zur rohen Form — und ein
// Stilpaket darf `magic-magic-foo.mse-style` heißen, wenn sein Kurzname
// `magic-foo` lautet. Ein Präfix ist ein Zeichenmuster und keine Auskunft
// darüber, welche der beiden Formen dasteht; §8 macht solche Daten zum
// Normalfall. Angestrichen vom Zweitblick über Phase A.
//
// **Was dieser Fehler gekostet hat:** bis zum 2026-08-16 wurde nur der rohe
// Wert gesucht. Jedes echte Set fand damit kein Stilpaket, und die Anwendung
// zeigte einen Behelfsplan mit den rohen Feldwerten
// (#292).
func stilpaketnamen(name, spiel string) []string {
	if spiel == "" {
		return []string{name}
	}

	return []string{spiel + "-" + name, name}
}

// ausPaket öffnet das Stilpaket und liest sein Modell.
//
// **Ohne die Auskunft „merken?"**, die [Stilspeicher.holen] nach außen gibt:
// wer bis hierher kommt, hat einen Setkopf gelesen und eine Wurzel — beides
// ändert sich nicht mehr, und das Ergebnis taugt zu merken, ob das Paket nun
// gefunden wurde oder nicht. Ein Rückgabewert, der immer dasselbe sagt, sähe
// aus wie eine Entscheidung und wäre keine.
//
// **Der vierte Wert nennt die Pakete, aus denen die Vorlage entstand** — Stil,
// Spiel und jedes Modul —, damit das Gedächtnis merkt, wenn sich eines ändert
// (`stilstempel.go`). Auch die Kandidaten, die es nicht gab: entsteht eines
// später, gilt es ohne Neustart.
func (s *Stilspeicher) ausPaket(name, spiel string) (
	vorlage Vorlage, befunde []domain.Befund, gefunden bool, pakete []string,
) {
	var oeffnungsbefunde []domain.Befund

	leser := &protokollleser{innen: msepaket.NeuerPaketleser(s.wurzeln...)}
	pakete = []string{spiel + "." + string(msepaket.ArtSpiel)}

	for _, kandidat := range stilpaketnamen(name, spiel) {
		pakete = append(pakete, kandidat+"."+string(msepaket.ArtStil))

		gelesen, weitere, da := s.unterNamen(kandidat, spiel, oeffnungsbefunde, leser)
		if da {
			return gelesen, weitere, true, append(pakete, leser.namen()...)
		}

		oeffnungsbefunde = weitere
	}

	return Vorlage{}, append(oeffnungsbefunde,
			domain.Befund{Meldung: fmt.Sprintf(
				"das Stilpaket %q liegt in keinem Paketverzeichnis", name+"."+string(msepaket.ArtStil))}),
		false, pakete
}

// unterNamen sucht **einen** Paketnamen in allen Wurzeln.
//
// `gefunden` sagt, ob das Paket aufging. Die Öffnungsbefunde der erfolglosen
// Wurzeln reicht der Aufrufer an den nächsten Kandidaten weiter — sonst
// verschwiege eine geglückte zweite Suche, dass die erste an einer kaputten
// Datei gescheitert ist.
func (s *Stilspeicher) unterNamen(name, spiel string, bisher []domain.Befund, leser msepaket.Paketquelle) (
	vorlage Vorlage, befunde []domain.Befund, gefunden bool,
) {
	paketname := name + "." + string(msepaket.ArtStil)
	oeffnungsbefunde := bisher

	for _, wurzel := range s.wurzeln {
		pfad := filepath.Join(wurzel, paketname)

		paket, err := s.oeffne(pfad)
		if err != nil {
			if befund := paketOeffnungsbefund(pfad, "Stilpaket", err); befund != nil {
				oeffnungsbefunde = append(oeffnungsbefunde, *befund)
			}

			continue
		}
		// **Die Module zuerst** (#186): ohne sie sieht der Planbauer die
		// Funktionen des `init script:` nicht, und aus 5.950 Feldstilen würden
		// wieder 2.016.
		befunde := append([]domain.Befund(nil), oeffnungsbefunde...)
		befunde = append(befunde,
			ausMsepaket(paket.ModuleEinbinden(leser))...)

		stil, modellbefunde := paket.Stilmodell()
		befunde = append(befunde, ausMsepaket(modellbefunde)...)

		// **Und das Spielpaket dazu** (#195): dort stehen die Funktionen, mit
		// denen ein Stil rechnet. Fehlt es, entstehen die Feldebenen, die ohne
		// sie auskommen — der Rest wird benannt (§8).
		spiel, spielbefunde := s.spielskript(spiel, leser)
		befunde = append(befunde, spielbefunde...)
		_ = paket.Schliessen() //nolint:errcheck // Modell und Skripte sind vollständig kopiert

		// **Und die Symbolschriften**, die der Stil nennt (#173). Der Planbauer
		// öffnet keine Pakete; er bekommt ein Modell und rechnet.
		schriften, schriftbefunde := s.Symbolschriften(stil)
		befunde = append(befunde, schriftbefunde...)

		// Die Bildquelle öffnet Bilddateien später selbst über den Paketpfad.
		// Der Stilspeicher hält deshalb keinen Archivdeskriptor im Cache.
		return Vorlage{
			Stil: mitErklaertenNamen(stil, spiel.Modell.Fields), Spielskript: spiel.Skript,
			Kartenfelder: spiel.Modell.Fields, Setfelder: spiel.Modell.SetFields,
			Symbolschriften:     schriften,
			Symbolfassungen:     spiel.Symbolfassungen,
			Schluesselwoerter:   spiel.Schluesselwoerter,
			Wortlisten:          spiel.Wortlisten,
			Rohwoerter:          spiel.Rohwoerter,
			Modi:                spiel.Modi,
			Parameterarten:      spiel.Parameterarten,
			Packs:               spiel.Packs,
			Kartenskripte:       spiel.Kartenskripte,
			Auszaehldimensionen: spiel.Auszaehldimensionen,
		}, befunde, true
	}

	return Vorlage{}, oeffnungsbefunde, false
}

// oeffne öffnet ein Stilpaket — mit dem Ersatz, wenn einer auf es zielt.
//
// **Verglichen wird der Pfad und nicht der Name.** Zwei Wurzeln können
// dasselbe Paket führen; gemeint ist das, dessen Datei jemand gerade
// bearbeitet (#1164).
func (s *Stilspeicher) oeffne(pfad string) (*msepaket.Paket, error) {
	if s.ersatz == nil || filepath.Clean(s.ersatz.Paket) != filepath.Clean(pfad) {
		return msepaket.Oeffne(pfad)
	}

	return msepaket.OeffneMitErsatz(pfad, s.ersatz.Datei, s.ersatz.Inhalt)
}
