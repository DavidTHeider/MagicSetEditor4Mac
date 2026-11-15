# 0156 — Eine Einladung genügt: Einlösen beim Einladenden über die Paarung

**Stand:** angenommen am 2026-10-01 · NET-01b.AK-01 bis AK-04
(#1954),
Folgeaufgabe aus [0150](0150-einladen-und-vergleichen.md), Satz 3

## Die Frage

Nach [0150](0150-einladen-und-vergleichen.md) tauschen **beide** ihre
Einladung: erst wer die des anderen angenommen hat, kennt beide Schlüssel.
Das sind zwei Codes, die hin und her gehen, wo einer genügen könnte. NET-01b
verlangt, dass wer annimmt, sich beim Einladenden über den gesicherten Weg
meldet (AK-01), dass eine verbrauchte, abgelaufene oder fremde Nonce nichts
anlegt und eine Ratengrenze greift (AK-02), dass eine ältere Gegenstelle
eine erklärende Absage bekommt (AK-03) und dass der Weg mit zwei Einladungen
bleibt (AK-04).

## Was MSE2 tut

Nichts — dieselbe Suche wie in 0150 (`socket`, `pairing`, `invit` in
`/Users/david/mse2-quelle/src`, Fassung f9d9356 vom 2021-07-06). Marke
**new**.

## Was gilt

1. **Die Einladung sagt selbst, ob sie einlösbar ist** — `einloesbar` in den
   **unterschriebenen** Daten, gesetzt, wenn sie eine Adresse nennt. Eine
   Fassung, die das Einlösen nicht kennt, schreibt das Feld nie; wer ihre
   Einladung annimmt, versucht nichts und liest den Rückfall. **So wird eine
   ältere Gegenstelle vor dem Wählen erkannt** und nicht an einer Frist.
2. **Der Einladende merkt sich, was er ausgestellt hat** (Wanderung 0096):
   Nonce, Zweck, Ablauf — und wer sie eingelöst hat. Die Unterschrift sagt
   nur, dass eine Einladung von hier stammt; ob sie **noch gilt**, weiß nur,
   wer sie ausgestellt hat. **Erst vermerkt, dann ausgegeben.**
3. **Eingelöst wird über eine Paarungssitzung.** Die Paarung darf jeder
   öffnen, der nicht gesperrt ist — sie **ist** das Kennenlernen. Kein neuer
   Zweck, keine neue Zeile im Handschlag: ein eigener Zweck wiese eine
   ältere Gegenstelle mit „kein Zweck dieses Programms" ab, bevor sie
   zuhört. Darüber gehen zwei Nachrichten — die Einlösung (Nonce, eigene
   Adresse, gewählter Name) und die Antwort mit einem Ergebnis —, beide mit
   einer eigenen **Fassung** (`domain.Einloesefassung`).
4. **Wer spricht, sagt der Handschlag, nicht die Nachricht.** Aufgenommen
   wird die Kennung, die dort bewiesen hat, dass sie ihren Schlüssel hält —
   **unbestätigt und ohne Recht**. Bestätigt wird weiterhin nur über den
   Vergleichscode; wer schon bestätigt ist, bleibt es; wer gesperrt ist,
   löst nichts ein, und die Nonce bleibt dann unverbraucht.
5. **Jedes Nein wird beantwortet** — `fremd`, `abgelaufen`, `verbraucht`,
   `gesperrt`, `gedrosselt`, `fassung` oder `gescheitert`. Wer annimmt, liest
   einen benannten Fehler und keinen Abriss. Was drüben ohne Namen scheitert,
   heißt `gescheitert`; der Grund bleibt im Protokoll des Einladenden.
6. **Drei Ratengrenzen, in dieser Reihenfolge**: insgesamt (60), je Absender
   (5) und je Einladung (5), je zehn Minuten, im Arbeitsspeicher. **Je
   Einladung**, weil eine Kennung einen Schlüssel kostet, also nichts;
   **insgesamt**, damit die zwei anderen Listen nicht mit jedem gewürfelten
   Versuch wachsen — was sie abweist, wird nicht vermerkt.
7. **Scheitert das Melden, bleibt die Annahme** (§7, Regel 3). Die Antwort
   trägt neben der Liste `gemeldet` — `gemeldet`, `ohne` oder `gescheitert`
   mit Grund —, und das Fenster nennt den Rückfall: die eigene Einladung
   zurückschicken.
8. **Der Name, der hinausgeht, wird gefragt** — leer vorgegeben und nur bei
   einer einlösbaren Einladung, aus demselben Grund, aus dem die eigene
   Einladung keinen vorausgefüllten Namen trägt.

## Was es nicht ist

**Kein neues Vertrauen** (NET-01.AK-03): ein eingelöster Kontakt ist
unbestätigt, bis der Code verglichen ist, und er darf nichts außer der
Paarung. **Kein Dienst im Hintergrund**: eingelöst werden kann nur, solange
der Einladende horcht — sonst ist die Adresse eine Tür, hinter der niemand
steht, und der Satz sagt es.

**Und was ein Mitleser kann** (Zweitblick über Phase 6.7): wer eine
Einladung mitliest, kann sie **zuerst** einlösen. Er wird dann unbestätigter
Kontakt ohne Recht, und die Einlösung des eigentlichen Empfängers sagt
„verbraucht". Mehr nicht: einen fremden Kontakt verändert er ohne dessen
Schlüssel nicht, und beim Vergleich zeigt er andere Ziffern. **Eine Einladung
ist deshalb ein Geheimnis bis zur Einlösung** — so steht es im Fenster neben
dem Code („wirkt nur einmal"), und wer „verbraucht" liest, ohne selbst
eingelöst zu haben, weiß, dass jemand mitgelesen hat.

**Und die Grenzen davor** (Z3): der Horchposten hält höchstens 32 Handschläge
zugleich, jeder hat zehn Sekunden, eine Paarung belegt ihren Platz bis zum
Ende ihrer Aufnahme, und die Liste der gesehenen Zufallszahlen hält höchstens
100.000 — darüber geht die älteste. Die Ratengrenze der Einlösung schützt
die Wirkung, diese Grenzen den Speicher.

## Verworfen

- **Ein eigener Zweck `einloesen`**: eine neue Zeile in `dieZwecke`, und eine
  ältere Gegenstelle wiese ihn ab, bevor sie zuhört — genau die Absage ohne
  Erklärung, die AK-03 ausschließt.
- **Die Nonce im Angebot des Handschlags**: das Drahtformat des Handschlags
  hätte sich geändert, und die Paarung prüft ohnehin erst danach, wer da ist.
- **Den Einladenden ohne Vermerk einlösen lassen**: jede je gesehene Einladung
  wäre ein Schlüssel, auch eine abgelaufene.
- **Den Namen aus dem Handschlag**: er wird dort nicht gesetzt, und ein
  vorausgefüllter wäre einer, den niemand gewählt hat.

## Gemessen

- Domäne, Ablage, Dienst und Anwendung mit ihren Läufen; die Ablage an einer
  echten Datenbank samt Abbrüchen durch Auslöser, nach denen nichts steht.
- **Zwei Anwendungen über eine echte Leitung** (`TestEineEinladungGenuegt`):
  TCP, TLS 1.3 festgenagelt auf die Kennung des Einladenden, Handschlag,
  Einlösung — danach kennt Anna Bert unbestätigt und mit demselben Code.
- **Am nativen Fenster** (Hüllenprobe `net01b`,
  `docs/belege/huellenprobe/net01b.md`): Bert horcht in einem eigenen
  `mse horchen` und lädt an der Kommandozeile mit `--adresse` ein; Anna gibt
  den Code im Fenster ein, das Namensfeld steht leer da, und nach „Später
  vergleichen" steht „Bert kennt dich jetzt auch". Bei Bert steht Anna
  danach unbestätigt und unter ihrem Namen — ohne dass ein zweiter Code
  zurückging.
- **Gegenprobe**: ohne die Aufnahme der Paarung im Horchposten — so verhält
  sich die Fassung davor — meldet jede Fläche sofort „unbeantwortet" und
  keine Frist.
