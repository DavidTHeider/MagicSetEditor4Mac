# 0090 — Zurücksetzen ist eine Familie, kein Knopf

**Stand:** entschieden · **Datum:** 2026-09-10 · **Werkstück:** O4 von Phase 5.2
(#1428)

## Worum es geht

Das Issue verlangt: „Zurücksetzen benennt genau, welche Einstellungen oder Daten
betroffen sind", und in der Umsetzung „getrennte Aktionen Layout,
Profileinstellungen, Nutzereinstellungen, Profil löschen, Nutzer löschen, ganze
Anwendung".

## Die Entscheidung

**Sechs Arten, in der Reihenfolge ihrer Schwere, und jede mit eigener Vorschau.**

Der naheliegende Weg wäre ein Knopf „Zurücksetzen" mit einer Rückfrage gewesen.
Er ist die Gestalt, in der jemand das eine drückt und das andere bekommt: das
Fenster wieder aufzuräumen und einen Menschen samt allem zu löschen haben nichts
miteinander zu tun außer dem Wort.

**Die Vorschau zählt und benennt.** „Dieses Profil und alles, was dazugehört"
ist keine Auskunft, mit der jemand entscheidet — „drei Sets, 107 Karten, zwei
Token, ein Abzeichen" ist eine.

## Vier Entscheidungen darin, die sich später hinterfragen lassen

### 1. Zurückgesetzt heißt gelöscht

Eine Profileinstellung, die es **nicht gibt**, erbt von der Nutzereinstellung.
Genau das verlangt die Abnahme („Profilreset stellt Vererbung her"), und ein
geschriebenes `erben` wäre eine Wahl, die wie keine aussieht.

**Der Preis**: es gibt keinen Unterschied mehr zwischen „nie gewählt" und
„zurückgesetzt". Wer ihn brauchte, bräuchte eine dritte Zeile im Bestand.

### 2. Der Neuanfang tauscht, statt zu löschen

`RuecksetzungAnwendung` legt die Bestandsdatei nach `bestand-vor-rueckspielung.db`
beiseite; der nächste Start legt einen frischen an.

**Verworfen**: eine Schleife über `DELETE FROM …` je Tabelle. Sie wäre
unumkehrbar **und** unvollständig — sie müsste jede Tabelle aufzählen, und die
nächste Wanderung fügte eine hinzu, die niemand nachträgt. Der Tausch braucht
keine Aufzählung und ist rückgängig zu machen, solange die Datei daliegt.

**Der Preis**: die Datei bleibt liegen und belegt Platz. Sie steht deshalb in
der Platzübersicht, und die Bilanz nennt ihren Ort.

### 3. Der Schlüsselbund wird nicht mit ausgeräumt

Das Issue sagt „Schlüsselbundbereinigung separat und begrenzt". Hier heißt das:
der Neuanfang fasst ihn **gar nicht** an.

**Der Grund ist der Tausch aus (2)**: die Einträge gehören zu dem Bestand, der
beiseiteliegt, und werden gebraucht, falls jemand ihn zurückholt. Wer ein
Geheimnis loswerden will, löscht **seinen Nutzer** — dann geht es mit, und nur
seines. Das ist die „begrenzte" Bereinigung, und sie stand schon vorher da.

**Und es ist zugleich die Zusage** „unbekannte Dateien und Daten anderer
Programme bleiben unangetastet": der Schlüsselbund gehört dem Betriebssystem.

### 4. Ein laufender Auftrag sperrt, statt gestoppt zu werden

Das Issue sagt „offene Peer-Sitzungen, Handelsreservierungen und Jobs vor
Löschung geordnet stoppen". Von den dreien gibt es heute nur die Jobs
(`service.Auftragsbuch`, U11), und **das Buch kennt keinen Abbruch**.

Die ehrliche Umsetzung ist deshalb ein **Hindernis**: es wird nicht mitten
hineingelöscht, und die Vorschau **nennt**, worauf zu warten ist. Einem
laufenden Schreiber den Bestand wegzunehmen wäre kein „geordnetes Stoppen",
sondern das Gegenteil.

**Die Reichweite gehört dazu** (§10): das Buch steht im Arbeitsspeicher **je
Prozess**. Was in einem zweiten Fenster läuft, sieht dieses nicht.

**Peer-Sitzungen und Handelsreservierungen gibt es noch nicht** — T2, T5 und T6
sind ungebaut. Ihre Stelle ist vorbereitet: `RuecksetzungOffeneRueckkehr` sperrt
schon heute, wenn über Mengen eines Profils nach einer Rückkehr nichts feststeht
([0088](0088-die-epoche-einer-rueckkehr.md)), und der belegwahrende Weg heraus
ist die Quittung — genau die Gestalt, die das Issue verlangt.

## Und der Nachweis, den O1 gebaut hat

`RuecksetzungAnwendung` verlangt einen [Nachweis](0078-der-wirksame-akteur.md).
Eine Rolle reicht dafür nicht: sie nähme **anderen Eigentümern** ihre Daten weg,
und das Issue sagt es ausdrücklich („Kein Profilverwalter erhält durch
umfassenden Reset Rechte über andere Eigentümer").

**Damit hat `Autorisierung.PruefeMitNachweis` seinen ersten Aufrufer.** Bis
hierher stand der Weg da, und niemand ging ihn — das ist die erste Dimension
aus §10, und sie hat hier zum zweiten Mal gegriffen.

## Was das kostet

**Sechs Arten sind sechs Wege**, und jeder braucht seine Vorschau, seine
Hindernisse und seinen Test. Ein Knopf wäre in einem Zehntel der Zeit fertig
gewesen — und hätte die Frage „was verliere ich?" unbeantwortet gelassen.

## Nachtrag vom 2026-09-11: die Revision ist Pflicht, und sie bindet im Geschäft

**Befund 9 des Nachlaufs zum Zweitblick über Phase 5.2.** Die Bindung an die
Vorschau kam mit Befund 12 des Zweitblicks dazu — und war dreifach löchrig:

1. **Die Marke las nur bis zur Zahl.** `Ruecksetzungsmarke` hashte Art und
   `Was=Anzahl`; die **Namen**, die die Vorschau nennt, gingen nicht ein. Ein
   Set »Alt« in der Vorschau und ein gleich großes Set »Neu« bei der Ausführung
   ergaben dieselbe Marke. Sie trägt die Namen jetzt mit — **sortiert**, denn
   eine Reihenfolge aus dem Bestand ist keine Aussage.
2. **Eine leere Revision hob die Bindung auf.** Hier stand „leer heißt: keine
   Vorschau gesehen", und das war als Weg für Aufrufer ohne Vorschau gedacht.
   Für die drei Arten, die **Daten wegnehmen**, ist das keine Wahl, sondern
   eine Tür neben der Zusage. Sie ist dort seither Pflicht; für Layout, Profil-
   und Nutzereinstellungen bleibt sie wahlfrei, weil dort nichts verschwindet,
   was zwischen Hinsehen und Klicken dazukommen könnte.
3. **Der Vergleich lag vor dem Schreiben, nicht darin.** Zwischen dem Zählen in
   `service` und dem Löschen in `storage` liegt eine Lücke; wer sie trifft,
   verliert etwas, das in keiner Vorschau stand.

**Die dritte ist die, die eine Naht gekostet hat.** `ProfilLoeschen` und
`NutzerLoeschen` nehmen seither eine **Bindung** entgegen — wahlfrei, weil die
vier anderen Flächen einen benannten Datensatz ohne Vorschau löschen. Sie läuft
**in** der Löschtransaktion: dort wird der Umfang je betroffenem Profil noch
einmal gezählt, und `service` entscheidet, ob das noch der Stand ist, den
jemand gesehen hat. Es ist dieselbe Bewegung wie bei Befund 11 des Zweitblicks,
wo die Revision der Einstellungen aus demselben Grund ins Geschäft gezogen
wurde.

**Zusammengezählt wird dabei in `service` und nicht im Bestand** (§2): was eine
Gruppe ist und wie zwei sich addieren, ist eine fachliche Regel. Die Bindung
bekommt die Umfänge **je Profil** und legt sie über **dieselbe** Funktion
zusammen, die auch die Vorschau benutzt — zwei Fassungen davon liefen
auseinander, sobald eine Gruppe dazukäme, und die Bindung wiese dann **jede**
Löschung ab.

### Was der Neuanfang nicht bekommt

`RuecksetzungAnwendung` legt den Bestand **beiseite** und öffnet einen frischen.
Dort gibt es kein Geschäft, in das eine Bindung passte — und es muss auch
keines geben: es wird nichts gelöscht. Der alte Bestand liegt danach als Datei
da, und die Bilanz sagt, wo. Die Revision ist trotzdem Pflicht, weil die
Vorschau zählt, was mitgeht.
