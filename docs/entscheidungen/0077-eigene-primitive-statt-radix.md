# 0077 — Eigene Interaktionsprimitive statt Radix

**Stand:** angenommen
**Datum:** 2026-09-08
**Anlass:** Werkstück U2 von Phase 5.1
(#1413) —
das Issue verlangt, Radix Primitives gegen die bestehenden Bausteine zu
bewerten und dabei Version, Peerbereiche, Wails-WebKit, Lizenz und Größe zu
**belegen**, bevor Field, Inspector, Dialog, Popover, Menü, Tooltip und Select
entstehen.

## Lage

**Was heute ist, ist gemessen** (2026-09-08, `frontend/src/bausteine/`):

| | |
| --- | ---: |
| Bausteine mit ARIA-Muster | **4** — `Laschen` (tablist, Pfeile, Fokus wandert), `Aufklappkasten` (aria-expanded, aria-controls nur bei Kasten), `Knopf` (bezeichnung, angehalten), `Pflichtfeld` (label/htmlFor, aria-invalid, aria-describedby) |
| Modale Dialoge | **2** — die Befehlspalette und die Einführung, beide von Hand: `role="dialog"`, `aria-modal`, Escape am Fenster, Fokus auf das erste Feld. **Ohne Fokusrückgabe und ohne Fokusfalle** |
| Rückfragen an Ort und Stelle | **10** (Set, Karte, Deck, Turnier, Nutzer, Profil, Stil, Ersetzen, Sicherung, Verschlüsselung), jede als eigene Handvoll Knöpfe — **kein** gemeinsamer Baustein |
| Menüs, Popover, Tooltips | **0** |
| Auswahlfelder | **51** native `<select>` in 36 Dateien |

**Und Radix ist gemessen**, in einem Scratch-Projekt mit dem gepinnten
Stand — React 19.2.8, jsdom 30.0.1, vitest 4.1.10, @testing-library/react
16.3.2, @testing-library/user-event 14.6.3:

| | |
| --- | --- |
| Paket | `radix-ui` **1.6.7**, MIT |
| Peerbereich | `react ^16.8 \|\| ^17 \|\| ^18 \|\| ^19`, `react-dom` dito, `@types/react *` — React 19.2.8 ist darin |
| Größe | **81 Pakete, 16 MB** in `node_modules`, wenn nur `radix-ui` installiert wird: 60 × `@radix-ui/*`, 4 × `@floating-ui/*`, `aria-hidden`, `react-remove-scroll`, `tslib`, `get-nonce`, `detect-node-es` … Einzeln: dialog 99 KB, popover 94 KB, dropdown-menu 107 KB, tooltip 139 KB, select 352 KB (unpacked) |
| jsdom | fünf Proben: **Dialog ✓** (öffnet, Escape schließt, Fokus kehrt zum Auslöser zurück), **Popover ✓**, **DropdownMenu ✓** (⏎ öffnet `role="menu"`), **Tooltip ✗** (erscheint bei Fokus in einer Sekunde nicht), **Select ✗** (⏎ öffnet keine `listbox` — jsdom kennt `hasPointerCapture` und `scrollIntoView` nicht) |
| Wails-WebKit | **nicht gemessen** — dafür hätte Radix in die Anwendung gemusst |
| Beschriftung | bleibt Sache der Anwendung: „For form controls, the native HTML `label` element is used. For non-form controls … WAI-ARIA specifies how to provide accessible names" (radix-ui.com/primitives/docs/overview/accessibility, gelesen am 2026-09-08 über Context7) |

## Entscheidung

**Eigene Primitive, kein Radix.** Acht Bausteine und ein Helfer, alle in
`frontend/src/bausteine/`, jeder nach dem Muster der WAI-ARIA Authoring
Practices, jeder mit Test, jeder ohne Text im Quelltext:

| Baustein | Muster | Was er zusagt |
| --- | --- | --- |
| `Feld` | Beschriftung, Bedienelement, Hinweis, Befund | `label`/`htmlFor`, `aria-describedby` auf Hinweis **und** Befund, `aria-invalid`, Befund als `role="alert"` am Feld (§6) |
| `Dialog` | modal | `role="dialog"`, `aria-modal`, `aria-labelledby`, Fokus auf das erste Bedienelement, Tab zyklisch im Kasten, Escape schließt, **Fokus kehrt zum Auslöser zurück** |
| `Entscheidung` | Rückfrage an Ort und Stelle, nicht modal | `role="group"` mit der Frage als Namen, Fokus springt auf den ersten Knopf, Escape ist Abbrechen, Fokus kehrt zurück — der eine Baustein für die zehn Rückfragen |
| `Menue` | `role="menu"` | ↑↓ Home End, Escape, ⏎ und Leertaste wählen, Fokus kehrt zurück, `aria-haspopup="menu"` am Auslöser |
| `Kontextmenue` | `Menue` über Rechtsklick **und** ⇧F10 | keine Aktion, die es nur dort gibt: der Aufrufer gibt dieselben Einträge einem sichtbaren `Menue` |
| `Tooltip` | `role="tooltip"` | bei Zeiger **und** Fokus, Escape schließt, `aria-describedby` am Auslöser |
| `Arbeitsplatz` | drei Regionen nach 0076 | Navigator 240, Zentrum, Inspector 320; je Region ein `aria-label`; ⌥⌘[ und ⌥⌘] klappen ein und aus; unter 1.200 Punkten Inhaltsbreite klappt der Navigator von selbst |
| `fokus.ts` | der Helfer dahinter | `fokusZurueck` gibt den Fokus an ein Element zurück, **wenn es das noch gibt** — nach einem Löschen ist die Zeile weg, und dann bleibt er, wo er ist; `relativZu` rechnet den Zeiger in den Kasten um |

**Select bleibt nativ.** `<select>` ist auf macOS das Bedienelement, das das
System zeichnet und vorliest; §6 verlangt, Eigenheiten zu bedienen statt
nachzuahmen. Es bekommt `Feld` um sich und sonst nichts.

**Und der Inspector ist keine Komponente.** Er ist die rechte Region des
`Arbeitsplatz`, und seine Gruppen sind `Aufklappkasten` — was U5 darüber
hinaus braucht, entsteht dort, an einem echten Formular, nicht hier auf
Vorrat. Das Issue nannte ihn in einer Reihe mit den anderen; die Reihe hat
ihn nicht gebraucht.

### Wer sie benutzt — und wer noch nicht

| Baustein | Verbraucher in U2 | Verbraucher später |
| --- | --- | --- |
| `Feld` | `Pflichtfeld` — und damit zwölf Formulare | U3 stellt Setangaben, Karteneditor, Zahlenfeld und Beschriftetesfeld um |
| `Dialog` | Befehlspalette, Einführung | — |
| `Entscheidung` | Karte löschen, Paket entfernen | U3: die acht übrigen Rückfragen |
| `Menue` | die Kartenhandlungen, die Paketzeile | U5, U8, U10 |
| `Kontextmenue` | — | U5 (Navigator), U8 (Tabelle) — die Paketliste ist eine `<table>`, und um eine `<tr>` gehört kein `div` |
| `Tooltip` | — | U5, U8: die Zeichenknöpfe der Navigatoren |
| `Arbeitsplatz` | — | U5 und U8 zuerst, das ist die Abnahme von U2; danach U7 und U10 |

**Ein Baustein ohne Verbraucher ist noch keine Anwendungsfunktion** (§2) —
deshalb steht diese Tabelle hier und nicht nur der Bestand. Was in U2 keinen
hat, hat ihn in dem Werkstück, das die Seite baut, für die er da ist.

### Nachtrag vom 2026-09-10: die Zeile hat gehalten, was sie versprach — bei zweien

**Der Zweitblick über Phase 5.1 hat nachgezählt** (Befund 1): nach U5, U7, U8
und U10 hatten `Popover`, `Kontextmenue` und `Tooltip` **keinen** produktiven
Aufrufer. Die Tabelle oben stand als Zusage da, und niemand hat sie geprüft.

- **`Kontextmenue`** hat seinen bekommen: die Platzliste des Deckbauplatzes
  (`Deckplaetze.tsx`), mit denselben vier Handlungen, die auch als Knöpfe
  dastehen — die Regel des Bausteins.
- **`Tooltip`** ebenso: die Zeichenknöpfe `−` und `+` derselben Liste. Dabei
  fiel auf, dass er die WCAG-Regel, die er zitiert, **nicht** einlöste — der
  Zeiger konnte den Text nicht erreichen (1.4.13, „hoverable"). Behoben.
- **`Popover` ist gefallen.** Er hatte nach vier Werkstücken keinen Aufrufer,
  und die zwei genannten Stellen brauchten ihn nicht: der Bezug auf einen
  Druck steht in U5 als Feld, der Filter der Sammlung in U8 als Formular.
  Einen Baustein auf Vorrat zu halten ist genau das, was dieser Vermerk beim
  Inspector abgelehnt hat. Wer ihn braucht, baut ihn mit seinem Aufrufer —
  das Muster steht in der APG, und `Menue` und `Kontextmenue` zeigen, wie es
  hier aussieht.

## Verworfen

- **Radix Primitives**, aus fünf Gründen, und jeder ist gemessen:
  1. **Das eine Teil, das schwer zu bauen ist, will niemand.** `Select` ist
     das größte Paket (352 KB) und das, das unter jsdom als Erstes fällt —
     und es ersetzt ein natives Element, das hier bleiben soll.
  2. **81 Pakete für drei Muster.** Popover, Menü und Tooltip sind nach APG je
     rund 150 Zeilen; `Laschen` zeigt, dass dieses Projekt sie schreiben und
     prüfen kann. Jedes Paket ist eine Zeile in `DRITTANBIETER.md`, eine
     Pinnung und eine Fläche für `npm audit` — dieses Projekt hat `jest-axe`
     für eine Zusicherung von fünf Zeilen abgelehnt, aus demselben Grund.
  3. **Zwei von fünf Proben brauchen Polyfills.** Eine Vorrichtung, die
     `hasPointerCapture` und `scrollIntoView` nachbaut, ist großzügiger als
     der Betrieb (§3) — und genau das Muster, gegen das die Leitlinien drei
     Absätze führen.
  4. **Was Radix nicht abnimmt, ist die Hälfte.** Beschriftung, Fokusziele
     und die Frage, ob eine Rückfrage überhaupt ein Dialog sein darf (§6),
     bleiben hier — die Doku sagt es selbst.
  5. **Wails-WebKit ist nicht belegt.** Ein Lauf im Fenster wäre nötig
     gewesen, um es zu belegen — für eine Bibliothek, gegen die vier Gründe
     stehen, wäre das eine Messung ohne Entscheidung.
- **`<details>` für Popover und Inspector.** Der Grund steht schon am
  Aufklappkasten: eigener Zustand, von außen nicht zu schließen, ein Dreieck
  je Browser.
- **`<dialog>` mit `showModal()`.** Es brächte die Fokusfalle mit — aber
  `::backdrop` und die Rückgabe des Fokus verhalten sich in WKWebView anders
  als in Chromium, und der Vorschau-Browser (`vite preview`) ist Chromium:
  ein Baustein, der in der Vorschau anders ist als im Fenster, ist keine
  Vorschau. **Nicht gemessen**, deshalb nicht behauptet; hier steht nur,
  warum die Frage nicht offen bleibt: der eigene Dialog verhält sich in
  beiden gleich, weil er nichts benutzt, was die beiden verschieden tun.

## Folgen

- Jede Rückfrage, die heute aus Knöpfen besteht, wird in ihrem Werkstück auf
  `Entscheidung` gestellt; U2 stellt Karte löschen und Paket entfernen um, U3
  die übrigen.
- `Palette` und `Einfuehrung` bekommen den `Dialog` — und damit erstmals die
  Fokusrückgabe.
- `docs/KONTRASTE.md` bekommt die Farben von Tooltip und Menü, gemessen.
- Die Kürzel ⌥⌘[ und ⌥⌘] stehen in der Kürzeltafel und im Handbuch.
