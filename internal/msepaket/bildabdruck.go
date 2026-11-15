package msepaket

// **Der Abdruck eines Paketbildes** — Werkstück G2 von Phase 5.5
// (#1459).
//
// # Wofür es ihn gibt
//
// **Damit ein gerastertes Kartenbild merkt, wenn sich ein Rahmen ändert.** Ein
// Zeichenplan trägt **Pfade**, keine Bytes; wer nur ihn abdruckt, hat einen
// Schlüssel, der einen ausgetauschten Rahmen nicht bemerkt — und der Spieltisch
// zeigte das Bild von vorhin.
//
// # Warum die Bytes und nicht das entschlüsselte Bild
//
// **Weil das entschlüsselte Bild von der Bibliothek abhängt**, die es
// entschlüsselt hat, und die Bytes nicht. Zwei Go-Fassungen könnten dasselbe
// PNG zu zwei `image.Image` machen, die sich in einem Randpixel unterscheiden;
// die Datei bliebe dieselbe. Was den Schlüssel trägt, soll die Datei sein.
//
// # Und warum es hier steht und nicht beim Rufer
//
// **Weil nur dieses Paket weiß, wo die Datei liegt.** Ein Rufer, der den Pfad
// selbst auflöste, hätte eine zweite Antwort auf dieselbe Frage — und liefe
// bei der ersten neuen Wurzel auseinander (§2).

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
)

// Bildabdruecke gibt den Abdruck von Paketbildern.
//
// **Ein eigener Typ neben [Bildquelle]** und kein Verfahren an ihr: jene hält
// entschlüsselte Bilder, dieser hält Abdrücke. Ein gemeinsamer Speicher
// verdrängte das eine für das andere — und ein Abdruck ist 32 Byte, ein Bild
// 780 KB.
type Bildabdruecke struct {
	wurzeln []string

	sperre  sync.RWMutex
	gemerkt map[string]string
}

// NeueBildabdruecke baut sie über denselben Wurzeln wie eine [Bildquelle].
func NeueBildabdruecke(wurzeln ...string) *Bildabdruecke {
	return &Bildabdruecke{
		wurzeln: echteWurzeln(wurzeln),
		gemerkt: map[string]string{},
	}
}

// Abdruck gibt den Abdruck der Datei hinter einem Verweis.
//
// **Ein fehlendes Bild ist kein Fehler, sondern ein Wert.** Wer eine Karte mit
// fehlendem Rahmen zeichnet, bekommt ein Bild mit einem Befund daneben — und
// dieses Bild soll wiedererkannt werden. Ein Fehler hier machte aus einem
// Befund einen Abbruch, und die Karte verschwände ganz.
//
// **Der Wert für „fehlt" ist trotzdem ein anderer als für jede Datei**: taucht
// die Datei wieder auf, ändert sich der Schlüssel, und das Bild wird neu
// gezeichnet.
func (b *Bildabdruecke) Abdruck(verweis string) string {
	if abdruck, da := b.nachschlagen(verweis); da {
		return abdruck
	}

	abdruck := b.lies(verweis)

	b.sperre.Lock()
	b.gemerkt[verweis] = abdruck
	b.sperre.Unlock()

	return abdruck
}

// Hat sagt, ob es die Datei hinter einem Verweis gibt.
//
// # Warum es das gibt und nicht ein Vergleich auf den Vorsatz
//
// **Weil `fehlt:` eine Sache dieses Pakets ist.** Ein Aufrufer, der den
// Vorsatz selbst prüfte, hinge an einer Zeichenkette, die hier steht — und
// am Tag, an dem sie sich ändert, fände er jede Datei. Gefragt wird deshalb
// hier, wo die Marke gesetzt wird.
//
// **Es kostet nichts mehr als ein [Bildabdruecke.Abdruck]** und merkt sich
// dasselbe: wer erst fragt und dann abdruckt, liest die Datei einmal.
func (b *Bildabdruecke) Hat(verweis string) bool {
	return !strings.HasPrefix(b.Abdruck(verweis), fehltVorsatz)
}

// fehltVorsatz steht vor dem Abdruck einer Datei, die es nicht gibt.
const fehltVorsatz = "fehlt:"

// lies sucht die Datei und druckt ihre Bytes ab.
func (b *Bildabdruecke) lies(verweis string) string {
	if ort, datei, ist := ZerlegeSetpaketverweis(verweis); ist {
		rohdaten, fehler := paketdatei(ort, datei)
		if fehler != nil {
			return fehlt(fehler)
		}

		return abdruckVon(rohdaten)
	}

	pfad, fehler := ZerlegePaketpfad(verweis)
	if fehler != nil {
		return fehlt(fehler)
	}

	versuche := make([]error, 0, len(b.wurzeln)+1)
	versuche = append(versuche, ErrKeinBild)

	for _, wurzel := range b.wurzeln {
		rohdaten, fehler := paketdatei(filepath.Join(wurzel, pfad.Paket), pfad.Datei)
		if fehler != nil {
			versuche = append(versuche, fehler)

			continue
		}

		return abdruckVon(rohdaten)
	}

	return fehlt(errors.Join(versuche...))
}

// nachschlagen sucht einen Abdruck im Speicher.
func (b *Bildabdruecke) nachschlagen(verweis string) (string, bool) {
	b.sperre.RLock()
	defer b.sperre.RUnlock()

	abdruck, da := b.gemerkt[verweis]

	return abdruck, da
}

// abdruckVon druckt Bytes ab.
func abdruckVon(rohdaten []byte) string {
	summe := sha256.Sum256(rohdaten)

	return hex.EncodeToString(summe[:])
}

// fehlt gibt den Wert für eine Datei, die es nicht gibt.
//
// **Mit dem Grund darin**, damit „die Wurzel ist unlesbar" und „es gibt sie
// nicht" zwei verschiedene Schlüssel geben: taucht die Wurzel wieder auf,
// ändert sich der Schlüssel, und die Karte wird neu gezeichnet.
func fehlt(grund error) string {
	return fehltVorsatz + abdruckVon([]byte(fmt.Sprint(grund)))
}
