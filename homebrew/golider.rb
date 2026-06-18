class Golider < Formula
  desc "AI-era Go backend scaffolding with production defaults"
  homepage "https://github.com/2859044775/Golider"
  url "https://github.com/2859044775/Golider/archive/refs/tags/v0.3.1.tar.gz"
  sha256 "5ae1584c182283e04102a849d1292bdb5fe94f0572434e4436423241df0381fe"
  license "MIT"
  head "https://github.com/2859044775/Golider.git"

  depends_on "go" => :build

  def install
    system "go", "build", "-ldflags",
           "-X github.com/2859044775/Golider/cmd.version=#{version}",
           "-o", bin/"golider", "."
  end

  test do
    system bin/"golider", "version"
  end
end
