package service

// **Woraus das Zeichnen eines Sets außer dem Bestand liest** — der Abdruck
// einer Problemprüfung (PERF02, #2352).
//
// # Was er faltet
//
// **Jede Datei jedes Pakets, aus dem gezeichnet wird**: die Pakete der
// Setvorlage samt ihren Einschlüssen, die der **eigenen Stile der Karten**
// (`stylesheet:` an der Karte, `kartenstil.go`) samt ihren, die der
// **Symbolschriften**, die einer dieser Stile nennt, samt ihren — und jedes
// Bild neben dem Set. Eine geänderte Datei in einem vorhandenen Paket ändert
// die Zeit seiner Wurzel nicht (Zweitblick über Phase 3, Befund 2).
//
// # Und aus demselben Gang wie das Zeichnen
//
// **Der Abdruck geht nicht an der [stempelpause] vorbei** (#2423, Nachlauf
// über Phase 3, Befund 2): er prüft die gemerkten Vorlagen frisch, und weicht
// ihr Stempel ab, werden sie neu gelesen. Bis dahin nahm das Zeichnen in der
// Pause die gemerkte Vorlage, während der Abdruck frisch ging — und ein neuer
// Abdruck stand über Befunden aus der alten.
//
// **Deterministisch, gleich ob schon gezeichnet wurde**: was ein Abdruck
// braucht und noch nicht gemerkt ist, liest er selbst. Sonst nennte der erste
// Abdruck weniger Pakete als jeder spätere, und eine Prüfung wäre nach dem
// ersten Zeichnen stets veraltet.
//
// # Was er kostet
//
// **Gemessen am 2026-10-16**, einmalig und nicht eingecheckt, auf Mac17,3
// (Apple M5, 32 GiB, Last unter 4): ein Set mit Stil M15, `archenemy` an den
// Karten, über `full-magic-pack/data` — **65 ms je Abdruck** über 20 Läufe.
// Gebildet wird er je Prüfung einmal und je Abfrage des Stands; die Grenze
// aus PERF02 gilt der ganzen Prüfung (60 s für 3.788 Karten).
//
// # Was er nicht sieht
//
// **Die Dateien eines eingelesenen Sets an seinem alten Ort** (`herkunft`):
// sie sind das Original, das diese Anwendung nur liest; ein anderer Ort hebt
// den Setstand (Wanderung 0100). **Eine Änderung binnen derselben Zeit auf
// dieselbe Größe** — wie beim Stempel.

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket"
)

// Kartenstilnamen nennt die eigenen Stile der Karten eines Sets — die Werte
// von `stylesheet:` an seinen Karten.
type Kartenstilnamen interface {
	Kartenstile(ctx context.Context, setID int64) ([]string, error)
}

// gemerkteSchrift ist, aus welchen Paketen eine Symbolschrift gelesen wird.
type gemerkteSchrift struct {
	pakete  []string
	stempel string
}

// MitKartenstilen setzt, woher die eigenen Stile der Karten kommen.
//
// **Beim Bau zu setzen**, wie [Stilspeicher.MitPaketwechsel].
func (s *Stilspeicher) MitKartenstilen(quelle Kartenstilnamen) *Stilspeicher {
	s.kartenstile = quelle

	return s
}

// Ressourcenstand faltet zusammen, woraus das Zeichnen eines Sets außer dem
// Bestand liest — der Kopf der Datei sagt, was und wie.
//
// **Ein Set, dessen Kopf sich nicht lesen lässt oder das dieses Profil nicht
// führt, hat den Stand „ohne Stil"** und keinen Fehler — der Stil wird dann
// nicht gemerkt. **Eines, dessen Stilpaket fehlt, hat den Stempel einer
// leeren Paketliste** (Nachlauf über Phase 3): die Befunde der Prüfung nennen
// das Fehlen, und entsteht das Paket, ändert sich die Wurzel — die steht im
// Stand der Schriftsuche.
func (s *Stilspeicher) Ressourcenstand(ctx context.Context, setID int64) string {
	s.sperre.Lock()

	s.frischPruefen(setID)
	s.gemerktenStil(ctx, setID)
	gemerkt, da := s.gemerkt[setID]

	if !da {
		s.sperre.Unlock()

		return "ohne Stil"
	}

	pakete := slices.Clone(gemerkt.pakete)
	stile := []msepaket.Stil{gemerkt.vorlage.Stil}
	namen, lesefehler := s.kartenstilnamen(ctx, setID)

	for _, name := range namen {
		schluessel := name + "\x00" + gemerkt.spiel
		if karte, gibts := s.gemerkteStile[schluessel]; gibts {
			karte.geprueft = time.Time{}
			s.gemerkteStile[schluessel] = karte
		}

		vorlage, _, _ := s.gemerkterKartenstil(ctx, setID, name)
		pakete = append(pakete, s.gemerkteStile[schluessel].pakete...)
		stile = append(stile, vorlage.Stil)
	}

	for _, stil := range stile {
		for _, schrift := range symbolschriftnamen(stil) {
			pakete = append(pakete, s.schriftpaketeFuer(schrift)...)
		}
	}

	ordner := gemerkt.vorlage.Bilderverzeichnis
	s.sperre.Unlock()

	stand := paketstempel(s.wurzeln, einmal(pakete))

	if ordner != "" {
		stand += "+" + paketstempel([]string{filepath.Dir(ordner)}, []string{filepath.Base(ordner)})
	}

	// **Ein Lesefehler steht im Stand**, statt ihn still zu verkürzen: der
	// nächste Abdruck, der lesen kann, ist ein anderer.
	if lesefehler {
		stand += "+kartenstile?"
	}

	return stand
}

// frischPruefen lässt die gemerkte Vorlage eines Sets beim nächsten Ruf ihren
// Stempel prüfen — ohne die Pause.
func (s *Stilspeicher) frischPruefen(setID int64) {
	if gemerkt, da := s.gemerkt[setID]; da {
		gemerkt.geprueft = time.Time{}
		s.gemerkt[setID] = gemerkt
	}
}

// kartenstilnamen gibt die eigenen Stile der Karten — getrimmt, einmal je
// Name, in fester Reihenfolge; der zweite Wert sagt, ob das Lesen scheiterte.
func (s *Stilspeicher) kartenstilnamen(ctx context.Context, setID int64) ([]string, bool) {
	if s.kartenstile == nil {
		return nil, false
	}

	roh, fehler := s.kartenstile.Kartenstile(ctx, setID)
	if fehler != nil {
		return nil, true
	}

	namen := make([]string, 0, len(roh))
	for _, name := range roh {
		if name = strings.TrimSpace(name); name != "" {
			namen = append(namen, name)
		}
	}

	slices.Sort(namen)

	return slices.Compact(namen), false
}

// schriftpaketeFuer gibt die Pakete, aus denen eine Symbolschrift gelesen
// wird: sie selbst und die Module, die sie einbindet.
//
// **Gemerkt mit Stempel** wie eine Vorlage — ein Abdruck je Übersicht öffnete
// sonst jede Schrift jedes Mal.
func (s *Stilspeicher) schriftpaketeFuer(name string) []string {
	if gemerkt, da := s.schriftpakete[name]; da && paketstempel(s.wurzeln, gemerkt.pakete) == gemerkt.stempel {
		return gemerkt.pakete
	}

	dateiname := name + "." + string(msepaket.ArtSymbolschrift)
	pakete := []string{dateiname}

	for _, wurzel := range s.wurzeln {
		paket, fehler := msepaket.Oeffne(filepath.Join(wurzel, dateiname))
		if fehler != nil {
			continue
		}

		leser := &protokollleser{innen: msepaket.NeuerPaketleser(s.wurzeln...)}
		_ = paket.ModuleEinbinden(leser)
		_ = paket.Schliessen() //nolint:errcheck // gelesen ist, was der Stempel braucht

		pakete = append(pakete, leser.namen()...)

		break
	}

	pakete = einmal(pakete)
	s.schriftpakete[name] = gemerkteSchrift{pakete: pakete, stempel: paketstempel(s.wurzeln, pakete)}

	return pakete
}

// einmal gibt jeden Namen einmal, in der Reihenfolge seines ersten Auftretens.
func einmal(namen []string) []string {
	gesehen := make(map[string]bool, len(namen))
	ergebnis := make([]string, 0, len(namen))

	for _, name := range namen {
		if !gesehen[name] {
			gesehen[name] = true
			ergebnis = append(ergebnis, name)
		}
	}

	return ergebnis
}
