# 0107 — Eine Einschätzung ist keine Regelauskunft

**Datum:** 2026-09-14 · **Stand:** entschieden ·
**Anlass:** Werkstück A13.2a von Phase 5.3
(#1452),
Ergänzung vom 2026-09-10

## Die Frage

Ein Modell beurteilt eine Karte und stützt sich dabei auf Stellen, die dieser
Dienst ihm gegeben hat. **Was tut das Programm, wenn die Antwort etwas nennt,
das in keiner dieser Stellen steht?**

Das Issue verlangt eine Richtung und lässt die Umsetzung offen:

> Fehlender/veralteter/widersprechender Korpus führt zu sichtbarer
> Einschränkung oder fehlendem Beleg, **niemals erfundenem Regelzitat**.

## Drei Wege, und zwei davon sind schlechter

| | Was es hieße | Warum nicht |
| --- | --- | --- |
| **Verwerfen** | die ganze Antwort fällt weg | Ein Satz mit einer falschen Nummer muss nicht falsch sein. Eine richtige Auskunft wegen eines Streifschusses zu verwerfen ist teuer und hilft niemandem |
| **Stillschweigend streichen** | die Nummer verschwindet aus dem Satz | **Der Rest sieht danach geprüft aus.** Das ist die gefährlichste der drei: sie erzeugt genau den Eindruck, den die Regel verhindern soll |
| **Zeigen und danebenschreiben** | Satz und Befunde stehen zusammen | Gewählt |

## Was gilt

**Die Antwort wird gezeigt, und was an ihrem Quellenanspruch nicht stimmt,
steht daneben.** [domain.PruefeEinschaetzung] hält jede genannte
Belegkennung, jede Regelnummer und jede Adresse gegen die vergebenen Belege;
was übrig bleibt, ist ein `Belegbefund` und keine Quelle.

**`belegt` ist eine Aussage über den Quellenanspruch und nicht über die
Wahrheit.** `false` heißt: die Antwort nennt etwas, das in keinem Beleg steht.
Sie kann trotzdem stimmen — nur nicht **von hier aus**.

**Eine Antwort ohne jede Fußnote ist kein Fehler.** „Dazu steht in den Belegen
nichts" ist eine richtige Antwort; sie ist nur keine belegte. Sie bekommt
deshalb den Befund `ohnebeleg` und nicht die Marke „belegt".

**Und sie ist nie amtlich.** Der Vorbehalt steht als Konstante am Domänentyp
([domain.EinschaetzungsVorbehalt]) und nicht in einer Oberfläche: vier Flächen
sind vier Gelegenheiten, ihn zu vergessen, und die eine, an der er fehlt, ist
die, an der jemand die Antwort für eine Regelauskunft hält.

**Die deterministische Formatprüfung bleibt ein eigener Weg.** Was ein Format
erlaubt, weiß dieses Programm aus `domain.Fragen` und fragt kein Modell. Eine
Einschätzung, die davon spräche, spräche über etwas, das schon feststeht.

## Warum es eine zweite Prüfung nicht gibt

**Weil sie eine wäre, die dieses Programm nicht führen kann.** Geprüft wird,
**ob** ein Beleg genannt wurde und ob es ihn gibt — nicht, ob er die Aussage
trägt. Der Satz steht seit A16 an [domain.PruefeBelege]:

> ID-Gültigkeit beweist noch keine inhaltliche Folgerichtigkeit.

Er bleibt stehen. Wer ihn für eine Zusage über den Inhalt hielte, hielte eine
Fußnote für ein Argument.

## Der Befund, der zu diesem Werkstück geführt hat

**`domain.PruefeBelege` hatte keinen Produktivaufrufer.** A16 hat die Prüfung
gebaut, mit Tests und voller Abdeckung — und niemand rief sie. Das ist die
erste Dimension aus AGENTS §10, und sie steht dort zuerst, weil sie schon
einmal gefehlt hat (`internal/svg` in Phase P). A13.2a ist ihr erster
Aufrufer; das ist keine Nebenbemerkung, sondern der Grund für den Zuschnitt
dieses Werkstücks.

## Was das kostet

**Eine unbelegte Antwort steht am Fenster** — mit einer roten Zeile darüber,
aber sie steht da. Wer sie überliest, liest eine Modellmeinung als Auskunft.
Das ist der Preis dafür, sie nicht zu verstümmeln, und er ist ausgesprochen.

**Und die Belege sind nur so gut wie der Index.** Ist keiner gebaut, gibt es
keine Belege — dann sagt die Recherche warum, und die Einschätzung entsteht
gar nicht erst. Ein Modell ohne Belege zu fragen wäre die stillste Art, aus
einer Meinung eine Quelle zu machen.

## Verworfen

**„Nur belegte Antworten zeigen."** Es klingt streng und ist es nicht: das
Modell lernte, weniger zu nennen. Eine Antwort ohne Fußnoten wäre dann die
sichere — und die schlechteste.

**„Die Befunde in eine Protokollzeile."** Dann sieht niemand sie, der die
Antwort liest. Ein Befund, der nicht dort steht, wo die Aussage steht, ist
eine Aufzeichnung und keine Warnung.
