class Harvest < Formula
  desc "Simple local-first Harvest time logging CLI"
  homepage "https://github.com/n3d1117/harvest-cli"
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
