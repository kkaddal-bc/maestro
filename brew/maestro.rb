class Maestro < Formula
  desc "Maestro CLI for installing and managing skills"
  homepage "https://github.com/kkaddal-bc/maestro"
  version "0.1.0"

  on_macos do
    on_arm do
      url "https://github.com/kkaddal-bc/maestro/releases/download/v0.1.0/maestro-darwin-arm64.tar.gz"
      sha256 "a2c821f2d54a019a03d72e525c958a6a6e0120471919cef4ab10011541a8b389"
    end

    on_intel do
      url "https://github.com/kkaddal-bc/maestro/releases/download/v0.1.0/maestro-darwin-amd64.tar.gz"
      sha256 "c87807123af9fc03d75ff02052d0509c96ca28b6b8e76876ade35578b8dd0aa3"
    end
  end

  def install
    bin.install "maestro"
  end

  test do
    system "#{bin}/maestro", "--help"
  end
end
