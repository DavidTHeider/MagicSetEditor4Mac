package service

// **Was ein Klient selbst schreibt** — Claude Code über `claude mcp`, der
// Dienst über `launchctl`, OPS-09b
// (#1961).
//
// # Warum über seinen Befehl
//
// **`~/.claude.json` schreibt Claude Code laufend selbst um** — wer
// daneben schriebe, verlöre gegen den nächsten Lauf. Gelesen wird die Datei
// trotzdem: `claude mcp get` prüft den Server über eine echte Verbindung,
// und ein Stand soll nichts starten.
//
// # Und der LaunchAgent
//
// **Die plist schreibt dieses Programm, laden tut `launchctl`.** Die plist
// ist die Datei des Ziels und sein Block; ihr Stand ist, ob sie Byte für
// Byte die eigene ist.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
)

// agentenstandAus liest den Stand eines LaunchAgents.
//
// **Eine eigene Altlast macht ihn veraltet** (Zweitblick über Phase 4, Befund
// 7): ein Dienst unter dem Etikett von vor 0.9.0 steht sonst neben dem
// heutigen — oder statt seiner —, und niemand sieht ihn. „Einrichten" räumt
// ihn dann, „Widerrufen" nimmt ihn mit.
func agentenstandAus(ziel Schnittstellenziel, inhalt []byte) Schnittstellenstand {
	stand := agentenstandOhneAltlast(ziel, inhalt)
	if (stand == SchnittstelleOffen || stand == SchnittstelleEingerichtet) && len(eigeneAltlasten(ziel)) > 0 {
		return SchnittstelleVeraltet
	}

	return stand
}

// agentenstandOhneAltlast ist der Stand der plist unter dem heutigen Etikett.
func agentenstandOhneAltlast(ziel Schnittstellenziel, inhalt []byte) Schnittstellenstand {
	switch {
	case inhalt == nil:
		return SchnittstelleOffen
	case bytes.Equal(inhalt, []byte(ziel.Block)):
		return SchnittstelleEingerichtet
	case bytes.Contains(inhalt, []byte("<string>"+ziel.Schluessel+"</string>")):
		return SchnittstelleVeraltet
	}

	return SchnittstelleFremd
}

// befehleEinrichten führt die Befehle des Klienten aus.
//
// **Ein veralteter Eintrag geht zuerst**: `claude mcp add` weist einen Namen
// ab, den es schon gibt, und ein geladener LaunchAgent lädt nicht ein
// zweites Mal.
//
// **Und ein Entladen, das scheitert, hält an** (Zweitblick über Phase 6.12,
// Z6). Hier wurde es übergangen; dann scheiterte das Laden am noch geladenen
// Dienst, die plist ging, und ein Dienst lief ohne Datei, die ihn erklärt.
// „Nicht geladen“ ist kein Scheitern — das sagt die Ausführung der App
// (Exit 3, gemessen am 2026-10-01).
func befehleEinrichten(ctx context.Context, ziel Schnittstellenziel, stand Schnittstellenstand) error {
	if stand == SchnittstelleVeraltet {
		if fehler := befehleWiderrufen(ctx, ziel); fehler != nil {
			return fehler
		}
	}

	if ziel.Art == ArtAgent {
		if fehler := PrivaterOrdner(ziel.Begleitordner); fehler != nil {
			return fmt.Errorf("den Ordner für %s anlegen: %w", ziel.Name, fehler)
		}

		if fehler := zielSchreiben(ziel.Datei, []byte(ziel.Block)); fehler != nil {
			return fehler
		}
	}

	if fehler := ausfuehren(ctx, ziel, ziel.Einrichtbefehle); fehler != nil {
		// **Kein halber Stand**: eine plist, die sich nicht laden ließ,
		// geht wieder — **über das Entladen** (Z6). Ob das Laden halb
		// gelang, sagt `launchctl` nicht; lässt der Dienst sich nicht
		// entladen, bleibt die plist, und der Fehler sagt beides.
		if ziel.Art == ArtAgent {
			if entladen := befehleWiderrufen(ctx, ziel); entladen != nil {
				return errors.Join(fehler, entladen)
			}
		}

		return fehler
	}

	return nil
}

// befehleWiderrufen führt die Widerrufbefehle aus — beim LaunchAgent nimmt
// es danach die plist.
func befehleWiderrufen(ctx context.Context, ziel Schnittstellenziel) error {
	if fehler := ausfuehren(ctx, ziel, ziel.Widerrufbefehle); fehler != nil {
		return fehler
	}

	if ziel.Art != ArtAgent {
		return nil
	}

	// **Erst jeden eigenen Dienst entladen, dann die Dateien** (Nachlauf
	// über Phase 4, Befund 7): stand nur der alte da (Zweitblick, Befund 7)
	// und ließ er sich nicht entladen, liefe er weiter — ohne die Ansage, die
	// beide teilen. Scheitert ein Entladen, bleibt jede Datei.
	alte := eigeneAltlasten(ziel)

	for _, alt := range alte {
		if fehler := ausfuehren(ctx, ziel, alt.Widerrufbefehle); fehler != nil {
			return fehler
		}
	}

	plists := []string{ziel.Datei}
	for _, alt := range alte {
		plists = append(plists, alt.Datei)
	}

	for _, plist := range plists {
		if fehler := os.Remove(echterOrt(plist)); fehler != nil && !os.IsNotExist(fehler) {
			return fmt.Errorf("%s entfernen: %w", plist, fehler)
		}
	}

	// **Die Marke geht mit dem Dienst** (SEC02, #2355, Vermerk 0192): er ist
	// entladen, sie gilt nicht mehr — und eine Datei, die eine tote Marke
	// nennt, sähe aus wie ein Zugang.
	if ziel.Ansage != "" {
		if fehler := os.Remove(ziel.Ansage); fehler != nil && !os.IsNotExist(fehler) {
			return fmt.Errorf("die Ansage von %s entfernen: %w", ziel.Name, fehler)
		}
	}

	return nil
}

// ausfuehren führt Befehle der Reihe nach aus und hält beim ersten Fehler.
func ausfuehren(ctx context.Context, ziel Schnittstellenziel, befehle [][]string) error {
	for _, befehl := range befehle {
		if fehler := ziel.Ausfuehren(ctx, befehl); fehler != nil {
			return fmt.Errorf("%s: %q: %w", ziel.Name, befehl[0], fehler)
		}
	}

	return nil
}
