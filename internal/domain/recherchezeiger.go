package domain

// **Der Zeiger auf die nächste Seite** — Werkstück A16 von Phase 5.3
// (#1515).
//
// # Was ein Zeiger binden muss
//
// **Fünf Dinge**, und das Issue zählt sie auf: „Begrenzte Pagination/Cursor
// binden Anfrage, Profil, Rechte, Generation und Ablauf; bei Wechsel
// konsistent fortsetzen oder explizit ungültig, **keine still gemischten
// Seiten**."
//
// Der letzte Halbsatz ist der Grund. Ein Zeiger, der nur eine Stelle trägt,
// setzt nach einem Neuaufbau des Index in einer **anderen** Liste fort — und
// die zweite Seite gehört dann zu einer Suche, die niemand gestellt hat.
//
// # Warum er sich selbst prüft und nicht der Bestand
//
// **Weil er nichts ist, was jemandem gehört.** Er trägt seinen eigenen
// Abdruck; wer ihn verändert, macht ihn ungültig, und wer ihn an einem
// anderen Stand vorlegt, bekommt eine Absage statt einer falschen Seite. Eine
// Tafel dafür wäre ein zweiter Ort, an dem etwas abläuft.

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Zeigerfrist ist, wie lange ein Zeiger gilt.
//
// **Fünfzehn Minuten.** Wer länger braucht, stellt die Frage neu — und
// bekommt dann die Liste, die heute gilt, statt einer von vorhin.
const Zeigerfrist = 15 * time.Minute

// Recherchezeiger sagt, wo eine Liste fortgesetzt wird.
type Recherchezeiger struct {
	// Stelle ist, wie viele Treffer schon gezeigt wurden.
	Stelle int `json:"stelle"`

	// Abdruck bindet Anfrage, Profil und Indexstand.
	Abdruck string `json:"abdruck"`

	// Erstellt ist, wann er entstanden ist.
	Erstellt time.Time `json:"erstellt"`
}

// Zeigerabdruck bindet eine Anfrage an Profil und Indexstand.
//
// **Die Rechte stecken im Profil.** Wer das Profil wechselt, bekommt einen
// anderen Abdruck — und damit eine Absage statt der Treffer des anderen.
func Zeigerabdruck(
	anfrage Rechercheanfrage, profil int64, indexabdruck string,
) string {
	teile := strings.Join([]string{
		string(anfrage.Fall), anfrage.Frage,
		strconv.FormatInt(anfrage.AusgangskarteID, 10), anfrage.Ausgangsstand,
		anfrage.Sprache, strconv.Itoa(anfrage.Grenze), anfrage.Korpusstand,
		strconv.FormatInt(profil, 10), indexabdruck,
	}, "\x1f")

	summe := sha256.Sum256([]byte(teile))

	return hex.EncodeToString(summe[:])
}

// AlsZeichenkette gibt den Zeiger in der Form, in der er hinausgeht.
//
// **Base64 über JSON, und nicht lesbar gemeint**: was darin steht, ist eine
// Stelle und ein Abdruck — beides sagt einem Aufrufer nichts, und wer es
// ändert, macht den Zeiger ungültig.
func (z Recherchezeiger) AlsZeichenkette() string {
	roh, _ := json.Marshal(z) //nolint:errcheck,errchkjson // drei Grundwerte

	return base64.RawURLEncoding.EncodeToString(roh)
}

// ZeigerAus liest einen Zeiger und prüft ihn gegen Abdruck und Frist.
//
// **Drei Absagen, und jede sagt etwas anderes**: unlesbar, fremd, abgelaufen.
// Wer sie zusammenfasste, ließe jemanden raten, ob er die Frage neu stellen
// oder das Profil wechseln muss.
func ZeigerAus(roh, abdruck string, jetzt time.Time) (Recherchezeiger, error) {
	gelesen, fehler := base64.RawURLEncoding.DecodeString(roh)
	if fehler != nil {
		return Recherchezeiger{}, fmt.Errorf(
			"%w: dieser Fortsetzungszeiger ist unlesbar", ErrValidation)
	}

	var zeiger Recherchezeiger

	if fehler := json.Unmarshal(gelesen, &zeiger); fehler != nil {
		return Recherchezeiger{}, fmt.Errorf(
			"%w: dieser Fortsetzungszeiger ist unlesbar", ErrValidation)
	}

	if zeiger.Abdruck != abdruck {
		return Recherchezeiger{}, fmt.Errorf(
			"%w: dieser Fortsetzungszeiger gehört zu einer anderen Frage, "+
				"einem anderen Profil oder einem anderen Indexstand — die "+
				"Liste wäre eine andere als die, in der er entstanden ist",
			ErrConflict)
	}

	if !jetzt.Before(zeiger.Erstellt.Add(Zeigerfrist)) {
		return Recherchezeiger{}, fmt.Errorf(
			"%w: dieser Fortsetzungszeiger ist älter als %v — die Frage "+
				"noch einmal stellen gibt die Liste, die heute gilt",
			ErrConflict, Zeigerfrist)
	}

	if zeiger.Stelle < 0 {
		return Recherchezeiger{}, fmt.Errorf(
			"%w: dieser Fortsetzungszeiger nennt die Stelle %d",
			ErrValidation, zeiger.Stelle)
	}

	return zeiger, nil
}
