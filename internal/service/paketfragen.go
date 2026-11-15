package service

import (
	"context"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/render"
)

// Dateiquelle ist eine Bildquelle, die auch nach anderen Dateien gefragt
// werden kann.
//
// **Eine eigene Schnittstelle neben [render.Bildquelle]**, weil die Frage eine
// andere ist: der Rasterer will Bilder, die Vorlage will wissen, ob eine Datei
// dasteht. Wer nur Bilder liefert, antwortet hier nie — und das ist besser als
// ein erfundenes Ja.
type Dateiquelle interface {
	Gibtes(verweis string) bool
}

// **`exists_in_package` fragt, ob eine Datei in einem Paket liegt.**
//
//	if exists_in_package(path) then path
//	else "/magic-modules.mse-include/loyalty/" + name
//
// `magic.mse-game/script`, Zeile 3614 — die Vorlagen wählen damit zwischen
// einem besonderen Bild und einem allgemeinen. **126 Befunde** über die 190
// Stile von `M15-Magic-Pack` lauteten auf diesen Namen
// (#270).
//
// # Warum sie hier steht und nicht in der Bibliothek
//
// Welche Pakete es gibt, weiß die Bildquelle — und `msescript` kennt dieses
// Projekt nicht (§2). Die Bibliothek trägt die Sprache; was die Umgebung
// darüber hinaus weiß, setzt der Aufrufer hinein.
//
// # Woher die Bedeutung stammt
//
// **Aus keiner der beiden portierbaren Quellen.** Weder MSE2 f9d9356 noch
// `fenhl/magic-set-generator` kennen den Namen; das Pack benutzt ihn trotzdem
// — es ist eine Funktion neuerer MSE-Fassungen. Was sie tut, ist an ihrer
// Verwendung **erschlossen**, nicht nachgelesen (§10).

// dateifrage baut die Funktion `exists_in_package` über einer Bildquelle.
//
// **Ohne Quelle ist die Antwort nein.** Ein Planbauer ohne Pakete hat keine
// Dateien; eine erfundene Zusage schickte die Vorlage auf ein Bild, das es
// nicht gibt.
func dateifrage(quelle render.Bildquelle) msescript.Wert {
	frage, _ := quelle.(Dateiquelle) //nolint:errcheck // nil heißt „kann nicht antworten"

	return msescript.Wert{
		Art: msescript.ArtFunktion,
		Funktion: &msescript.Funktionsteil{
			Eingebaut: func(_ context.Context, ruf msescript.Ruf) (msescript.Wert, error) {
				if frage == nil || len(ruf.Stellung) != 1 {
					return msescript.Wahrheit(false), nil
				}

				return msescript.Wahrheit(frage.Gibtes(ruf.Stellung[0].AlsText())), nil
			},
		},
	}
}
