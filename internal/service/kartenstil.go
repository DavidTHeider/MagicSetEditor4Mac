package service

// Der Stil **einer Karte** — ihr eigener, sonst der des Sets.
//
//	const StyleSheet& Set::stylesheetFor(const CardP& card) {
//	  if (card && card->stylesheet) return *card->stylesheet;
//	  else                          return *stylesheet;
//	}
//
// `src/data/set.cpp`, Zeilen 90–93, Fassung f9d9356 vom 2021-07-06 (§15).
//
// # Woran es aufgefallen ist
//
// In `Archenemy.mse-set` — einem kanonischen MSE2-Set — steht
// `stylesheet: archenemy` **70-mal an den Karten** und `m15-altered` einmal am
// Set. Alle 70 Karten wurden über die falsche Vorlage gezeichnet
// (#395).
//
// # Warum das Feld `stylesheet` heißt und kein Kartenfeld ist
//
// MSE führt es an der **Karte** (`Card::stylesheet`, `src/data/card.hpp:44`),
// nicht unter ihren Feldwerten. Dieses Projekt legt beim Import jeden
// Kartenschlüssel als Wert ab; der Stil steht damit unter dem Namen
// `stylesheet`. Ihn zu einer eigenen Spalte zu machen ist eigener Umfang —
// gelesen wird er hier, wo er gebraucht wird.

import (
	"context"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Kartenstilfeld ist der Name, unter dem der Stil einer Karte abgelegt ist.
const Kartenstilfeld = "stylesheet"

// StilFuerKarte gibt den Stil dieser Karte; der dritte Wert sagt, ob es einen
// gibt.
//
// **Die Setangaben bleiben die des Sets.** `set.symbol`, `set.copyright` und
// `set.border_color` hängen am Set und nicht am Stylesheet — dasselbe gilt für
// den Ort des Set-Pakets ([Vorlage.Setpaket]).
func (s *Stilspeicher) StilFuerKarte(ctx context.Context, karte *domain.CardRow) (
	Vorlage, []domain.Befund, bool,
) {
	eigener := ""
	if karte != nil {
		wert, _ := karte.Value(Kartenstilfeld)
		eigener = strings.TrimSpace(wert)
	}

	if eigener == "" {
		return s.StilFuerSet(ctx, karte.SetID)
	}

	s.sperre.Lock()
	defer s.sperre.Unlock()

	vorlage, befunde, da := s.gemerkterKartenstil(ctx, karte.SetID, eigener)
	if !da {
		return vorlage, befunde, false
	}

	// **Dieselbe Reihenfolge wie bei [Stilspeicher.StilFuerSet]**: erst der
	// Stil, dann die Angaben, und ein Fehler bei den Angaben hält den Stil
	// nicht auf (§8).
	werte, err := s.koepfe.Setwerte(ctx, karte.SetID)
	if err != nil {
		befunde = append(befunde, befundliste("die Setangaben lesen: %v", err)...)
	}

	vorlage.Setangaben = werte

	return vorlage, befunde, true
}

// gemerkterKartenstil holt einen Stil **nach Namen** und merkt ihn danach.
//
// **Nach Namen und nicht nach Karte**: ein Set hat wenige Stile und viele
// Karten. Wer je Karte merkte, öffnete dasselbe Paket siebzigmal.
func (s *Stilspeicher) gemerkterKartenstil(
	ctx context.Context, setID int64, name string,
) (Vorlage, []domain.Befund, bool) {
	kopf, befunde, da := s.setkopf(ctx, setID)
	if !da {
		return Vorlage{}, befunde, false
	}

	// **Nach Namen und Spiel** (#2005): `stilpaketnamen` setzt das Spiel
	// davor, und derselbe Name meint unter einem anderen Spiel ein anderes
	// Paket — und ein anderes Spielskript.
	schluessel := name + "\x00" + kopf.Game

	if gemerkt, drin := s.gemerkteStile[schluessel]; drin && s.stempelGilt(&gemerkt) {
		s.gemerkteStile[schluessel] = gemerkt
		vorlage := gemerkt.vorlage
		s.setort(&vorlage, kopf)

		return vorlage, gemerkt.befunde, gemerkt.da
	}

	if len(s.wurzeln) == 0 {
		return Vorlage{},
			befundliste("für den Stil %q ist kein Paketverzeichnis eingerichtet", name),
			false
	}

	vorlage, weitere, gefunden, pakete := s.ausPaket(name, kopf.Game)
	befunde = append(befunde, weitere...)

	s.gemerkteStile[schluessel] = s.gestempelt(vorlage, weitere, gefunden, pakete)

	s.setort(&vorlage, kopf)

	return vorlage, befunde, gefunden
}
