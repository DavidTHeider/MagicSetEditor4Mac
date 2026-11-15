package mtgjson

import (
	"encoding/json"
	"regexp"
	"slices"
	"strconv"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Art ist die JSON-Art, die ein Schlüssel tragen muss.
type Art string

const (
	ArtZeichenkette Art = "eine Zeichenkette"
	ArtWortliste    Art = "eine Liste von Zeichenketten"
	ArtZahl         Art = "eine Zahl"
	ArtWahrheit     Art = "ein Wahrheitswert"
	ArtObjekt       Art = "ein Objekt"

	// ArtListe ist eine Liste von **Objekten** — anders als [ArtWortliste].
	//
	// **Sie kam mit `rulings`** (#888): eine Liste aus `{date, text}`. Als
	// Wortliste geprüft fiele sie durch, als Objekt ebenso — beides meldete
	// einen Befund über eine Datei, die in Ordnung ist.
	ArtListe Art = "eine Liste von Objekten"
)

// bekannteSchluessel sagt, welche Art ein Kartenschlüssel trägt.
//
// # Woher die Tabelle stammt, und warum sie keine Erfindung ist
//
// **Ein maschinenlesbares Schema gibt es nicht.** Nachgesehen am 2026-08-15:
// weder mtgjson.com noch `mtgjson/mtgjson` auf GitHub veröffentlichen eines. Die
// Datenmodelle stehen als Dokumentation da (`mtgjson.com/data-models/card/card-set/`).
//
// Die Tabelle ist deshalb aus zwei Quellen gebaut — der Dokumentation und einer
// **bezogenen** Datei — und sie wird gegen die bezogene Datei gehalten:
// `fremdpakete_test.go` prüft, dass jeder Schlüssel, den eine echte Datei führt,
// dort die Art hat, die hier steht. Wäre sie erfunden, fiele die echte Datei
// durch.
//
// **Unbekannte Schlüssel sind kein Befund.** MTGJSON wächst; eine Prüfung, die
// jeden neuen abwiese, wäre nach dem nächsten Erscheinen rot, ohne dass sich hier
// etwas geändert hätte.
var bekannteSchluessel = map[string]Art{
	"artist":            ArtZeichenkette,
	"artistIds":         ArtWortliste,
	"availability":      ArtWortliste,
	"borderColor":       ArtZeichenkette,
	"boosterTypes":      ArtWortliste,
	"colorIdentity":     ArtWortliste,
	"colors":            ArtWortliste,
	"convertedManaCost": ArtZahl,
	"finishes":          ArtWortliste,
	"flavorText":        ArtZeichenkette,
	"frameVersion":      ArtZeichenkette,
	"identifiers":       ArtObjekt,
	"isReserved":        ArtWahrheit,
	"keywords":          ArtWortliste,
	"language":          ArtZeichenkette,
	"layout":            ArtZeichenkette,
	"legalities":        ArtObjekt,
	"loyalty":           ArtZeichenkette,
	"manaCost":          ArtZeichenkette,
	"manaValue":         ArtZahl,
	"name":              ArtZeichenkette,
	"number":            ArtZeichenkette,
	"originalText":      ArtZeichenkette,
	"originalType":      ArtZeichenkette,
	"power":             ArtZeichenkette,
	"printings":         ArtWortliste,
	"purchaseUrls":      ArtObjekt,

	// **Die vier für mehrseitige Karten und Regelhinweise** (#888).
	//
	// Ohne sie ginge eine falsche Art unbemerkt durch — „unbekannte Schlüssel
	// sind kein Befund". Das ist für Schlüssel richtig, die niemand benutzt;
	// für die, die dieses Programm **liest**, wäre es eine Lücke.
	"rulings":      ArtListe,
	"otherFaceIds": ArtWortliste,
	"faceName":     ArtZeichenkette,
	"side":         ArtZeichenkette,
	"rarity":       ArtZeichenkette,
	"setCode":      ArtZeichenkette,
	"subtypes":     ArtWortliste,
	"supertypes":   ArtWortliste,
	"text":         ArtZeichenkette,
	"toughness":    ArtZeichenkette,
	"type":         ArtZeichenkette,
	"types":        ArtWortliste,
	"uuid":         ArtZeichenkette,
}

// Schluesselarten gibt die Tabelle der bekannten Schlüssel.
//
// Exportiert für den Lauf hinter dem Build-Tag `fremdpakete`: er hält sie gegen
// eine **bezogene** Datei und ist damit der Beleg, dass sie keine Erfindung ist.
func Schluesselarten() map[string]Art {
	kopie := make(map[string]Art, len(bekannteSchluessel))
	for feld, art := range bekannteSchluessel {
		kopie[feld] = art
	}

	return kopie
}

// PasstZurArt sagt, ob ein roher JSON-Wert die erwartete Art hat.
func PasstZurArt(roh json.RawMessage, art Art) bool { return passt(roh, art) }

// manakostenForm ist die Gestalt, die `manaCost` in MTG JSON hat.
//
// Jedes Symbol steht in geschweiften Klammern: `{2}{W}{U/B}`. **Das ist der
// Fall, den der Rundlauf nicht finden kann** — wer die Rückwandlung vergisst,
// schreibt `2WU/B`, liest es wieder als `2WU/B` und ist grün.
var manakostenForm = regexp.MustCompile(`^(\{[^{}]+\})+$`)

// mseSeltenheiten sind die Namen, die **nur** MSE kennt.
//
// `card_fields` führt `mythic rare`, `basic land` und `masterpiece` als
// Auswahlmöglichkeiten; MTG JSON schreibt an ihrer Stelle `mythic` beziehungsweise
// `common`. Steht einer davon in der Ausgabe, ist die Rückwandlung nicht
// gelaufen.
var mseSeltenheiten = []string{"mythic rare", "basic land", "masterpiece"}

// Pruefe hält MTG JSON gegen das, was eine echte Datei zeigt.
//
// Sie gibt Befunde statt eines Fehlers: eine Datei kann an einer Karte fehlerhaft
// sein und an 300 nicht, und wer sie deshalb ganz abwiese, hülfe niemandem (§8).
func Pruefe(inhalt []byte, name string) []domain.Befund {
	sets, err := setsRoh(inhalt)
	if err != nil {
		// **Auch dieser Befund nennt seine Stelle** (§8) — das Feld, an dem
		// die Gestalt scheiterte. Hier stand keine (Fund der Fuzzkampagne vom
		// 2026-10-10, C02, #2227).
		return []domain.Befund{{Datei: name, Ort: err.Field, Meldung: err.Error()}}
	}

	// **Die Fassung zuerst** (#888): steht sie auf einer Hauptfassung, gegen
	// die hier niemand gelesen hat, ist das der Anlass, alles Folgende genauer
	// anzusehen. Erst danach zu melden hieße, den Rahmen hinter die Einzelfälle
	// zu stellen.
	_, befunde := Fassung(inhalt, name)

	for _, roh := range sets {
		befunde = append(befunde, pruefeSet(roh, name)...)
	}

	return befunde
}

// setsRoh holt die Sets als rohe Objekte — ohne die Deutung der Felder aus
// `setsAus`, aber mit seiner Regel, was ein einzelnes Set ist.
func setsRoh(inhalt []byte) ([]rohsatz, *domain.FieldError) {
	var umschlag struct {
		Daten json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(inhalt, &umschlag); err != nil {
		return nil, domain.NewFieldError("inhalt", domain.SchluesselValidierung,
			"Das ist kein JSON: %v", err)
	}

	// **Ohne Umschlag gilt der Inhalt selbst** — dieselbe Deutung wie in
	// [setsAus] (Zweitblick über Phase V, 2026-08-24): eine von Hand
	// geschriebene Datei hat den Umschlag nicht, und `docs/FORMATE.md` führt
	// drei Gestalten, von denen zwei ohne ihn auskommen.
	kern := umschlag.Daten
	if kern == nil {
		kern = inhalt
	}

	var eines map[string]json.RawMessage
	if err := json.Unmarshal(kern, &eines); err != nil {
		return nil, domain.NewFieldError("data", domain.SchluesselValidierung,
			"»data« ist weder ein Set noch eine Abbildung von Sets: %v", err)
	}

	// **Ein einzelnes Set erkennt man an `code` oder `cards`**, nicht allein an
	// `cards`: ein Set ohne Karten ist ein Befund, keine Abbildung von Sets. Der
	// erste Anlauf hielt es für eine — und meldete dann, dass „code" kein Set
	// sei, statt dass die Karten fehlen.
	// **Und nach derselben Regel wie der Leser** ([setsAus]): er dekodiert in
	// eine Struktur, und dort gilt ein Schlüssel auch in anderer Schreibung —
	// „Code" ist `code`. Las die Prüfung eine Abbildung von Sets, wo der Leser
	// ein Set las, meldete sie Dinge über eine Gestalt, die niemand gelesen
	// hatte (Fund der Fuzzkampagne vom 2026-10-10, C02, #2227).
	var alsSet rohSet
	leserSiehtEinSet := json.Unmarshal(kern, &alsSet) == nil && alsSet.Code != ""

	_, hatCode := eines["code"]
	if _, hatKarten := eines["cards"]; hatKarten || hatCode || leserSiehtEinSet {
		// **Der Schlüssel bleibt leer**: hier steht das Set unmittelbar unter
		// »data«, es gibt keinen. [ortVon] fällt dann auf »code« zurück, und
		// hat es auch das nicht, auf den benannten Platzhalter.
		return []rohsatz{{felder: eines}}, nil
	}

	codes := make([]string, 0, len(eines))
	for code := range eines {
		codes = append(codes, code)
	}

	slices.Sort(codes)

	sets := make([]rohsatz, 0, len(codes))

	for _, code := range codes {
		var set map[string]json.RawMessage
		if err := json.Unmarshal(eines[code], &set); err != nil {
			return nil, domain.NewFieldError("data", domain.SchluesselValidierung,
				"%q ist kein Set: %v", code, err)
		}

		sets = append(sets, rohsatz{schluessel: code, felder: set})
	}

	return sets, nil
}

// rohsatz ist ein Set samt dem Schlüssel, unter dem es in der Datei steht.
//
// **Der Schlüssel gehört dazu, weil `code` fehlen darf.** Ein Befund ohne Ort
// sagt nicht, welches von dreihundert Sets gemeint ist — und §8 verlangt die
// Stelle. Gefunden vom ersten Fuzzlauf über `LiesInhalt` am 2026-08-29, an der
// Eingabe `{"":{}}`
// (#1289).
type rohsatz struct {
	schluessel string
	felder     map[string]json.RawMessage
}

// pruefeSet prüft ein Set samt seinen Karten.
func pruefeSet(satz rohsatz, name string) []domain.Befund {
	roh := satz.felder
	code := ortVon(satz)

	var befunde []domain.Befund

	if roh["cards"] == nil {
		return append(befunde, domain.Befund{
			Datei: name, Ort: code,
			Meldung: "das Set führt kein »cards«",
		})
	}

	var karten []map[string]json.RawMessage
	if err := json.Unmarshal(roh["cards"], &karten); err != nil {
		return append(befunde, domain.Befund{
			Datei: name, Ort: code,
			Meldung: "»cards« ist keine Liste von Karten",
		})
	}

	for nummer, karte := range karten {
		befunde = append(befunde, pruefeKarte(karte, name, code, nummer+1)...)
	}

	return befunde
}

// pruefeKarte prüft eine einzelne Karte.
func pruefeKarte(karte map[string]json.RawMessage, name, code string, nummer int) []domain.Befund {
	ort := code + " → Karte " + strconv.Itoa(nummer)

	var befunde []domain.Befund

	melde := func(feld, meldung string) {
		befunde = append(befunde, domain.Befund{Datei: name, Ort: ort + " → " + feld, Meldung: meldung})
	}

	if _, hatNamen := karte["name"]; !hatNamen {
		melde("name", "eine Karte ohne Namen — MTG JSON verlangt ihn")
	}

	schluessel := make([]string, 0, len(karte))
	for feld := range karte {
		schluessel = append(schluessel, feld)
	}

	slices.Sort(schluessel)

	for _, feld := range schluessel {
		art, bekannt := bekannteSchluessel[feld]
		if !bekannt {
			continue
		}

		if !passt(karte[feld], art) {
			melde(feld, "erwartet ist "+string(art))

			continue
		}

		if befund := pruefeWert(feld, textVon(karte[feld])); befund != "" {
			melde(feld, befund)
		}
	}

	return befunde
}

// pruefeWert prüft die Werte, deren Form über die Art hinaus festliegt.
func pruefeWert(feld, wert string) string {
	switch {
	case feld == "manaCost" && wert != "" && !manakostenForm.MatchString(wert):
		return "jedes Symbol steht in geschweiften Klammern (»{2}{R}«), hier steht »" + wert + "«"

	case feld == "rarity" && slices.Contains(mseSeltenheiten, wert):
		return "»" + wert + "« ist eine Seltenheit aus MSE, nicht aus MTG JSON"
	}

	return ""
}

// passt sagt, ob ein roher Wert die erwartete Art hat.
func passt(roh json.RawMessage, art Art) bool {
	switch art {
	case ArtZeichenkette:
		var wert string

		return json.Unmarshal(roh, &wert) == nil

	case ArtWortliste:
		var wert []string

		return json.Unmarshal(roh, &wert) == nil

	case ArtZahl:
		var wert float64

		return json.Unmarshal(roh, &wert) == nil

	case ArtWahrheit:
		var wert bool

		return json.Unmarshal(roh, &wert) == nil

	case ArtListe:
		var wert []map[string]json.RawMessage

		return json.Unmarshal(roh, &wert) == nil

	case ArtObjekt:
		var wert map[string]json.RawMessage

		return json.Unmarshal(roh, &wert) == nil
	}

	return true
}

// textVon liest einen rohen Wert als Zeichenkette; sonst leer.
func textVon(roh json.RawMessage) string {
	var wert string
	if json.Unmarshal(roh, &wert) != nil {
		return ""
	}

	return wert
}
