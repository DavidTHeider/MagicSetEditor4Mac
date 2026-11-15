# 0133 — Seiten bleiben stehen

**Stand:** entschieden am 2026-09-24
(#1675, UX-03.AK-02).

## Was entschieden ist

**Jede besuchte Seite wird aufgehoben, statt beim Wiederkommen neu zu
entstehen.** `Seiteninhalt` zeichnet jede Anwendungsseite, die in dieser
Sitzung dastand, in einer eigenen `<Activity>` (React 19.2); verborgen ist
jede außer der aktuellen. Dasselbe gilt für die Laschen eines Sets und für
jede Schublade samt ihrer Laschen (`seiten/Setbereich.tsx`). Welche das sind,
führt `haken/useAufgehobeneSeiten.ts`.

**Dazu der Gang** (`haken/useSeitengang.ts`): ⌘[ und ⌘] — die Kürzel, die
der Finder für den vorigen und den nächsten Ordner führt (Apple,
„Mac-Tastaturkurzbefehle", support.apple.com/de-de/102650, nachgelesen am
2026-09-24) —, und dieselben unter *Darstellung*, wo schon die Seitenleiste
steht. Die Einträge heißen „Zurück" und „Vorwärts", wie die
Systemeinstellungen „Back" und „Forward" übersetzen (`Localizable.loctable`
in ihrem Programmbündel). Jeder Schritt verlässt zuerst das offene Feld (`feldwegschreiben.ts`,
#1290), denn ein Kartenfeld schreibt beim Verlassen, und ein Kürzel wechselte
die Seite ohne `blur`.

## Warum

UX-03.AK-02 verlangt: *„Zurück/Vorwärts, Direktaufruf und Wiederaufnahme
funktionieren ohne Verlust ungespeicherter Eingaben."* **Eine Seite hält mehr
als ihre Felder** — einen halb getippten Decknamen, einen aufgeklappten
Kasten, eine gewählte Zeile, eine Rückfrage. Bis hierher war das alles fort,
sobald man die Seite verließ.

**Was `<Activity>` dabei tut, und was das kostet:**

- Der **Zustand** eines verborgenen Baums bleibt; seine **Effekte ruhen**:
  Horcher am Fenster, Abfragen in Abständen, die Zeichenfläche des Tisches
  werden abgebaut und beim Wiederkommen neu aufgebaut. **Beim Wiederkommen
  laden die Seiten deshalb neu**, wie beim ersten Mal — was ein anderer
  Prozess inzwischen geändert hat, steht dann da.
- Der **DOM-Baum** bleibt, mit `display: none`. Rollenabfragen übergehen ihn,
  ein Vorleseprogramm auch; eine **Textabfrage im Test** nicht — dafür gibt es
  `test/verborgen.ts`.
- Ein verborgener Baum wird bei einer Änderung seiner Eigenschaften **mit
  niedriger Dringlichkeit weiter gezeichnet**. Das kostet Arbeit, die vorher
  nicht anfiel; gemessen ist sie nicht.

## Was nicht aufgehoben wird

- **Die Seiten eines Sets, wenn das Set wechselt** — und der Gang dorthin
  (AGENTS.md §7, Regel 1). *Karten* hieß vorhin „die Karten von Set A".
- **Nichts über einen Profilwechsel hinweg.** `App` baut die Anwendung dann
  neu (`key`); ein Lauf belegt es mit einer Gegenprobe
  (`profilwechsel.test.tsx`, „vergisst, was die verlassenen Seiten
  aufhoben").
- **Nichts über ein Zurückspielen hinweg** — außer der Seite, auf der es
  geschah. Es tauscht den ganzen Bestand, und `useBestandspflege` verlangt,
  den offenen Kontext zu räumen (§7, Regel 1); geräumt wurden Set und Karte,
  die verlassenen Seiten standen mit dem Zustand des ersetzten Bestands da.
  **Hier stand der Satz nicht**, und das war der Befund M9 (a) des
  Zweitblicks über Phase 6.3
  (#1666).
  Jetzt zählt `Bestandspflege.ersetzt`, wie oft zurückgespielt wurde, und
  ein neuer Stand räumt alles Aufgehobene (`useAufgehobeneSeiten.ts`). **Die
  Seite selbst bleibt**, weil sie den Ort des alten Bestands nennt; die
  Einstellungen liest sie neu (`appbestandspflege.test.tsx`). Belegt in
  `rueckspielseiten.test.tsx`.
- **Nichts über das Beenden hinweg.** Ein Formular, das bis ⌘Q nicht
  abgeschickt ist, ist fort; ein Kartenfeld, in dem beim ⌘Q noch der Cursor
  steht, ebenso, wenn WebKit beim Schließen kein `blur` mehr schickt — **das
  ist nicht gemessen**.

## Verworfen

- **Ein Entwurfsspeicher je Feld.** Über sechzig freie Eingaben
  (`docs/FELDPRUEFUNG.md`) hätten ihn bekommen müssen, und jede künftige
  ebenso — sonst wäre er die halbe Zusage (§10). Wo ein Entwurf die Seite
  **und** den Wechsel der Datei überleben soll, gibt es ihn trotzdem: die
  Vorlagenwerkstatt hebt ihn je Datei auf (`useVorlagenpuffer.ts`), weil dort
  eine andere Datei in derselben Seite aufgehen kann.
- **Nur eine begrenzte Zahl von Seiten aufheben.** Wer die älteste verwürfe,
  verlöre still, was dort stand — genau den Fehler, um den es geht. Es gibt
  zwölf Anwendungsseiten und dreizehn Setseiten (`navigation.ts`, gezählt am
  2026-09-24); mehr als das wird nie aufgehoben.
- **Pfeile in der Kopfleiste.** Der Entwurf der Zieloberfläche zeichnet
  keine, und eine Abweichung davon bräuchte einen Grund, den es nicht gibt:
  sichtbar sind die Kürzel in *Darstellung* und in der Kürzeltafel.
