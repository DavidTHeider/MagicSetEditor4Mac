# Güte des Gerätemodells am Testsatz

**AI-03.AK-04** (#1744):
Qualität, Fehlvorschläge und mehrsprachige Beispiele an einem fachlichen
Testsatz — acht **selbst geschriebene** Karten, je Name und Flavortext, auf
Deutsch und Englisch, am Gerätemodell dieses Macs (AI-05, macOS 27.0.1,
Mac17,3).

**Was davon das Formular ist und was nachgebaut** (Z6 des Zweitblicks über
Phase 6.10): die **Anweisung** ist die der Anwendung (`KIKartenanweisung`,
wörtlich), die **Auftragssätze** sind die des Formulars. **Kartenkontext und
Schema sind nachgebaut**: der Kontext in der Gestalt von `kifeldauftrag.ts`
(Satz, „Die Karte bisher:", je Feld eine Zeile), aber mit den zwei Zeilen
`type` und `rule text` statt der Textfelder einer echten Vorlage; das Schema
fragt das eine Feld als Pflicht, wie das Formular, ohne dessen weitere
Angaben. **Die Zahlen gelten für diesen Testsatz in dieser Form** — nicht für
jede Karte am Formular. So sieht eine Frage aus:

    Schlage einen passenden Eigennamen für diese Karte vor — nicht den Kartentyp und keine Beschreibung der Regel. Antworte auf Deutsch, nur mit dem Namen.

    Die Karte bisher:
    type: Kreatur — Hirsch
    rule text: Wachsamkeit. Wenn diese Kreatur ins Spiel kommt, erhältst du 2 Lebenspunkte.

    Schema: {"type":"object","required":["name"],"properties":{"name":{"type":"string"}}}

    go test -count=1 -tags fmlauf -run TestDieGueteAmTestsatz -v ./internal/applefm/

## Wer „passt?“ bewertet hat

**Claude (Opus 5.5) am 2026-10-04, nicht der Auftraggeber.** Gefragt ist:
gäbe ein Mensch, der ein Set baut, diesen Vorschlag so oder leicht geändert
auf die Karte? **ja** — so oder fast so; **teils** — brauchbarer Anstoß;
**nein** — falsch, Regeltext, Kartentyp oder falsche Sprache. Das ist eine
Einschätzung und keine Abnahme; die Spalte steht, damit jemand anders sie
nachprüfen und überstimmen kann.

## Zwei Läufe — und was zwischen ihnen geändert wurde

| | ja | teils | nein | davon Name (ja/teils/nein) | davon Flavortext |
| --- | ---: | ---: | ---: | --- | --- |
| **Lauf 1** — die Sätze des Formulars bis dahin | 9 | 11 | 12 | 4/7/5 | 5/4/7 |
| **Lauf 2** — mit geschärften Sätzen | 20 | 9 | 3 | 9/4/3 | 11/5/0 |

**Der Befund aus Lauf 1**: sieben von sechzehn Flavortexten wiederholten den
Regeltext, und fünf Namen waren der Kartentyp, eine Beschreibung der Regel
oder in der falschen Sprache. **Geändert** wurden daraufhin die zwei Sätze des
Formulars in allen acht Sprachen (`kiFeldAuftragName`, `kiFeldAuftragFlavor`
in `frontend/src/sprache/*/seitenki.ts`): ein Name soll ein **Eigenname**
sein, „nicht den Kartentyp und keine Beschreibung der Regel“; ein Flavortext
„Stimmung oder Geschichte, die den Regeltext nicht wiederholt“.

**Was bleibt**: zwei Namen in der falschen Sprache (Lauf 2, Karte 4) und
einer, der nur der Typ ist (Karte 8, englisch). **Die grobe Sprachprobe des
Laufs sieht davon nur das Umlautwort** — sie zählt einen kurzen Namen als
deutsch und erkennt Englisch nur an fehlenden Umlauten; die Spalte „passt?“
ist die eigentliche Prüfung. Der Vergleich mit zwei kleinen Ollama-Modellen
an denselben Fragen steht in [`modellvergleich.md`](modellvergleich.md)
(AI-02, #1743).
**Die Fragen stehen seit AI-02 in `testdaten/ki/testsatz.json`** — eine Datei
für beide Läufe.

## Lauf 2 — geschärfte Sätze

**2026-10-04 04:54** — AI-03.AK-04 (#1744), 32 Fragen: acht selbst geschriebene Karten, Name

| Ausgang | Zahl |
| --- | ---: |
| fertig, Sprache gehalten | 31 |
| fertig, Sprache **nicht** gehalten | 1 |
| leer | 0 |
| abgelehnt | 0 |
| Fehler | 0 |

| Karte | Sprache | Feld | Ausgang | Antwort | Dauer | passt? |
| ---: | --- | --- | --- | --- | ---: | --- |
| 1 | de | name | fertig | Silberhuf | 0.98 s | ja |
| 1 | de | flavor text | fertig | Der Hirsch lauert im Schatten und lauscht auf jedes Geräusch. Seine stille Präsenz verändert die Atmosphäre sofort. | 1.06 s | teils — „lauert“ passt nicht zu Wachsamkeit, der zweite Satz sagt nichts |
| 1 | en | name | fertig | Silvertail | 0.41 s | ja |
| 1 | en | flavor text | fertig | The forest holds its breath as the stag emerges, a silent watcher cloaked in moonlight. | 0.88 s | ja |
| 2 | de | name | fertig | Ignar der Drache | 0.47 s | ja |
| 2 | de | flavor text | fertig | Ein sanftes Zischen in der Luft, während der Drache über die Schlacht zieht. | 0.82 s | ja |
| 2 | en | name | fertig | Sky Drake | 0.41 s | ja |
| 2 | en | flavor text | fertig | The dragon glides through the air, leaving a trail of fire and fear. | 0.75 s | ja |
| 3 | de | name | fertig | Luminara | 0.40 s | teils — beliebig, ohne Bezug zur Karte |
| 3 | de | flavor text | fertig | Ein plötzlicher Wechsel der Stimmung – plötzlich wird alles anders. | 0.77 s | teils — leer, „plötzlich“ zweimal |
| 3 | en | name | fertig | Moment of Choice | 0.44 s | ja |
| 3 | en | flavor text | fertig | A sudden urge to explore unseen paths, pulling two cards from the deck before discarding one. | 0.80 s | teils — die zweite Hälfte ist die Regel |
| 4 | de | name | fertig | Eldritch Pact | 0.58 s | nein — ein englischer Name in der deutschen Antwort (die grobe Probe zählt kurze Namen als deutsch) |
| 4 | de | flavor text | fertig | Ein kalter Hauch von Magie durchdringt die Luft, als eine Kreatur mit unermesslicher Stärke zerbricht. | 0.97 s | ja |
| 4 | en | name | fertig, Sprache? | Erzwörung der Macht | 0.58 s | nein — deutsch und kein Wort |
| 4 | en | flavor text | fertig | A shadow stirs in the moonlight, seeking strength. | 0.63 s | teils — passt kaum zu einer Zerstörung |
| 5 | de | name | fertig | Luminar | 0.53 s | teils — beliebig |
| 5 | de | flavor text | fertig | Ein alter Schlüssel, der verborgene Energien freisetzt. | 0.80 s | ja |
| 5 | en | name | fertig | Arcane Essence | 0.39 s | ja |
| 5 | en | flavor text | fertig | A flicker of color emerges from the void, born of forgotten energy. | 0.72 s | ja |
| 6 | de | name | fertig | Luminar | 0.40 s | teils — beliebig, derselbe wie bei Karte 5 |
| 6 | de | flavor text | fertig | Die Luft ist erfüllt von einer magischen Präsenz, die jede Kreatur stärkt, die unter deiner Kontrolle steht. | 0.89 s | teils — endet mit der Regel |
| 6 | en | name | fertig | Fierce Presence | 0.51 s | ja |
| 6 | en | flavor text | fertig | The magic hums around them, making them stronger with every breath. | 0.74 s | ja |
| 7 | de | name | fertig | Schwarzwald | 0.59 s | teils — ein wirklicher Ortsname |
| 7 | de | flavor text | fertig | Ein geheimnisvolles Land erwacht langsam zum Leben, während die ersten Schritte im Sand widerhallen. | 0.98 s | ja |
| 7 | en | name | fertig | Sunlit Clearing | 0.50 s | ja |
| 7 | en | flavor text | fertig | The land hums with quiet power, hidden beneath a veil of silence. It awakens only when you whisper its name. | 1.02 s | ja |
| 8 | de | name | fertig | Steinwächter | 0.48 s | ja |
| 8 | de | flavor text | fertig | Ein alter Wächter aus Stein bewacht den Ort in stummer Präsenz. | 0.86 s | ja |
| 8 | en | name | fertig | Golem | 0.41 s | nein — nur der Kartentyp |
| 8 | en | flavor text | fertig | The golem stands silent, its body shaped by ancient forces, unable to fight yet ever watchful. | 0.89 s | ja |

## Lauf 1 — die Sätze bis dahin

**2026-10-04 04:53** — AI-03.AK-04 (#1744), 32 Fragen: acht selbst geschriebene Karten, Name

| Ausgang | Zahl |
| --- | ---: |
| fertig, Sprache gehalten | 31 |
| fertig, Sprache **nicht** gehalten | 1 |
| leer | 0 |
| abgelehnt | 0 |
| Fehler | 0 |

| Karte | Sprache | Feld | Ausgang | Antwort | Dauer | passt? |
| ---: | --- | --- | --- | --- | ---: | --- |
| 1 | de | name | fertig | Wachsamkeit des Hirschs | 1.15 s | teils — nennt die Fähigkeit statt eines Namens |
| 1 | de | flavor text | fertig | Der Hirsch lauert im Schatten und schlägt zu, sobald die Gefahr naht. Mit jedem Schritt bringt er neue Energie in das Spiel. | 1.17 s | teils — „lauert“ passt nicht zu Wachsamkeit; der zweite Satz redet über das Spiel |
| 1 | en | name | fertig | Alert Deer | 0.46 s | ja |
| 1 | en | flavor text | fertig | The deer arrives with caution, bringing you two points of life. | 0.61 s | nein — wiederholt die Regel (Lebenspunkte) |
| 2 | de | name | fertig | Drachenfluch | 0.47 s | teils — klingt nach einem Spruch, nicht nach einer Kreatur |
| 2 | de | flavor text | fertig | Ein majestätischer Drache erhebt sich in den Himmel und lässt sofort Schaden bei jedem Gegner entfesseln. | 0.99 s | teils — holprig und nah an der Regel |
| 2 | en | name | fertig | Dragon | 0.44 s | nein — nur der Kartentyp |
| 2 | en | flavor text | fertig | It soars above, striking each enemy with relentless precision from the skies. | 0.94 s | ja |
| 3 | de | name | fertig, Sprache? | Zwei Karten, eine Abweisung | 0.56 s | nein — beschreibt die Regel; „Abweisung“ ist nicht Abwerfen (die Sprache stimmt, die grobe Probe irrte) |
| 3 | de | flavor text | fertig | Ein kurzer Zauber, der zwei Karten zieht und dann eine wegschickt. Eine unerwartete Mischung. | 0.93 s | nein — Regelbeschreibung |
| 3 | en | name | fertig | Dual Draw and Discard | 0.53 s | nein — Regelbeschreibung |
| 3 | en | flavor text | fertig | Draw two cards, then discard one. | 0.61 s | nein — der Regeltext wörtlich |
| 4 | de | name | fertig | Hexenmeister | 0.41 s | nein — „Hexenmeister“ ist eine Person, die Karte eine Hexerei |
| 4 | de | flavor text | fertig | In der Dunkelheit erwacht eine uralte Macht, die kraftvolle Wesen niederstreckt. | 0.94 s | ja |
| 4 | en | name | fertig | Hexenwerk | 0.43 s | nein — ein deutsches Wort in der englischen Antwort (die grobe Probe sah es nicht) |
| 4 | en | flavor text | fertig | A single flame can end a creature's fight. | 0.68 s | ja |
| 5 | de | name | fertig | Farbe der Macht | 0.41 s | teils — allgemein |
| 5 | de | flavor text | fertig | Dieses Artefakt erzeugt ein Mana beliebiger Farbe – ein mächtiges Werkzeug für jede Situation. | 0.80 s | nein — Regelbeschreibung |
| 5 | en | name | fertig | Mana Generator | 0.40 s | teils — allgemein |
| 5 | en | flavor text | fertig | By tapping this artifact, you create one mana of any color. | 0.74 s | nein — der Regeltext umformuliert |
| 6 | de | name | fertig | Macht der Stärke | 0.55 s | teils — doppelt gemoppelt |
| 6 | de | flavor text | fertig | Diese Verzauberung verleiht deinen kontrollierten Kreaturen zusätzliche Stärke. | 0.76 s | nein — Regelbeschreibung |
| 6 | en | name | fertig | Power Surge | 0.42 s | ja |
| 6 | en | flavor text | fertig | Your creatures grow stronger, gaining more power with every turn. | 0.73 s | teils — „with every turn“ stimmt nicht, der Bonus ist fest |
| 7 | de | name | fertig | Geheimnisvolles Land | 0.61 s | teils — allgemein |
| 7 | de | flavor text | fertig | Dieses Land überrascht dich mit einem Geheimnis – es bringt {G} mit sich. | 0.81 s | nein — Regel samt Manasymbol |
| 7 | en | name | fertig | Emerald Clearing | 0.41 s | ja |
| 7 | en | flavor text | fertig | This land brings energy into the fold with every tap. | 0.76 s | teils |
| 8 | de | name | fertig | Wächter aus Stein | 0.50 s | ja |
| 8 | de | flavor text | fertig | Ein alter Golem, der die Schlacht meidet, während er die Verteidigung sichert. | 0.83 s | ja |
| 8 | en | name | fertig | Protector | 0.37 s | teils — allgemein |
| 8 | en | flavor text | fertig | It stands firm, unable to fight back. | 0.61 s | ja |
