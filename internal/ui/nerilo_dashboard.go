package ui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/x/ansi"
    "github.com/zsuroy/ctty/internal/config"
)

// The dashboard is strictly a view over the existing model. All shortcuts,
// SSH connections, searches and dialogs are handled by the original state machine.
func (m Model) renderNeriloDashboard() string {
    const sidebarWidth, inspectorWidth = 23, 39
    w := min(m.width-4, 148)
    h := min(m.height-4, 43)
    if w < 108 || h < 22 {
        return m.styles.App.Render(m.styles.Header.Render("◈ NERILO / Connections") + "\n" + m.renderTableView())
    }
    p := newNeriloPalette()
    centerWidth := w-sidebarWidth-inspectorWidth
    bodyHeight := h-5

    brand := neriloBlock(w, 2, p.chrome, "  "+p.title.Render("◈ NERILO")+"  "+p.dim.Render("/ remote workspace")+"      "+p.highlight.Render("CONNECTION MANAGER"))
    nav := neriloBlock(w, 2, p.background,
        "  "+p.key.Render("SSH")+"   "+p.dim.Render("t Serial   T Telnet   F FTP   W WebDAV   O S3   b Local Files")+
        "\n  "+p.dim.Render("WORKSPACE / Connections"))
    sidebar := m.neriloSidebar(sidebarWidth, bodyHeight, p)
    list := m.neriloHostList(centerWidth, bodyHeight, p)
    inspector := m.neriloInspector(inspectorWidth, bodyHeight, p)
    footer := neriloBlock(w, 1, p.chrome, "  "+p.key.Render("↑↓ / jk")+" Move   "+p.key.Render("Enter")+" Connect   "+p.key.Render("/")+" Search   "+p.key.Render("o")+" SFTP   "+p.key.Render("f")+" Forward   "+p.key.Render("a")+" Add   "+p.key.Render("?")+" Help")
    board := lipgloss.JoinVertical(lipgloss.Left, brand, nav, lipgloss.JoinHorizontal(lipgloss.Top, sidebar, list, inspector), footer)
    return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, board,
        lipgloss.WithWhitespaceBackground(p.background))
}

type neriloPalette struct {
    background, chrome, sidebar, panel, selected lipgloss.Color
    text, dim, title, highlight, key, success lipgloss.Style
}

func newNeriloPalette() neriloPalette {
    return neriloPalette{
        background: lipgloss.Color("#0B111B"),
        chrome: lipgloss.Color("#141C2A"),
        sidebar: lipgloss.Color("#101723"),
        panel: lipgloss.Color("#172132"),
        selected: lipgloss.Color("#263751"),
        text: lipgloss.NewStyle().Foreground(lipgloss.Color("#C5D0E2")),
        dim: lipgloss.NewStyle().Foreground(lipgloss.Color("#7E90A9")),
        title: lipgloss.NewStyle().Foreground(lipgloss.Color("#91B7F8")).Bold(true),
        highlight: lipgloss.NewStyle().Foreground(lipgloss.Color("#C2B2FF")),
        key: lipgloss.NewStyle().Foreground(lipgloss.Color("#AFCBFF")).Bold(true),
        success: lipgloss.NewStyle().Foreground(lipgloss.Color("#8ED8BB")),
    }
}

func neriloBlock(w, h int, bg lipgloss.Color, text string) string {
    return lipgloss.NewStyle().Width(w).Height(h).Background(bg).MaxWidth(w).MaxHeight(h).Render(text)
}

func (m Model) neriloSidebar(w, h int, p neriloPalette) string {
    rows := []string{
        "", "  "+p.title.Render("◈  nerilo"), "  "+p.dim.Render("REMOTE HUB"),
        "", "  "+p.dim.Render("WORKSPACE"),
        "  "+p.key.Render("▸")+" "+p.text.Render("SSH Hosts"),
        "    "+p.dim.Render("Serial"),
        "    "+p.dim.Render("Telnet"),
        "    "+p.dim.Render("FTP / WebDAV"),
        "    "+p.dim.Render("S3 Storage"),
        "    "+p.dim.Render("Local Files"),
        "", "  "+p.dim.Render("LIBRARY"),
        fmt.Sprintf("  Hosts       %d",len(m.hosts)),
        "", "  "+p.dim.Render("QUICK ACTIONS"),
        "  "+p.key.Render("a")+" Add host",
        "  "+p.key.Render("w")+" Tag filter",
        "  "+p.key.Render("p")+" Ping hosts",
        "  "+p.key.Render("?")+" Help",
    }
    if m.selectedTag != "" {
        rows = append(rows, "", "  "+p.highlight.Render(ansi.Truncate("#"+m.selectedTag,w-4,"…")))
    }
    return neriloBlock(w,h,p.sidebar,strings.Join(rows,"\n"))
}

func (m Model) neriloHostList(w,h int,p neriloPalette) string {
    hosts:=m.filteredHosts
    if hosts==nil {hosts=m.hosts}
    cursor:=m.table.Cursor()
    if cursor<0 {cursor=0}
    if cursor>=len(hosts)&&len(hosts)>0 {cursor=len(hosts)-1}
    inner:=w-5
    rows:=[]string{
        "", "  "+p.title.Render("Connections")+p.dim.Render(fmt.Sprintf("  %d hosts",len(hosts))),
        "  "+p.dim.Render("Manage your remote environments"),
        "", "  "+p.dim.Render("────────────────────────────────────"),
    }
    query:=m.searchInput.Value()
    if query=="" {query="Find host, IP or tag…"}
    query=ansi.Truncate(query,inner-5,"…")
    if m.searchMode {
        rows=append(rows,"  "+p.key.Render("⌕ ")+p.text.Render(query))
    } else {
        rows=append(rows,"  "+p.dim.Render("⌕ "+query)+"  "+p.key.Render("/"))
    }
    rows=append(rows,"","  "+p.dim.Render("ALL CONNECTIONS"),"")
    perPage:=(h-len(rows)-4)/3
    if perPage<1 {perPage=1}
    start:=max(0,cursor-perPage+1)
    if start+perPage>len(hosts) {start=max(0,len(hosts)-perPage)}
    end:=min(len(hosts),start+perPage)
    if len(hosts)==0 {
        rows=append(rows,"  "+p.dim.Render("No matching connections"),"  "+p.dim.Render("Press a to add a host"))
    }
    for i:=start;i<end;i++ {
        host:=hosts[i]
        prefix:="  "
        if m.selectedHosts[host.Name] {prefix="✓ "}
        top:=prefix+"▣ "+ansi.Truncate(host.Name,inner-7,"…")
        bottom:="     "+ansi.Truncate(neriloValue(host.Hostname,host.Name),inner-9,"…")
        if len(host.Tags)>0 {bottom+="  "+ansi.Truncate("#"+strings.Join(host.Tags,","),16,"…")}
        if i==cursor {
            selected:=lipgloss.NewStyle().Foreground(lipgloss.Color("#E3EAFA")).Background(p.selected)
            rows=append(rows,"  "+selected.Width(inner).Render("▸ "+ansi.Truncate(top,inner-2,"…")),
                "  "+selected.Width(inner).Render(ansi.Truncate(bottom,inner,"…")))
        } else {
            rows=append(rows,"  "+p.text.Render(ansi.Truncate(top,inner,"…")),
                "  "+p.dim.Render(ansi.Truncate(bottom,inner,"…")))
        }
        rows=append(rows,"")
    }
    if len(hosts)>perPage {
        rows=append(rows,"  "+p.dim.Render(fmt.Sprintf("Showing %d–%d of %d",start+1,end,len(hosts))))
    }
    if m.showingError&&m.errorMessage!="" {
        rows=append(rows,"  "+lipgloss.NewStyle().Foreground(lipgloss.Color("#F28B9D")).Render(ansi.Truncate(m.errorMessage,inner,"…")))
    } else if m.statusActive() {
        rows=append(rows,"  "+p.success.Render(ansi.Truncate(m.statusMessage,inner,"…")))
    }
    return neriloBlock(w,h,p.background,strings.Join(rows,"\n"))
}

func (m Model) neriloInspector(w,h int,p neriloPalette) string {
    rows:=[]string{"","  "+p.dim.Render("CONNECTION DETAILS"),""}
    hosts:=m.filteredHosts
    if hosts==nil {hosts=m.hosts}
    index:=m.table.Cursor()
    if index<0 {index=0}
    if index>=len(hosts) {index=len(hosts)-1}
    if index<0 {
        rows=append(rows,"  "+p.dim.Render("Select a connection"))
        return neriloBlock(w,h,p.panel,strings.Join(rows,"\n"))
    }
    host:=hosts[index]
    rows=append(rows,
        "  "+p.title.Render("▣  "+ansi.Truncate(host.Name,w-8,"…")),
        "  "+p.dim.Render(ansi.Truncate(neriloValue(host.User,"user")+"@"+neriloValue(host.Hostname,host.Name),w-5,"…")),
        "", "  "+p.dim.Render("SSH CONFIGURATION"),"",
        neriloField("Host",neriloValue(host.Hostname,host.Name),w),
        neriloField("Port",neriloValue(host.Port,"22"),w),
        neriloField("User",neriloValue(host.User,"default"),w),
        neriloField("Auth",neriloAuth(host),w),
        neriloField("Group",neriloValue(strings.Join(host.Tags,", "),"—"),w),
        "", "  "+p.dim.Render("QUICK ACTIONS"),
        "  "+p.key.Render("↵")+"  Connect via SSH",
        "  "+p.key.Render("o")+"  Open SFTP explorer",
        "  "+p.key.Render("f")+"  Port forwarding",
        "  "+p.key.Render("e")+"  Edit connection",
        "  "+p.key.Render("v")+"  Quick peek",
    )
    return neriloBlock(w,h,p.panel,strings.Join(rows,"\n"))
}

func neriloValue(value,fallback string) string {
    if value=="" {return fallback}
    return value
}

func neriloAuth(host config.SSHHost) string {
    if host.Identity!="" {return "SSH Key"}
    if host.IdentityAgent!="" {return "Agent"}
    return "Default SSH"
}

func neriloField(label,value string,width int) string {
    key:=lipgloss.NewStyle().Foreground(lipgloss.Color("#8195B1")).Render("  "+label)
    maxValueWidth:=max(5,width-lipgloss.Width(label)-8)
    val:=ansi.Truncate(value,maxValueWidth,"…")
    gap:=max(1,width-4-lipgloss.Width(label)-lipgloss.Width(val))
    return key+strings.Repeat(" ",gap)+val
}
