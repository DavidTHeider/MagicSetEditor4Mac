# 0176 — Eine Verbindung, eine Identität; die Nonceliste ist die zweite Reihe

**Stand:** angenommen am 2026-10-07 · SEC-15, SEC-16
(#2144,
#2145).

## Die Fragen

**SEC-16.** Eine gesicherte Leitung kennt zwei Auskünfte über ihr
Gegenüber: das Zeugnis im TLS-Handschlag und die Unterschrift im
Anwendungshandschlag. Die horchende Seite verlangte ein Zeugnis
(`RequireAnyClientCert`) und prüfte es gegen niemanden; die Unterschrift
prüfte sie gegen ihre Kontakte. **Gegeneinander hielt beides keiner** — TLS
mit dem Schlüssel A und ein gültiger Handschlag von B ergaben eine Sitzung
mit B. `service.Verbindung` sagte sogar ausdrücklich, der Transport nenne
kein Gegenüber, weil das „eine zweite Quelle für dieselbe Frage" wäre — die
zweite Quelle gab es längst.

**SEC-15.** Die Nonceliste hält höchstens 100.000 Zufallszahlen eine Stunde
lang; über der Grenze verdrängt die neue die älteste — **auch vor deren
Ablauf**. Eine verdrängte Zahl nähme die Liste danach wieder an.

## Die Entscheidungen

**Eine Verbindung hat eine Identität** (SEC-16). Die Leitung trägt die
Kennung aus dem geprüften Zeugnis weiter (`Leitung.Gegenstelle`, festgehalten
in `VerifyConnection`, das auch bei einer wiederaufgenommenen Sitzung läuft),
und `service.Peersitzungen` hält sie auf **beiden** Seiten gegen die Kennung
des Handschlags: die horchende vor dem Recht, die wählende nach dem
Schließen. Ein Widerspruch ist `ErrValidation`, die Leitung geht zu.

| Lage | Ausgang |
| --- | --- |
| Zeugnis A, Handschlag A | Sitzung |
| Zeugnis A, Handschlag B (horchend gesehen) | abgewiesen, vor dem Recht |
| Zeugnis A, Handschlag B (wählend gesehen) | abgewiesen — und schon vorher am Festnageln und an `Erwartet` |
| kein Zeugnis — eine Röhre im selben Prozess | allein der Handschlag entscheidet |

**Die Nonceliste bleibt eine begrenzte Zusatzwache** (SEC-15), keine strikte
Frist. Eine strikte Frist wiese bei voller Liste jedes neue Angebot ab: ein
Fremder, der in einer Stunde 100.000 gültig geformte Angebote schickt,
sperrte jeden Kontakt aus — er braucht dafür keine Freigabe, nur eine
Verbindung. **Der Schutz gegen ein Wiedereinspielen ruht auf der
Kanalbindung**: jedes Angebot trägt den Kanal seiner Verbindung, und
zwischen denselben beiden Rechnern hat jede Verbindung einen anderen
(`leitung/gegenstelle_test.go`, aus `ExportKeyingMaterial`). Die Liste fängt
dasselbe Angebot auf **derselben** Verbindung — solange sie es hält.

**Die Zusage, genau**: eine Zufallszahl wird innerhalb einer Stunde nicht
zweimal angenommen, **solange in dieser Stunde weniger als 100.000 neue
kamen**. Genau auf der Frist gilt sie noch, danach nicht mehr.

## Was es nicht ist

- **Keine Antwort auf einen gestohlenen Schlüssel.** Wer A und B hält, ist
  beide; SEC-16 verhindert nur, dass eine Verbindung zwei Auskünfte trägt.
- **Kein Wiedereinspielen auf einer Verbindung ohne eigenen Kanal belegt.**
  Ein Transport mit fester Bindung — die Röhre im selben Prozess — hängt
  allein an der Liste; im Betrieb gibt es ihn nicht.
