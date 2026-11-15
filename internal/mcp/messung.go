package mcp

// **Jeder Werkzeugruf wird gemessen** — OPS-05.AK-01/02
// (#1722).
//
// **Mit dem Kontext, den die Ausstattung gerichtet hat**: dort steht die
// Vorgangskennung. Ein Ruf, der nie dort ankam — ein unbekanntes Werkzeug,
// kaputte Argumente —, wird mit dem Kontext des Stroms gemessen und hat keine.

import (
	"context"
	"errors"
	"time"

	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain"
)

// UnbekanntesWerkzeug steht für einen Ruf, der kein Werkzeug trifft — **ein**
// Wert, damit erfundene Namen die Spur nicht füllen.
const UnbekanntesWerkzeug = "unbekannt"

// Messung ist ein gemessener Werkzeugruf.
type Messung struct {
	Weg        string
	Schreibend bool
	Dauer      time.Duration
	Ausgang    domain.Rufausgang
}

// MitMessung setzt die Naht, die jeden Werkzeugruf aufnimmt.
func (s *Server) MitMessung(messen func(context.Context, Messung)) *Server {
	s.messen = messen

	return s
}

// messlauf sammelt, was über einen Ruf bekannt wird, und meldet es am Ende.
type messlauf struct {
	ctx     context.Context
	beginn  time.Time
	messung Messung
}

func neuerMesslauf(ctx context.Context) *messlauf {
	return &messlauf{
		ctx: ctx, beginn: time.Now(),
		messung: Messung{Weg: UnbekanntesWerkzeug, Ausgang: domain.RufAbgewiesen},
	}
}

// werkzeug trägt ein, welches Werkzeug getroffen ist.
func (l *messlauf) werkzeug(werkzeug Werkzeug) {
	l.messung.Weg, l.messung.Schreibend = werkzeug.Name, werkzeug.Art != Lesend
}

// melden gibt die Messung an die Naht.
func (s *Server) melden(lauf *messlauf) {
	if s.messen == nil {
		return
	}

	lauf.messung.Dauer = time.Since(lauf.beginn)
	s.messen(lauf.ctx, lauf.messung)
}

// ausgangVon ordnet den Fehler eines Werkzeugs zu: **was der Aufrufer anders
// stellen müsste, ist abgewiesen**, alles andere gescheitert.
func ausgangVon(fehler error) domain.Rufausgang {
	if errors.Is(fehler, domain.ErrValidation) || errors.Is(fehler, domain.ErrNotFound) ||
		errors.Is(fehler, domain.ErrConflict) {
		return domain.RufAbgewiesen
	}

	return domain.RufGescheitert
}

// MitKennung sagt dem Server, wie er die Vorgangskennung eines Rufs findet.
//
// **Ohne sie steht keine im Fehlertext** — und das ist der Fall in jedem Test,
// der einen Server ohne Wurzel baut. Ein erfundener Wert wäre schlechter als
// keiner: er zeigte auf eine Protokollzeile, die es nicht gibt.
func (s *Server) MitKennung(kennung func(context.Context) string) *Server {
	s.kennung = kennung

	return s
}

// mitKennung hängt den Griff ins Protokoll an einen Fehlertext.
//
// **Nur an Fehler.** Ein geglückter Ruf gibt Nutzdaten zurück, und ein Modell,
// das sie liest, hat für einen Protokollgriff keine Verwendung; ein Fehler ist
// die Stelle, an der jemand nachsieht.
func (s *Server) mitKennung(ctx context.Context, text string) string {
	if s.kennung == nil {
		return text
	}

	if kennung := s.kennung(ctx); kennung != "" {
		return text + " (Vorgang: " + kennung + ")"
	}

	return text
}
