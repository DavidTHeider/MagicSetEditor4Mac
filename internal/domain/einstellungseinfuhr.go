// **Was aus einem Austausch angenommen wird** — Werkstück O3 von Phase 5.2
// (#1427).
//
// # Die Regel in einem Satz
//
// **Ein Schlüssel steht im Abschnitt seiner Ebene — sonst ist er ein Befund.**
// Das ist die Stelle, an der die Ebenentabelle arbeitet: ohne sie wäre
// „Abschnitt" nur eine Überschrift, und eine Datei könnte den Pfad dieser
// Platte unter `profil` unterbringen.
//
// # Warum hier nichts mit dem Bestand verglichen wird
//
// **Weil das eine andere Frage ist.** Hier steht, was **überhaupt** angenommen
// werden darf — eine Aussage über die Datei allein, ohne Bestand, ohne Profil,
// ohne Uhr. Was sich dadurch **ändert**, weiß erst `service`, und was eine
// Schranke lockert, auch: dafür braucht es die andere Ebene.

package domain

import "slices"

// PruefeAustausch urteilt über die Zeilen eines Austauschs — ohne den Bestand.
//
// **Die Reihenfolge ist die der Anzeige**: Abschnitte in der Reihenfolge der
// Ebenen, Zeilen in der Reihenfolge ihrer Reihe. Eine Vorschau, die sich
// zwischen zwei Blicken umsortiert, ist keine.
func PruefeAustausch(austausch Austausch) ([]Einfuhrpaar, []Einfuhrbefund) {
	paare := []Einfuhrpaar{}
	befunde := []Einfuhrbefund{}

	for _, ebene := range Ebenen() {
		werte, da := austausch.Bereiche[string(ebene)]
		if !da {
			continue
		}

		ausDiesem, dabei := ausAbschnitt(ebene, werte)
		paare = append(paare, ausDiesem...)
		befunde = append(befunde, dabei...)
	}

	return paare, append(befunde, fremdeAbschnitte(austausch.Bereiche)...)
}

// fremdeAbschnitte meldet, was keine Ebene ist.
//
// **Ein Befund je Abschnitt und nicht je Zeile darin**: der Grund ist derselbe
// für alle, und dreißig gleichlautende Zeilen sagen nicht mehr als eine.
func fremdeAbschnitte(bereiche map[string]map[string]string) []Einfuhrbefund {
	fremde := []string{}

	for name := range bereiche {
		if !KenntEbene(name) {
			fremde = append(fremde, name)
		}
	}

	slices.Sort(fremde)

	befunde := make([]Einfuhrbefund, 0, len(fremde))
	for _, name := range fremde {
		befunde = append(befunde, Einfuhrbefund{
			Ebene: name,
			Grund: "Diesen Abschnitt kennt dieses Programm nicht; " +
				"er wird übergangen.",
		})
	}

	return befunde
}

// ausAbschnitt urteilt über die Zeilen eines Abschnitts.
func ausAbschnitt(ebene Ebene, werte map[string]string) (
	[]Einfuhrpaar, []Einfuhrbefund,
) {
	paare := []Einfuhrpaar{}
	befunde := []Einfuhrbefund{}
	behandelt := map[string]bool{}

	for _, schluessel := range reiheFuer(ebene) {
		wert, da := werte[schluessel]
		if !da {
			continue
		}

		behandelt[schluessel] = true

		if grund := passtInAbschnitt(ebene, schluessel); grund != "" {
			befunde = append(befunde, Einfuhrbefund{
				Ebene: string(ebene), Schluessel: schluessel, Grund: grund,
			})

			continue
		}

		if fehler := pruefeAufEbene(ebene, schluessel, wert); fehler != nil {
			befunde = append(befunde, Einfuhrbefund{
				Ebene: string(ebene), Schluessel: schluessel,
				Grund: fehler.Error(),
			})

			continue
		}

		paare = append(paare, Einfuhrpaar{
			Ebene: ebene, Schluessel: schluessel, Wert: wert,
		})
	}

	return paare, append(befunde, uebrige(ebene, werte, behandelt)...)
}

// uebrige meldet, was in diesem Abschnitt steht und in keiner Reihe.
func uebrige(
	ebene Ebene, werte map[string]string, behandelt map[string]bool,
) []Einfuhrbefund {
	fremde := []string{}

	for schluessel := range werte {
		if !behandelt[schluessel] {
			fremde = append(fremde, schluessel)
		}
	}

	slices.Sort(fremde)

	befunde := make([]Einfuhrbefund, 0, len(fremde))
	for _, schluessel := range fremde {
		befunde = append(befunde, Einfuhrbefund{
			Ebene: string(ebene), Schluessel: schluessel,
			Grund: "Diesen Schlüssel kennt dieses Programm auf dieser Ebene " +
				"nicht; er wird übergangen.",
		})
	}

	return befunde
}

// reiheFuer gibt die Reihe, aus der ein Abschnitt seine Schlüssel nimmt.
//
// **Die Ebene `geraet` bekommt die Profilreihe** und nicht etwa keine: ihre
// Schlüssel stehen dort, und sie sollen einen **benannten** Befund bekommen
// („diese Ebene reist nicht mit") statt des allgemeinen („kennt dieses
// Programm nicht"). Ein Befund, der den Grund verfehlt, schickt jemanden an
// die falsche Stelle.
func reiheFuer(ebene Ebene) []string {
	if ebene == EbeneNutzer {
		return Nutzereinstellungsreihe()
	}

	return Einstellungsreihe()
}

// passtInAbschnitt gibt den Grund, warum ein Schlüssel hier nicht steht —
// leer, wenn er passt.
//
// # Die Regel in einem Satz
//
// **Ein Abschnitt nimmt die Schlüssel seiner Ebene** — und der Abschnitt
// `profil` zusätzlich die des **Menschen**, weil dort seine Übersteuerung
// steht („hier arbeite ich auf Englisch"). Alles andere ist ein Befund.
func passtInAbschnitt(ebene Ebene, schluessel string) string {
	if ebene == EbeneGeraet {
		return "Diese Ebene beschreibt einen Rechner; sie reist nicht mit."
	}

	ebeneDesSchluessels, _ := EbeneVon(schluessel)

	if ebeneDesSchluessels == ebene ||
		(ebene == EbeneProfil && ebeneDesSchluessels == EbeneNutzer) {
		return ""
	}

	return "Dieser Schlüssel gehört nicht in den Abschnitt »" +
		string(ebene) + "«, sondern auf die Ebene »" +
		string(ebeneDesSchluessels) + "«."
}

// pruefeAufEbene urteilt über den Wert — mit der Regel der Reihe, aus der der
// Schlüssel kommt.
func pruefeAufEbene(ebene Ebene, schluessel, wert string) error {
	if ebene == EbeneNutzer {
		return PruefeNutzereinstellung(schluessel, wert)
	}

	return PruefeEinstellung(schluessel, wert)
}
