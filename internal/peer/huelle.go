package peer

// **Die Nachrichtenhülle** — Werkstück N1 von Phase 5.4
// (#1454).
//
// # Warum der Kopf mit unterschrieben wird und nicht nur der Rumpf
//
// **Weil eine Nachricht sonst umziehen könnte.** Wer nur den Rumpf
// unterschreibt, schreibt eine Zusage über einen Inhalt — nicht darüber, in
// welchem Protokoll, in welcher Sitzung, an welcher Stelle der Reihe und von
// wem er gilt. Ein Mitleser nähme dieselben Bytes, setzte sie in eine andere
// Sitzung, und die Unterschrift hielte.

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// Huelle ist eine Nachricht dieses Programms an einen anderen Peer.
//
// # Was hier Fach und was Hülle ist
//
// **Der [Huelle.Rumpf] bleibt roh.** Diese Schicht liest ihn nicht; sie
// prüft, dass er JSON ist, und reicht ihn weiter. Was darin steht, entscheidet
// das Fachprotokoll — und es entscheidet es **nach** dieser Prüfung, nie
// davor.
//
// # Die Felder
//
//   - `Protokoll` ist eines der drei aus dem Issue, **mit seiner Fassung**.
//     `/mse4mac/pair/2` ist ein anderes Protokoll und keine Verhandlungssache.
//   - `Art` ist der Nachrichtentyp innerhalb des Protokolls.
//   - `ID` benennt diese eine Nachricht, `Sitzung` den Vorgang, zu dem sie
//     gehört — eine Paarung, eine Übertragung, eine Partie.
//   - `Absender` ist die [Kennung] in ihrer Textform. **Sie ist zugleich der
//     Prüfschlüssel**: wer sie austauscht, tauscht den Schlüssel mit, und die
//     Unterschrift bricht.
//   - `Folge` zählt innerhalb der Sitzung.
type Huelle struct {
	Protokoll    string          `json:"protokoll"`
	Art          string          `json:"art"`
	ID           string          `json:"id"`
	Sitzung      string          `json:"sitzung"`
	Absender     string          `json:"absender"`
	Folge        int             `json:"folge"`
	Rumpf        json.RawMessage `json:"rumpf"`
	Unterschrift string          `json:"unterschrift"`
}

// Verpacke macht aus einer Hülle die Bytes, die hinausgehen.
//
// # Was hier in welcher Reihenfolge geschieht
//
// **Erst der Kopf, dann die kanonische Form, dann die Unterschrift.** Der Kopf
// ist die billigste Prüfung; ein Rumpf wird nicht verdichtet und nichts wird
// unterschrieben, solange ein Pflichtfeld fehlt.
//
// **Der Rumpf wird verdichtet** (`json.Compact`), und zwar **bevor**
// unterschrieben wird. Das ist die „kanonische Codierung" aus dem Issue: was
// unterschrieben wird, ist genau das, was hinausgeht — Byte für Byte. Ein
// Vermittler, der die Nachricht neu einrückte, bräche die Unterschrift, und
// das ist gewollt: eine Änderung an den Bytes ist eine Änderung.
//
// **Unter fremdem Namen geht nichts hinaus.** Eine Hülle mit einem fremden
// `Absender` prüfte bei keinem Empfänger; sie wäre eine Nachricht, die
// unterwegs stirbt, und der Fehler entstünde hier.
func (i Identitaet) Verpacke(huelle Huelle) ([]byte, error) {
	if fehler := huelle.pruefeKopf(); fehler != nil {
		return nil, fehler
	}

	if huelle.Absender != string(i.kennung) {
		return nil, fmt.Errorf(
			"%w: diese Identität heißt %s und nicht %s", domain.ErrValidation,
			i.kennung, huelle.Absender)
	}

	huelle.Rumpf = verdichtet(huelle.Rumpf)
	huelle.Unterschrift = hex.EncodeToString(
		i.Signiere(huelle.unterschriftsform()))

	roh := huelle.alsJSON()

	// **Die zweite Grenze, und sie ist eine andere als die erste.** Der
	// Blick in [Huelle.pruefeKopf] misst den Rumpf; über die Leitung geht
	// die **Nachricht**, und die ist um ihren Kopf größer.
	if len(roh) > Hoechstlaenge {
		return nil, zuGross(len(roh))
	}

	return roh, nil
}

// Auspacke liest eine empfangene Nachricht und prüft sie.
//
// # Die Reihenfolge ist die Zusage
//
// **Zuerst die Länge, dann das Zerteilen, dann die Unterschrift** — und die
// Fachausführung erst danach, außerhalb dieses Pakets. Das Issue verlangt
// genau das: „unzulässige/zu große Typen vor Fachausführung abweisen". Wer
// erst zerteilt und dann misst, hat für jede zu große Nachricht schon den
// Speicher gestellt, gegen den die Grenze schützt.
//
// **Was hier *nicht* geprüft wird**: ob der Absender bekannt, erlaubt oder
// erwünscht ist. Diese Hülle sagt, dass eine Nachricht **echt** ist — nicht,
// dass sie **willkommen** ist. Wer das entscheidet, ist N1.4 (Kontakte,
// Zwecke, Rechte); wer es hier täte, hätte eine Rechteliste in einem Blatt.
func Auspacke(roh []byte) (Huelle, error) {
	if len(roh) > Hoechstlaenge {
		return Huelle{}, zuGross(len(roh))
	}

	var huelle Huelle

	if fehler := json.Unmarshal(roh, &huelle); fehler != nil {
		return Huelle{}, domain.MitGrund(domain.ErrValidation, fehler,
			"die Nachricht ließ sich nicht lesen")
	}

	if fehler := huelle.Pruefe(); fehler != nil {
		return Huelle{}, fehler
	}

	return huelle, nil
}

// Pruefe hält Kopf und Unterschrift gegeneinander.
//
// **Getrennt von [Auspacke], weil eine Hülle durch die Hand geht.** Sie reist
// durch `service`, wird weitergereicht und liegt zwischendurch in einer
// Struktur; wer sie dort ändert, soll sie erneut prüfen können, ohne sie
// zweimal zu kodieren.
func (h Huelle) Pruefe() error {
	if fehler := h.pruefeKopf(); fehler != nil {
		return fehler
	}

	unterschrift, fehler := hex.DecodeString(h.Unterschrift)
	if fehler != nil {
		return domain.MitGrund(domain.ErrValidation, fehler,
			"die Unterschrift ist kein Hex")
	}

	if len(unterschrift) != ed25519.SignatureSize {
		return fmt.Errorf(
			"%w: eine Unterschrift wiegt %d Byte, hier sind es %d",
			domain.ErrValidation, ed25519.SignatureSize, len(unterschrift))
	}

	return Pruefe(Kennung(h.Absender), h.unterschriftsform(), unterschrift)
}
