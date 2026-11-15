package domain

// **Weiterarbeiten** — Werkstück U11 von Phase 5.1
// (#1422).
//
// # Was hier steht
//
// Ein **Arbeitsobjekt** ist, woran jemand arbeitet — ein Set, ein Deck, ein
// Paket —, genannt über Art und Kennung, mit dem Namen als Schnappschuss.
// Ein **Arbeitsschritt** sagt, was damit geschah: geöffnet, eingeführt,
// ausgegeben, und wann. **Angeheftet** ist ein Objekt, das jemand oben halten
// will, gleich was zuletzt geschah.
//
// # Warum der Name mitreist
//
// Die Übersicht zeigt Namen, ohne jedes Objekt zu laden — und ein Objekt,
// das es nicht mehr gibt, wird vorher aussortiert (`service`), damit ein
// Name nie auf eine leere Zeile zeigt. Ein fremdes Profil sieht die Zeile
// gar nicht: die Ablage schneidet am Profil.

import (
	"slices"
	"sort"
	"strings"
	"time"
)

// Arbeitsart ist, was für ein Ding ein Arbeitsobjekt ist.
type Arbeitsart string

// Die drei Arten, die die Übersicht führt.
const (
	ArbeitsartSet   Arbeitsart = "set"
	ArbeitsartDeck  Arbeitsart = "deck"
	ArbeitsartPaket Arbeitsart = "paket"
)

// Arbeitsvorgang ist, was mit einem Objekt geschah.
type Arbeitsvorgang string

// Die drei Vorgänge, die die Übersicht unterscheidet.
const (
	VorgangGeoeffnet   Arbeitsvorgang = "geoeffnet"
	VorgangEingefuehrt Arbeitsvorgang = "eingefuehrt"
	VorgangAusgegeben  Arbeitsvorgang = "ausgegeben"
)

var (
	arbeitsarten     = []Arbeitsart{ArbeitsartSet, ArbeitsartDeck, ArbeitsartPaket}
	arbeitsvorgaenge = []Arbeitsvorgang{VorgangGeoeffnet, VorgangEingefuehrt, VorgangAusgegeben}
)

// Arbeitsobjekt ist, woran gearbeitet wurde.
type Arbeitsobjekt struct {
	Art     Arbeitsart
	Kennung string
	Name    string
}

// Validate prüft Art und Kennung.
func (o Arbeitsobjekt) Validate() []*FieldError {
	var befunde []*FieldError

	if !slices.Contains(arbeitsarten, o.Art) {
		befunde = append(befunde, NewFieldError("art", SchluesselValidierung,
			"%q ist keine Art, an der sich arbeiten lässt.", string(o.Art)))
	}

	if strings.TrimSpace(o.Kennung) == "" {
		befunde = append(befunde, NewFieldError("kennung", SchluesselPflichtfeld,
			"Ein Arbeitsobjekt hat eine Kennung."))
	}

	return befunde
}

// Arbeitsschritt ist ein Vorgang an einem Objekt zu einer Zeit.
type Arbeitsschritt struct {
	Objekt    Arbeitsobjekt
	Vorgang   Arbeitsvorgang
	Zeitpunkt time.Time
}

// Validate prüft Objekt, Vorgang und Zeitpunkt.
func (s Arbeitsschritt) Validate() []*FieldError {
	var befunde []*FieldError

	for _, befund := range s.Objekt.Validate() {
		befunde = append(befunde, NewFieldError("objekt."+befund.Field, befund.Key, "%s", befund.Message))
	}

	if !slices.Contains(arbeitsvorgaenge, s.Vorgang) {
		befunde = append(befunde, NewFieldError("vorgang", SchluesselValidierung,
			"%q ist kein Vorgang, den die Übersicht kennt.", string(s.Vorgang)))
	}

	if s.Zeitpunkt.IsZero() {
		befunde = append(befunde, NewFieldError("zeitpunkt", SchluesselPflichtfeld,
			"Ein Schritt hat einen Zeitpunkt."))
	}

	return befunde
}

// Weiterarbeiten ist die Übersicht: der jüngste Schritt, die Schritte je
// Vorgang, das Angeheftete.
type Weiterarbeiten struct {
	Zuletzt     *Arbeitsschritt
	Geoeffnet   []Arbeitsschritt
	Eingefuehrt []Arbeitsschritt
	Ausgegeben  []Arbeitsschritt
	Angeheftet  []Arbeitsobjekt
}

// Weiterarbeit ordnet Schritte und Angeheftetes zur Übersicht — nach Zeit,
// gruppiert nach Vorgang, je Gruppe auf `hoechstens` gekappt.
func Weiterarbeit(schritte []Arbeitsschritt, angeheftet []Arbeitsobjekt, hoechstens int) Weiterarbeiten {
	geordnet := make([]Arbeitsschritt, len(schritte))
	copy(geordnet, schritte)
	sort.SliceStable(geordnet, func(i, j int) bool {
		return geordnet[i].Zeitpunkt.After(geordnet[j].Zeitpunkt)
	})

	uebersicht := Weiterarbeiten{
		Geoeffnet: []Arbeitsschritt{}, Eingefuehrt: []Arbeitsschritt{}, Ausgegeben: []Arbeitsschritt{},
		Angeheftet: []Arbeitsobjekt{},
	}

	if len(geordnet) > 0 {
		juengster := geordnet[0]
		uebersicht.Zuletzt = &juengster
	}

	for _, eines := range geordnet {
		ziel := uebersicht.gruppe(eines.Vorgang)
		if len(*ziel) < hoechstens {
			*ziel = append(*ziel, eines)
		}
	}

	uebersicht.Angeheftet = append(uebersicht.Angeheftet, angeheftet...)

	return uebersicht
}

// gruppe gibt die Liste zu einem Vorgang.
func (w *Weiterarbeiten) gruppe(vorgang Arbeitsvorgang) *[]Arbeitsschritt {
	switch vorgang {
	case VorgangEingefuehrt:
		return &w.Eingefuehrt
	case VorgangAusgegeben:
		return &w.Ausgegeben
	default:
		return &w.Geoeffnet
	}
}

// IstAngeheftet sagt, ob ein Objekt angeheftet ist.
func (w Weiterarbeiten) IstAngeheftet(art Arbeitsart, kennung string) bool {
	for _, eines := range w.Angeheftet {
		if eines.Art == art && eines.Kennung == kennung {
			return true
		}
	}

	return false
}
