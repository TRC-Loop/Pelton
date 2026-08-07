#!/bin/sh
# Warns before a nightly package is unpacked. Nightlies are untested builds of
# the dev branch and must not be pointed at a real inbox.
#
# When the install is interactive, this asks for an explicit yes and aborts the
# install otherwise. When it is not (apt/dnf -y, a GUI package manager, an
# image build), there is no one to answer and blocking would hang the package
# manager forever, so it prints the warning and continues. Set
# PELTON_NIGHTLY_YES=1 to skip the prompt on an interactive install too.
set -e

# bold red, but only when something is there to render it; piped or captured
# output stays plain so log files do not fill up with escape sequences.
if [ -t 1 ]; then
  R=$(printf '\033[1;31m')
  W=$(printf '\033[1;97;41m')
  N=$(printf '\033[0m')
else
  R=''
  W=''
  N=''
fi

cat <<BANNER

$R  ##    ##  ####  ######  ##  ##  ######  ##      ##    ##
$R  ###   ##   ##   ##      ##  ##    ##    ##       ##  ##
$R  ####  ##   ##   ##      ##  ##    ##    ##        ####
$R  ## ## ##   ##   ####    ######    ##    ##         ##
$R  ##  ####   ##   ##      ##  ##    ##    ##         ##
$R  ##   ###   ##   ##      ##  ##    ##    ##         ##
$R  ##    ##  ####  ######  ##  ##    ##    ######     ##
$N
$W                                                                              $N
$W   HIGHLY EXPERIMENTAL BUILD. NOT A RELEASE. DO NOT USE WITH YOUR REAL INBOX.  $N
$W                                                                              $N

$R  This package is an automated nightly build from Pelton's development
  branch. It has not been reviewed, tested or released, and it is expected
  to break.

  IT CAN LOSE OR DAMAGE EMAIL. It may fail to send, send the wrong thing,
  delete messages on your server, or corrupt its local cache. Deletions on
  an email server can be permanent.

  DO NOT USE IT WITH YOUR REAL INBOX. Use a test account whose contents you
  can afford to lose entirely.

  There is NO WARRANTY OF ANY KIND and NO SUPPORT. You install and use this
  build entirely at your own risk. To the fullest extent permitted by law
  the authors and distributors of Pelton accept no liability for any loss
  or damage arising from its use.
$N
  It installs as "pelton-nightly", keeps its own separate accounts, mail and
  settings, and does not touch or replace a normally installed Pelton.

BANNER

if [ -n "$PELTON_NIGHTLY_YES" ]; then
  exit 0
fi

if [ ! -t 0 ]; then
  echo "  (non-interactive install, continuing; the warning above still applies)"
  echo
  exit 0
fi

printf '%s  Type "yes" to install this nightly build: %s' "$R" "$N"
read -r answer
case "$answer" in
  yes | YES | Yes)
    exit 0
    ;;
  *)
    echo "  Aborted. Install a normal release from https://pelton.app instead."
    exit 1
    ;;
esac
