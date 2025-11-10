class Ezysearch < Formula
  desc "Universal package manager and GitHub search tool written in Go"
  homepage "https://github.com/tumillanino/ezysearch"
  url "https://github.com/tumillanino/ezysearch/archive/v1.0.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000" # Replace with actual sha256
  license "MIT"
  
  depends_on "go" => :build
  
  # Runtime dependencies
  depends_on "fzf"
  
  def install
    system "go", "build", "-o", bin/"ezysearch", "cmd/ezysearch/main.go"
    
    # Install shell completion and key bindings
    bash_completion.install "shell/completion.bash" => "ezysearch"
    zsh_completion.install "shell/completion.zsh" => "_ezysearch"
  end
  
  def caveats
    <<~EOS
      To use ezysearch, run:
        ezysearch
      
      Usage:
        • Press Ctrl+P to search and install packages
        • Press Ctrl+G to search GitHub repositories
        • Press Ctrl+T to search directories
    EOS
  end
  
  test do
    assert_match "ezysearch version", shell_output("#{bin}/ezysearch --version")
  end
end