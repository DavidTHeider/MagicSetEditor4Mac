// Diese Datei bringt einen Fehler vor das Fenster, das es noch nicht gibt.
//
// # Warum es sie gibt
//
// `main.lauf` gab einen Startfehler zurück, `main` schrieb ihn nach stderr —
// und **eine doppelgeklickte `.app` hat kein stderr, das jemand sieht.** Das
// Symbol hüpfte einmal und war weg: kein Fenster, keine Meldung, kein Hinweis
// (#797).
//
// Damit war jede Mühe um eine gute Startmeldung folgenlos.
// #795 hatte
// gerade dafür gesorgt, dass ein Bestand aus einer neueren Fassung **benannt**
// abgewiesen wird — gelesen hatte diesen Satz noch nie jemand.
//
// # Warum nicht über Wails
//
// Unter Wails 2 verlangte `runtime.MessageDialog` den Kontext aus `OnStartup`
// oder `OnDomReady`; vor `wails.Run` gab es keinen (nachgelesen in der
// Wails-Dokumentation, §1). Unter Wails 3 zeigt ein Dialog sich über die
// laufende Anwendung (`application.App.Dialog`), und die entsteht erst in
// `fenster.Lauf` — ein Startfehler kommt davor.

package app

import (
	"fmt"
	"os"
)

// startmelder ist die Naht zum Betriebssystem.
//
// Als Variable, damit ein Test sie ersetzen kann: sonst öffnete jeder Lauf ein
// Fenster, und die Prüfung bliebe stehen, bis jemand klickt. Dieselbe Bauform
// wie `meldenEcht` daneben.
var startmelder = meldungZeigen

// Startfehler zeigt einen Fehler, der vor dem Fenster auftrat.
//
// **Ohne Rückgabe.** Der Aufrufer steht kurz vor dem Abbruch; ein zweiter
// Fehler, den er dann noch behandeln müsste, hätte keinen Weg mehr. Was hier
// schiefgeht, geht nach stderr — mehr ist an dieser Stelle nicht zu holen, und
// das steht hier, statt es zu verschweigen.
func Startfehler(fehler error) {
	if fehler == nil {
		return
	}

	if meldefehler := startmelder(
		"MagicSetEditor4Mac", startsatz(fehler)); meldefehler != nil {
		fmt.Fprintf(os.Stderr,
			"Der Start scheiterte: %v\nUnd die Meldung darüber ebenso: %v\n",
			fehler, meldefehler)
	}
}

// startsatz baut den Text des Dialogs — **erst der nächste Schritt, dann der
// Grund**.
//
// # Warum hier ein fester Satz steht und kein Katalogeintrag
//
// **Weil es noch keine Sprache gibt.** Der Katalog lebt im Fenster, und das
// gibt es an dieser Stelle nicht — die Sprache wird beim Aufbau der Oberfläche
// gewählt. Ein Satz auf Deutsch ist hier ehrlicher als einer, der so tut, als
// wäre er übersetzt.
//
// # Und warum der Grund trotzdem darunter steht
//
// **Weil die Anwendung danach weg ist.** `main` ruft diesen Weg und beendet
// sich; es gibt kein Fenster mehr, in dem jemand nachsehen könnte — anders als
// bei [Startbefund], wo die Anwendung läuft und der Satz aus dem Katalog kommt.
//
// # Der Verweis nennt die Protokolldatei
//
// **Das Protokoll gibt es wirklich**: `main` legt es vor dem Bestand an, und
// der Startfehler steht darin.
func startsatz(fehler error) string {
	return "MagicSetEditor4Mac konnte nicht starten.\n\n" +
		"Der Bestand liegt unter »Benutzer → Library → Application Support → " +
		"MagicSetEditor4Mac«. Prüfe, ob dieser Ordner erreichbar ist, und " +
		"starte erneut. Hilft das nicht, liegt im selben Ordner die Datei " +
		"»" + Protokollname + "« — sie und der Grund unten gehören in jede " +
		"Rückfrage.\n\nGrund: " + fehler.Error()
}
