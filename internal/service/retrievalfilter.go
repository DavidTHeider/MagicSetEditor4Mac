package service

// **Die harten Filter** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Drei Sätze, aus denen jede Zeile hier folgt
//
// **„Unbekannter Wert ist kein erfundener Treffer."** Wo ein Dokument ein
// Merkmal nicht hat, fällt es durch jeden Filter über dieses Merkmal — auch
// wenn nichts dagegen spricht.
//
// **„Querytexte ersetzen keine geprüfte Filterstruktur."** Hier steht kein
// Zerteiler für „rote Kreatur bis 3"; was gefiltert wird, kommt als Struktur
// herein und nicht als Satz.
//
// **„Farbe getrennt von Farbidentität."** Zwei Felder, zwei Bedeutungen —
// siehe [domain.Retrievalmerkmale].
//
// # Und warum das hier steht und nicht im SQL
//
// **Weil die Kandidaten ohnehin schon da sind.** Die semantische Hälfte
// rechnet über alle sichtbaren Zeilen (siehe `storage/retrievalsuche.go`);
// ein zweiter Filter im SQL spränge die Zahl der Zeilen nicht und brächte
// dieselbe Regel an eine zweite Stelle. Was er brächte, wäre ein größeres
// Kandidatenfenster — und das ist [Kandidatenfenster], mit dem, was es nicht
// leistet, daneben.

import (
	"slices"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Retrievalfilter schränkt die Treffer ein.
//
// **Eine leere Angabe heißt „nicht filtern"** — nicht „nichts zulassen".
type Retrievalfilter struct {
	Quellen  []domain.Retrievalquelle
	Spiele   []string
	Sprachen []string

	// Profile schränkt auf bestimmte Profile ein; die **Sichtprüfung** macht
	// die Ablage und nicht dieser Filter (Issue: „ACL bereits in
	// Kandidatenauswahl").
	Profile []int64

	Sets []string

	// Farben filtert nach der Farbe der Karte, Farbidentitaet nach ihrer
	// Identität — **beides muss zutreffen**, wenn beides gesetzt ist.
	Farben         []string
	Farbidentitaet []string

	// Typen müssen **alle** vorkommen: `land` und `basic` finden das
	// Standardland und nicht jedes Land.
	Typen []string

	// ManawertVon und ManawertBis sind einschließlich; `nil` heißt offen.
	ManawertVon *float64
	ManawertBis *float64

	// Format nennt das Format, Formatlage die verlangte Lage (`legal`).
	//
	// **Formatstand verlangt einen bestimmten Datenstand** — leer heißt: der
	// Stand des Dokuments genügt, er wird aber trotzdem verlangt
	// ([domain.Retrievalmerkmale.Format]).
	Format      string
	Formatlage  string
	Formatstand string
}

// leer sagt, ob gar nicht gefiltert wird.
func (f Retrievalfilter) leer() bool {
	return len(f.Quellen) == 0 && len(f.Spiele) == 0 && len(f.Sprachen) == 0 &&
		len(f.Profile) == 0 && len(f.Sets) == 0 && len(f.Farben) == 0 &&
		len(f.Farbidentitaet) == 0 && len(f.Typen) == 0 &&
		f.ManawertVon == nil && f.ManawertBis == nil && f.Format == ""
}

// laesstDurch sagt, ob ein Dokument bleibt.
func (f Retrievalfilter) laesstDurch(dokument domain.Retrievaldokument) bool {
	return passt(f.Quellen, dokument.Quelle) &&
		passt(f.Spiele, dokument.Spiel) &&
		passt(f.Sprachen, dokument.Sprache) &&
		passt(f.Profile, dokument.ProfilID) &&
		f.passtZuMerkmalen(dokument.Merkmale)
}

// passtZuDenFarben hält Farbe, Farbidentität und Typzeile.
func (f Retrievalfilter) passtZuDenFarben(merkmale domain.Retrievalmerkmale) bool {
	return enthaeltAlle(merkmale.Farben, f.Farben) &&
		enthaeltAlle(merkmale.Farbidentitaet, f.Farbidentitaet) &&
		enthaeltAlle(merkmale.Typen, f.Typen)
}

// passtZuMerkmalen prüft Farben, Typen, Set, Manawert und Format.
func (f Retrievalfilter) passtZuMerkmalen(merkmale domain.Retrievalmerkmale) bool {
	return f.passtZuDenFarben(merkmale) &&
		passt(f.Sets, merkmale.Set) &&
		f.passtZumManawert(merkmale.Manawert) &&
		f.passtZumFormat(merkmale)
}

// passtZumManawert hält die Spanne.
//
// **Ein Dokument ohne Manawert fällt durch jede Spanne.** Es hat keinen —
// und „kein Wert" liegt in keinem Bereich.
func (f Retrievalfilter) passtZumManawert(wert *float64) bool {
	if f.ManawertVon == nil && f.ManawertBis == nil {
		return true
	}

	if wert == nil {
		return false
	}

	return (f.ManawertVon == nil || *wert >= *f.ManawertVon) &&
		(f.ManawertBis == nil || *wert <= *f.ManawertBis)
}

// passtZumFormat hält Format, Lage und Datenstand zusammen.
func (f Retrievalfilter) passtZumFormat(merkmale domain.Retrievalmerkmale) bool {
	if f.Format == "" {
		return true
	}

	lage, stand, bekannt := merkmale.Format(f.Format)
	if !bekannt {
		return false
	}

	if f.Formatstand != "" && f.Formatstand != stand {
		return false
	}

	return f.Formatlage == "" || strings.EqualFold(f.Formatlage, lage)
}

// passt sagt, ob ein Wert zur Auswahl gehört — eine leere Auswahl filtert nicht.
func passt[T comparable](auswahl []T, wert T) bool {
	return len(auswahl) == 0 || slices.Contains(auswahl, wert)
}

// enthaeltAlle sagt, ob jedes Verlangte vorhanden ist.
//
// **Alle und nicht eines.** `land` und `basic` zusammen meinen ein
// Standardland; wer „eines davon" wollte, fragte zweimal.
func enthaeltAlle(vorhanden, verlangt []string) bool {
	for _, einzelnes := range verlangt {
		if !slices.Contains(vorhanden, einzelnes) {
			return false
		}
	}

	return true
}
