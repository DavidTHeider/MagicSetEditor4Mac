# 0018 — Ein Set aus einem Archiv wird nicht zurückgeschrieben

**Stand:** angenommen
**Datum:** 2026-08-15

## Lage

Ein `.mse-set` darf ein **Verzeichnis** oder eine **ZIP-Datei** sein.
`msepaket.Oeffne` liest beides seit Phase 2, und `docs/FORMATE.md` führt es.

`Setschreiber.Schreibe` nimmt ohne Ziel die Herkunft — das ist ⌘S. Bei einem
Archiv ist die Herkunft eine **Datei**, und `os.MkdirAll` traf dort auf eine
reguläre Datei: „not a directory". Wahr, und niemandem gesagt, was los ist.

Gefunden vom Zweitblick durch Codex über Phase 6
(#208).

## Entscheidung

**⌘S auf ein Set aus einem Archiv wird abgewiesen**, mit einem Satz, der das
Archiv nennt und auf „Sichern unter" zeigt.

## Warum nicht die beiden anderen Wege

**Als Archiv zurückschreiben.** Dann bräuchte `msepaket` einen ZIP-Schreiber —
und die Byte-Zusage aus #48
gälte nur noch für die **Hauptdatei**, nicht für das, was drumherum liegt.

Das ist der Kern: Kompression ist nicht eindeutig. Dieselben Bytes, durch einen
anderen Deflate-Encoder geschickt, ergeben eine andere Datei — andere
Blockgrenzen, andere Zeitstempel im Eintrag, andere Reihenfolge. Ein Rundlauf
„öffnen, sichern, Byte für Byte vergleichen" wäre für ein Archiv nicht mehr
herstellbar, und
[0011](0011-die-quelldatei-wird-mitgefuehrt.md) ist der Vermerk, der die
Byte-Zusage trägt — und sie ist die Zusage, an der dieses Projekt hängt.
Sie für ein Format aufzuweichen, in dem die meisten Sets **nicht** vorliegen,
ist ein schlechter Tausch.

**Als Verzeichnis daneben legen.** Dann läge nach ⌘S plötzlich ein zweites Set
da, das niemand angeordnet hat — und beim nächsten Öffnen wäre nicht mehr klar,
welches das gemeinte ist. Ein Programm, das aus „sichern" ein „kopieren" macht,
ist an der Stelle unbrauchbar, an der es darauf ankommt.

## Was das kostet, und der Preis gehört ausgesprochen

**Wer ein Set aus einem Archiv bearbeitet, kann es nicht dorthin
zurückgeben.** Es geht als Verzeichnis daneben hinaus — ein Schritt mehr, und
der Ort ist ein anderer.

**Hier stand, „Sichern unter" wechsle den Arbeitsort nicht**, und das war ein
Loch, kein Preis: das folgende ⌘S ging wieder gegen das Archiv und wurde wieder
abgewiesen. **Der Ausweg, den die Meldung nennt, führte nicht heraus.**
Angestrichen hat es der dritte Durchgang über Phase 6
(#226).

**Seit #226 wechselt er.** Nach jedem geglückten Schreiben an einen **anderen**
Ort führt der Bestand den neuen; das nächste ⌘S geht dorthin. Die Klasse war
größer als der Archivfall: bei einem Verzeichnis fiel es nur nicht auf, weil das
alte Ziel schreibbar bleibt — und dann schrieb ⌘S still an eine Stelle, die
niemand mehr meint.

**Ein Feld und keine Transaktion**, und der Grund gehört dazu: das Speichern
ändert die **Quelle nicht**. Sie steht seit dem Öffnen im Bestand
([0011](0011-die-quelldatei-wird-mitgefuehrt.md)), und was hinausgeht, entsteht
aus ihr. Auseinanderlaufen können Herkunft und Quelle deshalb nicht.

**„Sichern unter" steht in derselben Zeile der Setliste und in derselben
Menüleiste**; die Meldung nennt ihn.

**Nicht entschieden ist, ob das so bleibt.** Sobald jemand ein Set aus einem
Archiv wirklich bearbeitet und die Meldung ihn stört, ist die Frage neu zu
stellen — dann aber mit einem Fall, nicht mit einem Verdacht.

## Wer das ändern will

ändert zuerst diesen Vermerk, und danach
[0011](0011-die-quelldatei-wird-mitgefuehrt.md): ein ZIP-Schreiber ohne eine
Antwort auf die Byte-Zusage wäre eine Zusage, die zweierlei bedeutet.

---

## Nachtrag vom 2026-08-21: der Fall ist da, und die Antwort bleibt

„Nicht entschieden ist, ob das so bleibt. Sobald jemand ein Set aus einem
Archiv wirklich bearbeitet und die Meldung ihn stört, ist die Frage neu zu
stellen — dann aber mit einem Fall, nicht mit einem Verdacht."

**Der Fall ist gekommen**: seit #702
liefert das Bündel ein Beispielset mit, und die kanonischen Sets sind Archive.

**Die Antwort ist nicht der ZIP-Schreiber, sondern das Auspacken.**
`scripts/mitliefern.sh` legt das Beispielset **ausgepackt** ins Bündel; damit
ist es ein Verzeichnis, ⌘S trägt, und die Byte-Zusage aus
[0011](0011-die-quelldatei-wird-mitgefuehrt.md) bleibt, was sie ist. Gemessen
kostet es nichts: 8,6 MB gepackt gegen 9,0 MB ausgepackt — die Kartenbilder
sind bereits komprimiert.

**Und die Meldung kommt jetzt an.** Sie stand hier zugesagt („mit einem Satz,
der das Archiv nennt und auf ‚Sichern unter' zeigt") und erreichte niemanden:
`app.Uebersetze` bildet ein blankes `ErrValidation` auf den Rückfallsatz ab,
und der lautet „Es ist ein unerwarteter Fehler aufgetreten."
(#704).
