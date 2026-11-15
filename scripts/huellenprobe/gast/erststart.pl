#!/usr/bin/perl
# **Der Erststart aus dem Abbild** — HAND6
# (#2384),
# so weit ihn kein Mensch geben muss.
#
# Vier Teile, jeder mit seinem Ergebnis im Beleg, auch dem, der scheitert:
#
#  1. **Das Abbild wie heruntergeladen**: eine **kenntlich synthetische**
#     Quarantäne (`com.apple.quarantine` mit dem Absender `mse4mac-hand6`),
#     wie ein Browser sie setzt; eingehängt, das Bündel herauskopiert — die
#     Quarantäne erbt.
#  2. **Was Gatekeeper sagt**: `spctl`, ein `open`, das nicht startet, und die
#     Zeilen von `syspolicyd`. **Kein `xattr -d`**: die Quarantäne nimmt hier
#     niemand pauschal weg.
#  3. **Die Freigabe über *Datenschutz & Sicherheit → Dennoch öffnen*
#     verlangt das Kennwort eines Verwalters** — sie gibt ein Mensch
#     (REL, #2385), nicht dieser Treiber. Danach wie nach ihr: das Bündel ohne
#     Quarantäne ausgepackt (`ditto --noqtn`).
#  4. **Der erste Start über einem frischen, künstlichen Bestand** — eigenes
#     Zuhause, eigener Bestand, eigener Schlüsselbund —, und danach der Blick
#     auf das, was dabei entstand: Dateien, Rechte, Größen, das Schema.
#
#     perl erststart.pl <geteiltes Verzeichnis im Gast>
use strict;
use warnings;
use utf8;
use File::Path qw(make_path);
use File::Temp qw(tempdir);
use File::Find;
use FindBin;
use lib $FindBin::Bin;
use Encode qw(decode_utf8);
use JSON::PP;
use Time::HiRes qw(time sleep);

use Gast qw(lesen jetzt ort aufraeumen);

binmode STDOUT, ':encoding(UTF-8)';
binmode STDERR, ':encoding(UTF-8)';

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $D = "$T/hand6";
my $J = JSON::PP->new->utf8->canonical;

# **Ein eigenes Verzeichnis je Lauf**, und das Abbild hängt wieder ab, auch
# nach einem Abbruch.
my $W = tempdir('/tmp/hand6-XXXXXX');
make_path("$W/Programme");
my $app = "$W/Programme/MagicSetEditor4Mac.app";
END {
  system('pkill', '-f', "$W/") if $W;
  system('hdiutil', 'detach', '-quiet', '-force', "$W/mnt") if -d "$W/mnt";
  system('rm', '-rf', $W);
}

my %e = (maschine => {
  system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r,
  modell => scalar `sysctl -n hw.model` =~ s/\s+$//r, kerne => scalar `sysctl -n hw.ncpu` =~ s/\s+$//r });
my (@schritte, @grenzen);

sub schritt {
  my ($teil, $was, $tun) = @_;
  my $beginn = jetzt();
  my $wert = eval { $tun->() };
  my $ok = defined $wert && !$@;
  push @schritte, { teil => $teil, was => $was, ok => $ok ? JSON::PP::true : JSON::PP::false,
    wert => $ok ? "$wert" : ($@ =~ s/\s+$//r), ms => int(jetzt() - $beginn) };
  return $wert;
}

sub laeuft_noch { return `pgrep -f '$_[0]/Contents/MacOS/'` =~ /\d/ }

my $marke = sprintf('0083;%x;mse4mac-hand6;', time());

# --- 1. Das Abbild wie heruntergeladen
my $imAbbild;
eval {
  schritt('Abbild', 'Prüfsumme', sub {
    my $soll = (split ' ', lesen("$D/pruefsumme.txt"))[0];
    my $ist = (split ' ', `shasum -a 256 '$D/abbild.dmg'`)[0];
    $soll eq $ist or die "Prüfsumme $ist statt $soll\n";
    return substr($ist, 0, 16);
  });
  system('cp', "$D/abbild.dmg", "$W/abbild.dmg") == 0 or die "kopieren\n";
  system('xattr', '-w', 'com.apple.quarantine', $marke, "$W/abbild.dmg") == 0 or die "Quarantäne\n";
  schritt('Abbild', 'Quarantäne am Abbild', sub { `xattr -p com.apple.quarantine '$W/abbild.dmg' 2>&1` =~ s/\s+$//r });
  schritt('Abbild', 'einhängen', sub {
    system('hdiutil', 'attach', '-nobrowse', '-readonly', '-quiet', '-mountpoint', "$W/mnt", "$W/abbild.dmg") == 0 or die "hdiutil\n";
    return join ', ', map { decode_utf8(s{.*/}{}r) } glob("$W/mnt/*");
  });
  ($imAbbild) = glob("$W/mnt/*.app");
  schritt('Abbild', 'ins Programme-Verzeichnis ziehen', sub {
    system('ditto', $imAbbild, $app) == 0 or die "ditto\n";
    return 1;
  });
  $e{erbt} = schritt('Abbild', 'die Quarantäne erbt', sub {
    my $q = `xattr -p com.apple.quarantine '$app' 2>&1` =~ s/\s+$//r;
    $q =~ /mse4mac-hand6/ or die "am Bündel steht: $q\n";
    return $q;
  });
  1;
} or push @grenzen, "Abbild brach ab: $@";

# --- 2. Was Gatekeeper sagt
eval {
  schritt('Gatekeeper', 'spctl', sub {
    my $t = `spctl --assess --type execute -vv '$app' 2>&1`;
    my $rc = $? >> 8;
    $rc or die "spctl nahm an: $t\n";
    return "Rückgabe $rc: " . ($t =~ s/\s+/ /gr);
  });
  schritt('Gatekeeper', 'Siegel', sub { `codesign -dv '$app' 2>&1` =~ /(flags=\S+)/ ? $1 : die "kein Siegel\n" });
  schritt('Gatekeeper', 'ein Doppelklick startet nicht', sub {
    my $o = ort();
    system('open', '-n', '-a', $app, '--env', "HOME=$o->{heim}", '--env', "MSE_BESTAND=$o->{bestand}", '--env', "MSE_SCHLUESSELBUND=$o->{bund}");
    sleep 8;
    my $lief = laeuft_noch($app);
    system('pkill', '-f', "$app/Contents/MacOS/") if $lief;
    $lief and die "gestartet — Gatekeeper hielt nicht an\n";
    -e $o->{bestand} and die "ein Bestand entstand ohne Start\n";
    return 'nicht gestartet, kein Bestand';
  });
  $e{syspolicyd} = `log show --last 3m --style compact --predicate 'process == "syspolicyd" AND eventMessage CONTAINS[c] "MagicSetEditor4Mac"' 2>/dev/null | tail -6`;
  1;
} or push @grenzen, "Gatekeeper brach ab: $@";

push @grenzen, '»Dennoch öffnen« unter Datenschutz & Sicherheit verlangt das Kennwort eines Verwalters — gegeben von einem Menschen, nicht von diesem Lauf (REL, #2385). Teil 4 fährt das Bündel ohne Quarantäne, wie nach dieser Freigabe.';

# --- 4. Der erste Start über einem frischen Bestand
my $o = ort();
eval {
  # **Aus dem eingehängten Abbild, wie I03** (`aufstieg.pl`): ein `ditto
  # --noqtn` von der schon unter Quarantäne stehenden Kopie behielt sie.
  my $frei = "$W/Programme/Frei.app";
  system('ditto', '--noqtn', $imAbbild, $frei) == 0 or die "auspacken\n";
  schritt('Erststart', 'ohne Quarantäne', sub {
    my $l = `xattr -l '$frei' 2>&1`;
    $l =~ /com\.apple\.quarantine/ and die "noch da: " . ($l =~ s/\s+/ /gr) . "\n";
    return 'keine';
  });
  my $beginn = jetzt();
  system('open', '-n', '-a', $frei, '--env', "HOME=$o->{heim}", '--env', "MSE_BESTAND=$o->{bestand}", '--env', "MSE_SCHLUESSELBUND=$o->{bund}");
  schritt('Erststart', 'das Programm läuft', sub {
    for (1 .. 150) { return int(jetzt() - $beginn) . ' ms' if laeuft_noch($frei); sleep 0.1 }
    die "kein Prozess nach 15 s\n";
  });
  schritt('Erststart', 'der Bestand entsteht', sub {
    for (1 .. 300) { return int(jetzt() - $beginn) . ' ms' if -s $o->{bestand}; sleep 0.1 }
    die "kein Bestand nach 30 s\n";
  });
  sleep 3;
  schritt('Erststart', 'das Programm läuft noch', sub { laeuft_noch($frei) ? 1 : die "es endete von selbst\n" });
  schritt('Erststart', 'beenden', sub {
    system('pkill', '-TERM', '-f', "$frei/Contents/MacOS/");
    for (1 .. 150) { return 'beendet' unless laeuft_noch($frei); sleep 0.1 }
    die "läuft nach 15 s noch\n";
  });
  1;
} or push @grenzen, "Erststart brach ab: $@";

# --- Der Blick auf den Bestand
my @dateien;
for my $wurzel ($o->{ort}) {
  find({ no_chdir => 1, wanted => sub {
    return if $File::Find::name eq $wurzel;
    my @s = lstat($File::Find::name) or return;
    my $name = $File::Find::name =~ s{^\Q$wurzel\E/}{}r;
    return if $name =~ /^bund\.keychain-db/;
    push @dateien, { pfad => $name, rechte => sprintf('%04o', $s[2] & 07777), art => -d _ ? 'Verzeichnis' : 'Datei', groesse => -d _ ? 0 : $s[7] };
  } }, $wurzel);
}
$e{dateien} = [sort { $a->{pfad} cmp $b->{pfad} } @dateien];
$e{schema} = -e $o->{bestand} ? (`sqlite3 '$o->{bestand}' 'SELECT count(*) FROM migrationen' 2>&1` =~ s/\s+$//r) : '';
$e{info} = { map { $_ => (`defaults read '$app/Contents/Info' $_ 2>&1` =~ s/\s+$//r) } qw(CFBundleIdentifier CFBundleShortVersionString LSMinimumSystemVersion) };

$e{schritte} = \@schritte;
$e{grenzen} = \@grenzen;
open my $aus, '>', "$T/hand6-ergebnis.json" or die "Ergebnis: $!\n";
print {$aus} $J->encode(\%e);
close $aus;
aufraeumen();
exit((grep { !$_->{ok} } @schritte) ? 1 : 0);
