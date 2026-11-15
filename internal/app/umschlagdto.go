package app

// **Der Umschlag an der Naht** — Werkstück N4 von Phase 5.4
// (#1457).

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/austausch"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/peer"
)

// UmschlagobjektDTO ist ein Stück Inhalt, wie das Fenster es liest.
type UmschlagobjektDTO struct {
	// Art sagt, was es ist — „set", „karte", „vorlage", „deck",
	// „sammlungssicht".
	Art string `json:"art"`

	// Bezeichnung ist, was ein Mensch liest — **kein Nachweis**.
	Bezeichnung string `json:"bezeichnung"`

	// Groesse ist die Zahl der Bytes.
	Groesse int64 `json:"groesse"`

	// Braucht nennt, was dieses Objekt zum Vollständigsein braucht.
	Braucht []string `json:"braucht"`
}

// UmschlagAntwort ist, was in einem Umschlag steht.
type UmschlagAntwort struct {
	Objekte []UmschlagobjektDTO `json:"objekte"`

	// Urheber ist die Peerkennung dessen, der unterschrieben hat — leer
	// beim Schnüren, denn dort ist man es selbst.
	Urheber string `json:"urheber"`

	// Fingerabdruck ist die Form, die zwei Menschen vergleichen.
	//
	// **Und sie ist der einzige Nachweis, der zählt**: eine Unterschrift
	// sagt, dass der Umschlag von dieser Kennung kommt — ob man ihr traut,
	// sagt die Kontaktliste.
	Fingerabdruck string `json:"fingerabdruck"`

	// Pfad ist, wo der angesehene Umschlag liegt — leer beim Schnüren.
	//
	// **Er kommt zurück, weil das Übernehmen ihn braucht**: das Fenster
	// kennt ihn sonst nicht, denn gewählt hat ihn der native Dialog.
	Pfad string `json:"pfad"`

	// Abgebrochen heißt: der Dialog wurde weggeklickt.
	//
	// **Und das ist kein Fehler** — derselbe Unterschied wie beim Speichern
	// eines Sets.
	Abgebrochen bool `json:"abgebrochen"`

	Fehler *Fehler `json:"fehler"`
}

// objekteAus übersetzt ein Manifest in die Gestalt für das Fenster.
//
// **Eine leere Liste und kein `nil`** (§3).
func objekteAus(manifest austausch.Manifest) []UmschlagobjektDTO {
	objekte := make([]UmschlagobjektDTO, 0, len(manifest.Objekte))

	for _, objekt := range manifest.Objekte {
		braucht := objekt.Braucht
		if braucht == nil {
			braucht = []string{}
		}

		objekte = append(objekte, UmschlagobjektDTO{
			Art:         string(objekt.Art),
			Bezeichnung: objekt.Bezeichnung,
			Groesse:     objekt.Groesse,
			Braucht:     braucht,
		})
	}

	return objekte
}

// fingerabdruckVon gibt die vergleichbare Form einer Kennung.
func fingerabdruckVon(kennung peer.Kennung) string {
	return peer.Fingerabdruck(kennung)
}
