// **Die Angaben eines Profils** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkte 10 bis 13 der Kommentierung vom 2026-09-05.
//
// # Warum ein Profil eigene Angaben hat und nicht die des Menschen erbt
//
// **Weil es eine Rolle ist.** Wer ein Profil „Turnier" führt und eines
// „Sammlung", tut darin verschiedene Dinge — und die Beschreibung sagt,
// welche. Was über den **Menschen** gilt, steht in [Nutzer] und gilt in
// beiden; was hier steht, unterscheidet sie.
//
// # Und warum der Slogan hier »Subslogan« heißt
//
// **Weil er unter dem des Menschen steht und ihn nicht ersetzt.** Die
// Kommentierung nennt ihn so, und der Name hält die Richtung fest: zwei Felder
// namens `slogan` an zwei Tabellen wären beim Lesen nicht zu unterscheiden.

package domain

import "strings"

// Die Grenzen der Profilangaben.
//
// **Kürzer als beim Menschen, und das ist Absicht.** Ein Subslogan steht
// **neben** dem Slogan des Menschen in derselben Zeile; zwei Sätze in gleicher
// Länge machten daraus einen Absatz.
const (
	// HoechstlaengeSubslogan ist die Zeile unter dem Profilnamen.
	HoechstlaengeSubslogan = 120

	// HoechstlaengeKurzebeschreibung ist der Absatz, der auf eine Karte passt.
	HoechstlaengeKurzebeschreibung = 500

	// HoechstlaengeLangebeschreibung ist der Text, den niemand überfliegt.
	HoechstlaengeLangebeschreibung = 5000
)

// pruefeDateiname weist einen Wert ab, der ein Pfad ist.
//
// **Eine Stelle für zwei Felder** — `nutzer.bild` und `profile.badge`. Beide
// werden an ein Verzeichnis gehängt, und ein `../` darin läse eine Datei
// außerhalb; zwei Fassungen dieser Prüfung liefen beim ersten Sonderfall
// auseinander.
//
// **Leer ist erlaubt**: „kein Bild" ist eine Antwort.
func pruefeDateiname(feld, wert string) error {
	if wert == "" {
		return nil
	}

	if strings.ContainsAny(wert, `/\`) || wert == ".." || wert == "." {
		return NewFieldError(feld, SchluesselValidierung,
			"Ein Bildname ist ein Dateiname und kein Pfad; %q geht nicht.", wert)
	}

	return nil
}

// pruefeProfilangaben prüft die vier freiwilligen Angaben eines Profils.
func (p Profil) pruefeProfilangaben() error {
	for _, zeile := range []struct {
		feld       string
		wert       string
		hoechstens int
	}{
		{"subslogan", p.Subslogan, HoechstlaengeSubslogan},
		{"kurzebeschreibung", p.Kurzebeschreibung, HoechstlaengeKurzebeschreibung},
		{"langebeschreibung", p.Langebeschreibung, HoechstlaengeLangebeschreibung},
	} {
		if fehler := pruefeAngabe(zeile.feld, zeile.wert, zeile.hoechstens); fehler != nil {
			return fehler
		}
	}

	if fehler := pruefeDateiname("badge", p.Badge); fehler != nil {
		return fehler
	}

	return p.pruefeAblauf()
}

// pruefeAblauf weist einen Ablauf ab, den es so nicht geben soll.
//
// **Zwei Fälle**: ein Ablauf an einem persönlichen Profil, und einer, der kein
// Zeitpunkt ist. Der erste ist die Regel aus Werkstück 10 (#1351), der zweite
// dieselbe Prüfung wie beim Token.
func (p Profil) pruefeAblauf() error {
	if strings.TrimSpace(p.Ablauf) == "" {
		return nil
	}

	if !p.Technisch {
		return NewFieldError("ablauf", SchluesselValidierung,
			"Nur ein technisches Profil läuft ab — ein Mensch hört nicht auf, "+
				"dieser Mensch zu sein.")
	}

	return Zugriffstoken{Ablauf: p.Ablauf}.pruefeAblauf()
}
