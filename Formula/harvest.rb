class Harvest < Formula
  desc "CLI for Harvest time logging and week submission"
  homepage "https://github.com/n3d1117/harvest-cli"
  url "https://github.com/n3d1117/harvest-cli/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "647c6085729389fc438f548e3997c3bd063afd0c99e7e4b78391b21b573f287a"
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
