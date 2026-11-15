# Die Textmodelle an denselben Fragen

**2026-10-04 08:56** — AI-02.AK-03 (#1743),
Commit e03d48822, Mac17,3, macOS 27.0.1, Testsatz testsatz-1.

    go test -count=1 -tags fmlauf -run TestDieTextmodelleImVergleich -v ./internal/app/

**Gefragt wird, wie die Anwendung fragt**: Anweisung `KIKartenanweisung`, die
Auftragssätze des Formulars, ein Schema mit dem einen Pflichtfeld — über
dieselben Klienten, die die Anwendung einhüllt. **Kartenkontext und Schema
sind nachgebaut** wie in `fmguete.md`: die Zahlen gelten für diesen Testsatz
in dieser Form. Jedes Modell bekommt erst eine Aufwärmfrage (sie zählt als
„erste Antwort"), dann 2 Runden über 32 Fragen.

| Modell | Fragen | formtreu | Sprache gehalten | leer | kein JSON | abgelehnt | abgeschnitten | Fehler | erste Antwort | p50 | p95 | Runde 1 = Runde 2 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Gerätemodell (fm) | 64 | 64 | 64 | 0 | 0 | 0 | 0 | 0 | 1.16 s | 0.61 s | 1.01 s | 8 von 32 |
| qwen3:0.6b | 64 | 64 | 64 | 0 | 0 | 0 | 0 | 0 | 0.11 s | 0.12 s | 0.44 s | 32 von 32 |
| gemma3:1b | 64 | 64 | 62 | 0 | 0 | 0 | 0 | 0 | 0.22 s | 0.31 s | 0.69 s | 32 von 32 |

**formtreu** heißt: gültiges JSON mit dem verlangten Feld, nicht leer.
**Sprache gehalten** ist die grobe Probe aus `fmguete_test.go` und zählt nur
formtreue Antworten. **Ob ein Vorschlag passt, entscheidet der Lauf nicht** —
die Spalte „passt?" füllt, wer bewertet, und nennt sich.

## Wer „passt?" bewertet hat, und was dabei herauskam

**Claude (Opus 5.5) am 2026-10-04, nicht der Auftraggeber**, mit demselben
Maß wie in `fmguete.md`: **ja** — so oder fast so auf die Karte; **teils** —
ein brauchbarer Anstoß; **nein** — der Kartentyp, der Regeltext, eine
Beschreibung des Auftrags statt einer Antwort, oder die falsche Sprache. Das
ist eine Einschätzung und keine Abnahme; die Spalten stehen, damit jemand
anders sie nachprüfen und überstimmen kann.

| Modell | ja | teils | nein | davon Namen (ja/teils/nein) | davon Flavortexte |
| --- | ---: | ---: | ---: | --- | --- |
| Gerätemodell (fm) | 18 | 10 | 4 | 8/5/3 | 10/5/1 |
| gemma3:1b | 9 | 14 | 9 | 3/4/9 | 6/10/0 |
| qwen3:0.6b | 0 | 7 | 25 | 0/1/15 | 0/6/10 |

**Formtreu sind alle drei vollständig — passend ist nur eines.** Die Form
hält die Laufzeit (Ollama über `format`, das Gerätemodell über sein Schema);
was in der Form steht, hält niemand.

- **qwen3:0.6b gibt als Namen den Kartentyp** — „Golem", „Land",
  „Artefakt", „Hexerei" —, auf Englisch gefragt fünfmal von acht das
  deutsche Wort, und **beschreibt im Flavortext den Auftrag**, statt ihn zu
  erfüllen: „Eine kurze Flavortext: Stimmung, die den Regeltext nicht
  wiederholt." Zehn von sechzehn Flavortexten sind unbrauchbar, neun davon
  solche Echos.
- **gemma3:1b**: neun von sechzehn Namen sind unbrauchbar — der Typ, ein
  Gattungswort, die falsche Sprache —, aber kein einziger Flavortext. Und
  es ist schnell: p50 0,31 s gegen 0,61 s beim Gerätemodell, beides unter
  dem, was ein Mensch am Formular merkt.
- **Die grobe Sprachprobe sieht davon fast nichts**: in Runde 1 fand sie
  **eine** Antwort in der falschen Sprache, die Bewertung **neun** (zwei beim
  Gerätemodell, fünf bei qwen3, zwei bei gemma3). Ein deutsches Wort ohne
  Umlaut — „Spontanzauber", „Verbindung", „Eiswurm" — kommt durch.
- **Gleichlauf**: die zwei Ollama-Modelle antworten bei Temperatur 0 in
  beiden Runden wörtlich gleich; das Gerätemodell nur bei 8 von 32 Fragen.

## Ein erster Lauf, der nicht zu Ende kam

**Er stand vor diesem und wurde nach drei Minuten abgebrochen**: `qwen3:0.6b`
dachte auf die erste Frage über 7.400 Token lang, ohne Inhalt zu geben. Der
Klient schickte kein `think`, und **ohne das Feld schaltet Ollama 0.35.1 das
Denken bei qwen3 ein**. Wie lange, hängt an der Frage: über alle 32 im
Median 267 Token und 1,8 s, und eben diese eine nach drei Minuten noch nicht
fertig (`modellmessung.md`, „Ein Modell, das denken kann"). **Behoben im
selben Zug**: `/api/chat` schickt jetzt `"think": false`
(`internal/ollama/rede.go`, Vermerk 0098); dieser Lauf ist der danach.

## Runde 1, Frage für Frage

| Karte | Sprache | Feld | Gerätemodell (fm) | passt? | qwen3:0.6b | passt? | gemma3:1b | passt? |
| ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| 1 | de | name | Silberhahn (0.5 s) | teils — eine Henne für einen Hirsch | Hirn (0.1 s) | nein — „Gehirn“ | Hirsch (0.2 s) | nein — der Typ |
| 1 | de | flavor text | Der Hirsch lauert im Schatten und wartet auf den richtigen Moment, um zu handeln. (0.9 s) | teils — lauern passt nicht zu Wachsamkeit | Eine warme, scharfe und lebendige Stimmung, die die Regeltexte nicht wiederholt, sondern verbindet mit der Geschichte der Hirsche. (0.3 s) | nein — beschreibt den Auftrag | Der Hirsch zieht sich in den Schatten zurück, bereit für eine gefährliche Jagd. Seine Augen suchen nach neuen Beute. (0.6 s) | teils — ein Hirsch auf Beutejagd, „neuen Beute“ |
| 1 | en | name | Silvertail (0.5 s) | ja | Hirscher (0.1 s) | nein — kein Englisch, kein Name | *Sprache verfehlt*: Löwenkönig (0.3 s) | nein — deutsch, ein Löwe |
| 1 | en | flavor text | The forest holds its breath as the stag emerges, a silent guardian of the clearing. (0.9 s) | ja | A quiet, reflective mood that evokes the quiet strength of a hiker (0.2 s) | nein — beschreibt den Auftrag, ein Wanderer | A solitary hunter, tracking a distant deer through a shadowed forest. (0.4 s) | teils — der Jäger statt des Hirschs |
| 2 | de | name | Ignar (0.5 s) | ja | Drache (0.1 s) | nein — der Typ | Drache (0.3 s) | nein — der Typ |
| 2 | de | flavor text | Ein sanfter Wind trägt die Erinnerung an alte Schlachten durch die Luft. (0.8 s) | teils — Stimmung ohne den Drachen | Eine flügelfreie Kreatur, die fliegt, mit Fluch und Flügeln, die den Regeln folgen. Der Flavortext füllt die Karte mit: Stimmung, Geschichte. (0.4 s) | nein — beschreibt den Auftrag | Der Drache brennt mit einem unaufhaltsamen Feuer. Seine Wut ist unberechenbar und lässt die Luft in der Nähe schwerer werden. (0.6 s) | ja |
| 2 | en | name | Sky Dragon (0.4 s) | teils — fast der Typ | Drakeneater (0.1 s) | teils | Ignis (0.2 s) | ja |
| 2 | en | flavor text | The dragon glides through the air, leaving a trail of pain behind it. (0.8 s) | ja | A bold, fiery spirit that brings danger and power to the battlefield. (0.2 s) | teils | A fiery guardian, watching over the land. (0.3 s) | teils |
| 3 | de | name | Lina (0.4 s) | nein — ein Personenname für einen Zauber | Spontaner (0.1 s) | nein — aus dem Typ | Spontane Wurf (0.3 s) | nein — aus dem Typ, ungrammatisch |
| 3 | de | flavor text | Ein plötzlicher Wandel der Stimmung führt zu unerwarteten Entscheidungen. (0.8 s) | teils | Eine zartes, warmes und lebendiges Flavortext, das die Stimmung der Spontanzkarte mit einem kurzen, lebendigen Flavortext verbindet. (0.4 s) | nein — beschreibt den Auftrag | Ein warmer Sommer, ein leichter Wind... und ein bisschen Glück. (0.5 s) | teils |
| 3 | en | name | Ephemeral Moment (0.5 s) | ja | Spontanzauber (0.1 s) | nein — der Typ, deutsch | Verbindung (0.3 s) | nein — deutsch |
| 3 | en | flavor text | A sudden urge to explore the unknown, pulling two cards and discarding one at random. (0.9 s) | nein — wiederholt die Regel | A story of chance and unexpected twists (0.1 s) | teils | A bittersweet memory of autumn, tinged with longing. (0.3 s) | teils — ohne Bezug |
| 4 | de | name | Eldarion (0.6 s) | teils — klingt nach einer Figur | Hexerei (0.1 s) | nein — der Typ | Zauberer (0.3 s) | nein — ein Gattungswort |
| 4 | de | flavor text | Ein kalter Hauch von Magie durchdringt die Luft, als eine Kreatur mit großer Macht dem Verderben übergeben wird. (1.0 s) | ja | Eine grauere, dunkle Farbe, die den Text nicht wiederholt, und eine kurze Geschichte, die den Regeltext nicht wiederholt. (0.3 s) | nein — beschreibt den Auftrag | Verlorene Magie, die in den Schatten verweilt. (0.5 s) | teils |
| 4 | en | name | Eiswurm (0.5 s) | nein — deutsch | Hexe (0.1 s) | nein — deutsch | Shadow Weaver (0.3 s) | teils — eine Figur für einen Zauber |
| 4 | en | flavor text | A shadow stirs in the moonlight, seeking strength to break the silence. (0.8 s) | teils | A mysterious force begins to unravel (0.1 s) | teils | A lonely traveler, seeking solace in the shadows. (0.3 s) | teils — ohne Bezug |
| 5 | de | name | Luminara (0.4 s) | ja | Artefakt (0.1 s) | nein — der Typ | Magischer Artefakt (0.3 s) | nein — der Typ |
| 5 | de | flavor text | Ein vergessener Zauber, der Licht in die Dunkelheit bringt. (0.7 s) | ja | Eine kurze Flavortext: Stimmung, die den Regeltext nicht wiederholt. Geschichte, die den Regeltext nicht wiederholt. (0.3 s) | nein — beschreibt den Auftrag | Ein veralteter Kristall, der in der Dunkelheit schimmert. Er wurde von einer längst vergessenen Zivilisation erschaffen, um die Kraft des Magiers zu verstärken. (0.7 s) | ja |
| 5 | en | name | Arcane Sigil (0.5 s) | ja | Artefakt (0.1 s) | nein — der Typ, deutsch | Echo (0.2 s) | teils |
| 5 | en | flavor text | The shimmering relic hums with hidden energy, forging a new path of color into the void. (0.9 s) | ja | A story of a traveler who finds a hidden world of colors (0.2 s) | teils | Lost echoes of a forgotten kingdom. A shimmering relic, imbued with a melancholy beauty. (0.4 s) | ja |
| 6 | de | name | Luminara (0.4 s) | teils — derselbe Name wie Karte 5 | Kreaturen (0.1 s) | nein — aus dem Regeltext | Verzauberung (0.3 s) | nein — der Typ |
| 6 | de | flavor text | Die Luft knistert vor magischer Energie, während die Kreaturen stärker werden. (0.8 s) | ja | Stimmung: Ein lebendiger, zärtliches Flavortext, das die Karte verzaubert und die Regeltext verbindet. Geschichte: Ein kurzer, lebendiger Text, der die Regeltext nicht wiederholt. (0.4 s) | nein — beschreibt den Auftrag | Die Verzauberung flüstert Geheimnisse, während die Dunkelheit sich nähern. (0.6 s) | teils — „sich nähern“ |
| 6 | en | name | Flesche Magier (0.5 s) | nein — kein Wort | Verzauberung (0.1 s) | nein — der Typ, deutsch | Shadow Weaver (0.3 s) | teils — derselbe Name wie Karte 4 |
| 6 | en | flavor text | The magic hums around them, giving them strength in uncertain times. (0.7 s) | ja | A mystical realm where creatures rise from shadows, each with a unique power to influence the world around them. (0.2 s) | teils | A flicker of magic, a shadowed promise.  The air grows cold, and a strange beauty takes hold. (0.4 s) | teils |
| 7 | de | name | Schwarzwald (0.6 s) | teils — ein echter Ort | Land (0.1 s) | nein — der Typ | Wüstenland (0.2 s) | nein — Wüste für ein grünes Land |
| 7 | de | flavor text | Ein geheimer Bericht über verborgene Wege führt zu diesem Land. Es erwacht plötzlich in der Mitte der Schlacht. (1.1 s) | teils | Stimmung (0.2 s) | nein — ein Wort aus dem Auftrag | Die Steppenlandschaft ist bekannt für ihre reiche Vegetation und die Bewohner, die sich an die Elemente anpassen. Es gibt hier eine tiefe Verbindung zur Natur und eine starke Tradition der Selbstversorgung. (0.7 s) | teils — liest sich wie ein Lexikon |
| 7 | en | name | Emerald Clearing (0.4 s) | ja | Tippy (0.1 s) | nein — aus dem Regeltext | Veridia (0.3 s) | ja |
| 7 | en | flavor text | A hidden realm emerges, pulsing with unseen energy. The ground hums with quiet power. (0.9 s) | ja | A vibrant, lively land with a playful mood, where players can tap to create a unique and dynamic game element, such as a magical creature or a special item. (0.3 s) | nein — beschreibt das Spiel | A quiet, misty valley, filled with ancient trees and a sense of profound solitude. (0.5 s) | ja |
| 8 | de | name | Tharron (0.5 s) | ja | Golem (0.1 s) | nein — der Typ | Steinwerfer (0.3 s) | teils — wirft nicht, kann nicht angreifen |
| 8 | de | flavor text | Ein alter Golem wacht stumm in der Dunkelheit und bewacht sein Erbe. (1.0 s) | ja | Eine grauere, dunkle Farbe, die den Regeln verweist, und ein kurzer Flavortext, der die Stimmung der Karte verweist. (0.3 s) | nein — beschreibt den Auftrag | Ein alter Golem, der über die Ruinen wacht. Er ist ein stiller Beschützer, der seine Aufgabe mit unerschütterlicher Stärke ausführt. (0.6 s) | ja |
| 8 | en | name | Stone Sentinel (0.5 s) | ja | Golem (0.1 s) | nein — der Typ | Stoneguard (0.2 s) | ja |
| 8 | en | flavor text | The golem stands silent, a statue of stone and will. It waits, unmoving, for the storm to pass. (1.0 s) | ja | A quiet, mysterious force that moves with the wind, guarding the land from outside threats. (0.2 s) | teils | A silent guardian, weathered and slow, watching over the ruins. (0.4 s) | ja |
