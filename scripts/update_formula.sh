#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 2 ]]; then
  echo "usage: $0 <tag> <owner/repo>" >&2
  exit 1
fi

tag="$1"
repo="$2"
formula_path="Formula/harvest.rb"
source_url="https://github.com/${repo}/archive/refs/tags/${tag}.tar.gz"
tmp_tarball="$(mktemp)"

cleanup() {
  rm -f "$tmp_tarball"
}
trap cleanup EXIT

gh api \
  -H "Accept: application/vnd.github+json" \
  "repos/${repo}/tarball/${tag}" >"$tmp_tarball"

sha="$(shasum -a 256 "$tmp_tarball" | awk '{print $1}')"

cat >"$formula_path" <<EOF
class Harvest < Formula
  desc "CLI for Harvest time logging and week submission"
  homepage "https://github.com/${repo}"
  url "${source_url}"
  sha256 "${sha}"
  license "MIT"
  head "https://github.com/${repo}.git", branch: "main"

  depends_on "go" => :build

  def install
    cd "cli" do
      system "go", "build", *std_go_args, "./cmd/harvest"
    end
  end

  test do
    assert_match "harvest logs time to Harvest", shell_output("#{bin}/harvest help")
  end
end
EOF
