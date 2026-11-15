# 0021 — Das Referenzbild kommt vom Nutzenden, und das Bildfenster wird ausgespart

**Stand:** angenommen
**Datum:** 2026-08-16
**Gehört zu:** Phase A ([0020](0020-acht-phasen-fuer-die-paritaet.md)),
#290

## Lage

Das Ziel dieser Iteration ist **Parität zu MSE2**: dieselbe Vorlage, dieselben
Kartendaten, dasselbe Bild. Gemessen wird das an einem Goldbild — und dafür
braucht es eine Ausgabe von MSE2.

**Was es schon gibt**, und das gehört zuerst gesagt, damit diese Entscheidung
nicht mehr beansprucht, als sie entscheidet:

| Vergleich | Was er festhält |
| --- | --- |
| `render/goldbild_test.go` | den **Rasterer** — ein von Hand geschriebener Plan wird zum Bild |
| `app/saatgoldbild_test.go` | die **Naht** Stil → Plan → Bild, aus einem echten Stilpaket der eigenen Saat |

Beide vergleichen gegen **uns selbst**. Sie fangen jede Regression — und genau
das ist ihr Wert, denn drei der letzten fünf Fehler dieses Projekts waren am
Bild zu sehen und in keinem Test. Was sie nicht können, ist zu sagen, ob das,
was gleich bleibt, überhaupt richtig ist. Das ist die Lehre aus Phase 9: **ein
Rundlauf prüft Selbstkonsistenz, nicht Wahrheit.**

**Und ein Referenzbild darf nicht ins Repository** (§11). Es zeigt Kartenrahmen
und Symbole von Wizards of the Coast, dazu die Illustration und den Text eines
fremden Urhebers.

## Entscheidung

**Drei Teile, und der dritte ist der unbequeme.**

### 1 Das Referenzbild wird beim Lauf bereitgestellt

Ein Verzeichnis, genannt über `MSE_REFERENZBILDER`, hinter dem Build-Tag
`fremdpakete`. Es enthält je Referenz zwei Dateien:

```
<name>.jpg      die Ausgabe von MSE2, in der Auflösung der Vorlage
<name>.karte    die Kartendaten, die sie erzeugt haben — Feld für Feld
```

**Auch die `.karte` bleibt draußen**, nicht nur das Bild. Sie trägt Name, Typ
und Regeltext einer fremden Karte; das ist die schöpferische Leistung ihres
Urhebers und gehört so wenig ins Repository wie der Rahmen.

Fehlt das Verzeichnis, wird der Lauf **übersprungen** — wie jeder Lauf hinter
`fremdpakete`.

### 2 Der Vergleich läuft über die Naht, nicht über einen Plan von Hand

Er geht denselben Weg wie `saatgoldbild_test.go`: Stilpaket → Planbauer →
Rasterer, mit der Schriftsuche des Rechners.

### 3 Das Bildfenster wird ausgespart — und die Aussparung wird beziffert

**Die Illustration liegt uns nicht vor.** Ein Referenzbild aus einem Forum ist
ein fertig gerechnetes JPG; die Quelldatei der Illustration ist nicht dabei und
wird es nie sein. Unsere Ausgabe zeichnet dort nichts.

Der Vergleich meldet deshalb **zwei** Zahlen:

| Zahl | Was sie sagt |
| --- | --- |
| **Ganze Karte** | der wahre Abstand, ungeschönt. Er kann die Toleranz nie erreichen |
| **Ohne Bildfenster** | der Abstand an allem, was die Phasen B bis F bewegen: Rahmen, Titel, Manasymbole, Typzeile, Textkasten, Fußzeile |

**Die Aussparung wird nicht geraten.** Sie ist das Rechteck, das der
**Zeichenplan** für das Bildfeld führt — dieselbe Quelle, aus der auch gezeichnet
wird. Ein hart geschriebenes Rechteck wäre eine zweite Wahrheit neben dem Plan
und liefe bei jeder Stiländerung von ihm weg.

**Die Toleranz ist die bestehende** (`render.Toleranz` = 1,0 mittlerer Abstand,
`render.Spitzentoleranz` = 2 je Kanal). Sie ist in Phase 4 begründet worden und
bekommt hier keine neue Zahl — eine Toleranz, die zum Messobjekt passend
gewählt wird, misst nichts.

## Verworfen

| Möglichkeit | Warum nicht |
| --- | --- |
| **Nur gegen die eigene Saat vergleichen** | Ein Rundlauf. Er hielte jede Abweichung von MSE2 für richtig, sobald sie einmal drinsteht — und genau davon handelt diese Iteration |
| **Ein Referenzbild ins Repository legen** | §11. Rahmen und Symbole gehören Wizards of the Coast, Illustration und Text einem fremden Urheber |
| **Die `.karte` ins Repository, nur das Bild draußen** | Der Text einer fremden Karte ist die schöpferische Leistung, nicht ihr Beiwerk. Die Trennung wäre willkürlich |
| **`card-sample.png` aus den Stilpaketen als Referenz** | Nachgesehen am 2026-08-16: 75×104 Bildpunkte gegen 375×523 der Vorlage, und aus **unbekannten** Kartendaten. Parität heißt „dieselben Kartendaten"; ein Vorschaubildchen erfüllt keine Hälfte davon |
| **Ohne Aussparung vergleichen, eine Zahl** | Sie wäre ehrlich und **unbrauchbar**: die Illustration füllt rund ein Drittel der Karte und kann nie stimmen. Jede Verbesserung an Rahmen oder Text verschwände in ihrem Rauschen. Deshalb **beide** Zahlen, nicht die eine statt der anderen |
| **Nur ohne Bildfenster vergleichen, eine Zahl** | Das wäre das Wegdefinieren, vor dem der Auftrag warnt. Die ganze Karte steht mit im Protokoll |
| **Die Aussparung als festes Rechteck je Stil** | Eine zweite Wahrheit neben dem Zeichenplan. Sie liefe bei der ersten Stiländerung weg, und niemand merkte es |
| **Eine eigene, größere Toleranz für diesen Vergleich** | Eine Toleranz, die man wählt, nachdem man den Abstand kennt, ist keine Messung, sondern eine Ausrede (§10) |

## Folgen

- **Der Lauf ist rot, und er soll es sein.** Er beziffert den Abstand zu MSE2;
  jede Phase von B bis F wird an der Bewegung dieser Zahl gemessen. Das ist
  dieselbe Bauform wie `exportfremd_test.go`, der seit Phase 6 rot in der Ablage
  liegt und „13 von 19" zählt: was hinter `fremdpakete` steht, läuft nicht in
  der Pipeline und zählt nicht zur Abdeckung.
- **Was der Vergleich nicht prüft, steht neben ihm** (§10): JPG ist verlustbehaftet,
  und die Kompression allein erzeugt bereits einen Abstand über der Toleranz.
  Wer die Zahl liest, liest sie als Verlauf und nicht als absoluten Betrag.
- Das Verzeichnis ist in `docs/RELEASE.md` als einer der Läufe zu führen, die
  vor einer Freigabe gefahren werden.
