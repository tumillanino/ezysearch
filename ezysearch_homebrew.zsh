class EzySearch < Formula
  desc "Universal package manager and GitHub search tool"
  homepage "https://github.com/tumillanino/ezysearch"
  url "https://github.com/yourusername/ezysearch/archive/v1.0.0.tar.gz"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000" # Replace with actual sha256
  license "MIT"
  
  depends_on "fzf"
  depends_on "zsh"
  
  # Optional dependencies
  depends_on "gh" => :recommended
  depends_on "bat" => :recommended
  depends_on "fd" => :recommended
  
  def install
    bin.install "ezysearch.zsh" => "ezysearch"
    zsh_function_dir = pkgshare/"zsh"
    zsh_function_dir.install "ezysearch.plugin.zsh"
    
    # Create completion files
    zsh_completion.install "completions/_ezysearch" => "_ezysearch"
    
    # Create default config directory
    (prefix/"config").install "config.zsh" => "config.zsh"
  end
  
  def caveats
    <<~EOS
      To use ezy_search, add the following line to your ~/.zshrc:
      
        source "#{pkgshare}/zsh/ezysearch.plugin.zsh"
      
      Configuration files are stored in:
        $HOME/.config/ezysearch/
      
      Usage:
        • Press Ctrl+P to search and install packages
        • Press Ctrl+G to search GitHub repositories
        • Press Ctrl+D to search directories
    EOS
  end
  
  test do
    assert_match "ezysearch", shell_output("#{bin}/ezysearch --version")
  end
end
