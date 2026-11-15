// **Die Zeichenkette des MSE-Script-Lexers** — abgetrennt am 2026-08-26,
// weil `lexer.go` über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Sie ist das einzige Merkmal mit eigener Grammatik: ein Rückstrich flieht das
// nächste Zeichen, und `{…}` bettet einen Ausdruck ein. Alles andere im Lexer
// ist ein Zeichen oder eine Folge gleichartiger; das hier ist eine kleine
// Sprache in der Sprache.

package msescript

import "strings"

func (l *lexer) zeichenkette() (Merkmal, error) {
	zeile, spalte := l.zeile, l.spalte

	// **Der Rückstrich flieht das nächste Zeichen** — belegt in den
	// Exportvorlagen von `M15-Magic-Pack`: fünf schreiben `\"` mitten in einer
	// Zeichenkette, und ohne diese Regel endete sie dort. `docs/FORMATE.md`
	// führte die Frage seit Phase 3 als offen
	// (#51).
	//
	// **Und `\\` ist ein Rückstrich.** Dieselben Vorlagen schreiben `"\\n"`,
	// wenn sie die zwei Zeichen eines Regex meinen — das ergibt nur einen Sinn,
	// wenn ein einfaches `\n` etwas anderes ist.
	var gebaut strings.Builder

	ende := l.stelle + 1

	// **Umbrüche werden mitgezählt.** Eine Zeichenkette darf über Zeilen gehen
	// (siehe unten); ohne diese Zählung zeigte jeder Fehler danach auf eine
	// falsche Zeile — genau der Schaden, den der Zweitblick über Phase 6 an der
	// Fluchtfolge vor dem Zeilenende gefunden hat.
	umbrueche, letzterUmbruch := 0, -1

	// **Die geschweiften Klammern werden mitgezählt**
	// (#468).
	//
	// MSE2 hält dafür einen Klammerstapel: `"` legt `BRACE_STRING` auf
	// (`src/script/parser.cpp`, Zeilen 220–222), ein `{` in der Zeichenkette
	// kehrt zum normalen Lesen zurück (Zeilen 302–307) — **der Stapel behält
	// `BRACE_STRING`** —, ein `"` darin legt einen **zweiten** darauf
	// (Zeilen 268–272), und erst das `}` liest die äußere weiter (Zeilen
	// 224–228). Fassung f9d9356 vom 2021-07-06.
	//
	// Hier steht statt des Stapels eine Zählung, weil dieses Programm die
	// Zeichenkette als **eine** Marke liest und erst `textbaustein.go` sie
	// zerlegt: was zwischen den Klammern steht, kommt roh in die Marke und
	// wird dort für sich zerteilt.
	tiefe := 0

	for ende < len(l.zeichen) && (tiefe > 0 || l.zeichen[ende] != '"') {
		// **Ein Rückstrich flieht das nächste Zeichen** — auch einen Umbruch.
		// Fünf Exportvorlagen schreiben `\"` mitten in einer Zeichenkette.
		//
		// **Und er steht vor der Klammerzählung**: `\{` ist eine geschweifte
		// Klammer und keine Einbettung.
		if l.zeichen[ende] == '\\' && ende+1 < len(l.zeichen) {
			if l.zeichen[ende+1] == '\n' {
				umbrueche++
				letzterUmbruch = ende + 1
			}

			ende++
			gebaut.WriteString(geflohen(l.zeichen[ende]))

			ende++

			continue
		}

		// **Eine Zeichenkette in der Einbettung wird roh übernommen.** Sie ist
		// Teil eines Ausdrucks, den `gelesenerAusdruck` noch einmal zerteilt —
		// wer hier ihre Fluchtfolgen auflöste, gäbe dem zweiten Lauf einen
		// anderen Text als den geschriebenen.
		if l.zeichen[ende] == '"' {
			schluss, geschlossen := l.eingebetteteZeichenkette(ende)

			// **Eine Zeichenkette ohne Ende ist keine.** `"[\x{41]"` schreibt
			// eine Angabe für den Mustervergleich, keine Einbettung, und
			// `"a{b"` ist ein Kartentext mit halber Klammer. MSE2 bricht in
			// beiden Fällen ab (`Unexpected end of input in string constant`,
			// `src/script/parser.cpp`, Zeile 260, Fassung f9d9356 vom
			// 2021-07-06); dieses Programm liest
			// weiter, so weit es verantwortbar ist (§8) — die äußere
			// Zeichenkette endet hier, und was ohne Partner dasteht, ist Text.
			if !geschlossen {
				break
			}

			for stelle := ende; stelle <= schluss; stelle++ {
				if l.zeichen[stelle] == '\n' {
					umbrueche++
					letzterUmbruch = stelle
				}

				gebaut.WriteRune(l.zeichen[stelle])
			}

			ende = schluss + 1

			continue
		}

		switch l.zeichen[ende] {
		case '{':
			tiefe++

		case '}':
			if tiefe > 0 {
				tiefe--
			}

		case '\n':
			umbrueche++
			letzterUmbruch = ende
		}

		gebaut.WriteRune(l.zeichen[ende])
		ende++
	}

	if ende >= len(l.zeichen) {
		return Merkmal{}, NeuerFehler(zeile, spalte, SchluesselOffeneZeichenkette)
	}

	inhalt := gebaut.String()
	breite := ende - l.stelle + 1

	l.stelle += breite

	if umbrueche > 0 {
		l.zeile += umbrueche
		l.spalte = ende - letzterUmbruch + 1
	} else {
		l.spalte += breite
	}

	// Die Breite zählt die Anführungszeichen mit, der Text nicht.
	return Merkmal{
		Art: ArtZeichenkette, Text: inhalt,
		Zeile: zeile, Spalte: spalte, Breite: breite,
	}, nil
}

// eingebetteteZeichenkette sucht das Ende einer Zeichenkette in einer
// Einbettung. `auf` steht auf ihrem Anführungszeichen.
//
// Sie **schreibt nichts**: erst wenn feststeht, dass die Zeichenkette
// schließt, wird sie übernommen. Ohne diese Trennung verschluckte ein
// `"[\x{41]"` seinen Rest, bevor der Auffangfall greifen könnte.
func (l *lexer) eingebetteteZeichenkette(auf int) (int, bool) {
	for stelle := auf + 1; stelle < len(l.zeichen); stelle++ {
		if l.zeichen[stelle] == '\\' && stelle+1 < len(l.zeichen) {
			stelle++

			continue
		}

		if l.zeichen[stelle] == '"' {
			return stelle, true
		}
	}

	return 0, false
}
