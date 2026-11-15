package app

// **Die Zeilen eines Dienstes im Protokoll des Fensters** — Nachlauf über
// Phase 6.4, N3.
//
// Ein Dienst, den das Fenster startet, schreibt sein Protokoll geschwärzt und
// mit Vorgangskennung auf die Standardfehlerausgabe (`Serverprotokoll`). Ohne
// Empfänger ging sie ins Nullgerät, und die Kennung, die eine Antwort nennt,
// führte ins Leere. Jetzt landet jede Zeile in `mse.log` — **als Wert**, damit
// sie auch dort nach der Kennung zu finden ist.

import (
	"bytes"
	"log/slog"
	"sync"
)

// hoechsteZeile ist, was ohne Zeilenende gesammelt wird, bevor es trotzdem
// geschrieben wird — ein Dienst, der nie umbricht, soll den Speicher des
// Fensters nicht füllen.
const hoechsteZeile = 64 << 10

type kindprotokoll struct {
	dienst    string
	schreiber *slog.Logger

	schloss sync.Mutex
	rest    []byte
}

func neuesKindprotokoll(dienst string, schreiber *slog.Logger) *kindprotokoll {
	return &kindprotokoll{dienst: dienst, schreiber: schreiber}
}

// Write nimmt, was der Dienst schreibt, und gibt jede ganze Zeile weiter.
func (k *kindprotokoll) Write(stueck []byte) (int, error) {
	k.schloss.Lock()
	defer k.schloss.Unlock()

	k.rest = append(k.rest, stueck...)

	for {
		ende := bytes.IndexByte(k.rest, '\n')
		if ende < 0 && len(k.rest) < hoechsteZeile {
			break
		}

		if ende < 0 {
			ende = len(k.rest)
		}

		if zeile := string(bytes.TrimSpace(k.rest[:ende])); zeile != "" {
			k.schreiber.Info("ein Dienst meldet", "dienst", k.dienst, "zeile", zeile)
		}

		k.rest = k.rest[min(ende+1, len(k.rest)):]
	}

	return len(stueck), nil
}
