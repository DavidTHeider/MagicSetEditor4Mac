package app

// **Aus welchem Stand dieses Programm gebaut ist** — ENG-06.AK-04
// (#1731).
//
// # Warum neben der Fassung
//
// **Die Fassung benennt eine Freigabe, der Stand einen Bau.** Zwischen zwei
// Freigaben trägt jeder Bau dieselbe Fassung — `0.5.0` stand auch in einem
// Programm, das schon zwei Wanderungen mehr kannte als 0.5.0. Wer einen
// Fehlerbericht liest, braucht beides.
//
// # Und warum nur zur Diagnose
//
// **Eine zweite Produktfassung wäre eine Zusage, die niemand hält** — das
// Issue sagt es ausdrücklich. Der Stand wird deshalb nirgends verglichen,
// nirgends ausgehandelt und in keine Datei geschrieben, die jemand weitergibt;
// er steht dort, wo ein Mensch nachsieht: in der Meta-Auskunft und im
// Diagnosetext.

import "runtime/debug"

// Bau ist der Stand, aus dem dieses Programm gebaut wurde.
type Bau struct {
	// Stand ist der Commit — oder [unbekannterStand], wenn der Bau keinen
	// kennt.
	Stand string `json:"stand"`

	// Gebaut ist die Zeit dieses Commits, nicht die des Baus.
	Gebaut string `json:"gebaut"`

	// Veraendert sagt, dass der Arbeitsbaum beim Bauen nicht sauber war.
	Veraendert bool `json:"veraendert"`
}

// unbekannterStand steht, wo der Bau keinen Stand kennt.
//
// **Ein Wort und kein leerer Text**: ein leeres Feld sieht aus wie ein
// vergessenes. `go test` etwa schreibt keinen Stand in sein Binärprogramm.
const unbekannterStand = "unbekannt"

// DiesenBau sagt, aus welchem Stand dieses Programm gebaut ist.
func DiesenBau() Bau { return bauAus(debug.ReadBuildInfo()) }

// bauAus liest die Angaben der Versionsverwaltung, die `go build` einträgt.
func bauAus(info *debug.BuildInfo, gelesen bool) Bau {
	bau := Bau{Stand: unbekannterStand}

	if !gelesen {
		return bau
	}

	for _, angabe := range info.Settings {
		switch angabe.Key {
		case "vcs.revision":
			bau.Stand = angabe.Value
		case "vcs.time":
			bau.Gebaut = angabe.Value
		case "vcs.modified":
			bau.Veraendert = angabe.Value == "true"
		}
	}

	return bau
}

// bauzeile gibt den Bau für den Diagnosetext — Stand, Zeit und Vermerk.
func bauzeile(bau Bau, texte Dialogtexte) string {
	zeile := texte.DiagnoseBau + ": " + bau.Stand

	if bau.Gebaut != "" {
		zeile += " (" + bau.Gebaut + ")"
	}

	if bau.Veraendert {
		zeile += ", " + texte.DiagnoseVeraendert
	}

	return zeile
}
