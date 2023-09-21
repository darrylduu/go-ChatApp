package src

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
	"time"
)

const myAppVersion = "ZADRDJM"

type UI struct {
	*ChatRoom
	TerminalApp *tview.Application
	MsgInputs   chan string
	CmdInputs   chan uicommand
	peerBox     *tview.TextView
	messageBox  *tview.TextView
	inputBox    *tview.InputField
}
type uicommand struct {
	cmdtype string
	cmdarg  string
}

func NewUI(cr *ChatRoom) *UI {
	app := tview.NewApplication()
	cmdchan := make(chan uicommand)
	msgchan := make(chan string)

	titlebox := tview.NewTextView().
		SetText(fmt.Sprintf("PeerCht. A P2P Chat Application. %s", myAppVersion)).
		SetTextColor(tcell.ColorWhite).
		SetTextAlign(tview.AlignCenter)

	titlebox.SetBorder(true).SetBorderColor(tcell.ColorGreen)

	messagebox := tview.NewTextView().SetDynamicColors(true).SetChangedFunc(func() {
		app.Draw()
	})

	messagebox.SetBorder(true).SetBorderColor(tcell.ColorGreen).
		SetTitle(fmt.Sprintf("ChatRoom-%s", cr.RoomName)).
		SetTitleAlign(tview.AlignLeft).SetTitleColor(tcell.ColorWhite)

	usage := tview.NewTextView().
		SetDynamicColors(true).
		SetText(`[red]/quit[green] - quit the chat | [red]/room <roomname>[green] - change chat room | [red]/user <username>[green] - change user name | [red]/clear[green] - clear the chat`)
	usage.SetBorder(true).SetBorderColor(tcell.ColorGreen).
		SetTitle("Usage").SetTitleAlign(tview.AlignLeft).
		SetTitleColor(tcell.ColorWhite).SetBorderPadding(0, 0, 1, 0)

	peerbox := tview.NewTextView()
	peerbox.SetBorder(true).SetBorderColor(tcell.ColorGreen).
		SetTitle("Peers").SetTitleAlign(tview.AlignLeft).SetTitleColor(tcell.ColorWhite)

	input := tview.NewInputField().SetLabel(cr.UserName + " > ").
		SetLabelColor(tcell.ColorGreen).SetFieldWidth(0).SetFieldBackgroundColor(tcell.ColorBlack)

	input.SetBorder(true).SetBorderColor(tcell.ColorGreen).
		SetTitle("Input").SetTitleAlign(tview.AlignLeft).
		SetTitleColor(tcell.ColorWhite).SetBorderPadding(0, 0, 1, 0)

	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		line := input.GetText()
		if len(line) == 0 {
			return
		}
		if strings.HasPrefix(line, "/") {
			cmdparts := strings.Split(line, " ")

			if len(cmdparts) == 1 {
				cmdparts = append(cmdparts, "")
			}
			cmdchan <- uicommand{cmdtype: cmdparts[0], cmdarg: cmdparts[1]}
		} else {
			msgchan <- line
		}
		input.SetText(" ")
	})
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(titlebox, 3, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(messagebox, 0, 1, false).
			AddItem(peerbox, 20, 1, false),
			0, 8, false).AddItem(input, 3, 1, true).
		AddItem(usage, 3, 1, false)

	app.SetRoot(flex, true)
	return &UI{
		ChatRoom:    cr,
		TerminalApp: app,
		peerBox:     peerbox,
		messageBox:  messagebox,
		inputBox:    input,
		MsgInputs:   msgchan,
		CmdInputs:   cmdchan,
	}
}

func (ui *UI) Run() error {
	go ui.starteventhandler()
	defer ui.Close()
	return ui.TerminalApp.Run()
}
func (ui *UI) Close() {
	ui.pscancel()
}
func (ui *UI) starteventhandler() {
	refreshticker := time.NewTicker(time.Second)
	defer refreshticker.Stop()
	for {
		select {
		case msg := <-ui.MsgInputs:
			ui.Outbound <- msg
			ui.display_selfMessage(msg)
		case cmd := <-ui.CmdInputs:
			go ui.handleCommand(cmd)
		case msg := <-ui.Inbound:
			ui.dsiplay_chatMessage(msg)
		case log := <-ui.Logs:
			ui.display_logMessage(log)
		case <-refreshticker.C:
			ui.syncpeerbox()
		case <-ui.psctx.Done():
			return
		}
	}
}
func (ui *UI) display_selfMessage(msg string) {
	prompt := fmt.Sprintf("[blue]<%s>:[-]", ui.UserName)
	fmt.Fprintf(ui.messageBox, "%s %s\n", prompt, msg)
}

func (ui *UI) handleCommand(cmd uicommand) {
	switch cmd.cmdtype {
	case "/quit":
		ui.TerminalApp.Stop()
		return
	case "/clear":
		ui.messageBox.Clear()
	case "/room":
		if cmd.cmdarg == "" {
			ui.Logs <- chatlog{logprefix: "badcmd", logmsg: "missing room name for the cmd"}
		} else {
			ui.Logs <- chatlog{logprefix: "roomChnage",
				logmsg: fmt.Sprintf("Joining new room %s", cmd.cmdarg)}
			oldChatRoom := ui.ChatRoom
			newChatRoom, err := JoinChatRoom(ui.Host, ui.UserName, cmd.cmdarg)
			if err != nil {
				ui.Logs <- chatlog{logprefix: "exception",
					logmsg: fmt.Sprintf("Could not change chat room 0 %s", err)}
				return
			}
			ui.ChatRoom = newChatRoom
			oldChatRoom.Exit()
			ui.messageBox.Clear()
			ui.messageBox.SetTitle(fmt.Sprintf("ChatRoom-%s", ui.ChatRoom.RoomName))
		}
	case "/user":
		if cmd.cmdarg == "" {
			ui.Logs <- chatlog{logprefix: "badcmd", logmsg: "Missing user name for the cmd"}
		} else {
			ui.UpdateUser(cmd.cmdarg)
			ui.inputBox.SetTitle(ui.UserName + " > ")
		}
	default:
		ui.Logs <- chatlog{logprefix: "badcmd",
			logmsg: fmt.Sprintf("Unsupported command %s", cmd.cmdtype)}
	}
}
func (ui *UI) dsiplay_chatMessage(msg chatmessage) {
	prompt := fmt.Sprintf("[green]<%s>:[-]", msg.SenderName)
	fmt.Fprintf(ui.messageBox, "%s %s\n", prompt, msg.Message)
}
func (ui *UI) display_logMessage(log chatlog) {
	prompt := fmt.Sprintf("[yellow]<%s>:[-]", log.logprefix)
	fmt.Fprintf(ui.messageBox, "%s %s\n", prompt, log.logmsg)
}
func (ui *UI) syncpeerbox() {
	peers := ui.PeerList()
	ui.peerBox.Lock()
	ui.peerBox.Clear()
	ui.peerBox.Unlock()

	for _, p := range peers {
		peerid := p.Pretty()
		peerid = peerid[len(peerid)-8:]
		fmt.Fprintln(ui.peerBox, peerid)
	}
	ui.TerminalApp.Draw()
}
