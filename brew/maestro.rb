class Maestro < Formula
  desc "Maestro CLI for installing and managing skills"
  homepage "https://github.com/kkaddal-bc/maestro"
  version "0.1.0"

  on_macos do
    on_arm do
      url "https://github.com/kkaddal-bc/maestro/releases/download/v0.1.0/maestro-darwin-arm64.tar.gz"
      sha256 "PLACEHOLDER" # update with: shasum -a 256 maestro-darwin-arm64.tar.gz
    end

    on_intel do
      url "https://github.com/kkaddal-bc/maestro/releases/download/v0.1.0/maestro-darwin-amd64.tar.gz"
      sha256 "PLACEHOLDER" # update with: shasum -a 256 maestro-darwin-amd64.tar.gz
    end
  end

  def install
    bin.install "maestro"
  end

  test do
    system "#{bin}/maestro", "--help"
  end
end
