# typed: false
# frozen_string_literal: true

class Chotko < Formula
  desc "Terminal UI for Zabbix monitoring"
  homepage "https://github.com/harpchad/chotko"
  url "https://github.com/harpchad/chotko/archive/refs/tags/v0.7.0.tar.gz"
  sha256 "b7a7383679f761983e03ff732d38f624a5b0d55ab63526c0c730bd7d33aaf861"
  license "MIT"
  head "https://github.com/harpchad/chotko.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = %W[
      -s -w
      -X main.version=#{version}
      -X main.commit=#{tap.user}
      -X main.date=#{time.iso8601}
    ]
    system "go", "build", *std_go_args(ldflags:), "./cmd/chotko"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/chotko --version")
  end
end
