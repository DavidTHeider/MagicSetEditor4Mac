# 0125 — Mantine als Bauteilbibliothek, und was sie kostet

**Stand:** angenommen · **Datum:** 2026-09-20
**Schreibt fort:** [0077](0077-eigene-primitive-statt-radix.md) — und kehrt
seine Richtung um, mit demselben Maßstab
**Berührt:** [0065](0065-wie-diese-anwendung-aussieht.md),
#1679 (UX-07)

## Warum es diesen Vermerk gibt

**Weil [0077](0077-eigene-primitive-statt-radix.md) das Gegenteil entschieden
hat** — vor zwölf Tagen, mit einer Messung, die nichts schuldig blieb: 81
Pakete, 16 MB, fünf jsdom-Proben, zwei davon rot. Eine Entscheidung, die eine
andere umkehrt, braucht denselben Maßstab; sonst ist sie eine Meinung.

**UX-07.AK-01 nennt Mantine 9.6.1 als gewählte Bibliothek**, und AK-02 sagt
dazu: *„Die Mantine-Auswahl ist beschlossen, ihre native Eignung ist noch
nachzuweisen. Ein gescheiterter Pflichtpilot blockiert die Übernahme."* Dieser
Vermerk hält fest, **was gemessen ist** — und was noch nicht.

## Was gemessen wurde

**Am 2026-09-20**, in einem eigenen Projekt mit dem gepinnten Stand dieses
Repositoriums: React 19.2.8, jsdom 30.0.1, vitest 4.1.10,
@testing-library/react 16.3.2, user-event 14.6.3.

| | Mantine 9.6.1 | Radix (0077) |
| --- | --- | --- |
| Lizenz | **MIT** | MIT |
| Peerbereich | `react ^19.2.0`, `react-dom ^19.2.0` | `^16.8 \|\| ^17 \|\| ^18 \|\| ^19` |
| Neue Pakete | **16** | 81 |
| Zuwachs `node_modules` | **26 MB** (104 → 130) | 16 MB |
| Eigene Abhängigkeiten | `@floating-ui/react`, `clsx`, `react-number-format`, `react-remove-scroll`, `type-fest` | 60 × `@radix-ui/*`, 4 × `@floating-ui/*`, … |
| jsdom-Proben | **7 von 7 bestehen** | 3 von 5 — Tooltip und Select fielen aus |

**Die sieben Proben**: Modal (öffnet, Escape schließt), Menü (`role="menu"`,
Eintrag wählbar), Hinweisfähnchen bei Fokus, Auswahlliste (`role="listbox"`,
Wahl kommt an), Popover, Laschen (Pfeiltasten wandern), Eingabefeld
(Beschriftung, Befund, `aria-invalid`).

**Der Unterschied zu Radix ist der Punkt.** Dort fielen zwei von fünf aus, und
zwar genau die, die diese Anwendung am dringendsten braucht. Hier besteht
alles — **aber nur mit drei Vorkehrungen**, und die stehen unten, weil sie zum
Ergebnis gehören.

## Entscheidung

**Mantine 9.6.1 wird die Bauteilbibliothek dieser Anwendung**, auf die Stelle
genau gepinnt (`--save-exact`), zusammen mit `@mantine/hooks` in derselben
Fassung — sein Peerbereich verlangt sie ohne Spielraum.

**Die eigenen Bauteile bleiben, wo sie halten, was sie zusagen** (§15, Satz
12: was funktioniert, wird nicht ohne Grund neu geschrieben). Ersetzt wird,
wofür UX-07.AK-01 einen Nachweis verlangt, und das zuerst dort, wo die
Oberfläche ohnehin neu entsteht.

**Der erste Pilot ist `Dialog`**, und der Grund steht in AK-01: die nativen
Pflichtprüfungen gelten ausdrücklich „Palette" und „Overlays". Ob Mantines
Overlay in WKWebView trägt, erfährt man nur an einem echten — und die
Anwendung hat genau eines.

## Die drei Vorkehrungen, ohne die nichts davon stimmt

**1. `MantineProvider env="test"`.** Ohne sie bleibt unter jsdom **jedes**
Overlay auf `display: none` stehen: es steht im Baum, und jede rollenbasierte
Abfrage übersieht es. Der Grund steht in Mantines Quelltext, nicht in einer
Erinnerung — `Transition.mjs`, Zeile 22: `if (env === "test") return mounted ?
… : …`. Der Weg dahin war lang genug, dass er hier steht: erst `withinPortal`,
dann `transitionProps`, dann eigene Maße für jsdom — nichts davon half, weil
keines die Ursache war.

**2. `window.matchMedia` muss es geben.** Sonst wirft **schon der Provider**,
und damit jeder Lauf, der die Anwendung zeichnet. jsdom bringt es nicht mit.

**3. `ResizeObserver` muss es geben.** Sonst wirft die Auswahlliste. Dieses
Repositorium hatte ihn seit #891 ohnehin.

## Vier Funde, die beim Einbau herausfielen

Sie stehen hier, weil jeder von ihnen einen Lauf rot gemacht hat — und weil
sie zusammen sagen, was eine fremde Bibliothek wirklich kostet.

**`<Modal>` reicht Fremdeigenschaften an seine Wurzel, nicht an den Kasten.**
Ein `aria-label` landete am Schleiercontainer, und der Knoten mit
`role="dialog"` stand **ohne Namen** da. Nachgelesen in `Modal.mjs`, Zeile 34,
und `ModalBaseContent.mjs`, Zeile 33. Die zusammengesetzte Form
(`Modal.Root` / `Modal.Overlay` / `Modal.Content`) löst es.

**Die Rückgabe des Fokus greift beim Schließen, nicht beim Aushängen.** Die
drei Aufrufer schreiben `{offen && <Dialog …>}`; Mantines Effekt läuft dann
nie mit `opened=false`. Die Zusage bleibt deshalb **eigen** — sieben Zeilen,
die der Baustein schon hatte.

**Mantines Escape-Horcher fragt das Ziel nach einem Merkmal.** Ein
`fireEvent.keyDown(window, …)` gibt ihm das Fenster als Ziel, und
`window.getAttribute` gibt es nicht. Der Lauf war damit **großzügiger als der
Betrieb** (§3): ein echter Tastendruck hat immer ein Element als Ziel. Er
stellt es jetzt nach.

**Der Provider legt ein `<style data-mantine-styles>` in den Baum.** Drei
Läufe sagten über `toBeEmptyDOMElement()` zu, dass ein Bauteil gar nichts
zeichnet — der Behälter ist seitdem nie leer. Sie prüfen jetzt, was sie
meinen (`frontend/src/test/leererbaum.ts`).

## Wo Mantine angeschlossen ist — und warum dort

**Im `Sprachraum`**, also in dem Bauteil, das 344 Läufe ohnehin aufspannen.
Stünde der Provider nur in `App.tsx`, zeichneten alle diese Läufe in einer
**anderen** Welt als das Fenster — und jeder Fund über ein Overlay wäre eine
Aussage über eine Welt, die es nicht gibt (§3). Der Name sagt seither weniger,
als das Bauteil tut; ihn zu ändern kostete 344 Dateien und brächte nichts,
was der Kommentar dort nicht auch sagt.

## Was Mantine hier **nicht** bekommt

**Keine Farben.** Mantines Palette ist ein Zahlenraster von 0 bis 9 je Ton;
diese Anwendung hat acht benannte Merkmale, und jedes ist **gemessen**
([`KONTRASTE.md`](../KONTRASTE.md)). Eine Abbildung darauf wäre eine
Erfindung — welches Grau ist `--mse-leise`? Die Bauteile bekommen ihre Farben
weiter über Tailwind; Mantine steuert **Verhalten** bei: Fokusfalle, Rollen,
Tastatur.

**Keinen zweiten Fokusring.** `stil.css` setzt ihn für die ganze Anwendung,
und `internal/werkzeuge/fokusring_test.go` hält ihn.

**Keinen Zierrahmen und keinen Schatten.** Vermerk
[0065](0065-wie-diese-anwendung-aussieht.md), Festlegung 3 — und der
Dialogkasten ist deshalb durchsichtig: die Aufrufer bringen ihre Kante mit.

**Und seine Merkmalsdatei steht ganz oben in `stil.css`**, damit die
Gestaltungssprache dieser Anwendung das letzte Wort hat. Andersherum
überschriebe Mantine Farben, die gemessen sind — genau der CSS-Konflikt, den
UX-07.AK-04 fernhalten will.

## Was das kostet

**Eine fremde Bibliothek unter der Zugänglichkeit.** Fokusfalle und
Tastaturführung liegen jetzt nicht mehr im eigenen Quelltext. Dass sie tragen,
prüft weiterhin `dialog.test.tsx` — derselbe Lauf, derselbe Anspruch. Was ein
Lauf nicht prüft, ist VoiceOver in WKWebView, und das ist die offene Hälfte.

**26 MB und 16 Pakete.** Sie gehen nicht in die Anwendung, sondern in
`node_modules`; was im Bündel landet, entscheidet der Bau.

**Gemessen am 2026-09-25** (UX-07.AK-01, „Paket-/Pflegeaufwand"), gebaut mit
dem Vite dieses Projekts aus genau dem, was die Anwendung einbindet —
`Drawer`, `Modal`, `MantineProvider`, `createTheme`, React außen vor:

| Teil | minifiziert | gzip |
| --- | ---: | ---: |
| Skript | 144 KB | 34,5 KB |
| `@mantine/core/styles.css` | 277 KB | 39 KB |

**Die Stildatei ist der größere Teil**, und das ist ein Befund: `stil.css`
bindet sie **ganz** ein, für drei Bauteile. Mantine liefert sie auch je
Bauteil; der Tausch ist klein, aber er ist eine Änderung an der
Kaskade, und die gehört in den Migrationsplan (unten) und nicht nebenbei
hierher. **Die Pflege**: zwei Pakete, auf die Stelle gepinnt
(`--save-exact`), `@mantine/hooks` mit derselben Zahl; ein Heben ist ein
eigener Zug mit den Piloten als Probe.

**Und zwei Bauteile weniger Eigentum.** `Dialog` ist das erste; was folgt,
entscheidet die Fläche, an der es gebraucht wird.

## Die drei Pflichtpiloten — gefahren in WKWebView

**Nachgetragen am 2026-09-25.** Hier stand, keiner der drei Piloten aus
UX-07.AK-01 sei gefahren und WKWebView ungemessen. Beides gilt nicht mehr:
die WebKit-Probe (#1828)
fährt die Anwendung in einem WKWebView, eingerichtet wie in Wails, mit
echten AppKit-Ereignissen, und liest WebCores Zugänglichkeitsbaum.

| Pilot | Beleg | Was er zeigt |
| --- | --- | --- |
| Kartenformular | [`ux07-kartenformular`](../belege/webkitprobe/ux07-kartenformular.md) | Tastatur durch die Felder, Fokus sichtbar |
| Einstellungszentrum | [`ux07-einstellungen`](../belege/webkitprobe/ux07-einstellungen.md) | Suche, Bereiche, Schalter mit der Tastatur |
| Handelsassistent | [`ux07-handelsassistent`](../belege/webkitprobe/ux07-handelsassistent.md) | Laschen mit den Pfeiltasten, ein Gegenüber aufschlagen, jedes Ziel mit Tab, beide Erscheinungsbilder |
| Palette und Overlays | [`ux07-palette`](../belege/webkitprobe/ux07-palette.md), [`ux08-schubladen`](../belege/webkitprobe/ux08-schubladen.md) | Fokusfalle, Escape, Rückkehr zum Auslöser |

**Der erste Pilot war sofort rot**, und das ist der Grund, warum sie Pflicht
sind: ohne die macOS-Tastaturnavigation erreichte Tab in WKWebView keinen
Knopf, und aus dem Mantine-Dialog fiel der Fokus beim ersten Tab hinaus.
Behoben über `TabFocusesLinks` in `internal/fenster`; danach grün.

**Was die Probe nicht ersetzt**: VoiceOvers eigene Satzbildung und Betonung.
Sie liest den Baum, aus dem VoiceOver liest — `aria-describedby` etwa sieht
WebKit 1 nicht. Die Stichprobe mit VoiceOver selbst steht in
[`HANDPRUEFUNGEN.md`](../HANDPRUEFUNGEN.md).

## Eigene Bauteile neben Mantine — UX-07.AK-03

**Die Regel**: Mantine liefert **Verhalten** — Fokusfalle, Rollen,
Überlagerung, Tastaturführung. Ein eigenes Bauteil entsteht nur dort, wo es
**Fachsemantik** kapselt oder ein wiederkehrendes Muster dieser Anwendung ist,
und es nimmt Mantine dann **hinein**, statt es nachzubauen.

| Eigenes Bauteil | Was es kapselt | Mantine darin |
| --- | --- | --- |
| `Dialog` | Rückgabe des Fokus beim Aushängen, Name am Kasten | `Modal` (zusammengesetzt) |
| `Schublade` | Eigenschaften und Werkzeuge als Schubladen, Escape von innen nach außen | `Drawer` |
| `Pflichtfeld`, `Zahlenfeld`, `Vorschlagsfeld`, `Waehrungsfeld` | Befund am Feld, Rohtext bleibt stehen (UX-06) | — (native Eingaben) |
| `Laschen`, `Seitenkopf`, `Knopf` | die Gestaltungssprache aus [0065](0065-wie-diese-anwendung-aussieht.md) | — |
| `Suchauswahl` (mit `Zeitzonenwahl`, `Stilauswahl`, `Auswahlfeld`) | die Auswahl mit Suche ab fünfzehn Werten (R7), UXR-F1 | `Combobox` — **die Fläche**; Tastatur und Rollen eigen (unten) |

**Keine zweite Komponentenbibliothek.** `frontend/package.json` führt als
Oberflächenbibliothek genau eine, und jede neue Abhängigkeit braucht ihre
Zeile in `docs/ABHAENGIGKEITEN.md` samt Wächter. Ant Design, MUI oder Chakra
kämen dort vorbei, nicht still.

## Die Auswahl mit Suche — Schritt 2 des Migrationsplans, UXR-F1

**Gebaut am 2026-10-03** (#2007):
`frontend/src/bausteine/Suchauswahl.tsx`, das Kombinationsfeld nach dem ARIA-Muster
— Eingabe mit `role="combobox"`, Liste mit `listbox` und `option`, der Fokus
bleibt im Feld, die hervorgehobene Zeile steht in `aria-activedescendant`. Sie
löst die Notbehelfe ab, die R7 aus `docs/UIREVIEW.md` hinterlassen hatte: ein
Suchfeld über einem nativen `select` (Stil), ein natives `select` ohne Suche
(Testkarte, Spiel) und einen Freitext (Zeitzone).

**Was Mantine liefert und was eigen ist** — wieder nachgelesen und nicht
erinnert (Mantine 9.6.1):

- **Mantine: die Fläche.** `Combobox` mit `Combobox.Target`, `.Dropdown`,
  `.Options`, `.Option` und `.Empty`: das Portal, die Lage am Feld mit `flip`
  und `shift`, die Breite des Feldes, das Schließen auf einen Klick daneben.
  Dazu `unstyled`, wie beim `Popover`: Farben und Kante kommen aus den
  gemessenen Merkmalen ([`KONTRASTE.md`](../KONTRASTE.md)), nicht aus Mantines
  Palette.
- **Eigen: Tastatur und Rollen.** `use-combobox-target-props.mjs` merkt die
  Kennung der hervorgehobenen Zeile in einem Zustand, den nur ein Pfeil
  fortschreibt; die Zeile selbst führt es als Attribut im Baum. Filtert die
  Suche die Liste, zeigt `aria-activedescendant` auf eine Zeile, die es nicht
  mehr gibt. Dazu hält Mantine die Liste standardmäßig verborgen im Baum
  (`keepMounted`, `display: none`) — bei 420 Zeitzonen 420 Knoten je Feld.
  Hier ist die Stelle eine Zahl im Baustein, Attribut und Rolle folgen ihr, und
  die Liste steht nur da, solange sie offen ist.
- **Escape wird nur bei offener Liste verbraucht** (`preventDefault` und
  `stopPropagation`): die `Schublade` fragt nach `defaultPrevented` und schließt
  sonst — ein Feld darin ließe sich mit seiner offenen Liste nicht schließen,
  ohne dass dieselbe Taste die Schublade mitnähme.

**Wo sie steht**, in zwei Gestalten:

- **Immer** eine Auswahl mit Suche: Stil (Eigenschaften und neues Set),
  Testkarte der Vorlagenwerkstatt, Spiel beim Anlegen eines Pakets, Zeitzone.
  Die Zeitzone ist die einzige, deren Werte Go nicht aufzählt
  (`PruefeZeitzone` prüft gegen die Zonendatenbank): die Liste steht im
  Fenster (`sprache/zeitzonen.json`, aus `Intl.supportedValuesOf` und `UTC`),
  und `internal/werkzeuge/zeitzonenliste_test.go` hält, dass Go **jede** Zeile
  annimmt.
- **Ab `suchschwelle` (15) Werten**, darunter nativ — R7 wörtlich: das
  [`Auswahlfeld`](../../frontend/src/bausteine/Auswahlfeld.tsx) entscheidet an
  **einer** Stelle, und die Stellen, deren Listen mit dem Bestand wachsen,
  fragen die Schwelle nicht selbst: Feld und Set der Suche, Deck am Sitz,
  Deck der Leihe, Vorlage, Format und Deck des Turniers. Dieselbe Schwelle
  gilt in `Feldeingabe` (die Wahlen eines `choice:`-Feldes — `watermark` von
  `magic` hat 229), in der `Statistikwahl` (mit ihren zwei Gruppen als Zusatz)
  und in der `Kontaktwahl` (Optionsfelder bis vierzehn Gegenüber, darüber
  eine Auswahl, in die man den Namen tippt).

**Die Liste zeichnet höchstens `zeilengrenze` (200) Zeilen** und sagt unter
der Liste und in der Live-Region, wie viele Treffer es sind und wie viele
dastehen. Gesucht wird in allen. Der Grund ist der Knoten je Zeile: ein Set
trägt tausende Karten. Die Zeitzone (420) und das Wasserzeichen (229) liegen
darüber und zeigen deshalb zuerst die ersten zweihundert.

**Die Live-Region steht am Ende von `body`**, nicht im Baum des Feldes: viele
Stellen umschließen ihr Bedienelement mit dem `label`, dessen Text der Name des
Feldes ist **und alles, was in ihm steht** — auch ein Text, den man nicht sieht.
Dort stand sie zuerst, und aus „Zeitzone" wurde bei offener Liste „Zeitzone
Treffer: 420". Sie trägt `aria-live` ohne Rolle `status`: ein leeres `status`
auf jeder Seite träfe jede Abfrage nach einer Meldung der Seite.

**Was offen bleibt**: VoiceOver selbst (`HANDPRUEFUNGEN.md`), die Stellen,
deren Listen mit dem Bestand wachsen und die kein UIREVIEW-Eintrag nennt (Sets
und Decks der Deckausstattung, Exemplare der Zuordnung, Decks und Profile der
Übernahme, Formate der Deckwerkbank), und ein Wächter über *alle* nativen
`select` mit mehr als fünfzehn Werten — den gibt es nicht; die Zählung ist
Durchsicht je Issue.

## Migrationsplan — UX-07.AK-04

**Kein gleichzeitiger Umbau des ganzen Bestands** (§5, drei Züge; §15,
Satz 11). Ein eigenes Bauteil wird auf Mantine umgestellt, **wenn** seine
Fläche ohnehin angefasst wird und ein Pilot dafür grün ist — nicht vorher.

1. **Verhalten zuerst**: Overlays (Dialog, Schublade) — erledigt. Dazu
   gehört als nächster kleiner Schritt, die Stildatei **je Bauteil**
   einzubinden statt ganz (siehe „Was das kostet").
2. **Dann Auswahl**: Kombinationsfelder mit Vorschlägen, sobald `datalist`
   in WKWebView an eine Grenze stößt (UX-06 hat es gemessen: es trägt).
   **Der Anlass war R7** — die Auswahl mit Suche ab fünfzehn Werten
   (#2007).
   Gebaut als `Suchauswahl`, siehe oben.
3. **Formulare zuletzt**: die eigenen Felder tragen Fachbefunde; ein Tausch
   bringt dort nur, was sie schon haben.

**Gegen CSS-Konflikte** gilt, was oben steht: Mantines Merkmalsdatei steht
**ganz oben** in `stil.css`, Farben kommen weiter aus den gemessenen
Merkmalen ([`KONTRASTE.md`](../KONTRASTE.md)), und der Fokusring gehört
`stil.css` (`fokusring_test.go`). **Spezialbausteine bleiben**: der Tisch
zeichnet mit PixiJS, die Kartenliste ist eigene Tabelle, der Skripteditor
eigenes Feld — keiner davon bekommt ein Mantine-Gegenstück, solange er hält.

## Was ausdrücklich **nicht** belegt ist

**VoiceOver selbst.** Siehe oben — der Baum ist belegt, die Stimme nicht.

## Woran man merkt, dass die Entscheidung falsch war

- Ein Pflichtpilot fällt an VoiceOver oder an WKWebView aus. Dann blockiert
  er die Übernahme, und zwar nach AK-02 ausdrücklich.
- Die eigenen Farben und Mantines Merkmale geraten sich ins Gehege, und
  `KONTRASTE.md` hält nicht mehr, was es misst.
- Eine Fassung bricht die Fokusfalle, und niemand merkt es — dann war der
  Tausch „eigener Code gegen fremden" schlecht, und `dialog.test.tsx` hat
  nicht genug gehalten.

## Fortgeschrieben am 2026-10-09: die Grundmerkmale doch (A03)

**Der zweite Punkt der Liste darüber ist eingetreten.** Der Audit zu 0.8.0
sah in einem Gast „Erste Schritte" mit schwarzer Überschrift auf dunkler
Fläche (F-03, #2217).
Die Ursache stand hier schon als Grenze: Mantine bekam **keine** Farben —
auch nicht die Textfarbe, mit der es `body` färbt. Ohne `defaultColorScheme`
rechnet es „light", also `#000`, und jede Überlagerung, die es an `body`
hängt, erbte sie, wo kein Bauteil eine eigene Farbklasse setzte. Gemessen am
Fenster (Hüllenprobe `a03`): **1,27 : 1** an jeder Zeile der Befehlspalette
und an den Überschriften der ersten Schritte — im dunklen Erscheinungsbild,
auf jedem macOS. `KONTRASTE.md` hielt es nicht, weil es Paare misst und die
geerbte Farbe keines war.

**Jetzt bekommt Mantine seine Grundmerkmale** — Text, Grund, gedämpft, Rand,
Platzhalter, die Vorgabe der Knöpfe — aus den gemessenen Merkmalen dieser
Anwendung (`stil.css`, unter dem Import). **Die Palette von 0 bis 9 bleibt
unberührt**, und der Satz „welches Grau ist `--mse-leise`?" gilt weiter: für
die Grundmerkmale gibt es eine Entsprechung, für das Raster nicht. Danach
hielt jede der drei Überlagerungsarten in beiden Erscheinungsbildern
mindestens 6,99 : 1.

## Nachtrag vom 2026-10-16: Mantine in seiner Ebene

**Anlass:** #2451,
gefunden in AP-3. Mantine stand **ohne Ebene** in `stil.css`, mit der
Begründung, wer zuerst geladen wird, verliert bei gleicher Spezifität. Unter
Tailwind 4 trägt das nicht: Preflight und Utilities liegen in Ebenen, und
eine Regel ohne Ebene gewinnt gegen jede geschichtete. Mantines
`input, button, textarea, select { font: inherit }` schlug jedes
`font-medium` und `text-sm` an Knöpfen und Feldern — am Fenster gemessen,
400 statt 500 und 16 statt 14 Punkte.

**Jetzt** kommt Mantine aus `@mantine/core/styles.layer.css` (alles in
`@layer mantine`), und `stil.css` erklärt vor Tailwind die Reihenfolge
`theme, base, mantine, components, utilities` — Mantine über Preflight, die
Utilities über Mantine, wie Mantines eigene Anleitung es zeigt (mantine.dev,
*Mantine styles*, *CSS layers*, abgerufen am 2026-10-16). Die Gestalt der
vier Bauteile, die Mantine hier zeichnet, ist vorher und nachher am Fenster
verglichen (`mantineebene`); geändert hat sich nur die Schrift darin. Beleg:
`audits/0.9.0/umsetzung/MANTINE-2451.md` (Entwicklungsarchiv).

