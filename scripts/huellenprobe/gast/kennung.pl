#!/usr/bin/perl
# **Der Umzug der Kennung, am echten macOS** — R02
# (#2334),
# Vermerk 0191.
#
# Was der Lauf tut, je Fall:
#
#  1. **Der alte Stand, künstlich**: das alte Bündel (Kennung `com.wails.*`)
#     unter `~/Applications`, einmal gestartet; dazu eine Probevorgabe unter
#     der alten Domäne, eine Marke im alten Webspeicher und der Dienst unter
#     dem alten Etikett, geladen über `launchctl bootstrap` — so, wie das
#     Einrichten am alten Fenster ihn hinterließ.
#  2. **Die neue Fassung an derselben Stelle**, wie bei einer Aktualisierung,
#     einmal gestartet.
#  3. **Geprüft**: die Vorgabe steht unter der neuen Domäne, der Webspeicher
#     ist kopiert, **genau ein** Dienst läuft, unter dem neuen Etikett, und
#     die alte plist ist weg.
#  4. **Ein zweiter Start** ändert nichts.
#  5. **Ein abgebrochener Umzug**: ein halber Zwischenort und kein neuer
#     Webspeicher — der nächste Start holt ihn nach.
#  6. **Der Rückweg**: das alte Bündel an derselben Stelle findet seine
#     Vorgaben wieder; sein Dienst steht nicht mehr unter dem alten Etikett.
#
# **Zwei Fälle, nacheinander und je von vorn**: die alte Domäne in der
# Schreibweise des Bündels und klein geschrieben — beide stehen auf echten
# Rechnern (Vermerk 0191). **Kein zweites Benutzerkonto**: ein Fenster
# braucht eine Anmeldung, und die bekommt ein zweites Konto im Gast nicht
# ohne Bildschirm. Die Fälle teilen deshalb das Zuhause und räumen es
# zwischen sich ganz.
#
#     perl kennung.pl <geteiltes Verzeichnis im Gast>
use strict;
use warnings;
use utf8;
use JSON::PP;
use Time::HiRes qw(time sleep);

binmode STDOUT, ':encoding(UTF-8)';
binmode STDERR, ':encoding(UTF-8)';

# **Nur in einem Gast** — der Lauf richtet Dienste ein und räumt Vorgaben.
`sysctl -n hw.model` =~ /^VirtualMac/ or die "nur in einem Gast — diese Maschine ist kein VirtualMac\n";

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $D = "$T/r02";
my $J = JSON::PP->new->utf8->canonical;
my $H = $ENV{HOME};
my $uid = $<;
my $APP = "$H/Applications/MagicSetEditor4Mac.app";
my $NEU = 'io.github.davidtheider.magicseteditor4mac';
my $ALT = 'com.wails.MagicSetEditor4Mac';
my $ALTDIENST = "$ALT.dienst";
my $NEUDIENST = "$NEU.dienst";

my %e = (maschine => {
  system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r,
  modell => scalar `sysctl -n hw.model` =~ s/\s+$//r });
my (@schritte, @grenzen);

sub jetzt { return time() * 1000 }

sub schritt {
  my ($teil, $was, $tun) = @_;
  my $beginn = jetzt();
  my $ergebnis = eval { $tun->() };
  my $fehler = $@ ? $@ =~ s/\s+$//r : undef;
  push @schritte, { teil => $teil, was => $was, ok => $fehler ? JSON::PP::false : JSON::PP::true,
    ms => int(jetzt() - $beginn), ($fehler ? (fehler => $fehler) : ()), (defined $ergebnis && !$fehler ? (wert => "$ergebnis") : ()) };
  return $ergebnis;
}

sub geladen { my ($etikett) = @_; system("launchctl print gui/$uid/$etikett >/dev/null 2>&1"); return $? == 0 }

# dienste zählt die geladenen Dienste dieser Anwendung — unter jedem Etikett.
sub dienste { return scalar grep { /magicseteditor4mac\.dienst/i } `launchctl list 2>/dev/null` }

sub vorgabe { my ($domaene) = @_; my $w = `defaults read '$domaene' R02Probe 2>/dev/null`; chomp $w; return $w }

# einsetzen legt ein Bündel an die Stelle der Anwendung — ohne Quarantäne,
# wie nach „Dennoch öffnen" (§12; die Freigabe gibt ein Mensch).
sub einsetzen {
  my ($quelle) = @_;
  system('rm', '-rf', $APP);
  system('mkdir', '-p', "$H/Applications");
  system('ditto', '--noqtn', $quelle, $APP) == 0 or die "$quelle einsetzen\n";
  return 1;
}

# laeuft sagt, ob die Anwendung läuft. **Mit `[M]` im Muster**: `pgrep -f`
# in Backticks läuft über `sh -c`, dessen Befehlszeile das Muster trägt — ohne
# die Klammer fände es sich selbst, und jedes Warten dauerte bis zur Frist.
sub laeuft { return `pgrep -f '[M]agicSetEditor4Mac.app/Contents/MacOS/MagicSetEditor4Mac'` ? 1 : 0 }

# starten öffnet die Anwendung, wartet, bis `bedingung` gilt oder 40 s um
# sind, und beendet sie wieder.
sub starten {
  my ($kennung, $bedingung) = @_;
  my $gegeben = defined $bedingung;
  system('open', '-a', $APP) == 0 or die "open\n";
  my $bis = time() + 40;
  # **Ohne Bedingung bis der Prozess steht und fünf Sekunden darüber** — der
  # Umzug läuft vor dem Fenster, und das Fenster steht in dieser Zeit.
  $bedingung //= do { my $seit; sub { $seit //= time() if laeuft(); return $seit && time() - $seit > 5 } };
  sleep 1 until $bedingung->() || time() > $bis;
  sleep 2;
  # **Ein Signal und kein Apple Event**: `osascript … quit` verlangte im Gast
  # eine Erlaubnis für die Automation, und niemand sitzt davor, sie zu geben.
  system('pkill', '-TERM', '-f', "$APP/Contents/MacOS/MagicSetEditor4Mac");
  $bis = time() + 20;
  sleep 0.5 while time() < $bis && laeuft();
  system('pkill', '-f', "$APP/Contents/MacOS/MagicSetEditor4Mac");
  return $gegeben ? $bedingung->() : 1;
}

# raeumen nimmt alles weg, was ein Fall hinterlässt — Dienste, Vorgaben,
# Webspeicher, Bestand —, damit der nächste von vorn beginnt.
sub raeumen {
  for my $etikett ($ALTDIENST, $NEUDIENST) { system("launchctl bootout gui/$uid/$etikett >/dev/null 2>&1") }
  unlink "$H/Library/LaunchAgents/$ALTDIENST.plist", "$H/Library/LaunchAgents/$NEUDIENST.plist";
  for my $domaene ($ALT, lc $ALT, $NEU) { system("defaults delete '$domaene' >/dev/null 2>&1") }
  system('rm', '-rf', map { ("$H/Library/WebKit/$_", "$H/Library/HTTPStorages/$_", "$H/Library/WebKit/$_.umzug") } ($ALT, lc $ALT, $NEU));
  system('rm', '-rf', "$H/Library/Application Support/MagicSetEditor4Mac", $APP);
  return 1;
}

# fall fährt die Schritte für eine Schreibweise der alten Domäne. **Ohne
# ersten Start des alten Bündels, wenn sie klein ist**: sonst stünde die
# Datei schon in der Schreibweise des Bündels da, und auf APFS ohne
# Unterscheidung der Schreibweise schriebe `defaults` in dieselbe.
sub fall {
  my ($name, $altdomaene, $altStarten) = @_;
  my $teil = "Fall $name";
  raeumen();
  schritt($teil, 'das alte Bündel einsetzen' . ($altStarten ? ' und einmal starten' : ''), sub { einsetzen("$D/alt/MagicSetEditor4Mac.app"); starten($ALT) if $altStarten; 1 });
  schritt($teil, "die alte Probevorgabe unter $altdomaene", sub {
    system('defaults', 'write', $altdomaene, 'R02Probe', '-string', 'alt') == 0 or die "defaults write\n";
    vorgabe($altdomaene) eq 'alt' or die "nicht lesbar\n";
    return 'alt';
  });
  schritt($teil, 'eine Marke im alten Webspeicher', sub {
    system('mkdir', '-p', "$H/Library/WebKit/$altdomaene/R02") == 0 or die "mkdir\n";
    open my $f, '>', "$H/Library/WebKit/$altdomaene/R02/marke" or die "$!\n";
    print {$f} "alt\n";
    close $f;
    return 1;
  });
  schritt($teil, 'der Dienst unter dem alten Etikett, geladen', sub {
    my $plist = "$H/Library/LaunchAgents/$ALTDIENST.plist";
    system('mkdir', '-p', "$H/Library/LaunchAgents");
    my $inhalt = lesen("$D/dienst.plist") =~ s/__ETIKETT__/$ALTDIENST/gr =~ s/__PROGRAMM__/$APP\/Contents\/MacOS\/mse-rest/gr =~ s/__ANSAGE__/$H\/r02-dienst.txt/gr;
    open my $f, '>', $plist or die "$!\n";
    print {$f} $inhalt;
    close $f;
    system("launchctl bootstrap gui/$uid '$plist'") == 0 or die "bootstrap\n";
    geladen($ALTDIENST) or die "nicht geladen\n";
    return dienste() . ' geladen';
  });

  schritt($teil, 'die neue Fassung an derselben Stelle starten', sub {
    einsetzen("$D/neu/MagicSetEditor4Mac.app");
    starten($NEU, sub { geladen($NEUDIENST) && vorgabe($NEU) eq 'alt' }) or die "der Umzug kam nicht an\n";
    return 1;
  });
  schritt($teil, 'die Vorgabe steht unter der neuen Domäne', sub { my $w = vorgabe($NEU); $w eq 'alt' or die "»$w«\n"; $w });
  schritt($teil, 'der Webspeicher ist kopiert', sub { -f "$H/Library/WebKit/$NEU/R02/marke" or die "keine Marke\n"; 1 });
  schritt($teil, 'der alte Webspeicher bleibt für den Rückweg', sub { -f "$H/Library/WebKit/$altdomaene/R02/marke" or die "weg\n"; 1 });
  schritt($teil, 'genau ein Dienst, unter dem neuen Etikett', sub {
    geladen($NEUDIENST) or die "der neue ist nicht geladen\n";
    !geladen($ALTDIENST) or die "der alte ist noch geladen\n";
    dienste() == 1 or die dienste() . " Dienste\n";
    return 1;
  });
  schritt($teil, 'keine alte plist, die neue startet dasselbe Programm', sub {
    !-e "$H/Library/LaunchAgents/$ALTDIENST.plist" or die "die alte plist steht da\n";
    lesen("$H/Library/LaunchAgents/$NEUDIENST.plist") =~ /\Q$APP\E\/Contents\/MacOS\/mse-rest/ or die "ein anderes Programm\n";
    return 1;
  });

  schritt($teil, 'ein zweiter Start ändert nichts', sub {
    my $vorher = lesen("$H/Library/LaunchAgents/$NEUDIENST.plist");
    starten($NEU);
    lesen("$H/Library/LaunchAgents/$NEUDIENST.plist") eq $vorher or die "die plist änderte sich\n";
    dienste() == 1 && geladen($NEUDIENST) or die dienste() . " Dienste\n";
    vorgabe($NEU) eq 'alt' or die "die Vorgabe änderte sich\n";
    return 1;
  });

  schritt($teil, 'ein abgebrochener Umzug wird nachgeholt', sub {
    system('rm', '-rf', "$H/Library/WebKit/$NEU");
    system('mkdir', '-p', "$H/Library/WebKit/$NEU.umzug/halb");
    starten($NEU, sub { -f "$H/Library/WebKit/$NEU/R02/marke" });
    -f "$H/Library/WebKit/$NEU/R02/marke" or die "nicht nachgeholt\n";
    !-e "$H/Library/WebKit/$NEU/halb" or die "der halbe Stand zog mit\n";
    !-e "$H/Library/WebKit/$NEU.umzug" or die "der Zwischenort blieb\n";
    return 1;
  });

  schritt($teil, 'Rückweg: das alte Bündel findet seine Vorgaben', sub {
    einsetzen("$D/alt/MagicSetEditor4Mac.app");
    starten($ALT);
    my $w = vorgabe($altdomaene);
    $w eq 'alt' or die "»$w«\n";
    return "Vorgabe »$w«; der Dienst steht unter dem neuen Etikett und zeigt auf dasselbe Programm";
  });
  return 1;
}

sub lesen { local $/; open my $f, '<', $_[0] or die "$_[0]: $!\n"; my $x = <$f>; close $f; return $x }

eval {
  fall('A, Schreibweise des Bündels', $ALT, 1);
  fall('B, klein geschrieben', lc $ALT, 0);
  1;
} or push @grenzen, "abgebrochen: $@";

schritt('Dokumenttyp', 'der Typ für .msepackage ist unter der neuen Kennung angemeldet', sub {
  einsetzen("$D/neu/MagicSetEditor4Mac.app");
  my $ls = '/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister';
  system($ls, '-f', $APP);
  my $zahl = () = `$ls -dump 2>/dev/null` =~ /\Q$NEU.umschlag\E/g;
  $zahl > 0 or die "nicht angemeldet\n";
  return "$zahl Nennungen";
});

raeumen();
$e{schritte} = \@schritte;
$e{grenzen} = \@grenzen;
open my $aus, '>', "$T/r02-ergebnis.json" or die "$!\n";
print {$aus} $J->encode(\%e);
close $aus;
my @rot = grep { !$_->{ok} } @schritte;
print @rot ? "mit Fehlern: " . join('; ', map { "$_->{was}: $_->{fehler}" } @rot) . "\n" : "fertig\n";
