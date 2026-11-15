# 0183 — Im Kartentext prüft das System, und ersetzt wird nichts von selbst

**Stand:** angenommen am 2026-10-10 · G03
(#2250).
**Am nativen Fenster gemessen ist es noch nicht** — was das heißt, steht
unten unter „Was offen bleibt".

## Die Frage

G03 verlangt, MSE2s Rechtschreibhilfe nachzulesen und dann entweder
**native Hilfe samt fachlicher Wörterliste** zu bauen oder eine **klare
Architekturentscheidung** zu treffen — ohne ungewollte Autoersetzung von
Regeltext und Namen, mit Blick auf Rücknahme und Sprachen.

## Was MSE2 tut — nachgelesen, Fassung f9d9356 vom 2021-07-06

- **`check_spelling`** (`src/script/functions/spelling.cpp`) ist eine
  Skriptfunktion. Ein Spiel ruft sie in seinem Textfilter; sie setzt um jedes
  Wort, das kein Wörterbuch kennt, eine Marke `<error-spelling:…>`.
  **Sie ändert keinen Buchstaben.** Wörter in `<sym>`, `<atom>` und
  `<nospellcheck>` prüft sie nicht.
- **Die Wörterbücher sind Hunspell** (`src/util/spell_checker.cpp`, Zeile 61):
  `data/dictionaries/<sprache>.aff/.dic` und dazu eine **Wortliste des
  Spiels**. Im Magic-Paket ist das `magic.mse-game/dictionary/magic-words.<sprache>.dic`,
  **3.366 Zeilen über alle Sprachen** (gezählt am 2026-10-10 in
  `full-magic-pack`, Stand `622160f`).
- **Ob geprüft wird, entscheidet das Set**: Das Magic-Paket ruft
  `check_spelling` nur unter `set.mark_errors` und für Japanisch und
  Chinesisch gar nicht (`magic.mse-game/script`, Zeilen 1972–1978). Dazu kommt
  je Stil ein Schalter in den Einstellungen (`card_spellcheck_enabled`,
  `src/data/settings.cpp`, Zeile 130; Vorgabe an).
- **Korrigiert wird nur auf Wahl**: Ein Rechtsklick auf ein unterringeltes
  Wort zeigt Vorschläge, und wer einen wählt, ersetzt das Wort als **eine
  rücknehmbare Handlung** „correct" (`src/gui/value/text.cpp`, Zeilen
  561–581 und 597–612). **„Zum Wörterbuch hinzufügen" ist dort selbst nur ein
  `TODO`.**
- **Von selbst ersetzt MSE2 nichts.** Es zeichnet seinen Editor selbst, und
  die automatische Korrektur des Systems erreicht ihn nicht.

## Was hier gilt

1. **Ersetzt wird nichts von selbst.** Die Felder einer Karte und der Wert
   einer Massenänderung tragen `autocorrect="off"` und `autocapitalize="off"`
   (`seiten/Feldeingabe.tsx`, `seiten/Massenfelder.tsx`). Ein `textarea` in
   WebKit korrigiert sonst nach der Systemeinstellung, und die ist auf macOS
   ab Werk an. Ein Kartenname, den das System nicht kennt, würde beim Tippen
   still „berichtigt" — genau das, was MSE2 nie tat.
2. **Unterringelt wird nativ**, mit den Wörterbüchern des Systems.
   `spellcheck` bleibt an: Die Unterringelung ändert nichts am Text, und die
   Vorschläge per Rechtsklick sind dieselbe Bedienung wie in MSE2.
3. **`check_spelling` bleibt `adapted`**, ohne Wörterbuch
   (`internal/msescript/rechtschreibung.go`): Alte Marken fallen weg, neue
   kommen nicht. **Die gezeichnete Karte trägt also keine Unterringelung.**

## Was nicht gebaut wird, und warum

**Hunspell samt Spielwortliste.** Das wären eine neue Abhängigkeit für das
Lesen von `.aff/.dic` und eine zweite Unterringelung neben der des Systems.
Beide prüften dasselbe Wort mit verschiedenen Wörterbüchern, und wer sich
nach der einen richtet, bekommt die andere.

**Die Spielwortliste lässt sich der Prüfung des Systems über `learnWord`
nicht beibringen**: es schreibt in das Benutzerwörterbuch des Menschen, für
jedes Programm — eine Einstellung des Systems, und die ändert dieses Programm
nicht. **AppKit kennt daneben Ausnahmen je Dokument**:
`ignoreWord:inSpellDocumentWithTag:` und `setIgnoredWords:inSpellDocumentWithTag:`
gelten für einen Dokumenttag aus `uniqueSpellDocumentTag` und lassen das
Benutzerwörterbuch stehen (nachgelesen im SDK-Header `NSSpellChecker.h` am
2026-10-10). **Nicht nachgewiesen ist, dass dieser Weg diese WebView
erreicht**: die Prüfung im Kartentext führt WKWebView, und einen öffentlichen
Griff auf ihren Dokumenttag haben wir nicht gefunden. Ob es ihn über Wails
gibt, ist eine offene Frage mit eigenem Issue
(#2286,
Befund 7 des Zweitblicks über Phase G) — bis dahin bleibt die Spielwortliste
ungebaut.

**Nachgelesen für #2286
am 2026-10-10 — einen öffentlichen Griff gibt es nicht:**

| Wo | Was dort steht |
| --- | --- |
| SDK, `WebKit.framework/Headers/WKWebView.h` | **nichts** zur Rechtschreibung — keine Eigenschaft, kein Tag |
| SDK, `WebKit.framework/Headers/WebView.h:583` | `spellCheckerDocumentTag` — am **alten** `WebView` aus WebKit 1, nicht an `WKWebView` |
| WebKit, `Source/WebKit/UIProcess/API/Cocoa/WKWebViewPrivate.h:893` (Fassung `9a469598b602` vom 2026-10-10) | `_spellCheckerDocumentTag` — **SPI**: privat, ohne Zusage, mit Unterstrich |
| WebKit, `Source/WebKit/UIProcess/mac/TextCheckerMac.mm:361–386, 546–548` (dieselbe Fassung) | WebKit prüft **im Prozess der Anwendung** über `NSSpellChecker … inSpellDocumentWithTag:` und nimmt einen Tag aus `uniqueSpellDocumentTag`; ein `ignoreWord:` mit **diesem** Tag wirkte also |
| Wails v3.0.0-beta.27, `pkg/application/webview_window.go:1660` und `webview_window_darwin.h:10` | Go bekommt nur das `NSWindow` (`NativeWindow()`); die `WKWebView` hängt als Objective-C-Eigenschaft an Wails' eigener Fensterklasse |

**Der Weg dorthin wäre zweimal nicht öffentlich** — eine Eigenschaft einer
inneren Klasse von Wails und eine SPI von WebKit. Beides kann sich mit jeder
Fassung ändern, ohne dass es jemand ankündigt, und eine Prüfung, die dann
still nicht mehr greift, sähe aus wie eine, die greift. **Deshalb bleibt die
Spielwortliste ungebaut**, und der Preis darunter gilt weiter.

**Der Preis, ausgesprochen:** Fachwörter des Spiels, die das System nicht
kennt, werden unterringelt. Das ändert nichts am Text und lässt sich im
Kontextmenü als „Lernen" annehmen — **vom Menschen und für sein System**,
nicht von diesem Programm.

## Verworfen

- **`spellcheck="false"` auf dem Kartentext.** Ohne jede Hilfe wäre es
  weniger als MSE2.
- **Hunspell in Go nachbauen.** Siehe oben: ein zweites Urteil über dasselbe
  Wort.

## Was offen bleibt

**Am nativen Fenster gemessen**, mit der WebKit-Probe, sobald die Maschine
frei ist:

- Ob WKWebView unter macOS `autocorrect="off"` beachtet. MDN nennt das
  Attribut für `textarea` und `input` und sagt, unter macOS stütze sich ein
  Browser auf die Ersetzungen des Systems — **für WKWebView ist das nicht
  nachgelesen, sondern zu messen.**
- Ob eine Korrektur per Rechtsklick als ein Schritt zurückzunehmen ist.
- In welcher Sprache unterringelt wird. Das System prüft nach seiner Sprache,
  nicht nach der des Spiels.

Bis dahin ist dies eine Entscheidung mit einer **ungemessenen Annahme**, und
genau so steht sie in G03.
