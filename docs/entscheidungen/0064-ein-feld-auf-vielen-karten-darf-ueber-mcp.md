# 0064 — Ein Feld auf vielen Karten darf über MCP

**Stand:** angenommen
**Datum:** 2026-09-02
**Schreibt fort:** [0040](0040-der-lebenszyklus-einer-karte-gehoert-dazu.md),
und damit die zweite Hälfte von [0010](0010-mcp-als-eigener-prozess.md)
**Anlass:** #1390,
gefunden bei der Prüfung der 24 Flächenlücken (Werkstück E von
#1324)

## Lage

[0010](0010-mcp-als-eigener-prozess.md) hat den Umfang der MCP-Werkzeuge
beschränkt, und [0040](0040-der-lebenszyklus-einer-karte-gehoert-dazu.md) hat
den einen Satz davon ausdrücklich stehen lassen:

> **Die andere Hälfte von 0010 bleibt:** kein Werkzeug überschreibt ein Set,
> und keines ändert mehrere Karten auf einmal. Der `Massenschreiber` des
> Bestands bleibt außen vor.

Ein Wächter hält das — `TestJedesSchreibendeWerkzeugStehtInEinerZusage`.

**Und `docs/FAEHIGKEITEN.md` hat die Fähigkeit dazu falsch beschrieben.** Dort
stand sie als *Viele Felder setzen*, begründet damit, das Fenster schreibe „die
Eingaben **einer** Karte in einem Zug, weil eine Maske viele Felder hat". Die
Bindung heißt `FelderSetzen(setID, karteIDs, feld, wert)` und setzt **ein**
Feld auf **vielen** Karten — sie ist die Massenänderung, nicht ein Formular.
Die Begründung beschrieb eine andere Operation und verdeckte, dass hier eine
Entscheidung zu treffen ist.

## Entscheidung

**`schreiben.felder` kommt dazu**: ein Feld, mehrere benannte Karten, ein Set,
eine Transaktion.

**Was von 0010 bleibt:** kein Werkzeug überschreibt ein Set. Dieses hier tut es
auch nicht — es kann keine Karte anfassen, die der Aufrufer nicht nennt.

## Warum der Grund von 0010 hier nicht trägt

0010 nennt zwei Maßstäbe, und beide sind hier nachgesehen worden.

### Der Maßstab „ein ganzes Set gegen ein Feld"

Zwischen beidem liegt die **Karte**, und `AGENTS.md` §2 sagt, wofür es dieses
Paket gibt — „Werkzeuge, mit denen ein Sprachmodell **Karten** liest und
ändert".

**Der Umfang dieses Werkzeugs steht in seinen Argumenten.** Es ändert genau die
Karten, die der Aufrufer aufzählt, und genau **ein** Feld an ihnen. Wer zwei
nennt, ändert zwei. Ein Set hat es damit nicht in der Hand: eine Karte, die
nicht in der Liste steht, bleibt unberührt, und eine aus einem **anderen** Set
lässt den ganzen Aufruf scheitern — die Aufsicht fragt für jede einzeln
([0041](0041-die-setgrenze-liegt-im-dienst.md)).

### Der Maßstab Umkehrbarkeit — und was er hier wirklich hergibt

**Hier stand, er sei »besser erfüllt« als bei `schreiben.feld`. Das war
falsch, und der Zweitblick über die Phasen 4.0 bis 4.9 hat es angestrichen
(Befund 3, am 2026-09-03).**

Der Satz lautete: der Speicher legt die Massenänderung als **eine Gruppe im
Verlauf** ab (`internal/storage/massenschreiber.go`, `neueGruppe`, seit
#52), also
mache ein einziges `schreiben.zuruecknehmen` sie ganz rückgängig. Das stimmt
— **solange niemand dazwischenschreibt**.

**Und dazwischenschreiben ist ausdrücklich vorgesehen.**
[0010](0010-mcp-als-eigener-prozess.md) beschreibt mehrere schreibende
Prozesse nebeneinander. Die Rücknahme bekommt nur eine **Setkennung** und
nimmt den jüngsten geltenden Schritt dieses Sets
(`internal/storage/verlaufbewegen.go`, `naechsterSchritt`); eine
Vorgangskennung gibt es nicht. Schreibt das Fenster, die Kommandozeile oder
ein zweiter MCP-Prozess nach der Massenänderung in dasselbe Set, nimmt der
Aufruf **zuerst jene** Änderung zurück.

**Damit ist die Umkehrbarkeit hier nicht besser als bei `schreiben.feld`,
sondern schwächer**: dort hält der Aufrufer den vorigen Wert in der Hand und
ist von keinem gemeinsamen Stapel abhängig.

### Worauf diese Entscheidung deshalb steht

**Auf dem Umfang, nicht auf der Rücknahme.** Der Umfang steht in den
Argumenten des Aufrufs und nirgends sonst:

- Es wird **ein** Feld geändert.
- An **genau den Karten**, die der Aufrufer aufzählt.
- In **einem** Set — eine Karte aus einem anderen lässt den ganzen Aufruf
  scheitern ([0041](0041-die-setgrenze-liegt-im-dienst.md)).
- In **einer** Transaktion: alles oder nichts.

Das ist die Hälfte von 0010, die trägt, und sie trägt allein. Ein Set hat
dieses Werkzeug nicht in der Hand.

**Die Rücknahme bleibt eine Bequemlichkeit mit Vorbehalt**, und der Vorbehalt
steht in der Erklärung des Werkzeugs: wer sicher zurückwill, holt die Werte
vorher mit `lesen.karten`.

### Nachtrag vom 2026-09-04: der Vorbehalt ist fort

**Seit #1393
steht diese Entscheidung wieder auch auf der Rücknahme, und diesmal zu
Recht.** Die Massenänderung gibt an jeder Fläche einen **Vorgang** zurück, und
`schreiben.zuruecknehmen` nimmt ihn entgegen — dann geht genau diese Änderung
zurück, gleich wer inzwischen in dasselbe Set geschrieben hat.

**Die Kennung war die ganze Zeit da.** `neueGruppe` schreibt sie seit
#52 in jeden
Verlaufseintrag der Massenänderung; sie wurde nur nie zurückgegeben. Der
Aufwand lag im Durchreichen durch `service`, `app` und die drei Flächen, nicht
im Speicher.

**Von den drei Wegen, die das Issue nannte, ist es der dritte** — und der
Grund dafür ist §10: die ersten beiden hätten die Massenänderung geheilt, der
dritte heilt **jede** Änderung. `schrittkennung` gibt einem Eintrag ohne
Gruppe die Kennung `einzeln:<id>`; wer sie hat, nimmt auch eine einzelne
Änderung gezielt zurück.

**Am Fenster ändert sich nichts, und das ist Absicht.** ⌘Z heißt „das da
eben", und ein Fenster hat den Stand vor Augen; `App.Zuruecknehmen` gibt
deshalb weiter keinen Vorgang mit. Die gezielte Rücknahme ist die Antwort auf
ein Problem, das nur ohne Fenster besteht.

## Was das kostet, und es gehört ausgesprochen

**Ein Modell kann in einem Aufruf mehr anrichten als vorher.** Dreihundert
Karten in einem Zug sind dreihundert Karten. Was dagegen steht, ist nicht die
Zahl, sondern dass der Zug **begrenzt** ist — und dass die Alternative
schlechter wäre: dreihundert Einzelaufrufe, von denen die Hälfte durchgeht,
wenn einer scheitert.

**Und die vorigen Werte sind weiter nicht in der Antwort** — jetzt als
Entscheidung und nicht mehr als offene Frage. Wer sie einspielen wollte,
bräuchte je Karte einen Aufruf und hätte damit genau die halbe Änderung, die
dieses Werkzeug verhindert (siehe *Verworfen*). Was stattdessen zurückkommt,
ist der **Vorgang**, und der nimmt alles auf einmal zurück.

**Der Speicher liest sie ohnehin**, `massenzeile` holt den vorigen Wert für
den Verlaufseintrag — und dort bleibt er. Er ist der Rückweg, den die
Rücknahme geht; ein zweites Mal durch die Antwort zu reisen, hilft niemandem.

## Verworfen

**Die Ablehnung erneuern und MCP außen vor lassen.** Das wäre nach §16 zulässig
gewesen, aber die Begründung hätte den **Weg** betroffen und nicht die
Fähigkeit: „ein Sprachmodell soll das nicht können" ist eine Aussage über den
Aufrufer, und die trägt nicht, wenn dasselbe Modell dreihundert Einzelaufrufe
machen darf. Genau diese Unterscheidung ist der Gegenstand von Werkstück E.

**Die vorigen Werte zurückgeben.** Sie stünden als Liste in der Antwort und
wären in der Praxis unbrauchbar: wer sie einspielen will, braucht je Karte
einen Aufruf — und damit genau die halbe Änderung, die dieses Werkzeug
verhindert.
