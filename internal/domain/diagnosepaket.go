// **Was in ein Diagnosepaket gehört — und was nicht** — Werkstück O10 von
// Phase 5.2 (#1434).
//
// # Wozu eine Aufzählung
//
// Weil die Vorschau und das Paket **dieselbe** Liste sein müssen. Das Issue
// verlangt es wörtlich: „Vorschau stimmt mit ZIP-/Dateiinhalt überein." Zwei
// Listen — eine für die Anzeige, eine fürs Schreiben — liefen beim ersten
// neuen Teil auseinander, und die Zustimmung gälte dann einem anderen Inhalt
// als dem, der die Platte verlässt.
//
// # Und was hier **nicht** steht
//
// Kartentexte, Feldwerte, Bilder, die Sammlung, private Notizen, Token,
// Kennwörter, TOTP-Geheimnisse, OAuth-Antworten. Nicht, weil jemand sie
// vergessen hätte, sondern weil ein Diagnosepaket eine Auskunft über die
// **Installation** ist und über niemanden.

package domain

// Diagnoseteil benennt einen Teil des Pakets.
type Diagnoseteil string

// **Die Übersicht steht nicht darunter**, obwohl sie im Paket liegt: sie ist
// die **Vorschau** und zählt die Teile auf. Stünde sie in ihrer eigenen
// Aufzählung, müsste sie ihre eigene Größe nennen, bevor sie geschrieben ist.
// Was **alles** im Archiv liegt — die Übersicht eingeschlossen —, sagt das
// Manifest mit Größe und Prüfsumme je Eintrag.
const (
	// TeilFassung: Fassung, Plattform, Go-Fassung.
	TeilFassung Diagnoseteil = "fassung"

	// TeilBestand: Ort, Schemastand, verwaiste Zeilen — **kein Inhalt**.
	TeilBestand Diagnoseteil = "bestand"

	// TeilPakete: welche Vorlagenpakete in welchem Stand eingebunden sind.
	TeilPakete Diagnoseteil = "pakete"

	// TeilMetriken: der Zählstand aus O9 — vierzig mögliche Reihen.
	TeilMetriken Diagnoseteil = "metriken"

	// TeilProtokoll: die letzten Zeilen, geschwärzt — und **nur für den
	// Verwalter des gezeigten Profils** (O9): das Protokoll gehört dem
	// Prozess und nicht dem Profil.
	TeilProtokoll Diagnoseteil = "protokoll"
)

// Diagnoseteile gibt alle fünf, in der Reihenfolge des Pakets.
func Diagnoseteile() []Diagnoseteil {
	return []Diagnoseteil{
		TeilFassung, TeilBestand, TeilPakete, TeilMetriken, TeilProtokoll,
	}
}

// Diagnosepaketgrenze ist die Obergrenze eines Pakets in Bytes.
//
// **Acht Megabyte.** Der größte Teil ist das Protokoll, und das wird bei vier
// Megabyte gedreht (`app.Protokollgrenze`); mit seiner Vorfassung und den
// kleinen Teilen daneben liegt ein volles Paket knapp darunter. Die Grenze ist
// damit keine Schikane, sondern die Zusage, dass kein Paket entsteht, das
// niemand mehr verschicken kann — und wenn sie greift, ist etwas anderes
// schiefgegangen.
const Diagnosepaketgrenze = 8 << 20

// Diagnosepaketendung ist die Endung eines Pakets.
//
// **Nicht `.zip`** — das verschweigt, was es ist —, und nicht
// `.msesicherung`: eine Sicherung spielt man zurück, ein Diagnosepaket
// verschickt man. Zwei Formate unter einer Endung wären eine Einladung, das
// eine für das andere zu halten.
const Diagnosepaketendung = ".msediagnose"
