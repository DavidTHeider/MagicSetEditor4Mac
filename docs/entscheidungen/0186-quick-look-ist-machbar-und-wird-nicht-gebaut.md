# 0186 — Quick Look für Setdateien ist machbar und wird in 0.8.0 nicht gebaut

**Stand:** angenommen am 2026-10-11 · G06
(#2253).

## Die Frage

Soll der Finder für eine `.mse-set` ein Vorschaubild zeigen — die erste
Karte statt des Dokumentsymbols? G06 verlangt zuerst einen Explorationsbeleg
zu Dateipaket, Erweiterung, Größen- und Signaturkosten und Datenschutz, dann
eine Entscheidung.

## Was gemessen ist — am 2026-10-11

Im Lume-Gast mit macOS 15.2 ([`belege/g06-quicklook.md`](../belege/g06-quicklook.md),
`node scripts/quicklookgast.mjs`): eine minimale Vorschaubild-Erweiterung
(`com.apple.quicklook.thumbnail`), **ad hoc signiert** wie die Anwendung, mit
Sandbox, in einem Probebündel, das den Typ `.mse-set` als Zip-Archiv
erklärt.

| Frage | Antwort |
| --- | --- |
| **Lädt macOS eine ad hoc signierte Erweiterung?** | **Ja.** `QLThumbnailGenerator` antwortet vorher mit einem Symbol (256 × 256), nachher mit dem Vorschaubild der Erweiterung (183 × 256, Kartenformat) |
| **Sieht sie die Vorlagenpakete?** | **Nein.** Aus der Sandbox ist `~/Library/Application Support/MagicSetEditor4Mac/pakete` nicht lesbar |
| **Was wiegt sie?** | 68 KB ohne Renderer. **Mit Renderer nicht gemessen** — zum Vergleich: `mse`, das den ganzen Weg samt Renderer trägt, wiegt 53 MB; was eine Erweiterung mit eingebettetem Renderer und ihren Ressourcen wöge, sagt das nicht |
| **Netz?** | **Nicht gemessen.** Nach den Regeln der App-Sandbox braucht ausgehendes Netz die Berechtigung `network.client`, und die Probe trägt sie nicht; gebraucht würde es nicht |

**Und `.mse-set` hat heute keinen eigenen Typbezeichner** — `Info.plist`
nennt nur die Endung; eine Erweiterung braucht einen (`QLSupportedContentTypes`),
und weil das Format MSE2 gehört, wäre es eine **importierte** Erklärung.

## Die Entscheidung

**Machbar, aber nicht in 0.8.0.** Eine Vorschau, die die erste Karte zeigt,
braucht die Stile des Sets, und die liegen außerhalb der Sandbox. Die Wege
dahin haben jeder einen Preis, den 0.8.0 nicht zahlen soll:

- **Die Stile in die Erweiterung legen** — nur die mitgelieferten; ein Set
  mit einem bezogenen Stil zeigte dann ein Symbol oder ein falsches Bild.
- **Eine Ausnahmeberechtigung für den Anwendungsordner**
  (`temporary-exception`) — eine Lücke in der Sandbox, die jeder Prozess
  der Erweiterung erbt, für ein Vorschaubild.
- **Ein gemeinsamer Behälter mit der Anwendung** (App Group) — dessen
  Kennung trägt nach Apples Regeln die Team-Kennung vorn, und die gibt es
  ohne Entwicklerkonto nicht ([0003](0003-keine-notarisierung.md));
  **nicht gemessen**, ob ein ad hoc signiertes Paar einen Behälter teilt.
- **Den Renderer einbetten** — ungemessen viel mehr im Bündel (zum Vergleich: `mse` wiegt 53 MB), ein zweiter
  Bauweg mit cgo und Swift, und eine Fläche, die Dateien von fremder Hand
  im Finder liest, sobald jemand ein Verzeichnis öffnet.

Ohne Vorschau ändert sich nichts: der Finder zeigt das Dokumentsymbol.

## Verworfen

- **Bauen, ohne zu messen, ob es lädt.** Die erste Frage war, ob eine ad hoc
  signierte Erweiterung überhaupt geladen wird — vermutet war das Gegenteil.
- **Eine Vorschau, die nur den Titel zeigt.** Sie wäre ladbar und klein, aber
  keine Antwort auf die Frage, die G06 stellt: die erste Karte.
