class BvrCli < Formula
  desc "BVR is a keyboard-first terminal AI workspace"
  homepage "https://bvr-cli.com"
  url "https://github.com/rfaiii/BVR-CLI/archive/refs/tags/v1.2.8.tar.gz"
  sha256 "7be0e100b00fffe49b881ee639c21d563432ee6df3c9a2848f37b4439a660874"
  license "FSL-1.1-MIT"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X github.com/richavery/bvr-cli/internal/version.Version=#{version}"
    build_args = std_go_args(ldflags: ldflags, output: bin/"bvr-cli")
    system "go", "build", *build_args, "."

    generate_completions_from_executable(bin/"bvr-cli", "completion")

    require "zlib"
    (man1/"#{name}.1.gz").write Zlib.gzip(Utils.safe_popen_read(bin/"bvr-cli", "man"))
  end

  test do
    # --version prints only the plain version line when stdout is not a TTY,
    # because the mascot is gated behind the colorprofile writer.
    assert_match version.to_s, shell_output("#{bin}/bvr-cli --version")
  end
end
