package skill

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"openhue-cli/openhue"
	"openhue-cli/openhue/test/assert"
)

var content = []byte("---\nname: openhue\n---\n# openhue\n")

func TestNewCmdInstallSkill(t *testing.T) {
	cmd := NewCmdInstallSkill(openhue.NewTestIOStreamsDiscard(), content)

	assert.ThatCmdUseIs(t, cmd, "install-skill")
	assert.ThatCmdGroupIs(t, cmd, "config")
}

func TestInstall(t *testing.T) {
	dir := t.TempDir()
	expected := filepath.Join(dir, "openhue", "SKILL.md")

	path, status, err := Install(content, dir)
	require.NoError(t, err)
	require.Equal(t, expected, path)
	require.Equal(t, "installed", status)

	written, err := os.ReadFile(expected)
	require.NoError(t, err)
	require.Equal(t, content, written)

	_, status, err = Install(content, dir)
	require.NoError(t, err)
	require.Equal(t, "already up to date", status)

	newer := append(content, []byte("more\n")...)
	_, status, err = Install(newer, dir)
	require.NoError(t, err)
	require.Equal(t, "updated", status)

	written, err = os.ReadFile(expected)
	require.NoError(t, err)
	require.Equal(t, newer, written)
}

func TestInstall_EmptyContent(t *testing.T) {
	_, _, err := Install(nil, t.TempDir())
	require.Error(t, err)
}

func TestCmdInstallSkill_Dir(t *testing.T) {
	dir := t.TempDir()
	streams, _, out, _ := openhue.NewTestIOStreams()

	cmd := NewCmdInstallSkill(streams, content)
	cmd.SetArgs([]string{"--dir", dir})
	require.NoError(t, cmd.Execute())

	require.FileExists(t, filepath.Join(dir, "openhue", "SKILL.md"))
	require.Contains(t, out.String(), "[OK] Skill installed in")
}

func TestCmdInstallSkill_Print(t *testing.T) {
	streams, _, out, _ := openhue.NewTestIOStreams()

	cmd := NewCmdInstallSkill(streams, content)
	cmd.SetArgs([]string{"--print"})
	require.NoError(t, cmd.Execute())

	require.Equal(t, string(content), out.String())
}

func TestCmdInstallSkill_UserDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", dir)

	cmd := NewCmdInstallSkill(openhue.NewTestIOStreamsDiscard(), content)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	require.FileExists(t, filepath.Join(dir, "skills", "openhue", "SKILL.md"))
}
