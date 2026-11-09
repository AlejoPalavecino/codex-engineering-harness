package integrations

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLICatalogGuidesAndExplicitApply(t *testing.T) {
	s := Service{Home: t.TempDir()}
	var out bytes.Buffer
	if err := s.CLI(nil, &out); err != nil {
		t.Fatal(err)
	}
	for _, i := range Catalog() {
		if !strings.Contains(out.String(), i.ID) {
			t.Fatal(i.ID)
		}
		u, err := url.Parse(i.GuideURL)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			t.Fatal(i.GuideURL)
		}
		var guide bytes.Buffer
		if err := s.CLI([]string{"guide", i.ID}, &guide); err != nil || !strings.Contains(guide.String(), i.GuideURL) {
			t.Fatal(guide.String(), err)
		}
	}
	if err := s.CLI([]string{"configure", "exa"}, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(s.Home, "config.toml")); !os.IsNotExist(err) {
		t.Fatal("preview wrote config")
	}
	if err := s.CLI([]string{"configure", "exa", "--apply"}, &out); err != nil {
		t.Fatal(err)
	}
	if err := s.CLI([]string{"undo", "exa"}, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(s.Home, "config.toml")); err != nil {
		t.Fatal("undo preview removed config")
	}
	if err := s.CLI([]string{"undo", "exa", "--apply"}, &out); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"guide", "github", "--apply"}, {"configure", "unknown"}, {"configure", "egolite"}, {"list", "extra"}} {
		if err := s.CLI(args, &out); err == nil {
			t.Fatal("accepted invalid action", args)
		}
	}
}
