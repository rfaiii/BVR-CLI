class BvrCli < Formula
  desc "BVR is a keyboard-first terminal AI workspace"
  homepage "https://bvr-cli.com"
  url "https://github.com/richavery/bvr-cli-main/archive/refs/tags/v#{version}.tar.gz"
  version "1.1.9"
  license "FSL-1.1-MIT"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X github.com/richavery/bvr-cli/internal/version.Version=#{version}"
    build_args = std_go_args(ldflags: ldflags, output: bin/"bvr-cli")
    system "go", "build", *build_args, "."

    generate_completions_from_executable(bin/"bvr-cli", "completion")
    man1.install Utils.safe_popen_read(bin/"bvr-cli", "man").pipe("gzip -c"), "#{name}.1.gz"
  end

  test do
    assert_match "BVR", shell_output("#{bin}/bvr-cli --version")
  end
end
