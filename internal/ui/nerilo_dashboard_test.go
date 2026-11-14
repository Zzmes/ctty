package ui

import (
    "strings"
    "testing"

    "github.com/zsuroy/ctty/internal/config"
)

func TestNeriloAuth(t *testing.T) {
    cases := []struct {
        host config.SSHHost
        want string
    }{
        {config.SSHHost{}, "Default SSH"},
        {config.SSHHost{Identity: "~/.ssh/id_ed25519"}, "SSH Key"},
        {config.SSHHost{IdentityAgent: "SSH_AUTH_SOCK"}, "Agent"},
    }
    for _, tc := range cases {
        if got := neriloAuth(tc.host); got != tc.want {
            t.Errorf("neriloAuth(%+v) = %q, want %q", tc.host, got, tc.want)
        }
    }
}

func TestNeriloFieldTruncation(t *testing.T) {
    field := neriloField(" Host", strings.Repeat("a", 90), 30)
    if field == "" || !strings.Contains(field, "…") {
        t.Fatalf("expected truncated detail value, got %q", field)
    }
}

func TestNeriloValueFallback(t *testing.T) {
    if got := neriloValue("", "22"); got != "22" {
        t.Fatalf("empty value fallback = %q", got)
    }
    if got := neriloValue("2200", "22"); got != "2200" {
        t.Fatalf("nonempty value = %q", got)
    }
}
