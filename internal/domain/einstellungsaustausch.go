// **Das Austauschformat der Einstellungen** — Werkstück O3 von Phase 5.2
// (#1427),
// nach [Vermerk 0079](../../docs/entscheidungen/0079-einstellungsebenen-und-portabler-import.md).
//
// # Was mitreist und was nicht
//
// **Ein Abschnitt ist eine Ebene**, und was in ihm steht, ist auf dieser Ebene
// **gesetzt**. Die Ebene `geraet` hat keinen Abschnitt: sie beschreibt diesen
// Rechner — einen Pfad auf dieser Platte, die Kennung, unter der sich diese
// Installation ausweist, die Frage, ob diese Maschine ins Netz sieht. Drüben
// beschriebe das eine Maschine, die es nicht gibt.
//
// **Und der Arbeitsplatz reist nur auf Wunsch**: ein Layout hängt an einem
// Bildschirm, nicht an einem Menschen.
//
// # Warum eine Fassungsnummer und keine Verhandlung
//
// **Weil eine Datei nicht antwortet.** Ein Dienst kann sagen, was er kann;
// eine Datei liegt da. Die Zahl ist die einzige Auskunft, die sie über sich
// selbst geben kann — und die Regel dazu ist die aus
// [0033](../../docs/entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md):
// eine **kleinere** oder gleiche wird gelesen, eine **größere** wird
// verständlich abgewiesen. Verständlich heißt: die Meldung nennt beide Zahlen.

package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

// AustauschFassung ist die Fassung, die dieses Programm schreibt und liest.
const AustauschFassung = 1

// AustauschHoechstens ist die größte Datei, die gelesen wird — in Bytes.
//
// **64 KiB, und die Zahl ist gerechnet.** Der größte Austausch, den dieses
// Programm heute schreiben kann, sind alle Schlüssel beider Reihen: rund
// dreißig Zeilen mit je einem Schlüssel von höchstens 32 Zeichen und einem
// Wert von höchstens 16, samt Anführungszeichen und Komma — zusammen etwa
// **1,5 KB**. 64 KiB sind das Vierzigfache; wer sie überschreitet, schickt
// etwas anderes als Einstellungen.
//
// **Die Grenze greift vor dem Entziffern**, nicht danach: eine Datei, die
// nicht gelesen werden soll, wird nicht gelesen.
const AustauschHoechstens = 64 << 10

// AustauschTiefe ist die tiefste Verschachtelung, die dieses Format hat.
//
// **Drei, und es sind genau drei.** `{ fassung, bereiche: { ebene: { schlüssel:
// wert } } }` — das äußere Objekt ist eins, `bereiche` zwei, ein Abschnitt
// drei; Werte sind Zeichenketten. Was tiefer geht, ist nicht dieses Format,
// und es zu entziffern hieße, einer fremden Datei zu erlauben, wie tief sie
// den Stapel legt.
const AustauschTiefe = 3

// Austausch ist, was über die Naht geht.
type Austausch struct {
	Fassung int `json:"fassung"`

	// Bereiche hält je Ebene die **gesetzten** Werte.
	//
	// **Eine Abbildung und keine Liste von Verbünden**: ein Schlüssel steht
	// je Abschnitt einmal, und eine Liste könnte ihn zweimal tragen.
	Bereiche map[string]map[string]string `json:"bereiche"`
}

// Einfuhrbefund ist, was an einer Zeile nicht stimmt.
//
// **Ein Befund und kein Fehler** (§8): eine Datei mit dreißig guten Zeilen und
// zwei fremden wird übernommen, und die zwei werden **genannt**. Wer sie
// verschluckte, ließe jemanden glauben, es sei alles angekommen.
type Einfuhrbefund struct {
	Ebene      string `json:"ebene"`
	Schluessel string `json:"schluessel"`
	Grund      string `json:"grund"`
}

// Einfuhrpaar ist eine Zeile, die angenommen wird.
type Einfuhrpaar struct {
	Ebene      Ebene
	Schluessel string
	Wert       string
}

// AbschnittImProfil gibt den Abschnitt, in dem ein Wert aus der Profiltafel
// reist — der zweite Wert ist falsch, wenn er nicht reist.
//
// **Eine Übersteuerung reist im Abschnitt `profil`**, obwohl die Frage dem
// Menschen gehört: sie ist eine Aussage über **dieses Profil** („hier arbeite
// ich auf Englisch"), und dort steht sie auch drüben.
func AbschnittImProfil(schluessel string) (Ebene, bool) {
	if !slices.Contains(Einstellungsreihe(), schluessel) {
		return "", false
	}

	ebene, bekannt := EbeneVon(schluessel)
	if !bekannt || ebene == EbeneGeraet {
		return "", false
	}

	if ebene == EbeneArbeitsplatz {
		return EbeneArbeitsplatz, true
	}

	return EbeneProfil, true
}

// ReistBeimNutzer sagt, ob ein Wert aus der Nutzertafel mitreist.
//
// **Das Layout reist von hier nicht mit.** Beim Menschen steht es als
// **Rückfall** — was gilt, wenn ein Profil nichts sagt —, und ein Rückfall ist
// eine Gewohnheit an diesem Bildschirm. Was ausdrücklich für ein Profil
// gewählt wurde, reist über [AbschnittImProfil], und zwar auf Wunsch.
func ReistBeimNutzer(schluessel string) bool {
	if !slices.Contains(Nutzereinstellungsreihe(), schluessel) {
		return false
	}

	ebene, bekannt := EbeneVon(schluessel)

	return bekannt && ebene == EbeneNutzer
}

// Zerlege liest einen Austausch und weist ab, was gar nicht zu lesen ist.
//
// **Drei Ergebnisse und nicht zwei**: was gelesen wurde, was daran auffiel,
// und ob es überhaupt ging. Ein unbekanntes Feld ist ein **Befund** — die
// Datei bleibt lesbar; eine Fassung aus der Zukunft ist ein **Fehler** — was
// darin steht, weiß dieses Programm nicht.
func Zerlege(roh []byte) (Austausch, []Einfuhrbefund, error) {
	if len(roh) > AustauschHoechstens {
		return Austausch{}, nil, NewFieldError("datei", SchluesselValidierung,
			"Diese Datei ist %d Bytes groß; gelesen werden höchstens %d.",
			len(roh), AustauschHoechstens)
	}

	tiefe, fehler := tiefeVon(roh)
	if fehler != nil {
		return Austausch{}, nil, NewFieldError("datei", SchluesselValidierung,
			"Diese Datei ist kein JSON: %v", fehler)
	}

	if tiefe > AustauschTiefe {
		return Austausch{}, nil, NewFieldError("datei", SchluesselValidierung,
			"Diese Datei ist %d Stufen tief verschachtelt; dieses Format hat %d.",
			tiefe, AustauschTiefe)
	}

	return ausFeldern(roh)
}

// ausFeldern liest die Felder der obersten Ebene.
func ausFeldern(roh []byte) (Austausch, []Einfuhrbefund, error) {
	var felder map[string]json.RawMessage

	if fehler := json.Unmarshal(roh, &felder); fehler != nil {
		return Austausch{}, nil, NewFieldError("datei", SchluesselValidierung,
			"Diese Datei trägt kein Objekt: %v", fehler)
	}

	austausch := Austausch{Bereiche: map[string]map[string]string{}}

	if fehler := json.Unmarshal(mitVorgabe(felder["fassung"], "0"),
		&austausch.Fassung); fehler != nil {
		return Austausch{}, nil, NewFieldError("fassung", SchluesselValidierung,
			"Das Feld »fassung« ist keine Zahl: %v", fehler)
	}

	if austausch.Fassung < 1 || austausch.Fassung > AustauschFassung {
		return Austausch{}, nil, NewFieldError("fassung", SchluesselValidierung,
			"Diese Datei ist in Fassung %d geschrieben; dieses Programm liest "+
				"Fassung 1 bis %d.", austausch.Fassung, AustauschFassung)
	}

	if fehler := json.Unmarshal(mitVorgabe(felder["bereiche"], "{}"),
		&austausch.Bereiche); fehler != nil {
		return Austausch{}, nil, NewFieldError("bereiche", SchluesselValidierung,
			"Das Feld »bereiche« hat nicht die erwartete Gestalt: %v", fehler)
	}

	return austausch, fremdeFelder(felder, "fassung", "bereiche"), nil
}

// mitVorgabe gibt den Rohwert oder, wo keiner steht, eine Ersatzangabe.
func mitVorgabe(roh json.RawMessage, ersatz string) json.RawMessage {
	if len(roh) == 0 {
		return json.RawMessage(ersatz)
	}

	return roh
}

// fremdeFelder meldet, was oben steht und hier niemand kennt.
//
// **Sortiert** — Go durchläuft eine `map` absichtlich in wechselnder
// Reihenfolge, und eine Liste von Befunden, die sich zwischen zwei Blicken
// umsortiert, liest sich wie eine andere Antwort.
func fremdeFelder(
	felder map[string]json.RawMessage, bekannt ...string,
) []Einfuhrbefund {
	fremde := []string{}

	for name := range felder {
		if !slices.Contains(bekannt, name) {
			fremde = append(fremde, name)
		}
	}

	slices.Sort(fremde)

	befunde := make([]Einfuhrbefund, 0, len(fremde))

	for _, name := range fremde {
		befunde = append(befunde, Einfuhrbefund{
			Schluessel: name,
			Grund:      fremdesFeldGrund(name),
		})
	}

	return befunde
}

// fremdesFeldGrund sagt, was an einem Feld unbekannt ist.
//
// **Ein Feld ohne Namen bekommt einen eigenen Satz** — Werkstück O11 von
// Phase 5.2 (#1435),
// gefunden vom Fuzzlauf über `Zerlege`. JSON erlaubt `""` als Feldnamen, und
// die allgemeine Fassung las sich dann als »Das Feld »« kennt dieses Programm
// nicht« — ein Befund, der weder sagt, **wo** er herkommt, noch **was** dort
// stand (§8: ein Befund ohne Ort ist keiner).
func fremdesFeldGrund(name string) string {
	if strings.TrimSpace(name) == "" {
		return "Diese Datei trägt ein Feld ohne Namen; es wird übergangen."
	}

	return fmt.Sprintf("Das Feld »%s« kennt dieses Programm nicht; "+
		"es wird übergangen.", name)
}

// tiefeVon zählt die tiefste Verschachtelung, ohne den Inhalt zu bauen.
//
// **Über die Marken und nicht über einen Aufbau**: `encoding/json` legt beim
// Entziffern für jede Stufe einen Rahmen an, und wer erst baut und dann misst,
// hat die Tiefe schon bezahlt.
func tiefeVon(roh []byte) (int, error) {
	entzifferer := json.NewDecoder(bytes.NewReader(roh))

	tiefste, jetzt := 0, 0

	for {
		marke, fehler := entzifferer.Token()
		if errors.Is(fehler, io.EOF) {
			return tiefste, nil
		}

		if fehler != nil {
			return 0, fehler
		}

		klammer, ist := marke.(json.Delim)
		if !ist {
			continue
		}

		if klammer == '{' || klammer == '[' {
			jetzt++

			tiefste = max(tiefste, jetzt)

			continue
		}

		jetzt--
	}
}
