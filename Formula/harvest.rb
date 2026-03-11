class Harvest < Formula
  desc "CLI for Harvest time logging and week submission"
  homepage "https://github.com/n3d1117/harvest-cli"
  url "https://github.com/n3d1117/harvest-cli/archive/refs/tags/1.0.0.tar.gz"
  sha256 "495c43fb7aab844bf8679bcd67af7494fc38a11311cd11bb236878d86c776b92"
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
