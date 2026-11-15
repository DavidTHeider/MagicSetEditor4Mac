// **Den Bestand sichern und hineinsehen** —
// #1169,
// Werkstück 2.
//
// # Zwei Werkzeuge und nicht drei
//
// **Zurückgespielt wird hier nicht.** Der Maßstab ist die **Umkehrbarkeit**
// ([0044](../../docs/entscheidungen/0044-was-ein-werkzeug-schreiben-darf.md)):
// „Ein Werkzeug schreibt nur, was es nicht überschreibt." Ein Zurückspielen
// überschreibt den **ganzen** Bestand — jedes Set, jede Karte, jeden Feldwert —,
// und der alte liegt danach nur noch neben der Datei, in der Hoffnung, dass
// jemand hinsieht.
//
// **Nicht §16, Regel 6.** Der erste Anlauf zitierte sie („nichts wird still
// übernommen"), und der Zweitblick über Phase 3.8 hat es als Befund 6
// angestrichen: sie gilt fremden **Diensten** und Daten, die in ein Kartenfeld
// wollen. Eine Sicherung ist eine eigene Datei auf der eigenen Platte. Die
// Ablehnung stimmt, ihre Begründung stimmte nicht — und eine Regel auf einen
// Fall auszudehnen, für den sie nicht gemacht ist, ist die Ausrede, vor der §10
// warnt.
//
// **Sichern und Hineinsehen schon**: das eine legt daneben, das andere liest.
// Wer sie ruft, verliert nichts — und ein Modell, das vor einer großen
// Änderung sichert, tut genau das, wofür es die Sicherung gibt.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Bestandssicherung ist, was die zwei Werkzeuge dieser Datei brauchen.
type Bestandssicherung interface {
	// BestandSichern schreibt die Sicherung — **ohne Kennwort**, und aus
	// einem verschlüsselten Bestand nur mit `klartext` (OPS-06).
	BestandSichern(ctx context.Context, ziel string, klartext bool) error
	SicherungSchauen(ctx context.Context, pfad string) (Sicherungsangaben, error)

	// Verschluesselungsstand sagt, ob der Bestand verschlüsselt auf der Platte
	// liegt — Werkstück 14 von Phase 4.12.
	//
	// **Nur das Lesen und nicht das Umschalten**, und der Grund ist derselbe,
	// aus dem hier auch kein Zurückspielen steht: das Umschalten **tauscht die
	// Bestandsdatei**. Dieser Prozess hält eine Verbindung über seine ganze
	// Laufzeit, und jedes Repository hängt daran; nach einem Tausch hielten
	// sie alle einen Griff auf eine Datei, die es nicht mehr gibt.
	//
	// Geschaltet wird am Fenster und mit `mse bestand-verschluesselung` —
	// beides sind Stellen, die ihren Lebenszyklus danach neu aufbauen.
	Verschluesselungsstand() (bool, error)

	// Sicherungsstand nennt das Alter der letzten Sicherung und den Plan —
	// OPS01 (#2347). **Eine Gestalt, die dieses Paket nicht kennt**: sie
	// kommt aus `internal/app` und ist an jeder Fläche dieselbe, und eine
	// Abschrift hier liefe an dem Tag auseinander, an dem sie ein Feld
	// dazubekommt.
	Sicherungsstand(ctx context.Context) (any, error)
}

// Sicherungsangaben sind, was eine Sicherung über sich hergibt.
type Sicherungsangaben struct {
	AngelegtAm  string `json:"angelegtAm"`
	Schemastand string `json:"schemastand"`
	Sets        int    `json:"sets"`
	Karten      int    `json:"karten"`
	Bilder      string `json:"bilder"`

	Verschluesselt bool `json:"verschluesselt"`

	// PaketeEnthalten und PaketeFehlen: welche eigenen Pakete mitkommen und
	// welche bezogenen hier neu zu beziehen sind (OPS-06.AK-03).
	PaketeEnthalten []string `json:"paketeEnthalten"`
	PaketeFehlen    []string `json:"paketeFehlen"`
}

// sicherungswerkzeuge baut die beiden.
func sicherungswerkzeuge(quelle Bestandssicherung) (lesend, schreibend []Werkzeug) {
	if quelle == nil {
		return nil, nil
	}

	lesend = []Werkzeug{
		{
			Name: "lesen.sicherung", Art: Lesend, Idempotent: true,
			Erklaerung: "Sagt, was in einer Sicherungsdatei liegt: wann sie entstand, " +
				"welchen Schemastand sie hat, wie viele Sets und Karten darin sind und " +
				"was mit den Bildern geschähe. ÄNDERT NICHTS.",
			Schema: objekt(map[string]any{
				"pfad": text("Der Pfad der .msesicherung."),
			}, "pfad"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Pfad string `json:"pfad"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if argumente.Pfad == "" {
					return "", fmt.Errorf("%w: es wurde kein Pfad genannt", domain.ErrValidation)
				}

				angaben, err := quelle.SicherungSchauen(ctx, argumente.Pfad)
				if err != nil {
					return "", err
				}

				return alsJSON(angaben)
			},
		},
		{
			// **Nur lesend** — der Grund steht bei [Bestandssicherung].
			Name: "lesen.bestand.verschluesselung", Art: Lesend, Idempotent: true,
			Erklaerung: "Sagt, ob der Bestand verschlüsselt auf der Platte liegt. " +
				"UMSCHALTEN geht hier nicht: das tauscht die Bestandsdatei, und dieser " +
				"Server hält eine Verbindung darauf, solange er läuft. Geschaltet wird " +
				"am Fenster oder mit »mse bestand-verschluesselung an« bzw. »aus«.",
			Schema: objekt(nil),
			Ruf: func(context.Context, json.RawMessage) (string, error) {
				an, err := quelle.Verschluesselungsstand()
				if err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"verschluesselt": an})
			},
		},
		{
			Name: "lesen.bestand.sicherungsstand", Art: Lesend, Idempotent: true,
			Erklaerung: "Sagt, wann der Bestand zuletzt gesichert wurde — geschrieben UND zurückgelesen —, " +
				"über welche Fläche und in welche Datei, wie viele Tage das her ist und ob es länger als 14 Tage " +
				"sind (stufe: unbekannt, frisch, alt, unstimmig), dazu den freiwilligen Plan. " +
				"»unbekannt« heißt: über diese Fassung noch nie. ÄNDERT NICHTS.",
			Schema: objekt(nil),
			Ruf: func(ctx context.Context, _ json.RawMessage) (string, error) {
				stand, err := quelle.Sicherungsstand(ctx)
				if err != nil {
					return "", err
				}

				return alsJSON(stand)
			},
		},
	}

	schreibend = []Werkzeug{
		{
			Name: "schreiben.bestand.sichern", Art: Schreibend,
			Erklaerung: "ÄNDERT die Platte: schreibt den GANZEN Bestand samt eingesetzten " +
				"Bildern als .msesicherung an einen Ort, an dem noch nichts liegt. " +
				"Ein vorhandenes Ziel wird abgewiesen. Zurückgespielt wird hier nicht — " +
				"dafür gibt es das Fenster, »mse bestand-zurueckspielen« und den Dienst. " +
				"OHNE KENNWORT: ein Kennwort, das durch ein Modell geht, steht in dessen " +
				"Kontext und ist kein Geheimnis mehr. Ist der Bestand verschlüsselt, " +
				"verlangt die Sicherung klartext=true.",
			Schema: objekt(map[string]any{
				"ziel": text("Der Pfad, an den geschrieben wird — dort darf nichts liegen."),
				"klartext": jaNein("Wahr heißt ausdrücklich: ohne Kennwort, auch wenn " +
					"der Bestand verschlüsselt ist. Die Datei liegt dann im Klartext."),
			}, "ziel"),
			Ruf: func(ctx context.Context, roh json.RawMessage) (string, error) {
				var argumente struct {
					Ziel     string `json:"ziel"`
					Klartext bool   `json:"klartext"`
				}

				if err := lies(roh, &argumente); err != nil {
					return "", err
				}

				if argumente.Ziel == "" {
					return "", fmt.Errorf("%w: es wurde kein Ziel genannt", domain.ErrValidation)
				}

				if err := quelle.BestandSichern(ctx, argumente.Ziel, argumente.Klartext); err != nil {
					return "", err
				}

				return alsJSON(map[string]any{"pfad": argumente.Ziel})
			},
		},
	}

	return lesend, schreibend
}
