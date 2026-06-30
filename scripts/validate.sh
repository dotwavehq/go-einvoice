#!/usr/bin/env bash
# Generate the example invoice and validate it against the official EN 16931 and
# XRechnung (KoSIT) schematron. Requires: go, python3 with `saxonche` installed.
set -euo pipefail
cd "$(dirname "$0")/.."

XR_VERSION="v2.5.0"
XR_ZIP="xrechnung-3.0.2-schematron-2.5.0.zip"
cache="${XDG_CACHE_HOME:-$HOME/.cache}/go-einvoice-schematron"
cen="$cache/EN16931-CII-validation.xslt"
xr="$cache/XRechnung-CII-validation.xsl"
mkdir -p "$cache"

if [ ! -f "$cen" ]; then
	curl -sSL -o "$cen" \
		"https://raw.githubusercontent.com/ConnectingEurope/eInvoicing-EN16931/master/cii/xslt/EN16931-CII-validation.xslt"
fi
if [ ! -f "$xr" ]; then
	tmp="$(mktemp -d)"
	curl -sSL -o "$tmp/x.zip" \
		"https://github.com/itplr-kosit/xrechnung-schematron/releases/download/${XR_VERSION}/${XR_ZIP}"
	unzip -q "$tmp/x.zip" -d "$tmp"
	cp "$tmp/schematron/cii/XRechnung-CII-validation.xsl" "$xr"
	rm -rf "$tmp"
fi

out="$(mktemp -d)"
go run ./cmd/einvoice -in example/invoice.json -out "$out/invoice"
python3 scripts/validate.py "$out/invoice.xml" "$cen" "$xr"
