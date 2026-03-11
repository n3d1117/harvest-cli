class Harvest < Formula
  desc "CLI for Harvest time logging and week submission"
  homepage "https://github.com/n3d1117/harvest-cli"
  url "https://github.com/n3d1117/harvest-cli/archive/refs/tags/1.0.0.tar.gz"
  sha256 "96067bb01646c9ff7889633e759916d16f18383edaf8ba7517b09b159a7c9e05"
  license "MIT"
  head "https://github.com/n3d1117/harvest-cli.git", branch: "main"

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
