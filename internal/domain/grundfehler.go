package domain

import (
	"errors"
	"fmt"
)

// **Der Satz für Menschen und der Grund fürs Protokoll** — Werkstück 1 von
// Phase 4.13
// (#1352).
//
// # Warum es diese Trennung gibt
//
// **`fmt.Errorf("…: %w", ursache)` setzt die Ursache in den Satz.** Damit
// steht sie in `Error()` — und `Error()` ist genau das, was `internal/app` als
// `Meldung` weiterreicht. Das Fenster übersetzt über den Schlüssel und sieht
// sie nie; **Kommandozeile, MCP und REST haben keinen Katalog** und zeigen sie
// unverändert.
//
// **Am 2026-09-06 gezählt** (Punkt 1 von #1352):
// von 500 Rufen mit einem Wahrzeichen tragen **25** einen vollen Pfad und
// **42** einen zweiten `%w` — zusammen 54 Stellen, an denen ein Dateipfad,
// ein Go-Fehler oder ein SQLite-Text die Oberfläche erreichen kann. §8 sagt:
// „Rohe SQLite-Meldungen erreichen die Oberfläche nie."
//
// # Warum der Grund trotzdem nicht verschwindet
//
// **Wer einen Fehler sucht, braucht ihn.** „Die Bilddatei ließ sich nicht
// öffnen" sagt einem Menschen, was zu tun ist, und einem Suchenden nichts. Der
// Grund bleibt deshalb in der Kette — `errors.Is` findet ihn, und
// `internal/app` schreibt ihn ins Protokoll, bevor die Meldung hinausgeht.
//
// # Warum kein `%w` im Formatstring
//
// **Weil `%w` beides zugleich tut**: es hängt den Fehler in die Kette **und**
// seinen Text in den Satz. Genau diese Kopplung ist der Mangel. `MitGrund`
// löst sie: die Kette bekommt beides, der Satz nur das Eigene.
type grundfehler struct {
	art   error
	satz  string
	grund error
}

// Error gibt das Wahrzeichen und den eigenen Satz — und **nicht** den Grund.
func (g *grundfehler) Error() string { return g.art.Error() + ": " + g.satz }

// Unwrap gibt beide, damit `errors.Is` das Wahrzeichen **und** den Grund
// findet.
//
// **Ein `nil` in der Liste ist erlaubt** und wird von `errors` übergangen —
// deshalb braucht [MitGrund] keinen Zweig für „ohne Grund".
func (g *grundfehler) Unwrap() []error { return []error{g.art, g.grund} }

// MitGrund verbindet ein Wahrzeichen, die technische Ursache und einen Satz
// für Menschen.
//
// **`art` ist eines der Wahrzeichen** aus `fehler.go` nebenan —
// [ErrValidation], [ErrNotFound], [ErrConflict] und die übrigen. `grund` darf
// `nil` sein. `satz` ist, was jemand lesen soll, und wird wie ein
// Formatstring behandelt.
//
// **Der Grund steht vor dem Satz**, obwohl sich das andersherum besser läse:
// die Platzhalter gehören hinter ihren Satz, und ein variadischer Parameter
// steht in Go am Ende. Damit erkennt `go vet` die Funktion außerdem als
// Formatierer und prüft die Platzhalter mit.
func MitGrund(art, grund error, satz string, werte ...any) error {
	return &grundfehler{art: art, satz: fmt.Sprintf(satz, werte...), grund: grund}
}

// GrundVon gibt die verdeckte Ursache — oder nichts.
//
// **Für das Protokoll und nur dafür.** Wer sie in eine Meldung setzte, hätte
// die Trennung von oben wieder aufgehoben.
//
// **Auch aus einem Feldbefund** ([FieldError.MitGrund], ENG-05.AK-03).
func GrundVon(fehler error) error {
	var mit *grundfehler
	if errors.As(fehler, &mit) {
		return mit.grund
	}

	var befund *FieldError
	if errors.As(fehler, &befund) {
		return befund.grund
	}

	return nil
}

// SatzVon gibt den Satz für Menschen — und nur ihn.
//
// # Wofür es ihn gibt
//
// **Für die Stellen, an denen ein Fehler kein Fehler bleibt.** Scheitert die
// halbe Suche, trägt die andere Hälfte weiter; was fehlt, gehört trotzdem in
// die Antwort — und zwar als Satz und nicht als Fehlerkette. `Error()` wäre
// dafür falsch: es stellt das Wahrzeichen voran („Konflikt: …"), und das
// gehört in eine Meldung, nicht in einen Nebensatz.
//
// **Und ein fremder Fehler kommt hier nicht durch.** Wer keinen Satz mitbringt,
// bekommt keinen — `%w`-Ketten tragen Dateipfade und SQLite-Texte, und §8 sagt,
// dass die die Oberfläche nie erreichen.
func SatzVon(fehler error) string {
	var mit *grundfehler
	if !errors.As(fehler, &mit) {
		return ""
	}

	return mit.satz
}
