package app

import (
	"testing"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage"
)

// **Ein Set, das der Bestand nicht führt, hat keinen Ort** — und keinen
// Fehler.
//
// Seit #1815 erreicht keiner der beiden Rufer diesen Fall: beide lesen vorher
// die Karten des Sets, und ein fremdes oder fehlendes ist dort „nicht
// gefunden". Die Hilfe hält ihre Zusage trotzdem allein — fiele die Frage
// davor einmal weg, gäbe sie keinen erfundenen Ort zurück.
func TestEinUnbekanntesSetHatKeinenOrt(t *testing.T) {
	t.Parallel()

	anwendung := &App{}
	anwendung.MitBestand(storage.NeuerSetLeser(probedatenbank(t)))

	if ort := anwendung.setort(probekontext(t), 999); ort != "" {
		t.Errorf("ein unbekanntes Set hat den Ort %q", ort)
	}
}
