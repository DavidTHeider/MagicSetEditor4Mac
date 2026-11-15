// **Ein Doppelklick im Finder öffnet das Set** —
// #891,
// Werkstück 1 von Phase Y.
//
// # Warum das nicht über `os.Args` geht
//
// macOS übergibt eine doppelgeklickte Datei **nicht** in der Kommandozeile,
// sondern als Apple-Event `odoc` an die laufende Anwendung. Wails nimmt es im
// `NSApplicationDelegate` entgegen und meldet es als Ereignis
// `ApplicationOpenedWithFile` (`application_darwin_delegate.m`,
// `application:openFile:`, v3.0.0-beta.23; die Hülle reicht es hierher,
// `internal/fenster/wails.go`). Wer nur `os.Args` läse, bekäme vom Finder nie
// etwas.
//
// **`os.Args` gilt trotzdem** — aus einer Shell gestartet steht der Pfad dort,
// und das ist der Weg, den ein Skript nimmt. `main.go` reicht ihn an dieselbe
// Stelle.
//
// # Warum das Frontend entscheidet, was danach geschieht
//
// Dieselbe Überlegung wie bei der Menüleiste (`fensternaht.go`): Go öffnet das
// Set und **meldet** die Kennung; welche Seite dann dasteht, ist eine Frage der
// Bedienung und gehört ins Fenster (§2).

package app

// EreignisDateiGeoeffnet meldet, dass ein Set aus einer Datei offen ist.
//
// **Mit der Kennung als Nutzlast** — ohne sie wüsste das Fenster nicht, welches
// Set es zeigen soll, und müsste die Liste durchsuchen.
const EreignisDateiGeoeffnet = "datei:geoeffnet"

// EreignisDateiFehler meldet, dass eine übergebene Datei nicht zu lesen war.
//
// **Nicht stumm bleiben**: wer eine Datei doppelklickt und nichts geschieht,
// sucht den Fehler bei sich (§8).
//
// **Zwei Nutzlasten: der Pfad und der Fehler.** Der Satz entsteht im Fenster —
// Go kennt den Schlüssel, nicht die Sprache der Sitzung (§9). Ein deutscher
// Satz aus Go wäre in einer englischen Sitzung deutsch, und genau das war
// #133.
const EreignisDateiFehler = "datei:fehler"

// DateiOeffnen liest ein übergebenes Setpaket und meldet das Ergebnis.
//
// # Warum sie warten kann
//
// Der Finder schickt das Ereignis **beim Start** — es kann eintreffen, bevor
// die Anwendung `Bereit` bekommen hat und einen Kontext besitzt. Wer
// es dann meldete, meldete ins Leere: es hört niemand zu, und die Datei bliebe
// ungeöffnet, ohne dass irgendetwas fehlschlüge. Der Pfad wird deshalb
// aufgehoben und in [App.Bereit] nachgeholt.
//
// **Jeder Aufruf zählt.** Wails meldet jede Datei einzeln; wer zwei Sets
// zugleich aufschlägt, bekommt zwei Aufrufe. Alle werden geöffnet, und der
// letzte gewinnt die Anzeige — dieselbe Ordnung wie beim Öffnen über den
// Dialog.
//
// **Vor dem Fenster werden sie gesammelt.**
func (a *App) DateiOeffnen(pfad string) {
	if pfad == "" {
		return
	}

	if a.fenster == nil {
		a.wartendeDateien = append(a.wartendeDateien, pfad)

		return
	}

	// **Nacheinander mit einem Tausch des Bestands** (Nachlauf über Phase G,
	// zweiter Teil, Befund 3, #2292). Ein Tausch wartet laufende Aufrufe nicht
	// ab; las der Doppelklick noch in den alten Bestand, kam seine Kennung
	// nach dem Neuanfang an, und das Fenster schlug sie im neuen auf. Unter
	// derselben Sperre wie der Tausch ([App.imTausch]) wartet er ihn ab und
	// liest in den neuen — oder der Tausch wartet, bis die Kennung gemeldet
	// ist, und kommt damit **nach** ihr im Fenster an.
	a.lauf.wechsel.Lock()
	defer a.lauf.wechsel.Unlock()

	// **Nach dem Beginn des Schlusses nichts mehr** — der Bestand geht zu.
	if a.lauf.tor.Load() == torEndet {
		return
	}

	antwort := a.SetOeffnen(a.fenster, pfad)
	if antwort.Fehler != nil {
		a.melden(a.fenster, EreignisDateiFehler, pfad, antwort.Fehler)

		return
	}

	// **Mit der Epoche des Bestands** (J05, #2269): gelesen unter der Sperre
	// des Tauschs, also die, in der das Set eingelesen wurde.
	a.melden(a.fenster, EreignisDateiGeoeffnet, antwort.SetID, a.lauf.epoche.Load())
}

// wartendeNachholen öffnet die Dateien, die vor dem Fenster kamen.
//
// **In der Reihenfolge, in der sie kamen**, und damit gewinnt die letzte die
// Anzeige — so wie bei Aufrufen nach dem Fenster.
func (a *App) wartendeNachholen() {
	pfade := a.wartendeDateien
	a.wartendeDateien = nil

	for _, pfad := range pfade {
		a.DateiOeffnen(pfad)
	}
}
