#!/usr/bin/perl
# **Der Aufstieg mit einem eingerichteten Dienst unter launchd** — M02
# (#2346),
# der Lauf, den der Arbeitsrechner nicht tragen darf: ein LaunchAgent ist
# dauerhafte Einrichtung, und hier ist er Wegwerfware.
#
#  1. **0.8.0 mit Bestand und Dienst**: das Bündel von 0.8.0 unter
#     `~/Applications`, ein künstliches Set darin, der Dienst unter seinem
#     Etikett über `launchctl bootstrap` — so, wie das Einrichten am Fenster
#     ihn hinterließ. Er antwortet und hält den Bestand.
#  2. **Der Kandidat an derselben Stelle**, wie bei einer Aktualisierung.
#  3. **Der erste Ruf des Kandidaten wandert** — und vorher startet er den
#     Dienst neu (`launchctl kickstart -k`): der von 0.8.0 kennt die Frage
#     nach dem Bestand nicht. Geprüft: genau eine Sicherung mit dem Stand von
#     0.8.0, alle Wanderungen, der Dienst läuft danach **mit dem Programm des
#     Kandidaten** und antwortet.
#  4. **Ein zweiter Ruf** sichert nicht noch einmal und startet nichts neu.
#
#     perl m02.pl <geteiltes Verzeichnis im Gast>
use strict;
use warnings;
use utf8;
use JSON::PP;
use Time::HiRes qw(time sleep);

binmode STDOUT, ':encoding(UTF-8)';
binmode STDERR, ':encoding(UTF-8)';

`sysctl -n hw.model` =~ /^VirtualMac/ or die "nur in einem Gast — diese Maschine ist kein VirtualMac\n";

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $D = "$T/m02";
my $J = JSON::PP->new->utf8->canonical;
my $H = $ENV{HOME};
my $uid = $<;
my $APP = "$H/Applications/MagicSetEditor4Mac.app";
my $ORT = "$H/Library/Application Support/MagicSetEditor4Mac";
my $BESTAND = "$ORT/bestand.db";
my $ETIKETT = lesen("$D/etikett.txt") =~ s/\s+$//r;
my $PLIST = "$H/Library/LaunchAgents/$ETIKETT.plist";
my $ANSAGE = "$H/m02-dienst.txt";

my %e = (maschine => {
  system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r,
  modell => scalar `sysctl -n hw.model` =~ s/\s+$//r });
my (@schritte, @grenzen);

sub lesen { local $/; open my $f, '<', $_[0] or die "$_[0]: $!\n"; my $x = <$f>; close $f; return $x }
sub jetzt { return time() * 1000 }

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

# mse ruft die Kommandozeile des eingesetzten Bündels — Rückgabe und Text.
sub mse {
  my $befehl = join ' ', map { "'$_'" } ("$APP/Contents/MacOS/mse", @_);
  my $text = `$befehl 2>&1`;
  return ($? >> 8, $text);
}

sub einsetzen {
  my ($quelle) = @_;
  system('rm', '-rf', $APP);
  system('mkdir', '-p', "$H/Applications");
  system('ditto', '--noqtn', $quelle, $APP) == 0 or die "$quelle einsetzen\n";
  return 1;
}

# dienstpid nennt die Prozessnummer des geladenen Dienstes — 0, wenn keiner läuft.
sub dienstpid { my ($p) = `launchctl print gui/$uid/$ETIKETT 2>/dev/null` =~ /^\s*pid = (\d+)/m; return $p // 0 }

# ansage liest die letzte Ansage des Dienstes: Adresse und Marke.
sub ansage {
  my $text = -e $ANSAGE ? lesen($ANSAGE) : '';
  my @adressen = $text =~ /^Dienst auf (\S+)$/mg;
  my @marken = $text =~ /^Marke (\S+)$/mg;
  return (scalar @marken, $adressen[-1], $marken[-1]);
}

# antwortet fragt den Dienst nach den Sets — der HTTP-Status.
sub antwortet {
  my (undef, $adresse, $marke) = ansage();
  return 0 unless $adresse && $marke;
  my $status = `curl -s -o /dev/null -w '%{http_code}' -H 'Authorization: Bearer $marke' '$adresse/sets'`;
  return $status + 0;
}

sub warte {
  my ($was, $bedingung, $frist) = @_;
  my $bis = time() + ($frist // 30);
  until ($bedingung->()) { die "$was nicht nach " . ($frist // 30) . " s\n" if time() > $bis; sleep 0.5 }
  return 1;
}

sub wanderungen { my $z = `sqlite3 '$BESTAND' 'SELECT count(*) FROM migrationen'`; chomp $z; return $z + 0 }
sub zuletzt { my $z = `sqlite3 '$BESTAND' 'SELECT name FROM migrationen ORDER BY name DESC LIMIT 1'`; chomp $z; return $z }
sub sicherungen { opendir my $d, $ORT or return (); my @s = sort grep { /^bestand-vor-.*\.db$/ } readdir $d; closedir $d; return @s }

sub raeumen {
  system("launchctl bootout gui/$uid/$ETIKETT >/dev/null 2>&1");
  unlink $PLIST, $ANSAGE;
  system('rm', '-rf', $ORT, $APP);
  return 1;
}

raeumen();
eval {
  schritt('0.8.0', 'das Bündel einsetzen', sub { einsetzen("$D/alt/MagicSetEditor4Mac.app") });
  schritt('0.8.0', 'ein künstliches Set einfahren', sub {
    my ($rc, $t) = mse('oeffnen', "$D/probe-werkstatt.mse-set");
    $rc == 0 or die "Rückgabe $rc: $t\n";
    return wanderungen() . ' Wanderungen';
  });
  my $baseline = wanderungen();
  schritt('0.8.0', 'der Dienst unter launchd', sub {
    system('mkdir', '-p', "$H/Library/LaunchAgents");
    my $inhalt = lesen("$D/dienst.plist") =~ s/__ETIKETT__/$ETIKETT/gr =~ s/__PROGRAMM__/$APP\/Contents\/MacOS\/mse-rest/gr =~ s/__ANSAGE__/$ANSAGE/gr;
    open my $f, '>', $PLIST or die "$!\n";
    print {$f} $inhalt;
    close $f;
    system("launchctl bootstrap gui/$uid '$PLIST'") == 0 or die "bootstrap\n";
    warte('die Ansage', sub { (ansage())[0] >= 1 });
    return 'pid ' . dienstpid();
  });
  schritt('0.8.0', 'er antwortet und hält den Bestand', sub {
    my $status = antwortet();
    $status == 200 or die "Status $status\n";
    my $pid = dienstpid();
    `lsof -p $pid 2>/dev/null` =~ /bestand\.db/ or die "er hält den Bestand nicht\n";
    return "200, pid $pid";
  });
  my $altpid = dienstpid();

  schritt('Kandidat', 'an derselben Stelle einsetzen', sub { einsetzen("$D/neu/MagicSetEditor4Mac.app") });
  my $eingesetzt = time();
  schritt('Kandidat', 'der erste Ruf wandert', sub {
    my ($rc, $t) = mse('sets');
    $rc == 0 or die "Rückgabe $rc: $t\n";
    wanderungen() > $baseline or die "keine Wanderung (" . wanderungen() . ")\n";
    return wanderungen() . ' Wanderungen, zuletzt ' . zuletzt();
  });
  schritt('Kandidat', 'genau eine Sicherung, mit dem Stand von 0.8.0', sub {
    my @s = sicherungen();
    @s == 1 or die scalar(@s) . " Sicherungen: @s\n";
    my $z = `sqlite3 '$ORT/$s[0]' 'SELECT count(*) FROM migrationen'`;
    chomp $z;
    $z == $baseline or die "die Sicherung hält $z Wanderungen, 0.8.0 hatte $baseline\n";
    return $s[0];
  });
  schritt('Kandidat', 'der Dienst wurde neu gestartet und läuft wieder', sub {
    warte('ein neuer Dienst', sub { my $p = dienstpid(); $p && $p != $altpid }, 60);
    return "pid $altpid → " . dienstpid();
  });
  schritt('Kandidat', 'mit dem Programm des Kandidaten', sub {
    my $pid = dienstpid();
    my ($start) = `ps -p $pid -o lstart=`;
    my $txt = `lsof -p $pid -Fn 2>/dev/null` =~ /^n(\S*mse-rest)$/m ? $1 : '?';
    my $inode = (stat "$APP/Contents/MacOS/mse-rest")[1];
    `lsof -p $pid 2>/dev/null` =~ /\b$inode\b/ or die "der Dienst läuft nicht mit der Datei des Kandidaten ($txt)\n";
    return "gestartet $start";
  });
  schritt('Kandidat', 'er antwortet über seine neue Ansage', sub {
    warte('eine zweite Ansage', sub { (ansage())[0] >= 2 }, 60);
    my $status = antwortet();
    $status == 200 or die "Status $status\n";
    return 200;
  });
  my $neupid = dienstpid();
  schritt('Kandidat', 'ein zweiter Ruf sichert nicht und startet nichts neu', sub {
    my ($rc, $t) = mse('sets');
    $rc == 0 or die "Rückgabe $rc: $t\n";
    my @s = sicherungen();
    @s == 1 or die scalar(@s) . " Sicherungen\n";
    dienstpid() == $neupid or die "der Dienst wurde neu gestartet\n";
    return 1;
  });
  $e{wanderungen} = { vorher => $baseline, nachher => wanderungen(), zuletzt => zuletzt() };
  $e{protokoll} = substr(lesen($ANSAGE) =~ s/^Marke \S+$/Marke ***/mgr, -2000);
  1;
} or push @grenzen, "abgebrochen: $@";

raeumen();
$e{schritte} = \@schritte;
$e{grenzen} = \@grenzen;
open my $aus, '>', "$T/m02-ergebnis.json" or die "$!\n";
print {$aus} $J->encode(\%e);
close $aus;
my @rot = grep { !$_->{ok} } @schritte;
print @rot ? "mit Fehlern: " . join('; ', map { "$_->{was}: $_->{fehler}" } @rot) . "\n" : "fertig\n";
