package app

// **Wie ein übersetzter Fehler gebaut wird** — herausgelöst mit OPS-10
// (#1908), weil `fehler.go` mit dem neuen Code über die Zeilengrenze aus §2
// ging. Der Schnitt folgt der Frage: dort steht, **welcher** Fehler welche
// Übersetzung bekommt, hier, **woraus** sie besteht.

import "github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"

// einfach baut eine Übersetzung mit dem Rückfallsatz.
func einfach(code Fehlercode, schluessel domain.Schluessel) *Fehler {
	return &Fehler{Code: code, Schluessel: schluessel, Meldung: rueckfallmeldung}
}

// mitSatz ist dasselbe, behält aber den Satz aus Go.
//
// **Wofür**, steht am Aufrufort: die Flächen ohne Katalog sollen lesen können,
// was schiefging.
//
// **Kein Zweig für „leerer Satz"**, und kein Aufruf erreichte ihn: hierher
// kommt nur, wer eines der drei Wahrzeichen umschließt, und deren eigener Text
// steht in jedem `Error()`. Ein Zweig ohne Fall wäre eine Behauptung darüber,
// was im Ernstfall geschieht (§3).
func mitSatz(code Fehlercode, schluessel domain.Schluessel, fehler error) *Fehler {
	return &Fehler{Code: code, Schluessel: schluessel, Meldung: fehler.Error()}
}
