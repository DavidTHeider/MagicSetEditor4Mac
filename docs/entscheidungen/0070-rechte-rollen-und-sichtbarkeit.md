# 0070 — Rechte, Rollen und Sichtbarkeit

> **In einem Punkt abgelöst am 2026-09-06 durch
> [0073](0073-ein-leser-darf-wirklich-lesen.md).** `leser` heißt seither „kommt
> hinein und liest", nicht mehr „sieht, dass es das Profil gibt".
>
> **Der Grund, den dieser Vermerk nennt, ist damit eingelöst und nicht
> verworfen**: ein Nur-Lese-Modus **innerhalb** eines Profils braucht einen
> lückenlosen Schreibschutz, und den gibt es nur im Treiber. Den gibt es
> jetzt — `PRAGMA query_only` auf einer zweiten Verbindung.
>
> **Alles andere hier gilt unverändert**, insbesondere die Sichtbarkeit
> `bestand`: sie zeigt den Namen in der Übersicht und öffnet den Inhalt
> **nicht**.

**Entschieden am 2026-09-06**, Werkstück 11 von Phase 4.12
(#1351).

## Der Anlass

Die Kommentierung verlangt „ein Rechte- und Rollenmodell über Nutzer hinweg",
und die Roadmap sagt dazu: **ein Modell und nicht ein Feld**.

**Was heute geht und nicht gehen sollte**: `Wechseln` fragt den zweiten Faktor
des **Eigners** des Zielprofils. Hat er keinen — und das ist der Normalfall —,
kommt jeder Nutzer dieses Bestands in jedes Profil. Die Trennung aus Wanderung
0035 hält die Daten auseinander; **wer hinein darf**, hat bisher niemand
gefragt.

## 1. Drei Rollen, aufsteigend

| Rolle | darf |
| --- | --- |
| `leser` | das Profil in der Übersicht **sehen** — Name, Beschreibung, Eigner |
| `mitarbeiter` | dazu **hineinwechseln** und darin arbeiten |
| `verwalter` | dazu Freigaben vergeben und zurückziehen |

**Aufsteigend und nicht als Menge von Einzelrechten.** Eine Aufzählung wollte
bei jedem neuen Weg gepflegt sein, und ein ungepflegter Eintrag in einer
Rechteliste ist gefährlicher als keiner — dieselbe Überlegung wie beim Scope
eines Zugriffstokens ([0068](0068-ein-token-ist-kein-kennwort.md)).

**Der Eigner ist Verwalter seines eigenen Profils, ohne Zeile im Bestand.** Eine
Freigabe an sich selbst wäre eine Zeile, die niemand löschen dürfte.

## 2. `leser` heißt weniger, als das Wort verspricht

**Ein `leser` sieht, dass es das Profil gibt. Er sieht nicht, was darin steht.**

**Warum nicht mehr**: ein Nur-Lese-Modus **innerhalb** eines Profils bräuchte
einen lückenlosen Schreibschutz. Den gibt es an genau einer Stelle — im Treiber
(`_pragma=query_only(1)`). Alles andere hieße, 116 schreibende Bindungen und
205 Rufe von `imProfil` einzeln zu prüfen; eine Zusage, die daran hängt, bricht
beim ersten vergessenen Ruf **still**, und ein Wächter, der sie halten wollte,
müsste jede neue Bindung kennen (§10).

**Was das kostet, gehört ausgesprochen**: „teile dieses Profil zum Nachlesen"
geht heute nicht. Wer es will, gibt `mitarbeiter` — und muss dem anderen
vertrauen.

**Wie es ginge**: beim Wechsel in ein fremdes Profil als `leser` eine **zweite
Verbindung** mit `query_only` nehmen und sie halten, solange der Zeiger dort
steht. Das ist ein eigenes Werkstück, und es hat einen eigenen Preis — zwei
Verbindungen, zwei Lebenszyklen, und eine Fehlermeldung aus SQLite, die
übersetzt sein will.

## 3. Die Sichtbarkeit steht am Profil und nicht in der Freigabe

`privat` oder `bestand`. **Weil sie eine Aussage über das Profil ist und nicht
über einen Nutzer**: „das darf jeder sehen" braucht keine Zeile je Nutzer, und
eine Freigabe für jeden wäre eine Liste, die mit jedem neuen Nutzer
nachgeführt sein wollte.

**`privat` ist die Vorgabe.** Ein Bestand, der diese Wanderung durchläuft,
ändert sein Verhalten nicht.

## 4. Durchgesetzt wird an genau zwei Stellen

1. **`Wechseln`** verlangt `mitarbeiter` — **oder** den zweiten Faktor des
   Eigners. Der nächste Abschnitt sagt, warum es ein Oder ist.
2. **Die Freigabe selbst** verlangt `verwalter`.

**Beide sind einzeln und beide gegenprobierbar** — und es bleibt keine Fläche
halb gesichert. Das ist der Grund, aus dem `leser` nicht mehr darf: eine dritte
Durchsetzungsstelle gäbe es nicht.

## 5. Die Freigabe ersetzt den zweiten Faktor, statt vor ihm zu stehen

**Und das ist die Korrektur an der ersten Fassung dieses Werkstücks.** Sie
fragte die Freigabe **unbedingt**, vor dem Faktor — und das hat jeden zweiten
Nutzer ausgesperrt.

**Warum**: wer arbeitet, ist der Eigner des gezeigten Profils (Abschnitt 6).
Eine unbedingte Frage nach der Freigabe heißt dann, dass Bewegung nur entlang
vergebener Kanten möglich ist, und die gibt es am Anfang nicht:

- ein frisch angelegter zweiter Nutzer war **von Geburt an unerreichbar** —
  niemand konnte in sein Profil, auch er selbst nicht, denn um für ihn zu
  sprechen, müsste man schon darin stehen;
- und wer sein eigenes Profil einmal verließ, kam **nicht zurück**.

**Nachgemessen und nicht überlegt**: am 2026-09-06 wurden davon **sieben**
Läufe rot — vier Flächenläufe über den Profilwechsel und drei im Dienst —, und
keiner an einem Fehler des Bestands. Angestrichen hat es der Flächenlauf zu
diesem Werkstück selbst, beim ersten Versuch, ihn zu schreiben.

**Die Regel lautet jetzt**: ein Wechsel glückt, wer sich als Eigner
**ausweist** — oder eine Freigabe **hat**.

Das ist dieselbe Zusage, die der zweite Faktor seit Werkstück 7 macht: **wer
keinen hat, wird nicht gefragt.** Ein Profil ohne Schloss wird nicht bewacht.
Und wo eines steht, ist die Freigabe genau das, was sie sein soll — der Weg
hinein, ohne dass der Eigner seinen Code weitergeben muss.

**Was das kostet, gehört ausgesprochen** (§10): in einem Bestand, in dem
niemand einen zweiten Faktor eingerichtet hat, verhindert die Freigabe
**nichts**. Sie ist dort eine Auskunft darüber, wer wen hereinlässt, und keine
Sperre. Das ist dieselbe Grenze, die der zweite Faktor selbst hat, und dieselbe,
die [0066](0066-nutzer-profile-und-die-vier-entscheidungen.md) für die
Profiltrennung nennt: die Zusage über die **Datei** leistet allein die
Verschlüsselung.

**Und der Fehler, der zurückkommt, ist der des Faktors.** „Frag nach dem Code"
sagt dem heimkehrenden Eigner, was zu tun ist; „dir fehlt die Rolle" schickte
ihn zu einem Verwalter, der er selbst ist.

## 6. Wer „dieser Mensch" ist

**Der Eigner des gezeigten Profils.** Diese Anwendung hat keinen Begriff einer
Sitzung neben dem Profilzeiger, und einen zu erfinden hieße, zwei Wahrheiten
darüber zu führen, wer gerade arbeitet.

**Was daraus folgt und gesagt gehört**: wer im Profil eines anderen arbeitet,
**ist** für die Dauer dieser Mensch. Eine Herkunftsangabe an dem, was dabei
entsteht, ist Werkstück 12.

## Was verworfen wurde

**Rechte je Datensatz.** „Dieses eine Deck darfst du sehen" wäre eine zweite
Trennung neben der Profiltrennung, und die eine könnte die andere nicht mehr
halten.

**Eigentumsübergang.** Ein Profil gehört seinem Nutzer; wer es abgeben will,
legt ein neues an. Ein Übergang hieße, dass jede Freigabe daran ihre Bedeutung
ändert, ohne dass jemand sie angefasst hat.

**Eine Freigabe, die ein Löschen überlebt.** Anders als bei den Profilvorgaben
([0069](0069-eine-vorgabe-die-auf-eine-zeile-zeigt.md)) steht hier `ON DELETE
CASCADE`: eine Vorgabe, die ins Leere zeigt, ist eine Auskunft („das hattest du
gewählt"); eine Freigabe ohne Profil oder ohne Nutzer ist Müll — und sie wieder
gültig werden zu lassen, sobald jemand dieselbe Kennung bekommt, wäre eine
Rechteerweiterung durch Zufall.
