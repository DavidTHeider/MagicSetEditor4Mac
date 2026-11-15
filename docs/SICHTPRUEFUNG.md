# Sichtprüfung

**Verlegt aus `AGENTS.md` am 2026-09-22** — ENG-01.AK-04
(#1726):
„Detailkonzepte wandern in passende Dokumente statt beide Regeldateien
weiter mit ganzen Projektplänen zu überladen." **Die Regel steht weiter
dort** — hier steht, warum sie so lautet, samt jeder Messung.

---

Was weder ein DOM-Test noch axe zeigt — Abstände, Ausrichtung, **Farbkontraste**,
helles und dunkles Erscheinungsbild, und bei diesem Projekt zusätzlich die
**Kartenausgabe** —, wird angesehen.

**Und dafür gibt es zwei Wege, die nicht dasselbe können.**

| | zeigt | zeigt **nicht** |
| --- | --- | --- |
| `npm --prefix frontend run vorschau` | Aufbau, Abstände, Farben, beide Erscheinungsbilder | **jedes Datum und jede Karte** |
| `make dev-browser`, dann `http://localhost:34115` im Browser | alles davon **und** echte Sets, Karten, Vorschaubilder | Menüleiste und Fensterrahmen |
| **`scripts/webkitprobe.sh`** gegen dieselbe Adresse (#1828) | dieselbe Oberfläche in **WebKit** statt in Chromium — einem WKWebView, eingerichtet wie in Wails —, mit **echten** Tastatur- und Mausereignissen, WebKits eigenen Vorschlagslisten, WebGL und dem **Zugänglichkeitsbaum**, den VoiceOver liest | Menüleiste, Fensterrahmen, native Dateidialoge, Ablegen aus dem Finder — und VoiceOvers eigene Satzbildung |

**Die dritte Zeile ist die, die einen Menschen ersetzt, wo es geht.** Das
Browserfeld dieser Arbeitsumgebung ist Chromium; die Anwendung läuft in
WebKit, und genau dort liegen die Unterschiede, die bisher „nur am nativen
Fenster" hießen: ein Klick fokussiert in WebKit keinen Knopf, eine
`datalist` öffnet ein **natives** Fenster und kein Menü der Seite, und was
ein Vorleser sagt, kommt aus WebCore und nicht aus Blink. Die Probe braucht
**keine Freigabe**: sie schickt Ereignisse nur an ihr eigenes Fenster, wird
nie aktiv und liegt außerhalb des Bildschirms. Wie sie das schafft, steht in
den Köpfen von `scripts/webkitprobe/*.swift`.

**`make webkitabnahme` fährt alle Abnahmen auf einmal** — gegen die Anwendung
im Servermodus mit **eingebetteter** Oberfläche, über einem frischen
Prüfbestand, und legt je Drehbuch den Bericht nach
[`docs/belege/webkitprobe/`](belege/webkitprobe/uebersicht.md). **Was sie am
ersten Tag fand**, bei grüner Pipeline: Tab übersprang jeden Knopf, das
Suchfeld der Sammlung ragte in die Liste, die Messseite des Tischs warf in
jedem Browser, und auf einem Retina-Bildschirm lag der Tisch im linken oberen
Viertel seiner Fläche. Keiner der vier war in jsdom oder Chromium zu sehen.

Hier stand bis zum 2026-08-23 nur der erste, und der Satz darüber verlangte
ausdrücklich die **Kartenausgabe** — die er nicht zeigen kann. Gemessen an
diesem Tag: unter `vite preview` sind `window.go` und `window.runtime`
**`undefined`**, und die Fassade lehnt jeden Aufruf mit seinem Namen ab („Die
Methode Sets ist nicht belegt."). Der Aufbau steht, die Seite ist leer.

**`make dev-browser` stellt dieselbe Anwendung samt Bindungen unter `34115`
bereit**; wer dort einen Browser hinschickt, sieht dieselbe Anwendung mit
denselben Daten — und kann sie ablichten, was am Fenster ohne
Bildschirmaufnahme-Berechtigung nicht geht. Die Bildschirmabzüge im
`README.md` sind so entstanden.

**Seit Wails 3 sind das zwei Programme** (ENG-02,
[0132](entscheidungen/0132-wails-3-als-huelle.md)). Unter Wails 2 öffnete
`wails dev` das Fenster und diente zugleich dem Browser; Wails 3 baut ein
Programm mit Fenster **oder** mit HTTP-Server (Bauschalter `server`). Der
Browserweg ist deshalb ein eigener Aufruf — dieselbe Adresse, dieselben
Bindungen, dieselbe Oberfläche von Vite, aber **ohne** Menüleiste, Dialoge
und Zwischenablage: die gibt es nur am Fenster (`make dev`), und dort nimmt
sie ein Mensch ab (`docs/WAILS3.md`).

**Und wer dabei etwas ablegt, legt es auf diesem Rechner ab.** Ein Lauf, der
eine Peeridentität anlegt, schreibt sie in den **Anmeldebund** — und den gibt
man nicht weiter (§11). Dafür gibt es `wails-sichtpruefung` in
`.claude/launch.json`: eigener Bestand, eigener Schlüsselbund, beides unter
`$TMPDIR`. Die gewöhnliche Startkonfiguration `wails` bleibt daneben; sie ist
die richtige, solange nichts abgelegt wird.

> **Diese Zusage hat bis zum 2026-09-14 zur Hälfte gehalten, und das ist
> teuer geworden.** Der Bund griff, der Bestand nicht: `main.go` rief
> `app.Datenbankpfad()` unmittelbar, und nur die drei Flächen ohne Fenster
> gingen über `app.BestandsortFuerFlaechen`. **Jede Sichtprüfung schrieb
> damit in den echten Bestand dieses Rechners** — aufgefallen ist es, als
> eine Sichtprüfung zu N2 die Peeridentität des echten Profils abgelöst hat.
>
> Eine Zusage, die zur Hälfte hält, ist schlimmer als keine (§10): sie wird
> geglaubt. Gehalten wird sie jetzt von
> `internal/werkzeuge/bestandswurzeln_test.go` — **jeder** Einstieg, der
> einen Bestand öffnet, muss den Umgebungsschalter fragen, und wer
> `Datenbankpfad()` unmittelbar ruft, umgeht ihn.
>
> **Und danach hielt sie bis zum 2026-09-23 nicht für die Bilder.** Bestand
> und Protokoll folgten dem Schalter, die eingesetzten Bilder nicht — eine
> Sichtprüfung, die ein Bild einsetzte, legte es in das echte Profil.
> Gefunden bei QA-10.NAT-01
> (#1741);
> seitdem folgen sie dem Bestand, und derselbe Wächter lässt den Vorgabeort
> nur an benannten Stellen zu.
>
> **Pakete bleiben beim Rechner, und das ist gewollt**: sie gehören ihm
> (`docs/AUTORISIERUNG.md`), und eine Sichtprüfung braucht die echten
> Vorlagen. Wer darin ein Paket bezieht oder ändert, ändert es für diesen
> Rechner.

**`wails dev` fasste `go.mod` an — und das ist mit ihm gegangen.** Es fuhr
`go mod tidy` mit, und was dabei herausfiel, fiel **still** heraus: am
2026-09-14 waren es zehn indirekte Einträge. `make dev` baut mit `go build`,
und das räumt `go.mod` nicht auf (ENG-02). Die Regel dahinter gilt weiter für jedes
Werkzeug, das es anfasst: **ansehen, nicht ungesehen übernehmen und nicht
ungesehen zurücknehmen.**

Kontraste werden am gerenderten Element nachgemessen, nicht geschätzt. WCAG
1.4.11 gilt auch für Ränder ohne Fläche, und dafür hat axe keine Regel.

**Das Ergebnis steht in `docs/KONTRASTE.md`**, samt der Art, wie gemessen wurde.
Eine Messung von Hand veraltet still — deshalb hält
`internal/werkzeuge/kontraste_test.go` die Datei gegen den Quelltext: eine
benutzte Vordergrundfarbe ohne Zeile dort macht die Pipeline rot. Was er nicht
prüfen kann, steht neben ihm: ob die Zahl stimmt und ob der genannte Grund der
richtige ist.

**Gemessen wird über einen Canvas, nicht über ein Textmuster.** Tailwind 4 gibt
`oklch()`, und eine Regex, die Ziffern aus einer Farbangabe zieht, liest daraus
Unsinn — beim ersten Anlauf kam für schwarze Schrift auf Weiß 2,1:1 heraus. Wer
die Farbe in eine 1×1-Fläche malt und zurückliest, bekommt sRGB, gleich wie sie
geschrieben stand.

**Trennlinien halten 3:1 in beiden Erscheinungsbildern.** Sie sind bei jedem
bisherigen Anlauf zu blass gewesen: `slate-200/700` ergibt 1,7:1 im dunklen.
