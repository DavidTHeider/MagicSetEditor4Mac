#!/usr/bin/perl
# **Zehn Neustarts unter launchd, eine Marke** — SEC02
# (#2355),
# der Teil, den der Arbeitsrechner nicht tragen darf: ein LaunchAgent ist
# dauerhafte Einrichtung, und hier ist er Wegwerfware.
#
#  1. **Eingerichtet über das Fenster** — die Bindung `SchnittstelleEinrichten`
#     mit `dienst`, also die plist, die die Anwendung schreibt: `-ansage` als
#     Fahne, die Standardausgabe ins Leere, die Fehlerausgabe nach
#     `dienst.log`.
#  2. **Die erste Ansage**: genau eine Marke, die Datei 0600, der Ordner 0700,
#     und die Marke öffnet den Dienst.
#  3. **Zehn Neustarts mit `launchctl kickstart -k`**: je Start eine neue
#     Marke und nur sie in der Datei; die vorige weist der neue Dienst ab.
#  4. **`dienst.log`** trägt keine der elf Marken, und neben Ansage und
#     Protokoll entsteht keine Datei.
#  5. **Widerrufen über das Fenster**: der Dienst entladen, plist und Ansage
#     fort.
#
# **Gefragt wird die Bindung, nicht die Seite** — hinter dem Sperrbildschirm
# des Gastes ist die Seite verborgen; wie die Seite *Schnittstellen*
# aussieht, belegt `ops03` am Wirt.
#
#     perl sec02.pl <geteiltes Verzeichnis im Gast>
use strict;
use warnings;
use utf8;
use File::Temp qw(tempdir);
use FindBin;
use lib $FindBin::Bin;
use JSON::PP;
use Time::HiRes qw(sleep);

use Gast qw(lesen jetzt buendel messungen ort seite warte starte beende);

binmode STDOUT, ':encoding(UTF-8)';
binmode STDERR, ':encoding(UTF-8)';

`sysctl -n hw.model` =~ /^VirtualMac/ or die "nur in einem Gast — diese Maschine ist kein VirtualMac\n";

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $J = JSON::PP->new->utf8->canonical;
messungen($J->decode(lesen("$T/sec02/messungen.json")));

my $bau = tempdir('/tmp/sec02-abnahme-XXXXXX', CLEANUP => 1);
my $APP = "$bau/Abnahme.app";
system('ditto', '--noqtn', "$T/sec02/Abnahme.app", $APP) == 0 or die "Bündel\n";
buendel($APP);

my $uid = $<;
my $ETIKETT = 'io.github.davidtheider.magicseteditor4mac.dienst';
my %e = (maschine => {
  system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r,
  modell => scalar `sysctl -n hw.model` =~ s/\s+$//r, kerne => scalar `sysctl -n hw.ncpu` =~ s/\s+$//r });
my (@schritte, @ausgaben);

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

# bindung ruft eine Bindung des Fensters über seine Naht.
sub bindung {
  my ($name, @argumente) = @_;
  my $args = $J->encode(\@argumente);
  return seite(qq{const antwort = await fetch(window.location.origin + '/wails/runtime', { method: 'POST',
    headers: { 'x-wails-client-id': 'sec02gast', 'Content-Type': 'application/json' },
    body: JSON.stringify({ object: 0, method: 0, args: { 'call-id': 'sec02-' + String(Date.now()),
      methodName: 'github.com/DavidTHeider/MagicSetEditor4Mac/internal/app.Bindungen.$name', args: $args } }) });
    return await antwort.json()});
}

sub lage {
  my $text = `launchctl print gui/$uid/$ETIKETT 2>&1`;
  my %l = (code => $? >> 8);
  $l{zustand} = $1 if $text =~ /^\tstate = (.+)$/m;
  $l{laeufe} = $1 if $text =~ /^\truns = (\d+)$/m;
  return \%l;
}

sub rechte { my @s = stat $_[0] or die "$_[0]: $!\n"; return sprintf '%04o', $s[2] & 07777 }

# ansage liest die Datei und gibt Adresse und Marke — nur, wenn genau eine
# Marke darin steht.
sub ansage {
  my ($datei) = @_;
  return undef unless -e $datei;
  my $text = lesen($datei);
  my @marken = $text =~ /^Marke (\S+)$/mg;
  die "in der Ansage stehen " . scalar(@marken) . " Marken\n" if @marken > 1;
  my ($adresse) = $text =~ /^Dienst auf (\S+)$/m;
  return @marken && $adresse ? { adresse => $adresse, marke => $marken[0] } : undef;
}

# status fragt den Dienst mit einer Marke — der HTTP-Code, 000 ohne Antwort.
sub status {
  my ($adresse, $marke) = @_;
  my $code = `curl -s -o /dev/null -m 5 -w '%{http_code}' -H 'Authorization: Bearer $marke' '$adresse/sets'`;
  return $code;
}

eval {
  my $o = ort();
  my $eigene = "$o->{heim}/Library/Application Support/MagicSetEditor4Mac/schnittstellen";
  my $datei = "$eigene/dienst.txt";
  my $log = "$eigene/dienst.log";
  my $plist = "$o->{heim}/Library/LaunchAgents/$ETIKETT.plist";
  my @marken;

  schritt(1, 'eingerichtet über das Fenster: -ansage als Fahne, Ausgabe ins Leere', sub {
    starte($o);
    my $a = bindung('SchnittstelleEinrichten', 'dienst');
    my ($d) = grep { $_->{name} eq 'dienst' } @{ $a->{schnittstellen} // [] };
    !$a->{fehler} && $d && $d->{stand} eq 'eingerichtet' or die "Antwort: " . $J->encode($a) . "\n";
    my $inhalt = lesen($plist);
    $inhalt =~ m{<string>-ansage</string>\s*<string>\Q$datei\E</string>} or die "keine Fahne in der plist\n";
    $inhalt =~ m{<key>StandardOutPath</key>\s*<string>/dev/null</string>} or die "die Standardausgabe geht nicht ins Leere\n";
    $inhalt =~ m{<key>StandardErrorPath</key>\s*<string>\Q$log\E</string>} or die "die Fehlerausgabe geht nicht nach dienst.log\n";
    push @ausgaben, "plist der Anwendung:\n$inhalt";
    return 'eingerichtet';
  });

  schritt(2, 'die erste Ansage: eine Marke, 0600 im 0700-Ordner, und sie öffnet den Dienst', sub {
    my $an = warte('die Ansage', sub { ansage($datei) }, 30, 0.2);
    rechte($datei) eq '0600' or die "dienst.txt hat " . rechte($datei) . "\n";
    rechte($eigene) eq '0700' or die "der Ordner hat " . rechte($eigene) . "\n";
    my $mit = status($an->{adresse}, $an->{marke});
    $mit eq '200' or die "mit der Marke: $mit\n";
    my $ohne = status($an->{adresse}, 'falsch');
    $ohne eq '401' or die "mit falscher Marke: $ohne\n";
    push @marken, $an->{marke};
    push @ausgaben, "Start 1: $an->{adresse}, mit Marke $mit, mit falscher $ohne, " . join(', ', map { "$_ " . lage()->{$_} } qw(zustand laeufe));
    return "$an->{adresse}, 200 und 401";
  });

  schritt(3, 'zehn Neustarts mit kickstart -k: je Start eine neue Marke, die alte abgewiesen', sub {
    for my $start (2 .. 11) {
      my $vorher = $marken[-1];
      system('launchctl', 'kickstart', '-k', "gui/$uid/$ETIKETT") == 0 or die "kickstart: $?\n";
      my $an = warte("Start $start sagt an", sub { my $a = ansage($datei); $a && $a->{marke} ne $vorher ? $a : undef }, 30, 0.2);
      grep({ $_ eq $an->{marke} } @marken) and die "Start $start: eine Marke von früher\n";
      my $text = lesen($datei);
      for my $alte (@marken) { index($text, $alte) < 0 or die "Start $start: eine alte Marke steht noch darin\n" }
      rechte($datei) eq '0600' or die "Start $start: dienst.txt hat " . rechte($datei) . "\n";
      my $neu = warte("Start $start antwortet", sub { my $s = status($an->{adresse}, $an->{marke}); $s eq '200' ? $s : undef }, 15, 0.2);
      my $alt = status($an->{adresse}, $vorher);
      $alt eq '401' or die "Start $start: die vorige Marke bekam $alt\n";
      push @marken, $an->{marke};
      push @ausgaben, "Start $start: $an->{adresse}, neue Marke $neu, vorige $alt, Läufe " . (lage()->{laeufe} // '?');
    }
    return scalar(@marken) . ' verschiedene Marken, je eine in der Datei';
  });

  schritt(4, 'dienst.log trägt keine Marke, und daneben entsteht nichts', sub {
    -e $log or die "dienst.log fehlt\n";
    my $text = lesen($log);
    for my $marke (@marken) { index($text, $marke) < 0 or die "eine Marke steht in dienst.log\n" }
    opendir my $d, $eigene or die "$eigene: $!\n";
    my @da = sort grep { !/^\.\.?$/ } readdir $d;
    closedir $d;
    "@da" eq 'dienst.log dienst.txt' or die "im Ordner: @da\n";
    push @ausgaben, "dienst.log, " . length($text) . " Bytes:\n$text";
    return length($text) . ' Bytes, keine Marke; im Ordner nur dienst.log und dienst.txt';
  });

  schritt(5, 'widerrufen über das Fenster: der Dienst entladen, plist und Ansage fort', sub {
    my $letzte = ansage($datei);
    my $a = bindung('SchnittstelleWiderrufen', 'dienst');
    my ($d) = grep { $_->{name} eq 'dienst' } @{ $a->{schnittstellen} // [] };
    !$a->{fehler} && $d && $d->{stand} eq 'offen' or die "Antwort: " . $J->encode($a) . "\n";
    lage()->{code} == 113 or die "launchd kennt ihn noch: " . lage()->{code} . "\n";
    !-e $plist or die "die plist steht noch\n";
    !-e $datei or die "die Ansage steht noch\n";
    my $s = status($letzte->{adresse}, $letzte->{marke});
    $s eq '000' or die "die letzte Marke bekam noch $s\n";
    return 'entladen, plist und Ansage fort, die letzte Marke ohne Antwort';
  });

  beende();
};
my $fehler = $@;

# **Aufräumen, was der Lauf am Gast hinterließ** — auch nach einem Fehler.
system("launchctl bootout gui/$uid/$ETIKETT >/dev/null 2>&1");
$e{schritte} = \@schritte;
$e{ausgaben} = \@ausgaben;
$e{fehler} = $fehler =~ s/\s+$//r if $fehler;
open my $aus, '>:raw', "$T/sec02-ergebnis.json" or die "Ergebnis: $!\n";
print $aus $J->encode(\%e);
close $aus;
system('sync');
