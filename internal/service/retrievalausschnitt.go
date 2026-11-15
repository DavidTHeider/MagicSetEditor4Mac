package service

// **Der Ausschnitt eines Treffers** — Werkstück A15/A16 von Phase 5.3.
//
// # Warum eine eigene Datei
//
// **Weil `retrievaldienst.go` über die Zeilengrenze aus §2 ging.** Der Schnitt
// folgt der Frage: dort steht, **wie zwei Trefferlisten zusammenkommen**
// ([0103](../../docs/entscheidungen/0103-wie-zwei-trefferlisten-zusammenkommen.md)),
// hier, **womit ein Treffer erklärt wird**.

import "strings"

// Ausschnittzeichen ist die Länge eines Rückfallausschnitts.
//
// **Zweihundert Zeichen, und die Zahl ist eine Wahl.** Sie liegt in der
// Größenordnung dessen, was FTS5 aus einer Fundstelle schneidet; wer sie
// ändert, ändert nichts an einer Zusage.
const Ausschnittzeichen = 200

// anfangsausschnitt gibt den Anfang eines Abschnitts.
//
// # Warum es ihn gibt
//
// **Weil ein Treffer, den nur die Vektorhälfte fand, keine Fundstelle hat.**
// FTS5 schneidet eine heraus — der Cosinus kennt keine: er vergleicht zwei
// Vektoren, und wo im Text die Ähnlichkeit liegt, sagt er nicht.
//
// **Gesehen am 2026-09-12 an einem gebauten Index**: `mse recherche` gab vier
// Treffer und vier leere Zeilen darunter. Ein Treffer ohne Ausschnitt sagt
// nicht, warum er einer ist, und das Issue verlangt ihn ausdrücklich.
//
// **Und er behauptet nicht, die Fundstelle zu sein.** Er ist der Anfang des
// Abschnitts; wo eine echte Fundstelle vorliegt, bleibt die stehen — das hält
// `retrievalausschnitt_intern_test.go` mit einer Gegenprobe.
func anfangsausschnitt(text string) string {
	gekuerzt := strings.TrimSpace(text)

	if len([]rune(gekuerzt)) <= Ausschnittzeichen {
		return gekuerzt
	}

	return strings.TrimSpace(
		string([]rune(gekuerzt)[:Ausschnittzeichen])) + Kuerzungszeichen
}
