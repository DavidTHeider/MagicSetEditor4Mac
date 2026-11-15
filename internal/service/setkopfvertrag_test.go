package service

// **Die beiden Listen des Setkopfs bleiben zusammen** —
// #1325.
//
// # Wozu
//
// `importschluessel.go` sagt, was **abgelegt** wird; `setkopflesen.go` sagt,
// **wohin**. Wer in einer der beiden einen Schlüssel ergänzt und in der anderen
// nicht, bekommt einen Fehler, den niemand sieht:
//
//   - nur in `gespeichert`: die Zeile wird als verstanden markiert und danach
//     **verworfen** — genau der Fehler, den der Zweitblick über Phase 2 bei
//     `mse version` gefunden hat;
//   - nur in `kopfangaben`: die Zeile wird gelesen **und** reist roh mit, und
//     die Anwendung zeigt sie zweimal.
//
// **Er läuft im Paket**, weil beide Listen paketintern sind.

import "testing"

func TestJedeKopfangabeIstAuchGespeichert(t *testing.T) {
	t.Parallel()

	for schluessel := range kopfangaben {
		if !gespeichert[schluessel] {
			t.Errorf("%q wird gelesen, steht aber nicht in `gespeichert` —"+
				" die Zeile würde verworfen", schluessel)
		}
	}
}

// **Und die Gegenrichtung, soweit sie zu prüfen ist.**
//
// `gespeichert` führt mehr als den Kopf — `game`, `card`, `keyword` und die
// übrigen haben ihren eigenen Weg. Geprüft wird deshalb nur, dass keiner der
// **Kopfschlüssel** dort fehlt; was `gespeichert` sonst führt, prüft der Test
// darüber nicht. Das steht hier, damit niemand mehr hineinliest.
func TestDerSetkopfFuehrtFuenfAngaben(t *testing.T) {
	t.Parallel()

	// **Die Zahl steht hier**, damit eine leere Zuordnung nicht grün ist.
	if len(kopfangaben) != 5 {
		t.Errorf("%d Kopfangaben, erwartet waren 5", len(kopfangaben))
	}
}
