# Die Durchsicht der Flächentexte

**UX-15**
(#1771),
AK-01: *„Jede registrierte Seite ist durchgegangen; das Ergebnis steht je
Seite als geprüft, nichts zu tun oder als Liste der geänderten Stellen. Eine
Seite ohne Eintrag gilt als ungeprüft."*

**Ein Wächter hält diese Datei vollständig**
([`internal/werkzeuge/flaechentexte_test.go`](../internal/werkzeuge/flaechentexte_test.go)):
jeder Seitenschlüssel aus `navigation.ts` steht hier, und jede Zeile meint
eine Seite, die es gibt.

---

## Was gemessen wurde

**Am 2026-09-22**, über die Sprachkataloge unter `frontend/src/sprache/de/`:

| | Zeichen |
| --- | --- |
| Kopferklärung einer Seite | bis **340**, ihr erster Satz bis **114** |
| Hinweis unter einer Einstellung | im Mittel **171**, bis **371**; erster Satz im Mittel **25** |
| Abschnittserläuterung | **110** Stück über den Baum verteilt, bis **478** |

**Daraus folgt der Schnitt**: der erste Satz bleibt sichtbar, der Rest geht
hinter ein Hilfesymbol. Die Grenze für den sichtbaren Satz ist **140**
Zeichen und steht in
[`frontend/src/sprache/hinweislaenge.test.ts`](../frontend/src/sprache/hinweislaenge.test.ts).

## Drei Stellen, und keine vierte

| Bauteil | wofür |
| --- | --- |
| [`Seitenkopf`](../frontend/src/bausteine/Seitenkopf.tsx) | die Erklärung unter dem Titel — **jede** Seite |
| [`Wahl`](../frontend/src/seiten/Wahl.tsx) | der Hinweis unter einer Einstellung |
| [`Erklaerungsabsatz`](../frontend/src/bausteine/Erklaerungsabsatz.tsx) | jede Abschnittserläuterung dazwischen |

**Deshalb steht unten so oft „nichts Eigenes":** die drei Bauteile wirken für
alle Seiten zugleich. Was eine Seite eigens brauchte, steht bei ihr.

---

## Je Seite

| Seite | Ergebnis der Durchsicht am 2026-09-22 |
| --- | --- |
| `start` | geprüft — Kopf über `Seitenkopf`, nichts Eigenes |
| `sets` | geprüft — dazu `umschlagErklaerung` und `scryfallsatzErklaerung` |
| `pakete` | geprüft — dazu `bezugLizenz` und `quellenErklaerung` |
| `vorlagen` | geprüft — dazu die Erläuterungen des Inspectors |
| `sammlung` | geprüft — dazu die Erläuterungen der Sicht und der Freigaben |
| `decks` | geprüft — dazu `kurveHinweis`, `uebernahmeErklaerung`, `ungepruefterklaerung`, `einfuhrErklaerung` und die drei Sätze der Deckleihe |
| `turniere` | geprüft — dazu die Erläuterung der Paarung |
| `spielen` | geprüft — dazu `sitzprofilErklaerung` und die drei Sätze der Protokollfreigabe |
| `handeln` | geprüft — dazu `ausgangErklaerung`, `marktErklaerung` und die Erläuterungen der Buchung |
| `hilfe` | **geprüft, nichts zu tun** — hier **ist** der lange Text der Inhalt. Eine Hilfeseite hinter ein Hilfesymbol zu legen wäre die Umkehrung ihres Zwecks |
| `einstellungen` | geprüft — jeder Hinweis über `Wahl`, dazu `anmeldungHinweis`, `anmeldungGeheimnis`, `fragenHinweis` und der Dienstschalter. **Und der Proxy stand zweimal da** (AK-04) |
| `nutzer` | geprüft — zwölf Abschnittserläuterungen, die meisten dieser Seite. `anmeldungErklaerung` ist dabei **umgestellt** worden: der Weg („trägt die `client_id` unter … ein") stand im zweiten Satz und wäre hinter dem Symbol verschwunden |
| `karten` | geprüft — dazu die Erläuterungen der Marktangebote und der Marktlinke |
| `bilder` | geprüft — **eine Ausnahme**: `konfliktHilfe` bleibt ganz sichtbar. Der Absatz trägt eine `id` und ist über `aria-describedby` die **Beschreibung eines Feldes**; ihn zu teilen hieße, die halbe Feldbeschreibung zu verstecken (AK-02, letzter Satz) |
| `suche` | geprüft — Kopf über `Seitenkopf`, nichts Eigenes |
| `ersetzen` | geprüft — Kopf über `Seitenkopf`, nichts Eigenes |
| `stil` | geprüft — Kopf über `Seitenkopf`, nichts Eigenes |
| `symbol` | geprüft — dazu die Erläuterung der Symmetriebedienung |
| `setangaben` | geprüft — Kopf über `Seitenkopf`, nichts Eigenes |
| `schluesselwoerter` | geprüft — Kopf über `Seitenkopf`, nichts Eigenes |
| `statistik` | geprüft — Kopf über `Seitenkopf`, nichts Eigenes |
| `pack` | geprüft — Kopf über `Seitenkopf`, nichts Eigenes |
| `ausgeben` | geprüft — dazu die Erläuterungen der Bildausgabe |
| `druckbogen` | geprüft — Kopf über `Seitenkopf`, nichts Eigenes |
| `konsole` | geprüft — **eine Ausnahme**: `reihungSatz` bleibt ganz. Er wird aus Platzhaltern zusammengesetzt (`einsetzen`) und ist eine **Aussage über den gerade offenen Stand**, keine Erläuterung |

---

## Am laufenden Fenster — Sichtprüfung vom 2026-09-24

**Gemessen in WebKit** (WKWebView, 1.440 × 900) mit
[`ux15-seitenlaengen.probe`](../scripts/webkitprobe/abnahme/ux15-seitenlaengen.probe):
jede der 25 Seiten über die Befehlspalette geöffnet, die Seiten des Sets
einmal ohne und dann mit offenem Set, je Seite die Höhe, der Anteil im ersten
Blick, die langen und die doppelten Absätze und ein Abzug. Der Beleg liegt
unter `docs/belege/webkitprobe/`.

**Vorher** heißt: am selben Tag gegen den echten Bestand mit dem Set
„Green“; **nachher** gegen den Prüfbestand der Abnahme. Die **536
Vorlagenpakete** sind in beiden dieselben — sie trugen die zwei längsten
Seiten.

| Seite | vorher | nachher | was sich geändert hat |
| --- | --- | --- | --- |
| `start` | 24.827 | 1.169 | Unter „Was Aufmerksamkeit braucht“ standen **alle** Hinweise da, mit dem echten Bestand 536. Jetzt die ersten fünf, ihre Zahl neben dem Titel wie im Artboard *Main*, und ein Knopf „Alle … zeigen“ |
| `vorlagen` | 60.317 | 900 | Der Baum trug die Seite: wer unten eine Datei wählte, sah sie vierzig Bildschirme weiter oben aufgehen. Jetzt hat **jede Spalte des Arbeitsplatzes die Höhe bis an den Fensterrand und rollt für sich** — wie `.arbeitsplatz { height: 100% }` im Entwurf. Und derselbe Satz „Wähle links im Baum …“ stand in Zentrum **und** Inspector; der Inspector sagt jetzt, was er zeigen wird |
| `decks` | — | 900 | Derselbe Doppelsatz „Kein Deck aufgeschlagen …“ in Zentrum und Inspector — dieselbe Klasse, dieselbe Lösung |
| `sammlung` | 2.203 | 900 | Die Spalten rollen für sich; die Liste nimmt die Höhe des Zentrums statt einer zweiten Rechnung über `100vh`, die zwei Rollbalken übereinander gab |
| `karten` | — | 1.055 | Ohne offenes Set sagten die Setseiten „Öffne unter ‚Sets‘ eine .mse-set-Datei“ — auch über einem Bestand mit Sets. Jetzt „Wähle unter ‚Sets‘ ein Set, lege eines an oder lies eines ein“, und ein Knopf führt hin. **Und die Kartenliste fehlte bei 1.440 × 900**: der Arbeitsplatz maß 1.045 Punkte, unter der Grenze von 1.072, und der Navigator klappte von selbst ein. Rand und Innenabstand des Kastens um die Setlaschen nahmen 50 Punkte; im Artboard *Karten* steht der Arbeitsplatz ohne Kasten unter dem Kopf. Jetzt 1.095, alle drei Spalten stehen |
| `sets` | — | 900 | nichts zu tun |
| `nutzer` | — | 1.321 | Drei Zeilen Vorbehalt über den Laschen, an keinem Bauteil mit Grenze vorbei, und ihr Verweis nannte »Bestand« — eine Seite, die seit UX-04 eine Gruppe der Einstellungen ist. Jetzt ein Satz und das Hilfesymbol, der Weg „Einstellungen › Daten und Sicherung“ |
| `pakete` | 2.476 | 2.468 | „Ein Umschlag trägt …“ stand mit 321 Zeichen ganz da. Jetzt steht **im ersten Satz, wie der Name aussehen muss** — er ist nötig, um das Feld darunter zu füllen —, der Rest hinter dem Symbol |
| `handeln` | — | 900 | „Gehandelt wird nur …“ ganz sichtbar, und zwei Verweise auf eine Seite »Technik«, die es als Seite nicht mehr gibt. Jetzt „Einstellungen › Verbindungen“ |
| `einstellungen` | 2.149 | 2.161 | Die Belegung der ⌥-Kürzel (199 Zeichen) steht hinter dem Symbol. **Doppelt** bleibt „Wirksam ist, was der Rechner vorgibt …“ — das ist die Herkunft **je Einstellung** (AK-04), und zwei Einstellungen mit derselben Herkunft sagen dasselbe mit Recht |
| `ausgeben` | — | 1.227 | Der Satz zum Sichern gekürzt (164 → 108 Zeichen), ohne Angabe zu verlieren |
| `hilfe` | 3.077 | 3.109 | **Nichts gekürzt**, zwei Sätze berichtigt: „die dreizehn Setlaschen“ — es sind sechs Laschen und zwei Schubladen —, und die Karten stehen nicht „links die Liste, rechts die Karte“, sondern in drei Spalten |
| `spielen` | — | 1.974 | nichts zu tun: der Tisch hat die Höhe des Fensters, und darunter stehen die Partien |
| `turniere` | — | 900 | nichts zu tun |
| `bilder` | — | 914 | nichts zu tun; `konfliktHilfe` bleibt ganz — die Beschreibung eines Feldes (oben) |
| `suche` | — | 900 | nichts zu tun |
| `ersetzen` | — | 900 | nichts zu tun |
| `stil` | — | 1.055 | nichts zu tun — in der Schublade |
| `setangaben` | — | 1.055 | nichts zu tun — in der Schublade |
| `symbol` | — | 1.055 | nichts zu tun — in der Schublade |
| `schluesselwoerter` | — | 1.055 | nichts zu tun — in der Schublade |
| `statistik` | — | 1.055 | nichts zu tun — in der Schublade |
| `pack` | — | 1.055 | nichts zu tun — in der Schublade |
| `konsole` | — | 1.055 | nichts zu tun — in der Schublade |
| `druckbogen` | — | 900 | nichts zu tun; `anschnittHilfe` bleibt ganz — die Beschreibung eines Feldes |

**Und außerhalb der Seiten:** die Meldung eines unerwarteten Fehlers wies auf
»Hilfe → Diagnose«, die Diagnose steht unter „Einstellungen › Diagnose“
(Befund F14 der nativen Prüfung); die leere Recherche auf „Technik“. Beide
berichtigt.

### Was der Wächter jetzt hält, und was vorher durchging

**Die Grenze von 140 Zeichen galt für zwei Sorten Text** — Kopferklärungen
und Hinweise unter einer Einstellung. Eine Abschnittserläuterung ging durch
`Erklaerungsabsatz` und stand unter keinem Lauf: **zwölf** deutsche und
**sieben** englische Texte zeigten einen ersten Satz über der Grenze, bis
**290** Zeichen.
[`hinweislaenge.test.ts`](../frontend/src/sprache/hinweislaenge.test.ts)
misst jetzt **jeden Text im Katalog** außer der Hilfe, mit einer Ausnahme
samt Grund (ein Stück Konfiguration zum Abschreiben, kein Satz) — und eine
Ausnahme, die nicht mehr nötig ist, macht den Lauf rot.

**Was ein Katalog nicht zeigt**, ob ein langer Text ganz dasteht, hält die
Probe: auf jeder Seite außer der Hilfe darf **kein sichtbarer Absatz über 140
Zeichen** stehen, es sei denn, er ist über `aria-describedby` die
Beschreibung eines Feldes. **Sie sieht den ersten Blick** jeder Seite — eine
Lasche, die erst ein Klick öffnet, sieht sie nicht; dafür bleibt der Lauf
über den Katalog.

---

## Was diese Durchsicht nicht sagt

- **Ob der sichtbare Satz der richtige ist.** Der Wächter misst Zeichen; ob
  der erste Satz beim Entscheiden hilft, sieht nur ein Mensch (§10).
- **Ob hinter dem Fragezeichen etwas versteckt wurde, das man braucht.** Ein
  Fall ist gefunden und behoben worden (`anmeldungErklaerung`); dass es der
  einzige war, ist **nicht** belegt.
- **Wie lang die Seiten jetzt wirklich sind** — stand hier bis zum
  2026-09-24. Gemessen ist es seitdem, am laufenden Fenster (oben).
- **Ob ein Text „nach AI-Slop klingt".** Die Frage, mit der #1771 anfängt,
  beantwortet kein Lauf. Was dieser Durchgang tut, ist den **langen Teil aus
  dem Weg räumen** — wer ihn danach liest, liest ihn, weil er ihn wollte.
