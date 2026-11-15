#!/usr/bin/perl
# **Trägt ein selbst ausgestellter Herausgeber die Bindung über eine
# Fassung?** — I01 (#2259).
#
# 0181 hielt die Bindung an die Programme dieser Anwendung für untragbar,
# weil die ACL eines ad hoc signierten Programms dessen **cdhash** trägt und
# jede Fassung einen neuen hat (SEC-05, E3). Die Hypothese: ist das Programm
# mit einem **eigenen Herausgeberzeugnis** signiert, verankert die ACL die
# Anforderung am Zeugnis statt am cdhash, und eine neue Fassung desselben
# Herausgebers liest still — ein **anderer** Herausgeber mit demselben
# Bezeichner aber nicht.
#
# **E5 fragt, ob die Bindung fremden Code aufhält** (für die Entscheidung
# über #2295): eine Bibliothek (`fremd.c`) wird über `DYLD_INSERT_LIBRARIES`
# in den gebundenen Leser geladen und fragt vor dessen `main` dasselbe — einmal
# in eine Fassung wie heute, einmal in eine mit gehärteter Laufzeit
# (`--options runtime`).
#
# **Nur in einem Wegwerf-Gast** und **nur an einem künstlichen Bund** mit
# Probekennwort: die Herausgeber entstehen hier und werden mit dem Bund
# gelöscht; **auf dem Wirt wird nie ein Herausgeberzeugnis eingerichtet**.
# Kein Geheimnis steht im Ergebnis — nur Längen und Status.
#
#     perl bindung.pl <geteiltes Verzeichnis im Gast>
use strict;
use warnings;
use utf8;
use JSON::PP;
use Time::HiRes qw(time);

binmode STDOUT, ':encoding(UTF-8)';
binmode STDERR, ':encoding(UTF-8)';

# **Nur in einem Gast** — die zweite Sicherung neben `gastziel` (Befund 1 des
# Zweitblicks über Phase I): dieser Treiber richtet ein, was nie auf einen
# Wirt gehört, und bricht deshalb ab, bevor er etwas anfasst.
`sysctl -n hw.model` =~ /^VirtualMac/ or die "nur in einem Gast — diese Maschine ist kein VirtualMac\n";

my $T = shift // die "geteiltes Verzeichnis fehlt\n";
my $D = "$T/i01";
my $W = "$ENV{HOME}/i01";
my $J = JSON::PP->new->utf8->canonical;
my $bund = "$W/bund.keychain-db";
my $kennung = 'de.mse4mac.i01.leser';
system('rm', '-rf', $W);
mkdir $W or die "$W: $!\n";

my %e = (maschine => {
  system => scalar `sw_vers -productVersion` =~ s/\s+$//r, bau => scalar `sw_vers -buildVersion` =~ s/\s+$//r,
  openssl => scalar `openssl version` =~ s/\s+$//r });
my (@schritte, @protokoll);

# maske nimmt Probewerte und Probekennwort aus einem Text — **an jeder
# Stelle, die in den Beleg geht**: Protokoll, Ausnahme und Grenzen (Befund 6
# des Zweitblicks über Phase I: der Fehlerpfad trug sie unmaskiert hinaus).
sub maske { return $_[0] =~ s/(\s-[wpP])\s+\S+/$1 ***/gr =~ s/pass:\S+/pass:***/gr }

# laufe fährt einen Befehl und schreibt ihn maskiert ins Protokoll.
sub laufe {
  my $befehl = shift;
  my $t = `$befehl 2>&1`;
  my $rc = $? >> 8;
  push @protokoll, '$ ' . maske($befehl), maske($t =~ s/\s+$//r), "rc=$rc";
  return ($rc, $t);
}
sub muss { my ($rc, $t) = laufe(@_); die maske("$_[0]: rc=$rc " . substr($t, 0, 200)) . "\n" if $rc; return $t }

# erwarte vergleicht einen Lesezugriff mit dem, was die Hypothese sagt —
# »liest« oder »liest nicht«. Abweichung ist ein Befund, kein Abbruch.
sub erwarte {
  my ($fall, $wer, $soll, $befehl) = @_;
  my ($rc, $t) = laufe($befehl);
  # **Der security-Starter meldet Rückgabe und Länge**, nie den Wert. Bis
  # zum ersten Lauf zählte hier `| wc -c` — und das gibt auch ohne Treffer
  # »0« mit Rückgabe 0; E2 und E4 sahen dadurch wie »liest« aus.
  # **Fremder Code meldet sich mit FREMD** — ohne die Zeile lief er gar nicht.
  my $fremdcode = $wer =~ /^fremd/;
  my $liest = ($wer eq 'security') ? ($t =~ /rc=0 laenge=(\d+)/ && $1 > 0)
    : $fremdcode ? ($t =~ /FREMD GELESEN/) : ($t =~ /^v\d GELESEN/m);
  my $ist = $liest ? 'liest'
    : ($fremdcode && $t !~ /FREMD/) ? 'läuft nicht'
    : ($t =~ ($fremdcode ? qr/FREMD STATUS -25308/ : qr/-25308/)) ? 'müsste fragen' : "bekommt nichts (rc=$rc)";
  push @schritte, { fall => $fall, wer => $wer, soll => $soll, ist => $ist,
    ok => (($soll eq 'liest') == !!$liest) ? JSON::PP::true : JSON::PP::false };
  return;
}
my $fremd = sub { "sh -c 'w=\$(security find-generic-password -s $_[0] -w \"$bund\" 2>/dev/null); echo \"rc=\$? laenge=\${#w}\"'" };
my $leser = sub { "'$W/$_[0]' $_[1]" };
# fremd startet einen Leser mit der fremden Bibliothek; er selbst fragt nach
# einem Dienst, den es nicht gibt, damit nur die Bibliothek liest.
my $fremd_in = sub { "MSE_I01_FREMD=$_[1] DYLD_INSERT_LIBRARIES='$W/fremd.dylib' '$W/$_[0]' keiner" };

my @suchliste;

# aufraeumen nimmt den Bund aus der Suchliste, löscht ihn samt Herausgebern
# und das Arbeitsverzeichnis — **auch nach einem Abbruch** (Befund 1 des
# Nachlaufs über Phase I): `END` fängt `die` und jedes Ende, die Signale
# fangen `INT` und `TERM`. **Die Suchliste wird von heute gelesen**, nicht
# nur die gemerkte: so stimmt sie auch, wenn der Lauf abbrach, bevor er sie
# sich merkte. Ein zweiter Ruf tut nichts.
my $geraeumt = 0;
sub aufraeumen {
  return if $geraeumt++;
  my @jetzt = grep { $_ ne $bund } map { /"(.+)"/ ? $1 : () } split /\n/, `security list-keychains -d user`;
  laufe('security list-keychains -d user -s ' . join(' ', map { "'$_'" } @jetzt)) if @jetzt;
  laufe("security delete-keychain '$bund'") if -e $bund;
  system('rm', '-rf', $W);
  my $noch = grep { index($_, $bund) >= 0 } `security list-keychains -d user`;
  $e{aufgeraeumt} = ($noch || -e $bund || -e $W) ? 'nein' : 'ja';
  return;
}
$SIG{INT} = $SIG{TERM} = sub { aufraeumen(); exit 130 };
END { aufraeumen() }

eval {
  # --- der künstliche Bund, kurz vorn in der Suchliste des Gastes
  muss("security create-keychain -p leer '$bund'");
  muss("security unlock-keychain -p leer '$bund'");
  muss("security set-keychain-settings '$bund'");
  # **Ohne die Reste früherer Läufe**: ein abgebrochener Lauf ließ seinen
  # Bund in der Suchliste, und »wie vorher« trug ihn bei jedem weiteren Lauf
  # wieder ein (gemessen am 2026-10-11, zweimal derselbe Pfad).
  @suchliste = grep { $_ ne $bund } map { /"(.+)"/ ? $1 : () } split /\n/, muss('security list-keychains -d user');
  muss("security list-keychains -d user -s '$bund' " . join(' ', map { "'$_'" } @suchliste));

  # --- zwei selbst ausgestellte Herausgeber
  for my $h (qw(A B)) {
    muss("openssl req -x509 -newkey rsa:2048 -nodes -days 2 -keyout '$W/$h.key' -out '$W/$h.crt' -subj '/CN=MSE4Mac I01 Herausgeber $h' -addext 'keyUsage=critical,digitalSignature' -addext 'extendedKeyUsage=critical,codeSigning'");
    muss("openssl pkcs12 -export -inkey '$W/$h.key' -in '$W/$h.crt' -out '$W/$h.p12' -passout pass:leer");
    muss("security import '$W/$h.p12' -k '$bund' -P leer -T /usr/bin/codesign");
  }

  # --- fünf Fassungen des Lesers: drei von A (eine gehärtet), eine von B,
  # eine ad hoc — und die fremde Bibliothek, ad hoc wie jede am Wirt gebaute
  my %wer = ('leser-a1' => ['leser-v1', 'MSE4Mac I01 Herausgeber A'], 'leser-a2' => ['leser-v2', 'MSE4Mac I01 Herausgeber A'],
    'leser-a3' => ['leser-v1', 'MSE4Mac I01 Herausgeber A', '--options runtime'],
    'leser-b1' => ['leser-v1', 'MSE4Mac I01 Herausgeber B'], 'leser-adhoc' => ['leser-v1', '-']);
  muss("cp '$D/fremd.dylib' '$W/fremd.dylib'");
  for my $name (sort keys %wer) {
    my ($quelle, $herausgeber, $haertung) = @{ $wer{$name} };
    muss("cp '$D/$quelle' '$W/$name'");
    muss("codesign --force --sign '$herausgeber' --keychain '$bund' " . ($haertung // '') . " -i $kennung '$W/$name'");
    $e{flags}{$name} = (muss("codesign -dv '$W/$name'") =~ /flags=(\S+)/)[0];
    $e{anforderung}{$name} = muss("codesign -d -r- '$W/$name'") =~ s/.*designated => //sr =~ s/\s+$//r;
    $e{cdhash}{$name} = (muss("codesign -dvvv '$W/$name'") =~ /CDHash=(\w+)/)[0];
    $e{gatekeeper}{$name} = (laufe("spctl --assess --type execute -vv '$W/$name'"))[1] =~ s/\s+/ /gr;
  }

  # --- E1: wie das Produkt heute schreibt — ohne -T
  muss("security add-generic-password -a i01 -s dienst1 -w geheim-eins '$bund'");
  erwarte('E1 heute', 'security', 'liest', $fremd->('dienst1'));
  erwarte('E1 heute', 'leser-a1', 'liest nicht', $leser->('leser-a1', 'dienst1'));

  # --- E2: gebunden an leser-a1
  muss("security add-generic-password -a i01 -s dienst2 -w geheim-zwei -T '$W/leser-a1' '$bund'");
  $e{acl} = muss("security dump-keychain -a '$bund'") =~ s/.*"dienst2"//sr =~ s/(requirement:[^\n]*).*/$1/sr;
  erwarte('E2 gebunden', 'security', 'liest nicht', $fremd->('dienst2'));
  erwarte('E2 gebunden', 'leser-a1', 'liest', $leser->('leser-a1', 'dienst2'));

  # --- E3: eine neue Fassung — desselben Herausgebers, eines anderen, ad hoc
  erwarte('E3 neue Fassung, Herausgeber A', 'leser-a2', 'liest', $leser->('leser-a2', 'dienst2'));
  erwarte('E3 anderer Herausgeber, gleicher Bezeichner', 'leser-b1', 'liest nicht', $leser->('leser-b1', 'dienst2'));
  erwarte('E3 ad hoc, gleicher Bezeichner', 'leser-adhoc', 'liest nicht', $leser->('leser-adhoc', 'dienst2'));

  # --- E5: fremder Code im gebundenen Programm — wie heute, und gehärtet
  erwarte('E5 fremder Code in der gebundenen Fassung', 'fremd in leser-a1', 'liest', $fremd_in->('leser-a1', 'dienst2'));
  erwarte('E5 gehärtete Fassung von A', 'leser-a3', 'liest', $leser->('leser-a3', 'dienst2'));
  erwarte('E5 fremder Code in der gehärteten Fassung', 'fremd in leser-a3', 'liest nicht', $fremd_in->('leser-a3', 'dienst2'));

  # --- E4: gesperrt
  muss("security lock-keychain '$bund'");
  erwarte('E4 gesperrt', 'security', 'liest nicht', $fremd->('dienst2'));
  erwarte('E4 gesperrt', 'leser-a1', 'liest nicht', $leser->('leser-a1', 'dienst2'));
  1;
} or push @{ $e{grenzen} }, maske("abgebrochen: $@" =~ s/\s+$//r);

# --- aufräumen: Suchliste ohne den Bund, Bund und Herausgeber fort
aufraeumen();

$e{schritte} = \@schritte;
$e{protokoll} = \@protokoll;
open my $aus, '>', "$T/i01-ergebnis.json" or die "$!\n";
print {$aus} $J->encode(\%e);
close $aus;
my @rot = grep { !$_->{ok} } @schritte;
print @rot ? "abweichend: " . join('; ', map { "$_->{fall}/$_->{wer}: $_->{ist}" } @rot) . "\n" : "fertig\n";
