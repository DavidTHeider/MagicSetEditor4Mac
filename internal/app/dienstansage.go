package app

// **Wie ein Dienst sich meldet** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Warum Schreiber und Leser in einer Datei stehen
//
// **Weil es ein Vertrag zwischen zwei Prozessen ist.** `mse-rest` und
// `mse-mcp` schreiben ihn, der Dienstposten im Fenster liest ihn zurück
// ([leseAnsage]). Stünde die Form an drei Stellen, liefe sie beim ersten
// Zusatz auseinander — und der Posten wartete auf eine Zeile, die anders
// heißt, bis die Frist abläuft.
//
// # Und warum über die Ausgabe und nicht über eine Datei
//
// **Der Grund steht in `cmd/mse-rest/bedienen.go`**: eine Datei müsste
// angelegt, gefunden, gelesen und aufgeräumt werden, und sie überlebte den
// Prozess, dem sie gehört.

import (
	"strings"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Dienstansage ist, was ein Dienst über sich sagt.
type Dienstansage struct {
	// MitTLS entscheidet das Schema.
	MitTLS bool

	// Adresse ist Wirt und Port, so wie der Lauscher sie nennt.
	Adresse string

	// Pfad ist der Endpunkt darunter — leer, wo der Dienst an der Wurzel
	// sitzt.
	Pfad string

	// Fingerabdruck ist der des Zeugnisses — leer ohne TLS.
	Fingerabdruck string

	// Aussteller und GueltigBis kennt nur ein eigenes Zeugnis (NET-03,
	// #1707) — leer beim Zeugnis auf der Peeridentität.
	Aussteller string
	GueltigBis time.Time

	// Marke ist der Zugang, den dieser Lauf verlangt.
	Marke string

	// Weitere sind die übrigen Adressen der Lauschwahl, auf denen der Dienst
	// lauscht — G08 (#2272). **Je eine Zeile „Dienst auf" nach der ersten**:
	// wer nur die erste liest, liest weiter dasselbe.
	Weitere []string

	// Fehlende sind gewählte Adressen, auf denen kein Lauscher entstand —
	// je eine Zeile „Ohne".
	Fehlende []domain.Lauschbefund
}

// Text ist die Ansage, so wie [leseAnsage] sie erwartet.
//
// **„Marke" steht zuletzt**, und das ist die Zusage, auf die sich der Leser
// stützt: wer sie gelesen hat, hat alles davor auch gelesen. Ein Zusatz kommt
// deshalb **davor** und nicht dahinter.
func (a Dienstansage) Text() string {
	schema := "http"
	if a.MitTLS {
		schema = "https"
	}

	var gebaut strings.Builder

	gebaut.WriteString("Dienst auf " + schema + "://" + a.Adresse + a.Pfad + "\n")

	for _, weitere := range a.Weitere {
		gebaut.WriteString("Dienst auf " + schema + "://" + weitere + a.Pfad + "\n")
	}

	for _, fehlende := range a.Fehlende {
		gebaut.WriteString("Ohne " + fehlende.Adresse + ": " + fehlende.Grund + "\n")
	}

	// **Ohne TLS keine Zeile** statt einer leeren: eine Zeile
	// „Fingerabdruck " ohne Inhalt sähe aus wie ein Zeugnis, dessen Abdruck
	// verlorenging.
	if a.Fingerabdruck != "" {
		gebaut.WriteString("Fingerabdruck " + a.Fingerabdruck + "\n")
	}

	// **Und ein eigenes Zeugnis nennt seine Stelle und sein Ende** — vor der
	// Marke, denn nach ihr hört der Leser auf.
	if a.Aussteller != "" {
		gebaut.WriteString("Zeugnis von " + a.Aussteller + " bis " +
			a.GueltigBis.UTC().Format(time.RFC3339) + "\n")
	}

	gebaut.WriteString("Marke " + a.Marke + "\n")

	return gebaut.String()
}
