package app

// **Wer bin ich: Fassung, Plattform, Fläche** — Werkstück O9 von Phase 5.2
// (#1433).
//
// # Warum es das braucht
//
// Das Issue verlangt „Liveness, Readiness, **capabilities/meta** und optionale
// Metrikausgabe". Die Bereitschaft sagt, ob ein Prozess **arbeiten** kann; sie
// sagt nicht, **welcher** er ist. Wer einen Fehlerbericht liest oder zwei
// Prozesse nebeneinander laufen hat, braucht genau das zuerst.
//
// # Und die andere Hälfte des Wortes
//
// **„capabilities" ist nicht hier, sondern in O6** — dem Kompatibilitäts­manifest
// (#1430).
// Was eine Fläche kann, sagt sie heute schon selbst: `tools/list` über MCP,
// `mse hilfe` an der Kommandozeile, `GET /grenzen` und `GET /dienstkatalog` am
// Dienst. Eine fünfte Liste daneben wäre eine, die veraltet — und die Frage
// „passt mein Klient zu dieser Fassung?" ist eine Aushandlung und keine
// Auskunft. Sie steht deshalb im nächsten Werkstück und nicht in diesem.

import (
	"context"
	"runtime"
)

// MetaAntwort sagt, mit welchem Programm jemand spricht.
type MetaAntwort struct {
	// Fassung ist die Fassung dieser Anwendung — dieselbe, die das Bündel
	// trägt und die im Diagnosetext oben steht.
	Fassung string `json:"fassung"`

	// Plattform ist Betriebssystem und Rechnerart, `darwin/arm64` etwa.
	Plattform string `json:"plattform"`

	// Go ist die Fassung der Laufzeit — sie steht im Fehlerbericht neben der
	// Fassung, weil ein Unterschied dort einen Unterschied im Verhalten
	// erklären kann.
	Go string `json:"go"`

	// Flaeche ist die Fläche, über die **dieser** Ruf kam: `fenster`,
	// `kommandozeile`, `mcp` oder `dienst`.
	//
	// **Sie steht hier, weil sie nur hier zu erfahren ist.** Wer zwei
	// Prozesse an einem Bestand betreibt, sieht sonst nicht, welchen er
	// gerade fragt.
	Flaeche string `json:"flaeche"`

	// Bau ist der Stand, aus dem dieses Programm gebaut ist — zur Diagnose
	// und nicht als zweite Fassung (ENG-06.AK-04, [Bau]).
	Bau Bau `json:"bau"`

	// Bindungsvertrag ist der Abdruck dessen, was zwischen Fenster und Go
	// über den Draht geht ([Vertragsabdruck]).
	Bindungsvertrag string `json:"bindungsvertrag"`

	Fehler *Fehler `json:"fehler"`
}

// Meta sagt, welches Programm antwortet.
//
// **Ohne Bestand und ohne Naht**: alles hier steht in der Binärdatei. Ein Weg,
// der dafür eine Datenbank bräuchte, könnte die Frage genau dann nicht
// beantworten, wenn sie am nötigsten ist — beim Bestand, der nicht aufgeht.
func (a *App) Meta(ctx context.Context) MetaAntwort {
	ctx = a.imProfil(ctx)

	return MetaAntwort{
		Fassung:   Fassung,
		Plattform: runtime.GOOS + "/" + runtime.GOARCH,
		Go:        runtime.Version(),
		Flaeche:   string(DienstmodusAus(ctx)),

		Bau: DiesenBau(), Bindungsvertrag: Vertragsabdruck(),
	}
}
