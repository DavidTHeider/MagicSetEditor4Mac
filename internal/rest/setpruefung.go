// **Gibt es dieses Set?** — und die dritte Antwort
// (#1155).
//
// **Eine eigene Datei**, seit `endpunkte.go` über die Zeilengrenze aus §2
// ging. Der Schnitt folgt der Frage: dort stehen die Endpunkte, hier die eine
// Auskunft, die zwei von ihnen brauchen.

package rest

import (
	"fmt"
	"net/http"
)

// setGibtEs sagt, ob es ein Set mit dieser Kennung gibt.
//
// **Mit Fehler und nicht nur mit `false`.** Wer nicht nachsehen kann, weiß
// nicht, dass es das Set nicht gibt — er weiß gar nichts. Ein `false` an
// dieser Stelle machte aus einem Ausfall eine fachliche Auskunft.
func (d *Dienst) setGibtEs(anfrage *http.Request, setID int64) (bool, error) {
	koepfe, fehler := d.naehte.Leser.Setkoepfe(anfrage.Context())
	if fehler != nil {
		return false, fmt.Errorf("nachsehen, ob es das Set %d gibt: %w", setID, fehler)
	}

	for _, kopf := range koepfe {
		if kopf.ID == setID {
			return true, nil
		}
	}

	return false, nil
}
