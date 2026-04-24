package dsl

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	glazerunner "github.com/go-go-golems/glazed/pkg/cmds/runner"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestHostDiscoversEmbeddedVerbs(t *testing.T) {
	host, err := NewHost()
	require.NoError(t, err)

	paths := []string{}
	for _, verb := range host.registry.Verbs() {
		paths = append(paths, verb.FullPath())
	}
	sort.Strings(paths)

	require.Contains(t, paths, "script compare region")
	require.Contains(t, paths, "script compare brief")
}

func TestEmbeddedCompareCommandsExecute(t *testing.T) {
	left := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" style="padding: 12px 20px; color: rgb(255, 255, 255); background: rgb(96, 45, 72); border-radius: 8px;">Book now</button></body></html>`)
	}))
	defer left.Close()

	right := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><button id="cta" style="padding: 10px 16px; color: rgb(255, 255, 255); background: rgb(112, 61, 89); border-radius: 0px;">Book now</button></body></html>`)
	}))
	defer right.Close()

	host, err := NewHost()
	require.NoError(t, err)
	commandMap := mustCommandMap(t, host)
	regionOutDir := t.TempDir()
	briefOutDir := t.TempDir()

	rows := runCommand(t, commandMap["script compare region"], map[string]map[string]interface{}{
		"targets": {
			"leftUrl":  left.URL,
			"rightUrl": right.URL,
		},
		"viewport": {
			"width":  390,
			"height": 844,
		},
		"output": {
			"outDir":    regionOutDir,
			"writePngs": true,
		},
		"selectors": {
			"leftSelector":  "#cta",
			"rightSelector": "#cta",
		},
	})
	require.Len(t, rows, 1)
	require.NotNil(t, rows[0]["computed_diffs"])
	require.NotNil(t, rows[0]["pixel_diff"])

	text := runWriterCommand(t, commandMap["script compare brief"], map[string]map[string]interface{}{
		"default": {
			"question": "What should change?",
		},
		"targets": {
			"leftUrl":  left.URL,
			"rightUrl": right.URL,
		},
		"viewport": {
			"width":  390,
			"height": 844,
		},
		"output": {
			"outDir":    briefOutDir,
			"writePngs": true,
		},
		"selectors": {
			"leftSelector":  "#cta",
			"rightSelector": "#cta",
		},
	})
	require.Contains(t, text, "What should change?")
	require.Contains(t, text, "- ")
}

func mustCommandMap(t *testing.T, host *Host) map[string]cmds.Command {
	t.Helper()
	commands, err := host.Commands()
	require.NoError(t, err)
	ret := map[string]cmds.Command{}
	for i, command := range commands {
		ret[host.registry.Verbs()[i].FullPath()] = command
	}
	return ret
}

func runCommand(t *testing.T, command cmds.Command, valuesBySection map[string]map[string]interface{}) []map[string]interface{} {
	t.Helper()
	parsedValues, err := glazerunner.ParseCommandValues(command, glazerunner.WithValuesForSections(valuesBySection))
	require.NoError(t, err)

	glazeCommand, ok := command.(cmds.GlazeCommand)
	require.True(t, ok)

	gp := &captureProcessor{}
	err = glazeCommand.RunIntoGlazeProcessor(context.Background(), parsedValues, gp)
	require.NoError(t, err)
	err = gp.Close(context.Background())
	require.NoError(t, err)

	rows := make([]map[string]interface{}, 0, len(gp.rows))
	for _, row := range gp.rows {
		rows = append(rows, rowToMap(row))
	}
	return rows
}

func runWriterCommand(t *testing.T, command cmds.Command, valuesBySection map[string]map[string]interface{}) string {
	t.Helper()
	parsedValues, err := glazerunner.ParseCommandValues(command, glazerunner.WithValuesForSections(valuesBySection))
	require.NoError(t, err)

	writerCommand, ok := command.(cmds.WriterCommand)
	require.True(t, ok)

	var b strings.Builder
	err = writerCommand.RunIntoWriter(context.Background(), parsedValues, &b)
	require.NoError(t, err)
	return b.String()
}

func rowToMap(row types.Row) map[string]interface{} {
	ret := map[string]interface{}{}
	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		ret[pair.Key] = pair.Value
	}
	return ret
}

type captureProcessor struct {
	rows []types.Row
}

func (c *captureProcessor) AddRow(_ context.Context, row types.Row) error {
	c.rows = append(c.rows, row)
	return nil
}

func (c *captureProcessor) Close(context.Context) error {
	return nil
}
