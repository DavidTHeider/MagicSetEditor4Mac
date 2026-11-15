package service

// **Was hinausginge, bevor es hinausgeht** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Der Satz, um den herum das gebaut ist
//
// **„Vor externer Generierung konkreten Provider, Datenkategorien und Umfang
// des **zusammengestellten** Kontexts samt Kostenlimit zeigen/autorisiert
// binden."**
//
// Das Wort „zusammengestellten" ist der Punkt: gezeigt wird nicht, was in der
// Anfrage stand, sondern **was das Paket wirklich enthält**. Zwischen beidem
// liegt die Auswahl — und wer die Anfrage zeigte, zeigte etwas anderes, als
// er schickt.
//
// # Und warum eine Einbettung auch ein Export ist
//
// **Das Issue sagt es ausdrücklich**: „externe Embeddings/Reranking sind
// ebenfalls Datenexport, auch wenn der Fachvorgang lesend ist." Ein Text, der
// zu einem fremden Rechner geht, um dort zu einem Vektor zu werden, ist dort
// gewesen.

import (
	"sort"
	"strconv"
	"strings"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Datenkategorie sagt, **welche Art** von Inhalt hinausginge.
type Datenkategorie string

// Die vier Kategorien.
//
// **Vier und nicht zwei.** „Öffentlich" und „privat" reichte nicht: ein
// Oracle-Text ist öffentlich und trotzdem fremdes Material, und eine eigene
// Setnotiz ist privat und gehört einem Profil. Wer sie zusammenfasste, ließe
// jemanden über das eine entscheiden und das andere mitschicken.
const (
	// KategorieAufgabe ist, was ein Mensch getippt hat.
	KategorieAufgabe Datenkategorie = "aufgabe"

	// KategorieOeffentlich sind Oracle-Texte, Regelstellen und Rulings.
	KategorieOeffentlich Datenkategorie = "oeffentlich"

	// KategorieEigeneKarten sind Karten aus diesem Bestand.
	KategorieEigeneKarten Datenkategorie = "eigenekarten"

	// KategorieEigeneTexte sind Setnotizen und eigene Texte.
	KategorieEigeneTexte Datenkategorie = "eigenetexte"
)

// Kontextfreigabe ist, was vor dem Senden gezeigt wird.
type Kontextfreigabe struct {
	PaketID string

	Zielanbieter string
	Zielmodell   string

	// Ort sagt, **wo** gerechnet wird — das entscheidet, ob überhaupt etwas
	// hinausgeht.
	Ort domain.KIOrt

	// Zieladresse ist, **wohin** es ginge — Werkstück 3 von A18
	// (#1533).
	//
	// **Der Anbietername genügt nicht.** Seit die Adresse einstellbar ist,
	// kann hinter „openai" ein Firmengateway, OpenRouter oder Azure stehen —
	// und das ist ein **Datenexport an einen Dritten**, auch wenn dahinter
	// derselbe Anbieter rechnet. Wer nur den Namen zeigte, verschwiege genau
	// den Unterschied, um den es bei einer Freigabe geht.
	//
	// **Leer heißt: die Vorgabe des Anbieters.**
	Zieladresse string

	// Kategorien nennt, welche Arten von Inhalt im Paket stehen.
	Kategorien []Datenkategorie

	// Belege, Zeichen und Tokens sind der Umfang.
	Belege  int
	Zeichen int
	Tokens  int

	// Quellen nennt je Korpus, wie viele Belege daraus stammen.
	Quellen map[domain.Retrievalquelle]int

	// Privat sagt, ob **eigener** Inhalt dabei ist.
	//
	// **Eine eigene Zeile und nicht aus den Kategorien erschlossen**: das
	// ist die Frage, bei der jemand zweimal hinsieht, und sie soll nicht in
	// einer Liste stehen, die man überfliegt.
	Privat bool
}

// FreigabeFuer beschreibt, was ein Paket enthält.
//
// **Sie liest das Paket und nicht die Anfrage** — das Issue verlangt den
// Umfang des **zusammengestellten** Kontexts.
func FreigabeFuer(
	paket Kontextpaket, ort domain.KIOrt, adresse string,
) Kontextfreigabe {
	freigabe := Kontextfreigabe{
		Zieladresse:  adresse,
		PaketID:      paket.Manifest.PaketID,
		Zielanbieter: paket.Manifest.Zielanbieter,
		Zielmodell:   paket.Manifest.Zielmodell,
		Ort:          ort,
		Belege:       len(paket.Manifest.Belege),
		Zeichen:      len([]rune(paket.Kontext)) + len([]rune(paket.Manifest.Aufgabe)),
		Quellen:      map[domain.Retrievalquelle]int{},
	}

	freigabe.Tokens = Tokenschaetzung(paket.Kontext) +
		Tokenschaetzung(paket.Manifest.Aufgabe)

	gesehen := map[Datenkategorie]bool{KategorieAufgabe: true}

	for _, beleg := range paket.Manifest.Belege {
		freigabe.Quellen[beleg.Quelle]++

		kategorie := kategorieFuer(beleg.Quelle)
		gesehen[kategorie] = true

		if kategorie == KategorieEigeneKarten || kategorie == KategorieEigeneTexte {
			freigabe.Privat = true
		}
	}

	freigabe.Kategorien = geordneteKategorien(gesehen)

	return freigabe
}

// kategorieFuer ordnet eine Quelle ihrer Kategorie zu.
func kategorieFuer(quelle domain.Retrievalquelle) Datenkategorie {
	switch quelle {
	case domain.QuelleEigeneKarte:
		return KategorieEigeneKarten

	case domain.QuelleEigenerText:
		return KategorieEigeneTexte

	case domain.QuelleOracle, domain.QuelleRuling, domain.QuelleRegelwerk:
		return KategorieOeffentlich

	default:
		// **Was niemand kennt, gilt als eigen.** Eine unbekannte Quelle
		// stillschweigend als öffentlich zu führen wäre die Gestalt, in der
		// eine neue Quelle unbemerkt hinausginge.
		return KategorieEigeneTexte
	}
}

// geordneteKategorien gibt die Kategorien in fester Reihenfolge.
func geordneteKategorien(gesehen map[Datenkategorie]bool) []Datenkategorie {
	kategorien := make([]Datenkategorie, 0, len(gesehen))
	for kategorie := range gesehen {
		kategorien = append(kategorien, kategorie)
	}

	sort.Slice(kategorien, func(links, rechts int) bool {
		return kategorien[links] < kategorien[rechts]
	})

	return kategorien
}

// Zusammenfassung gibt einen Satz über den Umfang.
//
// **Für die Flächen ohne Fenster**, die keine Tabelle zeigen können — und als
// die eine Stelle, an der der Satz gebildet wird: zwei, die ihn bilden,
// laufen auseinander.
func (f Kontextfreigabe) Zusammenfassung() string {
	teile := make([]string, 0, len(f.Kategorien))

	for _, kategorie := range f.Kategorien {
		teile = append(teile, string(kategorie))
	}

	satz := "An " + f.Zielanbieter + "/" + f.Zielmodell + " gingen " +
		zahlwort(f.Belege, "Beleg", "Belege") + " (" +
		zahlwort(f.Tokens, "Token", "Tokens") + " geschätzt, " +
		strings.Join(teile, ", ") + ")."

	if f.Privat {
		satz += " Darunter sind eigene Inhalte."
	}

	if f.Ort == domain.KIOrtLokal {
		satz += " Gerechnet wird auf diesem Rechner."
	}

	return satz
}

// zahlwort setzt eine Zahl mit Ein- oder Mehrzahl.
func zahlwort(zahl int, einer, mehrere string) string {
	wort := mehrere
	if zahl == 1 {
		wort = einer
	}

	return strconv.Itoa(zahl) + " " + wort
}
