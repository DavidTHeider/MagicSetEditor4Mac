#!/usr/bin/perl
# **Quarantäne beim ersten Start und der Aufstieg mit befülltem Bestand** —
# I03 (#2261).
#
# Drei Teile, jeder mit seinem Ergebnis im Beleg, auch dem, der scheitert:
#
#  1. **Gatekeeper**: das Freigabeabbild von 0.7.0 und der Kandidat bekommen
#     eine **kenntlich synthetische** Quarantäne (`com.apple.quarantine` mit
#     dem Absender `mse4mac-i03`) — wie ein Download aus dem Browser sie
#     setzt. Gemessen wird, was `spctl` sagt und ob ein `open` startet. Die
#     Freigabe über *Datenschutz & Sicherheit → Dennoch öffnen* gibt ein
#     Mensch; dieser Treiber tut es nicht.
#  2. **Der Aufstieg**: 0.7.0 legt einen Bestand mit Set, Kartenbild, Deck und
#     einem zweiten Nutzer samt Profil an; der Kandidat öffnet ihn, sichert
#     vor der Wanderung, wandert, und liest alles wieder.
#  3. **Der Rückweg**, je nachdem, ob das Schema wanderte — gemessen an der
#     Buchführung `migrationen`, nicht angenommen. **Ohne Wanderung** ist
#     er das alte Bündel selbst: 0.7.0 liest den Bestand samt einem Wert,
#     den der Kandidat geschrieben hat. **Mit Wanderung** weist 0.7.0 den
#     Bestand ab und liest mit der Sicherung davor an seiner Stelle wieder
#     alles (`docs/ADMINISTRATION.md`, „Die Appfassung tauschen").
#
#     perl aufstieg.pl <geteiltes Verzeichnis im Gast>
use strict;
use warnings;
use utf8;
use File::Path qw(make_path);
use File::Temp qw(tempdir);
use FindBin;
use lib $FindBin::Bin;
use Digest::SHA qw(sha256_hex);
use File::Find;
use IPC::Open2;
use JSON::PP;
use Time::HiRes qw(time sleep);

use Gast qw(lesen jetzt buendel messungen ort starte beende aufraeumen);

binmode STDOUT, ':encoding(UTF-8)';
binmode STDERR, ':encoding(UTF-8)';

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $D = "$T/i03";
my $J = JSON::PP->new->utf8->canonical;
messungen($J->decode(lesen("$D/messungen.json")));

# **Ein eigenes Verzeichnis je Lauf**, und das Abbild hängt wieder ab, auch
# nach einem Abbruch (Befund 2 des Zweitblicks über Phase I).
my $W = tempdir('/tmp/i03-XXXXXX');
make_path("$W/alt", "$W/neu");
END { system('hdiutil', 'detach', '-quiet', '-force', "$W/alt-mnt") if -d "$W/alt-mnt"; system('rm', '-rf', $W) }

my %e = (maschine => {
  system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r,
  modell => scalar `sysctl -n hw.model` =~ s/\s+$//r, kerne => scalar `sysctl -n hw.ncpu` =~ s/\s+$//r });
my (@schritte, @grenzen);

sub schritt {
  my ($teil, $was, $tun) = @_;
  my $beginn = jetzt();
  my $ergebnis = eval { $tun->() };
  my $fehler = $@ ? $@ =~ s/\s+$//r : undef;
  push @schritte, { teil => $teil, was => $was, ok => $fehler ? JSON::PP::false : JSON::PP::true,
    ms => int(jetzt() - $beginn), ($fehler ? (fehler => $fehler) : ()), (defined $ergebnis && !$fehler ? (wert => "$ergebnis") : ()) };
  die "$was: $fehler\n" if $fehler;
  return $ergebnis;
}

# programm ruft ein Werkzeug aus einem Bündel in einem Ort — Rückgabe und Text.
sub programm {
  my ($app, $werkzeug, $o, @argumente) = @_;
  local $ENV{HOME} = $o->{heim};
  local $ENV{MSE_BESTAND} = $o->{bestand};
  local $ENV{MSE_SCHLUESSELBUND} = $o->{bund};
  my $text = `'$app/Contents/MacOS/$werkzeug' @{[ map { "'$_'" } @argumente ]} 2>&1`;
  return ($? >> 8, $text // '');
}
sub alt { my ($o, @a) = @_; return programm("$W/alt/MagicSetEditor4Mac.app", 'mse', $o, @a) }
sub neu { my ($o, @a) = @_; return programm("$W/neu/Abnahme.app", 'mse', $o, @a) }
sub muss { my ($rc, $text) = @_; die "Rückgabe $rc: " . substr($text, 0, 300) . "\n" if $rc; return $text }

# mcp fährt Rufe über `mse-mcp` auf stdio — eine Zeile JSON je Nachricht.
sub mcp {
  my ($app, $o, @rufe) = @_;
  local $ENV{HOME} = $o->{heim};
  local $ENV{MSE_BESTAND} = $o->{bestand};
  local $ENV{MSE_SCHLUESSELBUND} = $o->{bund};
  my $pid = open2(my $aus, my $ein, "$app/Contents/MacOS/mse-mcp");
  my $id = 0;
  my $sende = sub { print {$ein} $J->encode($_[0]), "\n" };
  my $lies = sub { my $z = <$aus>; die "mse-mcp schweigt\n" unless defined $z; return $J->decode($z) };
  $sende->({ jsonrpc => '2.0', id => ++$id, method => 'initialize',
    params => { protocolVersion => '2025-06-18', capabilities => {}, clientInfo => { name => 'i03', version => '1' } } });
  $lies->();
  $sende->({ jsonrpc => '2.0', method => 'notifications/initialized' });
  my @antworten;
  for my $ruf (@rufe) {
    $sende->({ jsonrpc => '2.0', id => ++$id, method => 'tools/call', params => { name => $ruf->[0], arguments => $ruf->[1] } });
    my $a = $lies->();
    my $text = $a->{result}{content}[0]{text} // ($a->{error}{message} // '');
    die "$ruf->[0]: $text\n" if $a->{error} || $a->{result}{isError};
    push @antworten, $text;
  }
  close $ein;
  waitpid $pid, 0;
  return @antworten;
}

# inhalt liest, was der Bestand trägt — mit einem Bündel; dieselbe Frage an
# beide Fassungen, damit der Vergleich nichts übersetzt.
# abdruck kürzt einen Abdruck für den Beleg.
sub abdruck { return substr(sha256_hex($_[0]), 0, 12) }

# plaetze sucht in einer Antwort jedes Objekt, das einen Deckplatz beschreibt.
sub plaetze {
  my ($wert) = @_;
  # **Nur Hashes und Listen** — `vorlaeufig: false` kommt als
  # JSON::PP::Boolean, eine Referenz, die kein Hash ist (gemessen am
  # 2026-10-11 an der Antwort von `lesen.deck`).
  return map { plaetze($_) } @$wert if ref $wert eq 'ARRAY';
  return () unless ref $wert eq 'HASH';
  my @eigene = exists $wert->{karteID} && exists $wert->{menge} ? ("$wert->{karteID}:$wert->{menge}:" . ($wert->{bereich} // '')) : ();
  return (@eigene, map { plaetze($_) } values %$wert);
}

# inhalt liest, was der Bestand trägt — **Inhalte, nicht Anzahlen** (Befund 5
# des Zweitblicks über Phase I): die Werte jeder Karte, die Plätze jedes
# Decks, die Bytes jedes Bilds, Nutzer samt Profilen. Dieselbe Frage an beide
# Fassungen, damit der Vergleich nichts übersetzt.
sub inhalt {
  my ($app, $o) = @_;
  my $mse = sub { muss(programm($app, 'mse', $o, @_)) };
  my %i;
  my $sets = $mse->('sets');
  $i{sets} = scalar grep { /\S/ } split /\n/, $sets;
  my ($set) = $sets =~ /^(\d+)\t/m;
  my $karten = JSON::PP->new->utf8->decode($mse->('karten', $set, '--json'))->{karten};
  $i{karten} = scalar @$karten . ' ' . abdruck($J->encode([map { [$_->{id}, $_->{werte}] } sort { $a->{id} <=> $b->{id} } @$karten]));

  # **Die Bilder folgen dem Bestand** (`beiwerk/` daneben) — gelesen wird,
  # was dort liegt, Byte für Byte.
  my @bilder;
  find({ no_chdir => 1, wanted => sub { push @bilder, $_ if -f } }, "$o->{ort}/beiwerk") if -d "$o->{ort}/beiwerk";
  $i{bilder} = scalar @bilder . ' ' . abdruck(join "\n", sort map { substr($_, length "$o->{ort}/") . ' ' . sha256_hex(lesen($_)) } @bilder);

  my ($decks) = mcp($app, $o, ['lesen.decks', {}]);
  # **Die Antworten von `mse-mcp` sind schon Zeichen** — `$lies` hat die Zeile
  # dekodiert; ein zweites `->utf8` verlangte Bytes und brach an jedem
  # Umlaut eines Kartennamens ab (gemessen am 2026-10-11).
  my $zeichen = JSON::PP->new;
  my @decks = @{ $zeichen->decode($decks) };
  $i{decks} = join ';', map {
    my ($deck) = mcp($app, $o, ['lesen.deck', { deckID => $_->{id} + 0 }]);
    # **Als Bytes** in den Abdruck, wie die Karten: `sha256_hex` nimmt keine
    # Zeichen über 255.
    "$_->{name}=" . abdruck($J->encode([sort(plaetze($zeichen->decode($deck)))]))
  } sort { $a->{id} <=> $b->{id} } @decks;

  # Nutzer stehen als `id⇥Name`, ihre Profile darunter als `⇥id⇥Name`.
  my (@nutzer, $zuletzt);
  for (split /\n/, $mse->('nutzer')) {
    if (/^\d+\t(.+)$/) { push @nutzer, $zuletzt = { name => $1, profile => [] } }
    elsif (/^\t\d+\t(.+)$/ && $zuletzt) { push @{ $zuletzt->{profile} }, $1 }
  }
  $i{nutzer} = join ';', map { "$_->{name}:" . join(',', sort @{ $_->{profile} }) } sort { $a->{name} cmp $b->{name} } @nutzer;
  return \%i;
}

# schema zählt die verbuchten Wanderungen — die Buchführung, nicht die Nummer
# der Fassung: zwischen zwei Freigaben trägt jeder Bau dieselbe
# (`docs/FASSUNGEN.md`).
sub schema { my ($o) = @_; my $n = `sqlite3 '$o->{bestand}' 'SELECT count(*) FROM migrationen' 2>&1`; return $n =~ /^(\d+)/ ? $1 + 0 : die "Schema: $n\n" }

# --- 1. Gatekeeper: kenntlich synthetische Quarantäne
#
# **Am Wirt nicht** (`MSE_I03_OHNE_GATEKEEPER`, die Positivkontrolle): ein
# `open` dort legte die Abweisung auf den Bildschirm dessen, der arbeitet.
my $marke = sprintf('0083;%x;mse4mac-i03;', time());
my $amWirt = $ENV{MSE_I03_OHNE_GATEKEEPER};
eval {
  schritt('Gatekeeper', '0.7.0-Abbild prüfen', sub {
    my $soll = (split ' ', lesen("$D/pruefsumme.txt"))[0];
    my $ist = (split ' ', `shasum -a 256 '$D/alt.dmg'`)[0];
    $soll eq $ist or die "Prüfsumme $ist statt $soll\n";
    return substr($ist, 0, 16);
  });
  system('cp', "$D/alt.dmg", "$W/alt.dmg") == 0 or die "kopieren\n";
  system('xattr', '-w', 'com.apple.quarantine', $marke, "$W/alt.dmg");
  schritt('Gatekeeper', 'Abbild einhängen', sub {
    system('hdiutil', 'attach', '-nobrowse', '-readonly', '-quiet', '-mountpoint', "$W/alt-mnt", "$W/alt.dmg") == 0 or die "hdiutil\n";
    return 1;
  });
  my ($app) = glob("$W/alt-mnt/*.app");
  system('cp', '-R', $app, "$W/quarantaene.app") == 0 or die "aus dem Abbild\n";
  $e{gatekeeper}{erbt} = `xattr -p com.apple.quarantine '$W/quarantaene.app' 2>&1` =~ s/\s+$//r;
  for my $fall ($amWirt ? () : (['0.7.0 aus dem Abbild', "$W/quarantaene.app"], ['Kandidat, Quarantäne am Bündel', "$W/kandidat-q.app"])) {
    if ($fall->[1] =~ /kandidat/) {
      system('ditto', "$D/Abnahme.app", $fall->[1]);
      system('xattr', '-w', 'com.apple.quarantine', $marke, $fall->[1]);
    }
    schritt('Gatekeeper', "spctl: $fall->[0]", sub { my $t = `spctl --assess --type execute -vv '$fall->[1]' 2>&1`; return "Rückgabe " . ($? >> 8) . ": " . ($t =~ s/\s+/ /gr) });
    schritt('Gatekeeper', "open: $fall->[0]", sub {
      system('open', '-n', '-a', $fall->[1], '--env', "MSE_BESTAND=$W/leer.db", '--env', "HOME=$W");
      sleep 8;
      my $laeuft = `pgrep -f '$fall->[1]/Contents/MacOS/'` =~ /\d/;
      system('pkill', '-f', "$fall->[1]/Contents/MacOS/") if $laeuft;
      return $laeuft ? 'gestartet' : 'nicht gestartet';
    });
  }
  $e{gatekeeper}{protokoll} = $amWirt ? '' : `log show --last 2m --style compact --predicate 'process == "syspolicyd" AND eventMessage CONTAINS[c] "i03"' 2>/dev/null | tail -5`;
  # Die Programme dieses Laufs ohne Quarantäne — wie nach »Dennoch öffnen«.
  system('ditto', '--noqtn', $app, "$W/alt/MagicSetEditor4Mac.app") == 0 or die "0.7.0 auspacken\n";
  system('ditto', '--noqtn', "$D/Abnahme.app", "$W/neu/Abnahme.app") == 0 or die "Kandidat\n";
  system('hdiutil', 'detach', '-quiet', "$W/alt-mnt");
  buendel("$W/neu/Abnahme.app");
  1;
} or push @grenzen, 'Gatekeeper brach ab — die folgenden Schritte fehlen';

# --- 2. Der Aufstieg
my $o = ort();
eval {
  schritt('0.7.0', 'Fassung', sub { (muss(alt($o, 'meta')) =~ /(\d+\.\d+\.\d+)/)[0] // die "keine Fassung\n" });
  my $set = schritt('0.7.0', 'Set einfahren', sub { (muss(alt($o, 'oeffnen', "$D/probe-werkstatt.mse-set", '--json')) =~ /"setID"\s*:\s*"?(\d+)/)[0] // die "keine Setkennung\n" });
  my $karte = schritt('0.7.0', 'Karte lesen', sub { (muss(alt($o, 'karten', $set, '--json')) =~ /"id"\s*:\s*"?(\d+)/)[0] // die "keine Karte\n" });
  schritt('0.7.0', 'Kartenbild einsetzen', sub {
    muss(alt($o, 'zeichnen', $karte, "$o->{ort}/bild.png"));
    mcp("$W/alt/MagicSetEditor4Mac.app", $o, ['schreiben.karte.bild', { setID => $set + 0, karteID => $karte + 0, feld => 'kunst', pfad => "$o->{ort}/bild.png" }]);
    return 1;
  });
  schritt('0.7.0', 'Deck aus dem Set', sub { mcp("$W/alt/MagicSetEditor4Mac.app", $o, ['schreiben.deck.ausset', { setID => $set + 0, name => 'Aufstieg', menge => 2 }]); 1 });
  schritt('0.7.0', 'zweiter Nutzer samt Profil', sub {
    muss(alt($o, 'nutzer-neu', 'Zweite'));
    my ($id) = muss(alt($o, 'nutzer')) =~ /^(\d+)\tZweite/m or die "kein Nutzer Zweite\n";
    muss(alt($o, 'profil-neu', $id, 'Turnier'));
    return $id;
  });
  $e{vorher} = schritt('0.7.0', 'Inhalt', sub { $J->encode(inhalt("$W/alt/MagicSetEditor4Mac.app", $o)) });

  my $vorher = schritt('0.7.0', 'Schema', sub { schema($o) });
  schritt('Kandidat', 'Aufstieg', sub { muss(neu($o, 'sets')); 1 });
  my $nachher = schritt('Kandidat', 'Schema', sub { schema($o) });
  $e{schema} = { vorher => $vorher, nachher => $nachher };
  $e{sicherungen} = [map { s{.*/}{}r } glob("$o->{ort}/bestand-vor-*.db")];
  $e{nachher} = schritt('Kandidat', 'Inhalt', sub { $J->encode(inhalt("$W/neu/Abnahme.app", $o)) });
  schritt('Kandidat', 'derselbe Inhalt', sub { $e{vorher} eq $e{nachher} or die "vorher $e{vorher}, nachher $e{nachher}\n"; 1 });
  # **Das Bild wird benutzt, nicht nur gezählt** — die Karte mit dem
  # eingesetzten Bild wird mit dem Kandidaten gezeichnet.
  schritt('Kandidat', 'die Karte mit dem Bild zeichnen', sub {
    muss(neu($o, 'zeichnen', $karte, "$o->{ort}/nachher.png"));
    open my $f, '<:raw', "$o->{ort}/nachher.png" or die "kein Bild\n";
    read $f, my $kopf, 8;
    close $f;
    $kopf eq "\x89PNG\r\n\x1a\n" or die "kein PNG\n";
    return -s "$o->{ort}/nachher.png";
  });
  schritt('Kandidat', 'das Fenster über dem Bestand', sub { my ($ms) = starte($o); beende() or die "⌘Q beendet nicht\n"; return int($ms) });

  if ($vorher == $nachher) {
    # **Ohne Wanderung gibt es keine Sicherung davor** — und der Rückweg ist
    # das alte Bündel selbst. Damit er nicht nur Lesen belegt, schreibt der
    # Kandidat vorher einen Wert, den 0.7.0 danach lesen muss.
    schritt('Kandidat', 'keine Wanderung, keine Sicherung davor', sub { @{ $e{sicherungen} } and die "eine Sicherung ohne Wanderung: @{ $e{sicherungen} }\n"; 1 });
    schritt('Kandidat', 'ein Wert nach dem Aufstieg', sub { muss(neu($o, 'feld', $set, $karte, 'name', 'Nach dem Aufstieg')); 1 });
    $e{geschrieben} = schritt('Kandidat', 'Inhalt nach dem Schreiben', sub { $J->encode(inhalt("$W/neu/Abnahme.app", $o)) });
    schritt('Rückweg', '0.7.0 liest den Bestand des Kandidaten', sub { muss(alt($o, 'karten', $set, '--json')) =~ /Nach dem Aufstieg/ or die "der Wert des Kandidaten fehlt\n"; 1 });
    $e{zurueck} = schritt('Rückweg', 'Inhalt mit 0.7.0', sub { $J->encode(inhalt("$W/alt/MagicSetEditor4Mac.app", $o)) });
    schritt('Rückweg', 'derselbe Inhalt wie nach dem Schreiben', sub { $e{geschrieben} eq $e{zurueck} or die "geschrieben $e{geschrieben}, zurück $e{zurueck}\n"; 1 });
  } else {
    schritt('Kandidat', 'Sicherung vor der Wanderung', sub { @{ $e{sicherungen} } or die "keine Sicherung neben dem Bestand\n"; join ', ', @{ $e{sicherungen} } });
    schritt('Rückweg', '0.7.0 weist den gewanderten Bestand ab', sub { my ($rc, $t) = alt($o, 'sets'); $rc or die "0.7.0 öffnete ihn\n"; return substr($t =~ s/\s+/ /gr, 0, 200) });
    schritt('Rückweg', 'Sicherung an die Stelle des Bestands', sub {
      my ($sicherung) = @{ $e{sicherungen} };
      system('mv', $o->{bestand}, "$o->{bestand}.gewandert") == 0 or die "beiseitelegen\n";
      unlink "$o->{bestand}-wal", "$o->{bestand}-shm";
      system('cp', "$o->{ort}/$sicherung", $o->{bestand}) == 0 or die "zurücklegen\n";
      return $sicherung;
    });
    $e{zurueck} = schritt('Rückweg', 'Inhalt mit 0.7.0', sub { $J->encode(inhalt("$W/alt/MagicSetEditor4Mac.app", $o)) });
    schritt('Rückweg', 'derselbe Inhalt', sub { $e{vorher} eq $e{zurueck} or die "vorher $e{vorher}, zurück $e{zurueck}\n"; 1 });
  }
  1;
} or push @grenzen, 'der Aufstieg brach ab — die folgenden Schritte fehlen';

aufraeumen();
system('rm', '-rf', $W);
$e{schritte} = \@schritte;
$e{grenzen} = \@grenzen;
open my $aus, '>', "$T/i03-ergebnis.json" or die "$!\n";
print {$aus} $J->encode(\%e);
close $aus;
my @rot = grep { !$_->{ok} } @schritte;
print @rot ? "mit Fehlern: " . join('; ', map { "$_->{was}: $_->{fehler}" } @rot) . "\n" : "fertig\n";
