package gitwrap

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gookit/goutil/cliutil"
	"github.com/gookit/goutil/fsutil"
	"github.com/gookit/goutil/sysutil"
	"github.com/gookit/slog"
)

// func Editor() (string, error) {
// 	varCmd := gitCmd("var", "GIT_EDITOR")
// 	varCmd.Stderr = nil
// 	output, err := varCmd.Output()
// 	if err != nil {
// 		return "", fmt.Errorf("can't load git var: GIT_EDITOR")
// 	}
//
// 	return os.ExpandEnv(firstLine(output)), nil
// }

var editorCmd string

// Editor returns program name of the editor.
func Editor() string {
	if editorCmd != "" {
		return editorCmd
	}

	var env, str string
	if env = os.Getenv("GIT_EDITOR"); env != "" {
		str = env
	} else if env = Var("GIT_EDITOR"); env != "" { // git var GIT_EDITOR
		str = env
	} else if env = Config("core.editor"); env != "" { // git config --get core.editer OR git config core.editer
		str = env
	} else if env = os.Getenv("VISUAL"); env != "" {
		str = env
	} else if env = os.Getenv("EDITOR"); env != "" {
		str = env
	} else if os.Getenv("TERM") == "dumb" {
		slog.Fatal(
			"No editor specified in GIT_EDITOR, core.editor, VISUAL or EDITOR.\n" +
				"Tried to fall back to vi but terminal is dumb.  Please configure at\n" +
				"least one of these before using this command.")
	} else {
		for _, c := range []string{"vim", "vi", "emacs", "nano"} {
			if path, err := exec.LookPath(c); err == nil {
				str = path
				break
			}
		}
	}

	// remove space and ':'
	editorCmd = strings.Trim(str, ": ")
	return editorCmd
}

// EditText starts an editor to edit data, and returns the edited data.
func EditText(data string) string {
	var (
		err    error
		editor string
	)

	editor = Editor()
	if !sysutil.IsTerminal(os.Stdout.Fd()) {
		slog.Println("no editor, input data unchanged")
		fmt.Println(data)
		return data
	}

	tmpfile, err := ioutil.TempFile("", "go-git-edit-file-*")
	if err != nil {
		slog.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.WriteString(data)
	if err != nil {
		slog.Fatal(err)
	}
	err = tmpfile.Close()
	if err != nil {
		slog.Fatal(err)
	}

	cmdArgs := editorCommands(editor, tmpfile.Name())
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		slog.Errorf("fail to run '%s' to edit script: %s",
			strings.Join(cmdArgs, " "),
			err)
	}

	f, err := os.Open(tmpfile.Name())
	if err != nil {
		slog.Fatal(err)
	}

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		slog.Fatal(err)
	}
	return string(buf)
}

func editorCommands(editor string, args ...string) []string {
	var cmdArgs []string

	if sysutil.IsWindows() {
		// Split on spaces, respecting quoted strings
		if len(editor) > 0 && (editor[0] == '"' || editor[0] == '\'') {
			cmdArgs = cliutil.ParseLine(editor)

			// if err != nil {
			// 	log.Errorf("fail to parse editor '%s': %s", editor, err)
			// 	cmdArgs = append(cmdArgs, editor)
			// }
		} else {
			for i, c := range editor {
				if c == ' ' || c == '\t' {
					if fsutil.PathExists(editor[:i]) {
						cmdArgs = append(cmdArgs, editor[:i])
						inArgs := cliutil.ParseLine(editor[i+1:])
						cmdArgs = append(cmdArgs, inArgs...)

						// inArgs, err := shellwords.Parse(editor[i+1:])
						// if err != nil {
						// 	log.Errorf("fail to parse args'%s': %s", editor[i+1:], err)
						// 	cmdArgs = append(cmdArgs, editor[i+1:])
						// } else {
						// 	cmdArgs = append(cmdArgs, inArgs...)
						// }
						break
					}
				}
			}
			if len(cmdArgs) == 0 {
				cmdArgs = append(cmdArgs, editor)
			}
		}
	} else if regexp.MustCompile(`^.*[$ \t'].*$`).MatchString(editor) {
		// See: https://gerrit-review.googlesource.com/c/git-repo/+/16156
		cmdArgs = append(cmdArgs, "sh", "-c", editor+` "$@"`, "sh")
	} else {
		cmdArgs = append(cmdArgs, editor)
	}

	cmdArgs = append(cmdArgs, args...)
	return cmdArgs
}

func parseRemoteUrl(url string, r *RemoteInfo) *RemoteInfo {
	if r == nil {
		r = &RemoteInfo{Name: DefaultRemoteName, URL: url}
	}

	// eg: "git@github.com:gookit/gitwrap.git"
	if strings.HasPrefix(url, "git@") {

	}


	return r
}
