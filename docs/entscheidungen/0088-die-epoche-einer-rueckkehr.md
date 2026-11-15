# 0088 — Die Epoche einer Rückkehr, und warum sie nichts bucht

**Stand:** angenommen · **Datum:** 2026-09-10 · **Werkstück:** O7 von Phase 5.2
(#1431)

## Der Anlass

> Ein historisches Backup kann bereits abgegebene Karten enthalten.
> Versionskompatibilität allein erlaubt keinen aktiven Handelsbestand.

Der Satz steht im Issue, und er beschreibt einen Schaden, der keinen Fehler
braucht: eine Sicherung von vor drei Wochen wird zurückgespielt, und der
Bestand ist wieder der von damals — **samt der Exemplare, die seither aus der
Sammlung hinausgegangen sind**. Sie stehen wieder da, und niemand weiß, ob sie
noch jemandem gehören.

Verlorene Karten kann man neu eintragen. Eine zweimal verkaufte nicht.

## Die Entscheidung

**Jede Rückkehr, die etwas offenlässt, hinterlässt eine Epoche im Bestand** —
wann zurückgegangen wurde, auf welchen Stand, aus welcher Datei, und welche
Mengen sie betrifft.

### Drei Antworten, und die dritte ist keine Notlösung

„Ist da" und „ist weg" reichen nicht. Die richtige Antwort ist **unklar** —
dieselbe Gestalt wie `ungeklaert` beim Verträglichkeitsurteil
([0087](0087-eine-kennung-je-vorgang.md) ist die Nachbarentscheidung desselben
Werkstücks; das Urteil selbst steht in O6, #1430).

Daraus folgt zweierlei:

1. **Unklar heißt nicht handelbar.** Bis jemand nachgesehen hat, ist die
   betroffene Menge aus dem Verkehr — `domain.Handelssperre` ist die eine
   Stelle, an der die Frage „darf das raus?" beantwortet wird.
2. **Unklar heißt ausdrücklich auch nicht ausgebucht.** Eine Anwendung, die
   nach einer Rückkehr selbsttätig abzöge, machte aus einer Unsicherheit eine
   Behauptung — und zwar eine, die niemand mehr rückgängig macht. Das Issue
   sagt es: *keine automatische Doppelbuchung.*

### Was die Epoche betrifft, ist gerechnet und nicht geschätzt

**Der Vergleich läuft vor dem Tausch**, weil es danach nur noch **eine** Seite
gibt. Verglichen wird die Sammlung Gegenstand für Gegenstand — Druck, Zustand,
Sprache, Ausführung, je Profil. Wo die Sicherung **mehr** hat als der laufende
Bestand, ist der Überschuss unklar.

**Und die andere Richtung ist der Datenverlust**: was seit der Sicherung
dazukam, verschwindet. Beides steht vor der Bestätigung; die Abnahme verlangt
„genau den wiederhergestellten Datenzeitpunkt und noch nicht übertragene
Änderungen".

### Eine ungelesene Sammlung ist nicht dasselbe wie eine leere

Eine sehr alte Sicherung kennt die Spalten nicht, die der Vergleich braucht.
Sie deswegen als unbedenklich zu führen wäre die gefährlichste Richtung: „keine
unklaren Mengen" ist die eine Auskunft, die niemals falsch sein darf.

Deshalb trägt jeder gelesene Stand ein `SammlungGelesen`, und wo eine der
beiden Seiten schweigt, steht `HandpruefungNoetig` — die Epoche entsteht dann
**ohne** Mengen und sagt genau das. Das Issue verlangt es: *bei fehlenden
Nachweisen Unsicherheit sichtbar und manuelle Bestandsprüfung erforderlich.*

### Sie verjährt nicht

Eine Epoche endet, wenn ein **Mensch** quittiert — und steht danach mit dem
Zeitpunkt weiter da. Sie verschwindet nicht mit der Zeit und nicht beim
nächsten Start; ein Nachweis, der verschwindet, ist keiner.

**Quittieren gibt es deshalb über MCP nicht.** Eine Quittung sagt „ich habe
nachgesehen", und nachsehen heißt hier: in die Kiste greifen und zählen. Ein
Sprachmodell kann das nicht, und ein Werkzeug, mit dem es trotzdem behaupten
könnte, es getan zu haben, machte aus der Unsicherheit eine Behauptung. An
Fenster, Kommandozeile und Dienst gibt es sie — dort sitzt ein Mensch davor.

### Und sie liegt im Bestand, nicht daneben

Eine Datei neben der Datenbank ginge beim nächsten Zurückspielen mit — und
dann wäre die Auskunft über die Rückkehr das Erste, was eine Rückkehr
verliert.

**Geschrieben wird sie deshalb ohne Profil im Kontext**, anders als jeder
andere Schreibweg des Bestands: ein Zurückspielen geschieht auch **beim
Start**, bevor jemand ein Profil gewählt hat — die Wiederherstellungsseite ist
genau dafür da. Ein Profilzwang hieße, dass ausgerechnet die Rückkehr aus einem
kaputten Bestand keinen Nachweis bekommt. **Gelesen** wird weiter nur mit
Profil: was *wem* fehlt, ist eine Frage an eine Sammlung, und die gehört einem.

## Was das kostet

**Es gibt heute keinen Handel, der die Sperre liest.** T2
(#1467) und
T6 (#1471)
bauen ihn in Phase 5.6. Was hier entsteht, ist der **Vertrag**, an dem sie
hängen — und ein Vertrag ohne Vertragspartner lässt sich nicht am Verhalten
prüfen, sondern nur an seiner Form.

**Das ist die Reichweite dieser Zusage, und sie gehört ausgesprochen** (§10):
geprüft ist, dass eine Rückkehr ihre Epoche hinterlässt, dass die Sperre daraus
die richtige Zahl rechnet und dass eine Quittung sie nimmt. **Nicht** geprüft
ist, dass ein Handel sie befragt — den gibt es nicht.

**Und der Weg, das zu halten, ist eine Aufzählung**: wer eine Menge als
verfügbar meldet, geht durch `domain.Handelssperre`. T2 trägt sich dort ein,
oder er meldet Bestände, über die nichts feststeht.

## Verworfen

**Automatisch ausbuchen.** Die Anwendung weiß, wie viele Stücke die Sicherung
mehr hat — sie könnte sie abziehen. Sie tut es nicht: die Differenz kann
**auch** daher rühren, dass jemand die Sammlung seither aufgeräumt hat. Aus
zwei möglichen Gründen den teureren zu wählen und still zu buchen wäre eine
Entscheidung, die niemand getroffen hat.

**Die Epoche nach einer Frist verfallen lassen.** Eine Unsicherheit, die von
selbst verschwindet, ist eine, die niemand geklärt hat.

**Die Sperre über den ganzen Bestand legen.** Nach einer Rückkehr wäre dann
gar nichts mehr handelbar. Das ist sicher und unbrauchbar — und es
unterscheidet die Fälle nicht mehr, um die es geht.
