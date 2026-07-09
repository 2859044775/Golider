class Golider < Formula
  desc "AI-era Go backend scaffolding with production defaults"
  homepage "https://github.com/2859044775/Golider"
  url "https://github.com/2859044775/Golider/archive/refs/tags/v0.6.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"
  license "MIT"
  head "https://github.com/2859044775/Golider.git"

  depends_on "go" => :build

  def install
    system "go", "build", "-ldflags",
           "-X github.com/2859044775/Golider/cmd.version=#{version}",
           "-o", bin/"Golider", "."
  end

  test do
    system bin/"Golider", "version"
  end
end
