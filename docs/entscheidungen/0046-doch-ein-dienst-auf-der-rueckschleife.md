# 0046 — Doch ein Dienst auf der Rückschleife

**Stand:** angenommen · **Datum:** 2026-08-23
**Schreibt fort:** [0039](0039-kein-dienst-auf-der-rueckschleife.md) — und
kehrt ihn um
**Berührt:** [0010](0010-mcp-als-eigener-prozess.md),
[0043](0043-neun-phasen-bis-0-2-0.md),
[0045](0045-zwei-programme-an-einem-bestand.md)

## Warum es diesen Vermerk gibt

[0039](0039-kein-dienst-auf-der-rueckschleife.md) hat REST am 2026-08-22
abgelehnt und den Fall, der jetzt eingetreten ist, selbst vorgesehen:

> Käme ein benannter Abnehmer, dem stdio nachweislich nicht reicht […] wäre die
> Rechnung neu aufzumachen. **Dieser Vermerk ist dann fortzuschreiben, nicht zu
> umgehen.**

Das ist der erste Kasten von Phase U
(#887), und er
steht dort **vor der ersten Zeile Code**.

## Die Entscheidung

**Es kommt ein REST-Dienst auf `127.0.0.1`.** Er startet nicht von selbst, er
bindet nirgendwo sonst hin, und er bringt keine eigene Fachlogik mit.

## Wer der benannte Abnehmer ist — und wer er nicht ist

**Er ist der Zielsatz für 0.2.0.** Der Auftraggeber hat REST am 2026-08-23
ausdrücklich verlangt, als eine von vier Flächen, deren Parität und Stabilität
das Ziel dieser Iteration ist. Die Entscheidung ist in
[0043](0043-neun-phasen-bis-0-2-0.md) festgehalten, samt der Feststellung, dass
sie **gegen** die Empfehlung des unabhängigen Zweitblicks und gegen die dieses
Arbeitszweigs gefallen ist.

**Das ist ein anderer Abnehmer, als 0039 sich vorgestellt hat**, und das gehört
gesagt. Dort stand „ein Werkzeug, das den Unterprozess nicht starten kann" —
also ein technischer Zwang. Ein solcher ist **nicht** aufgetreten und wird hier
auch nicht behauptet. Was aufgetreten ist, ist eine Anforderung an das Produkt.

Eine Anforderung ist ein gültiger Grund. Eine erfundene technische Not wäre
keiner — §10: *auch eine Ablehnung braucht Deckung*, und das gilt für ein Ja
genauso.

## Was sich seit dem 2026-08-22 wirklich geändert hat

**Punkt 3 von 0039 ist eingelöst**, und er war das stärkste Argument: *die
Lücke liegt im Umfang, nicht im Protokoll — ein REST-Dienst neben einem
unvollständigen MCP wäre eine zweite unvollständige Fläche.*

| | am 2026-08-22 | am 2026-08-23, nach Phase T |
| --- | --- | --- |
| MCP-Werkzeuge | 5, davon 1 schreibend | **12**, davon **8** schreibend |
| Kommandozeile | — | **7 Befehle** |
| Bindungen am Fenster | 61 | 67 |
| Fachliche Fähigkeiten, benannt | keine Liste | **37**, mit Wächter |

Ein REST-Dienst tritt also **nicht** mehr neben eine Fläche, die ein Feld
setzen kann und sonst nichts. Er tritt neben zwei Flächen, die denselben
Arbeitsablauf tragen und deren Lücke gezählt ist.

**Die anderen drei Punkte von 0039 gelten unverändert.** Punkt 1 ist durch die
Anforderung beantwortet, nicht widerlegt. Punkt 2 — 0010s Rechnung über
Herkunftsprüfung, Marken und Rückschleife — steht: sie sind jetzt zu bauen.
Punkt 4 ist der Preis, und er steht unten.

## Der Preis, und er ist dauerhaft

**Zwei Zusagen der Roadmap fallen.** „Kein Netzwerkdienst" wird berichtigt; die
Zeile darf nicht stehen bleiben, während das Gegenteil gebaut wird.

**Ein Wächter wird gelockert.** `internal/schichten/netz_test.go` lässt heute
`net/http` genau ein Paket erreichen. Der zweite Eintrag steht künftig mit
seinem Grund daneben — und er bekommt einen **eigenen** Wächter für das, was
die Lockerung nicht aufgibt: die Bindung an `127.0.0.1`.

**Je Endpunkt vier Testklassen, bei jeder künftigen Änderung.** Gültig, Grenze
überschritten, fremde Herkunft, fehlende oder abgelaufene Marke. Das ist nicht
die Arbeit der ersten Fassung, sondern die jeder folgenden.

**Und ein fünfter Prozess an derselben Datei.** Was dabei gilt, ist in
[0045](0045-zwei-programme-an-einem-bestand.md) gemessen; der Dienst ändert
daran nichts, außer dass es ihn häufiger trifft.

## Was die Marke gewährt — und was sie nicht begrenzt

**Der Zweitblick hat gefragt, und die Antwort gehört hierher**, weil sie eine
Entscheidung ist und keine Lücke.

**Wer die Marke hat, hat die Rechte des Prozesses.** `POST /sets`,
`POST /sets/{id}/sicherung` und `POST /karten/{id}/bild` nehmen einen **Pfad**
und arbeiten damit überall, wo der Dienst darf. Es gibt kein Basisverzeichnis
und keine Pfadfreigabe.

**Das ist Absicht**, und zwar aus §16: *dieselbe Fähigkeit, überall dieselbe
Antwort.* `mse sichern` nimmt jeden Pfad, das Fenster nimmt jeden Pfad — ein
Dienst, der nur in ein Verzeichnis schreiben dürfte, wäre eine **andere**
Fähigkeit unter demselben Namen. Wer die Marke weitergibt, gibt weiter, was
sein eigenes Benutzerkonto darf; das ist dieselbe Zusage wie bei einem
Terminal.

**Und die Marke steht auf der Standardausgabe.** Wer den Dienst in ein
mitgeschriebenes Protokoll startet, schreibt sie mit. Der Handel steht am Code:
eine Marke in einer Datei überlebt den Prozess, dem sie gehört, und muss
aufgeräumt werden — eine auf der Ausgabe gilt genau so lange, wie der Dienst
läuft. Wer das anders will, startet ihn ohne Umleitung.

## Unter welcher Bedingung die Entscheidung wieder fiele

0039 hat seine eigene Widerlegung benannt. Dieser Vermerk tut dasselbe, sonst
wäre er ein Ja ohne Rückweg.

**Sie fiele wieder, wenn eines von dreien einträte:**

1. **Der Dienst wird nicht benutzt.** Ist bei der Freigabe von 0.2.0 kein
   Gegenüber da, das ihn spricht — kein Skript, kein Werkzeug, keine
   Rückmeldung —, dann ist er wieder das, was 0039 abgelehnt hat: eine
   Angriffsfläche ohne Zweck. Dann wird er **entfernt**, nicht behalten, weil
   er schon gebaut ist.
2. **Die vier Testklassen werden einmal übergangen.** Sie sind der Preis, zu
   dem das Ja gegeben wurde. Ein Endpunkt ohne sie ist kein
   Dokumentationsmangel, sondern der Bruch dieser Entscheidung.
3. **Er bekommt eigene Fachlogik.** Der Grund, warum ein vierter Zugang
   vertretbar ist, ist der eine Fachkern dahinter. Ein `if` über einen
   Domänenzustand im HTTP-Handler nimmt ihm diesen Grund.

## Verworfen

**Den Wächter still lockern und keinen Vermerk schreiben.** Das ist genau der
Weg, den 0039 mit „fortzuschreiben, nicht zu umgehen" ausgeschlossen hat. Er
wäre auch der billigste — und in einem halben Jahr stünde ein gelockerter
Wächter da, dessen Grund niemand mehr kennt.

**Den Dienst im Fenster laufen lassen.** 0010 hat das für MCP abgelehnt, und
die beiden Gründe gelten hier genauso: die Standardeingabe des Fensters gehört
dem Fenster, und ein Dienst, der nur läuft, solange jemand das Fenster offen
hat, ist einer, den man vergisst einzuschalten. Der Dienst wird deshalb ein
**eigenes Programm**, wie `mse-mcp` und `mse`.

**REST ohne Marken, weil `127.0.0.1` doch schon schützt.** Tut es nicht: auf
einem Mehrbenutzerrechner erreicht jeder lokale Prozess die Rückschleife, und
ein Browser auf demselben Gerät auch. Das ist der Unterschied zwischen stdio
und einem Port, und er ist der Grund für Punkt 2 von 0039.
