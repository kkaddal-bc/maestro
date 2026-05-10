class Maestro < Formula
  desc "Maestro CLI for installing and managing skills"
  homepage "https://github.com/kkaddal-bc/maestro"
  version "0.1.0"

  on_macos do
    on_arm do
      url "https://github.com/kkaddal-bc/maestro/releases/download/v0.1.0/maestro-darwin-arm64.tar.gz"
      sha256 "938957e5ac72f194be3bbc79d864246d51fc77354f509588ec0467204151a166"
    end

    on_intel do
      url "https://github.com/kkaddal-bc/maestro/releases/download/v0.1.0/maestro-darwin-amd64.tar.gz"
      sha256 "943e4001be2ea33ddafde94922569425e43caf933ffa98d74489926d765a01ea"
    end
  end

  def install
    bin.install "maestro"
  end

  test do
    system "#{bin}/maestro", "--help"
  end
end
