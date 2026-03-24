class Harvest < Formula
  desc "CLI for Harvest time logging and week submission"
  homepage "https://github.com/n3d1117/harvest-cli"
  url "https://github.com/n3d1117/harvest-cli/archive/refs/tags/v1.0.1.tar.gz"
  sha256 "2f43912736131764fcf76af4210ab3cf37a3375d45f8af60e7ddeb0f896480d4"
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
