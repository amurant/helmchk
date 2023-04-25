package pkg_test

import (
	"fmt"
	"sort"
	"testing"
	"text/template"

	"github.com/amurant/helmchk/pkg"
	"github.com/amurant/helmchk/pkg/funcs_serdes"
	"github.com/stretchr/testify/require"
)

func TestListTemplatePathsFromTemplates(t *testing.T) {
	type testcase struct {
		templates     []string
		expectedPaths []string
	}

	testcases := []testcase{
		{
			templates: []string{
				"{{ .Values.foo }}",
			},
			expectedPaths: []string{
				".$",
				".$.foo",
			},
		},
		{
			templates: []string{
				"{{ .Values.foo }}",
				"{{ .Values.bar }}",
			},
			expectedPaths: []string{
				".$",
				".$.foo",
				".$.bar",
			},
		},
		{
			templates: []string{
				"{{ .Values.foo }}",
				"{{ .Values.foo }}",
			},
			expectedPaths: []string{
				".$",
				".$.foo",
			},
		},
		{
			templates: []string{
				"{{ .Values.foo }}",
				"{{ .Values.foo }}",
				"{{ .Values.bar }}",
			},
			expectedPaths: []string{
				".$",
				".$.foo",
				".$.bar",
			},
		},
		{
			templates: []string{
				"{{ .foo }}",
				"{{ .Values.bar }}",
			},
			expectedPaths: []string{
				".$",
				".$.bar",
			},
		},
		{
			templates: []string{
				"{{ range $key, $value := .Values.test }}{{ end }}",
			},
			expectedPaths: []string{
				".$",
				".$.test",
				".$.test.[*]",
			},
		},
		{
			templates: []string{
				"{{ $aa := .Values.test1.test2 }}",
			},
			expectedPaths: []string{
				".$",
				".$.test1",
				".$.test1.test2",
			},
		},
		{
			templates: []string{
				"{{ $aa := .Values.test1 }}{{ $bb := $aa.test2 }}{{ $bb.test3 }}",
			},
			expectedPaths: []string{
				".$",
				".$.test1",
				".$.test1.test2",
				".$.test1.test2.test3",
			},
		},
		{
			templates: []string{
				"{{ $value := .Values.test }}{{ $value.value }}",
			},
			expectedPaths: []string{
				".$",
				".$.test",
				".$.test.value",
			},
		},
		{
			templates: []string{
				"{{ $value := .Values.test1 }}{{ $value := .Values.test2 }}{{ $value.value }}",
			},
			expectedPaths: []string{
				".$",
				".$.test1",
				".$.test1.value",
				".$.test2",
				".$.test2.value",
			},
		},
		{
			templates: []string{
				"{{ range $key, $value := .Values.test }}{{ $key.key }}{{ $value.value.test1 }}{{ end }}",
			},
			expectedPaths: []string{
				".$",
				".$.test",
				".$.test.[*]",
				".$.test.[*].key",
				".$.test.[*].value",
				".$.test.[*].value.test1",
			},
		},
		{
			templates: []string{
				"{{ with .Values.test1 }}{{ .test2 }}{{ end }}",
			},
			expectedPaths: []string{
				".$",
				".$.test1",
				".$.test1.test2",
			},
		},
		{
			templates: []string{
				"{{ with .Values.test1 }}{{ . }}{{ end }}",
			},
			expectedPaths: []string{
				".$",
				".$.test1",
			},
		},
		{
			templates: []string{
				"{{ if .Values.test1 }}{{ . }}{{ .Values.test2 }}{{ end }}",
			},
			expectedPaths: []string{
				".$",
				".$.test1",
				".$.test2",
			},
		},
		{
			templates: []string{
				"{{define \"T1\" }}{{ .test2 }}{{end}} {{ .Values.foo }}",
				"{{ template \"T1\" .Values.test1 }}",
				"{{ .Values.bar }}",
			},
			expectedPaths: []string{
				".$",
				".$.test1",
				".$.test1.test2",
				".$.foo",
				".$.bar",
			},
		},
		{
			templates: []string{
				"{{define \"T1\" }}{{ .test1 }}{{end}}",
				"{{define \"T2\" }}{{ template \"T1\" .test2 }}{{end}}",
				"{{define \"T3\" }}{{ template \"T1\" .test2 }}{{ template \"T2\" .test3 }}{{end}}",
				"{{ template \"T1\" .Values.test1 }}{{ template \"T2\" .Values.test1 }}{{ template \"T3\" .Values.test1 }}",
			},
			expectedPaths: []string{
				".$",
				".$.test1",
				".$.test1.test1",
				".$.test1.test2",
				".$.test1.test2.test1",
				".$.test1.test3",
				".$.test1.test3.test2",
				".$.test1.test3.test2.test1",
			},
		},
		{
			templates: []string{
				"{{ $name := default .Values.test1 .Values.test2 }}{{ $name.test3 }}",
			},
			expectedPaths: []string{
				".$",
				".$.test1",
				".$.test1.test3",
				".$.test2",
				".$.test2.test3",
			},
		},
	}

	for _, tc := range testcases {
		tmpl := template.New("ROOT")

		tmpl.Funcs(funcs_serdes.FuncMap())

		templates := map[*template.Template]struct{}{}
		for idx, tem := range tc.templates {
			tpl, err := tmpl.New(fmt.Sprintf("input-item-%d", idx)).Parse(tem)
			if err != nil {
				t.Errorf("error parsing template: %s", err)
			}

			templates[tpl] = struct{}{}
		}

		paths, err := pkg.ListTemplatePathsFromTemplates(tmpl, templates)
		if err != nil {
			t.Errorf("error listing template paths: %s", err)
		}

		sort.Strings(tc.expectedPaths)
		sort.Strings(paths)

		require.EqualValues(t, tc.expectedPaths, paths)
	}
}
