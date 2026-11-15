package main

// **Geschrieben ist geschrieben — im Dienst** (§7, Regel 3;
// #1960,
// #1970).
//
// Scheitert nach einer geglückten Änderung nur das Nachlesen, antwortet der
// Dienst 200 und trägt den Grund unter `nachladefehler` — **in der Gestalt
// jeder Fehlerantwort** ([rest.Befund]). Ein 4xx oder 5xx hieße „nicht
// geschehen", und wer daraufhin wiederholte, liefe gegen den geschriebenen
// Stand.

import (
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/app"
	"github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest"
)

// alsNachladebefund gibt den Nachladefehler als Befund — `nil`, wenn keiner
// da ist. **Der Satz ist schon übersetzt** (§8).
func alsNachladebefund(nachladen *app.Fehler) *rest.Befund {
	if nachladen == nil {
		return nil
	}

	return &rest.Befund{Code: string(nachladen.Code), Meldung: nachladen.Meldung}
}
