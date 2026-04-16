#!/usr/bin/env sh
set -eu

# macOS LAN IP helper.
# Prints first non-empty IPv4 from best-guess interfaces.

get_ip() {
  # ipconfig exits non-zero when no address; silence errors.
  ipconfig getifaddr "$1" 2>/dev/null || true
}

default_iface="$(
  route get default 2>/dev/null | awk '/interface:/{print $2; exit}' || true
)"

if [ -n "${default_iface:-}" ]; then
  ip="$(get_ip "$default_iface")"
  if [ -n "${ip:-}" ]; then
    printf '%s\n' "$ip"
    exit 0
  fi
fi

for iface in en0 en1 en2 en3 bridge0; do
  ip="$(get_ip "$iface")"
  if [ -n "${ip:-}" ]; then
    printf '%s\n' "$ip"
    exit 0
  fi
done

# Fallback: scan all interfaces and return first with an address.
ifaces="$(ifconfig -l 2>/dev/null || true)"
for iface in $ifaces; do
  ip="$(get_ip "$iface")"
  if [ -n "${ip:-}" ]; then
    printf '%s\n' "$ip"
    exit 0
  fi
done

echo "failed to detect local IP (no IPv4 found)" >&2
exit 1

