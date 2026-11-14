package ui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/x/ansi"
    "github.com/zsuroy/ctty/internal/config"
)

// NERILO's dashboard is a presentation-only layer over the existing SSH list
// model. Keyboard handling, search, sorting, and connection commands stay
// untouched. The classic list remains the fallback for smaller terminals.
func (m Model) renderNeriloDashboard() string {
    const sidebarWidth = 20
    const inspectorWidth = 33
    contentWidth := m.width - sidebarWidth - inspectorWidth
    if contentWidth < 40 {
        return m.styles.App.Render(m.renderTableView())
    }

    bg := lipgloss.Color("#0B111B")
    sidebarBg := lipgloss.Color("#101723")
    panelBg := lipgloss.Color("#121B29")
    border := lipgloss.Color("#344052")
    fg := lipgloss.Color("#C5D0E2")
    muted := lipgloss.Color("#7E90A9")
    accent := lipgloss.Color("#91B7F8")
    purple := lipgloss.Color("#C2B2FF")

    title := lipgloss.NewStyle().Foreground(accent).Bold(true)
    dim := lipgloss.NewStyle().Foreground(muted)
    strong := lipgloss.NewStyle().Foreground(fg).Bold(true)
    strip := lipgloss.NewStyle().Foreground(fg).Background(panelBg).Padding(0, 1).
        Width(m.width).Render(title.Render("◇ NERILO") + dim.Render("  /  Remote Workspace"))
    tabs := lipgloss.NewStyle().Foreground(muted).Background(bg).Padding(0, 1).
        Width(m.width).Render(title.Render("SSH") + "   Serial   Telnet   FTP   WebDAV   S3")
    availableHeight := m.height - 5
    if availableHeight < 12 {
        availableHeight = 12
    }

    // Sidebar: intentionally static navigation labels. Protocol switching
    // keeps its established keyboard bindings and is not simulated by this UI.
    sidebarLines := []string{
        title.Render(" N  nerilo"),
        "",
        dim.Render(" WORKSPACE"),
        title.Render(" ▸ SSH Hosts"),
        "   Serial",
        "   Telnet",
        "   FTP / WebDAV",
        "   S3 Storage",
        "",
        dim.Render(" CONNECTIONS"),
        fmt.Sprintf("   Hosts   %d", len(m.hosts)),
        "",
        dim.Render(" SHORTCUTS"),
        "   /   Search",
        "   ↵   Connect",
        "   ?   Help",
    }
    sidebar := lipgloss.NewStyle().Width(sidebarWidth-1).Height(availableHeight).
        Background(sidebarBg).Foreground(fg).
        Border(lipgloss.NormalBorder(), false, true, false, false).
        BorderForeground(border).Render(strings.Join(sidebarLines, "\n"))

    hosts := m.filteredHosts
    if hosts == nil {
        hosts = m.hosts
    }
    cursor := m.table.Cursor()
    if cursor < 0 {
        cursor = 0
    }
    if len(hosts) > 0 && cursor >= len(hosts) {
        cursor = len(hosts) - 1
    }

    listWidth := contentWidth - 4
    if listWidth < 12 {
        listWidth = 12
    }
    centerLines := []string{
        strong.Render(" Connections") + " " + dim.Render(fmt.Sprintf("(%d)", len(hosts))),
        dim.Render(" Manage your remote environments"),
        "",
        title.Render(" ⌕ ") + ansi.Truncate(m.searchInput.Value(), listWidth-6, "") +
            dim.Render("  / search"),
        "",
        dim.Render(" HOSTS"),
    }
    maxRows := availableHeight - len(centerLines) - 2
    if maxRows < 1 {
        maxRows = 1
    }
    start := 0
    if cursor >= maxRows {
        start = cursor - maxRows + 1
    }
    if start > len(hosts)-maxRows {
        start = len(hosts)-maxRows
    }
    if start < 0 {
        start = 0
    }
    end := start + maxRows
    if end > len(hosts) {
        end = len(hosts)
    }
    for i := start; i < end; i++ {
        host := hosts[i]
        indicator := m.getPingStatusIndicator(host.Name)
        name := ansi.Truncate(host.Name, listWidth-9, "…")
        row := fmt.Sprintf(" %s %s", indicator, name)
        if i == cursor {
            row = lipgloss.NewStyle().Background(lipgloss.Color("#263751")).
                Foreground(lipgloss.Color("#E3EAFA")).Bold(true).
                Width(listWidth).Render("▸" + ansi.Truncate(row, listWidth-1, ""))
        } else {
            row = ansi.Truncate(row, listWidth, "")
        }
        centerLines = append(centerLines, row)
    }
    if len(hosts) == 0 {
        centerLines = append(centerLines, dim.Render(" No matching hosts"))
    }
    if m.showingError && m.errorMessage != "" {
        centerLines = append(centerLines, lipgloss.NewStyle().
            Foreground(lipgloss.Color("#F28B9D")).
            Render(ansi.Truncate(m.errorMessage, listWidth, "…")))
    } else if m.statusActive() {
        centerLines = append(centerLines, dim.Render(ansi.Truncate(m.statusMessage, listWidth, "…")))
    }
    center := lipgloss.NewStyle().Width(contentWidth-2).Height(availableHeight).
        Padding(0, 1).Background(bg).Render(strings.Join(centerLines, "\n"))

    // Inspector uses only existing configuration data: it never fabricates
    // online status, latency, or server CPU/memory statistics.
    detail := []string{dim.Render(" CONNECTION DETAILS"), ""}
    if len(hosts) > 0 {
        h := hosts[cursor]
        detail = append(detail,
            title.Render(" "+ansi.Truncate(h.Name, inspectorWidth-4, "…")),
            "",
            dim.Render(" SSH CONFIGURATION"),
            "",
            neriloField(" Host", h.Hostname, inspectorWidth),
            neriloField(" Port", neriloValue(h.Port, "22"), inspectorWidth),
            neriloField(" User", neriloValue(h.User, "default"), inspectorWidth),
            neriloField(" Auth", neriloAuth(h), inspectorWidth),
            neriloField(" Group", neriloValue(strings.Join(h.Tags, ", "), "—"), inspectorWidth),
            "",
            dim.Render(" ACTIONS"),
            lipgloss.NewStyle().Foreground(purple).Bold(true).Render(" ↵  Connect"),
            " o  Open SFTP",
            " e  Edit Connection",
        )
    } else {
        detail = append(detail, dim.Render(" Select a host to inspect"))
    }
    inspector := lipgloss.NewStyle().Width(inspectorWidth-2).Height(availableHeight).
        Padding(0, 1).Background(panelBg).
        Border(lipgloss.NormalBorder(), false, false, false, true).
        BorderForeground(border).Render(strings.Join(detail, "\n"))

    columns := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, center, inspector)
    help := lipgloss.NewStyle().Background(panelBg).Foreground(muted).
        Width(m.width).Padding(0, 1).Render("↑↓ Navigate   Enter Connect   / Search   ? Help   Esc Back")
    return lipgloss.JoinVertical(lipgloss.Left, strip, tabs, columns, help)
}

func neriloValue(value, fallback string) string {
    if value == "" {
        return fallback
    }
    return value
}

func neriloAuth(host config.SSHHost) string {
    if host.Identity != "" {
        return "SSH Key"
    }
    if host.IdentityAgent != "" {
        return "Agent"
    }
    return "Default SSH"
}

func neriloField(label, value string, width int) string {
    key := lipgloss.NewStyle().Foreground(lipgloss.Color("#8195B1")).Render(label)
    maxValueWidth := width - lipgloss.Width(label) - 5
    if maxValueWidth < 5 {
        maxValueWidth = 5
    }
    val := ansi.Truncate(value, maxValueWidth, "…")
    gap := width - 3 - lipgloss.Width(label) - lipgloss.Width(val)
    if gap < 1 {
        gap = 1
    }
    return key + strings.Repeat(" ", gap) + val
}
