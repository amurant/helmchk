package pkg

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func ListValuesPaths(valuesPath string) ([]string, error) {
	valuesBytes, err := os.ReadFile(valuesPath)
	if err != nil {
		return nil, err
	}

	var decoded interface{}
	err = yaml.Unmarshal(valuesBytes, &decoded)
	if err != nil {
		return nil, err
	}

	paths := map[string]struct{}{}
	walkValues(decoded, "$", func(path string, value interface{}) {
		paths[path] = struct{}{}
	})

	paths = makeUniform(paths)

	// sort paths
	values := []string{}
	for key := range paths {
		values = append(values, key)
	}

	return values, nil
}

func walkValues(
	node interface{},
	parentPath string,
	foundPathFn func(path string, value interface{}),
) {
	path := parentPath

	foundPathFn(path, node)

	if node == nil {
		return
	}

	switch tn := node.(type) {
	case map[interface{}]interface{}:
		for key, value := range tn {
			walkValues(value, fmt.Sprintf("%s.%s", path, key), foundPathFn)
		}
	case map[string]interface{}:
		for key, value := range tn {
			walkValues(value, fmt.Sprintf("%s.%s", path, key), foundPathFn)
		}
	case []interface{}:
		for _, value := range tn {
			walkValues(value, fmt.Sprintf("%s.[*]", path), foundPathFn)
		}
	}
}
