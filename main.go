package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/amurant/helmchk/pkg"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

const (
	helmchkCommand         = "helmchk [local Helm chart path]"
	helmchkDescription     = "Verify that the Helm chart values.yaml and template variables are in sync."
	helmchkLongDescription = "helmchk is a cli tool that can extract all the variables used in the templates of a" +
		"Helm chart and compare them with the default values configured in the values.yaml file."
	helmchkExample = `  $ helm pull jetstack/cert-manager --untar
  $ helmchk ./cert-manager/
  value missing from values.yaml: .$.acmesolver.image.tag
  value missing from values.yaml: .$.automountServiceAccountToken
  ...

  # Learn what the allowed exceptions are
  $ helmchk ./my-chart/ > exceptions.txt

  # Run helmchk and ignore the exceptions
  $ helmchk ./my-chart/ --exceptions exceptions.txt`
)

type options struct {
	valuesPath    string
	templatesPath string

	exceptionsPath string
}

func main() {
	o := &options{
		valuesPath:    path.Join(".", "values.yaml"),
		templatesPath: path.Join(".", "templates"),

		exceptionsPath: "",
	}

	cmd := &cobra.Command{
		Use:           helmchkCommand,
		Short:         helmchkDescription,
		Long:          helmchkLongDescription,
		Example:       helmchkExample,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires a single argument: the path to the Helm chart")
			}

			if _, err := os.Stat(args[0]); err != nil {
				return fmt.Errorf("invalid path: %s", args[0])
			}

			valuesPath := path.Join(args[0], o.valuesPath)
			templatesPath := path.Join(args[0], o.templatesPath)

			if _, err := os.Stat(valuesPath); err != nil {
				return fmt.Errorf("the Helm chart does not contain a values.yaml file: %s", valuesPath)
			}

			if _, err := os.Stat(templatesPath); err != nil {
				return fmt.Errorf("the Helm chart does not contain a templates directory: %s", templatesPath)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			valuesPath := path.Join(args[0], o.valuesPath)
			templatesPath := path.Join(args[0], o.templatesPath)

			return compareValues(valuesPath, templatesPath, o.exceptionsPath)
		},
	}

	cmd.Flags().StringVar(&o.valuesPath, "values", o.valuesPath, "path to the values.yaml file, relative to chart dir (default: ./values.yaml)")
	cmd.Flags().StringVar(&o.templatesPath, "templates", o.templatesPath, "path to the templates directory, relative to chart dir (default: ./templates)")

	cmd.Flags().StringVar(&o.exceptionsPath, "exceptions", o.exceptionsPath, "path to the file containing the list of exceptions")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func compareValues(
	valuesPath string,
	templatesPath string,
	exceptionsPath string,
) error {
	valuePaths, err := pkg.ListValuesPaths(valuesPath)
	if err != nil {
		return err
	}

	templatePaths, err := pkg.ListTemplatePaths(templatesPath)
	if err != nil {
		return err
	}

	exceptionStrings := []string{}
	if exceptionsPath != "" {
		exceptionsPathsRaw, err := os.ReadFile(exceptionsPath)
		if err != nil {
			return err
		}

		exceptionStrings = strings.Split(string(exceptionsPathsRaw), "\n")
	}

	sort.Strings(valuePaths)
	sort.Strings(templatePaths)
	sort.Strings(exceptionStrings)

	succeeded := true
	prefix := ""
	var i, j int
	for i < len(valuePaths) && j < len(templatePaths) {
		if valuePaths[i] == templatePaths[j] {
			prefix = valuePaths[i]
			i++
			j++
		} else if valuePaths[i] < templatePaths[j] {
			if !strings.HasPrefix(valuePaths[i], prefix) {
				exceptionString := fmt.Sprintf("value missing from templates: %s", valuePaths[i])

				if !slices.Contains(exceptionStrings, exceptionString) {
					fmt.Println(exceptionString)
					succeeded = false
				}
			}
			i++
		} else {
			if !strings.HasPrefix(templatePaths[j], prefix) {
				exceptionString := fmt.Sprintf("value missing from values.yaml: %s", templatePaths[j])

				if !slices.Contains(exceptionStrings, exceptionString) {
					fmt.Println(exceptionString)
					succeeded = false
				}
			}
			j++
		}
	}

	if !succeeded {
		return fmt.Errorf("values.yaml and templates are not in sync")
	}

	return nil
}
