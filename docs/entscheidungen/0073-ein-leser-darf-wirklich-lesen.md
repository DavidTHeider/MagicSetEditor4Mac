# 0073 — Ein `leser` darf wirklich lesen

**Entschieden am 2026-09-06**, Werkstück 16 von Phase 4.12
(#1351) ·
**Ergänzt** [0070](0070-rechte-rollen-und-sichtbarkeit.md)

## Der Anlass

[0070](0070-rechte-rollen-und-sichtbarkeit.md) hat `leser` bewusst klein
gehalten: er sieht, **dass** es das Profil gibt, nicht was darin steht. Der
Grund stand dort ausdrücklich — ein Nur-Lese-Modus **innerhalb** eines Profils
braucht einen lückenlosen Schreibschutz, und den gibt es an genau einer Stelle:
im Treiber. Eine Zusage, die an 116 Bindungen einzeln hängt — **am 2026-09-06
gezählt**, und die Zahl wächst —, bricht beim ersten vergessenen Ruf still.

Der Auftraggeber hat sich für die Treiberfassung entschieden.

## Portierungsmarke: **new** — und hier ist gesucht worden

§15 verlangt, dass eine `new`-Marke sagt, **wo** gesucht wurde. Am 2026-09-06
gemessen:

| Quelle | Treffer auf `role`, `permission`, `readonly` | Was es ist |
| --- | ---: | --- |
| `/Users/david/mse2-quelle/src` | **2** | beide unbeteiligt: ein Kommentar über das Observer-Muster (`action_stack.hpp`) und `wxCB_READONLY` an einer Combobox (`preferences_window.cpp`) |
| `/Users/david/magic-set-generator/src` | **0** | — |

**MSE2 kennt keine Nutzer** — es kennt keine Rollen, keine Freigaben und
keinen Nur-Lese-Modus über Menschen hinweg. Es gibt kein Altverhalten, das zu
portieren wäre; dasselbe stellt schon [0070](0070-rechte-rollen-und-sichtbarkeit.md)
für die Rollen selbst fest.

## 1. Der Schutz sitzt im Treiber, und zwar in `query_only`

`PRAGMA query_only` gilt für die **Verbindung**, nicht für eine Anweisung.
Damit gibt es keine Stelle, an der jemand ihn vergessen kann.

**Gemessen am 2026-09-06** mit der Produktivadresse:

| Weg | Ergebnis |
| --- | --- |
| `SELECT` | glückt |
| `INSERT`, `DELETE`, `CREATE TABLE` | `SQLITE_READONLY` |
| `BeginTx` mit `_txlock=immediate` | `SQLITE_READONLY` — **schon das Beginnen** |
| dieselbe Datei gewöhnlich geöffnet | nimmt den Schreibvorgang an |

Die letzte Zeile ist die Gegenprobe: ohne sie sagte die Messung nichts über das
Pragma, sondern nur über diese Datei.

**Und `SQLITE_READONLY` lief bis dahin roh nach oben.** `einordnen` kannte den
Code nicht; an der Oberfläche stand „attempt to write a readonly database" —
ein roher Treibertext, den §8 verbietet, und einer, der niemandem sagt, dass es
an einer **Freigabe** liegt und nicht an der Datei. Jetzt gibt es
`domain.ErrNurLesend`.

## 2. Es ist eine **zweite** Verbindung, und der Grund ist gemessen

Die naheliegende Fassung wäre, den Bestand gleich nur-lesend zu öffnen. Das
geht nicht: **`Migriere` beginnt unbedingt eine Transaktion**, auch wenn nichts
zu tun ist — gemessen am 2026-09-06:

```
Migriere auf einem aktuellen Bestand, nur lesend:
  Migration beginnen: in diesem Profil darf nur gelesen werden
```

Der Ablauf ist deshalb: **öffnen, migrieren, und erst dann** eine zweite
Verbindung mit `query_only(1)` nehmen, über der die Repositories gebaut werden.
Die erste bleibt — sie schreibt den Profilzeiger, und sonst nichts.

**Dass die Migration laut scheitert, ist richtig so.** Ein Bestand, dessen
Schema hinterherhinkt, soll nicht still halb gelesen werden. Im Betrieb tritt
der Fall nicht ein, weil auf der ersten Verbindung migriert wird.

## 3. Es setzt einen **angemeldeten Nutzer** voraus — und der kam erst dazu

**Das ist der Fund, der die Reihenfolge der Phase geändert hat.**

Beim Entwurf dieses Werkstücks stellte sich heraus: es gibt in dieser
Anwendung **keinen angemeldeten Nutzer**. Wer „arbeitet", wurde aus dem
*gezeigten Profil* abgeleitet — `freigabeReicht` fragte `gezeigt.NutzerID`.
Nach einem Wechsel in ein fremdes Profil ist das der **fremde** Eigner, und ein
Eigner ist `RolleVerwalter`.

**Daraus folgte: „ich bin hier nur `leser`" ließ sich aus dem Zeiger nicht
ablesen.** Wer es versuchte, bekäme unmittelbar nach dem Wechsel die Antwort
„Verwalter" — und schriebe.

**Gemessen am 2026-09-06** an `internal/service`, mit drei Nutzern und zwei
Freigaben: Erster darf auf Zweites, Zweiter darf auf Drittes, Erster darf
**nicht** auf Drittes. Der Erste wechselt auf Zweites — erlaubt. Von dort auf
Drittes: **glückt**, ohne Nachweis. Geprüft wurde als *Zweiter*.

Der erste Entwurf dieses Werkstücks wollte das umgehen, indem er die
**Eintrittsart** neben dem Zeiger festhält. Das hätte für den Schreibschutz
gereicht und den Befund darunter stehen gelassen: **auch jeder weitere Wechsel**
wird als der fremde Eigner geprüft, und damit lassen sich Freigaben
aneinanderhängen
(#1400).

**Der Auftraggeber hat entschieden, #1400 in diese Phase zu nehmen.** Damit
kehrt sich die Reihenfolge um, und das Werkstück wird **kleiner**:

- **Werkstück 17 zuerst** — „wer arbeitet" wird festgehalten und ändert sich
  durch einen Profilwechsel nicht mehr.
- **Werkstück 16 danach** — die Rolle im gezeigten Profil ist dann wieder
  ableitbar (`Rolle(werArbeitet, gezeigt)`), und es braucht **keinen**
  festgehaltenen Eintrittsvermerk. Jeder Prozess kann die Frage beim Öffnen
  selbst stellen.

**Was das kostet, gehört gesagt**: Phase 4.12 wächst auf siebzehn Werkstücke,
und 0.4.0 rückt entsprechend nach hinten. Der Grund steht in `docs/ROADMAP.md`
(§13).

## 4. Geschaltet wird an zwei Flächen, geschützt an vier

**Der Schutz gilt überall**, denn er hängt an einem festgehaltenen Stand: jeder
Prozess liest ihn beim Öffnen und nimmt die zweite Verbindung, wenn er gesetzt
ist. Wer die Anwendung neu startet, kommt nicht daran vorbei.

**Der Wechsel *in* ein `leser`-Profil geht am Fenster und an der
Kommandozeile.** Dienst und MCP-Server halten eine Verbindung über ihre ganze
Laufzeit, an der jede Naht hängt; die Repositories mitten im Betrieb auf eine
zweite umzuhängen hieße, sie alle neu zu bauen. **Das ist wörtlich der Grund,
aus dem dort schon das Zurückspielen fehlt** und aus dem die Verschlüsselung
dort nicht geschaltet wird ([0072](0072-ein-treiber-der-verschluesseln-kann.md)).

**Was das kostet, gehört gesagt**: wer über MCP oder den Dienst in ein Profil
wechseln will, in dem er nur `leser` ist, bekommt eine Absage mit diesem Grund
— und tut es am Fenster oder mit `mse profil-wechseln`.

## 5. Was `leser` jetzt heißt

| | vorher (0070) | jetzt |
| --- | --- | --- |
| Sieht das Profil in der Übersicht | ja | ja |
| Kommt hinein | **nein** | **ja** |
| Sieht, was darin steht | nein | **ja** |
| Schreibt | nein | nein — und zwar im Treiber |

**Der Satz in `domain.RolleLeser` wird damit falsch** („Er kommt nicht
hinein") und ist fortgeschrieben.
