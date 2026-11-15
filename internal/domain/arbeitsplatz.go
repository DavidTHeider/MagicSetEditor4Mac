// **Wie das Fenster steht** — Werkstück O3 von Phase 5.2
// (#1427),
// nach [Vermerk 0079](../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md).
//
// # Warum das eine Einstellung ist und keine eigene Bindung
//
// **Weil es dieselbe Frage ist wie »welche Schriftgröße«** — nur über das
// Fenster statt über die Schrift. Zwei neue Bindungen hießen zwei neue Wege an
// jeder der vier Flächen (§16), und beide täten genau das, was
// `EinstellungSetzen` schon tut. Was hier dazukommt, sind **Schlüssel**, kein
// Mechanismus.
//
// # Warum sie trotzdem eine eigene Ebene haben
//
// **Weil sie anders reisen.** Ein Arbeitsplatz hängt an einem Bildschirm: wer
// auf einem 13-Zoll-Gerät einen schmalen Navigator eingestellt hat, will ihn
// auf einem 32-Zoll-Schirm nicht wiederhaben. Deshalb steht die Ebene
// `arbeitsplatz` im Austausch **nur auf Wunsch**.
//
// # Wer sie benutzt
//
// **Noch niemand — Werkstück U4** (#1415).
// Hier stehen die Schlüssel, ihre Vorgaben und die Prüfung; das Fenster liest
// sie dort über die Einstellungen, die es ohnehin lädt. Dass hier eine Form
// steht und noch keine Fähigkeit, ist gesagt und nicht verschwiegen (§10,
// erste Dimension).

package domain

import "strconv"

// Die vier Schlüssel des Arbeitsplatzes.
const (
	// ArbeitsplatzNavigator sagt, ob die linke Spalte steht.
	ArbeitsplatzNavigator = "arbeitsplatz.navigator"

	// ArbeitsplatzInspektor sagt, ob die rechte Spalte steht.
	ArbeitsplatzInspektor = "arbeitsplatz.inspektor"

	// ArbeitsplatzNavigatorbreite ist die Breite der linken, in Punkten.
	ArbeitsplatzNavigatorbreite = "arbeitsplatz.navigatorbreite"

	// ArbeitsplatzInspektorbreite ist die der rechten.
	ArbeitsplatzInspektorbreite = "arbeitsplatz.inspektorbreite"
)

// Die zwei Stellungen einer Spalte.
const (
	SpalteSichtbar  = "sichtbar"
	SpalteVerborgen = "verborgen"
)

// Die Grenzen einer Spaltenbreite, in Punkten.
//
// **Beide Zahlen sind eine Entscheidung und keine Messung.** Unter 120 Punkten
// bricht ein Setname in jeder Zeile um, und die Spalte zeigt Silben statt
// einer Liste; über 800 bleibt auf einem 1280 Punkte breiten Fenster für die
// Karte weniger übrig als für ihre Nachbarspalte — und die Karte ist das
// Hauptobjekt (§6). Wer sie ändert, ändert eine Zusage über die Bedienbarkeit;
// deshalb stehen sie hier und nicht im Fenster.
const (
	BreiteMindestens = 120
	BreiteHoechstens = 800
)

// Die Vorgabebreiten, beim Menschen.
//
// **Die Zahlen kommen aus Vermerk 0076**, nicht von hier: 240 Punkte für den
// Navigator, 320 für den Inspektor — so setzt sie der Baustein
// `Arbeitsplatz` (Werkstück U2), und so stehen sie in `UX-ARCHITEKTUR.md`.
// Der Navigator ist schmaler, weil er Namen führt und jener Felder mit
// Beschriftung. Wer hier eine andere Zahl will, ändert den Vermerk — eine
// Vorgabe, die vom gebauten Baustein abweicht, wäre zwei Zusagen für eine
// Frage.
const (
	BreiteNavigatorVorgabe = "240"
	BreiteInspektorVorgabe = "320"
)

// pruefeBreite urteilt über eine Spaltenbreite.
//
// **Über eine Zeichenkette und nicht über eine Zahl**, weil eine Einstellung
// eine Zeichenkette ist: die Tafel hält Text. Wer hier eine Zahl nähme, hätte
// die Umwandlung an der Naht statt an der Regel — und eine Naht urteilt nicht.
func pruefeBreite(wert string) error {
	punkte, fehler := strconv.Atoi(wert)
	if fehler != nil {
		return NewFieldError("wert", SchluesselValidierung,
			"%q ist keine Breite — erwartet wird eine ganze Zahl in Punkten.",
			wert)
	}

	if punkte < BreiteMindestens || punkte > BreiteHoechstens {
		return NewFieldError("wert", SchluesselValidierung,
			"%d Punkte liegen außerhalb von %d bis %d.",
			punkte, BreiteMindestens, BreiteHoechstens)
	}

	return nil
}

// pruefeBreiteOderErben lässt zusätzlich „wie der Nutzer" zu.
//
// **Auf der Profilebene und nur dort**: der Mensch ist der Boden, und wer dort
// erben könnte, zeigte auf sich selbst (`uebersteuerung.go`).
func pruefeBreiteOderErben(wert string) error {
	if wert == Erben {
		return nil
	}

	return pruefeBreite(wert)
}

// arbeitsplatzschluessel gibt die vier — für die Ebenentabelle nebenan.
func arbeitsplatzschluessel() []string {
	return append([]string{
		ArbeitsplatzNavigator, ArbeitsplatzInspektor,
		ArbeitsplatzNavigatorbreite, ArbeitsplatzInspektorbreite,
	}, ansichtsschluessel()...)
}

// haengeArbeitsplatzAn hängt die vier Schlüssel an beide Reihen.
//
// **In beide**, weil das Layout je Profil gilt und vom Menschen erbt: wer zwei
// Profile hat und in beiden gleich arbeitet, sagt es einmal. Dieselbe Bauform
// wie bei den vier Erscheinungseinstellungen — und derselbe Grund: eine
// Antwort so oft zu geben, wie es Profile gibt, ist keine Antwort.
//
// **Gerufen aus dem `init` von `einstellung.go` und nicht aus einem eigenen.**
// Go führt die `init` einer Datei in der Reihenfolge aus, in der der Übersetzer
// die Dateien bekommt — dem Dateinamen nach. Eine Anzeigereihenfolge, die an
// einem Dateinamen hängt, ändert sich beim Umbenennen, und niemand sieht es.
func haengeArbeitsplatzAn() {
	for _, schluessel := range []string{ArbeitsplatzNavigator, ArbeitsplatzInspektor} {
		einstellungen = append(einstellungen, einstellung{
			Schluessel: schluessel,
			Wirkung:    WirkungSofort,
			Bereich:    BereichArbeitsplatz,
			Werte:      []string{Erben, SpalteSichtbar, SpalteVerborgen},
			Vorgabe:    Erben,
		})

		nutzereinstellungen = append(nutzereinstellungen, nutzereinstellung{
			Schluessel: schluessel,
			Werte:      []string{SpalteSichtbar, SpalteVerborgen},
			Vorgabe:    SpalteSichtbar,
		})
	}

	for _, breite := range []struct{ schluessel, vorgabe string }{
		{ArbeitsplatzNavigatorbreite, BreiteNavigatorVorgabe},
		{ArbeitsplatzInspektorbreite, BreiteInspektorVorgabe},
	} {
		// **Ohne Werteliste**: eine Breite entsteht am Ziehgriff, und eine
		// Aufzählung von 681 Punkten wäre keine Wahl, sondern eine Liste.
		// Dasselbe Muster wie beim Hochladepfad — nur mit einer Regel statt
		// ohne.
		einstellungen = append(einstellungen, einstellung{
			Schluessel: breite.schluessel,
			Wirkung:    WirkungSofort,
			Bereich:    BereichArbeitsplatz,
			Vorgabe:    Erben,
			Pruefung:   pruefeBreiteOderErben,
		})

		nutzereinstellungen = append(nutzereinstellungen, nutzereinstellung{
			Schluessel: breite.schluessel,
			Vorgabe:    breite.vorgabe,
			Pruefung:   pruefeBreite,
		})
	}
}
