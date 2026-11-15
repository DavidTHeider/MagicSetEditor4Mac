// **Ein Schreibfehler in der Nachbardatei lässt das Paket unberührt** — der
// Befund des Zweitblicks über Phase 4.2 zu
// #1289.
//
// # Der Befund, wörtlich
//
// „Injiziert werden ausschließlich die drei Gestalten der Umbenennung. Fehler
// aus `Write`, `Sync` und `Close` in `ablegen` bleiben unbetreten; der Test
// erklärt sie ausdrücklich für ungeprüft."
//
// Er hat recht, und die Begründung von damals — *auf einer gesunden Platte
// nicht herbeizuführen* — war eine Aussage über `os.CreateTemp`, nicht über den
// Weg. Das Werkstück heißt „**abgebrochenes Schreiben**"; ein Lauf, der nur
// Umbenennungen abbricht, löst seine eigene Abnahme nicht ein.
//
// # Warum die Abdeckung das nicht gezeigt hat
//
// Weil sie Anweisungen zählt und keine Fälle. `fehler = schliessfehler` steht
// in `if schliessfehler := neben.Close(); fehler == nil` — die Zeile **läuft**
// bei jedem geglückten Schreiben. Gezählt ist sie damit; betreten wurde sie
// nie mit einem Fehler darin.
//
// # Was hier zugesichert wird
//
//   - Der eingespeiste Fehler kommt beim Aufrufer an.
//   - Das Paket steht danach **unverändert** da — jede Datei mit ihren alten
//     Bytes, keine dazu, keine weg.
//   - Keine Nachbardatei bleibt liegen.
//
// # Was nicht
//
// **Der Stromausfall.** `ablegen` ruft [os.File.Sync]; was danach im Journal
// des Dateisystems steht, entscheidet nicht dieser Lauf.
package msepaket

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"
)

// abbruchstelle sagt, an welchem Griff die Nachbardatei versagt.
type abbruchstelle int

const (
	beimAnlegen abbruchstelle = iota
	beimEinschreiben
	beimSynchronisieren
	beimSchliessen
	beimFreigeben
)

func (a abbruchstelle) String() string {
	return [...]string{
		"beim Anlegen", "beim Schreiben",
		"beim Synchronisieren", "beim Schließen",
		"beim Freigeben des Archivs",
	}[a]
}

// versagendeNachbardatei tut das Echte — außer an einer Stelle.
type versagendeNachbardatei struct {
	*os.File

	stelle abbruchstelle
}

func (v *versagendeNachbardatei) Write(teil []byte) (int, error) {
	if v.stelle == beimEinschreiben {
		return 0, errAbbruch
	}

	return v.File.Write(teil) //nolint:wrapcheck // der Griff reicht durch
}

func (v *versagendeNachbardatei) Sync() error {
	if v.stelle == beimSynchronisieren {
		return errAbbruch
	}

	return v.File.Sync() //nolint:wrapcheck // dito
}

func (v *versagendeNachbardatei) Close() error {
	// **Geschlossen wird trotzdem.** Ein Lauf, der den Griff offen ließe,
	// prüfte nicht den Fehlerweg, sondern erzeugte einen zweiten.
	fehler := v.File.Close()
	if v.stelle == beimSchliessen {
		return errAbbruch
	}

	return fehler //nolint:wrapcheck // dito
}

// versagenderBauer gibt einen Nachbarn, der an der genannten Stelle versagt.
func versagenderBauer(stelle abbruchstelle) func(string) (nachbardatei, error) {
	return func(verzeichnis string) (nachbardatei, error) {
		if stelle == beimAnlegen {
			return nil, errAbbruch
		}

		datei, err := echterNachbar(verzeichnis)
		if err != nil {
			return nil, err
		}

		// **Der Typwechsel ist hier sicher**, weil [echterNachbar] genau eine
		// Gestalt gibt: eine `*os.File` aus [os.CreateTemp]. Ein Lauf, der
		// stattdessen die Schnittstelle einwickelte, verlöre `Name()` — und
		// den braucht der Austausch.
		roh, passt := datei.(*os.File)
		if !passt {
			return nil, errAbbruch
		}

		return &versagendeNachbardatei{File: roh, stelle: stelle}, nil
	}
}

// hilfenMitAbbruch gibt die Griffe, von denen einer an der Stelle versagt.
//
// **Das Freigeben des Archivs betrifft nur die Archivform.** Ein
// Verzeichnispaket hält nichts offen; dort geht der Lauf dann durch, und das
// ist keine Lücke, sondern die Auskunft, dass es diesen Griff nicht gibt.
func hilfenMitAbbruch(stelle abbruchstelle) schreibhilfen {
	hilfen := echteHilfen()
	hilfen.nachbar = versagenderBauer(stelle)

	if stelle == beimFreigeben {
		hilfen.nachbar = echterNachbar
		hilfen.schliessen = func(io.Closer) error { return errAbbruch }
	}

	return hilfen
}

// **Ein Fehlschlag an der Nachbardatei lässt das Paket, wie es war.**
func TestEinSchreibfehlerLaesstDasPaketUnberuehrt(t *testing.T) {
	t.Parallel()

	for _, form := range paketformen() {
		for _, stelle := range []abbruchstelle{
			beimAnlegen, beimEinschreiben, beimSynchronisieren, beimSchliessen,
			beimFreigeben,
		} {
			t.Run(form.name+", "+stelle.String(), func(t *testing.T) {
				t.Parallel()

				ort := form.bauen(t)
				vorher := paketstand(t, ort)

				fehler := dateiSchreiben(context.Background(), ort, form.hauptdatei,
					[]byte("mse version: 2.0.0\ngame: anders\n"),
					summeVon(vorher[form.hauptdatei]),
					hilfenMitAbbruch(stelle))

				// **Ein Verzeichnispaket hat kein Archiv freizugeben.** Dann
				// gelingt das Schreiben, und geprüft wird, dass es wirklich
				// ankam — sonst prüfte dieser Zweig gar nichts.
				if stelle == beimFreigeben && !form.archiv {
					if fehler != nil {
						t.Fatalf("das Verzeichnispaket kennt diesen Griff nicht, "+
							"und trotzdem kam ein Fehler: %v", fehler)
					}

					if stand := paketstand(t, ort)[form.hauptdatei]; bytes.Equal(
						stand, vorher[form.hauptdatei]) {
						t.Error("nichts ist geschrieben worden")
					}

					return
				}

				if !errors.Is(fehler, errAbbruch) {
					t.Fatalf("der Abbruch %s wurde nicht durchgereicht: %v", stelle, fehler)
				}

				standVergleichen(t, vorher, paketstand(t, ort))
				ohneNachbardatei(t, ort)
			})
		}
	}
}
